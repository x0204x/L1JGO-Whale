package scripting

import "testing"

func TestCalcHealAmountUsesYiweiDiceCountFromValueAndMagicBonus(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	got := engine.callIntFunc("calc_heal_amount", 2, 1, 99, 18, 0, 10)
	if got != 5 {
		t.Fatalf("yiwei 治療量應以 damage_value + INT magicBonus 作為骰數，且忽略 damage_dice_count；got=%d want=5", got)
	}
}

func TestCalcHealAmountAppliesYiweiLawfulAndLeverage(t *testing.T) {
	engine := newStatCombatTestEngine(t)

	gotLawful := engine.callIntFunc("calc_heal_amount", 2, 1, 0, 18, 32768, 10)
	if gotLawful != 10 {
		t.Fatalf("yiwei lawful 正值應放大治療量；got=%d want=10", gotLawful)
	}

	gotLeverage := engine.callIntFunc("calc_heal_amount", 2, 1, 0, 18, 0, 20)
	if gotLeverage != 10 {
		t.Fatalf("yiwei leverage 應在 lawful 後套用；got=%d want=10", gotLeverage)
	}
}
