package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
)

const skillThunderGrab = int32(192)

func (s *SkillSystem) applyDragonKnightBindAttackEffect(caster, target *world.PlayerInfo, skill *data.SkillInfo) {
	if caster == nil || target == nil || skill == nil || target.Dead {
		return
	}
	if skill.SkillID == skillThunderGrab && checkDragonKnightDebuffSuccess(caster, target, skill) {
		s.applyThunderGrabBind(caster, target)
	}
}

func checkDragonKnightDebuffSuccess(caster, target *world.PlayerInfo, skill *data.SkillInfo) bool {
	return world.RandInt(100) < calcDragonKnightDebuffProbability(caster, target, skill)
}

func calcDragonKnightDebuffProbability(caster, target *world.PlayerInfo, skill *data.SkillInfo) int {
	if caster == nil || target == nil || skill == nil {
		return 0
	}
	levelDiff := int(caster.Level - target.Level)
	probability := int(float64(skill.ProbabilityDice)/10.0*float64(levelDiff)) + skill.ProbabilityValue
	if probability < 0 {
		return 0
	}
	if probability > 100 {
		return 100
	}
	return probability
}

func (s *SkillSystem) applyThunderGrabBind(caster, target *world.PlayerInfo) {
	if target.HasBuff(skillThunderGrab) {
		return
	}
	bindSeconds := world.RandInt(4) + 1
	buff := &world.ActiveBuff{
		SkillID:      skillThunderGrab,
		TicksLeft:    bindSeconds * 5,
		SetParalyzed: true,
	}
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}
	target.Paralyzed = true
	handler.SendParalysis(target.Session, handler.BindApply)

	nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, 4184))
}

const skillFreezingBreath = int32(194)

func (s *SkillSystem) applyDragonKnightFreezeAttackEffect(caster, target *world.PlayerInfo, skill *data.SkillInfo) {
	if caster == nil || target == nil || skill == nil || target.Dead {
		return
	}
	if skill.SkillID != skillFreezingBreath {
		return
	}
	if checkDragonKnightDebuffSuccess(caster, target, skill) {
		s.applyFreezingBreathFreeze(target, skill)
	}
	s.revealInvisibleTarget(target)
}

func (s *SkillSystem) applyFreezingBreathFreeze(target *world.PlayerInfo, skill *data.SkillInfo) {
	if target.Paralyzed || target.HasBuff(50) || target.HasBuff(80) || target.HasBuff(30) || target.HasBuff(157) || target.HasBuff(skillFreezingBreath) {
		return
	}
	dur := skill.BuffDuration + 1
	if dur <= 0 {
		dur = 4
	}
	buff := &world.ActiveBuff{
		SkillID:      skillFreezingBreath,
		TicksLeft:    dur * 5,
		SetParalyzed: true,
	}
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}
	target.Paralyzed = true
	handler.SendParalysis(target.Session, handler.FreezeApply)
	broadcastPlayerPoison(target, 2, s.deps)

	nearby := s.deps.World.GetNearbyPlayersAt(target.X, target.Y, target.MapID)
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, 81168))
}

const (
	skillFoeSlayer     = int32(187)
	skillCopyShockStun = int32(508)

	foeSlayerHitCount        = 3
	foeSlayerDefaultBonusMax = 10
	foeSlayerDefaultStunRate = 15
	foeSlayerDefaultStunSec  = 30
)

func (s *SkillSystem) executeFoeSlayerOnPlayer(sess *net.Session, caster *world.PlayerInfo, skill *data.SkillInfo, target *world.PlayerInfo, nearby []*world.PlayerInfo) {
	if caster == nil || target == nil || skill == nil || target.Dead {
		return
	}
	defer clearDragonKnightWeakness(caster)
	for range foeSlayerHitCount {
		damage := s.calcFoeSlayerPlayerHitDamage(caster, target)
		s.broadcastFoeSlayerAttack(caster, target.CharID, damage, nearby)
		s.applyFoeSlayerPlayerDamage(sess, caster, target, damage)
		if target.Dead || target.HP <= 0 {
			return
		}
	}
	bonus := foeSlayerRandomBonus(skill)
	if bonus > 0 {
		s.applyFoeSlayerPlayerDamage(sess, caster, target, bonus)
	}
	s.broadcastFoeSlayerEffects(caster.CharID, target.CharID, nearby)
	s.applyFoeSlayerPlayerStun(target, skill, nearby)
}

