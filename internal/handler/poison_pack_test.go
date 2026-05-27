package handler

import (
	stdnet "net"
	"testing"

	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newPoisonPackTestSession(t *testing.T) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, 1, 8, 8, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func readFlushedPacket(t *testing.T, sess *l1net.Session) []byte {
	t.Helper()
	sess.FlushOutput()
	select {
	case pkt := <-sess.OutQueue:
		return pkt
	default:
		t.Fatal("預期 session 有封包輸出")
		return nil
	}
}

func readPutObjectStatus(t *testing.T, pkt []byte) byte {
	t.Helper()
	if len(pkt) == 0 || pkt[0] != packet.S_OPCODE_PUT_OBJECT {
		t.Fatalf("預期 S_PUT_OBJECT，got=%v", pkt)
	}
	r := packet.NewReader(pkt)
	_ = r.ReadH()
	_ = r.ReadH()
	_ = r.ReadD()
	_ = r.ReadH()
	_ = r.ReadC()
	_ = r.ReadC()
	_ = r.ReadC()
	_ = r.ReadC()
	_ = r.ReadD()
	_ = r.ReadH()
	_ = r.ReadS()
	_ = r.ReadS()
	return r.ReadC()
}

func readPutObjectActionStatus(t *testing.T, pkt []byte) byte {
	t.Helper()
	if len(pkt) == 0 || pkt[0] != packet.S_OPCODE_PUT_OBJECT {
		t.Fatalf("預期 S_PUT_OBJECT，got=%v", pkt)
	}
	r := packet.NewReader(pkt)
	_ = r.ReadH()
	_ = r.ReadH()
	_ = r.ReadD()
	_ = r.ReadH()
	return r.ReadC()
}

func TestSendPutObjectIncludesPoisonStatusBit(t *testing.T) {
	sess := newPoisonPackTestSession(t)
	player := &world.PlayerInfo{
		CharID:     1001,
		Name:       "target",
		X:          100,
		Y:          100,
		MapID:      4,
		PoisonType: 1,
	}

	SendPutObject(sess, player)

	status := readPutObjectStatus(t, readFlushedPacket(t, sess))
	if status&0x01 == 0 {
		t.Fatalf("中毒玩家的 S_PUT_OBJECT status 應帶 poison bit，status=0x%02x", status)
	}
}

func TestSendNpcPackIncludesPoisonStatusBit(t *testing.T) {
	sess := newPoisonPackTestSession(t)
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45213,
		NameID:       "$45213",
		Name:         "npc",
		X:            100,
		Y:            100,
		MapID:        4,
		PoisonDmgAmt: 5,
	}

	SendNpcPack(sess, npc)

	status := readPutObjectStatus(t, readFlushedPacket(t, sess))
	if status&0x01 == 0 {
		t.Fatalf("中毒 NPC 的 S_PUT_OBJECT status 應帶 poison bit，status=0x%02x", status)
	}
}

func TestSendNpcPackUsesSinkHiddenActionStatusLikeJava(t *testing.T) {
	sess := newPoisonPackTestSession(t)
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45161,
		NameID:       "$318",
		Name:         "史巴托",
		X:            100,
		Y:            100,
		MapID:        4,
		HiddenStatus: world.NpcHiddenSink,
	}

	SendNpcPack(sess, npc)

	status := readPutObjectActionStatus(t, readFlushedPacket(t, sess))
	if status != 13 {
		t.Fatalf("Java S_NPCPack 會把遁地 NPC 的 action status 寫成 13，got=%d", status)
	}
}
