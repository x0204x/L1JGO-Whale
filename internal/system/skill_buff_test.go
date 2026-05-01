package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func newSkillBuffTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	s := newSkillTestSystem(t, ws)

	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入技能資料失敗: %v", err)
	}
	buffIcons, err := data.LoadBuffIconTable("../../data/yaml/buff_icon_map.yaml")
	if err != nil {
		t.Fatalf("載入 buff 圖示資料失敗: %v", err)
	}

	s.deps.Skills = skills
	s.deps.BuffIcons = buffIcons
	return s
}

func TestSkillBuffSilenceBuffSetsAndClearsCastingLock(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:      64,
		BuffDuration: 16,
	}

	s.applyBuffEffect(target, skill)

	if !target.Silenced {
		t.Fatal("沉默技能套用後應禁止玩家施法")
	}
	if !target.HasBuff(64) {
		t.Fatal("沉默技能應註冊 active buff")
	}

	s.removeBuffAndRevert(target, 64)

	if target.Silenced {
		t.Fatal("沉默 buff 移除後應解除施法禁止")
	}
}

func TestSkillBuffImmuneToHarmHalvesPlayerMagicDamage(t *testing.T) {
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
	target.AddBuff(&world.ActiveBuff{SkillID: 68, TicksLeft: 150})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:  4,
		ActionID: 18,
		CastGfx:  167,
	}

	s.applySkillDamageToPlayer(caster.Session, caster, target, skill, 20, []*world.PlayerInfo{caster, target})

	if target.HP != 90 {
		t.Fatalf("聖結界應將玩家魔法傷害 20 減半為 10，HP=%d", target.HP)
	}
}

func TestSkillBuffImmuneToHarmHelperHalvesPhysicalDamage(t *testing.T) {
	target := &world.PlayerInfo{}
	target.AddBuff(&world.ActiveBuff{SkillID: 68, TicksLeft: 150})

	damage := applyImmuneToHarmDamage(target, 21)

	if damage != 10 {
		t.Fatalf("聖結界共用 helper 應將物理/魔法傷害向下減半，got=%d", damage)
	}
}

func TestSkillBuffCounterMagicConsumesBuffOnPlayerAttackSkill(t *testing.T) {
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
	target.AddBuff(&world.ActiveBuff{SkillID: 31, TicksLeft: 150})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         4,
		SkillLevel:      1,
		Target:          "attack",
		Type:            64,
		DamageValue:     20,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          10,
		ActionID:        18,
		CastGfx:         167,
	}

	s.executeAttackSkill(caster.Session, caster, skill, target.CharID)

	if target.HP != 100 {
		t.Fatalf("魔法屏障抵消後不應受到攻擊魔法傷害，HP=%d", target.HP)
	}
	if target.HasBuff(31) {
		t.Fatal("魔法屏障抵消攻擊魔法後應被消耗")
	}
}