func (s *SkillSystem) executeFoeSlayerOnNpc(sess *net.Session, caster *world.PlayerInfo, skill *data.SkillInfo, npc *world.NpcInfo, nearby []*world.PlayerInfo) {
	if caster == nil || skill == nil || npc == nil || npc.Dead {
		return
	}
	defer clearDragonKnightWeakness(caster)
	for range foeSlayerHitCount {
		damage := s.calcFoeSlayerNpcHitDamage(caster, npc)
		s.broadcastFoeSlayerAttack(caster, npc.ID, damage, nearby)
		s.applyFoeSlayerNpcDamage(sess, caster, npc, damage, nearby)
		if npc.Dead || npc.HP <= 0 {
			return
		}
	}
	bonus := foeSlayerRandomBonus(skill)
	if bonus > 0 {
		s.applyFoeSlayerNpcDamage(sess, caster, npc, bonus, nearby)
	}
	s.broadcastFoeSlayerEffects(caster.CharID, npc.ID, nearby)
	s.applyFoeSlayerNpcStun(npc, skill, nearby)
}

func (s *SkillSystem) calcFoeSlayerPlayerHitDamage(caster, target *world.PlayerInfo) int32 {
	if s.deps == nil || s.deps.Scripting == nil || target.AbsoluteBarrier {
		return 0
	}
	result := s.deps.Scripting.CalcMeleeAttack(scripting.CombatContext{
		AttackerLevel:  int(caster.Level),
		AttackerSTR:    int(caster.Str),
		AttackerDEX:    int(caster.Dex),
		AttackerWeapon: s.foeSlayerWeaponDamage(caster, "small"),
		AttackerHitMod: int(caster.HitMod),
		AttackerDmgMod: int(caster.DmgMod),
		TargetAC:       int(target.AC),
		TargetLevel:    int(target.Level),
		TargetMR:       0,
	})
	if !result.IsHit || result.Damage <= 0 {
		return 0
	}
	damage := int32(result.Damage)
	if target.HasBuff(112) {
		damage = int32(float64(damage) * 1.58)
	}
	damage += dragonKnightWeaknessFoeSlayerBonus(caster, target.CharID)
	return applyImmuneToHarmDamage(target, damage)
}

func (s *SkillSystem) calcFoeSlayerNpcHitDamage(caster *world.PlayerInfo, npc *world.NpcInfo) int32 {
	if s.deps == nil || s.deps.Scripting == nil {
		return 0
	}
	targetSize := npc.Size
	if targetSize == "" {
		targetSize = "small"
	}
	result := s.deps.Scripting.CalcMeleeAttack(scripting.CombatContext{
		AttackerLevel:   int(caster.Level),
		AttackerSTR:     int(caster.Str),
		AttackerDEX:     int(caster.Dex),
		AttackerWeapon:  s.foeSlayerWeaponDamage(caster, targetSize),
		AttackerHitMod:  int(caster.HitMod),
		AttackerDmgMod:  int(caster.DmgMod),
		TargetAC:        int(npc.AC),
		TargetLevel:     int(npc.Level),
		TargetMR:        int(npc.MR),
		TargetClassType: -1,
	})
	if !result.IsHit || result.Damage <= 0 {
		return 0
	}
	damage := int32(result.Damage)
	if npc.HasDebuff(112) {
		damage = int32(float64(damage) * 1.58)
	}
	damage += dragonKnightWeaknessFoeSlayerBonus(caster, npc.ID)
	return damage
}

func (s *SkillSystem) foeSlayerWeaponDamage(caster *world.PlayerInfo, targetSize string) int {
	weaponDmg := 4
	if caster == nil || caster.Equip.Weapon() == nil || s.deps == nil || s.deps.Items == nil {
		return weaponDmg
	}
	info := s.deps.Items.Get(caster.Equip.Weapon().ItemID)
	if info == nil {
		return weaponDmg
	}
	if targetSize == "large" && info.DmgLarge > 0 {
		return info.DmgLarge
	}
	if info.DmgSmall > 0 {
		return info.DmgSmall
	}
	return weaponDmg
}

