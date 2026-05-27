package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleActionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	sameSess := newHandlerTestSession(t, 2)
	otherSess := newHandlerTestSession(t, 3)

	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    7101,
		Name:      "actor",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	}
	sameShow := &world.PlayerInfo{
		SessionID: sameSess.ID,
		Session:   sameSess,
		CharID:    7102,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	}
	otherShow := &world.PlayerInfo{
		SessionID: otherSess.ID,
		Session:   otherSess,
		CharID:    7103,
		Name:      "other",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	}
	ws.AddPlayer(player)
	ws.AddPlayer(sameShow)
	ws.AddPlayer(otherShow)

	HandleAction(sess, actionShowIDReader(68), &Deps{World: ws, Log: zap.NewNop()})

	if !hasHandlerActionGfxPacket(drainHandlerTestPackets(sess), player.CharID, 68) {
		t.Fatal("yiwei sendPacketsAll 會把 C_ExtraCommand 動作送給施放者自己")
	}
	if !hasHandlerActionGfxPacket(drainHandlerTestPackets(sameSess), player.CharID, 68) {
		t.Fatal("同 ShowID 玩家應收到 C_ExtraCommand 動作")
	}
	if hasHandlerActionGfxPacket(drainHandlerTestPackets(otherSess), player.CharID, 68) {
		t.Fatal("不同 ShowID 玩家不應收到 C_ExtraCommand 動作")
	}
}

func TestHandleActionRejectsInvalidAndBlockedStatesLikeJava(t *testing.T) {
	cases := []struct {
		name   string
		action byte
		mutate func(*world.PlayerInfo)
	}{
		{name: "invalid_action", action: 70},
		{name: "pending_teleport", action: 68, mutate: func(p *world.PlayerInfo) { p.HasTeleport = true }},
		{name: "invisible", action: 68, mutate: func(p *world.PlayerInfo) { p.Invisible = true }},
		{name: "shape_change_disallowed_gfx", action: 68, mutate: func(p *world.PlayerInfo) {
			p.TempCharGfx = 95
			p.AddBuff(&world.ActiveBuff{SkillID: 67, TicksLeft: 10})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := world.NewState()
			sess := newHandlerTestSession(t, 11)
			viewerSess := newHandlerTestSession(t, 12)
			player := &world.PlayerInfo{
				SessionID: sess.ID,
				Session:   sess,
				CharID:    7201,
				Name:      "actor",
				X:         100,
				Y:         100,
				MapID:     900,
				ShowID:    77,
			}
			viewer := &world.PlayerInfo{
				SessionID: viewerSess.ID,
				Session:   viewerSess,
				CharID:    7202,
				Name:      "viewer",
				X:         101,
				Y:         100,
				MapID:     900,
				ShowID:    77,
			}
			if tc.mutate != nil {
				tc.mutate(player)
			}
			ws.AddPlayer(player)
			ws.AddPlayer(viewer)

			HandleAction(sess, actionShowIDReader(tc.action), &Deps{World: ws, Log: zap.NewNop()})

			if hasHandlerActionGfxPacket(drainHandlerTestPackets(sess), player.CharID, tc.action) {
				t.Fatalf("Java C_ExtraCommand 在 %s 不應送自己動作封包", tc.name)
			}
			if hasHandlerActionGfxPacket(drainHandlerTestPackets(viewerSess), player.CharID, tc.action) {
				t.Fatalf("Java C_ExtraCommand 在 %s 不應廣播動作封包", tc.name)
			}
		})
	}
}

func actionShowIDReader(action byte) *packet.Reader {
	return packet.NewReader([]byte{packet.C_OPCODE_ACTION, action})
}

func hasHandlerActionGfxPacket(packets [][]byte, objectID int32, actionCode byte) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ACTION {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID && pkt[5] == actionCode {
			return true
		}
	}
	return false
}
