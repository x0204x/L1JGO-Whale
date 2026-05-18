package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestWarehouseItemStateRoundTripPreservesInvItemState(t *testing.T) {
	player := newTradeTestPlayer(t, 1, 100, "倉庫玩家")
	item := addTradeStateItem(player)
	objectID := item.ObjectID

	whItem := warehouseItemFromInvItem("test_account", player.Name, handler.WhTypePersonal, item, 1)
	if whItem.ItemObjID != objectID {
		t.Fatalf("倉庫存入快照應保留原 ObjectID，got %d want %d", whItem.ItemObjID, objectID)
	}

	itemTable := testItemTable(t)
	cache := warehouseCacheFromPersistItem(whItem, itemTable.Get(item.ItemID))
	if cache.TempObjID != objectID {
		t.Fatalf("倉庫快取應沿用原 ObjectID，got %d want %d", cache.TempObjID, objectID)
	}

	inv := world.NewInventory()
	restored := inv.AddItemWithID(cache.TempObjID, cache.ItemID, cache.Count, cache.Name, cache.InvGfx, cache.Weight, cache.Stackable, byte(cache.Bless))
	copyWarehouseCacheState(restored, cache)

	assertTradeItemStatePreserved(t, restored, objectID)
}

func TestWarehousePartialStackDepositUsesSeparateWarehouseObjectID(t *testing.T) {
	player := newTradeTestPlayer(t, 1, 100, "倉庫玩家")
	item := player.Inv.AddItem(1001, 10, "測試藥水", 101, 10, true, 1)

	whItem := warehouseItemFromInvItem("test_account", player.Name, handler.WhTypePersonal, item, 3)
	if whItem.ItemObjID == 0 {
		t.Fatal("部分堆疊存入倉庫應建立可追蹤的新 ObjectID")
	}
	if whItem.ItemObjID == item.ObjectID {
		t.Fatalf("部分堆疊存入倉庫不可重用來源背包 ObjectID，got %d", whItem.ItemObjID)
	}
}
