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

func TestSkillResurrectionRejectsDifferentShowPlayerTargetLikeJava(t *testing.T) {
	cases := []struct {
		name    string
		skillID int32
	}{
		{name: "resurrection", skillID: 61},
		{name: "greater resurrection", skillID: 75},
		{name: "call of nature", skillID: 165},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "caster",
				X:         100,
				Y:         100,
				MapID:     4,
				ShowID:    30,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 2,
				Session:   newSkillTestSession(t, 2),
				CharID:    1002,
				Name:      "other-show-dead",
				X:         101,
				Y:         100,
				MapID:     4,
				ShowID:    31,
				Dead:      true,
				HP:        0,
				MaxHP:     100,
				MaxMP:     50,
			})
			s := newSkillTestSystem(t, ws)

			s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: tc.skillID, ActionID: 19, CastGfx: 3944}, target.CharID)

			if target.PendingResSkill != 0 || target.PendingResCaster != 0 {
				t.Fatalf("不同 ShowID 玩家不應收到 %s 復活同意，Pending=(%d,%d)",
					tc.name, target.PendingResSkill, target.PendingResCaster)
			}
		})
	}
}

func TestSkillResurrectionCastBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    40,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "dead",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    40,
		Dead:      true,
		HP:        0,
		MaxHP:     100,
		MaxMP:     50,
	})
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    40,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    41,
	})
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 61, ActionID: 19, CastGfx: 3944}, target.CharID)

	if !hasActionGfxPacket(drainSkillTestPackets(sameViewer.Session), caster.CharID, 19) {
		t.Fatal("同 ShowID 玩家應收到返生術施法動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherViewer.Session), caster.CharID, 19) {
		t.Fatal("不同 ShowID 玩家不應收到返生術施法動作")
	}
}

func TestSkillResurrectPlayerBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    50,
		Level:     50,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "dead",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    50,
		Dead:      true,
		HP:        0,
		MaxHP:     100,
		MaxMP:     50,
	})
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    50,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    51,
	})
	s := newSkillTestSystem(t, ws)

	s.resurrectPlayer(target, caster, &data.SkillInfo{SkillID: 61})

	if !hasPutObjectPacket(drainSkillTestPackets(sameViewer.Session), target.CharID) {
		t.Fatal("同 ShowID 玩家應收到復活玩家 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherViewer.Session), target.CharID) {
		t.Fatal("不同 ShowID 玩家不應收到復活玩家 put object")
	}
}

func TestSkillResurrectPetBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    60,
	})
	pet := &world.PetInfo{
		ID:          3001,
		OwnerCharID: caster.CharID,
		NpcID:       45046,
		Name:        "pet",
		X:           101,
		Y:           100,
		MapID:       4,
		ShowID:      60,
		MaxHP:       80,
		Status:      world.PetStatusAggressive,
	}
	ws.AddPet(pet)
	pet.Dead = true
	pet.HP = 0
	ws.PetDied(pet)
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    60,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    61,
	})
	s := newSkillTestSystem(t, ws)

	if !s.resurrectPetWithHP(caster.Session, caster, &data.SkillInfo{SkillID: 75, CastGfx: 3944}, pet, pet.MaxHP/4) {
		t.Fatal("寵物復活應成功")
	}

	samePackets := drainSkillTestPackets(sameViewer.Session)
	otherPackets := drainSkillTestPackets(otherViewer.Session)
	if !hasRemoveObjectPacket(samePackets, pet.ID) {
		t.Fatal("同 ShowID 玩家應收到寵物復活 remove object")
	}
	if !hasSkillEffectPacket(samePackets, pet.ID, 3944) {
		t.Fatal("同 ShowID 玩家應收到寵物復活特效")
	}
	if hasRemoveObjectPacket(otherPackets, pet.ID) {
		t.Fatal("不同 ShowID 玩家不應收到寵物復活 remove object")
	}
	if hasSkillEffectPacket(otherPackets, pet.ID, 3944) {
		t.Fatal("不同 ShowID 玩家不應收到寵物復活特效")
	}
}
