package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestSkillGMClearAllStatusesInvisibilityBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8651,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Invisible: true,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8652,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8653,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})

	newSkillTestSystem(t, ws).GMClearAllStatuses(player)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("same ShowID viewer should receive GM clear revealed player put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("different ShowID viewer must not receive GM clear revealed player put object")
	}
}
