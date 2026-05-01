package system

import (
	"fmt"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

const immuneToHarmSkillID int32 = 68

func applyImmuneToHarmDamage(target *world.PlayerInfo, damage int32) int32 {
	if target == nil || damage <= 0 || !target.HasBuff(immuneToHarmSkillID) {
		return damage
	}
	return damage / 2
}

func (s *SkillSystem) validateSolidCarriage(player *world.PlayerInfo) bool {
	if player == nil {
		return false
	}
	return player.Equip.Get(world.SlotShield) != nil || player.Equip.Get(world.SlotGuarder) != nil
}

// ========================================================================
//  Buff 管理
// ========================================================================

// sendBuffIcon 發送適當的 buff 圖示封包。
func (s *SkillSystem) sendBuffIcon(target *world.PlayerInfo, skillID int32, durationSec uint16) {
	if s.deps.BuffIcons == nil {
		return
	}
	icon := s.deps.BuffIcons.Get(skillID)
	if icon == nil {
		return
	}
	sess := target.Session
	switch icon.Type {
	case "shield":
		handler.SendIconShield(sess, durationSec, icon.Param)
	case "strup":
		handler.SendIconStrup(sess, durationSec, byte(target.Str), icon.Param)
	case "dexup":
		handler.SendIconDexup(sess, durationSec, byte(target.Dex), icon.Param)
	case "aura":
		handler.SendIconAura(sess, byte(skillID-1), durationSec)
	case "invis":
		handler.SendInvisible(sess, target.CharID, durationSec > 0)
	case "wisdom":
		handler.SendWisdomPotionIcon(sess, durationSec)
	case "blue_potion":
		handler.SendBluePotionIcon(sess, durationSec)
	}
}

// cancelBuffIcon 取消 buff 圖示。
func (s *SkillSystem) cancelBuffIcon(target *world.PlayerInfo, skillID int32) {
	s.sendBuffIcon(target, skillID, 0)
}

// applyBuffEffect 套用屬性變化並註冊 buff 計時器。
func (s *SkillSystem) applyBuffEffect(target *world.PlayerInfo, skill *data.SkillInfo) {
	if skill.BuffDuration <= 0 {
		return
	}

	buff := &world.ActiveBuff{
		SkillID:   skill.SkillID,
		TicksLeft: skill.BuffDuration * 5,
	}

	eff := s.deps.Scripting.GetBuffEffect(int(skill.SkillID), int(target.Level))

	if eff != nil {
		// 移除衝突 buff
		for _, exID := range eff.Exclusions {
			s.removeBuffAndRevert(target, int32(exID))
		}

		// 設定屬性差值
		buff.DeltaAC = int16(eff.AC)
		buff.DeltaStr = int16(eff.Str)
		buff.DeltaDex = int16(eff.Dex)
		buff.DeltaCon = int16(eff.Con)
		buff.DeltaWis = int16(eff.Wis)
		buff.DeltaIntel = int16(eff.Intel)
		buff.DeltaCha = int16(eff.Cha)
		buff.DeltaMaxHP = int32(eff.MaxHP)
		buff.DeltaMaxMP = int32(eff.MaxMP)
		buff.DeltaHitMod = int16(eff.HitMod)
		buff.DeltaDmgMod = int16(eff.DmgMod)
		buff.DeltaSP = int16(eff.SP)
		buff.DeltaMR = int16(eff.MR)
		buff.DeltaHPR = int16(eff.HPR)
		buff.DeltaMPR = int16(eff.MPR)
		buff.DeltaBowHit = int16(eff.BowHit)
		buff.DeltaBowDmg = int16(eff.BowDmg)
		buff.DeltaDodge = int16(eff.Dodge)
		buff.DeltaRegistSustain = int16(eff.RegistSustain)
		buff.DeltaRegistFreeze = int16(eff.RegistFreeze)
		buff.DeltaRegistStun = int16(eff.RegistStun)
		buff.DeltaRegistStone = int16(eff.RegistStone)
		buff.DeltaRegistBlind = int16(eff.RegistBlind)
		buff.DeltaRegistSleep = int16(eff.RegistSleep)
		buff.DeltaMagicCritical = int16(eff.MagicCritical)
		buff.DeltaFireRes = int16(eff.FireRes)
		buff.DeltaWaterRes = int16(eff.WaterRes)
		buff.DeltaWindRes = int16(eff.WindRes)
		buff.DeltaEarthRes = int16(eff.EarthRes)
		if skill.SkillID == 147 {
			applyElementalProtectionDelta(target, buff)
		}

		// 套用屬性差值
		target.AC += buff.DeltaAC
		target.Str += buff.DeltaStr
		target.Dex += buff.DeltaDex
		target.Con += buff.DeltaCon
		target.Wis += buff.DeltaWis
		target.Intel += buff.DeltaIntel
		target.Cha += buff.DeltaCha
		target.MaxHP += buff.DeltaMaxHP
		target.MaxMP += buff.DeltaMaxMP
		target.HitMod += buff.DeltaHitMod
		target.DmgMod += buff.DeltaDmgMod
		target.SP += buff.DeltaSP
		target.MR += buff.DeltaMR
		target.HPR += buff.DeltaHPR
		target.MPR += buff.DeltaMPR
		target.BowHitMod += buff.DeltaBowHit
		target.BowDmgMod += buff.DeltaBowDmg
		target.Dodge += buff.DeltaDodge
		target.RegistSustain += buff.DeltaRegistSustain
		target.RegistFreeze += buff.DeltaRegistFreeze
		target.RegistStun += buff.DeltaRegistStun
		target.RegistStone += buff.DeltaRegistStone
		target.RegistBlind += buff.DeltaRegistBlind
		target.RegistSleep += buff.DeltaRegistSleep
		target.MagicCritical += buff.DeltaMagicCritical
		target.FireRes += buff.DeltaFireRes
		target.WaterRes += buff.DeltaWaterRes
		target.WindRes += buff.DeltaWindRes
		target.EarthRes += buff.DeltaEarthRes

		// 速度互抵邏輯
		if eff.MoveSpeed > 0 {
			if eff.MoveSpeed == 2 && target.MoveSpeed == 1 {
				s.cancelSpeedBuffs(target, 1)
				target.MoveSpeed = 0
				target.HasteTicks = 0
				s.sendSpeedToAll(target, 0, 0)
			} else if eff.MoveSpeed == 1 && target.MoveSpeed == 2 {
				s.cancelSpeedBuffs(target, 2)
				target.MoveSpeed = 0
				target.HasteTicks = 0
				s.sendSpeedToAll(target, 0, 0)
			} else {
				buff.SetMoveSpeed = byte(eff.MoveSpeed)
				target.MoveSpeed = byte(eff.MoveSpeed)
				target.HasteTicks = buff.TicksLeft
				s.sendSpeedToAll(target, byte(eff.MoveSpeed), uint16(skill.BuffDuration))
			}
		}
		if eff.BraveSpeed > 0 {
			buff.SetBraveSpeed = byte(eff.BraveSpeed)
			target.BraveSpeed = byte(eff.BraveSpeed)
			s.sendBraveToAll(target, byte(eff.BraveSpeed), uint16(skill.BuffDuration))
		}
		// Dodge 變化通知（龍之眼等）
		if buff.DeltaDodge > 0 {
			handler.SendDodgeIcon(target.Session, target.Dodge, true)
		}
		if eff.Invisible {
			buff.SetInvisible = true
			target.Invisible = true
		}
		if eff.Paralyzed {
			buff.SetParalyzed = true
			target.Paralyzed = true
			switch skill.SkillID {
			case 87, 508:
				handler.SendParalysis(target.Session, handler.StunApply)
			case 157, 50, 80, 30, 194:
				handler.SendParalysis(target.Session, handler.FreezeApply)
				// 凍結類：廣播灰色色調給附近所有玩家（Java: S_Poison type=2）
				broadcastPlayerPoison(target, 2, s.deps)
			case 192:
				handler.SendParalysis(target.Session, handler.BindApply)
			default:
				handler.SendParalysis(target.Session, handler.ParalysisApply)
			}
		}
		if eff.Sleeped {
			buff.SetSleeped = true
			target.Sleeped = true
			handler.SendParalysis(target.Session, handler.SleepApply)
		}
	}
	if skill.SkillID == 64 {
		buff.SetSilenced = true
		target.Silenced = true
	}

	// 註冊 buff（替換舊的）
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}

	// 屬性變化時發送更新
	if buff.DeltaStr != 0 || buff.DeltaDex != 0 || buff.DeltaCon != 0 ||
		buff.DeltaWis != 0 || buff.DeltaIntel != 0 || buff.DeltaCha != 0 ||
		buff.DeltaMaxHP != 0 || buff.DeltaMaxMP != 0 || buff.DeltaAC != 0 ||
		buff.DeltaDmgMod != 0 || buff.DeltaHitMod != 0 {
		handler.SendPlayerStatus(target.Session, target)
	}

	s.sendBuffIcon(target, skill.SkillID, uint16(skill.BuffDuration))

	// 日光術（技能 2）：上 buff 後更新光源
	if skill.SkillID == 2 {
		handler.UpdatePlayerLight(target, s.deps.World)
	}
}

