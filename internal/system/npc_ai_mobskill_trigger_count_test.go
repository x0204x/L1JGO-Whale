package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillTriggerCountLimitsUsesLikeJava(t *testing.T) {
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
		HP:          10,
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
	withNpcSelfHealMobSkillOptions(t, s, 1, 1)

	s.tickMonsterAI(npc)
	firstHP := npc.HP
	if firstHP <= 10 {
		t.Fatalf("TriCount 測試需要第一次 mob skill 成功補血，HP=%d", firstHP)
	}

	npc.AttackTimer = 0
	s.tickMonsterAI(npc)

	if npc.HP != firstHP {
		t.Fatalf("Java TriCount=1 應限制同一 mob skill 只使用一次，firstHP=%d afterSecond=%d", firstHP, npc.HP)
	}
}
