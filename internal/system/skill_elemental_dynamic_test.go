package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillElementalDynamicElementalFallDownUsesCasterElfAttr(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 133)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ElfAttr:   8,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		FireRes:   10,
		WaterRes:  11,
		WindRes:   12,
		EarthRes:  13,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      133,
		BuffDuration: 32,
		Target:       "buff",
		ActionID:     19,
	}, target.CharID)

	if !target.HasBuff(133) || target.FireRes != 10 || target.WaterRes != 11 || target.WindRes != -38 || target.EarthRes != 13 {
		t.Fatalf("弱化屬性應依施法者 ElfAttr 只降低單一屬性 50，四抗=%d/%d/%d/%d buff=%v",
			target.FireRes, target.WaterRes, target.WindRes, target.EarthRes, target.GetBuff(133))
	}
	s.removeBuffAndRevert(target, 133)
	if target.WindRes != 12 {
		t.Fatalf("弱化屬性解除後應還原風抗，WindRes=%d", target.WindRes)
	}
}

func TestSkillElementalDynamicElementalFallDownRestoresNpcResistance(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ElfAttr:   1,
	})
	npc := &world.NpcInfo{
		ID:       2001,
		NpcID:    45000,
		Name:     "mob",
		X:        101,
		Y:        100,
		MapID:    4,
		HP:       100,
		MaxHP:    100,
		EarthRes: 20,
	}
	ws.AddNpc(npc)

	applyElementalFallDownToNpc(caster, npc, 32)

	if !npc.HasDebuff(133) || npc.EarthRes != -30 || npc.ElementalFallDownAttr != 1 {
		t.Fatalf("NPC 弱化屬性應降低施法者屬性對應抗性 50，EarthRes=%d attr=%d debuff=%v",
			npc.EarthRes, npc.ElementalFallDownAttr, npc.HasDebuff(133))
	}
	removeNpcDebuffEffect(npc, 133, ws)
	if npc.EarthRes != 20 || npc.ElementalFallDownAttr != 0 {
		t.Fatalf("NPC 弱化屬性解除後應還原抗性與 attr，EarthRes=%d attr=%d", npc.EarthRes, npc.ElementalFallDownAttr)
	}
}

func TestSkillElementalDynamicWaterLifeAndPolluteWaterModifyHealing(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Intel:     18,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     200,
	})
	s := newSkillTestSystem(t, ws)
	healSkill := &data.SkillInfo{SkillID: 1, BuffDuration: 0, Target: "buff", Type: 16, DamageValue: 2, DamageDice: 1, ActionID: 19}

	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 170, BuffDuration: 64})
	s.executeBuffSkill(caster.Session, caster, healSkill, target.CharID)
	if target.HP != 60 || target.HasBuff(170) {
		t.Fatalf("水之元氣應讓下一次治療加倍並被移除，HP=%d buff170=%v", target.HP, target.GetBuff(170))
	}

	target.HP = 50
	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 173, BuffDuration: 32})
	s.executeBuffSkill(caster.Session, caster, healSkill, target.CharID)
	if target.HP != 52 || !target.HasBuff(173) {
		t.Fatalf("污濁之水應讓治療量減半且不被治療移除，HP=%d buff173=%v", target.HP, target.GetBuff(173))
	}
}

func TestSkillElementalDynamicElfCombatBuffsAreFlagsAndApplyDamageHelpers(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
		X:         100,
		Y:         100,
		MapID:     4,
		Dodge:     2,
		DmgMod:    3,
		BowHitMod: 4,
		BowDmgMod: 5,
		SP:        6,
		Intel:     18,
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

	for _, skillID := range []int32{160, 171, 174, 175} {
		s.applyBuffEffect(player, &data.SkillInfo{SkillID: skillID, BuffDuration: 64})
	}
	if player.Dodge != 7 {
		t.Fatalf("水之防護應依 Java getEr 給 ER/Dodge +5，Dodge=%d", player.Dodge)
	}
	if player.DmgMod != 3 || player.BowHitMod != 4 || player.BowDmgMod != 5 || player.SP != 6 || player.Intel != 18 {
		t.Fatalf("屬性之火/精準射擊/烈焰之魂應為戰鬥旗標，不應給固定面板，Dmg=%d BowHit=%d BowDmg=%d SP=%d Int=%d",
			player.DmgMod, player.BowHitMod, player.BowDmgMod, player.SP, player.Intel)
	}

	if got := elfMeleeDamageWithRoll(player, 100, "sword", 0); got != 225 {
		// Java `SOUL_OF_FLAME_DAMAGE=1.5` * `ELEMENTAL_FIRE=1.5` = 2.25x（兩者皆 round-down 整數）
		t.Fatalf("烈焰之魂(1.5x)與屬性之火(1.5x)應可套用近戰增傷，期望 225，got=%d", got)
	}
	if got := elfMeleeDamageWithRoll(player, 100, "bow", 0); got != 100 {
		t.Fatalf("烈焰之魂/屬性之火不應套用弓類武器，got=%d", got)
	}
	target.AddBuff(&world.ActiveBuff{SkillID: 174, TicksLeft: 10})
	if got := strikerGaleRangedDamage(target, 100); got != 110 {
		t.Fatalf("精準射擊應讓目標承受遠程傷害 1.1 倍，got=%d", got)
	}
}
