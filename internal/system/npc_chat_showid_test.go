package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcChatBroadcastUsesYiweiRangeAndShowID(t *testing.T) {
	ws := world.NewState()
	npc := &world.NpcInfo{
		ID:     9101,
		NpcID:  45000,
		Name:   "talker",
		NameID: "$talker",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 77,
	}
	ws.AddNpc(npc)
	nearSame := addSkillTestPlayer(ws, npcChatTestPlayer(t, 1, 1001, "near_same", 104, 100, 77))
	farSame := addSkillTestPlayer(ws, npcChatTestPlayer(t, 2, 1002, "far_same", 130, 100, 77))
	otherShow := addSkillTestPlayer(ws, npcChatTestPlayer(t, 3, 1003, "other_show", 104, 100, 88))
	sys := NewNpcChatSystem(ws, &handler.Deps{})

	sys.broadcastNpcChat(npc, &data.NpcChat{}, "$normal")

	if !hasNpcChatPacket(drainSkillTestPackets(nearSame.Session), npc.ID, 0) {
		t.Fatal("同 ShowID 且 8 格內玩家應收到 yiwei broadcastPacketX8 一般 NPC 聊天")
	}
	if hasNpcChatPacket(drainSkillTestPackets(farSame.Session), npc.ID, 0) {
		t.Fatal("同 ShowID 但超過 8 格玩家不應收到一般 NPC 聊天")
	}
	if hasNpcChatPacket(drainSkillTestPackets(otherShow.Session), npc.ID, 0) {
		t.Fatal("不同 ShowID 玩家不應收到一般 NPC 聊天")
	}

	sys.broadcastNpcChat(npc, &data.NpcChat{IsShout: true}, "$shout")

	if !hasNpcChatPacket(drainSkillTestPackets(farSame.Session), npc.ID, 2) {
		t.Fatal("同 ShowID 且 50 格內玩家應收到 yiwei wideBroadcastPacket 大喊 NPC 聊天")
	}
	if hasNpcChatPacket(drainSkillTestPackets(otherShow.Session), npc.ID, 2) {
		t.Fatal("不同 ShowID 玩家不應收到大喊 NPC 聊天")
	}
}

func npcChatTestPlayer(t *testing.T, sessionID uint64, charID int32, name string, x, y int32, showID int32) *world.PlayerInfo {
	t.Helper()
	return &world.PlayerInfo{
		SessionID: sessionID,
		Session:   newSkillTestSession(t, sessionID),
		CharID:    charID,
		Name:      name,
		X:         x,
		Y:         y,
		MapID:     900,
		ShowID:    showID,
	}
}

func hasNpcChatPacket(packets [][]byte, npcID int32, chatType byte) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_NPCSHOUT {
			continue
		}
		if pkt[1] == chatType && int32(binary.LittleEndian.Uint32(pkt[2:6])) == npcID {
			return true
		}
	}
	return false
}
