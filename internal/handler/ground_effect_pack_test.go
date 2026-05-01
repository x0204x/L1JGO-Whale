package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestSendGroundEffectPackUsesOwnerNameAndLawfulForTomb(t *testing.T) {
	sess := newResurrectionTestSession(t, 1)
	tomb := &world.GroundEffect{
		ID:        300001,
		NpcID:     world.TombEffectNpcID,
		GfxID:     world.TombEffectGfxID,
		Type:      world.GroundEffectTomb,
		X:         32767,
		Y:         32768,
		MapID:     4,
		OwnerName: "dead",
		Lawful:    1234,
	}

	SendGroundEffectPack(sess, tomb)
	sess.FlushOutput()

	data := <-sess.OutQueue
	if data[0] != packet.S_OPCODE_PUT_OBJECT {
		t.Fatalf("墓碑應使用 S_NPCPack_Eff 相容的 PUT_OBJECT，opcode=%d", data[0])
	}
	if binary.LittleEndian.Uint16(data[19:21]) != uint16(1234) {
		t.Fatalf("墓碑 lawful 應等於死亡玩家 lawful，bytes=%v", data[19:21])
	}
	if got := string(data[21:25]); got != "dead" {
		t.Fatalf("墓碑名稱應顯示死亡玩家名稱，got=%q", got)
	}
}
