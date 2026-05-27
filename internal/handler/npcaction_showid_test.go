package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func hasNpcActionShowIDPutObjectPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 9 || pkt[0] != packet.S_OPCODE_PUT_OBJECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[5:9])) == objectID {
			return true
		}
	}
	return false
}

func hasNpcActionShowIDRemoveObjectPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_REMOVE_OBJECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}

func addNpcActionShowIDPlayer(ws *world.State, p *world.PlayerInfo) *world.PlayerInfo {
	if p.Known == nil {
		p.Known = world.NewKnownEntities()
	}
	ws.AddPlayer(p)
	return p
}

func TestTeleportPlayerRebuildsOnlySameShowObjectsLikeJava(t *testing.T) {
	ws := world.NewState()

	player := addNpcActionShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    6101,
		Name:      "teleporter",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newHandlerTestSession(t, 1),
	})
	sameOldViewer := addNpcActionShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    6102,
		Name:      "same_old",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newHandlerTestSession(t, 2),
	})
	otherOldViewer := addNpcActionShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    6103,
		Name:      "other_old",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newHandlerTestSession(t, 3),
	})
	sameNewViewer := addNpcActionShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    6104,
		Name:      "same_new",
		X:         201,
		Y:         200,
		MapID:     900,
		ShowID:    77,
		SessionID: 4,
		Session:   newHandlerTestSession(t, 4),
	})
	otherNewViewer := addNpcActionShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    6105,
		Name:      "other_new",
		X:         202,
		Y:         200,
		MapID:     900,
		ShowID:    88,
		SessionID: 5,
		Session:   newHandlerTestSession(t, 5),
	})

	ownedPet := &world.PetInfo{ID: 6201, OwnerCharID: player.CharID, Name: "owned_pet", X: 101, Y: 101, MapID: 900, ShowID: 77, HP: 10, MaxHP: 10}
	ownedSummon := &world.SummonInfo{ID: 6202, OwnerCharID: player.CharID, Name: "owned_summon", NameID: "$1", X: 102, Y: 101, MapID: 900, ShowID: 77, HP: 10, MaxHP: 10}
	ownedDoll := &world.DollInfo{ID: 6203, OwnerCharID: player.CharID, Name: "owned_doll", NameID: "$1", X: 103, Y: 101, MapID: 900, ShowID: 77}
	ownedFollower := &world.FollowerInfo{ID: 6204, OwnerCharID: player.CharID, Name: "owned_follower", NameID: "$1", X: 104, Y: 101, MapID: 900, ShowID: 77}
	ownedHierarch := &world.HierarchInfo{ID: 6205, OwnerCharID: player.CharID, Name: "owned_hierarch", NameID: "$1", X: 105, Y: 101, MapID: 900, ShowID: 77}
	ws.AddPet(ownedPet)
	ws.AddSummon(ownedSummon)
	ws.AddDoll(ownedDoll)
	ws.AddFollower(ownedFollower)
	ws.AddHierarch(ownedHierarch)

	sameNpc := &world.NpcInfo{ID: 6301, NpcID: 45000, Name: "same_npc", X: 201, Y: 201, MapID: 900, ShowID: 77}
	otherNpc := &world.NpcInfo{ID: 6302, NpcID: 45000, Name: "other_npc", X: 202, Y: 201, MapID: 900, ShowID: 88}
	ws.AddNpc(sameNpc)
	ws.AddNpc(otherNpc)

	sameNearbySummon := &world.SummonInfo{ID: 6401, OwnerCharID: sameNewViewer.CharID, Name: "same_summon", NameID: "$1", X: 201, Y: 202, MapID: 900, ShowID: 77}
	otherNearbySummon := &world.SummonInfo{ID: 6402, OwnerCharID: otherNewViewer.CharID, Name: "other_summon", NameID: "$1", X: 202, Y: 202, MapID: 900, ShowID: 88}
	ws.AddSummon(sameNearbySummon)
	ws.AddSummon(otherNearbySummon)

	sameNearbyHierarch := &world.HierarchInfo{ID: 6501, OwnerCharID: sameNewViewer.CharID, Name: "same_hierarch", NameID: "$1", X: 201, Y: 203, MapID: 900, ShowID: 77}
	otherNearbyHierarch := &world.HierarchInfo{ID: 6502, OwnerCharID: otherNewViewer.CharID, Name: "other_hierarch", NameID: "$1", X: 202, Y: 203, MapID: 900, ShowID: 88}
	ws.AddHierarch(sameNearbyHierarch)
	ws.AddHierarch(otherNearbyHierarch)

	sameGround := &world.GroundItem{ID: 6601, ItemID: 1001, Count: 1, Name: "same_ground", X: 201, Y: 204, MapID: 900, ShowID: 77}
	otherGround := &world.GroundItem{ID: 6602, ItemID: 1001, Count: 1, Name: "other_ground", X: 202, Y: 204, MapID: 900, ShowID: 88}
	ws.AddGroundItem(sameGround)
	ws.AddGroundItem(otherGround)

	TeleportPlayer(player.Session, player, 200, 200, 900, 5, &Deps{World: ws})

	sameOldPackets := drainHandlerTestPackets(sameOldViewer.Session)
	for _, objectID := range []int32{player.CharID, ownedPet.ID, ownedSummon.ID, ownedDoll.ID, ownedFollower.ID, ownedHierarch.ID} {
		if !hasNpcActionShowIDRemoveObjectPacket(sameOldPackets, objectID) {
			t.Fatalf("同 ShowID 舊視野玩家應收到 %d 的 remove", objectID)
		}
	}

	otherOldPackets := drainHandlerTestPackets(otherOldViewer.Session)
	for _, objectID := range []int32{player.CharID, ownedPet.ID, ownedSummon.ID, ownedDoll.ID, ownedFollower.ID, ownedHierarch.ID} {
		if hasNpcActionShowIDRemoveObjectPacket(otherOldPackets, objectID) {
			t.Fatalf("不同 ShowID 舊視野玩家不應收到 %d 的 remove", objectID)
		}
	}

	sameNewPackets := drainHandlerTestPackets(sameNewViewer.Session)
	for _, objectID := range []int32{player.CharID, ownedPet.ID, ownedSummon.ID, ownedDoll.ID, ownedFollower.ID, ownedHierarch.ID} {
		if !hasNpcActionShowIDPutObjectPacket(sameNewPackets, objectID) {
			t.Fatalf("同 ShowID 新視野玩家應收到 %d 的 put object", objectID)
		}
	}

	otherNewPackets := drainHandlerTestPackets(otherNewViewer.Session)
	for _, objectID := range []int32{player.CharID, ownedPet.ID, ownedSummon.ID, ownedDoll.ID, ownedFollower.ID, ownedHierarch.ID} {
		if hasNpcActionShowIDPutObjectPacket(otherNewPackets, objectID) {
			t.Fatalf("不同 ShowID 新視野玩家不應收到 %d 的 put object", objectID)
		}
	}

	if _, ok := player.Known.Players[sameNewViewer.CharID]; !ok {
		t.Fatalf("傳送後應感知同 ShowID 玩家")
	}
	if _, ok := player.Known.Players[otherNewViewer.CharID]; ok {
		t.Fatalf("傳送後不應感知不同 ShowID 玩家")
	}
	if _, ok := player.Known.Npcs[sameNpc.ID]; !ok {
		t.Fatalf("傳送後應感知同 ShowID NPC")
	}
	if _, ok := player.Known.Npcs[otherNpc.ID]; ok {
		t.Fatalf("傳送後不應感知不同 ShowID NPC")
	}
	if _, ok := player.Known.Summons[sameNearbySummon.ID]; !ok {
		t.Fatalf("傳送後應感知同 ShowID 召喚物")
	}
	if _, ok := player.Known.Summons[otherNearbySummon.ID]; ok {
		t.Fatalf("傳送後不應感知不同 ShowID 召喚物")
	}
	if _, ok := player.Known.Hierarchs[sameNearbyHierarch.ID]; !ok {
		t.Fatalf("傳送後應感知同 ShowID 祭司")
	}
	if _, ok := player.Known.Hierarchs[otherNearbyHierarch.ID]; ok {
		t.Fatalf("傳送後不應感知不同 ShowID 祭司")
	}
	if _, ok := player.Known.GroundItems[sameGround.ID]; !ok {
		t.Fatalf("傳送後應感知同 ShowID 地上物")
	}
	if _, ok := player.Known.GroundItems[otherGround.ID]; ok {
		t.Fatalf("傳送後不應感知不同 ShowID 地上物")
	}
}
