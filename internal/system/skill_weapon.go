package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

func calcWeaponBreakDurabilityDamage(caster *world.PlayerInfo) int8 {
	maxDamage := 1
	if caster != nil {
		maxDamage = int(caster.Intel) / 3
		if maxDamage < 1 {
			maxDamage = 1
		}
	}
	return int8(world.RandInt(maxDamage) + 1)
}

func applyWeaponBreakDurability(weapon *world.InvItem, amount int8) bool {
	if weapon == nil || amount <= 0 {
		return false
	}
	before := weapon.Durability
	next := int16(weapon.Durability) + int16(amount)
	maxDurability := int16(weapon.EnchantLvl) + 5
	if maxDurability < 0 {
		maxDurability = 0
	}
	if maxDurability > 127 {
		maxDurability = 127
	}
	if next > maxDurability {
		next = maxDurability
	}
	weapon.Durability = int8(next)
	return weapon.Durability != before
}

func clearNpcCancellationState(npc *world.NpcInfo) {
	if npc == nil {
		return
	}
	npc.PoisonDmgAmt = 0
	npc.PoisonDmgTimer = 0
	npc.Paralyzed = false
	npc.Sleeped = false
	npc.WeaponBroken = false
	removeShapeChangeFromNpc(npc)
}

func applyNpcWeaponBreakDamage(npc *world.NpcInfo, damage int32) int32 {
	if npc == nil || !npc.WeaponBroken || damage <= 0 {
		return damage
	}
	return damage / 2
}

// ========================================================================
//  鎧甲強化技能
// ========================================================================

// executeArmorEnchant 處理鎧甲護持（skill 21）— 物品強化技能。
// Java: targetID = 背包物品 ObjectID。檢查物品是否為身體鎧甲（type2=2, type=2），
// 是 → AC-3 buff + 訊息 161；否 → 訊息 79「沒有任何事情發生。」
func (s *SkillSystem) executeArmorEnchant(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, itemObjID int32) {
	// 查找背包物品
	invItem := player.Inv.FindByObjectID(itemObjID)
	if invItem == nil {
		handler.SendServerMessage(sess, 79) // 沒有任何事情發生。
		return
	}

	// 查詢物品模板 — 必須為身體鎧甲（Java: type2==2 && type==2）
	itemInfo := s.deps.Items.Get(invItem.ItemID)
	if itemInfo == nil || itemInfo.Category != data.CategoryArmor || itemInfo.Type != "armor" {
		handler.SendServerMessage(sess, 79) // 沒有任何事情發生。
		return
	}

	// 施法動畫 + GFX
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}

	applySkillArmorEnchant(invItem, skill)
	player.Dirty = true
	if invItem.Equipped || player.Equip.Get(world.SlotArmor) == invItem {
		handler.RecalcEquipStats(sess, player, s.deps)
	}

	handler.SendServerMessageArgs(sess, 161, invItem.Name, "$245", "$247")
}

// executeWeaponEnchant 處理擬似魔法武器（skill 12）和暗影之牙（skill 107）— 武器強化 buff。
// Java: targetID = 背包物品 ObjectID。檢查物品是否為武器（type2=1），
// 是 → 套用武器強化 buff + icon；否 → 訊息 79。
func (s *SkillSystem) executeWeaponEnchant(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, itemObjID int32) {
	s.executeTargetedWeaponEnchant(sess, player, skill, itemObjID)
}

// executeCreateMagicalWeapon 處理創造魔法武器（skill 73）— 武器強化 +1。
// Java: 僅可對 safe_enchant > 0 且 enchant_level == 0 的武器使用。
// Go 簡化：驗證物品為武器即可，完整強化邏輯待後續實作。
func (s *SkillSystem) executeCreateMagicalWeapon(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, itemObjID int32) {
	invItem := player.Inv.FindByObjectID(itemObjID)
	if invItem == nil {
		handler.SendServerMessage(sess, 79)
		return
	}

	itemInfo := s.deps.Items.Get(invItem.ItemID)
	if itemInfo == nil || itemInfo.Category != data.CategoryWeapon {
		handler.SendServerMessage(sess, 79)
		return
	}

	// safe_enchant 檢查（Java: safe_enchant <= 0 → msg 79）
	if itemInfo.SafeEnchant <= 0 {
		handler.SendServerMessage(sess, 79)
		return
	}

	// 只對未強化武器有效（Java: enchant_level != 0 → msg 79）
	if invItem.EnchantLvl != 0 {
		handler.SendServerMessage(sess, 79)
		return
	}

	// 廣播施法動畫
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}

	// 成功：強化 +1（100% 成功率）
	invItem.EnchantLvl = 1

	// 發送 msg 161：「%0 閃耀 %1 %2 的光芒」（Java: item_name, "$245", "$247"）
	itemName := invItem.Name
	if invItem.Identified {
		itemName = "+0 " + invItem.Name
	}
	handler.SendServerMessageArgs(sess, 161, itemName, "$245", "$247")

	// 更新物品名稱顯示
	handler.SendItemNameUpdate(sess, invItem, itemInfo)
	player.Dirty = true
}

