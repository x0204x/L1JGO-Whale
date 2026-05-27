package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillIllusionistStatusArmBreakerRevealBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1011,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    77,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1012,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    77,
		HP:        100,
		MaxHP:     100,
		Invisible: true,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1013,
		Name:      "same",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1014,
		Name:      "other",
		X:         103,
		Y:         100,
		MapID:     4,
		ShowID:    88,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:     213,
		Target:      "attack",
		Type:        64,
		Ranged:      3,
		DamageValue: 15,
		ActionID:    18,
		CastGfx:     6551,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), target.CharID) {
		t.Fatalf("same ShowID viewer should receive revealed target put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), target.CharID) {
		t.Fatalf("different ShowID viewer must not receive revealed target put object")
	}
}
