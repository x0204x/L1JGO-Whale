package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestVisibilityPerceivesOnlySameShowObjectsLikeJava(t *testing.T) {
	ws := world.NewState()

	viewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    4001,
		Name:      "viewer",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Known:     world.NewKnownEntities(),
	})

	samePlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    4002,
		Name:      "same_player",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherPlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    4003,
		Name:      "other_player",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	sameNpc := &world.NpcInfo{ID: 4101, NpcID: 45000, Name: "same_npc", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherNpc := &world.NpcInfo{ID: 4102, NpcID: 45000, Name: "other_npc", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddNpc(sameNpc)
	ws.AddNpc(otherNpc)

	sameSummon := &world.SummonInfo{ID: 4201, OwnerCharID: viewer.CharID, Name: "same_summon", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherSummon := &world.SummonInfo{ID: 4202, OwnerCharID: viewer.CharID, Name: "other_summon", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddSummon(sameSummon)
	ws.AddSummon(otherSummon)

	sameDoll := &world.DollInfo{ID: 4301, OwnerCharID: viewer.CharID, Name: "same_doll", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherDoll := &world.DollInfo{ID: 4302, OwnerCharID: viewer.CharID, Name: "other_doll", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddDoll(sameDoll)
	ws.AddDoll(otherDoll)

	sameHierarch := &world.HierarchInfo{ID: 4401, OwnerCharID: viewer.CharID, Name: "same_hierarch", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherHierarch := &world.HierarchInfo{ID: 4402, OwnerCharID: viewer.CharID, Name: "other_hierarch", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddHierarch(sameHierarch)
	ws.AddHierarch(otherHierarch)

	sameFollower := &world.FollowerInfo{ID: 4501, OwnerCharID: viewer.CharID, Name: "same_follower", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherFollower := &world.FollowerInfo{ID: 4502, OwnerCharID: viewer.CharID, Name: "other_follower", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddFollower(sameFollower)
	ws.AddFollower(otherFollower)

	samePet := &world.PetInfo{ID: 4601, OwnerCharID: viewer.CharID, Name: "same_pet", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherPet := &world.PetInfo{ID: 4602, OwnerCharID: viewer.CharID, Name: "other_pet", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddPet(samePet)
	ws.AddPet(otherPet)

	sameGround := &world.GroundItem{ID: 4701, ItemID: 1001, Count: 1, Name: "same_ground", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherGround := &world.GroundItem{ID: 4702, ItemID: 1001, Count: 1, Name: "other_ground", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddGroundItem(sameGround)
	ws.AddGroundItem(otherGround)

	sameGroundEffect := &world.GroundEffect{ID: 4801, NpcID: 81169, GfxID: 2231, Type: world.GroundEffectLifeStream, X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherGroundEffect := &world.GroundEffect{ID: 4802, NpcID: 81169, GfxID: 2231, Type: world.GroundEffectLifeStream, X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddGroundEffect(sameGroundEffect)
	ws.AddGroundEffect(otherGroundEffect)

	vis := NewVisibilitySystem(ws, &handler.Deps{})
	vis.Update(0)
	vis.Update(0)

	if _, ok := viewer.Known.Players[samePlayer.CharID]; !ok {
		t.Fatalf("同 ShowID 玩家應寫入 Known")
	}
	if _, ok := viewer.Known.Npcs[sameNpc.ID]; !ok {
		t.Fatalf("同 ShowID NPC 應寫入 Known")
	}
	if _, ok := viewer.Known.Summons[sameSummon.ID]; !ok {
		t.Fatalf("同 ShowID 召喚物應寫入 Known")
	}
	if _, ok := viewer.Known.Dolls[sameDoll.ID]; !ok {
		t.Fatalf("同 ShowID 娃娃應寫入 Known")
	}
	if _, ok := viewer.Known.Hierarchs[sameHierarch.ID]; !ok {
		t.Fatalf("同 ShowID 祭司應寫入 Known")
	}
	if _, ok := viewer.Known.Followers[sameFollower.ID]; !ok {
		t.Fatalf("同 ShowID 跟隨物應寫入 Known")
	}
	if _, ok := viewer.Known.Pets[samePet.ID]; !ok {
		t.Fatalf("同 ShowID 寵物應寫入 Known")
	}
	if _, ok := viewer.Known.GroundItems[sameGround.ID]; !ok {
		t.Fatalf("同 ShowID 地上物應寫入 Known")
	}
	if _, ok := viewer.Known.GroundEffects[sameGroundEffect.ID]; !ok {
		t.Fatalf("同 ShowID 地面效果應寫入 Known")
	}

	packets := drainSkillTestPackets(viewer.Session)
	for _, objectID := range []int32{
		otherPlayer.CharID,
		otherNpc.ID,
		otherSummon.ID,
		otherDoll.ID,
		otherHierarch.ID,
		otherFollower.ID,
		otherPet.ID,
		otherGround.ID,
		otherGroundEffect.ID,
	} {
		if hasPutObjectPacket(packets, objectID) {
			t.Fatalf("不同 ShowID 物件 %d 不應被感知", objectID)
		}
	}

	if _, ok := viewer.Known.Players[otherPlayer.CharID]; ok {
		t.Fatalf("不同 ShowID 玩家不應寫入 Known")
	}
	if _, ok := viewer.Known.Npcs[otherNpc.ID]; ok {
		t.Fatalf("不同 ShowID NPC 不應寫入 Known")
	}
	if _, ok := viewer.Known.Summons[otherSummon.ID]; ok {
		t.Fatalf("不同 ShowID 召喚物不應寫入 Known")
	}
	if _, ok := viewer.Known.Dolls[otherDoll.ID]; ok {
		t.Fatalf("不同 ShowID 娃娃不應寫入 Known")
	}
	if _, ok := viewer.Known.Hierarchs[otherHierarch.ID]; ok {
		t.Fatalf("不同 ShowID 祭司不應寫入 Known")
	}
	if _, ok := viewer.Known.Followers[otherFollower.ID]; ok {
		t.Fatalf("不同 ShowID 跟隨物不應寫入 Known")
	}
	if _, ok := viewer.Known.Pets[otherPet.ID]; ok {
		t.Fatalf("不同 ShowID 寵物不應寫入 Known")
	}
	if _, ok := viewer.Known.GroundItems[otherGround.ID]; ok {
		t.Fatalf("不同 ShowID 地上物不應寫入 Known")
	}
	if _, ok := viewer.Known.GroundEffects[otherGroundEffect.ID]; ok {
		t.Fatalf("不同 ShowID 地面效果不應寫入 Known")
	}
}
