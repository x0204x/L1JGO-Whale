package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestPlayerDamagePoisonBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1101,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1102,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1103,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	})

	if !applyDamagePoisonToPlayer(target, 0, 5, &handler.Deps{World: ws}) {
		t.Fatal("player damage poison should apply")
	}

	if !hasPoisonPacket(drainSkillTestPackets(target.Session), target.CharID, 1) {
		t.Fatal("poisoned player should receive own poison packet")
	}
	if !hasPoisonPacket(drainSkillTestPackets(sameShow.Session), target.CharID, 1) {
		t.Fatal("same ShowID viewer should receive player poison packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(otherShow.Session), target.CharID, 1) {
		t.Fatal("different ShowID viewer must not receive player poison packet")
	}
}

func TestNpcDamagePoisonApplyBroadcastsX8OnlySameShowLikeJava(t *testing.T) {
	ws, npc, sameNear, sameFar, otherShow := newNpcDamagePoisonShowFixture(t)

	if !applyDamagePoisonToNpc(npc, 0, 5, &handler.Deps{World: ws}) {
		t.Fatal("NPC damage poison should apply")
	}

	if !hasPoisonPacket(drainSkillTestPackets(sameNear.Session), npc.ID, 1) {
		t.Fatal("same ShowID viewer inside 8 tiles should receive NPC poison packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(sameFar.Session), npc.ID, 1) {
		t.Fatal("same ShowID viewer outside broadcastPacketX8 range must not receive NPC poison packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(otherShow.Session), npc.ID, 1) {
		t.Fatal("different ShowID viewer must not receive NPC poison packet")
	}
}

func TestNpcDamagePoisonClearBroadcastsX8OnlySameShowLikeJava(t *testing.T) {
	ws, npc, sameNear, sameFar, otherShow := newNpcDamagePoisonShowFixture(t)
	npc.PoisonDmgAmt = 5

	tickNpcPoison(npc, ws, nil)

	if !hasPoisonPacket(drainSkillTestPackets(sameNear.Session), npc.ID, 0) {
		t.Fatal("same ShowID viewer inside 8 tiles should receive NPC poison clear packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(sameFar.Session), npc.ID, 0) {
		t.Fatal("same ShowID viewer outside broadcastPacketX8 range must not receive NPC poison clear packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(otherShow.Session), npc.ID, 0) {
		t.Fatal("different ShowID viewer must not receive NPC poison clear packet")
	}
}

func TestNpcDamagePoisonDebuffExpireBroadcastsX8OnlySameShowLikeJava(t *testing.T) {
	ws, npc, sameNear, sameFar, otherShow := newNpcDamagePoisonShowFixture(t)
	npc.PoisonDmgAmt = 5
	npc.ActiveDebuffs = map[int32]int{11: 1}

	tickNpcDebuffs(npc, ws, nil)

	if !hasPoisonPacket(drainSkillTestPackets(sameNear.Session), npc.ID, 0) {
		t.Fatal("same ShowID viewer inside 8 tiles should receive debuff-expired NPC poison clear packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(sameFar.Session), npc.ID, 0) {
		t.Fatal("same ShowID viewer outside broadcastPacketX8 range must not receive debuff-expired NPC poison clear packet")
	}
	if hasPoisonPacket(drainSkillTestPackets(otherShow.Session), npc.ID, 0) {
		t.Fatal("different ShowID viewer must not receive debuff-expired NPC poison clear packet")
	}
}

func newNpcDamagePoisonShowFixture(t *testing.T) (*world.State, *world.NpcInfo, *world.PlayerInfo, *world.PlayerInfo, *world.PlayerInfo) {
	t.Helper()
	ws := world.NewState()
	sameNear := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1201,
		Name:      "same_near",
		X:         108,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	})
	sameFar := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1202,
		Name:      "same_far",
		X:         109,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1203,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	})
	npc := &world.NpcInfo{
		ID:     2201,
		Impl:   "L1Monster",
		Name:   "poison_npc",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 77,
		HP:     100,
		MaxHP:  100,
	}
	ws.AddNpc(npc)
	return ws, npc, sameNear, sameFar, otherShow
}