func applyElementalProtectionDelta(target *world.PlayerInfo, buff *world.ActiveBuff) {
	if target == nil || buff == nil {
		return
	}
	switch target.ElfAttr {
	case 1:
		buff.DeltaEarthRes = 50
	case 2:
		buff.DeltaFireRes = 50
	case 4:
		buff.DeltaWaterRes = 50
	case 8:
		buff.DeltaWindRes = 50
	}
}

// ApplyNpcDebuff NPC 對玩家施放 debuff 技能（麻痺/睡眠/減速等）。
// 實際委派給 applyBuffEffect，由 NpcAISystem 透過 SkillManager 介面呼叫。
func (s *SkillSystem) ApplyNpcDebuff(target *world.PlayerInfo, skill *data.SkillInfo) {
	s.applyBuffEffect(target, skill)
}

// cancelAbsoluteBarrier 解除絕對屏障效果（Java: L1BuffUtil.cancelAbsoluteBarrier）。
// 被攻擊/施法/使用道具時呼叫。移動時不解除。
func (s *SkillSystem) cancelAbsoluteBarrier(player *world.PlayerInfo) {
	s.removeBuffAndRevert(player, 78)
	// removeBuffAndRevert → revertBuffStats 會清除 AbsoluteBarrier flag
}

// CancelAbsoluteBarrier 匯出版本，供 handler（movement/item）呼叫。
func (s *SkillSystem) CancelAbsoluteBarrier(player *world.PlayerInfo) {
	if player.AbsoluteBarrier {
		s.cancelAbsoluteBarrier(player)
	}
}

