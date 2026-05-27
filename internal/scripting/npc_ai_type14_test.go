package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAIMobSkillTypeFourteenIgnoresMpAndReturnsAreaDebuffLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      97257,
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
				Type:          14,
				TriggerRandom: 1,
				TriggerRange:  -14,
				MpConsume:     25,
				Leverage:      5,
				SkillID:       71,
				GfxID:         11030,
				ActID:         18,
				ReuseDelay:    15000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areadebuff 沒有 MP 檢查，MP 不足仍應回傳 type 14 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_debuff" {
		t.Fatalf("Java type 14 areadebuff 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].SkillID != 71 {
		t.Fatalf("area_debuff 應保留 mobskill skill_id，got=%d want=71", cmds[0].SkillID)
	}
	if cmds[0].Leverage != 5 {
		t.Fatalf("area_debuff 應保留 mobskill leverage 作為 buff 秒數，got=%d want=5", cmds[0].Leverage)
	}
	if cmds[0].GfxID != 11030 {
		t.Fatalf("area_debuff 應保留 mobskill gfx_id，got=%d want=11030", cmds[0].GfxID)
	}
	if cmds[0].ActID != 18 {
		t.Fatalf("area_debuff 應保留 mobskill act_id，got=%d want=18", cmds[0].ActID)
	}
	if cmds[0].ReuseDelay != 15000 {
		t.Fatalf("area_debuff 應保留 mobskill reuse_delay，got=%d want=15000", cmds[0].ReuseDelay)
	}
}
