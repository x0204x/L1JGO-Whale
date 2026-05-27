package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillDamageNpcKirtasBarrierThreeAbsoluteBarrierBlocksDamageLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     900,
		Intel:     30,
		SP:        20,
		MP:        100,
		MaxMP:     100,
		HP:        1000,
		MaxHP:     1000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  81163,
		Impl:   "L1Monster",
		Name:   "kirtas_absolute_barrier_npc",
		X:      101,
		Y:      100,
		MapID:  900,
		Level:  80,
		HP:     100000,
		MaxHP:  100000,
		ShowID: caster.ShowID,
		MR:     0,
	}
	npc.AddDebuff(11058, 120)
	npc.AddDebuff(78, 120)
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)

	initialNpcHP := npc.HP
	s.executeAttackSkill(caster.Session, caster, kirtasBarrierTestAttackSkill(), npc.ID)

	if npc.HP != initialNpcHP {
		t.Fatalf("yiwei NPC 持有 ABSOLUTE_BARRIER(78) 時魔法傷害應被 dmg0 擋下，NPC HP=%d want=%d", npc.HP, initialNpcHP)
	}
}

func TestSkillDamageNpcKirtasBarrierTwoReflectsMagicDamageLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     900,
		Intel:     30,
		SP:        20,
		MP:        100,
		MaxMP:     100,
		HP:        1000,
		MaxHP:     1000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  81163,
		Impl:   "L1Monster",
		Name:   "kirtas_mirror_npc",
		X:      101,
		Y:      100,
		MapID:  900,
		Level:  80,
		HP:     100000,
		MaxHP:  100000,
		ShowID: caster.ShowID,
		MR:     0,
	}
	npc.AddDebuff(11059, 120)
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)

	initialCasterHP := caster.HP
	initialNpcHP := npc.HP
	s.executeAttackSkill(caster.Session, caster, kirtasBarrierTestAttackSkill(), npc.ID)
	packets := drainSkillTestPackets(caster.Session)

	if npc.HP != initialNpcHP {
		t.Fatalf("yiwei KIRTAS_BARRIER2(11059) 反射後原魔法傷害歸零，NPC HP=%d want=%d", npc.HP, initialNpcHP)
	}
	if caster.HP >= initialCasterHP {
		t.Fatalf("yiwei KIRTAS_BARRIER2(11059) 會把魔法傷害反彈給施法者，caster HP=%d initial=%d", caster.HP, initialCasterHP)
	}
	if !hasActionGfxPacket(packets, caster.CharID, 2) {
		t.Fatalf("yiwei 11059 魔法反射會送施法者 action=2，packets=%v", packets)
	}
	if !hasSkillEffectPacket(packets, npc.ID, 4395) {
		t.Fatalf("yiwei 11059 魔法反射會對 NPC 廣播 4395，packets=%v", packets)
	}
}

func kirtasBarrierTestAttackSkill() *data.SkillInfo {
	return &data.SkillInfo{
		SkillID:         4,
		Target:          "attack",
		Type:            64,
		DamageValue:     50,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          10,
		ActionID:        18,
		CastGfx:         167,
	}
}
