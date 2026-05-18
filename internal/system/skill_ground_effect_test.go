package system

import (
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newGroundEffectTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	npcs, err := data.LoadNpcTable("../../data/yaml/npc_list.yaml")
	if err != nil {
		t.Fatalf("讀取 NPC 資料失敗: %v", err)
	}
	return &SkillSystem{deps: &handler.Deps{
		World:  ws,
		Skills: skills,
		Npcs:   npcs,
		Log:    zap.NewNop(),
	}}
}

func TestSkillGroundEffectLifeStreamCreatesGroundEffect(t *testing.T) {
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
			63,
		},
		Inv: world.NewInventory(),
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
		t.Fatalf("生命之泉應建立 1 個地面效果，got=%d", len(effects))
	}
	if effects[0].NpcID != 81169 || effects[0].GfxID != 2231 || effects[0].Type != world.GroundEffectLifeStream {
		t.Fatalf("生命之泉效果資料錯誤: %+v", effects[0])
	}
}

func TestSkillGroundEffectLifeStreamStoresJavaEffectSkillIDZero(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{63},
		Inv:         world.NewInventory(),
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
		t.Fatalf("生命之泉應建立 1 個地面效果，got=%d", len(effects))
	}
	if effects[0].SkillID != 0 {
		t.Fatalf("Java 生命之泉 spawnEffect skill id 應為 0，got=%d", effects[0].SkillID)
	}
}

func TestSkillGroundEffectFireWallCreatesLineGroundEffects(t *testing.T) {
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
			58,
		},
		Inv: world.NewInventory(),
	})
	s := newGroundEffectTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   58,
		TargetX:   108,
		TargetY:   100,
	})

	effects := ws.GetNearbyGroundEffects(104, 100, 4)
	if len(effects) != 8 {
		t.Fatalf("火牢應沿方向建立 8 格效果，got=%d", len(effects))
	}
	if !ws.HasGroundEffectAt(101, 100, 4, 81157) {
		t.Fatal("火牢第一格應出現在施法者前方")
	}
}

func TestSkillGroundEffectFireWallRecalculatesDirectionFromLastEffect(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{58},
		Inv:         world.NewInventory(),
	})
	s := newGroundEffectTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   58,
		TargetX:   103,
		TargetY:   101,
	})

	if !ws.HasGroundEffectAt(101, 101, 4, 81157) {
		t.Fatal("火牢第一格應依施法者到目標方向生成在東南格")
	}
	if !ws.HasGroundEffectAt(102, 101, 4, 81157) {
		t.Fatal("Java 火牢第二格會從上一個效果重新朝目標計算方向，不應固定斜線到 102,102")
	}
	if ws.HasGroundEffectAt(102, 102, 4, 81157) {
		t.Fatal("Go 不應用初始方向一路斜向生成火牢")
	}
}

func TestSkillGroundEffectFireWallTickDamagesNearbyPlayer(t *testing.T) {
	ws := world.NewState()
	owner := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Intel:     18,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         102,
		Y:         100,
		MapID:     4,
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
		OwnerCharID:  owner.CharID,
		OwnerSession: owner.SessionID,
		OwnerName:    owner.Name,
		OwnerIntel:   owner.Intel,
	})
	sys := NewGroundEffectSystem(ws, &handler.Deps{World: ws, Log: zap.NewNop()})

	for i := 0; i < fireWallDamageIntervalTicks; i++ {
		sys.Update(200 * time.Millisecond)
	}

	if target.HP >= 100 {
		t.Fatalf("火牢 tick 應傷害範圍內玩家，HP=%d", target.HP)
	}
}

func TestSkillGroundEffectLifeStreamAddsHPRegenBonus(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     4,
	})
	ws.AddGroundEffect(&world.GroundEffect{
		ID:      world.NextGroundEffectID(),
		SkillID: 63,
		NpcID:   81169,
		GfxID:   2231,
		Type:    world.GroundEffectLifeStream,
		X:       100,
		Y:       100,
		MapID:   4,
	})
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	regen := NewRegenSystem(ws, engine, nil, nil)

	if bonus := regen.lifeStreamHPBonus(player); bonus != 3 {
		t.Fatalf("生命之泉 4 格內 HPR bonus=%d，want=3", bonus)
	}
}
