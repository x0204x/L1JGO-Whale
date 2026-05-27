package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestBuffNonCancellableExcludesRemovedWarriorSkillsFor380C(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	t.Cleanup(engine.Close)

	for _, skillID := range []int{226, 228, 230} {
		if engine.IsNonCancellable(skillID) {
			t.Fatalf("3.80C 已剔除的戰士技能 %d 不應留在不可相消表", skillID)
		}
	}

	if !engine.IsNonCancellable(219) {
		t.Fatal("既有 3.80C 幻術化身 219 仍應維持不可相消")
	}
}
