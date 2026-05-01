package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillIllusionistControlBoneBreakParalyzesPlayerTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     50,
		Intel:     300,
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
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:     208,
		Target:      "attack",
		Type:        64,
		Ranged:      1,
		DamageValue: 1,
		ActionID:    18,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if !target.Paralyzed || !target.HasBuff(208) {
		t.Fatalf("骷髏毀壞應讓玩家目標麻痺，Paralyzed=%v buff208=%v", target.Paralyzed, target.GetBuff(208))
	}
}

func TestSkillIllusionistControlJoyOfPainCastPrimesCasterWithoutDamagingTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:  218,
		Target:   "attack",
		Type:     64,
		Ranged:   2,
		ActionID: 19,
		CastGfx:  6528,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if target.HP != 100 {
		t.Fatalf("疼痛的歡愉施放時不應直接傷害目標，HP=%d", target.HP)
	}
	buff := caster.GetBuff(218)
	if buff == nil || buff.TicksLeft != 16*5 {
		t.Fatalf("疼痛的歡愉應給施法者 16 秒狀態，buff=%v", buff)
	}
}

func TestSkillIllusionistControlJoyOfPainBacklashDamagesCasterOnce(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	caster.AddBuff(&world.ActiveBuff{SkillID: 218, TicksLeft: 16 * 5})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 1, ActionID: 18}

	s.applySkillDamageToPlayer(caster.Session, caster, target, skill, 5, []*world.PlayerInfo{caster, target})

	if target.HP != 45 {
		t.Fatalf("原本傷害仍應套用到目標，HP=%d", target.HP)
	}
	if caster.HP != 90 {
		t.Fatalf("疼痛的歡愉應依目標既有失血量反傷施法者 10，HP=%d", caster.HP)
	}
	if caster.HasBuff(218) {
		t.Fatalf("疼痛的歡愉反傷後應移除一次性狀態，buff=%v", caster.GetBuff(218))
	}
}
