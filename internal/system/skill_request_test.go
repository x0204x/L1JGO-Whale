package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestResolveSkillRequestTargetIDKeepsExplicitTargetID(t *testing.T) {
	ws := world.NewState()
	s := &SkillSystem{deps: &handler.Deps{World: ws}}

	got := s.resolveSkillRequestTargetID(handler.SkillRequest{
		TargetID:   1234,
		TargetName: "Alice",
	})

	if got != 1234 {
		t.Fatalf("target id = %d, want explicit target id 1234", got)
	}
}

func TestResolveSkillRequestTargetIDUsesTargetName(t *testing.T) {
	ws := world.NewState()
	ws.AddPlayer(&world.PlayerInfo{
		SessionID: 88,
		CharID:    4321,
		Name:      "Alice",
	})
	s := &SkillSystem{deps: &handler.Deps{World: ws}}

	got := s.resolveSkillRequestTargetID(handler.SkillRequest{
		TargetName: "Alice",
	})

	if got != 4321 {
		t.Fatalf("target id = %d, want char id 4321", got)
	}
}

func TestResolveSkillRequestTargetIDReturnsZeroForMissingName(t *testing.T) {
	ws := world.NewState()
	s := &SkillSystem{deps: &handler.Deps{World: ws}}

	got := s.resolveSkillRequestTargetID(handler.SkillRequest{
		TargetName: "Missing",
	})

	if got != 0 {
		t.Fatalf("target id = %d, want 0", got)
	}
}
