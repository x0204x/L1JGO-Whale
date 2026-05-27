package scripting

import (
	"fmt"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func newStatCombatTestEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	t.Cleanup(engine.Close)
	return engine
}

func luaStatLookup(t *testing.T, engine *Engine, tableName string, stat int) int {
	t.Helper()
	code := fmt.Sprintf("__test_value = table_lookup(%s, %d)", tableName, stat)
	if err := engine.vm.DoString(code); err != nil {
		t.Fatalf("執行 Lua 查表失敗: %v", err)
	}
	value := engine.vm.GetGlobal("__test_value")
	n, ok := value.(lua.LNumber)
	if !ok {
		t.Fatalf("Lua 查表 %s[%d] 回傳型別錯誤: %T", tableName, stat, value)
	}
	return int(n)
}

func TestCombatStatTablesMatchYiweiJava(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	cases := []struct {
		table string
		stat  int
		want  int
	}{
		{"STR_HIT", 7, 4},
		{"STR_HIT", 8, 5},
		{"STR_HIT", 10, 6},
		{"STR_DMG", 9, 2},
		{"STR_DMG", 10, 3},
		{"STR_DMG", 12, 4},
		{"DEX_HIT", 7, -3},
		{"DEX_HIT", 8, -2},
		{"DEX_HIT", 20, 10},
		{"DEX_DMG", 8, 2},
		{"DEX_DMG", 9, 3},
		{"DEX_DMG", 12, 4},
		{"INT_DMG", 14, 0},
		{"INT_DMG", 15, 1},
		{"INT_DMG", 20, 2},
		{"INT_CRIT", 34, 0},
		{"INT_CRIT", 35, 1},
		{"INT_CRIT", 40, 2},
		{"INT_MAGIC_HIT", 22, 0},
		{"INT_MAGIC_HIT", 23, 1},
		{"INT_MAGIC_HIT", 26, 2},
	}

	for _, tc := range cases {
		if got := luaStatLookup(t, engine, tc.table, tc.stat); got != tc.want {
			t.Fatalf("%s[%d] 應對齊 yiwei Java L1AttackList=%d，實際 %d", tc.table, tc.stat, tc.want, got)
		}
	}
}

func TestMagicDamageMrDefenseUsesIntMagicHitLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := SkillDamageContext{
		SkillID:         10,
		DamageValue:     100,
		DamageDice:      0,
		DamageDiceCount: 0,
		SkillLevel:      7,
		AttackerINT:     23,
		AttackerBaseINT: 23,
		TargetMR:        80,
	}

	got := engine.CalcSkillDamage(ctx)
	if got.Damage != 123 {
		t.Fatalf("yiwei calcMrDefense 會用 INT_MAGIC_HIT 抵銷 MR：INT23/MR80 got=%d want=123", got.Damage)
	}
}

func TestMagicDamageMrDefenseUsesOriginalMagicHitLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := SkillDamageContext{
		SkillID:                  10,
		DamageValue:              100,
		DamageDice:               0,
		DamageDiceCount:          0,
		SkillLevel:               7,
		AttackerINT:              23,
		AttackerBaseINT:          23,
		AttackerOriginalMagicHit: 20,
		TargetMR:                 80,
	}

	got := engine.CalcSkillDamage(ctx)
	if got.Damage != 144 {
		t.Fatalf("yiwei calcMrDefense 會用 OriginalMagicHit 抵銷 MR：INT23+Original20/MR80 got=%d want=144", got.Damage)
	}
}

func TestMeleeDamageUsesStrengthLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := CombatContext{
		AttackerLevel:  50,
		AttackerSTR:    10,
		AttackerWeapon: 1,
		AttackerHitMod: 80,
		TargetAC:       10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	lowStr := engine.CalcMeleeAttack(ctx)
	if !lowStr.IsHit {
		t.Fatal("高命中測試前提失敗：低 STR 近戰應命中")
	}

	ctx.AttackerSTR = 20
	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	highStr := engine.CalcMeleeAttack(ctx)
	if !highStr.IsHit {
		t.Fatal("高命中測試前提失敗：高 STR 近戰應命中")
	}
	if highStr.Damage <= lowStr.Damage {
		t.Fatalf("yiwei 近戰傷害應吃 STR_DMG，low=%d high=%d", lowStr.Damage, highStr.Damage)
	}
}

