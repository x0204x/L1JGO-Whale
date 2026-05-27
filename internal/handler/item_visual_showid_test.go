package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestBroadcastVisualUpdateBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	playerSess := newHandlerTestSession(t, 31)
	sameSess := newHandlerTestSession(t, 32)
	otherSess := newHandlerTestSession(t, 33)

	player := &world.PlayerInfo{
		SessionID:     playerSess.ID,
		Session:       playerSess,
		CharID:        9301,
		Name:          "visual",
		X:             100,
		Y:             100,
		MapID:         900,
		ShowID:        77,
		CurrentWeapon: 4,
	}
	sameShow := &world.PlayerInfo{
		SessionID: sameSess.ID,
		Session:   sameSess,
		CharID:    9302,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	}
	otherShow := &world.PlayerInfo{
		SessionID: otherSess.ID,
		Session:   otherSess,
		CharID:    9303,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	}
	ws.AddPlayer(player)
	ws.AddPlayer(sameShow)
	ws.AddPlayer(otherShow)

	BroadcastVisualUpdate(playerSess, player, &Deps{World: ws, Log: zap.NewNop()})

	if count := countHandlerVisualPackets(drainHandlerTestPackets(playerSess), player.CharID); count != 1 {
		t.Fatalf("yiwei sendPacketsAll 會送自己一次，角色視覺更新自我封包數量 = %d", count)
	}
	if !hasHandlerVisualPacket(drainHandlerTestPackets(sameSess), player.CharID) {
		t.Fatal("同 ShowID 玩家應收到角色視覺更新")
	}
	if hasHandlerVisualPacket(drainHandlerTestPackets(otherSess), player.CharID) {
		t.Fatal("yiwei S_CharVisualUpdate 走 broadcastPacketAll，不同 ShowID 玩家不應收到角色視覺更新")
	}
}

func countHandlerVisualPackets(packets [][]byte, objectID int32) int {
	count := 0
	for _, pkt := range packets {
		if isHandlerVisualPacket(pkt, objectID) {
			count++
		}
	}
	return count
}

func hasHandlerVisualPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if isHandlerVisualPacket(pkt, objectID) {
			return true
		}
	}
	return false
}

func isHandlerVisualPacket(pkt []byte, objectID int32) bool {
	if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_CHANGE_DESC {
		return false
	}
	return int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID
}
