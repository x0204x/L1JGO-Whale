package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestAddPlayerHateLinksFamilyAgroFamilyLikeJava(t *testing.T) {
	const attackerSID uint64 = 100
	const otherSID uint64 = 200

	ws := world.NewState()
	player := &world.PlayerInfo{SessionID: attackerSID, X: 100, Y: 100, MapID: 4}
	ws.AddPlayer(player)

	source := &world.NpcInfo{ID: 200_000_010, NpcID: 45001, X: 101, Y: 100, MapID: 4, ShowID: 7, Family: "orc"}
	sameFamily := &world.NpcInfo{ID: 200_000_011, NpcID: 45002, X: 102, Y: 100, MapID: 4, ShowID: 7, Family: "orc", AgroFamily: 1}
	otherFamily := &world.NpcInfo{ID: 200_000_012, NpcID: 45003, X: 103, Y: 100, MapID: 4, ShowID: 7, Family: "wolf", AgroFamily: 1}
	globalFamily := &world.NpcInfo{ID: 200_000_013, NpcID: 45004, X: 104, Y: 100, MapID: 4, ShowID: 7, AgroFamily: 2}
	busy := &world.NpcInfo{
		ID:          200_000_014,
		NpcID:       45005,
		X:           105,
		Y:           100,
		MapID:       4,
		ShowID:      7,
		Family:      "orc",
		AgroFamily:  1,
		AggroTarget: otherSID,
		HateList:    map[uint64]int32{otherSID: 5},
	}
	otherShow := &world.NpcInfo{ID: 200_000_015, NpcID: 45006, X: 106, Y: 100, MapID: 4, ShowID: 9, Family: "orc", AgroFamily: 1}
	emptyFamily := &world.NpcInfo{ID: 200_000_016, NpcID: 45007, X: 107, Y: 100, MapID: 4, ShowID: 7, AgroFamily: 1}

	for _, npc := range []*world.NpcInfo{source, sameFamily, otherFamily, globalFamily, busy, otherShow, emptyFamily} {
		ws.AddNpc(npc)
	}

	AddPlayerHateLikeJava(ws, source, player, 30)

	if source.HateList[attackerSID] != 30 || source.AggroTarget != attackerSID {
		t.Fatalf("被攻擊 NPC 應取得直接 hate：hate=%v target=%d", source.HateList, source.AggroTarget)
	}
	if sameFamily.HateList[attackerSID] != 0 || sameFamily.AggroTarget != attackerSID {
		t.Fatalf("同族 agrofamily=1 NPC 應以 0 hate 連動：hate=%v target=%d", sameFamily.HateList, sameFamily.AggroTarget)
	}
	if globalFamily.HateList[attackerSID] != 0 || globalFamily.AggroTarget != attackerSID {
		t.Fatalf("agrofamily>1 NPC 應以 0 hate 連動：hate=%v target=%d", globalFamily.HateList, globalFamily.AggroTarget)
	}
	if otherFamily.HateList != nil || otherFamily.AggroTarget != 0 {
		t.Fatalf("不同 family 的 agrofamily=1 NPC 不應連動：hate=%v target=%d", otherFamily.HateList, otherFamily.AggroTarget)
	}
	if busy.HateList[otherSID] != 5 || busy.AggroTarget != otherSID {
		t.Fatalf("已有 hate 的 NPC 不應被覆蓋：hate=%v target=%d", busy.HateList, busy.AggroTarget)
	}
	if otherShow.HateList != nil || otherShow.AggroTarget != 0 {
		t.Fatalf("不同 ShowID NPC 不應連動：hate=%v target=%d", otherShow.HateList, otherShow.AggroTarget)
	}
	if emptyFamily.HateList != nil || emptyFamily.AggroTarget != 0 {
		t.Fatalf("agrofamily=1 但 family 空白的 NPC 不應連動：hate=%v target=%d", emptyFamily.HateList, emptyFamily.AggroTarget)
	}
}