// executeBringStone 處理提煉魔石（skill 100）— 魔石升級鏈。
// Java: 40320→40321→40322→40323→40324，各有不同成功率。
// Go 簡化：驗證物品為魔石即可，完整升級邏輯待後續實作。
func (s *SkillSystem) executeBringStone(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, itemObjID int32) {
	invItem := player.Inv.FindByObjectID(itemObjID)
	if invItem == nil {
		handler.SendServerMessage(sess, 79)
		return
	}

	// 檢查是否為可升級的魔石（Java: 40320/40321/40322/40323）
	switch invItem.ItemID {
	case 40320, 40321, 40322, 40323:
		// 有效的魔石
	default:
		handler.SendServerMessage(sess, 79)
		return
	}

	// 計算成功率與結果物品 ID
	rate, resultID, msgArg := calcBringStoneRate(player, invItem.ItemID)
	if resultID == 0 {
		handler.SendServerMessage(sess, 79)
		return
	}

	// 廣播施法動畫
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}

	// 消耗原石（Java: 無論成功失敗都消耗）
	removed := player.Inv.RemoveItem(invItem.ObjectID, 1)
	if removed {
		handler.SendRemoveInventoryItem(sess, invItem.ObjectID)
	} else {
		handler.SendItemCountUpdate(sess, invItem)
	}
	// 更新負重
	handler.SendWeightUpdate(sess, player)

	// 擲骰判定（Java: random.nextInt(100)+1，即 1~100）
	if world.RandInt(100)+1 <= rate {
		// 成功：新增升級石到背包
		resultInfo := s.deps.Items.Get(resultID)
		if resultInfo == nil {
			handler.SendServerMessage(sess, 280)
			player.Dirty = true
			return
		}
		if s.deps.ItemCreate != nil {
			if _, ok := s.deps.ItemCreate.GiveItem(sess, player, resultID, 1); !ok {
				player.Dirty = true
				return
			}
		} else {
			newItem := player.Inv.AddItem(resultID, 1, resultInfo.Name, resultInfo.InvGfx, resultInfo.Weight, resultInfo.Stackable, byte(resultInfo.Bless))
			handler.SendAddItem(sess, newItem, resultInfo)
		}
		handler.SendServerMessageStr(sess, 403, msgArg)
	} else {
		// 失敗：魔法失敗了
		handler.SendServerMessage(sess, 280)
	}
	handler.SendWeightUpdate(sess, player)
	player.Dirty = true
}

// calcBringStoneRate 計算提煉魔石的成功率。
// Java 公式：dark = floor(10 + level*0.8 + (wis-6)*1.2)，逐級除以常數。
func calcBringStoneRate(p *world.PlayerInfo, itemID int32) (rate int, resultID int32, msgArg string) {
	dark := int(10 + float64(p.Level)*0.8 + float64(p.Wis-6)*1.2)
	brave := int(float64(dark) / 2.1)
	wise := int(float64(brave) / 2.0)
	kayser := int(float64(wise) / 1.9)

	switch itemID {
	case 40320:
		return dark, 40321, "$2475"
	case 40321:
		return brave, 40322, "$2476"
	case 40322:
		return wise, 40323, "$2477"
	case 40323:
		return kayser, 40324, "$2478"
	}
	return 0, 0, ""
}

func applySkillWeaponEnchant(weapon *world.InvItem, skill *data.SkillInfo) {
	if weapon == nil || skill == nil {
		return
	}
	weapon.DmgByMagic = 0
	weapon.HitByMagic = 0
	switch skill.SkillID {
	case 12:
		weapon.DmgByMagic = 2
	case 48:
		weapon.DmgByMagic = 2
		weapon.HitByMagic = 2
	case 107:
		weapon.DmgByMagic = 5
	}
	weapon.DmgMagicExpiry = skill.BuffDuration * 5
}

func applySkillArmorEnchant(armor *world.InvItem, skill *data.SkillInfo) {
	if armor == nil || skill == nil {
		return
	}
	armor.AcByMagic = 0
	if skill.SkillID == 21 {
		armor.AcByMagic = 3
	}
	armor.AcMagicExpiry = skill.BuffDuration * 5
}

func (s *SkillSystem) executeTargetedWeaponEnchant(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, itemObjID int32) {
	weapon := player.Inv.FindByObjectID(itemObjID)
	if weapon == nil {
		handler.SendServerMessage(sess, 79)
		return
	}
	itemInfo := s.deps.Items.Get(weapon.ItemID)
	if itemInfo == nil || itemInfo.Category != data.CategoryWeapon {
		handler.SendServerMessage(sess, 79)
		return
	}

	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}

	applySkillWeaponEnchant(weapon, skill)
	player.Dirty = true
	handler.RecalcEquipStats(sess, player, s.deps)
	if skill.SkillID == 12 {
		handler.SendServerMessageArgs(sess, 161, weapon.Name, "$245", "$247")
		handler.SendWeaponEnchantIcon(sess, 747, uint16(skill.BuffDuration), true)
	} else if skill.SkillID == 107 {
		handler.SendWeaponEnchantIcon(sess, 2951, uint16(skill.BuffDuration), true)
	}
}

func (s *SkillSystem) executeBlessWeaponEnchant(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, targetID int32) {
	targetPlayer := player
	if targetID != 0 && targetID != player.CharID {
		other := s.deps.World.GetByCharID(targetID)
		if other == nil || other.Dead || other.MapID != player.MapID ||
			chebyshevDist(player.X, player.Y, other.X, other.Y) > 20 {
			handler.SendServerMessage(sess, 79)
			return
		}
		targetPlayer = other
	}
	weapon := targetPlayer.Equip.Weapon()
	if weapon == nil {
		handler.SendServerMessage(sess, 79)
		return
	}
	itemInfo := s.deps.Items.Get(weapon.ItemID)
	if itemInfo == nil || itemInfo.Category != data.CategoryWeapon {
		handler.SendServerMessage(sess, 79)
		return
	}

	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(targetPlayer.CharID, skill.CastGfx))
	}

	applySkillWeaponEnchant(weapon, skill)
	targetPlayer.Dirty = true
	handler.RecalcEquipStats(targetPlayer.Session, targetPlayer, s.deps)
	handler.SendServerMessageArgs(targetPlayer.Session, 161, weapon.Name, "$245", "$247")
}
