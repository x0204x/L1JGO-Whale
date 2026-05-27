package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestCompanionSummonMoveBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	maps := newSkillLOSTestMap(t)

	master := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1001,
		Name:      "master",
		X:         103,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1002,
		Name:      "same",
		X:         104,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1003,
		Name:      "other",
		X:         104,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	sum := &world.SummonInfo{
		ID:          3001,
		OwnerCharID: master.CharID,
		NpcID:       45000,
		Name:        "summon",
		X:           100,
		Y:           100,
		MapID:       900,
		ShowID:      master.ShowID,
		Status:      world.SummonDefensive,
		MoveTimer:   0,
	}
	ws.AddSummon(sum)

	ai := NewCompanionAISystem(ws, &handler.Deps{MapData: maps})
	ai.tickSummons()

	if !hasNpcMovePacket(drainSkillTestPackets(sameShow.Session), sum.ID) {
		t.Fatalf("同 ShowID 玩家應收到召喚物移動封包")
	}
	if hasNpcMovePacket(drainSkillTestPackets(otherShow.Session), sum.ID) {
		t.Fatalf("不同 ShowID 玩家不應收到召喚物移動封包")
	}
}

func TestCompanionExpiredSummonRemoveBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()

	master := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1101,
		Name:      "master",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 11,
		Session:   newSkillTestSession(t, 11),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1102,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 12,
		Session:   newSkillTestSession(t, 12),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1103,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 13,
		Session:   newSkillTestSession(t, 13),
	})

	sum := &world.SummonInfo{
		ID:          3101,
		OwnerCharID: master.CharID,
		NpcID:       45000,
		Name:        "summon",
		X:           100,
		Y:           100,
		MapID:       900,
		ShowID:      master.ShowID,
		Status:      world.SummonRest,
		TimerTicks:  1,
	}
	ws.AddSummon(sum)

	ai := NewCompanionAISystem(ws, &handler.Deps{})
	ai.tickSummons()

	if !hasRemoveObjectPacket(drainSkillTestPackets(sameShow.Session), sum.ID) {
		t.Fatalf("同 ShowID 玩家應收到過期召喚物移除封包")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(otherShow.Session), sum.ID) {
		t.Fatalf("不同 ShowID 玩家不應收到過期召喚物移除封包")
	}
}

func TestCompanionDollTeleportBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()

	master := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1201,
		Name:      "master",
		X:         300,
		Y:         100,
		MapID:     901,
		ShowID:    77,
		SessionID: 21,
		Session:   newSkillTestSession(t, 21),
	})
	oldSame := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1202,
		Name:      "old_same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 22,
		Session:   newSkillTestSession(t, 22),
	})
	oldOther := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1203,
		Name:      "old_other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 23,
		Session:   newSkillTestSession(t, 23),
	})
	newSame := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1204,
		Name:      "new_same",
		X:         301,
		Y:         100,
		MapID:     901,
		ShowID:    77,
		SessionID: 24,
		Session:   newSkillTestSession(t, 24),
	})
	newOther := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1205,
		Name:      "new_other",
		X:         302,
		Y:         100,
		MapID:     901,
		ShowID:    88,
		SessionID: 25,
		Session:   newSkillTestSession(t, 25),
	})

	doll := &world.DollInfo{
		ID:          3201,
		OwnerCharID: master.CharID,
		Name:        "doll",
		NameID:      "$1",
		X:           100,
		Y:           100,
		MapID:       900,
		ShowID:      master.ShowID,
		TimerTicks:  100,
	}
	ws.AddDoll(doll)

	ai := NewCompanionAISystem(ws, &handler.Deps{})
	ai.tickDolls()

	if !hasRemoveObjectPacket(drainSkillTestPackets(oldSame.Session), doll.ID) {
		t.Fatalf("同 ShowID 舊位置玩家應收到娃娃瞬移移除封包")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(oldOther.Session), doll.ID) {
		t.Fatalf("不同 ShowID 舊位置玩家不應收到娃娃瞬移移除封包")
	}
	if !hasPutObjectPacket(drainSkillTestPackets(newSame.Session), doll.ID) {
		t.Fatalf("同 ShowID 新位置玩家應收到娃娃瞬移顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(newOther.Session), doll.ID) {
		t.Fatalf("不同 ShowID 新位置玩家不應收到娃娃瞬移顯示封包")
	}
}

func TestCompanionHierarchTeleportBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()

	master := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1301,
		Name:      "master",
		X:         300,
		Y:         100,
		MapID:     903,
		ShowID:    77,
		SessionID: 31,
		Session:   newSkillTestSession(t, 31),
	})
	oldSame := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1302,
		Name:      "old_same",
		X:         101,
		Y:         100,
		MapID:     902,
		ShowID:    77,
		SessionID: 32,
		Session:   newSkillTestSession(t, 32),
	})
	oldOther := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1303,
		Name:      "old_other",
		X:         102,
		Y:         100,
		MapID:     902,
		ShowID:    88,
		SessionID: 33,
		Session:   newSkillTestSession(t, 33),
	})
	newSame := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1304,
		Name:      "new_same",
		X:         301,
		Y:         100,
		MapID:     903,
		ShowID:    77,
		SessionID: 34,
		Session:   newSkillTestSession(t, 34),
	})
	newOther := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1305,
		Name:      "new_other",
		X:         302,
		Y:         100,
		MapID:     903,
		ShowID:    88,
		SessionID: 35,
		Session:   newSkillTestSession(t, 35),
	})

	h := &world.HierarchInfo{
		ID:          3301,
		OwnerCharID: master.CharID,
		Name:        "hierarch",
		NameID:      "$1",
		X:           100,
		Y:           100,
		MapID:       902,
		ShowID:      master.ShowID,
		TimerTicks:  100,
		BuffTimer:   100,
	}
	ws.AddHierarch(h)

	ai := NewCompanionAISystem(ws, &handler.Deps{})
	ai.tickHierarchs()

	if !hasRemoveObjectPacket(drainSkillTestPackets(oldSame.Session), h.ID) {
		t.Fatalf("同 ShowID 舊位置玩家應收到祭司瞬移移除封包")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(oldOther.Session), h.ID) {
		t.Fatalf("不同 ShowID 舊位置玩家不應收到祭司瞬移移除封包")
	}
	if !hasPutObjectPacket(drainSkillTestPackets(newSame.Session), h.ID) {
		t.Fatalf("同 ShowID 新位置玩家應收到祭司瞬移顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(newOther.Session), h.ID) {
		t.Fatalf("不同 ShowID 新位置玩家不應收到祭司瞬移顯示封包")
	}
}

