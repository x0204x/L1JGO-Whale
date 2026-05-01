package system

import (
	"fmt"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

const (
	skillCurseBlindEffect = int32(40)
	skillStatusHaste      = int32(1001)
)

func (s *SkillSystem) applyCurseBlindEffect(target *world.PlayerInfo, skill *data.SkillInfo) {
	if target == nil || skill == nil {
		return
	}
	s.removeBuffAndRevert(target, skillCurseBlindEffect)
	dur := skill.BuffDuration
	if dur <= 0 {
		dur = 8
	}
	target.AddBuff(&world.ActiveBuff{
		SkillID:   skillCurseBlindEffect,
		TicksLeft: dur * 5,
	})
	handler.SendCurseBlind(target.Session, 1)
}

func (s *SkillSystem) removeCurseBlindEffect(target *world.PlayerInfo) {
	if target == nil {
		return
	}
	s.removeBuffAndRevert(target, skillCurseBlindEffect)
	handler.SendCurseBlind(target.Session, 0)
}

func (s *SkillSystem) handleOppositeMoveSpeedSkill(target *world.PlayerInfo, skillID int32) bool {
	if target == nil {
		return false
	}
	switch skillID {
	case 43, 54:
		if target.MoveSpeed != 2 {
			return false
		}
		s.removeMoveSpeedBuffs(target, []int32{29, 76, 152})
		return true
	case 29, 76, 152:
		if target.MoveSpeed != 1 {
			return false
		}
		s.removeMoveSpeedBuffs(target, []int32{43, 54, skillStatusHaste})
		return true
	default:
		return false
	}
}

func (s *SkillSystem) removeMoveSpeedBuffs(target *world.PlayerInfo, skillIDs []int32) {
	for _, skillID := range skillIDs {
		s.removeBuffAndRevert(target, skillID)
	}
	target.MoveSpeed = 0
	target.HasteTicks = 0
	s.sendSpeedToAll(target, 0, 0)
}

func (s *SkillSystem) applyShockStunToPlayer(sess *net.Session, caster, target *world.PlayerInfo, skill *data.SkillInfo, nearby []*world.PlayerInfo) bool {
	if caster.CharID == target.CharID {
		return true
	}
	if !s.hasTwoHandSwordEquipped(caster) {
		handler.SendGlobalChat(sess, 9, "\\f3請使用雙手劍。")
		return true
	}
	if target.HasBuff(87) {
		return true
	}
	dur := shockStunDurationSeconds()
	stunSkill := *skill
	stunSkill.BuffDuration = dur
	s.applyBuffEffect(target, &stunSkill)
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
	}
	if s.deps != nil && s.deps.Log != nil {
		s.deps.Log.Info(fmt.Sprintf("衝擊之暈  施法者=%s  玩家=%s  持續=%d秒", caster.Name, target.Name, dur))
	}
	return true
}

func shockStunDurationSeconds() int {
	return 1 + world.RandInt(6)
}

func (s *SkillSystem) hasTwoHandSwordEquipped(player *world.PlayerInfo) bool {
	if player == nil {
		return false
	}
	wpn := player.Equip.Weapon()
	if wpn == nil || s.deps == nil || s.deps.Items == nil {
		return false
	}
	info := s.deps.Items.Get(wpn.ItemID)
	return info != nil && info.Type == "tohandsword"
}

// ========================================================================
//  NPC Debuff
// ========================================================================

