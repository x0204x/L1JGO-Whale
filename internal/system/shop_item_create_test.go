package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type shopItemCreateStub struct {
	itemID int32
	count  int32
	calls  int
}

func (s *shopItemCreateStub) GiveItem(_ *l1net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.calls++
	return &world.InvItem{ItemID: itemID, Count: count}, true
}

func TestShopBuyFromNpcUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewShopSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "shop-test",
		Str:  50,
		Con:  50,
		Inv:  world.NewInventory(),
	}
	player.Inv.AddItem(world.AdenaItemID, 1000, "Adena", 318, 0, true, 1)
	shop := &data.Shop{
		SellingItems: []*data.ShopItem{
			{ItemID: 1003, SellingPrice: 50, PackCount: 2},
		},
	}
	w := packet.NewWriterWithOpcode(0)
	w.WriteD(0)
	w.WriteD(3)

	sys.BuyFromNpc(sess, packet.NewReader(w.RawBytes()), 1, player, shop, nil)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1003 || itemCreate.count != 6 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1003 count=6", itemCreate.itemID, itemCreate.count)
	}
}

func TestShopSellToNpcUsesItemCreateForAdena(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewShopSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "shop-test",
		Str:  50,
		Con:  50,
		Inv:  world.NewInventory(),
	}
	item := player.Inv.AddItem(1001, 4, "test-item", 101, 10, true, 1)
	shop := &data.Shop{
		PurchasingItems: []*data.ShopItem{
			{ItemID: 1001, PurchasingPrice: 12},
		},
	}
	w := packet.NewWriterWithOpcode(0)
	w.WriteD(item.ObjectID)
	w.WriteD(3)

	sys.SellToNpc(sess, packet.NewReader(w.RawBytes()), 1, player, shop)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 36 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=36", itemCreate.itemID, itemCreate.count, world.AdenaItemID)
	}
	if item.Count != 1 {
		t.Fatalf("販賣後物品數量錯誤：got %d want 1", item.Count)
	}
}
