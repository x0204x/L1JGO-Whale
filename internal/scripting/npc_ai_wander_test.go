package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAINoTargetCanReturnJavaPauseDirection(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	seenPause := false
	for i := 0; i < 80; i++ {
		cmds := engine.RunNpcAI(AIContext{
			NpcID:   45039,
			HP:      100,
			MaxHP:   100,
			MP:      100,
			MaxMP:   100,
			CanMove: true,
		})
		if len(cmds) != 1 || cmds[0].Type != "wander" {
			t.Fatalf("無目標且可移動時應回傳 wander 指令，round=%d cmds=%+v", i, cmds)
		}
		if cmds[0].Dir >= 8 && cmds[0].Dir <= 39 {
			seenPause = true
			break
		}
	}

	if !seenPause {
		t.Fatal("Java noTarget 使用 random 0-39，Lua AI 應可能回傳 8-39 的停留方向")
	}
}
