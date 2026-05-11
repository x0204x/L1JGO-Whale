package data

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ResolventEntry 是溶解劑可轉換的物品設定。
// Java: resolvent.item_id / crystal_count
type ResolventEntry struct {
	ItemID       int32  `yaml:"item_id"`
	Note         string `yaml:"note"`
	CrystalCount int32  `yaml:"crystal_count"`
}

type resolventFile struct {
	Items []ResolventEntry `yaml:"items"`
}

// ResolventTable 以物品 ID 查詢溶解後的基礎魔法結晶體數量。
type ResolventTable struct {
	byItemID map[int32]int32
}

// LoadResolventTable 從 YAML 載入溶解表。
func LoadResolventTable(path string) (*ResolventTable, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read resolvent_list: %w", err)
	}
	var f resolventFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parse resolvent_list: %w", err)
	}

	t := &ResolventTable{byItemID: make(map[int32]int32, len(f.Items))}
	for _, item := range f.Items {
		if item.ItemID <= 0 || item.CrystalCount <= 0 {
			continue
		}
		t.byItemID[item.ItemID] = item.CrystalCount
	}
	return t, nil
}

// CrystalCount 回傳物品的基礎結晶數量；查無資料時回傳 0。
func (t *ResolventTable) CrystalCount(itemID int32) int32 {
	if t == nil {
		return 0
	}
	return t.byItemID[itemID]
}

// Count 回傳已載入的溶解條目數量。
func (t *ResolventTable) Count() int {
	if t == nil {
		return 0
	}
	return len(t.byItemID)
}
