package system

import (
	"testing"
	"time"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestSkillGroundEffectCreationInheritsCasterShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   31,
		Session:     newSkillTestSession(t, 31),
		CharID:      3101,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		ShowID:      77,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{63},
		Inv:         world.NewInventory(),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 32,
		Session:   newSkillTestSession(t, 32),
		CharID:    3102,
		Name:      "same_show",
		X:         105,
		Y:         100,
		MapID:     4,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 33,
		Session:   newSkillTestSession(t, 33),
		CharID:    3103,
		Name:      "other_show",
		X:         105,
		Y:         100,
		MapID:     4,
		ShowID:    88,
	})
	player.Inv.AddItem(40318, 1, "magic gem", 0, 0, true, 1)
	s := newGroundEffectTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   63,
		TargetX:   105,
		TargetY:   100,
	})

	effects := ws.GetNearbyGroundEffects(105, 100, 4)
	if len(effects) != 1 {
		t.Fatalf("應建立一個生命之泉地面效果，got=%d", len(effects))
	}
	if effects[0].ShowID != player.ShowID {
		t.Fatalf("Yiwei L1SpawnUtil.spawnEffect 會讓效果繼承施法者 ShowID，got=%d want=%d", effects[0].ShowID, player.ShowID)
	}
	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), effects[0].ID) {
		t.Fatalf("同 ShowID 玩家應收到地面效果顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), effects[0].ID) {
		t.Fatalf("不同 ShowID 玩家不應收到地面效果顯示封包")
	}
}

func TestGroundEffectLifeStreamRegenSkipsDifferentShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 34,
		Session:   newSkillTestSession(t, 34),
		CharID:    3201,
		Name:      "same_show",
		X:         103,
		Y:         100,
		MapID:     4,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 35,
		Session:   newSkillTestSession(t, 35),
		CharID:    3202,
		Name:      "other_show",
		X:         103,
		Y:         100,
		MapID:     4,
		ShowID:    88,
	})
	ws.AddGroundEffect(&world.GroundEffect{
		ID:      world.NextGroundEffectID(),
		SkillID: 0,
		NpcID:   81169,
		GfxID:   2231,
		Type:    world.GroundEffectLifeStream,
		X:       100,
		Y:       100,
		MapID:   4,
		ShowID:  77,
	})
	regen := NewRegenSystem(ws, nil, nil, nil)

	if bonus := regen.lifeStreamHPBonus(sameShow); bonus != 3 {
		t.Fatalf("同 ShowID 玩家應取得生命之泉 HPR bonus，got=%d", bonus)
	}
	if bonus := regen.lifeStreamHPBonus(otherShow); bonus != 0 {
		t.Fatalf("不同 ShowID 玩家不應取得生命之泉 HPR bonus，got=%d", bonus)
	}
}

func TestGroundEffectFireWallDamageSkipsDifferentShowLikeJava(t *testing.T) {
	ws := world.NewState()
	owner := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 36,
		Session:   newSkillTestSession(t, 36),
		CharID:    3301,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    77,
		Intel:     18,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 37,
		Session:   newSkillTestSession(t, 37),
		CharID:    3302,
		Name:      "same_show",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    77,
		HP:        100,
		MaxHP:     100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 38,
		Session:   newSkillTestSession(t, 38),
		CharID:    3303,
		Name:      "other_show",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    88,
		HP:        100,
		MaxHP:     100,
	})
	ws.AddGroundEffect(&world.GroundEffect{
		ID:           world.NextGroundEffectID(),
		SkillID:      58,
		NpcID:        81157,
		GfxID:        168,
		Type:         world.GroundEffectFireWall,
		X:            101,
		Y:            100,
		MapID:        4,
		ShowID:       77,
		OwnerCharID:  owner.CharID,
		OwnerSession: owner.SessionID,
		OwnerName:    owner.Name,
		OwnerIntel:   owner.Intel,
	})
	sys := NewGroundEffectSystem(ws, &handler.Deps{World: ws, Log: zap.NewNop()})

	for i := 0; i < fireWallDamageIntervalTicks; i++ {
		sys.Update(200 * time.Millisecond)
	}

	if sameShow.HP >= 100 {
		t.Fatalf("同 ShowID 玩家應受到火牆傷害，HP=%d", sameShow.HP)
	}
	if otherShow.HP != 100 {
		t.Fatalf("不同 ShowID 玩家不應受到火牆傷害，HP=%d", otherShow.HP)
	}
}
