package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcSelfBuffMobSkillSkipsWhenTargetBehindWallLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "self_healer",
		X:           101,
		Y:           100,
		MapID:       900,
		HP:          50,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		Intel:       18,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcSelfHealMobSkill(t, s)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if npc.HP != 50 {
		t.Fatalf("隔牆目標不應讓怪物自補，HP=%d want=50", npc.HP)
	}
	if hasSkillEffectPacket(packets, npc.ID, 744) {
		t.Fatalf("Java L1NpcInstance.attack() 在目標隔牆時不呼叫 L1MobSkillUse.skillUse，不應送出自補效果封包")
	}
}

func TestNpcSelfBuffMobSkillStillWorksWithLineOfSightLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "self_healer",
		X:           101,
		Y:           100,
		MapID:       900,
		HP:          50,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		Intel:       18,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcSelfHealMobSkill(t, s)

	s.tickMonsterAI(npc)

	if npc.HP <= 50 {
		t.Fatalf("目標可視時怪物仍應能使用自放輔助技能，HP=%d want>50", npc.HP)
	}
}
