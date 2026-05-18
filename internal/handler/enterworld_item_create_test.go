package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestLoadInventoryFromDBGivesStartingGoldWithItemCreate(t *testing.T) {
	player := &world.PlayerInfo{
		CharID: 1001,
		Name:   "new-player",
		Inv:    world.NewInventory(),
	}
	itemCreate := &cookingItemCreateStub{}
	deps := &Deps{
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	}

	loadInventoryFromDB(player, deps)

	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 20000 {
		t.Fatalf("初始金幣應走 ItemCreate: item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
	if player.Inv.Size() != 0 {
		t.Fatalf("ItemCreate 成功時不應再走 handler 直接 AddItem fallback: size=%d", player.Inv.Size())
	}
}
