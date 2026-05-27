package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleLocalChatBroadcastsOnlySameShowLikeJava(t *testing.T) {
	cases := []struct {
		name     string
		chatType byte
		farShout bool
	}{
		{name: "normal", chatType: ChatNormal},
		{name: "shout", chatType: ChatShout, farShout: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := world.NewState()
			speakerSess := newHandlerTestSession(t, 1)
			sameSess := newHandlerTestSession(t, 2)
			otherSess := newHandlerTestSession(t, 3)
			farSess := newHandlerTestSession(t, 4)

			speaker := &world.PlayerInfo{
				SessionID: speakerSess.ID,
				Session:   speakerSess,
				CharID:    7001,
				Name:      "speaker",
				X:         100,
				Y:         100,
				MapID:     900,
				ShowID:    77,
			}
			sameShow := &world.PlayerInfo{
				SessionID: sameSess.ID,
				Session:   sameSess,
				CharID:    7002,
				Name:      "same",
				X:         101,
				Y:         100,
				MapID:     900,
				ShowID:    77,
			}
			otherShow := &world.PlayerInfo{
				SessionID: otherSess.ID,
				Session:   otherSess,
				CharID:    7003,
				Name:      "other",
				X:         101,
				Y:         100,
				MapID:     900,
				ShowID:    88,
			}
			farSameShow := &world.PlayerInfo{
				SessionID: farSess.ID,
				Session:   farSess,
				CharID:    7004,
				Name:      "far",
				X:         140,
				Y:         100,
				MapID:     900,
				ShowID:    77,
			}
			ws.AddPlayer(speaker)
			ws.AddPlayer(sameShow)
			ws.AddPlayer(otherShow)
			ws.AddPlayer(farSameShow)

			HandleChat(speakerSess, chatShowIDReader(tc.chatType, "hello"), &Deps{
				World: ws,
				Log:   zap.NewNop(),
			})

			if !hasHandlerChatPacket(drainHandlerTestPackets(sameSess), tc.chatType, speaker.CharID) {
				t.Fatalf("同 ShowID 玩家應收到 %d 聊天封包", tc.chatType)
			}
			if hasHandlerChatPacket(drainHandlerTestPackets(otherSess), tc.chatType, speaker.CharID) {
				t.Fatalf("不同 ShowID 玩家不應收到 %d 聊天封包", tc.chatType)
			}
			farPackets := drainHandlerTestPackets(farSess)
			if tc.farShout && !hasHandlerChatPacket(farPackets, tc.chatType, speaker.CharID) {
				t.Fatalf("Java 大喊 50 格內同 ShowID 玩家應收到聊天封包")
			}
			if !tc.farShout && hasHandlerChatPacket(farPackets, tc.chatType, speaker.CharID) {
				t.Fatalf("一般聊天不應送給 20 格外玩家")
			}
		})
	}
}

func chatShowIDReader(chatType byte, text string) *packet.Reader {
	w := packet.NewWriterWithOpcode(packet.C_OPCODE_CHAT)
	w.WriteC(chatType)
	w.WriteS(text)
	return packet.NewReader(w.RawBytes())
}

func hasHandlerChatPacket(packets [][]byte, chatType byte, senderID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_SAY || pkt[1] != chatType {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[2:6])) == senderID {
			return true
		}
	}
	return false
}
