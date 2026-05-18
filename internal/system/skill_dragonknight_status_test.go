package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillDragonKnightStatusResistFearAppliesDodgePenaltyOnly(t *testing.T) {
	ws := world.NewState()
	// Java skillmode/RESIST_FEAR.java:18 `if (!cha.hasSkillEffect(188) && cha instanceof L1PcInstance)` 對 PC 目標套用 dodge_down+5。
	// 走自我施放路徑：skill_buff.go:940 的 MR 抗性閘對 `caster.CharID == target.CharID` 自動跳過，
	// 與 playerDebuffSkills[188] 新增的 MR 機率（50±）解耦，避免測試變成概率性。
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Str:       12,
		Intel:     13,
		Dodge:     3,
	})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:      188,
		Target:       "buff",
		BuffDuration: 60,
		Ranged:       3,
		ActionID:     19,
		CastGfx:      6586,
	}

	s.executeBuffSkill(caster.Session, caster, skill, caster.CharID)

	if caster.Dodge != -2 {
		t.Fatalf("恐懼無助應降低閃避 5，Dodge=%d", caster.Dodge)
	}
	if caster.Str != 12 || caster.Intel != 13 {
		t.Fatalf("恐懼無助不應調整 STR/INT，Str=%d Int=%d", caster.Str, caster.Intel)
	}
}

func TestSkillDragonKnightStatusThunderGrabBindsPlayerTarget(t *testing.T) {
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
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:          192,
		Target:           "attack",
		Type:             64,
		Ranged:           8,
		DamageValue:      1,
		ProbabilityValue: 10,
		ProbabilityDice:  25,
		ActionID:         18,
		CastGfx:          6512,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if !target.Paralyzed || !target.HasBuff(192) {
		t.Fatalf("奪命之雷應束縛玩家目標，Paralyzed=%v buff192=%v", target.Paralyzed, target.GetBuff(192))
	}
}
