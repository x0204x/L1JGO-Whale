package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newPoisonVisualTestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 8, 8, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func newPoisonVisualTestPlayer(t *testing.T, ws *world.State, poisonType byte, ticks int, paralyzed bool) *world.PlayerInfo {
	t.Helper()
	player := &world.PlayerInfo{
		SessionID:       1,
		Session:         newPoisonVisualTestSession(t, 1),
		CharID:          1001,
		Name:            "target",
		X:               100,
		Y:               100,
		MapID:           4,
		HP:              100,
		MaxHP:           100,
		PoisonType:      poisonType,
		PoisonTicksLeft: ticks,
		Paralyzed:       paralyzed,
	}
	ws.AddPlayer(player)
	return player
}

func drainPoisonTestPackets(p *world.PlayerInfo) [][]byte {
	p.Session.FlushOutput()
	var packets [][]byte
	for {
		select {
		case pkt := <-p.Session.OutQueue:
			packets = append(packets, pkt)
		default:
			return packets
		}
	}
}

func hasParalysisSubtype(packets [][]byte, subtype byte) bool {
	for _, pkt := range packets {
		if len(pkt) >= 2 && pkt[0] == packet.S_OPCODE_PARALYSIS && pkt[1] == subtype {
			return true
		}
	}
	return false
}

func TestParalysisPoisonUsesYiweiApplySubtype(t *testing.T) {
	ws := world.NewState()
	target := newPoisonVisualTestPlayer(t, ws, 3, 1, false)
	deps := &handler.Deps{World: ws}

	TickPlayerPoison(target, deps)

	packets := drainPoisonTestPackets(target)
	if !hasParalysisSubtype(packets, handler.ParalysisApply) {
		t.Fatalf("毒性麻痺進入麻痺期應依 yiwei 送 S_Paralysis subtype 0x02，packets=%v", packets)
	}
	if hasParalysisSubtype(packets, handler.ParalysisMobApply) {
		t.Fatalf("毒性麻痺不應送 TYPE_PARALYSIS2 subtype 0x04，packets=%v", packets)
	}
}

func TestParalysisPoisonUsesYiweiRemoveSubtype(t *testing.T) {
	ws := world.NewState()
	target := newPoisonVisualTestPlayer(t, ws, 4, 0, true)
	deps := &handler.Deps{World: ws}

	CurePoison(target, deps)

	packets := drainPoisonTestPackets(target)
	if !hasParalysisSubtype(packets, handler.ParalysisRemove) {
		t.Fatalf("毒性麻痺解除應依 yiwei 送 S_Paralysis subtype 0x03，packets=%v", packets)
	}
	if hasParalysisSubtype(packets, handler.ParalysisMobRemove) {
		t.Fatalf("毒性麻痺解除不應送 TYPE_PARALYSIS2 subtype 0x05，packets=%v", packets)
	}
}
