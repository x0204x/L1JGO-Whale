package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestDeathRestartRebuildsGroundItemsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7800,
		Name:      "dead_player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 78,
		Session:   newSkillTestSession(t, 78),
		Known:     world.NewKnownEntities(),
		Dead:      true,
		HP:        0,
		Level:     50,
		MaxHP:     100,
		MaxMP:     50,
	})
	deps := newDeathTombDeps(t, ws, false)
	rx, ry, rmap := getBackLocation(player.MapID, deps)

	sameGround := &world.GroundItem{ID: 7801, ItemID: 1001, Count: 1, Name: "same_ground", X: rx + 1, Y: ry, MapID: rmap, ShowID: 77}
	otherGround := &world.GroundItem{ID: 7802, ItemID: 1001, Count: 1, Name: "other_ground", X: rx + 2, Y: ry, MapID: rmap, ShowID: 88}
	ws.AddGroundItem(sameGround)
	ws.AddGroundItem(otherGround)

	NewDeathSystem(deps).ProcessRestart(player.Session, player)

	if _, ok := player.Known.GroundItems[sameGround.ID]; !ok {
		t.Fatalf("重生後應感知同 ShowID 地上物")
	}
	if _, ok := player.Known.GroundItems[otherGround.ID]; ok {
		t.Fatalf("重生後不應感知不同 ShowID 地上物")
	}
	if hasPutObjectPacket(drainSkillTestPackets(player.Session), otherGround.ID) {
		t.Fatalf("重生後不應重送不同 ShowID 地上物顯示封包")
	}
}
