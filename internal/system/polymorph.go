package system

import (
	"fmt"
	"strings"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

// PolymorphSystem 處理玩家變身與解除變身邏輯。
type PolymorphSystem struct {
	deps *handler.Deps
}

func NewPolymorphSystem(deps *handler.Deps) *PolymorphSystem {
	return &PolymorphSystem{deps: deps}
}

const (
	directPolyScrollDurationSec = 1800
	directPolyWeaponEquipAll    = 2047
	directPolyArmorEquipAll     = 8191
)

var directSharnaPolyIDs = map[int32][7][2]int32{
	handler.ItemSharnaPolyLevel30: {
		{6822, 6823}, {6824, 6825}, {6826, 6827}, {6828, 6829},
		{6830, 6831}, {7139, 7140}, {7141, 7142},
	},
	handler.ItemSharnaPolyLevel40: {
		{6832, 6833}, {6834, 6835}, {6836, 6837}, {6838, 6839},
		{6840, 6841}, {7143, 7144}, {7145, 7146},
	},
	handler.ItemSharnaPolyLevel52: {
		{6842, 6843}, {6844, 6845}, {6846, 6847}, {6848, 6849},
		{6850, 6851}, {7147, 7148}, {7149, 7150},
	},
	handler.ItemSharnaPolyLevel55: {
		{6852, 6853}, {6854, 6855}, {6856, 6857}, {6858, 6859},
		{6860, 6861}, {7151, 7152}, {7153, 7154},
	},
	handler.ItemSharnaPolyLevel60: {
		{6862, 6863}, {6864, 6865}, {6866, 6867}, {6868, 6869},
		{6870, 6871}, {7155, 7156}, {7157, 7158},
	},
	handler.ItemSharnaPolyLevel65: {
		{6872, 6873}, {6874, 6875}, {6876, 6877}, {6878, 6879},
		{6880, 6881}, {7159, 7160}, {7161, 7162},
	},
	handler.ItemSharnaPolyLevel70: {
		{6882, 6883}, {6884, 6885}, {6886, 6887}, {6888, 6889},
		{6890, 6891}, {7163, 7164}, {7165, 7166},
	},
}

// ==================== 變身 ====================

// DoPoly implements handler.PolymorphManager — 將玩家變身為指定形態。
// cause: PolyCauseMagic(1), PolyCauseGM(2), PolyCauseNPC(4). cause=0 bypasses cause check.
// durationSec: buff 持續秒數（0 = 永久直到取消）。
func (s *PolymorphSystem) DoPoly(player *world.PlayerInfo, polyID int32, durationSec int, cause int) {
	if player == nil || player.Dead {
		return
	}
	if s.deps.Polys == nil {
		return
	}

	poly := s.deps.Polys.GetByID(polyID)
	if poly == nil {
		return
	}

	// Cause check (Java: isMatchCause)
	if !poly.IsMatchCause(cause) {
		return
	}

	// 英雄變身限制（Java: L1PolyMorph GFX 13715-13745 需排名上榜）
	if polyID >= 13715 && polyID <= 13745 {
		if s.deps.Ranking == nil || !s.deps.Ranking.IsHero(player.Name) {
			handler.SendServerMessage(player.Session, 181) // "你無法使用。"
			return
		}
	}

	s.applyPoly(player, poly, durationSec)
}

func (s *PolymorphSystem) applyPoly(player *world.PlayerInfo, poly *data.PolymorphInfo, durationSec int) {
	if player == nil || player.Dead || poly == nil {
		return
	}
	polyID := poly.PolyID

	// 已有變身則先解除
	if player.TempCharGfx > 0 {
		s.UndoPoly(player)
	}

	// 設定變身狀態
	player.TempCharGfx = polyID
	player.PolyID = polyID

	// 檢查武器相容性 — 變身形態不允許的武器則隱藏視覺
	if player.CurrentWeapon != 0 && s.deps.Items != nil {
		wpn := player.Equip.Weapon()
		if wpn != nil {
			wpnInfo := s.deps.Items.Get(wpn.ItemID)
			if wpnInfo != nil && !poly.IsWeaponEquipable(wpnInfo.Type) {
				player.CurrentWeapon = 0
			}
		}
	}

	// 廣播外觀變更
	if s.deps.World != nil {
		nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
		for _, viewer := range nearby {
			handler.SendChangeShape(viewer.Session, player.CharID, polyID, player.CurrentWeapon)
		}
	}

	// 強制脫下不相容裝備
	if s.deps.Items != nil {
		s.forceUnequipIncompat(player, poly)
	}

	// 註冊為 buff（skillID=67 變形術）
	if durationSec > 0 {
		buff := &world.ActiveBuff{
			SkillID:   handler.SkillShapeChange,
			TicksLeft: durationSec * 5, // 秒 → tick（每 tick 200ms）
		}
		old := player.AddBuff(buff)
		if old != nil && s.deps.Skill != nil {
			s.deps.Skill.RevertBuffStats(player, old)
		}

		// 發送變身計時圖示：S_PacketBox sub 35
		if player.Session != nil {
			handler.SendPolyIcon(player.Session, uint16(durationSec))
		}
	}

	if s.deps.Log != nil {
		s.deps.Log.Info(fmt.Sprintf("玩家變身  角色=%s  形態=%s(GFX:%d)  持續=%d秒",
			player.Name, poly.Name, polyID, durationSec))
	}
}

// ==================== 解除變身 ====================

// UndoPoly implements handler.PolymorphManager — 解除玩家變身，恢復原始外觀。
func (s *PolymorphSystem) UndoPoly(player *world.PlayerInfo) {
	if player.TempCharGfx == 0 {
		return // 未變身
	}

	player.TempCharGfx = 0
	player.PolyID = 0

	// 恢復武器視覺
	if wpn := player.Equip.Weapon(); wpn != nil {
		wpnInfo := s.deps.Items.Get(wpn.ItemID)
		if wpnInfo != nil {
			player.CurrentWeapon = world.WeaponVisualID(wpnInfo.Type)
		}
	}

	// 廣播原始外觀
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	for _, viewer := range nearby {
		handler.SendChangeShape(viewer.Session, player.CharID, player.ClassID, player.CurrentWeapon)
	}

	// 取消變身計時圖示
	handler.SendPolyIcon(player.Session, 0)

	// 移除變形術 buff
	player.RemoveBuff(handler.SkillShapeChange)

	s.deps.Log.Info(fmt.Sprintf("玩家解除變身  角色=%s", player.Name))
}

// ==================== 內部輔助函式 ====================

// ==================== 變身卷軸 / 技能 ====================

// UsePolyScroll 處理變身卷軸使用（業務邏輯從 handler/polymorph.go 搬入）。
// monsterName="" 表示取消目前變身。
func (s *PolymorphSystem) UsePolyScroll(sess *net.Session, player *world.PlayerInfo, invItem *world.InvItem, monsterName string) {
	// 空字串 = 取消變身（Java: s.equals("")）
	if monsterName == "" {
		if player.TempCharGfx != 0 {
			s.UndoPoly(player)
			// 消耗卷軸
			removed := player.Inv.RemoveItem(invItem.ObjectID, 1)
			if removed {
				handler.SendRemoveInventoryItem(sess, invItem.ObjectID)
			} else {
				handler.SendItemCountUpdate(sess, invItem)
			}
			handler.SendWeightUpdate(sess, player)
		}
		return
	}

	if s.deps.Polys == nil {
		return
	}

	// 查詢變身形態
	poly := s.deps.Polys.GetByName(monsterName)
	if poly == nil {
		handler.SendServerMessage(sess, 181) // "無法變成你指定的怪物。"
		return
	}

	// 等級檢查（Java: poly.getMinLevel() <= pc.getLevel()）
	if poly.MinLevel > 0 && int(player.Level) < poly.MinLevel {
		handler.SendServerMessage(sess, 181)
		return
	}

	// 原因檢查 — 卷軸使用 PolyCauseMagic (1)
	if !poly.IsMatchCause(data.PolyCauseMagic) {
		handler.SendServerMessage(sess, 181)
		return
	}

	// 計算持續時間
	duration := handler.PolyScrollDuration(invItem.ItemID)

	// 執行變身
	s.DoPoly(player, poly.PolyID, duration, data.PolyCauseMagic)

	// 消耗卷軸
	removed := player.Inv.RemoveItem(invItem.ObjectID, 1)
	if removed {
		handler.SendRemoveInventoryItem(sess, invItem.ObjectID)
	} else {
		handler.SendItemCountUpdate(sess, invItem)
	}
	handler.SendWeightUpdate(sess, player)
}

// UseDirectPolyScroll 處理使用後直接變身的特殊卷軸。
func (s *PolymorphSystem) UseDirectPolyScroll(sess *net.Session, player *world.PlayerInfo, invItem *world.InvItem) {
	if sess == nil || player == nil || invItem == nil || player.Dead {
		return
	}
	if isDirectPolyBannedMap(player.MapID) {
		handler.SendServerMessage(sess, 1170) // 這裡不可以變身。
		return
	}

	poly, ok := directPolyScrollInfo(player, invItem.ItemID)
	if !ok {
		handler.SendServerMessage(sess, 79) // 沒有任何事情發生。
		return
	}
	if poly.MinLevel > 0 && int(player.Level) < poly.MinLevel {
		handler.SendServerMessage(sess, 181)
		return
	}

	// TODO: PlayerInfo 尚未暴露龍騎士覺醒狀態；待覺醒狀態資料化後需對齊 Java 185/190/195 變身限制。
	s.applyPoly(player, poly, directPolyScrollDurationSec)
	consumePolyScroll(sess, player, invItem)
}

func directPolyScrollInfo(player *world.PlayerInfo, itemID int32) (*data.PolymorphInfo, bool) {
	if itemID == handler.ItemOrcEmissaryPolyScroll {
		return &data.PolymorphInfo{
			PolyID:      6984,
			Name:        "orc spy",
			MinLevel:    1,
			WeaponEquip: directPolyWeaponEquipAll,
			ArmorEquip:  directPolyArmorEquipAll,
			CanUseSkill: true,
			Cause:       data.PolyCauseMagic | data.PolyCauseGM | data.PolyCauseNPC,
		}, true
	}

	forms, ok := directSharnaPolyIDs[itemID]
	if !ok {
		return nil, false
	}
	classType := int(player.ClassType)
	sex := int(player.Sex)
	if classType < 0 || classType >= len(forms) || sex < 0 || sex > 1 {
		return nil, false
	}
	polyID := forms[classType][sex]
	if polyID == 0 {
		return nil, false
	}
	minLevel := directSharnaMinLevel(itemID)
	return &data.PolymorphInfo{
		PolyID:      polyID,
		Name:        fmt.Sprintf("夏納的變身卷軸(等級%d)", minLevel),
		MinLevel:    minLevel,
		WeaponEquip: directPolyWeaponEquipAll,
		ArmorEquip:  directPolyArmorEquipAll,
		CanUseSkill: true,
		Cause:       data.PolyCauseMagic | data.PolyCauseGM | data.PolyCauseNPC,
	}, true
}

func directSharnaMinLevel(itemID int32) int {
	switch itemID {
	case handler.ItemSharnaPolyLevel30:
		return 30
	case handler.ItemSharnaPolyLevel40:
		return 40
	case handler.ItemSharnaPolyLevel52:
		return 52
	case handler.ItemSharnaPolyLevel55:
		return 55
	case handler.ItemSharnaPolyLevel60:
		return 60
	case handler.ItemSharnaPolyLevel65:
		return 65
	case handler.ItemSharnaPolyLevel70:
		return 70
	}
	return 0
}

func isDirectPolyBannedMap(mapID int16) bool {
	switch mapID {
	case 5300, 9000, 9100, 9101, 9102, 9202:
		return true
	}
	return false
}

func consumePolyScroll(sess *net.Session, player *world.PlayerInfo, invItem *world.InvItem) {
	removed := player.Inv.RemoveItem(invItem.ObjectID, 1)
	if removed {
		handler.SendRemoveInventoryItem(sess, invItem.ObjectID)
	} else {
		handler.SendItemCountUpdate(sess, invItem)
	}
	handler.SendWeightUpdate(sess, player)
}

// UsePolySkill 處理變形術技能選擇對話框結果（業務邏輯從 handler/polymorph.go 搬入）。
func (s *PolymorphSystem) UsePolySkill(sess *net.Session, player *world.PlayerInfo, monsterName string) {
	if player == nil || !player.PendingPolySkill {
		return
	}
	if isPolyListNavigation(monsterName) {
		handler.SendHypertext(sess, player.CharID, monsterName)
		return
	}
	player.PendingPolySkill = false
	if monsterName == "" {
		return
	}
	if s.deps.Polys == nil {
		return
	}

	poly := s.deps.Polys.GetByName(monsterName)
	if poly == nil {
		handler.SendServerMessage(sess, 181) // "此怪物名稱不正確。"
		return
	}

	// 等級檢查
	if poly.MinLevel > 0 && int(player.Level) < poly.MinLevel {
		handler.SendServerMessage(sess, 181)
		return
	}

	// 原因檢查 — 技能 67 為魔法原因
	if !poly.IsMatchCause(data.PolyCauseMagic) {
		handler.SendServerMessage(sess, 181)
		return
	}

	// 執行變身：7200 秒 = 2 小時（Java 預設）
	s.DoPoly(player, poly.PolyID, 7200, data.PolyCauseMagic)
	handler.SendCloseList(sess, player.CharID)
}

func isPolyListNavigation(action string) bool {
	action = strings.ToLower(strings.TrimSpace(action))
	return strings.HasPrefix(action, "monlist") || strings.HasPrefix(action, "sh_monlist")
}

// ==================== 內部輔助函式 ====================

// forceUnequipIncompat 強制脫下與當前變身形態不相容的所有裝備。
// Java: L1PolyMorph.doPoly() → takeoff loop
func (s *PolymorphSystem) forceUnequipIncompat(player *world.PlayerInfo, poly *data.PolymorphInfo) {
	sess := player.Session

	for slot := world.EquipSlot(1); slot < world.SlotMax; slot++ {
		item := player.Equip.Get(slot)
		if item == nil {
			continue
		}

		itemInfo := s.deps.Items.Get(item.ItemID)
		if itemInfo == nil {
			continue
		}

		shouldUnequip := false

		if slot == world.SlotWeapon {
			// 檢查武器相容性
			if !poly.IsWeaponEquipable(itemInfo.Type) {
				shouldUnequip = true
			}
		} else {
			// 檢查防具相容性
			if !poly.IsArmorEquipable(itemInfo.Type) {
				shouldUnequip = true
			}
		}

		if shouldUnequip {
			// 詛咒物品（bless == 2）即使變身也無法脫下
			if item.Bless == 2 {
				continue
			}
			if s.deps.Equip != nil {
				s.deps.Equip.UnequipSlot(sess, player, slot)
			}
		}
	}
}
