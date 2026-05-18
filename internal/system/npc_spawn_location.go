package system

import (
	"math/rand"

	"github.com/l1jgo/server/internal/world"
)

type npcSpawnMap interface {
	IsInMap(mapID int16, x, y int32) bool
	IsPassablePoint(mapID int16, x, y int32) bool
}

// NpcSpawnRule 保存 Java spawnlist 對生成點有影響的欄位。
type NpcSpawnRule struct {
	MapID         int16
	X             int32
	Y             int32
	Count         int
	RandomX       int32
	RandomY       int32
	LocX1         int32
	LocY1         int32
	LocX2         int32
	LocY2         int32
	AvoidPC       bool
	RespawnScreen bool
}

// FindNpcSpawnPoint 依 yiwei L1Spawn 的順序抽候選點，並額外保證不進牆、不重疊。
func FindNpcSpawnPoint(rule NpcSpawnRule, ws *world.State, maps npcSpawnMap, excludeID int32, rng *rand.Rand) (int32, int32, bool) {
	var fallbackX, fallbackY int32
	hasFallback := false
	for i := 0; i < 50; i++ {
		x, y := randomSpawnCandidate(rule, rng)
		if isValidNpcSpawnPoint(rule.MapID, x, y, ws, maps, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
			return x, y, true
		}
		if !hasFallback && isUsableNpcFallbackPoint(rule.MapID, x, y, ws, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
			fallbackX, fallbackY, hasFallback = x, y, true
		}
	}

	if isAreaSpawnRule(rule) {
		for x := rule.LocX1; x < rule.LocX2; x++ {
			for y := rule.LocY1; y < rule.LocY2; y++ {
				if isValidNpcSpawnPoint(rule.MapID, x, y, ws, maps, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
					return x, y, true
				}
				if !hasFallback && isUsableNpcFallbackPoint(rule.MapID, x, y, ws, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
					fallbackX, fallbackY, hasFallback = x, y, true
				}
			}
		}
	}

	maxRadius := spawnSearchRadius(rule)
	for radius := int32(1); radius <= maxRadius; radius++ {
		for _, pt := range ringCandidates(rule.X, rule.Y, radius) {
			if isValidNpcSpawnPoint(rule.MapID, pt[0], pt[1], ws, maps, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
				return pt[0], pt[1], true
			}
			if !hasFallback && isUsableNpcFallbackPoint(rule.MapID, pt[0], pt[1], ws, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
				fallbackX, fallbackY, hasFallback = pt[0], pt[1], true
			}
		}
	}

	if isValidNpcSpawnPoint(rule.MapID, rule.X, rule.Y, ws, maps, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
		return rule.X, rule.Y, true
	}
	if isUsableNpcFallbackPoint(rule.MapID, rule.X, rule.Y, ws, excludeID, rule.AvoidPC && !rule.RespawnScreen) {
		return rule.X, rule.Y, true
	}
	if hasFallback {
		return fallbackX, fallbackY, true
	}
	return rule.X, rule.Y, false
}

// NpcSpawnRuleFromNpc 以 NPC 保存的原始 spawnlist 規則建立重生選點規則。
func NpcSpawnRuleFromNpc(npc *world.NpcInfo) NpcSpawnRule {
	if npc == nil {
		return NpcSpawnRule{}
	}
	x, y := npc.SpawnBaseX, npc.SpawnBaseY
	if x == 0 && y == 0 {
		x, y = npc.SpawnX, npc.SpawnY
	}
	return NpcSpawnRule{
		MapID:         npc.SpawnMapID,
		X:             x,
		Y:             y,
		Count:         npc.SpawnCount,
		RandomX:       npc.SpawnRandomX,
		RandomY:       npc.SpawnRandomY,
		LocX1:         npc.SpawnLocX1,
		LocY1:         npc.SpawnLocY1,
		LocX2:         npc.SpawnLocX2,
		LocY2:         npc.SpawnLocY2,
		AvoidPC:       npc.SpawnAvoidPC,
		RespawnScreen: npc.SpawnScreen,
	}
}

// ApplyNpcSpawnRule 保存原始 spawn 規則，讓後續重生能重新抽點。
func ApplyNpcSpawnRule(npc *world.NpcInfo, rule NpcSpawnRule) {
	if npc == nil {
		return
	}
	npc.SpawnBaseX = rule.X
	npc.SpawnBaseY = rule.Y
	npc.SpawnMapID = rule.MapID
	npc.SpawnCount = rule.Count
	npc.SpawnRandomX = rule.RandomX
	npc.SpawnRandomY = rule.RandomY
	npc.SpawnLocX1 = rule.LocX1
	npc.SpawnLocY1 = rule.LocY1
	npc.SpawnLocX2 = rule.LocX2
	npc.SpawnLocY2 = rule.LocY2
	npc.SpawnAvoidPC = rule.AvoidPC
	npc.SpawnScreen = rule.RespawnScreen
}

