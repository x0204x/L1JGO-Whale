package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

const (
	targetToClan  = 4
	targetToParty = 8
)

func (s *SkillSystem) applyRoyalAuraSkill(player *world.PlayerInfo, skill *data.SkillInfo, nearby []*world.PlayerInfo) bool {
	if player == nil || skill == nil {
		return false
	}
	switch skill.SkillID {
	case 114, 115, 117:
	default:
		return false
	}

	targets := s.royalAuraTargets(player, skill.TargetTo, nearby)
	for _, target := range targets {
		s.applyBuffEffect(target, skill)
	}
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}
	if skill.SysMsgHappen > 0 {
		handler.SendServerMessage(player.Session, uint16(skill.SysMsgHappen))
	}
	return true
}

func (s *SkillSystem) royalAuraTargets(player *world.PlayerInfo, targetTo int, nearby []*world.PlayerInfo) []*world.PlayerInfo {
	targets := []*world.PlayerInfo{player}
	seen := map[int32]bool{player.CharID: true}
	for _, other := range nearby {
		if other == nil || other.Dead || seen[other.CharID] {
			continue
		}
		if targetTo&targetToParty != 0 && sameParty(s.deps.World, player, other) {
			targets = append(targets, other)
			seen[other.CharID] = true
			continue
		}
		if targetTo&targetToClan != 0 && player.ClanID != 0 && player.ClanID == other.ClanID {
			targets = append(targets, other)
			seen[other.CharID] = true
		}
	}
	return targets
}

func sameParty(ws *world.State, a, b *world.PlayerInfo) bool {
	if ws == nil || a == nil || b == nil {
		return false
	}
	party := ws.Parties.GetParty(a.CharID)
	if party == nil {
		return false
	}
	for _, memberID := range party.Members {
		if memberID == b.CharID {
			return true
		}
	}
	return false
}

func braveAuraDamage(attacker *world.PlayerInfo, damage int32) int32 {
	return braveAuraDamageWithRoll(attacker, damage, world.RandInt(100))
}

func braveAuraDamageWithRoll(attacker *world.PlayerInfo, damage int32, roll int) int32 {
	if attacker == nil || damage <= 0 || !attacker.HasBuff(117) || roll >= 33 {
		return damage
	}
	return damage * 3 / 2
}

func (s *SkillSystem) applyTrueTargetEffect(caster, target *world.PlayerInfo, skill *data.SkillInfo, text string) {
	if caster == nil || target == nil || skill == nil {
		return
	}
	if !target.HasBuff(113) {
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 5
		}
		target.AddBuff(&world.ActiveBuff{
			SkillID:   113,
			TicksLeft: dur * 5,
		})
	}
	s.sendTrueTargetToClan(caster, target, text)
}

func (s *SkillSystem) sendTrueTargetToClan(caster, target *world.PlayerInfo, text string) {
	if caster.ClanID == 0 || s.deps == nil || s.deps.World == nil {
		handler.SendTrueTarget(caster.Session, target.CharID, caster.CharID, text)
		return
	}
	sent := false
	s.deps.World.AllPlayers(func(player *world.PlayerInfo) {
		if player != nil && player.ClanID == caster.ClanID {
			handler.SendTrueTarget(player.Session, target.CharID, caster.CharID, text)
			sent = true
		}
	})
	if !sent {
		handler.SendTrueTarget(caster.Session, target.CharID, caster.CharID, text)
	}
}

func (s *SkillSystem) executeClanTargetSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, targetID int32, targetName string, consume bool) {
	target := s.resolveClanSkillTarget(targetID, targetName)
	if target == nil {
		if targetName != "" {
			handler.SendServerMessageArgs(sess, 73, targetName)
			return
		}
		handler.SendServerMessage(sess, skillMsgCastFail)
		return
	}
	if target.CharID == player.CharID || player.ClanID == 0 || player.ClanID != target.ClanID {
		handler.SendServerMessage(sess, 414)
		return
	}

	switch skill.SkillID {
	case 116:
		if consume {
			s.consumeSkillResources(sess, player, skill)
		}
		nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
		target.PendingYesNoType = 729
		target.PendingYesNoData = player.CharID
		handler.SendYesNoDialog(target.Session, 729)

	case 118:
		if !s.canRunClanTeleport(player, target) {
			handler.SendServerMessage(sess, s.runClanRejectMessage(player, target))
			handler.SendParalysis(sess, handler.TeleportUnlock)
			return
		}
		if consume {
			s.consumeSkillResources(sess, player, skill)
		}
		nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
		handler.TeleportPlayer(sess, player, target.X, target.Y, target.MapID, 5, s.deps)
	}
}

func (s *SkillSystem) resolveClanSkillTarget(targetID int32, targetName string) *world.PlayerInfo {
	if s == nil || s.deps == nil || s.deps.World == nil {
		return nil
	}
	if targetID != 0 {
		if target := s.deps.World.GetByCharID(targetID); target != nil {
			return target
		}
	}
	if targetName != "" {
		return s.deps.World.GetByName(targetName)
	}
	return nil
}

func (s *SkillSystem) canRunClanTeleport(player, target *world.PlayerInfo) bool {
	if !s.isEscapableForRunClan(player) {
		return false
	}
	if !isRunClanAllowedTargetMap(target.MapID) {
		return false
	}
	return !s.isInAnyCastleWarArea(target.X, target.Y, target.MapID)
}

func (s *SkillSystem) runClanRejectMessage(player, target *world.PlayerInfo) uint16 {
	if !s.isEscapableForRunClan(player) {
		return 647
	}
	if !isRunClanAllowedTargetMap(target.MapID) || s.isInAnyCastleWarArea(target.X, target.Y, target.MapID) {
		return 1192
	}
	return skillMsgCastFail
}

func (s *SkillSystem) isEscapableForRunClan(player *world.PlayerInfo) bool {
	if player.AccessLevel >= 200 {
		return true
	}
	if s.deps == nil || s.deps.MapData == nil {
		return true
	}
	if mi := s.deps.MapData.GetInfo(player.MapID); mi != nil {
		return mi.Escapable
	}
	return true
}

func isRunClanAllowedTargetMap(mapID int16) bool {
	return mapID == 0 || mapID == 4 || mapID == 304
}

func (s *SkillSystem) isInAnyCastleWarArea(x, y int32, mapID int16) bool {
	return s.deps != nil && s.deps.Castles != nil && s.deps.Castles.GetCastleIDByArea(x, y, mapID) != 0
}
