package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillIllusionistStatusMindBreakPlayerUsesJavaDamageAndDrainsMP(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ClassType: 6,
		Level:     60,
		Intel:     18,
		SP:        2,
		Wis:       10,
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
		MP:        20,
		MaxMP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:  207,
		Target:   "attack",
		Type:     64,
		Ranged:   3,
		ActionID: 18,
		CastGfx:  6553,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if target.MP != 15 {
		t.Fatalf("心靈破壞應扣目標 5 MP，MP=%d", target.MP)
	}
	if target.HP != 43 {
		t.Fatalf("心靈破壞應使用 Java SP*3.8 傷害，HP=%d", target.HP)
	}
}

func TestSkillIllusionistStatusConfusionSilencesPlayerTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
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
		SkillID:      202,
		Target:       "attack",
		Type:         64,
		Ranged:       4,
		BuffDuration: 8,
		DamageValue:  1,
		ActionID:     18,
		CastGfx:      6525,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if !target.Silenced || !target.HasBuff(64) {
		t.Fatalf("混亂應對玩家套用沉默狀態，Silenced=%v buff64=%v", target.Silenced, target.GetBuff(64))
	}
}

func TestSkillIllusionistStatusPhantasmSleepsPlayerTarget(t *testing.T) {
	// PHANTASM 已加入 playerDebuffSkills，需停用 MR 機率以避免測試 flaky。
	disablePlayerDebuffMRForStatusTest(t, 212)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:      212,
		Target:       "buff",
		BuffDuration: 5,
		ActionID:     19,
		CastGfx:      6530,
	}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if !target.Sleeped || target.Paralyzed {
		t.Fatalf("幻想應造成睡眠而不是一般麻痺，Sleeped=%v Paralyzed=%v", target.Sleeped, target.Paralyzed)
	}
	// Java skillmode/PHANTASM.java:22 對 PC `setSkillEffect(66, integer*1000)` ——
	// 實際 buff key 是 FOG_OF_SLEEPING(66) 而非 PHANTASM(212)，與 case 103 暗黑盲咒同模式。
	if !target.HasBuff(66) {
		t.Fatal("幻想應註冊 66 (FOG_OF_SLEEPING) active buff 以便到期解除睡眠（對齊 Java skillmode）")
	}
}

func TestSkillIllusionistStatusArmBreakerRevealsInvisiblePlayerTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
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
		Invisible: true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:     213,
		Target:      "attack",
		Type:        64,
		Ranged:      3,
		DamageValue: 15,
		ActionID:    18,
		CastGfx:     6551,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if target.Invisible || target.HasBuff(60) {
		t.Fatalf("武器破壞者應揭示隱身目標，Invisible=%v buff60=%v", target.Invisible, target.GetBuff(60))
	}
}