func TestCompanionManualDismissBroadcastsOnlySameShowLikeJava(t *testing.T) {
	tests := []struct {
		name     string
		objectID int32
		run      func(ws *world.State, master *world.PlayerInfo, objectID int32)
	}{
		{
			name:     "summon",
			objectID: 3401,
			run: func(ws *world.State, master *world.PlayerInfo, objectID int32) {
				sum := &world.SummonInfo{
					ID:          objectID,
					OwnerCharID: master.CharID,
					NpcID:       45000,
					Name:        "summon",
					X:           100,
					Y:           100,
					MapID:       900,
					ShowID:      master.ShowID,
				}
				ws.AddSummon(sum)
				NewSummonSystem(&handler.Deps{World: ws}).DismissSummon(sum, master)
			},
		},
		{
			name:     "pet",
			objectID: 3402,
			run: func(ws *world.State, master *world.PlayerInfo, objectID int32) {
				pet := &world.PetInfo{
					ID:          objectID,
					OwnerCharID: master.CharID,
					ItemObjID:   9901,
					NpcID:       45000,
					Name:        "pet",
					X:           100,
					Y:           100,
					MapID:       900,
					ShowID:      master.ShowID,
				}
				ws.AddPet(pet)
				NewPetSystem(&handler.Deps{World: ws}).CollectPet(pet, master)
			},
		},
		{
			name:     "doll",
			objectID: 3403,
			run: func(ws *world.State, master *world.PlayerInfo, objectID int32) {
				doll := &world.DollInfo{
					ID:          objectID,
					OwnerCharID: master.CharID,
					Name:        "doll",
					NameID:      "$1",
					X:           100,
					Y:           100,
					MapID:       900,
					ShowID:      master.ShowID,
				}
				ws.AddDoll(doll)
				NewDollSystem(&handler.Deps{World: ws}).DismissDoll(doll, master)
			},
		},
		{
			name:     "hierarch",
			objectID: 3404,
			run: func(ws *world.State, master *world.PlayerInfo, objectID int32) {
				h := &world.HierarchInfo{
					ID:          objectID,
					OwnerCharID: master.CharID,
					Name:        "hierarch",
					NameID:      "$1",
					X:           100,
					Y:           100,
					MapID:       900,
					ShowID:      master.ShowID,
				}
				ws.AddHierarch(h)
				NewHierarchSystem(&handler.Deps{World: ws}).dismissHierarch(h, master)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := world.NewState()
			master := addSkillTestPlayer(ws, &world.PlayerInfo{
				CharID:    1401,
				Name:      "master",
				X:         103,
				Y:         100,
				MapID:     900,
				ShowID:    77,
				SessionID: 41,
				Session:   newSkillTestSession(t, 41),
			})
			sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
				CharID:    1402,
				Name:      "same",
				X:         101,
				Y:         100,
				MapID:     900,
				ShowID:    77,
				SessionID: 42,
				Session:   newSkillTestSession(t, 42),
			})
			otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
				CharID:    1403,
				Name:      "other",
				X:         102,
				Y:         100,
				MapID:     900,
				ShowID:    88,
				SessionID: 43,
				Session:   newSkillTestSession(t, 43),
			})

			tt.run(ws, master, tt.objectID)

			if !hasRemoveObjectPacket(drainSkillTestPackets(sameShow.Session), tt.objectID) {
				t.Fatalf("同 ShowID 玩家應收到 %s 手動解散移除封包", tt.name)
			}
			if hasRemoveObjectPacket(drainSkillTestPackets(otherShow.Session), tt.objectID) {
				t.Fatalf("不同 ShowID 玩家不應收到 %s 手動解散移除封包", tt.name)
			}
		})
	}
}

func TestCompanionPetDeathBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	master := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1501,
		Name:      "master",
		X:         103,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 51,
		Session:   newSkillTestSession(t, 51),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1502,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 52,
		Session:   newSkillTestSession(t, 52),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    1503,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 53,
		Session:   newSkillTestSession(t, 53),
	})

	pet := &world.PetInfo{
		ID:          3501,
		OwnerCharID: master.CharID,
		ItemObjID:   9902,
		NpcID:       45000,
		Name:        "pet",
		HP:          10,
		MaxHP:       10,
		X:           100,
		Y:           100,
		MapID:       900,
		ShowID:      master.ShowID,
	}
	ws.AddPet(pet)

	NewPetSystem(&handler.Deps{World: ws}).PetDie(pet)

	if !hasActionGfxPacket(drainSkillTestPackets(sameShow.Session), pet.ID, 8) {
		t.Fatalf("同 ShowID 玩家應收到寵物死亡動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherShow.Session), pet.ID, 8) {
		t.Fatalf("不同 ShowID 玩家不應收到寵物死亡動作")
	}
}
