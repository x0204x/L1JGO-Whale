package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestShopCnBuyUsesItemCreateWithEnchant(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &boxItemCreateStub{}
	sys := NewShopCnSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "shop-cn-test",
		Inv:  world.NewInventory(),
	}
	player.Inv.AddItem(handler.CnCurrencyItemID, 1000, "商城幣", 610, 0, true, 1)
	cnItem := &data.ShopCnItem{
		ItemID:       1002,
		SellingPrice: 50,
		EnchantLevel: 5,
	}

	sys.BuyCnItem(sess, player, cnItem, 2, 2)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤: got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1002 || itemCreate.count != 2 {
		t.Fatalf("ItemCreate 參數錯誤: got item=%d count=%d want item=1002 count=2", itemCreate.itemID, itemCreate.count)
	}
	if itemCreate.opts.EnchantLvl != 5 {
		t.Fatalf("ItemCreate 強化選項錯誤: got %d want 5", itemCreate.opts.EnchantLvl)
	}
	if got := player.Inv.FindByItemID(handler.CnCurrencyItemID).Count; got != 900 {
		t.Fatalf("商城幣扣除錯誤: got %d want 900", got)
	}
}

func TestShopCnSellUsesItemCreateForCurrency(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewShopCnSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "shop-cn-test",
		Inv:  world.NewInventory(),
	}
	item := player.Inv.AddItem(1001, 5, "test-item", 101, 10, true, 1)

	sys.SellCnItem(sess, player, item, 3, 7)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != handler.CnCurrencyItemID || itemCreate.count != 21 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=21", itemCreate.itemID, itemCreate.count, handler.CnCurrencyItemID)
	}
	if item.Count != 2 {
		t.Fatalf("回收後物品數量錯誤：got %d want 2", item.Count)
	}
}

func TestShopCnSellRejectsFullInventoryWhenCurrencyNeedsNewSlot(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewShopCnSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "shop-cn-test",
		Inv:  world.NewInventory(),
	}
	item := player.Inv.AddItem(1001, 5, "test-item", 101, 10, true, 1)
	for i := player.Inv.Size(); i < world.MaxInventorySize; i++ {
		player.Inv.Items = append(player.Inv.Items, &world.InvItem{
			ObjectID: int32(50_000 + i),
			ItemID:   int32(60_000 + i),
			Count:    1,
		})
	}

	sys.SellCnItem(sess, player, item, 3, 7)

	if itemCreate.calls != 0 {
		t.Fatalf("背包滿且商城幣需要新格時不應發幣：got calls=%d", itemCreate.calls)
	}
	if item.Count != 5 {
		t.Fatalf("拒絕回收時不應扣物品：got %d want 5", item.Count)
	}
}