// cancelInvisibility 解除隱身效果（Java: L1BuffUtil.cancelInvisibility）。
// 攻擊/施法時呼叫。移除隱身 buff 並通知周圍玩家重新顯示此角色。
func (s *SkillSystem) cancelInvisibility(player *world.PlayerInfo) {
	// 移除隱身術 (60) 和暗隱術 (97) 的 buff
	s.removeBuffAndRevert(player, 60)
	s.removeBuffAndRevert(player, 97)
	// removeBuffAndRevert → revertBuffStats 會清除 Invisible flag

	// 通知玩家自己已解除隱身
	handler.SendInvisible(player.Session, player.CharID, false)

	// 通知周圍玩家重新顯示此角色（下一 tick VisibilitySystem 也會處理，
	// 但主動 SendPutObject 讓解除更即時）
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	for _, viewer := range nearby {
		if viewer.CharID != player.CharID {
			handler.SendPutObject(viewer.Session, player)
		}
	}
}

// CancelInvisibility 匯出版本，供 combat/handler 呼叫。
func (s *SkillSystem) CancelInvisibility(player *world.PlayerInfo) {
	if player.Invisible {
		s.cancelInvisibility(player)
	}
}

// ApplyGMBuff GM 強制套用 buff（繞過已學/MP/材料驗證）。
func (s *SkillSystem) ApplyGMBuff(player *world.PlayerInfo, skillID int32) bool {
	skill := s.deps.Skills.Get(skillID)
	if skill == nil {
		return false
	}
	s.applyBuffEffect(player, skill)
	dur := uint16(skill.BuffDuration)
	if dur == 0 {
		dur = 300 // 預設 5 分鐘
	}
	s.sendBuffIcon(player, skillID, dur)
	handler.SendPlayerStatus(player.Session, player)
	// 負重強化：套用時更新負重
	if skillID == 14 || skillID == 218 {
		handler.SendWeightUpdate(player.Session, player)
	}
	return true
}

// counterMagicExempt 魔法屏障不可抵擋的技能清單（Java: EXCEPT_COUNTER_MAGIC[]）。
// 這些技能穿透魔法屏障，不會被抵消。
var counterMagicExempt = map[int32]bool{
	1: true, 2: true, 3: true, 5: true, 8: true, 9: true, 12: true, 13: true, 14: true,
	19: true, 21: true, 26: true, 31: true, 32: true, 35: true, 37: true, 42: true,
	43: true, 44: true, 48: true, 49: true, 52: true, 54: true, 55: true, 57: true,
	60: true, 61: true, 63: true, 67: true, 68: true, 69: true, 72: true, 73: true,
	75: true, 78: true, 79: true, 87: true, 88: true, 89: true, 90: true, 91: true,
	97: true, 98: true, 99: true, 100: true, 101: true, 102: true, 104: true, 105: true,
	106: true, 107: true, 109: true, 110: true, 111: true, 113: true, 114: true, 115: true,
	116: true, 117: true, 118: true, 129: true, 130: true, 131: true, 132: true, 134: true,
	137: true, 138: true, 146: true, 147: true, 148: true, 149: true, 150: true, 151: true,
	155: true, 156: true, 158: true, 159: true, 161: true, 163: true, 164: true, 165: true,
	166: true, 168: true, 169: true, 170: true, 171: true, 175: true, 176: true, 181: true,
	185: true, 190: true, 194: true, 195: true, 201: true, 204: true, 209: true, 211: true,
	213: true, 214: true, 216: true, 219: true, 228: true, 230: true,
	10026: true, 10027: true, 10028: true, 10029: true, 41472: true,
}

// tryCounterMagic 檢查目標是否有魔法屏障（buff 31），若有則觸發抵消。
// 回傳 true 表示技能被抵消，呼叫方應跳過該目標的效果。
// Java 參考: L1SkillUse.isUseCounterMagic()
func (s *SkillSystem) tryCounterMagic(target *world.PlayerInfo, skillID int32) bool {
	// 豁免技能不受魔法屏障影響
	if counterMagicExempt[skillID] {
		return false
	}
	// 目標沒有魔法屏障
	if !target.HasBuff(31) {
		return false
	}
	// 觸發：移除魔法屏障 buff + 播放 GFX
	s.removeBuffAndRevert(target, 31)
	// 取得 castGfx2（魔法屏障觸發動畫）
	gfx := int32(10702) // 預設值
	if sk := s.deps.Skills.Get(31); sk != nil && sk.CastGfx2 > 0 {
		gfx = sk.CastGfx2
	}
	// 廣播觸發動畫給附近玩家 + 目標自己
	nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
	data := handler.BuildSkillEffect(target.CharID, gfx)
	handler.BroadcastToPlayers(nearby, data)
	return true
}

