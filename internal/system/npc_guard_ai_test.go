package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestNpcGuardSearchTargetSkipsDifferentShowWantedPlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "other_show_wanted",
		X:           101,
		Y:           100,
		MapID:       900,
		ShowID:      8,
		HP:          1000,
		MaxHP:       1000,
		WantedTicks: 100,
	})
	addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same_show_clean",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    3,
		HP:        1000,
		MaxHP:     1000,
	})
	guard := &world.NpcInfo{
		ID:         2001,
		Impl:       "L1Guard",
		Name:       "guard",
		X:          100,
		Y:          100,
		MapID:      900,
		ShowID:     3,
		SpawnX:     100,
		SpawnY:     100,
		SpawnMapID: 900,
		HP:         100,
		MaxHP:      100,
		Level:      50,
		STR:        30,
		DEX:        30,
		AtkDmg:     20,
		Ranged:     1,
	}
	ws.AddNpc(guard)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickGuardAI(guard)

	if guard.AggroTarget != 0 {
		t.Fatalf("yiwei L1GuardInstance.searchTarget() 會跳過不同 ShowID 通緝玩家，got AggroTarget=%d", guard.AggroTarget)
	}
}
