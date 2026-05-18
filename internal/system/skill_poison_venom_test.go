package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillPoisonVenomVenomResistBlocksNpcPoison(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 104, BuffDuration: 320})
	npc := &world.NpcInfo{ID: 2001, PoisonAtk: 1}

	ApplyNpcPoisonAttackWithRoll(npc, target, s.deps, 0)

	if target.PoisonType != 0 {
		t.Fatalf("毒性抵抗 104 應阻擋 NPC 施毒，PoisonType=%d", target.PoisonType)
	}
}

func TestSkillPoisonVenomDragonLifeEyeBlocksPoison(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	target.AddBuff(&world.ActiveBuff{SkillID: 6687, TicksLeft: 100})
	npc := &world.NpcInfo{ID: 2001, PoisonAtk: 1}

	ApplyNpcPoisonAttackWithRoll(npc, target, s.deps, 0)

	if target.PoisonType != 0 {
		t.Fatalf("生命魔眼 6687 應阻擋中毒，PoisonType=%d", target.PoisonType)
	}
}

func TestSkillPoisonVenomCursePoisonRespectsPlayerPoisonResistance(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 11)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 104, TicksLeft: 100})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 11, ActionID: 19, CastGfx: 745}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.PoisonType != 0 {
		t.Fatalf("毒咒應依 Java L1Poison.isValidTarget 受毒性抵抗阻擋，PoisonType=%d", target.PoisonType)
	}
}

func TestSkillPoisonVenomCursePoisonDoesNotOverwriteNpcExistingPoison(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:                2001,
		X:                 101,
		Y:                 100,
		MapID:             4,
		HP:                100,
		MaxHP:             100,
		PoisonDmgAmt:      20,
		PoisonDmgTimer:    7,
		PoisonAttackerSID: 99,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 11, ActionID: 19, CastGfx: 745}

	s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)

	if npc.PoisonDmgAmt != 20 || npc.PoisonDmgTimer != 7 || npc.PoisonAttackerSID != 99 {
		t.Fatalf("毒咒不應覆寫既有毒狀態，amount=%d timer=%d attacker=%d",
			npc.PoisonDmgAmt, npc.PoisonDmgTimer, npc.PoisonAttackerSID)
	}
}

func TestSkillPoisonVenomEnchantVenomPoisonsPlayerAndRespectsResistance(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	attacker.AddBuff(&world.ActiveBuff{SkillID: 98, TicksLeft: 100})
	attacker.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 1, Equipped: true})

	if !applyEnchantVenomPoisonToPlayerWithRoll(attacker, target, s.deps, 0) {
		t.Fatalf("附加劇毒在機率內且有武器時應可讓玩家目標中傷害毒")
	}
	if target.PoisonType != 1 || target.PoisonDmgAmount != 5 || target.PoisonTicksLeft != 150 {
		t.Fatalf("附加劇毒玩家目標應為 30 秒、每次 5 傷害毒，type=%d amount=%d ticks=%d",
			target.PoisonType, target.PoisonDmgAmount, target.PoisonTicksLeft)
	}

	CurePoison(target, s.deps)
	target.AddBuff(&world.ActiveBuff{SkillID: 104, TicksLeft: 100})
	if applyEnchantVenomPoisonToPlayerWithRoll(attacker, target, s.deps, 0) || target.PoisonType != 0 {
		t.Fatalf("毒性抵抗應阻擋附加劇毒，PoisonType=%d", target.PoisonType)
	}
}

func TestSkillPoisonVenomEnchantVenomPoisonsNpc(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{ID: 2001, X: 101, Y: 100, MapID: 4, HP: 100, MaxHP: 100}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attacker.AddBuff(&world.ActiveBuff{SkillID: 98, TicksLeft: 100})
	attacker.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 1, Equipped: true})

	if !applyEnchantVenomPoisonToNpcWithRoll(attacker, npc, s.deps, 0) {
		t.Fatalf("附加劇毒在機率內且有武器時應可讓 NPC 中傷害毒")
	}
	if npc.PoisonDmgAmt != 5 || npc.PoisonDmgTimer != 0 || npc.PoisonAttackerSID != attacker.SessionID {
		t.Fatalf("附加劇毒 NPC 目標應為每次 5 傷害毒且記錄攻擊者，amount=%d timer=%d attacker=%d",
			npc.PoisonDmgAmt, npc.PoisonDmgTimer, npc.PoisonAttackerSID)
	}
}
