package system

import (
	"github.com/l1jgo/server/internal/net"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/persist"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type innKeyItemCreateStub struct {
	itemID int32
	count  int32
	calls  int
	item   *world.InvItem
}

func (s *innKeyItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.calls++
	s.item = &world.InvItem{ObjectID: 777, ItemID: itemID, Name: "旅館鑰匙", Count: count}
	return s.item, true
}

func TestInnReturnRoomUsesItemCreateForRefundAdena(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "inn-refund-test",
	})
	key := player.Inv.AddItem(innKeyItemID, 1, "旅館鑰匙", 11, 10, false, 1)
	key.InnNpcID = 70012
	itemCreate := &shopItemCreateStub{}
	sys := NewInnSystem(&handler.Deps{
		World: ws,
		InnRooms: map[int32]map[int32]*persist.InnRoom{
			70012: {},
		},
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.ReturnRoom(player.Session, player, 90001, 70012)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 20 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=20", itemCreate.itemID, itemCreate.count, world.AdenaItemID)
	}
}

func TestInnRentRoomUsesItemCreateForKeyAndKeepsInnFields(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:         1,
		Session:           newSkillTestSession(t, 1),
		CharID:            1001,
		Name:              "inn-rent-test",
		PendingInnRoomNum: 1,
		PendingInnHall:    true,
	})
	player.Inv.AddItem(world.AdenaItemID, 1000, "金幣", 5, 0, true, 1)
	itemCreate := &innKeyItemCreateStub{}
	deps := &handler.Deps{
		World: ws,
		Items: testItemTable(t),
		InnRooms: map[int32]map[int32]*persist.InnRoom{
			70012: {
				1: {NpcID: 70012, RoomNumber: 1, DueTime: time.Now().Add(-time.Minute)},
			},
		},
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	}
	sys := NewInnSystem(deps)

	sys.RentRoom(player.Session, player, 90001, 70012, 2)

	key := itemCreate.item
	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != innKeyItemID || itemCreate.count != 2 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=2", itemCreate.itemID, itemCreate.count, innKeyItemID)
	}
	if key == nil {
		t.Fatal("租房後應取得旅館鑰匙")
	}
	if key.Count != 2 {
		t.Fatalf("旅館鑰匙應維持單一物件與 count=2，got count=%d size=%d", key.Count, player.Inv.Size())
	}
	if key.InnKeyID != key.ObjectID || key.InnNpcID != 70012 || !key.InnHall || key.InnDueTime == 0 {
		t.Fatalf("旅館鑰匙欄位錯誤：key=%+v", key)
	}
	if room := deps.InnRooms[70012][1]; room.KeyID != key.ObjectID || room.LodgerID != player.CharID || !room.Hall {
		t.Fatalf("房間狀態未對應鑰匙：room=%+v key=%+v", room, key)
	}
}
