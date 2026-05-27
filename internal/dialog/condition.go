package dialog

import (
	"time"

	"github.com/l1jgo/server/internal/world"
)

// EvalCondition 評估單一條件子句對玩家的真值。
// 同一個 Condition 物件實務上只會啟用一個欄位（YAML 鍵唯一）；若多欄位都填，
// 全部 AND 起來。
func EvalCondition(c *Condition, p *world.PlayerInfo) bool {
	if c == nil || p == nil {
		return c == nil // nil condition = always true（default 分支）
	}
	now := time.Now().Unix()

	if c.LevelLt != nil {
		if !(p.Level < *c.LevelLt) {
			return false
		}
	}
	if c.LevelGte != nil {
		if !(p.Level >= *c.LevelGte) {
			return false
		}
	}
	if c.CooldownActive != "" {
		v := readPlayerInt64Field(p, c.CooldownActive)
		if !(v > now) {
			return false
		}
	}
	if c.CooldownPassed != "" {
		v := readPlayerInt64Field(p, c.CooldownPassed)
		if !(v <= now) {
			return false
		}
	}
	if c.HasItem != nil {
		count := c.HasItem.Count
		if count <= 0 {
			count = 1
		}
		if !inventoryHasItem(p, c.HasItem.ID, count) {
			return false
		}
	}
	if c.LacksItem != nil {
		count := c.LacksItem.Count
		if count <= 0 {
			count = 1
		}
		if inventoryHasItem(p, c.LacksItem.ID, count) {
			return false
		}
	}
	if c.InDungeon != nil {
		// 暫不檢查具體 dungeon ID 比對；只判 ShowID > 0（在副本內）
		// TODO: 加 quest world 查 dungeon ID 對齊
		if p.ShowID <= 0 {
			return false
		}
		_ = *c.InDungeon
	}
	if c.NotInDungeon {
		if p.ShowID > 0 {
			return false
		}
	}
	if c.ClassMaskHas != nil {
		mask := int32(1) << uint(p.ClassType)
		if mask&*c.ClassMaskHas == 0 {
			return false
		}
	}
	return true
}

// EvalAll 對一組條件做 AND 評估。空陣列 = true。
func EvalAll(conds []Condition, p *world.PlayerInfo) bool {
	for i := range conds {
		if !EvalCondition(&conds[i], p) {
			return false
		}
	}
	return true
}

// readPlayerInt64Field 由欄位名稱讀 PlayerInfo 的 int64 欄位（白名單，避免反射開放整個結構）。
// 未列入白名單的欄位回 0。
func readPlayerInt64Field(p *world.PlayerInfo, field string) int64 {
	if p == nil {
		return 0
	}
	switch field {
	case "NextHansBagAt":
		return p.NextHansBagAt
	// 未來新增其他冷卻/時間欄位請補上面 case
	default:
		return 0
	}
}

// writePlayerInt64Field 由欄位名稱寫 int64 欄位（同樣白名單）。
// 失敗回 false。
func writePlayerInt64Field(p *world.PlayerInfo, field string, value int64) bool {
	if p == nil {
		return false
	}
	switch field {
	case "NextHansBagAt":
		p.NextHansBagAt = value
		return true
	default:
		return false
	}
}

// inventoryHasItem 檢查玩家是否擁有指定 itemID 至少 count 個。
// 處理 stackable（單 slot 多數量）與 non-stackable（多個 slot）兩種情況。
func inventoryHasItem(p *world.PlayerInfo, itemID, count int32) bool {
	if p == nil || p.Inv == nil {
		return false
	}
	var total int32
	for _, it := range p.Inv.Items {
		if it.ItemID == itemID {
			c := it.Count
			if c <= 0 {
				c = 1
			}
			total += c
			if total >= count {
				return true
			}
		}
	}
	return total >= count
}
