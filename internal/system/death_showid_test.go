package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestDeathKillPlayerBroadcastsDeathActionOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7001,
		Name:      "dying_player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Known:     world.NewKnownEntities(),
		HP:        100,
		MaxHP:     100,
		Level:     50,
		MaxMP:     50,
	})
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7002,
		Name:      "same_show_death_viewer",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7003,
		Name:      "other_show_death_viewer",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	NewDeathSystem(newDeathTombDeps(t, ws, false)).KillPlayer(player)

	if !hasActionGfxPacket(drainSkillTestPackets(sameViewer.Session), player.CharID, 8) {
		t.Fatalf("同 ShowID 玩家應收到玩家死亡動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherViewer.Session), player.CharID, 8) {
		t.Fatalf("不同 ShowID 玩家不應收到玩家死亡動作")
	}
}

func TestDeathRestartRebuildsOnlySameShowObjectsLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7101,
		Name:      "dead_player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Known:     world.NewKnownEntities(),
		Dead:      true,
		HP:        0,
		Level:     50,
		MaxHP:     100,
		MaxMP:     50,
	})
	sameOldViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7102,
		Name:      "same_old",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherOldViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7103,
		Name:      "other_old",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	deps := newDeathTombDeps(t, ws, false)
	rx, ry, rmap := getBackLocation(player.MapID, deps)

	sameNewViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7104,
		Name:      "same_new",
		X:         rx + 1,
		Y:         ry,
		MapID:     rmap,
		ShowID:    77,
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
	})
	otherNewViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    7105,
		Name:      "other_new",
		X:         rx + 2,
		Y:         ry,
		MapID:     rmap,
		ShowID:    88,
		SessionID: 5,
		Session:   newSkillTestSession(t, 5),
	})

	sameNpc := &world.NpcInfo{ID: 7201, NpcID: 45000, Name: "same_npc", X: rx + 1, Y: ry + 1, MapID: rmap, ShowID: 77}
	otherNpc := &world.NpcInfo{ID: 7202, NpcID: 45000, Name: "other_npc", X: rx + 2, Y: ry + 1, MapID: rmap, ShowID: 88}
	ws.AddNpc(sameNpc)
	ws.AddNpc(otherNpc)

	sameSummon := &world.SummonInfo{ID: 7301, OwnerCharID: sameNewViewer.CharID, Name: "same_summon", NameID: "$1", X: rx + 1, Y: ry + 2, MapID: rmap, ShowID: 77}
	otherSummon := &world.SummonInfo{ID: 7302, OwnerCharID: otherNewViewer.CharID, Name: "other_summon", NameID: "$1", X: rx + 2, Y: ry + 2, MapID: rmap, ShowID: 88}
	ws.AddSummon(sameSummon)
	ws.AddSummon(otherSummon)

	sameDoll := &world.DollInfo{ID: 7401, OwnerCharID: sameNewViewer.CharID, Name: "same_doll", NameID: "$1", X: rx + 1, Y: ry + 3, MapID: rmap, ShowID: 77}
	otherDoll := &world.DollInfo{ID: 7402, OwnerCharID: otherNewViewer.CharID, Name: "other_doll", NameID: "$1", X: rx + 2, Y: ry + 3, MapID: rmap, ShowID: 88}
	ws.AddDoll(sameDoll)
	ws.AddDoll(otherDoll)

	sameHierarch := &world.HierarchInfo{ID: 7501, OwnerCharID: sameNewViewer.CharID, Name: "same_hierarch", NameID: "$1", X: rx + 1, Y: ry + 4, MapID: rmap, ShowID: 77}
	otherHierarch := &world.HierarchInfo{ID: 7502, OwnerCharID: otherNewViewer.CharID, Name: "other_hierarch", NameID: "$1", X: rx + 2, Y: ry + 4, MapID: rmap, ShowID: 88}
	ws.AddHierarch(sameHierarch)
	ws.AddHierarch(otherHierarch)

	sameFollower := &world.FollowerInfo{ID: 7601, OwnerCharID: sameNewViewer.CharID, Name: "same_follower", NameID: "$1", X: rx + 1, Y: ry + 5, MapID: rmap, ShowID: 77}
	otherFollower := &world.FollowerInfo{ID: 7602, OwnerCharID: otherNewViewer.CharID, Name: "other_follower", NameID: "$1", X: rx + 2, Y: ry + 5, MapID: rmap, ShowID: 88}
	ws.AddFollower(sameFollower)
	ws.AddFollower(otherFollower)

	samePet := &world.PetInfo{ID: 7701, OwnerCharID: sameNewViewer.CharID, Name: "same_pet", X: rx + 1, Y: ry + 6, MapID: rmap, ShowID: 77}
	otherPet := &world.PetInfo{ID: 7702, OwnerCharID: otherNewViewer.CharID, Name: "other_pet", X: rx + 2, Y: ry + 6, MapID: rmap, ShowID: 88}
	ws.AddPet(samePet)
	ws.AddPet(otherPet)

	NewDeathSystem(deps).ProcessRestart(player.Session, player)

	if !hasRemoveObjectPacket(drainSkillTestPackets(sameOldViewer.Session), player.CharID) {
		t.Fatalf("同 ShowID 舊視野玩家應收到死亡重啟 remove")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(otherOldViewer.Session), player.CharID) {
		t.Fatalf("不同 ShowID 舊視野玩家不應收到死亡重啟 remove")
	}
	if !hasPutObjectPacket(drainSkillTestPackets(sameNewViewer.Session), player.CharID) {
		t.Fatalf("同 ShowID 新視野玩家應收到死亡重啟 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherNewViewer.Session), player.CharID) {
		t.Fatalf("不同 ShowID 新視野玩家不應收到死亡重啟 put object")
	}

	if _, ok := player.Known.Players[sameNewViewer.CharID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 玩家")
	}
	if _, ok := player.Known.Players[otherNewViewer.CharID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 玩家")
	}
	if _, ok := player.Known.Npcs[sameNpc.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID NPC")
	}
	if _, ok := player.Known.Npcs[otherNpc.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID NPC")
	}
	if _, ok := player.Known.Summons[sameSummon.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 召喚物")
	}
	if _, ok := player.Known.Summons[otherSummon.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 召喚物")
	}
	if _, ok := player.Known.Dolls[sameDoll.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 娃娃")
	}
	if _, ok := player.Known.Dolls[otherDoll.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 娃娃")
	}
	if _, ok := player.Known.Hierarchs[sameHierarch.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 祭司")
	}
	if _, ok := player.Known.Hierarchs[otherHierarch.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 祭司")
	}
	if _, ok := player.Known.Followers[sameFollower.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 跟隨物")
	}
	if _, ok := player.Known.Followers[otherFollower.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 跟隨物")
	}
	if _, ok := player.Known.Pets[samePet.ID]; !ok {
		t.Fatalf("死亡重啟後應感知同 ShowID 寵物")
	}
	if _, ok := player.Known.Pets[otherPet.ID]; ok {
		t.Fatalf("死亡重啟後不應感知不同 ShowID 寵物")
	}
}
