package data

// MISS-P1-005：物品強化加成載入器測試。
//
// 涵蓋：YAML 解析、Bonus() 計算（線性 + 門檻線性 + offset）、驗證失敗情境。

import (
	"os"
	"path/filepath"
	"testing"
)

func writeItemPowerYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "item_power.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("寫 YAML: %v", err)
	}
	return path
}

// TestItemPowerBonusLinear 純線性物品（如抗魔法頭盔）：bonus = enchant × per_enchant。
func TestItemPowerBonusLinear(t *testing.T) {
	r := &ItemPowerRule{ItemID: 20011, Stat: ItemPowerStatMR, PerEnchant: 1}
	cases := []struct {
		enchant int
		want    int
	}{
		{0, 0},
		{1, 1},
		{6, 6},
		{12, 12},
		{-3, 0}, // 負值不應加分
	}
	for _, c := range cases {
		if got := r.Bonus(c.enchant); got != c.want {
			t.Errorf("線性 enchant=%d 應為 %d，實際 %d", c.enchant, c.want, got)
		}
	}
}

// TestItemPowerBonusThresholdWithOffset 門檻+offset（如巫妖斗篷）：
//
//	min_enchant=3, enchant_offset=2 → +3 給 1、+4 給 2、+5 給 3。
//	+1/+2 不觸發。
func TestItemPowerBonusThresholdWithOffset(t *testing.T) {
	r := &ItemPowerRule{
		ItemID:        20107,
		Stat:          ItemPowerStatSP,
		PerEnchant:    1,
		MinEnchant:    3,
		EnchantOffset: 2,
	}
	cases := []struct {
		enchant int
		want    int
	}{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 1},
		{5, 3},
		{8, 6},
	}
	for _, c := range cases {
		if got := r.Bonus(c.enchant); got != c.want {
			t.Errorf("巫妖斗篷 enchant=%d 應為 %d，實際 %d", c.enchant, c.want, got)
		}
	}
}

// TestItemPowerBonusMultiplierGreaterThanOne 高 per_enchant（如混沌斗篷 MR×3）。
func TestItemPowerBonusMultiplierGreaterThanOne(t *testing.T) {
	r := &ItemPowerRule{ItemID: 20078, Stat: ItemPowerStatMR, PerEnchant: 3}
	if got := r.Bonus(6); got != 18 {
		t.Errorf("混沌斗篷 enchant=6 應為 18，實際 %d", got)
	}
	if got := r.Bonus(0); got != 0 {
		t.Errorf("enchant=0 應為 0，實際 %d", got)
	}
}

// TestItemPowerNilSafe nil rule 不應 panic。
func TestItemPowerNilSafe(t *testing.T) {
	var r *ItemPowerRule
	if got := r.Bonus(5); got != 0 {
		t.Errorf("nil rule 應回 0，實際 %d", got)
	}
}

// TestLoadItemPowerValidYAML 載入合法 YAML 並驗證 Get + Count + ItemCount。
func TestLoadItemPowerValidYAML(t *testing.T) {
	path := writeItemPowerYAML(t, `
rules:
  - { item_id: 20011, stat: MR, per_enchant: 1, note: "抗魔法頭盔" }
  - { item_id: 20078, stat: MR, per_enchant: 3, note: "混沌斗篷" }
  - { item_id: 20107, stat: SP, per_enchant: 1, min_enchant: 3, enchant_offset: 2, note: "巫妖斗篷" }
`)
	tbl, err := LoadItemPowerTable(path)
	if err != nil {
		t.Fatalf("LoadItemPowerTable: %v", err)
	}
	if tbl.Count() != 3 {
		t.Fatalf("Count 應為 3，實際 %d", tbl.Count())
	}
	if tbl.ItemCount() != 3 {
		t.Fatalf("ItemCount 應為 3，實際 %d", tbl.ItemCount())
	}
	rules := tbl.Get(20011)
	if len(rules) != 1 || rules[0].Stat != ItemPowerStatMR {
		t.Fatalf("抗魔法頭盔規則異常: %+v", rules)
	}
	if rules[0].Bonus(6) != 6 {
		t.Fatalf("抗魔法頭盔 +6 應為 6，實際 %d", rules[0].Bonus(6))
	}
	if got := tbl.Get(20107)[0].Bonus(3); got != 1 {
		t.Fatalf("巫妖斗篷 +3 應為 1，實際 %d", got)
	}
}

// TestLoadItemPowerMultipleRulesPerItem 一個 item_id 可有多個 stat 規則。
func TestLoadItemPowerMultipleRulesPerItem(t *testing.T) {
	// 馬昆斯斗篷 同時有 MR 與 SP 加成
	path := writeItemPowerYAML(t, `
rules:
  - { item_id: 21221, stat: MR, per_enchant: 3, note: "馬昆斯斗篷 MR" }
  - { item_id: 21221, stat: SP, per_enchant: 1, min_enchant: 7, enchant_offset: 6, note: "馬昆斯斗篷 SP" }
`)
	tbl, err := LoadItemPowerTable(path)
	if err != nil {
		t.Fatalf("LoadItemPowerTable: %v", err)
	}
	if tbl.ItemCount() != 1 {
		t.Fatalf("ItemCount 應為 1（同 item_id 多規則），實際 %d", tbl.ItemCount())
	}
	if tbl.Count() != 2 {
		t.Fatalf("Count 應為 2（規則總數），實際 %d", tbl.Count())
	}
	rules := tbl.Get(21221)
	if len(rules) != 2 {
		t.Fatalf("馬昆斯斗篷應有 2 條規則，實際 %d", len(rules))
	}
}

// TestLoadItemPowerRejectsInvalid 無效規則應被 validateItemPowerRule 拒絕。
func TestLoadItemPowerRejectsInvalid(t *testing.T) {
	cases := []struct {
		name string
		yaml string
	}{
		{
			name: "item_id 為 0",
			yaml: "rules:\n  - { item_id: 0, stat: MR, per_enchant: 1 }\n",
		},
		{
			name: "未支援的 stat",
			yaml: "rules:\n  - { item_id: 100, stat: BOGUS, per_enchant: 1 }\n",
		},
		{
			name: "per_enchant 為 0",
			yaml: "rules:\n  - { item_id: 100, stat: MR, per_enchant: 0 }\n",
		},
		{
			name: "min_enchant 為負",
			yaml: "rules:\n  - { item_id: 100, stat: MR, per_enchant: 1, min_enchant: -1 }\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := writeItemPowerYAML(t, c.yaml)
			if _, err := LoadItemPowerTable(path); err == nil {
				t.Fatalf("應拒絕 %s，實際成功載入", c.name)
			}
		})
	}
}

// TestLoadItemPowerMissingNoOp Get 未登錄 item 應回 nil，nil table 不應 panic。
func TestLoadItemPowerMissingNoOp(t *testing.T) {
	path := writeItemPowerYAML(t, "rules: []\n")
	tbl, err := LoadItemPowerTable(path)
	if err != nil {
		t.Fatalf("empty rules 應可載入: %v", err)
	}
	if tbl.Get(99999) != nil {
		t.Fatal("未登錄 item 應回 nil")
	}

	var nilTbl *ItemPowerTable
	if nilTbl.Get(20011) != nil {
		t.Fatal("nil table Get 應回 nil")
	}
	if nilTbl.Count() != 0 || nilTbl.ItemCount() != 0 {
		t.Fatal("nil table Count/ItemCount 應為 0")
	}
}
