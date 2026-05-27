package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDurabilitySchemaLoadsNpcHardFlag(t *testing.T) {
	path := writeTempYAML(t, "npc_list.yaml", `
npcs:
  - npc_id: 45001
    name: hard monster
    impl: L1Monster
    hp: 10
    hard: true
`)

	table, err := LoadNpcTable(path)
	if err != nil {
		t.Fatalf("載入 NPC YAML 失敗: %v", err)
	}

	tmpl := table.Get(45001)
	if tmpl == nil {
		t.Fatal("找不到測試 NPC")
	}
	if !tmpl.Hard {
		t.Fatal("hard 欄位應載入為 true")
	}
}

func TestNpcSchemaLoadsSubMagicSpeedLikeJava(t *testing.T) {
	path := writeTempYAML(t, "npc_list.yaml", `
npcs:
  - npc_id: 990002
    name: shock stun caster
    impl: L1Monster
    hp: 10
    atk_magic_speed: 1200
    sub_magic_speed: 1600
`)

	table, err := LoadNpcTable(path)
	if err != nil {
		t.Fatalf("載入 NPC YAML 失敗: %v", err)
	}

	tmpl := table.Get(990002)
	if tmpl == nil {
		t.Fatal("找不到測試 NPC")
	}
	if tmpl.SubMagicSpeed != 1600 {
		t.Fatalf("Java NpcTable 會載入 sub_magic_speed，SubMagicSpeed=%d want=1600", tmpl.SubMagicSpeed)
	}
	if tmpl.AtkMagicSpeed != 1200 {
		t.Fatalf("Java NpcTable 會載入 atk_magic_speed，AtkMagicSpeed=%d want=1200", tmpl.AtkMagicSpeed)
	}
}

func TestNpcYamlIncludesAtkMagicSpeedLikeJava(t *testing.T) {
	table, err := LoadNpcTable(filepath.Join("..", "..", "data", "yaml", "npc_list.yaml"))
	if err != nil {
		t.Fatalf("載入 NPC YAML 失敗: %v", err)
	}

	tmpl := table.Get(45005)
	if tmpl == nil {
		t.Fatal("找不到 yiwei 測試 NPC 45005")
	}
	if tmpl.AtkMagicSpeed != 1000 {
		t.Fatalf("yiwei npc 45005 應保留 atk_magic_speed，AtkMagicSpeed=%d want=1000", tmpl.AtkMagicSpeed)
	}
	if tmpl.SubMagicSpeed != 1000 {
		t.Fatalf("yiwei npc 45005 應保留 sub_magic_speed，SubMagicSpeed=%d want=1000", tmpl.SubMagicSpeed)
	}
}

func TestDurabilitySchemaLoadsWeaponCanBeDamagedFlag(t *testing.T) {
	weaponPath := writeTempYAML(t, "weapon_list.yaml", `
weapons:
  - item_id: 1
    name: fragile sword
    type: sword
    can_be_damaged: true
  - item_id: 2
    name: protected sword
    type: sword
    can_be_damaged: false
  - item_id: 3
    name: legacy sword
    type: sword
`)
	armorPath := writeTempYAML(t, "armor_list.yaml", "armors: []\n")
	etcPath := writeTempYAML(t, "etcitem_list.yaml", "items: []\n")

	table, err := LoadItemTable(weaponPath, armorPath, etcPath)
	if err != nil {
		t.Fatalf("載入物品 YAML 失敗: %v", err)
	}

	if !table.Get(1).CanBeDamaged {
		t.Fatal("can_be_damaged=true 應載入為可損壞")
	}
	if table.Get(2).CanBeDamaged {
		t.Fatal("can_be_damaged=false 應載入為不可損壞")
	}
	if table.Get(3).CanBeDamaged {
		t.Fatal("未宣告 can_be_damaged 的資料應對齊 Java 預設為不可損壞")
	}
}

func writeTempYAML(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}
	return path
}