func (s *SkillSystem) applyFoeSlayerPlayerDamage(sess *net.Session, caster, target *world.PlayerInfo, damage int32) {
	if damage <= 0 || target.Dead {
		return
	}
	if target.Sleeped {
		breakPlayerSleepBySkill(target)
	}
	target.HP -= damage
	target.Dirty = true
	if target.HP <= 0 {
		target.HP = 0
		if s.deps.Death != nil {
			s.deps.Death.KillPlayer(target)
		}
		return
	}
	sendHpUpdate(target.Session, target)
	if caster.AttackView {
		handler.SendDamageNumbers(sess, target.CharID, damage)
	}
}

func (s *SkillSystem) applyFoeSlayerNpcDamage(sess *net.Session, caster *world.PlayerInfo, npc *world.NpcInfo, damage int32, nearby []*world.PlayerInfo) {
	if damage <= 0 || npc.Dead {
		return
	}
	npc.HP -= damage
	if npc.HP < 0 {
		npc.HP = 0
	}
	if npc.Sleeped {
		BreakNpcSleep(npc, s.deps.World)
	}
	AddHate(npc, sess.ID, damage)
	hpRatio := int16(0)
	if npc.MaxHP > 0 {
		hpRatio = int16((npc.HP * 100) / npc.MaxHP)
	}
	handler.BroadcastToPlayers(nearby, handler.BuildHpMeter(npc.ID, hpRatio))
	if caster.AttackView {
		handler.SendDamageNumbers(sess, npc.ID, damage)
	}
	if npc.HP <= 0 {
		handleNpcDeath(npc, caster, nearby, s.deps)
	}
}

func (s *SkillSystem) applyFoeSlayerPlayerStun(target *world.PlayerInfo, skill *data.SkillInfo, nearby []*world.PlayerInfo) {
	if target.HasBuff(skillCopyShockStun) || !foeSlayerStunSuccess(skill) {
		return
	}
	buff := &world.ActiveBuff{
		SkillID:      skillCopyShockStun,
		TicksLeft:    foeSlayerStunSeconds(skill) * 5,
		SetParalyzed: true,
	}
	old := target.AddBuff(buff)
	if old != nil {
		s.revertBuffStats(target, old)
	}
	target.Paralyzed = true
	handler.SendParalysis(target.Session, handler.StunApply)
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(target.CharID, 81162))
}

func (s *SkillSystem) applyFoeSlayerNpcStun(npc *world.NpcInfo, skill *data.SkillInfo, nearby []*world.PlayerInfo) {
	if npc.HasDebuff(skillCopyShockStun) || !foeSlayerStunSuccess(skill) {
		return
	}
	npc.Paralyzed = true
	npc.AddDebuff(skillCopyShockStun, foeSlayerStunSeconds(skill)*5)
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, 81162))
}

func (s *SkillSystem) broadcastFoeSlayerAttack(caster *world.PlayerInfo, targetID int32, damage int32, nearby []*world.PlayerInfo) {
	for _, viewer := range nearby {
		handler.SendAttackPacket(viewer.Session, caster.CharID, targetID, damage, caster.Heading)
	}
}

func (s *SkillSystem) broadcastFoeSlayerEffects(casterID, targetID int32, nearby []*world.PlayerInfo) {
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(casterID, 7020))
	handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(targetID, 12119))
}

func foeSlayerRandomBonus(skill *data.SkillInfo) int32 {
	maxBonus := foeSlayerDefaultBonusMax
	if skill != nil {
		if skill.DamageDice > 0 {
			maxBonus = skill.DamageDice
		} else if skill.DamageValue > 0 {
			maxBonus = int(skill.DamageValue)
		}
	}
	if maxBonus <= 0 {
		return 0
	}
	return int32(world.RandInt(maxBonus) + 1)
}

