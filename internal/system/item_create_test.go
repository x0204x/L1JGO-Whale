package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func testItemTable(t *testing.T) *data.ItemTable {
	t.Helper()
	dir := t.TempDir()
	weaponPath := filepath.Join(dir, "weapon.yaml")
	armorPath := filepath.Join(dir, "armor.yaml")
	etcPath := filepath.Join(dir, "etc.yaml")

	if err := os.WriteFile(weaponPath, []byte("weapons: []\n"), 0644); err != nil {
		t.Fatalf("寫入武器測試資料失敗: %v", err)
	}
	if err := os.WriteFile(armorPath, []byte("armors: []\n"), 0644); err != nil {
		t.Fatalf("寫入防具測試資料失敗: %v", err)
	}
	etcYAML := `items:
  - item_id: 40308
    name: 金幣
    use_type: other
    weight: 0
    inv_gfx: 5
    grd_gfx: 5
    stackable: true
    bless: 1
    tradeable: true
  - item_id: 40312
    name: 旅館鑰匙
    use_type: other
    weight: 10
    inv_gfx: 11
    grd_gfx: 11
    stackable: false
    bless: 1
    tradeable: true
  - item_id: 1001
    name: 測試藥水
    use_type: potion
    weight: 10
    inv_gfx: 101
    grd_gfx: 201
    stackable: true
    bless: 1
    tradeable: true
  - item_id: 1002
    name: 測試劍
    use_type: weapon
    weight: 10
    inv_gfx: 102
    grd_gfx: 202
    stackable: false
    bless: 0
    tradeable: true
  - item_id: 1003
    name: 測試魔杖
    use_type: wand
    weight: 100
    inv_gfx: 103
    grd_gfx: 203
    stackable: false
    max_charge_count: 15
    bless: 0
    tradeable: true
  - item_id: 6100
    name: 測試商城幣
    use_type: other
    weight: 0
    inv_gfx: 610
    grd_gfx: 610
    stackable: true
    bless: 1
    tradeable: true
`
	if err := os.WriteFile(etcPath, []byte(etcYAML), 0644); err != nil {
		t.Fatalf("寫入道具測試資料失敗: %v", err)
	}

	table, err := data.LoadItemTable(weaponPath, armorPath, etcPath)
	if err != nil {
		t.Fatalf("載入道具測試資料失敗: %v", err)
	}
	return table
}

func testItemCreateSystem(t *testing.T) (*ItemCreateSystem, *world.PlayerInfo) {
	t.Helper()
	player := &world.PlayerInfo{
		Name: "測試角色",
		Str:  10,
		Con:  10,
		Inv:  world.NewInventory(),
	}
	deps := &handler.Deps{
		Items: testItemTable(t),
		Log:   zap.NewNop(),
	}
	return NewItemCreateSystem(deps), player
}

func TestItemCreateSystemStacksExistingItem(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	first, ok := sys.GiveItem(nil, player, 1001, 3)
	if !ok {
		t.Fatal("第一次給可堆疊物品應成功")
	}

	second, ok := sys.GiveItem(nil, player, 1001, 2)
	if !ok {
		t.Fatal("第二次給可堆疊物品應成功")
	}

	if first.ObjectID != second.ObjectID {
		t.Fatalf("可堆疊物品應合併到同一格，got first=%d second=%d", first.ObjectID, second.ObjectID)
	}
	if second.Count != 5 {
		t.Fatalf("合併後數量錯誤，got %d want 5", second.Count)
	}
	if player.Inv.Size() != 1 {
		t.Fatalf("可堆疊物品應只占 1 格，got %d", player.Inv.Size())
	}
}

func TestItemCreateSystemCreatesSeparateSlotsForNonStackableItems(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	item, ok := sys.GiveItem(nil, player, 1002, 2)
	if !ok {
		t.Fatal("給不可堆疊物品應成功")
	}

	if item.ItemID != 1002 {
		t.Fatalf("回傳最後建立物品錯誤，got %d want 1002", item.ItemID)
	}
	if player.Inv.Size() != 2 {
		t.Fatalf("不可堆疊物品數量 2 應占 2 格，got %d", player.Inv.Size())
	}
	if player.Inv.Items[0].ObjectID == player.Inv.Items[1].ObjectID {
		t.Fatal("不可堆疊物品應有不同 ObjectID")
	}
}

func TestItemCreateSystemInitializesChargeCount(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	item, ok := sys.GiveItem(nil, player, 1003, 1)
	if !ok {
		t.Fatal("給有 charge 的物品應成功")
	}
	if item.ChargeCount != 15 {
		t.Fatalf("ChargeCount 錯誤，got %d want 15", item.ChargeCount)
	}
}

func TestItemCreateSystemAppliesOptions(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	item, ok := sys.GiveItemWithOptions(nil, player, 1002, 1, ItemCreateOptions{
		EnchantLvl: 7,
		Bless:      2,
		BlessSet:   true,
	})
	if !ok {
		t.Fatal("給有覆寫欄位的物品應成功")
	}
	if item.EnchantLvl != 7 || item.Bless != 2 {
		t.Fatalf("覆寫欄位錯誤，got enchant=%d bless=%d want enchant=7 bless=2", item.EnchantLvl, item.Bless)
	}
}

func TestItemCreateSystemSingleItemOptionKeepsCount(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	item, ok := sys.GiveItemWithOptions(nil, player, innKeyItemID, 2, ItemCreateOptions{
		SingleItem: true,
		BeforeSend: func(item *world.InvItem) {
			item.InnKeyID = item.ObjectID
			item.InnNpcID = 70012
		},
	})
	if !ok {
		t.Fatal("單一物件給物品應成功")
	}
	if player.Inv.Size() != 1 || item.Count != 2 {
		t.Fatalf("單一物件應只占 1 格且保留 count=2，got size=%d count=%d", player.Inv.Size(), item.Count)
	}
	if item.InnKeyID != item.ObjectID || item.InnNpcID != 70012 {
		t.Fatalf("BeforeSend 應在回傳前套用欄位，got key=%+v", item)
	}
}

func TestItemCreateSystemRejectsUnknownItem(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	_, ok := sys.GiveItem(nil, player, 9999, 1)
	if ok {
		t.Fatal("未知 itemID 不應給物品成功")
	}
	if player.Inv.Size() != 0 {
		t.Fatalf("未知 itemID 不應改變背包，got size=%d", player.Inv.Size())
	}
}

func TestItemCreateSystemRejectsFullInventoryForNewSlot(t *testing.T) {
	sys, player := testItemCreateSystem(t)
	for i := 0; i < world.MaxInventorySize; i++ {
		player.Inv.Items = append(player.Inv.Items, &world.InvItem{
			ObjectID: int32(10_000 + i),
			ItemID:   int32(20_000 + i),
			Count:    1,
		})
	}

	_, ok := sys.GiveItem(nil, player, 1002, 1)
	if ok {
		t.Fatal("背包滿時新增不可堆疊物品不應成功")
	}
	if player.Inv.Size() != world.MaxInventorySize {
		t.Fatalf("背包滿拒絕後不應改變格數，got %d", player.Inv.Size())
	}
}
