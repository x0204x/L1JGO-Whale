package handler

import (
	"encoding/binary"
	stdnet "net"
	"testing"

	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestBuildResurrectionPacketMatchesJavaPlayerFormat(t *testing.T) {
	target := &world.PlayerInfo{CharID: 1002, ClassID: 61}
	caster := &world.PlayerInfo{CharID: 1001}

	got := BuildResurrection(target, caster.CharID, 0)

	if len(got) != 16 {
		t.Fatalf("S_Resurrection 應 padding 到 16 bytes，len=%d", len(got))
	}
	if got[0] != packet.S_OPCODE_RESURRECTION {
		t.Fatalf("opcode=%d, want %d", got[0], packet.S_OPCODE_RESURRECTION)
	}
	if binary.LittleEndian.Uint32(got[1:5]) != uint32(target.CharID) {
		t.Fatalf("target id bytes 錯誤: %v", got[1:5])
	}
	if got[5] != 0 {
		t.Fatalf("type=%d, want 0", got[5])
	}
	if binary.LittleEndian.Uint32(got[6:10]) != uint32(caster.CharID) {
		t.Fatalf("use id bytes 錯誤: %v", got[6:10])
	}
	if binary.LittleEndian.Uint32(got[10:14]) != uint32(target.ClassID) {
		t.Fatalf("class id bytes 錯誤: %v", got[10:14])
	}
}

func TestResurrectionResponseSendsResurrectionPacket(t *testing.T) {
	ws := world.NewState()
	targetSess := newResurrectionTestSession(t, 2)
	casterSess := newResurrectionTestSession(t, 1)
	death := &fakeResurrectionDeathManager{}
	caster := &world.PlayerInfo{
		SessionID: 1,
		Session:   casterSess,
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	}
	target := &world.PlayerInfo{
		SessionID:        2,
		Session:          targetSess,
		CharID:           1002,
		Name:             "dead",
		X:                101,
		Y:                100,
		MapID:            4,
		ClassID:          61,
		Dead:             true,
		HP:               0,
		MaxHP:            100,
		MaxMP:            50,
		Inv:              world.NewInventory(),
		PendingResSkill:  61,
		PendingResCaster: caster.CharID,
		TombEffectID:     300001,
	}
	ws.AddPlayer(caster)
	ws.AddPlayer(target)
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}

	handleResurrectionResponse(targetSess, target, true, &Deps{
		World:     ws,
		Scripting: engine,
		Death:     death,
		Log:       zap.NewNop(),
	})
	targetSess.FlushOutput()

	if target.Dead || target.HP != 50 {
		t.Fatalf("返生術同意後應半血復活，Dead=%v HP=%d", target.Dead, target.HP)
	}
	if !outQueueHasOpcode(targetSess, packet.S_OPCODE_RESURRECTION) {
		t.Fatal("復活同意後應送 S_Resurrection")
	}
	if !death.clearTombCalled {
		t.Fatal("復活同意後應依 yiwei 先清除死亡玩家墓碑")
	}
}

type fakeResurrectionDeathManager struct {
	clearTombCalled bool
}

func (f *fakeResurrectionDeathManager) KillPlayer(_ *world.PlayerInfo) {}

func (f *fakeResurrectionDeathManager) ProcessRestart(_ *l1net.Session, _ *world.PlayerInfo) {}

func (f *fakeResurrectionDeathManager) ClearPlayerTomb(player *world.PlayerInfo) {
	f.clearTombCalled = true
	player.TombEffectID = 0
}

func newResurrectionTestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 32, 32, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func outQueueHasOpcode(sess *l1net.Session, opcode byte) bool {
	for {
		select {
		case data := <-sess.OutQueue:
			if len(data) > 0 && data[0] == opcode {
				return true
			}
		default:
			return false
		}
	}
}
