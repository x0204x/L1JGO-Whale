package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

type captureSkillManager struct {
	reqs []SkillRequest
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

func (m *captureSkillManager) CancelAbsoluteBarrier(_ *world.PlayerInfo) {}

func (m *captureSkillManager) CancelInvisibility(_ *world.PlayerInfo) {}

func (m *captureSkillManager) ApplyGMBuff(_ *world.PlayerInfo, _ int32) bool { return false }

func (m *captureSkillManager) RevertBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

func (m *captureSkillManager) ConsumeSkillResources(_ *net.Session, _ *world.PlayerInfo, _ *data.SkillInfo) {
}

func (m *captureSkillManager) ApplyBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

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
