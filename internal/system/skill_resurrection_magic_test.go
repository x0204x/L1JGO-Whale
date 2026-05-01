package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillResurrectionMagicResurrectionResurrectsDeadNpcQuarterHP(t *testing.T) {
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
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		X:     101,
		Y:     100,
		MapID: 4,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	npc.Dead = true
	npc.HP = 0
	ws.NpcDied(npc)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 61, ActionID: 19, CastGfx: 3944}, npc.ID)

	if npc.Dead || npc.HP != 25 {
		t.Fatalf("返生術應以 1/4 HP 復活 NPC，Dead=%v HP=%d", npc.Dead, npc.HP)
	}
}

func TestSkillResurrectionMagicGreaterResurrectionResurrectsDeadPetQuarterHP(t *testing.T) {
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
	pet := &world.PetInfo{
		ID:          3001,
		OwnerCharID: caster.CharID,
		NpcID:       45046,
		Name:        "pet",
		X:           101,
		Y:           100,
		MapID:       4,
		MaxHP:       80,
		Status:      world.PetStatusAggressive,
	}
	ws.AddPet(pet)
	pet.Dead = true
	pet.HP = 0
	ws.PetDied(pet)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 75, ActionID: 19, CastGfx: 3944}, pet.ID)

	if pet.Dead || pet.HP != 20 || pet.Status != world.PetStatusRest {
		t.Fatalf("終極返生術對寵物應依 Java 以 1/4 HP 復活，Dead=%v HP=%d Status=%d", pet.Dead, pet.HP, pet.Status)
	}
}

func TestSkillResurrectionMagicResurrectionRejectsCantResurrectNpc(t *testing.T) {
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
		ID:            2001,
		NpcID:         45000,
		Impl:          "L1Monster",
		Name:          "no_rez",
		X:             101,
		Y:             100,
		MapID:         4,
		MaxHP:         100,
		CantResurrect: true,
	}
	ws.AddNpc(npc)
	npc.Dead = true
	npc.HP = 0
	ws.NpcDied(npc)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 61, ActionID: 19, CastGfx: 3944}, npc.ID)

	if !npc.Dead || npc.HP != 0 {
		t.Fatalf("cant_resurrect NPC 不應被返生術復活，Dead=%v HP=%d", npc.Dead, npc.HP)
	}
}

func TestSkillResurrectionMagicResurrectionRejectsPetCorpseTileOccupiedByAlivePlayer(t *testing.T) {
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
	pet := &world.PetInfo{
		ID:          3001,
		OwnerCharID: caster.CharID,
		NpcID:       45046,
		Name:        "pet",
		X:           101,
		Y:           100,
		MapID:       4,
		MaxHP:       80,
		Dead:        true,
	}
	ws.AddPet(pet)
	ws.PetDied(pet)
	addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "blocker",
		X:         pet.X,
		Y:         pet.Y,
		MapID:     pet.MapID,
	})
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 61, ActionID: 19, CastGfx: 3944}, pet.ID)

	if !pet.Dead || pet.HP != 0 {
		t.Fatalf("寵物屍體格有活人時返生術應拒絕，Dead=%v HP=%d", pet.Dead, pet.HP)
	}
}
