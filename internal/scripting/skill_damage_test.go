package scripting

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
)

func TestCalcSkillDamageDisintegrateCanMagicCriticalLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	ctx := SkillDamageContext{
		SkillID:     77,
		DamageValue: 100,
		SkillLevel:  10,
		AttackerINT: 12,
		TargetMR:    0,
	}
	const normalDamage = 109
	criticalSeen := false

	for seed := 1; seed <= 500; seed++ {
		if err := engine.vm.DoString(fmt.Sprintf("math.randomseed(%d)", seed)); err != nil {
			t.Fatalf("設定 Lua random seed 失敗: %v", err)
		}
		result := engine.CalcSkillDamage(ctx)
		if result.Damage > normalDamage {
			criticalSeen = true
			break
		}
	}

	if !criticalSeen {
		t.Fatal("Java DISINTEGRATE 即使是 10 級魔法仍可觸發魔法爆擊，Go Lua 應允許出現 1.5 倍傷害")
	}
}