// removeBuffAndRevert 移除衝突 buff 並還原屬性。
func (s *SkillSystem) removeBuffAndRevert(target *world.PlayerInfo, skillID int32) {
	old := target.RemoveBuff(skillID)
	if old != nil {
		s.revertBuffStats(target, old)
		s.cancelBuffIcon(target, skillID)
		// 日光術（技能 2）：buff 被取消後更新光源
		if skillID == 2 {
			handler.UpdatePlayerLight(target, s.deps.World)
		}
		// 風之枷鎖（技能 167）：buff 被取消後清除客戶端效果
		if skillID == 167 {
			handler.SendWindShackle(target.Session, target.CharID, 0)
		}

		if s.deps.Skills != nil {
			if sk := s.deps.Skills.Get(skillID); sk != nil && sk.SysMsgStop > 0 {
				handler.SendServerMessage(target.Session, uint16(sk.SysMsgStop))
			}
		}
	}
}

// cancelSpeedBuffs 移除指定速度類型的所有 buff。
func (s *SkillSystem) cancelSpeedBuffs(target *world.PlayerInfo, speedType byte) {
	if target.ActiveBuffs == nil {
		return
	}
	for skillID, b := range target.ActiveBuffs {
		if b.SetMoveSpeed == speedType {
			s.revertBuffStats(target, b)
			delete(target.ActiveBuffs, skillID)
		}
	}
}

// revertBuffStats 還原 buff 的所有屬性修改。
func (s *SkillSystem) revertBuffStats(target *world.PlayerInfo, buff *world.ActiveBuff) {
	target.AC -= buff.DeltaAC
	target.Str -= buff.DeltaStr
	target.Dex -= buff.DeltaDex
	target.Con -= buff.DeltaCon
	target.Wis -= buff.DeltaWis
	target.Intel -= buff.DeltaIntel
	target.Cha -= buff.DeltaCha
	target.MaxHP -= buff.DeltaMaxHP
	target.MaxMP -= buff.DeltaMaxMP
	target.HitMod -= buff.DeltaHitMod
	target.DmgMod -= buff.DeltaDmgMod
	target.SP -= buff.DeltaSP
	target.MR -= buff.DeltaMR
	target.HPR -= buff.DeltaHPR
	target.MPR -= buff.DeltaMPR
	target.BowHitMod -= buff.DeltaBowHit
	target.BowDmgMod -= buff.DeltaBowDmg
	target.FireRes -= buff.DeltaFireRes
	target.WaterRes -= buff.DeltaWaterRes
	target.WindRes -= buff.DeltaWindRes
	target.EarthRes -= buff.DeltaEarthRes
	target.Dodge -= buff.DeltaDodge
	// Dodge 減少通知（龍之眼解除等）
	if buff.DeltaDodge > 0 && target.Session != nil {
		handler.SendDodgeIcon(target.Session, target.Dodge, false)
	}
	target.RegistSustain -= buff.DeltaRegistSustain
	target.RegistFreeze -= buff.DeltaRegistFreeze
	target.RegistStun -= buff.DeltaRegistStun
	target.RegistStone -= buff.DeltaRegistStone
	target.RegistBlind -= buff.DeltaRegistBlind
	target.RegistSleep -= buff.DeltaRegistSleep
	target.MagicCritical -= buff.DeltaMagicCritical
	if target.HP > target.MaxHP && target.MaxHP > 0 {
		target.HP = target.MaxHP
	}
	if target.MP > target.MaxMP && target.MaxMP > 0 {
		target.MP = target.MaxMP
	}
	if buff.SetInvisible {
		target.Invisible = false
	}
	if buff.SetParalyzed {
		target.Paralyzed = false
	}
	if buff.SetSleeped {
		target.Sleeped = false
	}
	if buff.SetSilenced {
		target.Silenced = false
	}
	if buff.SetAbsoluteBarrier {
		target.AbsoluteBarrier = false
	}
}

// RevertBuffStats implements handler.SkillManager — 還原 buff 的所有屬性修改（Exported）。
func (s *SkillSystem) RevertBuffStats(target *world.PlayerInfo, buff *world.ActiveBuff) {
	s.revertBuffStats(target, buff)
}

// ConsumeSkillResources implements handler.SkillManager — 扣除 MP/HP/材料並設定冷卻（Exported）。
func (s *SkillSystem) ConsumeSkillResources(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo) {
	s.consumeSkillResources(sess, player, skill)
}

// ApplyBuffStats implements handler.SkillManager — 套用 buff 屬性加成（靜默，不發送封包）。
func (s *SkillSystem) ApplyBuffStats(player *world.PlayerInfo, buff *world.ActiveBuff) {
	player.AC += buff.DeltaAC
	player.Str += buff.DeltaStr
	player.Dex += buff.DeltaDex
	player.Con += buff.DeltaCon
	player.Wis += buff.DeltaWis
	player.Intel += buff.DeltaIntel
	player.Cha += buff.DeltaCha
	player.MaxHP += buff.DeltaMaxHP
	player.MaxMP += buff.DeltaMaxMP
	player.HitMod += buff.DeltaHitMod
	player.DmgMod += buff.DeltaDmgMod
	player.SP += buff.DeltaSP
	player.MR += buff.DeltaMR
	player.HPR += buff.DeltaHPR
	player.MPR += buff.DeltaMPR
	player.BowHitMod += buff.DeltaBowHit
	player.BowDmgMod += buff.DeltaBowDmg
	player.Dodge += buff.DeltaDodge
	player.FireRes += buff.DeltaFireRes
	player.WaterRes += buff.DeltaWaterRes
	player.WindRes += buff.DeltaWindRes
	player.EarthRes += buff.DeltaEarthRes
	player.RegistSustain += buff.DeltaRegistSustain
	player.RegistFreeze += buff.DeltaRegistFreeze
	player.RegistStun += buff.DeltaRegistStun
	player.RegistStone += buff.DeltaRegistStone
	player.RegistBlind += buff.DeltaRegistBlind
	player.RegistSleep += buff.DeltaRegistSleep
	player.MagicCritical += buff.DeltaMagicCritical
}

