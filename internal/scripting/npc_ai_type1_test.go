package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAIMobSkillTypeOnePreservesPhysicalRangeLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      45040,
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
				ActNo:         1,
				Type:          1,
				TriggerRandom: 1,
				TriggerRange:  -14,
				Range:         1,
				AreaWidth:     3,
				AreaHeight:    2,
				ActID:         30,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("type 1 物理技能應回傳一個指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "skill" || cmds[0].SkillType != 1 {
		t.Fatalf("type 1 物理技能應保留 skill_type=1 交給 System 走 physicalAttack，got=%+v", cmds[0])
	}
	if cmds[0].Range != 1 || cmds[0].AreaWidth != 3 || cmds[0].AreaHeight != 2 {
		t.Fatalf("type 1 物理技能應保留 yiwei range/area 欄位，got=%+v", cmds[0])
	}
}
