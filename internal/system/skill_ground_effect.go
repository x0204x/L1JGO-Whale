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
	nearby := s.deps.World.GetNearbyPlayersInShow(player.X, player.Y, player.MapID, 0, player.ShowID)
	if skill.SkillID == skillFireWall {
		player.Heading = CalcHeading(player.X, player.Y, targetX, targetY)
		handler.BroadcastToPlayers(nearby, handler.BuildChangeHeading(player.CharID, player.Heading))
	}
	if skill.ActionID > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
	}

	switch skill.SkillID {
	case skillLifeStream:
		s.spawnGroundEffectWithSkillID(player, skill, lifeStreamNpcID, world.GroundEffectLifeStream, targetX, targetY, 0)
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
	for _, effect := range s.deps.World.GetNearbyGroundEffectsInShow(player.X, player.Y, player.MapID, player.ShowID) {
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
	baseX := player.X
	baseY := player.Y

	for i := 0; i < 8; i++ {
		if s.countOwnerFireWalls(player) >= 24 {
			return
		}
		heading := CalcHeading(baseX, baseY, targetX, targetY)
		if s.deps.MapData != nil && !s.deps.MapData.IsPassableIgnoreOccupant(player.MapID, baseX, baseY, int(heading)) {
			return
		}
		x := baseX + combatHeadingDX[heading]
		y := baseY + combatHeadingDY[heading]
		if s.deps.World.HasGroundEffectAtInShow(x, y, player.MapID, fireWallNpcID, player.ShowID) {
			continue
		}
		s.spawnGroundEffect(player, skill, fireWallNpcID, world.GroundEffectFireWall, x, y)
		baseX = x
		baseY = y
	}
}

func (s *SkillSystem) countOwnerFireWalls(player *world.PlayerInfo) int {
	count := 0
	for _, effect := range s.deps.World.GetNearbyGroundEffectsInShow(player.X, player.Y, player.MapID, player.ShowID) {
		if effect.Type == world.GroundEffectFireWall && effect.OwnerCharID == player.CharID {
			count++
		}
	}
	return count
}

func (s *SkillSystem) spawnGroundEffect(player *world.PlayerInfo, skill *data.SkillInfo, npcID int32, typ world.GroundEffectType, x, y int32) {
	s.spawnGroundEffectWithSkillID(player, skill, npcID, typ, x, y, skill.SkillID)
}

func (s *SkillSystem) spawnGroundEffectWithSkillID(player *world.PlayerInfo, skill *data.SkillInfo, npcID int32, typ world.GroundEffectType, x, y int32, effectSkillID int32) {
	if s.deps.Npcs == nil {
		return
	}
	tpl := s.deps.Npcs.Get(npcID)
	if tpl == nil {
		return
	}
	effect := &world.GroundEffect{
		ID:           world.NextGroundEffectID(),
		SkillID:      effectSkillID,
		NpcID:        npcID,
		GfxID:        tpl.GfxID,
		Type:         typ,
		X:            x,
		Y:            y,
		MapID:        player.MapID,
		ShowID:       player.ShowID,
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

func (s *SkillSystem) spawnGroundEffectFromNpc(caster *world.NpcInfo, skill *data.SkillInfo, npcID int32, typ world.GroundEffectType, x, y int32) {
	if s == nil || s.deps == nil || skill == nil {
		return
	}
	spawnNpcGroundEffectLikeJava(s.deps.World, s.deps.Npcs, caster, skill.SkillID, npcID, typ, x, y, skill.BuffDuration*groundEffectTickSec)
}

func (s *SkillSystem) broadcastGroundEffect(effect *world.GroundEffect) {
	nearby := s.deps.World.GetNearbyPlayersInShow(effect.X, effect.Y, effect.MapID, 0, effect.ShowID)
	broadcastGroundEffectToPlayers(nearby, effect)
}

func spawnNpcGroundEffectLikeJava(ws *world.State, npcs *data.NpcTable, caster *world.NpcInfo, skillID, npcID int32, typ world.GroundEffectType, x, y int32, ticksLeft int) (*world.GroundEffect, bool) {
	if ws == nil || npcs == nil || caster == nil || npcID == 0 {
		return nil, false
	}
	tpl := npcs.Get(npcID)
	if tpl == nil {
		return nil, false
	}
	if ticksLeft <= 0 {
		ticksLeft = groundEffectTickSec
	}
	effect := &world.GroundEffect{
		ID:          world.NextGroundEffectID(),
		SkillID:     skillID,
		NpcID:       npcID,
		GfxID:       tpl.GfxID,
		Type:        typ,
		X:           x,
		Y:           y,
		MapID:       caster.MapID,
		ShowID:      caster.ShowID,
		OwnerCharID: caster.ID,
		OwnerName:   caster.Name,
		TicksLeft:   ticksLeft,
	}
	ws.AddGroundEffect(effect)
	broadcastGroundEffectInShow(ws, effect, caster.ShowID)
	return effect, true
}

func broadcastGroundEffectInShow(ws *world.State, effect *world.GroundEffect, showID int32) {
	if ws == nil || effect == nil {
		return
	}
	nearby := ws.GetNearbyPlayersInShow(effect.X, effect.Y, effect.MapID, 0, showID)
	broadcastGroundEffectToPlayers(nearby, effect)
}

func broadcastGroundEffectToPlayers(nearby []*world.PlayerInfo, effect *world.GroundEffect) {
	if effect == nil {
		return
	}
	for _, viewer := range nearby {
		handler.SendGroundEffectPack(viewer.Session, effect)
		if viewer.Known != nil {
			viewer.Known.GroundEffects[effect.ID] = world.KnownPos{X: effect.X, Y: effect.Y}
		}
	}
}