// sendSpeedToAll 向自己和附近玩家發送速度封包。
func (s *SkillSystem) sendSpeedToAll(target *world.PlayerInfo, speedType byte, duration uint16) {
	sendSpeedPacket(target.Session, target.CharID, speedType, duration)
	nearby := s.deps.World.GetNearbyPlayers(target.X, target.Y, target.MapID, target.SessionID)
	for _, other := range nearby {
		sendSpeedPacket(other.Session, target.CharID, speedType, 0)
	}
}

// sendBraveToAll 向自己和附近玩家發送勇敢封包。
func (s *SkillSystem) sendBraveToAll(target *world.PlayerInfo, braveType byte, duration uint16) {
	sendBravePacket(target.Session, target.CharID, braveType, duration)
	nearby := s.deps.World.GetNearbyPlayers(target.X, target.Y, target.MapID, target.SessionID)
	for _, other := range nearby {
		sendBravePacket(other.Session, target.CharID, braveType, 0)
	}
}

// cancelAllBuffs 移除所有可取消的 buff。
func (s *SkillSystem) cancelAllBuffs(target *world.PlayerInfo) {
	if target.ActiveBuffs == nil {
		return
	}

	// 追蹤需要發送的客戶端通知（迴圈結束後統一發送）
	needFreezeRemove := false
	needStunRemove := false
	needParalysisRemove := false
	needSleepRemove := false
	needBindRemove := false
	needInvisRemove := false

	for skillID, buff := range target.ActiveBuffs {
		if s.deps.Scripting.IsNonCancellable(int(skillID)) {
			continue
		}
		s.revertBuffStats(target, buff)
		delete(target.ActiveBuffs, skillID)
		s.cancelBuffIcon(target, skillID)

		if skillID == handler.SkillShapeChange && s.deps.Polymorph != nil {
			s.deps.Polymorph.UndoPoly(target)
		}

		if buff.SetMoveSpeed > 0 {
			target.MoveSpeed = 0
			target.HasteTicks = 0
			s.sendSpeedToAll(target, 0, 0)
		}
		if buff.SetBraveSpeed > 0 {
			target.BraveSpeed = 0
			s.sendBraveToAll(target, 0, 0)
		}

		// 追蹤麻痺/凍結/暈眩類型
		if buff.SetParalyzed {
			switch skillID {
			case 87, 508:
				needStunRemove = true
			case 157, 50, 80, 30, 194:
				needFreezeRemove = true
			case 192:
				needBindRemove = true
			default:
				needParalysisRemove = true
			}
		}
		if buff.SetSleeped {
			needSleepRemove = true
		}
		if buff.SetInvisible {
			needInvisRemove = true
		}
	}

	// 凍結解除通知（控制鎖 + 灰色色調）
	if needFreezeRemove {
		handler.SendParalysis(target.Session, handler.FreezeRemove)
		broadcastPlayerPoison(target, 0, s.deps)
	}
	if needStunRemove {
		handler.SendParalysis(target.Session, handler.StunRemove)
	}
	if needParalysisRemove {
		handler.SendParalysis(target.Session, handler.ParalysisRemove)
	}
	// 睡眠解除通知
	if needSleepRemove {
		handler.SendParalysis(target.Session, handler.SleepRemove)
	}
	if needBindRemove {
		handler.SendParalysis(target.Session, handler.BindRemove)
	}
	// 隱身解除通知 + 周圍玩家重新顯示
	if needInvisRemove {
		handler.SendInvisible(target.Session, target.CharID, false)
		nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
		for _, viewer := range nearby {
			if viewer.CharID != target.CharID {
				handler.SendPutObject(viewer.Session, target)
			}
		}
	}

	// 重新檢查是否仍有非 buff 來源的麻痺（毒麻痺/詛咒麻痺）
	if shouldStayParalyzed(target, false, false) {
		target.Paralyzed = true
	}

	handler.SendPlayerStatus(target.Session, target)
}

// ========================================================================
//  Buff 計時器
// ========================================================================

