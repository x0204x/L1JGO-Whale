package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type petItemCreateStub struct {
	itemID int32
	count  int32
	calls  int
	nextID int32
}

func (s *petItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.calls++
	if s.nextID == 0 {
		s.nextID = 9001
	}
	return &world.InvItem{ObjectID: s.nextID, ItemID: itemID, Count: count}, true
}

func TestPetCollarCreationUsesItemCreate(t *testing.T) {
	for _, itemID := range []int32{petCollarNormal, petCollarHigher} {
		ws := world.NewState()
		sess := newSkillTestSession(t, uint64(itemID))
		itemCreate := &petItemCreateStub{}
		sys := NewPetSystem(&handler.Deps{
			World:      ws,
			ItemCreate: itemCreate,
			Log:        zap.NewNop(),
		})
		player := addSkillTestPlayer(ws, &world.PlayerInfo{
			SessionID: uint64(itemID),
			Session:   sess,
			CharID:    1001,
			Name:      "pet-test",
			Inv:       world.NewInventory(),
		})

		collar := sys.givePetCollarItem(sess, player, itemID)

		if itemCreate.calls != 1 {
			t.Fatalf("itemID=%d ItemCreate 呼叫次數錯誤: got %d want 1", itemID, itemCreate.calls)
		}
		if itemCreate.itemID != itemID || itemCreate.count != 1 {
			t.Fatalf("ItemCreate 參數錯誤: got item=%d count=%d want item=%d count=1",
				itemCreate.itemID, itemCreate.count, itemID)
		}
		if collar == nil || collar.ObjectID != 9001 {
			t.Fatalf("項圈物件錯誤: %+v", collar)
		}
	}
}