// executeNpcDebuffSkill 對 NPC 施加 debuff 技能。
func (s *SkillSystem) executeNpcDebuffSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, npc *world.NpcInfo) {
	ws := s.deps.World

	maxRange := int32(skill.Ranged)
	if maxRange <= 0 {
		maxRange = 10
	}
	if chebyshevDist(player.X, player.Y, npc.X, npc.Y) > maxRange+2 {
		return
	}

	player.Heading = CalcHeading(player.X, player.Y, npc.X, npc.Y)

	// 對 NPC 施放 debuff 技能 → 累加仇恨（讓 NPC 追擊施法者）
	AddHate(npc, sess.ID, 1)

	nearby := ws.GetNearbyPlayersAt(npc.X, npc.Y, npc.MapID)

	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))

	switch skill.SkillID {
	case handler.SkillShapeChange: // Owner: skill_polymorph.go
		s.executeShapeChangeNpc(sess, player, skill, npc, nearby)

	case 133: // 弱化屬性 — Owner: skill_elemental.go
		if player.ElfAttr == 0 {
			handler.SendServerMessage(sess, 79)
			return
		}
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 32
		}
		applyElementalFallDownToNpc(player, npc, dur)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("弱化屬性  施法者=%s  NPC=%s  屬性=%d  持續=%d秒", player.Name, npc.Name, player.ElfAttr, dur))

	case 27: // 壞物術 — Owner: skill_weapon.go；NPC 近戰傷害減半，直到相消術清除
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		npc.WeaponBroken = true
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("壞物術  施法者=%s  NPC=%s", player.Name, npc.Name))

	case 87: // 衝擊之暈 — 需要雙手劍
		if !s.hasTwoHandSwordEquipped(player) {
			handler.SendGlobalChat(sess, 9, "\\f3請使用雙手劍。")
			return
		}
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := shockStunDurationSeconds()
		npc.Paralyzed = true
		npc.AddDebuff(87, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("衝擊之暈  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 157: // 大地屏障 — 凍結 + 灰色色調
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := 1 + world.RandInt(12)
		npc.Paralyzed = true
		npc.AddDebuff(157, dur*5)
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("大地屏障  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 103: // 暗黑盲咒
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 3
		}
		npc.Sleeped = true
		npc.AddDebuff(66, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("暗黑盲咒  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 66: // 沉睡之霧
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 10
		}
		npc.Sleeped = true
		npc.AddDebuff(66, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("沉睡之霧  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 33: // 木乃伊詛咒（NPC 版）— 階段一：灰色延遲
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		if npc.Paralyzed || npc.HasDebuff(33) || npc.HasDebuff(4001) {
			return
		}
		npc.AddDebuff(33, 25)
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("木乃伊詛咒(階段一)  施法者=%s  NPC=%s  延遲=5秒", player.Name, npc.Name))

	case 11: // 毒咒 — 對 NPC 施加傷害毒（Java: L1DamagePoison.doInfection, 3000ms, 5dmg）
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		npc.PoisonDmgAmt = 5
		npc.PoisonDmgTimer = 0
		npc.PoisonAttackerSID = sess.ID // 仇恨歸屬
		AddHate(npc, sess.ID, 1)
		npc.AddDebuff(11, 150) // 30 秒 = 150 ticks
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 1))
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("毒咒  施法者=%s  NPC=%s  持續=30秒  每次=5傷害", player.Name, npc.Name))

	case 29, 76, 152: // 緩速系列
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 64
		}
		npc.AddDebuff(skill.SkillID, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("緩速術  施法者=%s  NPC=%s  技能=%d  持續=%d秒", player.Name, npc.Name, skill.SkillID, dur))

	case 50: // 冰矛 — NPC 凍結（Java: setFrozen + S_Poison 灰色）
		if npc.Paralyzed || npc.HasDebuff(50) || npc.HasDebuff(80) || npc.HasDebuff(22) || npc.HasDebuff(30) {
			break // 已被凍結
		}
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 8
		}
		npc.Paralyzed = true
		npc.AddDebuff(50, (dur+1)*5)
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("冰矛凍結  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur+1))

	case 80: // 冰雪颶風 — 對 NPC 施加凍結（Java: setFrozen + S_Poison 灰色）
		if npc.Paralyzed || npc.HasDebuff(50) || npc.HasDebuff(80) {
			break // 已被凍結/冰矛
		}
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			break // 抗性判定失敗不阻止傷害，只是不凍結
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 16
		}
		npc.Paralyzed = true
		npc.AddDebuff(80, (dur+1)*5) // Java: buffDuration + 1
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("冰雪颶風凍結  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur+1))

	case 47: // 弱化術 — NPC debuff（Java: DMG-5, HIT-1）
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 64
		}
		npc.AddDebuff(47, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("弱化術  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 56: // 疾病術 — NPC debuff（Java: DMG-6, AC+12）
		if !s.checkNpcMRResist(player, npc, skill.SkillID) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 64
		}
		npc.AddDebuff(56, dur*5)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("疾病術  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 112: // 破壞盔甲（NPC debuff）— Java: ARMOR_BREAK.java 對怪物/召喚/寵物
		if !s.calcArmorBreakProbNpc(player, npc) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 8
		}
		// 移除舊效果 → 重新套用
		npc.AddDebuff(112, dur*5) // 8 秒 = 40 ticks
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		handler.SendGlobalChat(sess, 9, "\\f2破壞盔甲 施放成功!")
		s.deps.Log.Info(fmt.Sprintf("破壞盔甲  施法者=%s  NPC=%s  持續=%d秒", player.Name, npc.Name, dur))

	case 44: // 魔法相消術 — 解除 NPC 所有 debuff + 狀態（Java: CANCELLATION.java:158-167）
		// 清除所有 debuffs
		for debuffID := range npc.ActiveDebuffs {
			delete(npc.ActiveDebuffs, debuffID)
		}
		hadNpcPoly := npc.PolyOriginalGfxID != 0
		clearNpcCancellationState(npc)
		// 清除所有視覺效果（毒色/灰色）
		handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 0))
		if hadNpcPoly {
			for _, viewer := range nearby {
				handler.SendChangeShape(viewer.Session, npc.ID, npc.GfxID, 0)
			}
		}
		// 施法特效
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
		s.deps.Log.Info(fmt.Sprintf("魔法相消術(NPC)  施法者=%s  NPC=%s", player.Name, npc.Name))

	default:
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
		}
	}
}

