package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

const (
	skillConfusion  = int32(202)
	skillMindBreak  = int32(207)
	skillArmBreaker = int32(213)

	skillSilence = int32(64)
)

func calcMindBreakDamage(caster *world.PlayerInfo) int32 {
	if caster == nil {
		return 0
	}
	return int32(float64(caster.SP) * 3.8)
}

func applyMindBreakMPDrain(target *world.PlayerInfo) {
	if target == nil {
		return
	}
	target.MP -= 5
	if target.MP < 0 {
		target.MP = 0
	}
	sendMpUpdate(target.Session, target)
}

func (s *SkillSystem) applyIllusionistStatusAttackEffect(target *world.PlayerInfo, skill *data.SkillInfo) {
	if target == nil || skill == nil || target.Dead {
		return
	}
	switch skill.SkillID {
	case skillConfusion:
		s.applyConfusionSilence(target, skill)
	case skillArmBreaker:
		s.revealInvisibleTarget(target)
	}
}

func (s *SkillSystem) applyConfusionSilence(target *world.PlayerInfo, skill *data.SkillInfo) {
	if target.Silenced || target.HasBuff(skillSilence) {
		target.Silenced = true
		return
	}
	dur := skill.BuffDuration
	if dur <= 0 {
		dur = 8
	}
	buff := &world.ActiveBuff{
		SkillID:     skillSilence,
		TicksLeft:   dur * 5,
		SetSilenced: true,
	}
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}
	target.Silenced = true
}

func (s *SkillSystem) revealInvisibleTarget(target *world.PlayerInfo) {
	if !target.Invisible && !target.HasBuff(60) && !target.HasBuff(97) {
		return
	}
	s.removeBuffAndRevert(target, 60)
	s.removeBuffAndRevert(target, 97)
	target.Invisible = false
	handler.SendInvisible(target.Session, target.CharID, false)

	nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
	for _, viewer := range nearby {
		if viewer.CharID != target.CharID {
			handler.SendPutObject(viewer.Session, target)
		}
	}
}

const (
	skillBoneBreak = int32(208)
	skillJoyOfPain = int32(218)

	joyOfPainTicks   = 16 * 5
	joyOfPainDivisor = 5
	joyOfPainMaxDmg  = 1000

	boneBreakHigherLevelProb = 30
	boneBreakEqualLevelProb  = 20
	boneBreakLowerLevelProb  = 10
	boneBreakIntFactor       = 0.5
	boneBreakMRFactor        = 0.1
)

func (s *SkillSystem) applyIllusionistControlAttackEffect(caster, target *world.PlayerInfo, skill *data.SkillInfo) {
	if target == nil || skill == nil || target.Dead {
		return
	}
	if skill.SkillID == skillBoneBreak && checkBoneBreakPlayerSuccess(caster, target) {
		s.applyBoneBreakParalysis(target)
	}
}

func checkBoneBreakPlayerSuccess(caster, target *world.PlayerInfo) bool {
	return world.RandInt(100) < calcBoneBreakPlayerProbability(caster, target)
}

func calcBoneBreakPlayerProbability(caster, target *world.PlayerInfo) int {
	if caster == nil || target == nil {
		return 0
	}
	probability := boneBreakEqualLevelProb
	if caster.Level > target.Level {
		probability = boneBreakHigherLevelProb
	} else if caster.Level < target.Level {
		probability = boneBreakLowerLevelProb
	}
	probability += int(float64(caster.Intel) * boneBreakIntFactor)
	probability -= int(float64(target.MR) * boneBreakMRFactor)
	if probability < 0 {
		return 0
	}
	if probability > 100 {
		return 100
	}
	return probability
}

func (s *SkillSystem) applyBoneBreakParalysis(target *world.PlayerInfo) {
	if target.Paralyzed || target.HasBuff(skillBoneBreak) {
		return
	}
	durTicks := (world.RandInt(2) + 1) * 5
	buff := &world.ActiveBuff{
		SkillID:      skillBoneBreak,
		TicksLeft:    durTicks,
		SetParalyzed: true,
	}
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}
	target.Paralyzed = true
	handler.SendParalysis(target.Session, handler.ParalysisApply)
	nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, 13119))
}

func (s *SkillSystem) applyJoyOfPainReady(caster *world.PlayerInfo) {
	if caster == nil {
		return
	}
	if caster.HasBuff(skillJoyOfPain) {
		handler.SendSystemMessage(caster.Session, "已經準備疼痛的歡愉。")
		return
	}
	buff := &world.ActiveBuff{
		SkillID:   skillJoyOfPain,
		TicksLeft: joyOfPainTicks,
	}
	old := caster.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(caster, old)
	}
}

func (s *SkillSystem) applyJoyOfPainBacklash(attacker, target *world.PlayerInfo, nearby []*world.PlayerInfo) {
	if attacker == nil || target == nil || attacker.CharID == target.CharID || !attacker.HasBuff(skillJoyOfPain) {
		return
	}
	damage := (target.MaxHP - target.HP) / joyOfPainDivisor
	if damage <= 0 {
		return
	}
	if damage > joyOfPainMaxDmg {
		damage = joyOfPainMaxDmg
	}
	if attacker.HP-damage <= 0 {
		attacker.HP = 1
	} else {
		attacker.HP -= damage
	}
	attacker.Dirty = true
	s.removeBuffAndRevert(attacker, skillJoyOfPain)
	handler.SendWeightUpdate(attacker.Session, attacker)
	sendHpUpdate(attacker.Session, attacker)
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(attacker.CharID, 2))
}
