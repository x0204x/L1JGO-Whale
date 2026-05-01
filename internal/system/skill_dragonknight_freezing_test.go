package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillDragonKnightFreezingFreezingBreathFreezesAndRevealsPlayerTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     1,
		HP:        100,
		MaxHP:     100,
		Invisible: true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:          194,
		Target:           "attack",
		Type:             64,
		Ranged:           10,
		BuffDuration:     3,
		DamageValue:      1,
		ProbabilityValue: 100,
		ProbabilityDice:  100,
		ActionID:         18,
		CastGfx:          6988,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if !target.Paralyzed || !target.HasBuff(194) {
		t.Fatalf("寒冰噴吐應凍結玩家目標，Paralyzed=%v buff194=%v", target.Paralyzed, target.GetBuff(194))
	}
	if target.Invisible || target.HasBuff(60) {
		t.Fatalf("寒冰噴吐應揭示隱身目標，Invisible=%v buff60=%v", target.Invisible, target.GetBuff(60))
	}
}
