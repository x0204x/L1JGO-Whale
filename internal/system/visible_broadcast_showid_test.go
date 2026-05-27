package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestFishingStartBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "fisher",
		X:         100,
		Y:         100,
		MapID:     fishingMap,
		ShowID:    10,
		Heading:   2,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-show",
		X:         101,
		Y:         100,
		MapID:     fishingMap,
		ShowID:    10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-show",
		X:         102,
		Y:         100,
		MapID:     fishingMap,
		ShowID:    99,
	})

	sys := NewFishingSystem(&handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
	})

	sys.StartFishing(player, &world.InvItem{ItemID: fishingPoleA})

	if !hasFishingActionPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatal("同 ShowID 玩家應收到釣魚動作")
	}
	if hasFishingActionPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatal("yiwei sendPacketsAll 受 ShowID 約束，不同 ShowID 玩家不應收到釣魚動作")
	}
}

func TestTrapEffectBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 11,
		Session:   newSkillTestSession(t, 11),
		CharID:    1101,
		Name:      "trap-trigger",
		X:         200,
		Y:         200,
		MapID:     4,
		ShowID:    20,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 12,
		Session:   newSkillTestSession(t, 12),
		CharID:    1102,
		Name:      "same-show",
		X:         201,
		Y:         200,
		MapID:     4,
		ShowID:    20,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 13,
		Session:   newSkillTestSession(t, 13),
		CharID:    1103,
		Name:      "other-show",
		X:         202,
		Y:         200,
		MapID:     4,
		ShowID:    99,
	})

	sys := NewTrapSystem(&handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
	})
	trap := &world.TrapInstance{
		Template: &data.TrapTemplate{GfxID: 2231},
		X:        200,
		Y:        200,
		MapID:    4,
		Alive:    true,
	}

	sys.sendTrapEffect(player.Session, player, trap)

	if !hasEffectLocationPacket(drainSkillTestPackets(sameShow.Session), trap.X, trap.Y, trap.Template.GfxID) {
		t.Fatal("同 ShowID 玩家應收到陷阱地點特效")
	}
	if hasEffectLocationPacket(drainSkillTestPackets(otherShow.Session), trap.X, trap.Y, trap.Template.GfxID) {
		t.Fatal("yiwei broadcastPacketAll 受 ShowID 約束，不同 ShowID 玩家不應收到陷阱地點特效")
	}
}

func TestNpcTeleportDepartureEffectBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 21,
		Session:   newSkillTestSession(t, 21),
		CharID:    2101,
		Name:      "npc-teleport",
		X:         300,
		Y:         300,
		MapID:     4,
		ShowID:    30,
		Inv:       world.NewInventory(),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 22,
		Session:   newSkillTestSession(t, 22),
		CharID:    2102,
		Name:      "same-show",
		X:         301,
		Y:         300,
		MapID:     4,
		ShowID:    30,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 23,
		Session:   newSkillTestSession(t, 23),
		CharID:    2103,
		Name:      "other-show",
		X:         302,
		Y:         300,
		MapID:     4,
		ShowID:    99,
	})
	sys := NewNpcServiceSystem(&handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
	})

	sys.NpcTeleportWithCost(player.Session, player, &data.TeleportDest{
		X:     33000,
		Y:     33001,
		MapID: 4,
	}, 0)

	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), player.CharID, 169) {
		t.Fatal("same ShowID viewer should receive npc teleport departure effect like yiwei broadcastPacketAll")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), player.CharID, 169) {
		t.Fatal("other ShowID viewer must not receive npc teleport departure effect")
	}
}

func TestPetHealPotionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 31,
		Session:   newSkillTestSession(t, 31),
		CharID:    3101,
		Name:      "pet-owner",
		X:         400,
		Y:         400,
		MapID:     4,
		ShowID:    40,
		Inv:       world.NewInventory(),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 32,
		Session:   newSkillTestSession(t, 32),
		CharID:    3102,
		Name:      "same-show",
		X:         401,
		Y:         400,
		MapID:     4,
		ShowID:    40,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 33,
		Session:   newSkillTestSession(t, 33),
		CharID:    3103,
		Name:      "other-show",
		X:         402,
		Y:         400,
		MapID:     4,
		ShowID:    99,
	})
	pet := &world.PetInfo{
		ID:          5101,
		OwnerCharID: player.CharID,
		NpcID:       45034,
		HP:          10,
		MaxHP:       100,
		X:           400,
		Y:           400,
		MapID:       4,
		ShowID:      40,
	}
	item := player.Inv.AddItemWithID(8101, 40010, 1, "red potion", 0, 0, true, 1)
	sys := NewPetSystem(&handler.Deps{
		World:    ws,
		PetTypes: loadVisibleBroadcastPetTypes(t),
		Log:      zap.NewNop(),
	})

	sys.GiveToPet(player.Session, player, pet, item)

	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), pet.ID, 189) {
		t.Fatal("same ShowID viewer should receive pet heal potion effect like yiwei broadcastPacketAll")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), pet.ID, 189) {
		t.Fatal("other ShowID viewer must not receive pet heal potion effect")
	}
}

func TestPetHastePotionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 41,
		Session:   newSkillTestSession(t, 41),
		CharID:    4101,
		Name:      "pet-owner",
		X:         500,
		Y:         500,
		MapID:     4,
		ShowID:    50,
		Inv:       world.NewInventory(),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 42,
		Session:   newSkillTestSession(t, 42),
		CharID:    4102,
		Name:      "same-show",
		X:         501,
		Y:         500,
		MapID:     4,
		ShowID:    50,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 43,
		Session:   newSkillTestSession(t, 43),
		CharID:    4103,
		Name:      "other-show",
		X:         502,
		Y:         500,
		MapID:     4,
		ShowID:    99,
	})
	pet := &world.PetInfo{
		ID:          5201,
		OwnerCharID: player.CharID,
		NpcID:       45034,
		HP:          100,
		MaxHP:       100,
		X:           500,
		Y:           500,
		MapID:       4,
		ShowID:      50,
	}
	item := player.Inv.AddItemWithID(8201, 40013, 1, "green potion", 0, 0, true, 1)
	sys := NewPetSystem(&handler.Deps{
		World:    ws,
		PetTypes: loadVisibleBroadcastPetTypes(t),
		Log:      zap.NewNop(),
	})

	sys.GiveToPet(player.Session, player, pet, item)

	samePackets := drainSkillTestPackets(sameShow.Session)
	if !hasVisualShowIDPacket(samePackets, packet.S_OPCODE_SPEED, pet.ID) {
		t.Fatal("same ShowID viewer should receive pet haste packet like yiwei broadcastPacketAll")
	}
	if !hasSkillEffectPacket(samePackets, pet.ID, 191) {
		t.Fatal("same ShowID viewer should receive pet haste sound like yiwei broadcastPacketAll")
	}

	otherPackets := drainSkillTestPackets(otherShow.Session)
	if hasVisualShowIDPacket(otherPackets, packet.S_OPCODE_SPEED, pet.ID) {
		t.Fatal("other ShowID viewer must not receive pet haste packet")
	}
	if hasSkillEffectPacket(otherPackets, pet.ID, 191) {
		t.Fatal("other ShowID viewer must not receive pet haste sound")
	}
}

func hasFishingActionPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ACTION {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID && pkt[5] == 71 {
			return true
		}
	}
	return false
}

func loadVisibleBroadcastPetTypes(t *testing.T) *data.PetTypeTable {
	t.Helper()
	petTypes, err := data.LoadPetTypeTable("../../data/yaml/pet_types.yaml")
	if err != nil {
		t.Fatalf("load pet types: %v", err)
	}
	return petTypes
}

func hasEffectLocationPacket(packets [][]byte, x, y, gfxID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_EFFECTLOCATION {
			continue
		}
		px := int32(binary.LittleEndian.Uint16(pkt[1:3]))
		py := int32(binary.LittleEndian.Uint16(pkt[3:5]))
		pgfx := int32(binary.LittleEndian.Uint16(pkt[5:7]))
		if px == x && py == y && pgfx == gfxID {
			return true
		}
	}
	return false
}