// checkNpcMRResist 檢查 NPC 魔法抗性。
func (s *SkillSystem) checkNpcMRResist(caster *world.PlayerInfo, npc *world.NpcInfo, _ int32) bool {
	prob := 50 + (int(caster.Level)-int(npc.Level))*5 + int(caster.Intel)*2 - int(npc.MR)
	if prob < 5 {
		prob = 5
	}
	if prob > 95 {
		prob = 95
	}
	return world.RandInt(100) < prob
}

// playerDebuffSkills 需要對玩家目標進行 MR 抗性判定的 debuff 技能。
// 這些技能對其他玩家施放時，必須通過魔法抗性檢查才能命中。
var playerDebuffSkills = map[int32]bool{
	11:  true, // 毒咒
	20:  true, // 闇盲咒術
	27:  true, // 壞物術
	29:  true, // 緩速術
	33:  true, // 木乃伊詛咒
	40:  true, // 黑闇之影
	47:  true, // 弱化術
	56:  true, // 疾病術
	66:  true, // 沉睡之霧
	71:  true, // 藥水霜化術
	76:  true, // 集體緩速術
	87:  true, // 衝擊之暈
	103: true, // 暗黑盲咒
	133: true, // 弱化屬性
	152: true, // 究極緩速術
	157: true, // 大地屏障
	161: true, // 封印禁地
	173: true, // 污濁之水
	174: true, // 精準射擊
}

// checkPlayerMRResist 對玩家目標的魔法抗性判定（debuff 用）。
// 簡化版公式（Java L1MagicPc.calcProbabilityMagic 的核心概念）：
//
//	prob = 50 + (casterLevel - targetLevel) * 3 + casterINT - targetMR
//	clamp(prob, 10, 90)
//	success = rand(100) < prob
func (s *SkillSystem) checkPlayerMRResist(caster, target *world.PlayerInfo) bool {
	prob := 50 + (int(caster.Level)-int(target.Level))*3 + int(caster.Intel) - int(target.MR)
	if prob < 10 {
		prob = 10
	}
	if prob > 90 {
		prob = 90
	}
	return world.RandInt(100) < prob
}

// calcArmorBreakProb 破壞盔甲對玩家目標的機率判定。
// Java: L1MagicPc.calcProbabilityMagic(ARMOR_BREAK) — 非標準 MR 判定，使用等級比較系統。
// 攻擊者等級 > 防禦者 → 5%；相等 → 10%；攻擊者 < 防禦者 → 15%
// 加上純智力加成（INT 25-44: +(INT-15)/10, INT 45+: +5）
func (s *SkillSystem) calcArmorBreakProb(caster, target *world.PlayerInfo) bool {
	atkLv := int(caster.Level)
	defLv := int(target.Level)

	var prob int
	if atkLv > defLv {
		prob = 5
	} else if atkLv == defLv {
		prob = 10
	} else {
		prob = 15
	}

	// Java: probability += magichit + INT 加成
	baseInt := int(caster.Intel)
	if baseInt >= 25 && baseInt <= 44 {
		prob += (baseInt - 15) / 10
	} else if baseInt >= 45 {
		prob += 5
	}

	if prob < 1 {
		prob = 1
	}
	return world.RandInt(100) < prob
}

// calcArmorBreakProbNpc 破壞盔甲對 NPC 目標的機率判定。
// 與玩家版本相同的機率系統，但使用 NPC 等級。
func (s *SkillSystem) calcArmorBreakProbNpc(caster *world.PlayerInfo, npc *world.NpcInfo) bool {
	atkLv := int(caster.Level)
	defLv := int(npc.Level)

	var prob int
	if atkLv > defLv {
		prob = 5
	} else if atkLv == defLv {
		prob = 10
	} else {
		prob = 15
	}

	baseInt := int(caster.Intel)
	if baseInt >= 25 && baseInt <= 44 {
		prob += (baseInt - 15) / 10
	} else if baseInt >= 45 {
		prob += 5
	}

	if prob < 1 {
		prob = 1
	}
	return world.RandInt(100) < prob
}
