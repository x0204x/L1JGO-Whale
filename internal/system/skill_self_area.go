package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
)

func isSelfAreaAttackSkill(skill *data.SkillInfo) bool {
	return skill.Type == 64 && skill.Area > 0 && (skill.DamageValue > 0 || skill.DamageDice > 0)
}

func (s *SkillSystem) executeSelfAreaAttackSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, nearby []*world.PlayerInfo) {
	targets := make([]handler.RangeSkillTarget, 0, 8)
	for _, target := range nearby {
		if target.CharID == player.CharID || target.Dead {
			continue
		}
		if chebyshevDist(player.X, player.Y, target.X, target.Y) > int32(skill.Area) {
			continue
		}
		if s.tryCounterMagic(target, skill.SkillID) {
			continue
		}
		res := s.deps.Scripting.CalcSkillDamage(s.buildPlayerSkillDamageContext(player, target, skill))
		dmg := int32(res.Damage)
		targets = append(targets, handler.RangeSkillTarget{ObjectID: target.CharID, Hit: dmg > 0, Damage: dmg})
		s.applySelfAreaSkillDamageToPlayerNoVisual(sess, player, target, skill, dmg, nearby)
	}

	nearbyNpcs := s.deps.World.GetNearbyNpcs(player.X, player.Y, player.MapID)
	for _, npc := range nearbyNpcs {
		if npc.Dead {
			continue
		}
		if chebyshevDist(player.X, player.Y, npc.X, npc.Y) > int32(skill.Area) {
			continue
		}
		dmg := s.calcSelfAreaSkillNpcDamage(player, npc, skill)
		targets = append(targets, handler.RangeSkillTarget{ObjectID: npc.ID, Hit: dmg > 0, Damage: dmg})
		s.applySelfAreaSkillDamageToNpcNoVisual(sess, player, skill, npc, dmg, nearby)
	}

	if skill.CastGfx > 0 {
		data := handler.BuildRangeSkill(
			player.CharID,
			player.X,
			player.Y,
			player.Heading,
			skill.CastGfx,
			byte(skill.ActionID),
			handler.RangeSkillTypeNoDir,
			targets,
		)
		handler.BroadcastToPlayers(nearby, data)
	}
}

func (s *SkillSystem) applySelfAreaSkillDamageToPlayerNoVisual(sess *net.Session, player, target *world.PlayerInfo, skill *data.SkillInfo, dmg int32, nearby []*world.PlayerInfo) {
	if dmg < 0 {
		dmg = 0
	}
	if target.AbsoluteBarrier {
		dmg = 0
	}
	dmg = applyImmuneToHarmDamage(target, dmg)
	dmg = s.applyCounterMirrorMagicDamage(player, target, dmg, world.RandInt(100), nearby)

	if player.AttackView {
		handler.SendDamageNumbers(sess, target.CharID, dmg)
	}
	if dmg <= 0 {
		return
	}

	if target.Sleeped {
		breakPlayerSleepBySkill(target)
	}

	s.applyJoyOfPainBacklash(player, target, nearby)

	target.HP -= dmg
	target.Dirty = true
	if target.HP <= 0 {
		target.HP = 0
		if s.deps.Death != nil {
			s.deps.Death.KillPlayer(target)
		}
		return
	}
	sendHpUpdate(target.Session, target)
}

func (s *SkillSystem) calcSelfAreaSkillNpcDamage(player *world.PlayerInfo, npc *world.NpcInfo, skill *data.SkillInfo) int32 {
	ctx := scripting.SkillDamageContext{
		SkillID:            int(skill.SkillID),
		DamageValue:        skill.DamageValue,
		DamageDice:         skill.DamageDice,
		DamageDiceCount:    skill.DamageDiceCount,
		SkillLevel:         skill.SkillLevel,
		Attr:               skill.Attr,
		AttackerLevel:      int(player.Level),
		AttackerSTR:        int(player.Str),
		AttackerDEX:        int(player.Dex),
		AttackerINT:        int(player.Intel),
		AttackerWIS:        int(player.Wis),
		AttackerSP:         int(player.SP),
		AttackerDmgMod:     int(player.DmgMod),
		AttackerHitMod:     int(player.HitMod),
		AttackerMagicLevel: calcMagicLevel(int(player.ClassType), int(player.Level)),
		TargetAC:           int(npc.AC),
		TargetLevel:        int(npc.Level),
		TargetMR:           int(npc.MR),
		TargetFireRes:      int(npc.FireRes),
		TargetWaterRes:     int(npc.WaterRes),
		TargetWindRes:      int(npc.WindRes),
		TargetEarthRes:     int(npc.EarthRes),
	}
	res := s.deps.Scripting.CalcSkillDamage(ctx)
	return int32(res.Damage)
}

func (s *SkillSystem) applySelfAreaSkillDamageToNpcNoVisual(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, npc *world.NpcInfo, dmg int32, nearby []*world.PlayerInfo) {
	if dmg < 0 {
		dmg = 0
	}
	if player.AttackView {
		handler.SendDamageNumbers(sess, npc.ID, dmg)
	}
	npc.HP -= dmg
	if npc.HP < 0 {
		npc.HP = 0
	}
	AddHate(npc, sess.ID, dmg)
	hpRatio := int16(0)
	if npc.MaxHP > 0 {
		hpRatio = int16((npc.HP * 100) / npc.MaxHP)
	}
	handler.BroadcastToPlayers(nearby, handler.BuildHpMeter(npc.ID, hpRatio))
	if npc.HP <= 0 {
		handleNpcDeath(npc, player, nearby, s.deps)
		return
	}

	if skill.SkillID == 80 && !npc.Paralyzed && !npc.HasDebuff(50) && !npc.HasDebuff(80) {
		if s.checkNpcMRResist(player, npc, skill.SkillID) {
			dur := skill.BuffDuration
			if dur <= 0 {
				dur = 16
			}
			npc.Paralyzed = true
			npc.AddDebuff(80, (dur+1)*5)
			handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
		}
	}
}
