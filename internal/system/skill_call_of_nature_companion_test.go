package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillCallOfNatureCompanionCallOfNatureResurrectsDeadNpcToFullHP(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
		HP:    1,
		MaxHP: 150,
		MP:    0,
		MaxMP: 30,
	}
	ws.AddNpc(npc)
	npc.Dead = true
	npc.HP = 0
	ws.NpcDied(npc)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, npc.ID)

	if npc.Dead || npc.HP != npc.MaxHP {
		t.Fatalf("自然呼喚應滿血復活死亡 NPC，Dead=%v HP=%d/%d", npc.Dead, npc.HP, npc.MaxHP)
	}
}

func TestSkillCallOfNatureCompanionCallOfNatureResurrectsDeadPetToFullHP(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
		HP:          1,
		MaxHP:       80,
		MP:          0,
		MaxMP:       20,
		Status:      world.PetStatusAggressive,
	}
	ws.AddPet(pet)
	pet.Dead = true
	pet.HP = 0
	ws.PetDied(pet)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, pet.ID)

	if pet.Dead || pet.HP != pet.MaxHP || pet.Status != world.PetStatusRest {
		t.Fatalf("自然呼喚應滿血復活死亡寵物並改休息，Dead=%v HP=%d/%d Status=%d",
			pet.Dead, pet.HP, pet.MaxHP, pet.Status)
	}
}

func TestSkillCallOfNatureCompanionCallOfNatureRejectsPetCorpseTileOccupiedByAlivePlayer(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
		HP:          0,
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

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, pet.ID)

	if !pet.Dead || pet.HP != 0 {
		t.Fatalf("自然呼喚遇到寵物屍體格有活人時應拒絕，Dead=%v HP=%d", pet.Dead, pet.HP)
	}
}
