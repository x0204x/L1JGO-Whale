package scripting

import (
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
	if err := engine.vm.DoString("math.randomseed(1)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	normal := engine.CalcSkillDamage(ctx)

	ctx.AttackerMagicCrit = 100
	if err := engine.vm.DoString("math.randomseed(1)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	critical := engine.CalcSkillDamage(ctx)

	if critical.Damage <= normal.Damage {
		t.Fatalf("Java DISINTEGRATE 即使是 10 級魔法仍可依魔法爆擊加成觸發 1.5 倍傷害，normal=%d critical=%d", normal.Damage, critical.Damage)
	}
}
