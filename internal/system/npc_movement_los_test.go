package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcMoveTowardDoesNotMoveWithoutMapData(t *testing.T) {
	ws := world.NewState()
	npc := &world.NpcInfo{
		ID:    2001,
		Impl:  "L1Monster",
		Name:  "npc",
		X:     101,
		Y:     100,
		MapID: 900,
	}
	ws.AddNpc(npc)

	npcMoveToward(ws, npc, 103, 100, nil)

	if npc.X != 101 || npc.Y != 100 {
		t.Fatalf("沒有地圖通行資料時 NPC 不應該移動，got (%d,%d)", npc.X, npc.Y)
	}
}
