package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillSelfBuffConsumesSilenceWithoutUseLikeJava(t *testing.T) {
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
	npc.AddDebuff(64, 80)
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcSelfHealMobSkillOptions(t, s, 1, 1)

	s.tickMonsterAI(npc)

	if npc.HasDebuff(64) {
		t.Fatalf("Java L1SkillUse.checkUseSkill() 會移除 NPC 身上的 SILENCE(64)")
	}
	if npc.HP != 50 {
		t.Fatalf("NPC 被 SILENCE(64) 時該次 mobskill 不應成功回血，HP=%d want=50", npc.HP)
	}
	if got := npc.MobSkillUseCounts[0]; got != 0 {
		t.Fatalf("Java mobskill 只有成功 useSkill 才累積 TriCount，got=%d want=0", got)
	}

	npc.AttackTimer = 0
	s.tickMonsterAI(npc)

	if npc.HP <= 50 {
		t.Fatalf("SILENCE(64) 被消費後下一次 mobskill 應可正常自補，HP=%d want>50", npc.HP)
	}
	if got := npc.MobSkillUseCounts[0]; got != 1 {
		t.Fatalf("第二次成功施法後 TriCount 應累積一次，got=%d want=1", got)
	}
}
