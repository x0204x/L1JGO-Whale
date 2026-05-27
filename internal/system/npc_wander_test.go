package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcWanderDirectionEightPausesLikeJavaNoTarget(t *testing.T) {
	ws := world.NewState()
	npc := &world.NpcInfo{
		ID:        2002,
		Impl:      "L1Monster",
		Name:      "npc",
		X:         101,
		Y:         100,
		MapID:     900,
		MoveSpeed: 800,
	}
	ws.AddNpc(npc)

	npcWander(ws, npc, 8, nil)

	if npc.X != 101 || npc.Y != 100 {
		t.Fatalf("Java noTarget random 8-39 應停留不移動，got (%d,%d)", npc.X, npc.Y)
	}
	if npc.MoveTimer <= 0 {
		t.Fatalf("Java noTarget 停留也應進入 passive 移動節奏冷卻，MoveTimer=%d", npc.MoveTimer)
	}
	if npc.WanderDist != 0 {
		t.Fatalf("停留方向不應保留連續遊走距離，WanderDist=%d", npc.WanderDist)
	}
}
