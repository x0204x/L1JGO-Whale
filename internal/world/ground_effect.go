package world

import "sync/atomic"

type GroundEffectType byte

const (
	GroundEffectFireWall       GroundEffectType = 1
	GroundEffectLifeStream     GroundEffectType = 2
	GroundEffectCubeIgnition   GroundEffectType = 3
	GroundEffectCubeQuake      GroundEffectType = 4
	GroundEffectCubeShock      GroundEffectType = 5
	GroundEffectCubeBalance    GroundEffectType = 6
	GroundEffectTomb           GroundEffectType = 7
	GroundEffectShockStun      GroundEffectType = 8
	GroundEffectThunderGrab    GroundEffectType = 9
	GroundEffectFreezingBreath GroundEffectType = 10
	GroundEffectPoisonCloud    GroundEffectType = 11
	GroundEffectNpcEffect      GroundEffectType = 12
)

const (
	TombEffectNpcID int32 = 86126
	TombEffectGfxID int32 = 13600
)

var groundEffectIDCounter atomic.Int32

func init() {
	groundEffectIDCounter.Store(300_000_000)
}

func NextGroundEffectID() int32 {
	return groundEffectIDCounter.Add(1)
}

type GroundEffect struct {
	ID            int32
	SkillID       int32
	NpcID         int32
	GfxID         int32
	Type          GroundEffectType
	X             int32
	Y             int32
	MapID         int16
	ShowID        int32
	Heading       int16
	LightSize     byte
	OwnerCharID   int32
	OwnerSession  uint64
	OwnerName     string
	OwnerIntel    int16
	OwnerClanID   int32
	Lawful        int32
	TicksLeft     int
	DamageTickAcc int
}