// tickPlayerBuffs 每 tick 遞減 buff 計時器並處理到期。
func (s *SkillSystem) tickPlayerBuffs(p *world.PlayerInfo) {
	if p.ActiveBuffs == nil {
		return
	}
	for skillID, buff := range p.ActiveBuffs {
		if buff.TicksLeft <= 0 {
			continue
		}
		buff.TicksLeft--
		if buff.TicksLeft <= 0 {
			s.revertBuffStats(p, buff)
			delete(p.ActiveBuffs, skillID)

			s.cancelBuffIcon(p, skillID)

			if skillID == handler.SkillShapeChange && s.deps.Polymorph != nil {
				s.deps.Polymorph.UndoPoly(p)
			}

			if buff.SetMoveSpeed > 0 {
				p.MoveSpeed = 0
				p.HasteTicks = 0
				s.sendSpeedToAll(p, 0, 0)
			}
			if buff.SetBraveSpeed > 0 {
				p.BraveSpeed = 0
				p.BraveTicks = 0
				s.sendBraveToAll(p, 0, 0)
			}

			// 麻痺/睡眠/致盲到期
			if buff.SetParalyzed {
				switch skillID {
				case 87, 508:
					handler.SendParalysis(p.Session, handler.StunRemove)
				case 157, 50, 80, 30, 194:
					handler.SendParalysis(p.Session, handler.FreezeRemove)
					// 清除灰色色調
					broadcastPlayerPoison(p, 0, s.deps)
				case 192:
					handler.SendParalysis(p.Session, handler.BindRemove)
				default:
					handler.SendParalysis(p.Session, handler.ParalysisRemove)
				}
			}
			if buff.SetSleeped {
				handler.SendParalysis(p.Session, handler.SleepRemove)
			}
			if skillID == 20 || skillID == 40 {
				handler.SendCurseBlind(p.Session, 0)
			}

			// 慎重藥水到期
			if skillID == handler.SkillStatusWisdomPotion {
				p.WisdomSP = 0
				p.WisdomTicks = 0
			}

			// 負重強化到期：更新負重顯示
			if skillID == 14 || skillID == 218 {
				handler.SendWeightUpdate(p.Session, p)
			}

			// 日光術（技能 2）：到期後更新光源
			if skillID == 2 {
				handler.UpdatePlayerLight(p, s.deps.World)
			}

			// 風之枷鎖（技能 167）：到期後清除客戶端效果
			if skillID == 167 {
				handler.SendWindShackle(p.Session, p.CharID, 0)
			}

			// 法利昂覺醒（190）到期時連帶清除 Physical Power（169）
			if skillID == 190 {
				s.removeBuffAndRevert(p, 169)
			}

			if s.deps.Skills != nil {
				if sk := s.deps.Skills.Get(skillID); sk != nil && sk.SysMsgStop > 0 {
					handler.SendServerMessage(p.Session, uint16(sk.SysMsgStop))
				}
			}

			handler.SendPlayerStatus(p.Session, p)
		} else if buff.SetParalyzed && buff.TicksLeft%25 == 0 {
			// 3.80C 客戶端灰色色調會自動淡出，每 5 秒重發維持視覺
			switch skillID {
			case 157, 50, 80, 30, 194:
				broadcastPlayerPoison(p, 2, s.deps)
			}
		}
	}

	// 同步藥水倒數
	if p.HasteTicks > 0 {
		p.HasteTicks--
	}
	if p.BraveTicks > 0 {
		p.BraveTicks--
	}
	if p.WisdomTicks > 0 {
		p.WisdomTicks--
	}

	// PK 粉紅名到期
	if p.PinkNameTicks > 0 {
		p.PinkNameTicks--
		if p.PinkNameTicks <= 0 {
			p.PinkName = false
		}
	}

	// 通緝狀態到期
	if p.WantedTicks > 0 {
		p.WantedTicks--
	}
}

// ========================================================================
//  Buff 技能
// ========================================================================