func foeSlayerStunSuccess(skill *data.SkillInfo) bool {
	chance := foeSlayerDefaultStunRate
	if skill != nil && skill.ProbabilityValue > 0 {
		chance = skill.ProbabilityValue
	}
	if chance <= 0 {
		return false
	}
	if chance > 100 {
		chance = 100
	}
	return world.RandInt(100) < chance
}

func foeSlayerStunSeconds(skill *data.SkillInfo) int {
	if skill != nil && skill.BuffDuration > 0 {
		return skill.BuffDuration
	}
	return foeSlayerDefaultStunSec
}

const (
	dragonKnightClassType = int16(5)
	chainSwordType        = "chainsword"

	dragonKnightWeaknessBaseChance  = 15
	dragonKnightWeaknessOtherChance = 10
	dragonKnightWeaknessOtherItemID = int32(410189)
)

var dragonKnightWeaknessRoll = world.RandInt

func applyDragonKnightWeaknessExposure(player *world.PlayerInfo, targetID int32, weaponItemID int32, weaponType string) {
	if player == nil || player.ClassType != dragonKnightClassType || weaponType != chainSwordType {
		return
	}
	if player.WeaknessTargetID != targetID {
		clearDragonKnightWeakness(player)
		player.WeaknessTargetID = targetID
	}

	chance := dragonKnightWeaknessBaseChance
	if weaponItemID == dragonKnightWeaknessOtherItemID {
		chance += dragonKnightWeaknessOtherChance
	}
	roll := dragonKnightWeaknessRoll(100)
	switch player.WeaknessLevel {
	case 0:
		if roll < chance {
			setDragonKnightWeakness(player, 1)
		}
	case 1:
		if roll < chance {
			setDragonKnightWeakness(player, 1)
		} else if roll < chance*2 {
			setDragonKnightWeakness(player, 2)
		}
	case 2:
		if roll < chance {
			setDragonKnightWeakness(player, 2)
		} else if roll < chance*2 {
			setDragonKnightWeakness(player, 3)
		}
	case 3:
		if roll < chance {
			setDragonKnightWeakness(player, 3)
		}
	}
}

func setDragonKnightWeakness(player *world.PlayerInfo, level int16) {
	player.WeaknessLevel = level
	handler.SendPacketBoxDk(player.Session, level)
}

func clearDragonKnightWeakness(player *world.PlayerInfo) {
	if player == nil {
		return
	}
	player.WeaknessLevel = 0
	player.WeaknessTargetID = 0
	handler.SendPacketBoxDk(player.Session, 0)
}

func dragonKnightWeaknessFoeSlayerBonus(player *world.PlayerInfo, targetID int32) int32 {
	if player == nil || player.WeaknessTargetID != targetID {
		return 0
	}
	switch player.WeaknessLevel {
	case 1:
		return 20 + player.FoeSlayerBonusDmg
	case 2:
		return 40 + player.FoeSlayerBonusDmg
	case 3:
		return 60 + player.FoeSlayerBonusDmg
	default:
		return 0
	}
}

func (s *CombatSystem) applyDragonKnightWeaknessFromMelee(player *world.PlayerInfo, targetID int32) {
	itemID, weaponType, ok := equippedWeaponForWeakness(s.deps, player)
	if !ok {
		return
	}
	applyDragonKnightWeaknessExposure(player, targetID, itemID, weaponType)
}

func (s *PvPSystem) applyDragonKnightWeaknessFromMelee(player *world.PlayerInfo, targetID int32) {
	itemID, weaponType, ok := equippedWeaponForWeakness(s.deps, player)
	if !ok {
		return
	}
	applyDragonKnightWeaknessExposure(player, targetID, itemID, weaponType)
}

func equippedWeaponForWeakness(deps *handler.Deps, player *world.PlayerInfo) (int32, string, bool) {
	if deps == nil || deps.Items == nil || player == nil || player.Equip.Weapon() == nil {
		return 0, "", false
	}
	weapon := player.Equip.Weapon()
	info := deps.Items.Get(weapon.ItemID)
	if info == nil {
		return 0, "", false
	}
	return weapon.ItemID, info.Type, true
}
