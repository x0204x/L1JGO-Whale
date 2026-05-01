package system

import (
	"testing"
	"time"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestSkillCubeGroundEffectCubeIgnitionCreatesGroundEffect(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		MP:        100,
		MaxMP:     100,
		KnownSpells: []int32{
			205,
		},
		Inv: world.NewInventory(),
	})
	player.Inv.AddItem(49156, 2, "cube material", 0, 0, true, 1)
	s := newGroundEffectTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   205,
	})

	effects := ws.GetNearbyGroundEffects(player.X, player.Y, player.MapID)
	if len(effects) != 1 {
		t.Fatalf("立方：燃燒應建立 1 個地面效果，got=%d", len(effects))
	}
	if effects[0].NpcID != 80149 || effects[0].GfxID != 6706 || effects[0].Type != world.GroundEffectCubeIgnition {
		t.Fatalf("立方：燃燒效果資料錯誤: %+v", effects[0])
	}
}

func TestSkillCubeGroundEffectCubeDuplicateWithinRangeDoesNotConsumeMaterial(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		MP:        200,
		MaxMP:     200,
		KnownSpells: []int32{
			205,
		},
		Inv: world.NewInventory(),
	})
	player.Inv.AddItem(49156, 2, "cube material", 0, 0, true, 1)
	s := newGroundEffectTestSystem(t, ws)

	req := handler.SkillRequest{SessionID: player.SessionID, SkillID: 205}
	s.processSkill(req)
	player.SkillDelayUntil = time.Time{}
	s.processSkill(req)

	effects := ws.GetNearbyGroundEffects(player.X, player.Y, player.MapID)
	if len(effects) != 1 {
		t.Fatalf("3 格內已有同型立方時不應重複召喚，got=%d", len(effects))
	}
	slot := player.Inv.FindByItemID(49156)
	if slot == nil || slot.Count != 1 {
		t.Fatalf("第二次失敗召喚不應消耗材料，剩餘=%v", slot)
	}
}

func TestSkillCubeGroundEffectCubeIgnitionDamagesEnemyButNotClanAlly(t *testing.T) {
	ws := world.NewState()
	owner := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
	})
	enemy := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "enemy",
		X:         102,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	ally := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "ally",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		ClanID:    7,
	})
	ws.AddGroundEffect(&world.GroundEffect{
		ID:           world.NextGroundEffectID(),
		SkillID:      205,
		NpcID:        80149,
		GfxID:        6706,
		Type:         world.GroundEffectCubeIgnition,
		X:            100,
		Y:            100,
		MapID:        4,
		OwnerCharID:  owner.CharID,
		OwnerSession: owner.SessionID,
		OwnerClanID:  owner.ClanID,
	})
	sys := NewGroundEffectSystem(ws, &handler.Deps{World: ws, Log: zap.NewNop()})

	for i := 0; i < cubeEffectIntervalTicks; i++ {
		sys.Update(200 * time.Millisecond)
	}

	if enemy.HP != 90 {
		t.Fatalf("立方：燃燒應每輪對敵人造成 10 傷害，HP=%d", enemy.HP)
	}
	if ally.HP != 100 {
		t.Fatalf("同血盟盟友不應被燃燒立方傷害，HP=%d", ally.HP)
	}
	if ally.FireRes != 30 || !ally.HasBuff(cubeStatusIgnitionAlly) {
		t.Fatalf("同血盟盟友應取得火抗 +30 立方狀態，FireRes=%d buff=%v", ally.FireRes, ally.GetBuff(cubeStatusIgnitionAlly))
	}
}

func TestSkillCubeGroundEffectCubeBalanceRestoresMPAndDamagesTarget(t *testing.T) {
	ws := world.NewState()
	owner := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        10,
		MaxMP:     100,
	})
	ws.AddGroundEffect(&world.GroundEffect{
		ID:           world.NextGroundEffectID(),
		SkillID:      220,
		NpcID:        80152,
		GfxID:        6724,
		Type:         world.GroundEffectCubeBalance,
		X:            100,
		Y:            100,
		MapID:        4,
		OwnerCharID:  owner.CharID,
		OwnerSession: owner.SessionID,
	})
	sys := NewGroundEffectSystem(ws, &handler.Deps{World: ws, Log: zap.NewNop()})

	for i := 0; i < cubeBalanceDamageIntervalTicks; i++ {
		sys.Update(200 * time.Millisecond)
	}

	if target.MP != 15 {
		t.Fatalf("立方：和諧第 4 秒應恢復 MP 5，MP=%d", target.MP)
	}
	if target.HP != 75 {
		t.Fatalf("立方：和諧第 5 秒應造成 25 傷害，HP=%d", target.HP)
	}
}
