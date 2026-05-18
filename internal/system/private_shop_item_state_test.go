package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestPrivateShopTransferItemPreservesItemState(t *testing.T) {
	seller := newTradeTestPlayer(t, 1, 100, "賣家")
	buyer := newTradeTestPlayer(t, 2, 200, "買家")
	item := addTradeStateItem(seller)
	objectID := item.ObjectID
	sys := NewPrivateShopSystem(&handler.Deps{
		Items: testItemTable(t),
		Log:   zap.NewNop(),
	})

	sys.TransferItem(seller, buyer, item, 1)

	if seller.Inv.FindByObjectID(objectID) != nil {
		t.Fatal("個人商店轉移後來源背包不應保留原物品")
	}
	assertTradeItemStatePreserved(t, buyer.Inv.FindByObjectID(objectID), objectID)
}

func TestBuildPrivateShopWALEntriesUsesObjectIDForNonStackableItem(t *testing.T) {
	seller := newTradeTestPlayer(t, 1, 100, "賣家")
	buyer := newTradeTestPlayer(t, 2, 200, "買家")
	item := addTradeStateItem(seller)

	entries := buildPrivateShopWALEntries(seller, buyer, buyer, seller, item, 1, 300)
	if len(entries) != 2 {
		t.Fatalf("個人商店應產生物品與金幣兩筆 WAL，got %d", len(entries))
	}
	if entries[0].TxType != "private_shop" || entries[0].FromChar != seller.CharID || entries[0].ToChar != buyer.CharID {
		t.Fatalf("物品 WAL 來源/目標錯誤: %+v", entries[0])
	}
	if entries[0].ItemID != item.ObjectID || entries[0].Count != 1 {
		t.Fatalf("非堆疊物品 WAL 應記錄 ObjectID 與數量，entry=%+v objectID=%d", entries[0], item.ObjectID)
	}
	if entries[1].FromChar != buyer.CharID || entries[1].ToChar != seller.CharID ||
		entries[1].ItemID != world.AdenaItemID || entries[1].GoldAmount != 300 {
		t.Fatalf("金幣 WAL 錯誤: %+v", entries[1])
	}
}

func TestBuildPrivateShopWALEntriesUsesItemIDForStackableItem(t *testing.T) {
	seller := newTradeTestPlayer(t, 1, 100, "賣家")
	buyer := newTradeTestPlayer(t, 2, 200, "買家")
	item := seller.Inv.AddItem(1001, 10, "測試藥水", 101, 10, true, 1)

	entries := buildPrivateShopWALEntries(seller, buyer, buyer, seller, item, 3, 90)
	if len(entries) != 2 {
		t.Fatalf("個人商店應產生物品與金幣兩筆 WAL，got %d", len(entries))
	}
	if entries[0].ItemID != item.ItemID || entries[0].Count != 3 {
		t.Fatalf("堆疊物品 WAL 應記錄模板 item_id 與轉移數量，entry=%+v itemID=%d", entries[0], item.ItemID)
	}
}
