package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillDarkElfBuffShadowArmorAddsMROnly(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
		AC:        10,
		MR:        7,
	})
	s := newSkillTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 99, BuffDuration: 960})

	if !player.HasBuff(99) || player.MR != 12 || player.AC != 10 {
		t.Fatalf("影之防護應依 Java 只加 MR +5，不改 AC，buff=%v MR=%d AC=%d", player.GetBuff(99), player.MR, player.AC)
	}
}

func TestSkillDarkElfBuffDressBuffsUseJavaValues(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
		Str:       12,
		Dex:       13,
		Dodge:     2,
		AC:        10,
	})
	s := newSkillTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 109, BuffDuration: 960})
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 110, BuffDuration: 960})
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 111, BuffDuration: 960})

	if player.Str != 15 || player.Dex != 16 || player.Dodge != 20 || player.AC != 10 {
		t.Fatalf("黑妖提升技能應為 STR+3、DEX+3、ER/Dodge+18 且不改 AC，Str=%d Dex=%d Dodge=%d AC=%d",
			player.Str, player.Dex, player.Dodge, player.AC)
	}
}

func TestSkillDarkElfBuffBurningSpiritAndDoubleBreakAreProcFlags(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     50,
		SP:        3,
		HitMod:    2,
		DmgMod:    4,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 102, BuffDuration: 320})
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 105, BuffDuration: 320})

	if !player.HasBuff(102) || !player.HasBuff(105) || player.SP != 3 || player.HitMod != 2 || player.DmgMod != 4 {
		t.Fatalf("燃燒鬥志與雙重破壞應是觸發旗標，不應給固定數值，SP=%d Hit=%d Dmg=%d buff102=%v buff105=%v",
			player.SP, player.HitMod, player.DmgMod, player.GetBuff(102), player.GetBuff(105))
	}
	if got := darkElfPhysicalDamageWithRolls(player, 100, "edoryu", 0, 0); got != 300 {
		t.Fatalf("燃燒鬥志與雙重破壞同時觸發時應可連乘，got=%d", got)
	}
	if got := darkElfPhysicalDamageWithRolls(player, 100, "edoryu", 99, 99); got != 100 {
		t.Fatalf("燃燒鬥志與雙重破壞未觸發時不應增傷，got=%d", got)
	}
}

func TestSkillDarkElfBuffDarkBlindUsesSleepEffect66(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 103)
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
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 103, BuffDuration: 8, Target: "buff", ActionID: 19}, target.CharID)

	if !target.Sleeped || !target.HasBuff(66) || target.HasBuff(103) {
		t.Fatalf("暗黑盲咒應依 Java 使用 66 睡眠效果，Sleeped=%v buff66=%v buff103=%v",
			target.Sleeped, target.GetBuff(66), target.GetBuff(103))
	}
}
