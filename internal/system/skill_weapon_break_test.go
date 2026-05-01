package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestCalcWeaponBreakDurabilityDamageKeepsAtLeastOne(t *testing.T) {
	caster := &world.PlayerInfo{Intel: 2}

	damage := calcWeaponBreakDurabilityDamage(caster)

	if damage != 1 {
		t.Fatalf("低 INT 施放壞物術仍應至少造成 1 點耐久傷害，got=%d", damage)
	}
}

func TestCalcWeaponBreakDurabilityDamageUsesIntDividedByThreeRange(t *testing.T) {
	caster := &world.PlayerInfo{Intel: 12}

	for i := 0; i < 50; i++ {
		damage := calcWeaponBreakDurabilityDamage(caster)
		if damage < 1 || damage > 4 {
			t.Fatalf("INT 12 的壞物術耐久傷害應落在 1..4，got=%d", damage)
		}
	}
}

func TestApplyWeaponBreakDurabilityCapsAtClientRange(t *testing.T) {
	weapon := &world.InvItem{Durability: 126}

	changed := applyWeaponBreakDurability(weapon, 4)

	if !changed {
		t.Fatal("耐久未滿時應回報有變更")
	}
	if weapon.Durability != 127 {
		t.Fatalf("耐久應封頂在 127，got=%d", weapon.Durability)
	}
}

func TestApplyWeaponBreakDurabilityIgnoresMissingWeapon(t *testing.T) {
	if applyWeaponBreakDurability(nil, 3) {
		t.Fatal("沒有裝備武器時不應回報耐久變更")
	}
}

func TestApplyNpcWeaponBreakDamageHalvesBrokenNpcMeleeDamage(t *testing.T) {
	npc := &world.NpcInfo{WeaponBroken: true}

	damage := applyNpcWeaponBreakDamage(npc, 15)

	if damage != 7 {
		t.Fatalf("NPC 被壞物術後近戰傷害應減半並向下取整，got=%d", damage)
	}
}

func TestClearNpcCancellationStateClearsWeaponBreak(t *testing.T) {
	npc := &world.NpcInfo{
		WeaponBroken: true,
		Paralyzed:    true,
		Sleeped:      true,
	}

	clearNpcCancellationState(npc)

	if npc.WeaponBroken {
		t.Fatal("相消術應清除 NPC 壞物術狀態")
	}
}
