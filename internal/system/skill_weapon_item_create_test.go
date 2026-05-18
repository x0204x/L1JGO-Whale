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

func testBringStoneItemTable(t *testing.T) *data.ItemTable {
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
  - item_id: 40321
    name: 高品質的黑魔石
    use_type: other
    weight: 10
    inv_gfx: 40321
    grd_gfx: 40321
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

func TestBringStoneSuccessUsesItemCreate(t *testing.T) {
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	itemCreate := &shopItemCreateStub{}
	sys := &SkillSystem{deps: &handler.Deps{
		World:      ws,
		Items:      testBringStoneItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	}}
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID: 1,
		Name:   "bring-stone-test",
		Level:  200,
		Wis:    50,
		Str:    50,
		Con:    50,
		X:      100,
		Y:      100,
		MapID:  4,
		Inv:    world.NewInventory(),
	})
	stone := player.Inv.AddItemWithID(7001, 40320, 1, "黑魔石", 40320, 10, true, 1)

	sys.executeBringStone(sess, player, &data.SkillInfo{}, stone.ObjectID)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤: got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 40321 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤: got item=%d count=%d want item=40321 count=1", itemCreate.itemID, itemCreate.count)
	}
	if player.Inv.FindByObjectID(stone.ObjectID) != nil {
		t.Fatal("提煉成功後來源魔石未被消耗")
	}
}
