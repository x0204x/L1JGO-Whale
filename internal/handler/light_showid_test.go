package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestUpdatePlayerLightBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws, player, sameShow, otherShow := newLightShowIDWorld(t, false)

	UpdatePlayerLight(player, ws)

	if !hasHandlerLightPacket(drainHandlerTestPackets(player.Session), player.CharID, 14) {
		t.Fatal("日光術光源變更應先送給自己")
	}
	if !hasHandlerLightPacket(drainHandlerTestPackets(sameShow.Session), player.CharID, 14) {
		t.Fatal("同 ShowID 玩家應收到玩家光源變更")
	}
	if hasHandlerLightPacket(drainHandlerTestPackets(otherShow.Session), player.CharID, 14) {
		t.Fatal("yiwei S_Light 走 broadcastPacketAll，不同 ShowID 玩家不應收到玩家光源變更")
	}
}

func TestUpdatePlayerLightSkipsInvisibleBroadcastLikeJava(t *testing.T) {
	ws, player, sameShow, otherShow := newLightShowIDWorld(t, true)

	UpdatePlayerLight(player, ws)

	if !hasHandlerLightPacket(drainHandlerTestPackets(player.Session), player.CharID, 14) {
		t.Fatal("隱身玩家光源變更仍應送給自己")
	}
	if hasHandlerLightPacket(drainHandlerTestPackets(sameShow.Session), player.CharID, 14) {
		t.Fatal("yiwei 隱身玩家 turnOnOffLight 不會廣播 S_Light")
	}
	if hasHandlerLightPacket(drainHandlerTestPackets(otherShow.Session), player.CharID, 14) {
		t.Fatal("不同 ShowID 玩家不應收到隱身玩家光源變更")
	}
}

func newLightShowIDWorld(t *testing.T, invisible bool) (*world.State, *world.PlayerInfo, *world.PlayerInfo, *world.PlayerInfo) {
	t.Helper()
	ws := world.NewState()
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newHandlerTestSession(t, 1),
		CharID:    9501,
		Name:      "light",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		Invisible: invisible,
	}
	player.AddBuff(&world.ActiveBuff{SkillID: 2})
	sameShow := &world.PlayerInfo{
		SessionID: 2,
		Session:   newHandlerTestSession(t, 2),
		CharID:    9502,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	}
	otherShow := &world.PlayerInfo{
		SessionID: 3,
		Session:   newHandlerTestSession(t, 3),
		CharID:    9503,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	}
	ws.AddPlayer(player)
	ws.AddPlayer(sameShow)
	ws.AddPlayer(otherShow)
	return ws, player, sameShow, otherShow
}

func hasHandlerLightPacket(packets [][]byte, objectID int32, light byte) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_CHANGE_LIGHT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID && pkt[5] == light {
			return true
		}
	}
	return false
}
