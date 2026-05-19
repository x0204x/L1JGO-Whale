package system

// MISS-P1-005：L1ItemPower 強化加成整合測試。
//
// 驗證 applyItemPowerBonuses 將 ItemPowerTable 規則正確套用到 world.EquipStats：
//   - 各 stat dimension 路由到正確欄位（MR→MDef、MPR→AddMPR、…）
//   - 多 stat 同 item 同時生效
//   - nil itemPowers / 未登錄 item / enchant<min 都不應變更 stats

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func loadItemPowerTableForTest(t *testing.T, content string) *data.ItemPowerTable {
	t.Helper()
	path := filepath.Join(t.TempDir(), "item_power.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("寫 YAML: %v", err)
	}
	tbl, err := data.LoadItemPowerTable(path)
	if err != nil {
		t.Fatalf("載入 YAML: %v", err)
	}
	return tbl
}

// TestApplyItemPowerBonusesLinearMR 抗魔法頭盔 +6 應加 MDef +6。
func TestApplyItemPowerBonusesLinearMR(t *testing.T) {
	tbl := loadItemPowerTableForTest(t, `
rules:
  - { item_id: 20011, stat: MR, per_enchant: 1 }
`)
	var stats world.EquipStats
	applyItemPowerBonuses(&stats, tbl, 20011, 6)
	if stats.MDef != 6 {
		t.Fatalf("抗魔法頭盔 +6 應 MDef=6，實際 %d", stats.MDef)
	}
}

// TestApplyItemPowerBonusesAllStatDimensions 6 個 stat 維度各路由到正確欄位。
func TestApplyItemPowerBonusesAllStatDimensions(t *testing.T) {
	tbl := loadItemPowerTableForTest(t, `
rules:
  - { item_id: 1, stat: MR,         per_enchant: 2 }
  - { item_id: 2, stat: MPR,        per_enchant: 3 }
  - { item_id: 3, stat: SP,         per_enchant: 4 }
  - { item_id: 4, stat: HIT,        per_enchant: 5 }
  - { item_id: 5, stat: HP,         per_enchant: 10 }
  - { item_id: 6, stat: DMG_REDUCE, per_enchant: 1 }
`)
	cases := []struct {
		itemID  int32
		check   func(s *world.EquipStats) int
		want    int
		enchant int
		name    string
	}{
		{1, func(s *world.EquipStats) int { return s.MDef }, 6, 3, "MR→MDef"},
		{2, func(s *world.EquipStats) int { return s.AddMPR }, 9, 3, "MPR→AddMPR"},
		{3, func(s *world.EquipStats) int { return s.AddSP }, 12, 3, "SP→AddSP"},
		{4, func(s *world.EquipStats) int { return s.HitMod }, 15, 3, "HIT→HitMod"},
		{5, func(s *world.EquipStats) int { return s.AddHP }, 30, 3, "HP→AddHP"},
		{6, func(s *world.EquipStats) int { return s.DmgReduction }, 3, 3, "DMG_REDUCE→DmgReduction"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stats world.EquipStats
			applyItemPowerBonuses(&stats, tbl, c.itemID, c.enchant)
			if got := c.check(&stats); got != c.want {
				t.Fatalf("%s 應為 %d，實際 %d", c.name, c.want, got)
			}
		})
	}
}

// TestApplyItemPowerBonusesMultiStatSameItem 同 item 多個 stat 規則同時生效（如馬昆斯斗篷 MR+SP）。
func TestApplyItemPowerBonusesMultiStatSameItem(t *testing.T) {
	tbl := loadItemPowerTableForTest(t, `
rules:
  - { item_id: 21221, stat: MR, per_enchant: 3 }
  - { item_id: 21221, stat: SP, per_enchant: 1, min_enchant: 7, enchant_offset: 6 }
`)
	var stats world.EquipStats
	applyItemPowerBonuses(&stats, tbl, 21221, 8) // +8 馬昆斯斗篷
	if stats.MDef != 24 {                        // 3 * 8
		t.Fatalf("馬昆斯 +8 MR 應 MDef=24，實際 %d", stats.MDef)
	}
	if stats.AddSP != 2 { // (8 - 6) * 1
		t.Fatalf("馬昆斯 +8 SP 應 AddSP=2，實際 %d", stats.AddSP)
	}
}

// TestApplyItemPowerBonusesThresholdBlocked enchant 低於 min_enchant 不應加分。
func TestApplyItemPowerBonusesThresholdBlocked(t *testing.T) {
	tbl := loadItemPowerTableForTest(t, `
rules:
  - { item_id: 20107, stat: SP, per_enchant: 1, min_enchant: 3, enchant_offset: 2 }
`)
	var stats world.EquipStats
	applyItemPowerBonuses(&stats, tbl, 20107, 2) // 巫妖斗篷 +2 不觸發
	if stats.AddSP != 0 {
		t.Fatalf("巫妖斗篷 +2 不應加 SP，實際 %d", stats.AddSP)
	}
	applyItemPowerBonuses(&stats, tbl, 20107, 3) // 巫妖斗篷 +3 → +1
	if stats.AddSP != 1 {
		t.Fatalf("巫妖斗篷 +3 應 AddSP=1，實際 %d", stats.AddSP)
	}
}

// TestApplyItemPowerBonusesNilSafe nil table / 未登錄 item / nil stats 都不應 panic。
func TestApplyItemPowerBonusesNilSafe(t *testing.T) {
	var stats world.EquipStats
	applyItemPowerBonuses(&stats, nil, 20011, 6) // nil table
	if stats.MDef != 0 {
		t.Fatalf("nil table 不應變更 stats，實際 MDef=%d", stats.MDef)
	}

	applyItemPowerBonuses(nil, nil, 20011, 6) // nil stats — 不應 panic

	tbl := loadItemPowerTableForTest(t, "rules: []\n")
	applyItemPowerBonuses(&stats, tbl, 99999, 6) // 未登錄 item
	if stats.MDef != 0 {
		t.Fatalf("未登錄 item 不應變更 stats，實際 MDef=%d", stats.MDef)
	}
}
