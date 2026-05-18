package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestNpcUpgradeSuccessUsesItemCreate(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "upgrade-test",
	})
	player.Inv.AddItem(1001, 1, "測試材料", 101, 10, true, 1)
	itemCreate := &shopItemCreateStub{}
	sys := NewNpcServiceSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	upg := &data.ItemUpgrade{
		UpgradeChance: 100,
		NewItemID:     1002,
		MainItemID:    1001,
		MainItemCount: 1,
	}

	sys.NpcUpgrade(player.Session, player, upg)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1002 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1002 count=1", itemCreate.itemID, itemCreate.count)
	}
}

func TestNpcRefineUsesItemCreateForCrystals(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "refine-test",
	})
	item := player.Inv.AddItem(1002, 1, "測試劍", 102, 10, false, 1)
	itemCreate := &shopItemCreateStub{}
	sys := NewNpcServiceSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.Refine(player.Session, player, item, 1001, 7)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1001 || itemCreate.count != 7 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1001 count=7", itemCreate.itemID, itemCreate.count)
	}
}
