package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

const (
	skillFireWall       int32 = 58
	skillLifeStream     int32 = 63
	skillCubeIgnition   int32 = 205
	skillCubeQuake      int32 = 210
	skillCubeShock      int32 = 215
	skillCubeBalance    int32 = 220
	fireWallNpcID       int32 = 81157
	lifeStreamNpcID     int32 = 81169
	cubeIgnitionNpcID   int32 = 80149
	cubeQuakeNpcID      int32 = 80150
	cubeShockNpcID      int32 = 80151
	cubeBalanceNpcID    int32 = 80152
	groundEffectTickSec       = 5
)

func isGroundTargetSkill(skillID int32) bool {
	return skillID == skillFireWall || skillID == skillLifeStream || isCubeSkill(skillID)
}

func isCubeSkill(skillID int32) bool {
	switch skillID {
	case skillCubeIgnition, skillCubeQuake, skillCubeShock, skillCubeBalance:
		return true
	default:
		return false
	}
}

func (s *SkillSystem) executeGroundTargetSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo, targetX, targetY int32) {
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)
	if skill.SkillID == skillFireWall {
		player.Heading = CalcHeading(player.X, player.Y, targetX, targetY)
		handler.BroadcastToPlayers(nearby, handler.BuildChangeHeading(player.CharID, player.Heading))
	}
	if skill.ActionID > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	}

	switch skill.SkillID {
	case skillLifeStream:
		s.spawnGroundEffect(player, skill, lifeStreamNpcID, world.GroundEffectLifeStream, targetX, targetY)
	case skillFireWall:
		s.spawnFireWallEffects(player, skill, targetX, targetY)
	case skillCubeIgnition:
		s.spawnGroundEffect(player, skill, cubeIgnitionNpcID, world.GroundEffectCubeIgnition, player.X, player.Y)
	case skillCubeQuake:
		s.spawnGroundEffect(player, skill, cubeQuakeNpcID, world.GroundEffectCubeQuake, player.X, player.Y)
	case skillCubeShock:
		s.spawnGroundEffect(player, skill, cubeShockNpcID, world.GroundEffectCubeShock, player.X, player.Y)
	case skillCubeBalance:
		s.spawnGroundEffect(player, skill, cubeBalanceNpcID, world.GroundEffectCubeBalance, player.X, player.Y)
	}
}

func (s *SkillSystem) hasNearbySameCube(player *world.PlayerInfo, skillID int32) bool {
	cubeType := cubeEffectType(skillID)
	if cubeType == 0 {
		return false
	}
	for _, effect := range s.deps.World.GetNearbyGroundEffects(player.X, player.Y, player.MapID) {
		if effect.Type != cubeType {
			continue
		}
		if chebyshevDist(effect.X, effect.Y, player.X, player.Y) <= 3 {
			return true
		}
	}
	return false
}

func cubeEffectType(skillID int32) world.GroundEffectType {
	switch skillID {
	case skillCubeIgnition:
		return world.GroundEffectCubeIgnition
	case skillCubeQuake:
		return world.GroundEffectCubeQuake
	case skillCubeShock:
		return world.GroundEffectCubeShock
	case skillCubeBalance:
		return world.GroundEffectCubeBalance
	default:
		return 0
	}
}

func (s *SkillSystem) spawnFireWallEffects(player *world.PlayerInfo, skill *data.SkillInfo, targetX, targetY int32) {
	heading := CalcHeading(player.X, player.Y, targetX, targetY)
	dx := combatHeadingDX[heading]
	dy := combatHeadingDY[heading]
	x := player.X
	y := player.Y

	for i := 0; i < 8; i++ {
		if s.countOwnerFireWalls(player) >= 24 {
			return
		}
		if s.deps.MapData != nil && !s.deps.MapData.IsPassableIgnoreOccupant(player.MapID, x, y, int(heading)) {
			return
		}
		x += dx
		y += dy
		if s.deps.World.HasGroundEffectAt(x, y, player.MapID, fireWallNpcID) {
			continue
		}
		s.spawnGroundEffect(player, skill, fireWallNpcID, world.GroundEffectFireWall, x, y)
	}
}

func (s *SkillSystem) countOwnerFireWalls(player *world.PlayerInfo) int {
	count := 0
	for _, effect := range s.deps.World.GetNearbyGroundEffects(player.X, player.Y, player.MapID) {
		if effect.Type == world.GroundEffectFireWall && effect.OwnerCharID == player.CharID {
			count++
		}
	}
	return count
}

func (s *SkillSystem) spawnGroundEffect(player *world.PlayerInfo, skill *data.SkillInfo, npcID int32, typ world.GroundEffectType, x, y int32) {
	if s.deps.Npcs == nil {
		return
	}
	tpl := s.deps.Npcs.Get(npcID)
	if tpl == nil {
		return
	}
	effect := &world.GroundEffect{
		ID:           world.NextGroundEffectID(),
		SkillID:      skill.SkillID,
		NpcID:        npcID,
		GfxID:        tpl.GfxID,
		Type:         typ,
		X:            x,
		Y:            y,
		MapID:        player.MapID,
		OwnerCharID:  player.CharID,
		OwnerSession: player.SessionID,
		OwnerName:    player.Name,
		OwnerIntel:   player.Intel,
		OwnerClanID:  player.ClanID,
		TicksLeft:    skill.BuffDuration * groundEffectTickSec,
	}
	if effect.TicksLeft <= 0 {
		effect.TicksLeft = groundEffectTickSec
	}
	s.deps.World.AddGroundEffect(effect)
	s.broadcastGroundEffect(effect)
}

func (s *SkillSystem) broadcastGroundEffect(effect *world.GroundEffect) {
	nearby := s.deps.World.GetNearbyPlayersAt(effect.X, effect.Y, effect.MapID)
	for _, viewer := range nearby {
		handler.SendGroundEffectPack(viewer.Session, effect)
		if viewer.Known != nil {
			viewer.Known.GroundEffects[effect.ID] = world.KnownPos{X: effect.X, Y: effect.Y}
		}
	}
}
