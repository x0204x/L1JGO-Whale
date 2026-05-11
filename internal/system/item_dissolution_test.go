package system

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newDissolutionTestSystem(t *testing.T, ws *world.State) *ItemUseSystem {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	return NewItemUseSystem(&handler.Deps{
		World:      ws,
		Items:      items,
		Resolvents: loadDissolutionTestResolvents(t, map[int32]int32{31: 60, 40010: 1}),
		Log:        zap.NewNop(),
	})
}

func TestItemDissolutionSuccessConsumesTargetAndSolventAndGivesCrystals(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
	})
	target := player.Inv.AddItemWithID(7001, 31, 1, "長劍", 0, 1000, false, 1)
	solvent := player.Inv.AddItemWithID(7002, 41245, 2, "溶解劑", 2608, 0, true, 1)
	s := newDissolutionTestSystem(t, ws)

	if !s.UseDissolutionWithRoll(player.Session, player, solvent, target.ObjectID, 75) {
		t.Fatal("可溶解物品應成功")
	}

	if player.Inv.FindByObjectID(target.ObjectID) != nil {
		t.Fatal("成功溶解後目標物品應移除")
	}
	if solvent.Count != 1 {
		t.Fatalf("成功溶解後應消耗 1 個溶解劑，got=%d want=1", solvent.Count)
	}
	crystal := player.Inv.FindByItemID(41246)
	if crystal == nil || crystal.Count != 90 {
		t.Fatalf("roll=75 應給 1.5 倍魔法結晶體 90 個，got=%v", crystal)
	}
}

func TestItemDissolutionRejectsEnchantedWeapon(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
	})
	target := player.Inv.AddItemWithID(7001, 31, 1, "長劍", 0, 1000, false, 1)
	target.EnchantLvl = 1
	solvent := player.Inv.AddItemWithID(7002, 41245, 2, "溶解劑", 2608, 0, true, 1)
	s := newDissolutionTestSystem(t, ws)

	if s.UseDissolutionWithRoll(player.Session, player, solvent, target.ObjectID, 10) {
		t.Fatal("強化過的武器不得溶解")
	}

	if player.Inv.FindByObjectID(target.ObjectID) == nil {
		t.Fatal("拒絕溶解時目標物品不應移除")
	}
	if solvent.Count != 2 {
		t.Fatalf("拒絕溶解時溶解劑不應消耗，got=%d want=2", solvent.Count)
	}
	if player.Inv.FindByItemID(41246) != nil {
		t.Fatal("拒絕溶解時不應給魔法結晶體")
	}
	if !hasServerMessage(drainSkillTestPackets(player.Session), 1161) {
		t.Fatal("強化武器不得溶解時應送 S_ServerMessage 1161")
	}
}

func TestItemDissolutionRejectsEquippedArmor(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
	})
	target := player.Inv.AddItemWithID(7001, 20043, 1, "鋼盔", 0, 1000, false, 1)
	target.Equipped = true
	solvent := player.Inv.AddItemWithID(7002, 41245, 2, "溶解劑", 2608, 0, true, 1)
	s := NewItemUseSystem(&handler.Deps{
		World: ws,
		Items: mustLoadDissolutionTestItems(t),
		Resolvents: loadDissolutionTestResolvents(t, map[int32]int32{
			20043: 4,
		}),
		Log: zap.NewNop(),
	})

	if s.UseDissolutionWithRoll(player.Session, player, solvent, target.ObjectID, 10) {
		t.Fatal("裝備中的防具不得溶解")
	}
	if target.Count != 1 || solvent.Count != 2 {
		t.Fatalf("拒絕溶解時不得消耗物品，target=%d solvent=%d", target.Count, solvent.Count)
	}
	if !hasServerMessage(drainSkillTestPackets(player.Session), 1161) {
		t.Fatal("裝備中防具不得溶解時應送 S_ServerMessage 1161")
	}
}

func mustLoadDissolutionTestItems(t *testing.T) *data.ItemTable {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	return items
}

func loadDissolutionTestResolvents(t *testing.T, entries map[int32]int32) *data.ResolventTable {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "resolvent_list.yaml")
	content := "items:\n"
	for itemID, count := range entries {
		content += "  - item_id: " + int32ToString(itemID) + "\n"
		content += "    note: 測試物品\n"
		content += "    crystal_count: " + int32ToString(count) + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入溶解測試 YAML 失敗: %v", err)
	}
	table, err := data.LoadResolventTable(path)
	if err != nil {
		t.Fatalf("載入溶解測試表失敗: %v", err)
	}
	return table
}

func int32ToString(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}
