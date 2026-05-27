package system

import (
	"time"

	coresys "github.com/l1jgo/server/internal/core/system"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

// GroundItemSystem removes expired ground items and broadcasts S_RemoveObject
// to nearby players. Phase 3 (PostUpdate).
type GroundItemSystem struct {
	world *world.State
}

func NewGroundItemSystem(ws *world.State) *GroundItemSystem {
	return &GroundItemSystem{world: ws}
}

func (s *GroundItemSystem) Phase() coresys.Phase { return coresys.PhasePostUpdate }

func (s *GroundItemSystem) Update(_ time.Duration) {
	expired := s.world.TickGroundItems()
	for _, g := range expired {
		nearby := s.world.GetNearbyPlayersInShow(g.X, g.Y, g.MapID, 0, g.ShowID)
		data := handler.BuildRemoveObject(g.ID)
		handler.BroadcastToPlayers(nearby, data)
	}
}