func TestMeleeDamageUsesBaseStrengthBonusLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := CombatContext{
		AttackerLevel:   50,
		AttackerSTR:     25,
		AttackerBaseSTR: 24,
		AttackerWeapon:  1,
		AttackerHitMod:  80,
		TargetAC:        10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	base24 := engine.CalcMeleeAttack(ctx)
	if !base24.IsHit {
		t.Fatal("測試前提錯誤：Base STR 24 近戰應命中")
	}

	ctx.AttackerBaseSTR = 25
	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	base25 := engine.CalcMeleeAttack(ctx)
	if !base25.IsHit {
		t.Fatal("測試前提錯誤：Base STR 25 近戰應命中")
	}
	if base25.Damage <= base24.Damage {
		t.Fatalf("yiwei 近戰傷害應吃 Base STR 25+ 純能力加成，base24=%d base25=%d", base24.Damage, base25.Damage)
	}
}

func TestMagicDamageDoesNotRollTrueSPDiceLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := SkillDamageContext{
		SkillID:           10,
		DamageValue:       100,
		DamageDice:        10,
		DamageDiceCount:   0,
		SkillLevel:        7,
		AttackerINT:       18,
		AttackerTrueSP:    13,
		AttackerFullSP:    13,
		AttackerMagicCrit: 0,
		TargetMR:          0,
	}

	if err := engine.vm.DoString("math.randomseed(17)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	got := engine.CalcSkillDamage(ctx)

	// yiwei L1MagicPc.calcMagicDiceDamage() 只擲 damageDiceCount 次；
	// true SP 不會額外增加攻擊魔法傷害骰數。
	if got.Damage != 156 {
		t.Fatalf("攻擊魔法不應因 true SP 額外擲骰，got=%d want=156", got.Damage)
	}
}

func TestMagicDamageUsesIntAndExtraSPLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := SkillDamageContext{
		SkillID:         10,
		DamageValue:     100,
		DamageDice:      10,
		DamageDiceCount: 0,
		SkillLevel:      7,
		AttackerINT:     12,
		AttackerSP:      0,
		TargetMR:        0,
	}

	low := engine.CalcSkillDamage(ctx)

	ctx.AttackerINT = 18
	highInt := engine.CalcSkillDamage(ctx)
	if highInt.Damage <= low.Damage {
		t.Fatalf("yiwei 攻擊魔法傷害應吃 INT 係數，low=%d highInt=%d", low.Damage, highInt.Damage)
	}

	ctx.AttackerINT = 12
	ctx.AttackerSP = 3
	extraSP := engine.CalcSkillDamage(ctx)
	if extraSP.Damage <= low.Damage {
		t.Fatalf("yiwei 攻擊魔法傷害應吃額外 SP 係數，low=%d extraSP=%d", low.Damage, extraSP.Damage)
	}
}

func TestMagicDamageUsesBaseIntDamageBonusLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := SkillDamageContext{
		SkillID:         10,
		DamageValue:     100,
		DamageDice:      10,
		DamageDiceCount: 0,
		SkillLevel:      7,
		AttackerINT:     25,
		AttackerBaseINT: 24,
		TargetMR:        0,
	}

	base24 := engine.CalcSkillDamage(ctx)
	ctx.AttackerBaseINT = 25
	base25 := engine.CalcSkillDamage(ctx)
	if base25.Damage <= base24.Damage {
		t.Fatalf("yiwei 攻擊魔法傷害應吃 Base INT 25+ 純能力加成，base24=%d base25=%d", base24.Damage, base25.Damage)
	}
}

func TestRangedDamageUsesDexButNotStrengthLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := RangedCombatContext{
		AttackerLevel:     50,
		AttackerDEX:       20,
		AttackerSTR:       10,
		AttackerBowDmg:    1,
		AttackerArrowDmg:  0,
		AttackerBowHitMod: 80,
		TargetAC:          10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	lowStr := engine.CalcRangedAttack(ctx)
	if !lowStr.IsHit {
		t.Fatal("測試前提錯誤：低 STR 遠攻應命中")
	}

	ctx.AttackerSTR = 50
	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	highStr := engine.CalcRangedAttack(ctx)
	if !highStr.IsHit {
		t.Fatal("測試前提錯誤：高 STR 遠攻應命中")
	}
	if highStr.Damage != lowStr.Damage {
		t.Fatalf("yiwei 遠距離傷害只吃 DEX_DMG/弓傷/箭傷，不應因 STR 變動；low=%d high=%d", lowStr.Damage, highStr.Damage)
	}
}

func TestCombatCriticalTablesMatchYiweiJava(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	cases := []struct {
		table string
		stat  int
		want  int
	}{
		{"STR_CRIT", 39, 0},
		{"STR_CRIT", 40, 1},
		{"STR_CRIT", 50, 2},
		{"DEX_CRIT", 39, 0},
		{"DEX_CRIT", 40, 1},
		{"DEX_CRIT", 50, 2},
	}

	for _, tc := range cases {
		if got := luaStatLookup(t, engine, tc.table, tc.stat); got != tc.want {
			t.Fatalf("%s[%d] 應對齊 yiwei Java L1AttackList=%d，got=%d", tc.table, tc.stat, tc.want, got)
		}
	}
}

func TestMeleeDamageUsesStrCriticalLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := CombatContext{
		AttackerLevel:  50,
		AttackerSTR:    10,
		AttackerWeapon: 10,
		AttackerHitMod: 80,
		TargetAC:       10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	normal := engine.CalcMeleeAttack(ctx)
	if !normal.IsHit {
		t.Fatal("測試前提錯誤：一般近戰應命中")
	}

	if err := engine.vm.DoString("STR_CRIT = { _max_index = 127 }; for i = 0, 127 do STR_CRIT[i] = 100 end; math.randomseed(23)"); err != nil {
		t.Fatalf("設定 STR_CRIT 測試表失敗: %v", err)
	}
	critical := engine.CalcMeleeAttack(ctx)
	if !critical.IsHit {
		t.Fatal("測試前提錯誤：爆擊近戰應命中")
	}
	if critical.Damage <= normal.Damage {
		t.Fatalf("yiwei 近戰爆擊應將武器傷害提升到最大值兩倍，normal=%d critical=%d", normal.Damage, critical.Damage)
	}
}

func TestRangedDamageUsesDexCriticalLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := RangedCombatContext{
		AttackerLevel:     50,
		AttackerDEX:       10,
		AttackerBowDmg:    10,
		AttackerArrowDmg:  0,
		AttackerBowHitMod: 80,
		TargetAC:          10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	normal := engine.CalcRangedAttack(ctx)
	if !normal.IsHit {
		t.Fatal("測試前提錯誤：一般遠攻應命中")
	}

	if err := engine.vm.DoString("DEX_CRIT = { _max_index = 127 }; for i = 0, 127 do DEX_CRIT[i] = 100 end; math.randomseed(23)"); err != nil {
		t.Fatalf("設定 DEX_CRIT 測試表失敗: %v", err)
	}
	critical := engine.CalcRangedAttack(ctx)
	if !critical.IsHit {
		t.Fatal("測試前提錯誤：爆擊遠攻應命中")
	}
	if critical.Damage <= normal.Damage {
		t.Fatalf("yiwei 遠攻爆擊應將武器傷害提升到最大值，normal=%d critical=%d", normal.Damage, critical.Damage)
	}
}

func TestRangedAttackUsesTargetErEvasionLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	ctx := RangedCombatContext{
		AttackerLevel:     100,
		AttackerDEX:       50,
		AttackerBaseDEX:   50,
		AttackerBowDmg:    1,
		AttackerArrowDmg:  0,
		AttackerBowHitMod: 100,
		TargetAC:          10,
		TargetClassType:   0,
		TargetDodge:       3000,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	got := engine.CalcRangedAttack(ctx)
	if got.IsHit {
		t.Fatalf("yiwei 遠距離攻擊命中後會再跑 calcErEvasion，ER=3000 應必定閃避：got hit damage=%d", got.Damage)
	}
}

func TestNpcRangedAttackUsesTargetErEvasionLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	ctx := CombatContext{
		AttackerLevel:   100,
		AttackerSTR:     10,
		AttackerDEX:     50,
		AttackerWeapon:  1,
		AttackerHitMod:  100,
		TargetAC:        10,
		TargetClassType: 0,
		TargetDodge:     3000,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	got := engine.CalcNpcRanged(ctx)
	if got.IsHit {
		t.Fatalf("yiwei NPC 遠距離攻擊命中後會再跑 calcErEvasion，ER=3000 應必定閃避：got hit damage=%d", got.Damage)
	}
}

func TestRangedDamageUsesBaseDexBonusLikeYiwei(t *testing.T) {
	engine := newStatCombatTestEngine(t)
	ctx := RangedCombatContext{
		AttackerLevel:     50,
		AttackerDEX:       25,
		AttackerBaseDEX:   24,
		AttackerBowDmg:    1,
		AttackerArrowDmg:  0,
		AttackerBowHitMod: 80,
		TargetAC:          10,
	}

	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	base24 := engine.CalcRangedAttack(ctx)
	if !base24.IsHit {
		t.Fatal("測試前提錯誤：Base DEX 24 遠攻應命中")
	}

	ctx.AttackerBaseDEX = 25
	if err := engine.vm.DoString("math.randomseed(23)"); err != nil {
		t.Fatalf("設定 Lua random seed 失敗: %v", err)
	}
	base25 := engine.CalcRangedAttack(ctx)
	if !base25.IsHit {
		t.Fatal("測試前提錯誤：Base DEX 25 遠攻應命中")
	}
	if base25.Damage <= base24.Damage {
		t.Fatalf("yiwei 遠攻傷害應吃 Base DEX 25+ 純能力加成，base24=%d base25=%d", base24.Damage, base25.Damage)
	}
}