func randomSpawnCandidate(rule NpcSpawnRule, rng *rand.Rand) (int32, int32) {
	if isAreaSpawnRule(rule) {
		return randomInHalfOpenRange(rule.LocX1, rule.LocX2, rng),
			randomInHalfOpenRange(rule.LocY1, rule.LocY2, rng)
	}

	x, y := rule.X, rule.Y
	if rule.RandomX > 0 || rule.RandomY > 0 {
		x += randomJavaDelta(rule.RandomX, rng)
		y += randomJavaDelta(rule.RandomY, rng)
		return x, y
	}

	if rule.Count > 1 {
		radius := int32(rule.Count * 6)
		if radius > 30 {
			radius = 30
		}
		x = randomInHalfOpenRange(rule.X-radius, rule.X+radius, rng)
		y = randomInHalfOpenRange(rule.Y-radius, rule.Y+radius, rng)
	}
	return x, y
}

func isAreaSpawnRule(rule NpcSpawnRule) bool {
	return rule.LocX1 != 0 && rule.LocY1 != 0 && rule.LocX2 != 0 && rule.LocY2 != 0 &&
		rule.LocX2 > rule.LocX1 && rule.LocY2 > rule.LocY1
}

func randomInHalfOpenRange(min, max int32, rng *rand.Rand) int32 {
	if max <= min {
		return min
	}
	return min + int32(randIntn(rng, int(max-min)))
}

func randomJavaDelta(limit int32, rng *rand.Rand) int32 {
	if limit <= 0 {
		return 0
	}
	return int32(randIntn(rng, int(limit))) - int32(randIntn(rng, int(limit)))
}

func randIntn(rng *rand.Rand, n int) int {
	if n <= 0 {
		return 0
	}
	if rng != nil {
		return rng.Intn(n)
	}
	return rand.Intn(n)
}

func isValidNpcSpawnPoint(mapID int16, x, y int32, ws *world.State, maps npcSpawnMap, excludeID int32, avoidPC bool) bool {
	if maps != nil {
		if !maps.IsInMap(mapID, x, y) || !maps.IsPassablePoint(mapID, x, y) {
			return false
		}
	}
	if ws != nil {
		if ws.IsOccupied(x, y, mapID, excludeID) {
			return false
		}
		if avoidPC && len(ws.GetNearbyPlayersAt(x, y, mapID)) > 0 {
			return false
		}
	}
	return true
}

func isUsableNpcFallbackPoint(mapID int16, x, y int32, ws *world.State, excludeID int32, avoidPC bool) bool {
	if x == 0 && y == 0 {
		return false
	}
	if ws == nil {
		return true
	}
	if ws.IsOccupied(x, y, mapID, excludeID) {
		return false
	}
	if avoidPC && len(ws.GetNearbyPlayersAt(x, y, mapID)) > 0 {
		return false
	}
	return true
}

func spawnSearchRadius(rule NpcSpawnRule) int32 {
	radius := max32(rule.RandomX, rule.RandomY)
	if isAreaSpawnRule(rule) {
		radius = max32(abs32(rule.LocX2-rule.LocX1), abs32(rule.LocY2-rule.LocY1))
	} else if rule.Count > 1 {
		radius = int32(rule.Count * 6)
		if radius > 30 {
			radius = 30
		}
	}
	if radius < 3 {
		return 3
	}
	if radius > 50 {
		return 50
	}
	return radius
}

func ringCandidates(cx, cy, radius int32) [][2]int32 {
	result := make([][2]int32, 0, radius*8)
	result = append(result,
		[2]int32{cx + radius, cy},
		[2]int32{cx - radius, cy},
		[2]int32{cx, cy + radius},
		[2]int32{cx, cy - radius},
	)
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if abs32(dx) != radius && abs32(dy) != radius {
				continue
			}
			if (dx == radius && dy == 0) || (dx == -radius && dy == 0) ||
				(dx == 0 && dy == radius) || (dx == 0 && dy == -radius) {
				continue
			}
			result = append(result, [2]int32{cx + dx, cy + dy})
		}
	}
	return result
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
