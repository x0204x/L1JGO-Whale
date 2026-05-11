package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestGMCommandAdjustLawfulClampsAndSendsLawful(t *testing.T) {
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   sess,
		CharID:    1001,
		Name:      "gm",
		Lawful:    32760,
		Inv:       world.NewInventory(),
	}
	sys := NewGMCommandSystem(&handler.Deps{
		World: world.NewState(),
		Log:   zap.NewNop(),
	})

	sys.AdjustLawful(sess, player, 1000)

	if player.Lawful != 32767 {
		t.Fatalf("正義值應 clamp 到 32767: got=%d", player.Lawful)
	}
	if !player.Dirty {
		t.Fatal("調整正義值後應標記角色 Dirty")
	}

	packets := drainSkillTestPackets(sess)
	if !hasLawfulPacket(packets, player.CharID, player.Lawful) {
		t.Fatalf("未送出 S_Lawful: packets=%d", len(packets))
	}
}

func TestGMCommandAdjustLawfulClampsNegative(t *testing.T) {
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   sess,
		CharID:    1001,
		Name:      "gm",
		Lawful:    -32760,
		Inv:       world.NewInventory(),
	}
	sys := NewGMCommandSystem(&handler.Deps{
		World: world.NewState(),
		Log:   zap.NewNop(),
	})

	sys.AdjustLawful(sess, player, -1000)

	if player.Lawful != -32768 {
		t.Fatalf("正義值應 clamp 到 -32768: got=%d", player.Lawful)
	}
}

func hasLawfulPacket(packets [][]byte, charID int32, lawful int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_LAWFUL {
			continue
		}
		gotID := int32(binary.LittleEndian.Uint32(pkt[1:5]))
		gotLawful := int32(int16(binary.LittleEndian.Uint16(pkt[5:7])))
		if gotID == charID && gotLawful == lawful {
			return true
		}
	}
	return false
}
