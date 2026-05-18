package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

type captureSkillManager struct {
	reqs                       []SkillRequest
	cancelAbsoluteBarrierCalls int
}

func (m *captureSkillManager) QueueSkill(req SkillRequest) {
	m.reqs = append(m.reqs, req)
}

func (m *captureSkillManager) CancelAllBuffs(_ *world.PlayerInfo) {}

func (m *captureSkillManager) ClearAllBuffsOnDeath(_ *world.PlayerInfo) {}

func (m *captureSkillManager) GMClearAllStatuses(_ *world.PlayerInfo) {}

func (m *captureSkillManager) TickPlayerBuffs(_ *world.PlayerInfo) {}

func (m *captureSkillManager) RemoveBuffAndRevert(_ *world.PlayerInfo, _ int32) {}

func (m *captureSkillManager) ApplyNpcDebuff(_ *world.PlayerInfo, _ *data.SkillInfo) {}

func (m *captureSkillManager) CancelAbsoluteBarrier(_ *world.PlayerInfo) {
	m.cancelAbsoluteBarrierCalls++
}

func (m *captureSkillManager) CancelInvisibility(_ *world.PlayerInfo) {}

func (m *captureSkillManager) ApplyGMBuff(_ *world.PlayerInfo, _ int32) bool { return false }

func (m *captureSkillManager) RevertBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

func (m *captureSkillManager) ConsumeSkillResources(_ *net.Session, _ *world.PlayerInfo, _ *data.SkillInfo) {
}

func (m *captureSkillManager) ApplyBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

func TestHandleUseSpellBlockedByShockStunBeforeSkillQueueLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "stunned",
		Paralyzed: true,
		Inv:       world.NewInventory(),
	}
	player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會以 isParalyzedX 擋下 SHOCK_STUN 中的施法，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
	if !hasHandlerServerMessage(drainHandlerTestPackets(sess), 285) {
		t.Fatal("Java C_UseSkill 在 isParalyzedX 擋下施法時會送 S_ServerMessage(285)")
	}
}

func TestHandleUseSpellBlockedByPrivateShopLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID:   sess.ID,
		Session:     sess,
		CharID:      1001,
		Name:        "seller",
		PrivateShop: true,
		Inv:         world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會在 pc.isPrivateShop() 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellParsesTeleportBookmark(t *testing.T) {
	req := parseUseSpellForTest(69, func(w *packet.Writer) {
		w.WriteH(4)
		w.WriteD(98765)
	})

	if req.TargetID != 98765 {
		t.Fatalf("TargetID = %d, want bookmark id 98765", req.TargetID)
	}
	if req.BookmarkID != 98765 {
		t.Fatalf("BookmarkID = %d, want 98765", req.BookmarkID)
	}
	if req.MapID != 4 {
		t.Fatalf("MapID = %d, want 4", req.MapID)
	}
}

func TestHandleUseSpellParsesTrueTargetTextAndPosition(t *testing.T) {
	req := parseUseSpellForTest(113, func(w *packet.Writer) {
		w.WriteD(1234)
		w.WriteH(33010)
		w.WriteH(33011)
		w.WriteS("mark")
	})

	if req.TargetID != 1234 {
		t.Fatalf("TargetID = %d, want 1234", req.TargetID)
	}
	if req.TargetX != 33010 || req.TargetY != 33011 {
		t.Fatalf("Target position = (%d,%d), want (33010,33011)", req.TargetX, req.TargetY)
	}
	if req.Text != "mark" {
		t.Fatalf("Text = %q, want %q", req.Text, "mark")
	}
}

func TestHandleUseSpellParsesGroundTargetPosition(t *testing.T) {
	req := parseUseSpellForTest(58, func(w *packet.Writer) {
		w.WriteH(32768)
		w.WriteH(32769)
	})

	if req.TargetX != 32768 || req.TargetY != 32769 {
		t.Fatalf("Target position = (%d,%d), want (32768,32769)", req.TargetX, req.TargetY)
	}
}

func TestHandleUseSpellParsesClanTargetName(t *testing.T) {
	req := parseUseSpellForTest(116, func(w *packet.Writer) {
		w.WriteS("Alice[Clan]")
	})

	if req.TargetName != "Alice" {
		t.Fatalf("TargetName = %q, want %q", req.TargetName, "Alice")
	}
}

func TestHandleUseSpellPreservesSummonSelectionValue(t *testing.T) {
	req := parseUseSpellForTest(51, func(w *packet.Writer) {
		w.WriteD(812)
	})

	if req.TargetID != 812 {
		t.Fatalf("TargetID = %d, want 812", req.TargetID)
	}
	if req.SummonID != 812 {
		t.Fatalf("SummonID = %d, want 812", req.SummonID)
	}
}

func parseUseSpellForTest(skillID int32, writeRest func(*packet.Writer)) SkillRequest {
	row := byte((skillID - 1) / 8)
	column := byte((skillID - 1) % 8)

	w := packet.NewWriterWithOpcode(6)
	w.WriteC(row)
	w.WriteC(column)
	writeRest(w)

	mgr := &captureSkillManager{}
	deps := &Deps{Skill: mgr}
	sess := &net.Session{ID: 777}
	HandleUseSpell(sess, packet.NewReader(w.RawBytes()), deps)

	if len(mgr.reqs) != 1 {
		panic("expected exactly one queued skill request")
	}
	return mgr.reqs[0]
}

func useSpellReader(skillID int32) *packet.Reader {
	row := byte((skillID - 1) / 8)
	column := byte((skillID - 1) % 8)

	w := packet.NewWriterWithOpcode(packet.C_OPCODE_USE_SPELL)
	w.WriteC(row)
	w.WriteC(column)
	w.WriteD(0)
	w.WriteH(0)
	w.WriteH(0)
	return packet.NewReader(w.RawBytes())
}

func hasHandlerServerMessage(packets [][]byte, msgID uint16) bool {
	for _, pkt := range packets {
		if len(pkt) >= 3 && pkt[0] == packet.S_OPCODE_MESSAGE_CODE &&
			binary.LittleEndian.Uint16(pkt[1:3]) == msgID {
			return true
		}
	}
	return false
}
