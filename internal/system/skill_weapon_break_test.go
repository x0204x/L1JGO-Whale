package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
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

func TestApplyWeaponBreakDurabilityCapsAtJavaEnchantPlusFive(t *testing.T) {
	weapon := &world.InvItem{Durability: 4}

	changed := applyWeaponBreakDurability(weapon, 4)

	if !changed {
		t.Fatal("耐久未滿時應回報有變更")
	}
	if weapon.Durability != 5 {
		t.Fatalf("Java 武器損壞應封頂在 enchant+5，got=%d", weapon.Durability)
	}
}

func TestApplyWeaponBreakDurabilityIgnoresMissingWeapon(t *testing.T) {
	if applyWeaponBreakDurability(nil, 3) {
		t.Fatal("沒有裝備武器時不應回報耐久變更")
	}
}

func TestSkillWeaponBreakPlayerWeaponCapsAtJavaEnchantPlusFive(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "weapon-break-caster",
		X:           100,
		Y:           100,
		MapID:       4,
		Level:       99,
		MP:          100,
		MaxMP:       100,
		Intel:       3,
		KnownSpells: []int32{27},
	})
	caster.Inv.AddItemWithID(6001, 40318, 1, "魔法寶石", 0, 0, true, 0)
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "weapon-break-target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     1,
	})
	weapon := target.Inv.AddItemWithID(7001, 1, 1, "sword", 0, 0, false, 0)
	weapon.Equipped = true
	weapon.Durability = 5
	target.Equip.Set(world.SlotWeapon, weapon)
	s := newSkillBuffTestSystem(t, ws)
	s.deps.Items = loadDurabilityTestItems(t)

	s.processSkill(handler.SkillRequest{
		SessionID: caster.SessionID,
		SkillID:   27,
		TargetID:  target.CharID,
	})

	if weapon.Durability != 5 {
		t.Fatalf("Java receiveDamage 應將 +0 武器損壞封頂在 enchant+5=5，got=%d", weapon.Durability)
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasServerMessage(packets, 268) {
		t.Fatal("壞物術對玩家目標即使武器已達損壞上限仍應送 Java 訊息 268")
	}
	if !hasDurabilityStatusPacket(packets, weapon.ObjectID, 5) {
		t.Fatal("壞物術應送 S_ItemStatus 更新損壞度，而不是只送數量更新")
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
