package data

// 物品強化加成（MISS-P1-005）
//
// Java 對照：com.lineage.server.model.Instance.L1ItemPower
//
// L1ItemPower 為 50+ 個特殊物品提供「依強化等級遞增的額外屬性」，
// 例如「抗魔法頭盔」（ID 20011）每強化 +1 額外提供 +1 抗魔。
// 本檔載入這些靜態加成規則，供 system/equip.go 在計算裝備加成時疊加。
//
// 支援的加成類型（對應 L1ItemPower 的 8 個 dimension 中的子集）：
//   MR / MPR / SP / HIT / HP / DMG_REDUCE
// 暫不支援（單一/極少數物品 + 需要架構額外欄位）：
//   X2（雙擊率，1 個物品咆哮雙刀） — EquipStats 無對應欄位
//   PVPDMG（PvP 額外傷害，8 個火神武器） — EquipStats 無對應欄位且皆為 step-function
//
// 規則模型：
//   - 線性：enchant_level >= 1 時，bonus = (enchant_level - enchant_offset) * per_enchant
//   - 門檻：enchant_level < min_enchant 時 bonus = 0
//   - 預設 min_enchant=1、enchant_offset=0（純線性，最常見情況）
//
// 範例：
//   抗魔法頭盔 (20011)：per_enchant=1（強化+6 → MR+6）
//   巫妖斗篷 (20107)：per_enchant=1, min_enchant=3, enchant_offset=2
//                    （強化+3 → SP+1；+5 → SP+3）
//
// 未涵蓋的 step-function 物品（體力臂甲/法師臂甲/守護臂甲/火神武器）已知缺口，
// 待後續子任務以複雜 schema 補上。

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ItemPowerStatKind 強化加成的屬性類型。
type ItemPowerStatKind string

const (
	ItemPowerStatMR        ItemPowerStatKind = "MR"
	ItemPowerStatMPR       ItemPowerStatKind = "MPR"
	ItemPowerStatSP        ItemPowerStatKind = "SP"
	ItemPowerStatHIT       ItemPowerStatKind = "HIT"
	ItemPowerStatHP        ItemPowerStatKind = "HP"
	ItemPowerStatDmgReduce ItemPowerStatKind = "DMG_REDUCE"
)

// ItemPowerRule 單一物品的強化加成規則。
type ItemPowerRule struct {
	ItemID        int32             `yaml:"item_id"`
	Stat          ItemPowerStatKind `yaml:"stat"`
	PerEnchant    int               `yaml:"per_enchant"`              // 每強化 +1 額外加 N 點
	MinEnchant    int               `yaml:"min_enchant,omitempty"`    // 最低觸發強化等級（預設 1）
	EnchantOffset int               `yaml:"enchant_offset,omitempty"` // 計算前先扣除（預設 0）
	Note          string            `yaml:"note,omitempty"`
}

// Bonus 依當前強化等級回傳此規則的加成值。
// enchant 為負或低於 MinEnchant 時回 0。
func (r *ItemPowerRule) Bonus(enchant int) int {
	if r == nil || enchant <= 0 {
		return 0
	}
	minE := r.MinEnchant
	if minE <= 0 {
		minE = 1
	}
	if enchant < minE {
		return 0
	}
	return (enchant - r.EnchantOffset) * r.PerEnchant
}

// itemPowerFile YAML 根結構。
type itemPowerFile struct {
	Rules []ItemPowerRule `yaml:"rules"`
}

// ItemPowerTable 依 item_id 索引的強化加成規則表。
// 一個 item_id 可有多個 stat 規則（不同維度）。
type ItemPowerTable struct {
	byItem map[int32][]*ItemPowerRule
}

// LoadItemPowerTable 從 YAML 載入強化加成規則。
func LoadItemPowerTable(path string) (*ItemPowerTable, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取 item_power: %w", err)
	}
	var f itemPowerFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("解析 item_power: %w", err)
	}
	t := &ItemPowerTable{byItem: make(map[int32][]*ItemPowerRule, len(f.Rules))}
	for i := range f.Rules {
		r := &f.Rules[i]
		if err := validateItemPowerRule(r); err != nil {
			return nil, fmt.Errorf("規則 item_id=%d stat=%s 驗證失敗: %w", r.ItemID, r.Stat, err)
		}
		t.byItem[r.ItemID] = append(t.byItem[r.ItemID], r)
	}
	return t, nil
}

// validateItemPowerRule 檢查單一規則欄位合法。
func validateItemPowerRule(r *ItemPowerRule) error {
	if r.ItemID <= 0 {
		return fmt.Errorf("item_id 必須 > 0")
	}
	switch r.Stat {
	case ItemPowerStatMR, ItemPowerStatMPR, ItemPowerStatSP,
		ItemPowerStatHIT, ItemPowerStatHP, ItemPowerStatDmgReduce:
	default:
		return fmt.Errorf("未支援的 stat=%q（允許 MR/MPR/SP/HIT/HP/DMG_REDUCE）", r.Stat)
	}
	if r.PerEnchant == 0 {
		return fmt.Errorf("per_enchant 必須非零")
	}
	if r.MinEnchant < 0 || r.EnchantOffset < 0 {
		return fmt.Errorf("min_enchant 與 enchant_offset 不可為負")
	}
	return nil
}

// Get 取得指定物品 ID 的所有強化加成規則；不存在回 nil。
func (t *ItemPowerTable) Get(itemID int32) []*ItemPowerRule {
	if t == nil {
		return nil
	}
	return t.byItem[itemID]
}

// Count 規則總數（多個 stat 共用一個 item_id 也算多筆）。
func (t *ItemPowerTable) Count() int {
	if t == nil {
		return 0
	}
	n := 0
	for _, list := range t.byItem {
		n += len(list)
	}
	return n
}

// ItemCount 已登錄 item_id 數量。
func (t *ItemPowerTable) ItemCount() int {
	if t == nil {
		return 0
	}
	return len(t.byItem)
}

// String debug helper.
func (r *ItemPowerRule) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "ItemPower{id=%d stat=%s per=%d", r.ItemID, r.Stat, r.PerEnchant)
	if r.MinEnchant > 0 {
		fmt.Fprintf(&sb, " minE=%d", r.MinEnchant)
	}
	if r.EnchantOffset > 0 {
		fmt.Fprintf(&sb, " offset=%d", r.EnchantOffset)
	}
	if r.Note != "" {
		fmt.Fprintf(&sb, " note=%q", r.Note)
	}
	sb.WriteByte('}')
	return sb.String()
}
