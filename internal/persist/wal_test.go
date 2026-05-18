package persist

import (
	"strings"
	"testing"
)

func TestWALReplayClassifiesGoldAndItemEntries(t *testing.T) {
	itemEntry := WALEntry{FromChar: 100, ToChar: 200, ItemID: 500000001, GoldAmount: 0}
	if !isWALItemTransfer(itemEntry) {
		t.Fatal("一般物品 WAL 應被視為物品轉移")
	}

	stackEntry := WALEntry{FromChar: 100, ToChar: 200, ItemID: 1001, Count: 3, GoldAmount: 0}
	if isWALItemTransfer(stackEntry) {
		t.Fatal("堆疊物品 WAL 不應被 ObjectID 轉移 replay 處理")
	}
	if !isWALStackItemTransfer(stackEntry) {
		t.Fatal("堆疊物品 WAL 應被數量扣加 replay 處理")
	}

	goldEntry := WALEntry{FromChar: 100, ToChar: 200, ItemID: defaultWALGoldItemID, GoldAmount: 30}
	if isWALItemTransfer(goldEntry) {
		t.Fatal("金幣 WAL 不應再被物品轉移 replay 處理")
	}
	if walGoldItemID(goldEntry) != defaultWALGoldItemID {
		t.Fatalf("金幣 WAL item_id 錯誤，got %d", walGoldItemID(goldEntry))
	}
}

func TestRecoverWALItemTransferSQLUsesPersistedObjectID(t *testing.T) {
	if !strings.Contains(recoverWALItemTransferSQL, "obj_id = $2") {
		t.Fatalf("WAL 物品 replay 必須使用 character_items.obj_id: %s", recoverWALItemTransferSQL)
	}
	if strings.Contains(recoverWALItemTransferSQL, "WHERE id = $2") {
		t.Fatalf("WAL 物品 replay 不可使用 character_items.id: %s", recoverWALItemTransferSQL)
	}
}
