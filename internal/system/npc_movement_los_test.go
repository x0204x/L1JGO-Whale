package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMoveTowardDoesNotMoveWithoutMapData(t *testing.T) {
	ws := world.NewState()
	npc := &world.NpcInfo{
		ID:    2001,
		Impl:  "L1Monster",
		Name:  "npc",
		X:     101,
		Y:     100,
		MapID: 900,
	}
	ws.AddNpc(npc)

	npcMoveToward(ws, npc, 103, 100, nil)

	if npc.X != 101 || npc.Y != 100 {
		t.Fatalf("沒有地圖通行資料時 NPC 不應該移動，got (%d,%d)", npc.X, npc.Y)
	}
}

func TestNpcExecuteMoveBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "same_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    3,
		HP:        100,
		MaxHP:     100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    8,
		HP:        100,
		MaxHP:     100,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		Impl:   "L1Monster",
		Name:   "npc",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 3,
	}
	ws.AddNpc(npc)

	npcExecuteMove(ws, npc, 101, 100, calcNpcHeading(100, 100, 101, 100), newSkillLOSTestMap(t))

	if !hasNpcMovePacket(drainSkillTestPackets(sameShow.Session), npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC 移動封包")
	}
	if hasNpcMovePacket(drainSkillTestPackets(otherShow.Session), npc.ID) {
		t.Fatal("yiwei broadcastPacketAll 只送同 ShowID，其他 ShowID 不應收到 NPC 移動封包")
	}
}

func TestNpcWanderBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "same_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    3,
		HP:        100,
		MaxHP:     100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    8,
		HP:        100,
		MaxHP:     100,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		Impl:   "L1Monster",
		Name:   "npc",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 3,
	}
	ws.AddNpc(npc)

	npcWander(ws, npc, int(calcNpcHeading(100, 100, 101, 100)), newSkillLOSTestMap(t))

	if !hasNpcMovePacket(drainSkillTestPackets(sameShow.Session), npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC 遊走移動封包")
	}
	if hasNpcMovePacket(drainSkillTestPackets(otherShow.Session), npc.ID) {
		t.Fatal("yiwei broadcastPacketAll 只送同 ShowID，其他 ShowID 不應收到 NPC 遊走移動封包")
	}
}

func TestNpcTeleportHomeBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	oldSameShow := addNpcTeleportHomeTestPlayer(t, ws, 1, 1001, "old_same_show", 130, 3)
	oldOtherShow := addNpcTeleportHomeTestPlayer(t, ws, 2, 1002, "old_other_show", 130, 8)
	newSameShow := addNpcTeleportHomeTestPlayer(t, ws, 3, 1003, "new_same_show", 100, 3)
	newOtherShow := addNpcTeleportHomeTestPlayer(t, ws, 4, 1004, "new_other_show", 100, 8)
	npc := &world.NpcInfo{
		ID:         2001,
		Impl:       "L1Monster",
		Name:       "npc",
		X:          130,
		Y:          100,
		MapID:      900,
		ShowID:     3,
		SpawnX:     100,
		SpawnY:     100,
		SpawnMapID: 900,
		HP:         1,
		MaxHP:      100,
		MP:         2,
		MaxMP:      50,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.npcTeleportHome(npc)

	oldSamePackets := drainSkillTestPackets(oldSameShow.Session)
	oldOtherPackets := drainSkillTestPackets(oldOtherShow.Session)
	newSamePackets := drainSkillTestPackets(newSameShow.Session)
	newOtherPackets := drainSkillTestPackets(newOtherShow.Session)
	if !hasRemoveObjectPacket(oldSamePackets, npc.ID) {
		t.Fatal("同 ShowID 舊位置玩家應收到 NPC 移除封包")
	}
	if hasRemoveObjectPacket(oldOtherPackets, npc.ID) {
		t.Fatal("yiwei 只通知同 ShowID，可是不同 ShowID 舊位置玩家收到 NPC 移除封包")
	}
	if !hasPutObjectPacket(newSamePackets, npc.ID) {
		t.Fatal("同 ShowID 新位置玩家應收到 NPC 顯示封包")
	}
	if hasPutObjectPacket(newOtherPackets, npc.ID) {
		t.Fatal("yiwei 只通知同 ShowID，可是不同 ShowID 新位置玩家收到 NPC 顯示封包")
	}
	if npc.HP != npc.MaxHP || npc.MP != npc.MaxMP {
		t.Fatalf("NPC 回家應重置 HP/MP，got HP=%d/%d MP=%d/%d", npc.HP, npc.MaxHP, npc.MP, npc.MaxMP)
	}
}

func TestNpcGuardTeleportHomeBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	oldSameShow := addNpcTeleportHomeTestPlayer(t, ws, 1, 1001, "old_same_show", 130, 3)
	oldOtherShow := addNpcTeleportHomeTestPlayer(t, ws, 2, 1002, "old_other_show", 130, 8)
	newSameShow := addNpcTeleportHomeTestPlayer(t, ws, 3, 1003, "new_same_show", 100, 3)
	newOtherShow := addNpcTeleportHomeTestPlayer(t, ws, 4, 1004, "new_other_show", 100, 8)
	npc := &world.NpcInfo{
		ID:         2002,
		Impl:       "L1Guard",
		Name:       "guard",
		X:          130,
		Y:          100,
		MapID:      900,
		ShowID:     3,
		SpawnX:     100,
		SpawnY:     100,
		SpawnMapID: 900,
		HP:         20,
		MaxHP:      100,
		MP:         10,
		MaxMP:      50,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.guardTeleportHome(npc)

	oldSamePackets := drainSkillTestPackets(oldSameShow.Session)
	oldOtherPackets := drainSkillTestPackets(oldOtherShow.Session)
	newSamePackets := drainSkillTestPackets(newSameShow.Session)
	newOtherPackets := drainSkillTestPackets(newOtherShow.Session)
	if !hasRemoveObjectPacket(oldSamePackets, npc.ID) {
		t.Fatal("同 ShowID 舊位置玩家應收到守衛移除封包")
	}
	if hasRemoveObjectPacket(oldOtherPackets, npc.ID) {
		t.Fatal("yiwei 只通知同 ShowID，可是不同 ShowID 舊位置玩家收到守衛移除封包")
	}
	if !hasPutObjectPacket(newSamePackets, npc.ID) {
		t.Fatal("同 ShowID 新位置玩家應收到守衛顯示封包")
	}
	if hasPutObjectPacket(newOtherPackets, npc.ID) {
		t.Fatal("yiwei 只通知同 ShowID，可是不同 ShowID 新位置玩家收到守衛顯示封包")
	}
}

func addNpcTeleportHomeTestPlayer(t *testing.T, ws *world.State, sessionID uint64, charID int32, name string, x int32, showID int32) *world.PlayerInfo {
	t.Helper()
	return addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: sessionID,
		Session:   newSkillTestSession(t, sessionID),
		CharID:    charID,
		Name:      name,
		X:         x,
		Y:         100,
		MapID:     900,
		ShowID:    showID,
		HP:        100,
		MaxHP:     100,
	})
}

func hasNpcMovePacket(packets [][]byte, npcID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_MOVE_OBJECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == npcID {
			return true
		}
	}
	return false
}

func hasNpcAttackPacket(packets [][]byte, npcID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ATTACK {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[2:6])) == npcID {
			return true
		}
	}
	return false
}

func hasRemoveObjectPacket(packets [][]byte, objectID int32) bool {
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

func hasPutObjectPacket(packets [][]byte, objectID int32) bool {
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

func hasHpMeterPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_HP_METER {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}
