package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAIMobSkillTypeFifteenIgnoresMpAndReturnsAreaPoisonLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      45625,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          15,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         19,
				Leverage:      5,
				SummonID:      86125,
				ReuseDelay:    12000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.area_poison 不檢查 MP，MP 不足仍應回傳 type 15 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_poison" {
		t.Fatalf("Java type 15 area_poison 不應退回一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 19 {
		t.Fatalf("area_poison 應保留 mobskill act_id，got=%d want=19", cmds[0].ActID)
	}
	if cmds[0].Leverage != 5 {
		t.Fatalf("area_poison 應保留 mobskill leverage 作為存在秒數，got=%d want=5", cmds[0].Leverage)
	}
	if cmds[0].SummonID != 86125 {
		t.Fatalf("area_poison 應保留 mobskill summon_id，got=%d want=86125", cmds[0].SummonID)
	}
	if cmds[0].ReuseDelay != 12000 {
		t.Fatalf("area_poison 應保留 mobskill reuse_delay，got=%d want=12000", cmds[0].ReuseDelay)
	}
}
