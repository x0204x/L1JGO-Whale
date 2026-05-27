package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func newNpcAITargetTestMonster(sessionID uint64) *world.NpcInfo {
	return &world.NpcInfo{
		ID:          2001,
		NpcID:       45161,
		Impl:        "L1Monster",
		Name:        "target_test_monster",
		X:           101,
		Y:           100,
		MapID:       900,
		SpawnX:      101,
		SpawnY:      100,
		SpawnMapID:  900,
		HP:          100,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		AtkDmg:      20,
		Ranged:      1,
		Agro:        true,
		AggroTarget: sessionID,
		MoveTimer:   10,
	}
}

func TestNpcAggroScanSkipsInvisiblePlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "invisible",
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
		Invisible: true,
	})
	npc := newNpcAITargetTestMonster(0)
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)

	if npc.AggroTarget == target.SessionID {
		t.Fatal("yiwei searchTarget 會略過一般隱身玩家，怪物不應自動鎖定")
	}
}

func TestNpcAggroScanSkipsGMPlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm",
		X:           100,
		Y:           100,
		MapID:       900,
		HP:          5000,
		MaxHP:       5000,
		AccessLevel: 200,
	})
	npc := newNpcAITargetTestMonster(0)
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	s.deps.MapData.SetImpassable(900, 102, 100, true)

	s.tickMonsterAI(npc)

	if npc.AggroTarget == target.SessionID {
		t.Fatal("yiwei searchTarget 會略過 GM，怪物不應自動鎖定 GM")
	}
}

func TestNpcAggroScanSkipsDifferentShowIDLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "other_show",
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
		ShowID:    100,
	})
	npc := newNpcAITargetTestMonster(0)
	npc.ShowID = 200
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)

	if npc.AggroTarget == target.SessionID {
		t.Fatal("yiwei searchTarget 要求 showId 相同，怪物不應鎖定不同副本玩家")
	}
}

func TestNpcAggroTargetClearsDifferentShowIDLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "other_show",
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
		ShowID:    100,
	})
	npc := newNpcAITargetTestMonster(target.SessionID)
	npc.ShowID = 200
	npc.HateList = map[uint64]int32{target.SessionID: 10}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)

	if npc.AggroTarget != 0 {
		t.Fatalf("yiwei checkTarget 會清掉 showId 不同的既有目標，AggroTarget=%d", npc.AggroTarget)
	}
	if _, ok := npc.HateList[target.SessionID]; ok {
		t.Fatal("showId 不同的既有目標應從 hate list 移除")
	}
}

func TestNpcAggroScanSkipsUnreachablePlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "behind_wall",
		X:         103,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := newNpcAITargetTestMonster(0)
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)

	if npc.AggroTarget == target.SessionID {
		t.Fatal("yiwei searchTarget 會先以 moveDirection 排除不可到達目標，怪物不應隔牆鎖定")
	}
}