// executeBuffSkill 處理治療與 buff 類技能。
func (s *SkillSystem) executeBuffSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, targetID int32, textArg ...string) {
	ws := s.deps.World
	text := ""
	if len(textArg) > 0 {
		text = textArg[0]
	}

	if skill.SkillID == 116 || skill.SkillID == 118 {
		// Owner: skill_clan.go
		s.executeClanTargetSkill(sess, player, skill, targetID, "", false)
		return
	}

	// 檢查目標是否為 NPC（debuff 路徑）
	if targetID != 0 && targetID != player.CharID {
		if npc := ws.GetNpc(targetID); npc != nil && !npc.Dead {
			s.executeNpcDebuffSkill(sess, player, skill, npc)
			return
		}
	}

	// 檢查目標是否為自己的寵物/召喚物
	// Java: TARGET_TO_PET — 治療/復活技能可對自己的寵物/召喚物生效
	if targetID != 0 && targetID != player.CharID {
		isResurrect := skill.SkillID == 61 || skill.SkillID == 75 // 返生術 / 終極返生術
		if pet := ws.GetPet(targetID); pet != nil && pet.OwnerCharID == player.CharID {
			if isResurrect && pet.Dead {
				s.resurrectPet(sess, player, skill, pet)
				return
			}
			if !isResurrect && !pet.Dead {
				s.healCompanion(sess, player, skill, pet.ID, &pet.HP, pet.MaxHP, pet.X, pet.Y, pet.MapID)
				return
			}
		}
		if sum := ws.GetSummon(targetID); sum != nil && sum.OwnerCharID == player.CharID {
			if !sum.Dead {
				s.healCompanion(sess, player, skill, sum.ID, &sum.HP, sum.MaxHP, sum.X, sum.Y, sum.MapID)
			}
			return
		}
	}

	target := player
	if targetID != 0 && targetID != player.CharID {
		if other := ws.GetByCharID(targetID); other != nil {
			if other.MapID != player.MapID || other.Dead {
				return
			}
			if chebyshevDist(player.X, player.Y, other.X, other.Y) > 20 {
				return
			}
			target = other
		}
	}

	// 魔法屏障攔截：對其他玩家施放非豁免技能時，檢查目標是否有 Counter Magic（buff 31）
	if target.CharID != player.CharID && s.tryCounterMagic(target, skill.SkillID) {
		// 技能被抵消，仍播放施法動畫但不產生效果
		nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
		return
	}

	// 玩家 debuff MR 抗性判定：對其他玩家施放 debuff 時必須通過 MR 檢查
	if target.CharID != player.CharID && playerDebuffSkills[skill.SkillID] {
		if !s.checkPlayerMRResist(player, target) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			// 仍播放施法動畫（Java: 即使 miss 也會播放動畫）
			nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
			handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
			if skill.CastGfx > 0 {
				handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
			}
			return
		}
	}

	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)

	// 廣播施法動畫
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))

	// 變形術：開啟怪物列表對話框
	if skill.SkillID == handler.SkillShapeChange {
		// Owner: skill_polymorph.go
		s.executeShapeChangePlayer(sess, player, target, skill)
		return
	}

	// 即時效果技能
	switch skill.SkillID {
	case 9: // 解毒術
		CurePoison(target, s.deps)
		return

	case 11: // 毒咒 — 對玩家施加傷害毒（Java: L1DamagePoison.doInfection(attacker, target, 3000, 5)）
		if target.CharID != player.CharID && target.PoisonType == 0 {
			target.PoisonType = 1
			target.PoisonTicksLeft = 150 // 30 秒 = 150 ticks
			target.PoisonDmgTimer = 0
			target.PoisonDmgAmount = 5 // 毒咒：每 3 秒扣 5 HP
			target.PoisonAttacker = sess.ID
			BroadcastPlayerPoison(target, 1, s.deps) // 綠色
		}

	case 23: // 能量感測 — 顯示目標的等級、HP/MP、屬性抗性等資訊
		if target.CharID != player.CharID {
			msg := fmt.Sprintf("\\f2【%s】 Lv.%d  HP:%d/%d  MP:%d/%d  AC:%d  MR:%d  火:%d 水:%d 風:%d 地:%d",
				target.Name, target.Level, target.HP, target.MaxHP, target.MP, target.MaxMP,
				target.AC, target.MR, target.FireRes, target.WaterRes, target.WindRes, target.EarthRes)
			handler.SendGlobalChat(sess, 9, msg)
		}

	case 27: // 壞物術 — Owner: skill_weapon.go；破壞玩家目標已裝備武器耐久
		weapon := target.Equip.Weapon()
		if applyWeaponBreakDurability(weapon, calcWeaponBreakDurabilityDamage(player)) {
			handler.SendServerMessageArgs(target.Session, 268, weapon.Name)
			handler.SendItemCountUpdate(target.Session, weapon)
		}

	case 20, 40: // 闇盲咒術 / 黑闇之影 — Owner: skill_status.go
		s.applyCurseBlindEffect(target, skill)
		return

	case 87: // 衝擊之暈 — Owner: skill_status.go；玩家目標需要雙手劍
		if s.applyShockStunToPlayer(sess, player, target, skill, nearby) {
			return
		}

	case 33: // 木乃伊詛咒 — 對玩家施加詛咒麻痺
		if target.CharID != player.CharID && !target.Paralyzed && target.CurseType == 0 &&
			!target.HasBuff(157) && !target.HasBuff(50) && !target.HasBuff(80) {
			target.CurseType = 1
			target.CurseTicksLeft = 25
			BroadcastPlayerPoison(target, 2, s.deps)
			handler.SendServerMessage(target.Session, 212)
		}

	case 37: // 聖潔之光 — 解毒 + 解詛咒 + 解麻痺/睡眠/致盲
		CurePoison(target, s.deps)
		if target.CurseType > 0 {
			CureCurseParalysis(target, s.deps)
		}
		if target.Paralyzed {
			target.Paralyzed = false
			handler.SendParalysis(target.Session, handler.ParalysisRemove)
		}
		if target.Sleeped {
			target.Sleeped = false
			handler.SendParalysis(target.Session, handler.SleepRemove)
		}
		s.removeCurseBlindEffect(target)
		return

	case 39: // 魔力奪取
		drain := int32(5 + world.RandInt(10))
		if target.MP >= drain {
			target.MP -= drain
			player.MP += drain
			if player.MP > player.MaxMP {
				player.MP = player.MaxMP
			}
			sendMpUpdate(target.Session, target)
			sendMpUpdate(sess, player)
		}

	case 44: // 魔法相消術 — 解除目標所有 buff + 毒 + 詛咒
		if target != nil {
			CurePoison(target, s.deps)
			CureCurseParalysis(target, s.deps)
			s.cancelAllBuffs(target)
		}
		// 施法時自身也解除隱身（Java: if srcpc.isInvisble() → srcpc.delInvis()）
		if player.Invisible {
			s.cancelInvisibility(player)
		}
		return

	case 71: // 藥水霜化術 — 通知目標無法使用藥水
		if target.CharID != player.CharID {
			handler.SendServerMessage(target.Session, 698) // "喉嚨灼熱，無法喝東西。"
		}

	case 103: // 暗黑盲咒 — Java 使用 66 睡眠效果與睡眠操作鎖
		sleepSkill := *skill
		sleepSkill.SkillID = 66
		s.applyBuffEffect(target, &sleepSkill)
		return

	case 113: // 精準目標 — Owner: skill_clan.go；目標狀態 + S_TrueTarget 給血盟成員
		s.applyTrueTargetEffect(player, target, skill, text)
		return

	case 133: // 弱化屬性 — Owner: skill_elemental.go
		if player.ElfAttr == 0 {
			handler.SendServerMessage(sess, 79)
			return
		}
		s.applyElementalFallDownToPlayer(player, target, skill)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
		}
		return

	case 157: // 大地屏障：Java EARTH_BIND，命中後 1-12 秒凍結。
		if target.HasBuff(157) || target.Paralyzed {
			return
		}
		earthBind := *skill
		earthBind.BuffDuration = 1 + world.RandInt(12)
		s.applyBuffEffect(target, &earthBind)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
		}
		return

	case 112: // 破壞盔甲（黑暗妖精 debuff）
		// Java: ARMOR_BREAK.java — 自定義機率系統，非標準 MR 判定
		if target.CharID == player.CharID {
			return // 不可對自己施放
		}
		if !s.calcArmorBreakProb(player, target) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}
		// 移除舊的破壞盔甲效果（Java: killSkillEffectTimer + 重新 setSkillEffect）
		if target.HasBuff(112) {
			s.removeBuffAndRevert(target, 112)
		}
		// 套用 buff 效果（8 秒計時器）
		s.applyBuffEffect(target, skill)
		// 廣播技能音效 GFX 3400
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, 3400))
		// Buff 圖示（Java: S_PacketBoxIconAura(119, 8)）
		handler.SendIconAura(target.Session, 119, 8)
		// 成功訊息
		handler.SendGlobalChat(sess, 9, "\\f2破壞盔甲 施放成功!")
		return

	case 153: // 魔法消除 — 解除 buff
		s.cancelAllBuffs(target)

	case 167: // 風之枷鎖 — 降低目標攻擊速度
		// Java: WIND_SHACKLE.java — 不可重複施加；發送 S_PacketBoxWindShackle
		if target.HasBuff(167) {
			return // 已有效果，不重複
		}
		s.applyBuffEffect(target, skill)
		handler.SendWindShackle(target.Session, target.CharID, skill.BuffDuration)
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
		return
	}

	// 覺醒系統（185/190/195）：再施放同一覺醒 → 解除；互斥由 buffs.lua exclusions 處理
	if skill.SkillID == 185 || skill.SkillID == 190 || skill.SkillID == 195 {
		if target.HasBuff(skill.SkillID) {
			// 再施放 → 解除覺醒（Java: toggle off）
			s.removeBuffAndRevert(target, skill.SkillID)
			// 法利昂（190）解除時連帶清除 Physical Power（169）
			if skill.SkillID == 190 {
				s.removeBuffAndRevert(target, 169)
			}
			// 播放解除 GFX
			if skill.CastGfx > 0 {
				handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
			}
			handler.SendPlayerStatus(target.Session, target)
			return
		}
		// 首次施放覺醒 → 走正常流程（exclusions 會先清除其他覺醒）
		s.applyBuffEffect(target, skill)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
		}
		// 法利昂（190）啟動時同時設定 Physical Power（169）timer
		if skill.SkillID == 190 {
			if sk169 := s.deps.Skills.Get(169); sk169 != nil {
				s.applyBuffEffect(target, sk169)
			}
		}
		handler.SendPlayerStatus(target.Session, target)
		if skill.SysMsgHappen > 0 {
			handler.SendServerMessage(target.Session, uint16(skill.SysMsgHappen))
		}
		return
	}

	if s.handleOppositeMoveSpeedSkill(target, skill.SkillID) {
		return
	}

	// 治療效果
	if skill.Type == 16 || skill.DamageValue > 0 || skill.DamageDice > 0 {
		casterINT := int(player.Intel)
		casterSP := int(player.SP)

		if skill.Area == -1 {
			// 範圍治療
			for _, p := range nearby {
				heal := int32(s.deps.Scripting.CalcHeal(skill.DamageValue, skill.DamageDice, skill.DamageDiceCount, casterINT, casterSP))
				heal = s.applyElfWaterHealingModifiers(p, heal)
				if heal > 0 && p.HP < p.MaxHP {
					p.HP += heal
					if p.HP > p.MaxHP {
						p.HP = p.MaxHP
					}
					sendHpUpdate(p.Session, p)
				}
			}
		} else {
			// 單目標治療
			heal := int32(s.deps.Scripting.CalcHeal(skill.DamageValue, skill.DamageDice, skill.DamageDiceCount, casterINT, casterSP))
			heal = s.applyElfWaterHealingModifiers(target, heal)
			if heal > 0 && target.HP < target.MaxHP {
				target.HP += heal
				if target.HP > target.MaxHP {
					target.HP = target.MaxHP
				}
				sendHpUpdate(target.Session, target)
			}
		}
	}

	// 套用 buff 效果
	s.applyBuffEffect(target, skill)

	// 效果 GFX
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, skill.CastGfx))
	}

	if skill.SysMsgHappen > 0 {
		handler.SendServerMessage(target.Session, uint16(skill.SysMsgHappen))
	}
}
