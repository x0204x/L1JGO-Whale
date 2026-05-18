package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAIMobSkillTypeFiveReturnsAreaShockStun(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  2,
				MpConsume:     25,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("type 5 mob skill 應回傳一個 AI 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_shock_stun" {
		t.Fatalf("Java type 5 areashock_stun 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
}

func TestRunNpcAIMobSkillTypeFiveIgnoresMpConsumeLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
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
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  2,
				MpConsume:     25,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areashock_stun 沒有 MP 檢查，MP 不足仍應回傳 type 5 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_shock_stun" {
		t.Fatalf("Java type 5 MP 不足仍應走 area_shock_stun，got=%q", cmds[0].Type)
	}
}

func TestRunNpcAIMobSkillTypeFiveTriggerRandomOneAlwaysPassesLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	for i := 0; i < 20; i++ {
		cmds := engine.RunNpcAI(AIContext{
			NpcID:      231008,
			HP:         100,
			MaxHP:      100,
			MP:         100,
			MaxMP:      100,
			TargetID:   1001,
			TargetDist: 1,
			CanAttack:  true,
			CanMove:    true,
			Skills: []MobSkillEntry{
				{
					Type:          5,
					TriggerRandom: 1,
					TriggerRange:  2,
				},
			},
		})

		if len(cmds) != 1 || cmds[0].Type != "area_shock_stun" {
			t.Fatalf("Java L1MobSkillUse 以 trigger_random=1 必定通過，round=%d cmds=%+v", i, cmds)
		}
	}
}
