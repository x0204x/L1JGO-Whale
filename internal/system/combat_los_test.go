package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type fakePvPManager struct {
	meleeCalls int
	farCalls   int
}

func (m *fakePvPManager) HandlePvPAttack(_, _ *world.PlayerInfo) {
	m.meleeCalls++
}

func (m *fakePvPManager) HandlePvPFarAttack(_, _ *world.PlayerInfo) {
	m.farCalls++
}

func (m *fakePvPManager) TriggerPinkName(_, _ *world.PlayerInfo) {}

func (m *fakePvPManager) AddLawfulFromNpc(_ *world.PlayerInfo, _ int32) {}

func newCombatLOSTestSystem(t *testing.T, ws *world.State, pvp *fakePvPManager) *CombatSystem {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	items, err := data.LoadItemTable(
		"../../data/yaml/weapon_list.yaml",
		"../../data/yaml/armor_list.yaml",
		"../../data/yaml/etcitem_list.yaml",
	)
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	return &CombatSystem{deps: &handler.Deps{
		World:     ws,
		Scripting: engine,
		MapData:   newSkillLOSTestMap(t),
		Items:     items,
		Log:       zap.NewNop(),
		PvP:       pvp,
	}}
}

func TestMeleeAttackSkipsNpcBehindWall(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         101,
		Y:         100,
		MapID:     900,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		Impl:  "L1Monster",
		Name:  "behind_wall",
		X:     103,
		Y:     100,
		MapID: 900,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newCombatLOSTestSystem(t, ws, &fakePvPManager{})

	s.processMeleeAttack(player.SessionID, npc.ID)
	packets := drainSkillTestPackets(player.Session)

	if len(packets) != 0 {
		t.Fatalf("隔牆近戰不應該送出攻擊封包，packets=%d", len(packets))
	}
	if npc.HP != 100 {
		t.Fatalf("隔牆近戰不應該傷害 NPC，HP=%d", npc.HP)
	}
}

func TestRangedAttackSkipsNpcBehindWall(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         101,
		Y:         100,
		MapID:     900,
	})
	player.Equip.Set(world.SlotWeapon, &world.InvItem{ItemID: 190})
	npc := &world.NpcInfo{
		ID:    2001,
		Impl:  "L1Monster",
		Name:  "behind_wall",
		X:     103,
		Y:     100,
		MapID: 900,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newCombatLOSTestSystem(t, ws, &fakePvPManager{})

	s.processRangedAttack(player.SessionID, npc.ID)
	packets := drainSkillTestPackets(player.Session)

	if len(packets) != 0 {
		t.Fatalf("隔牆遠攻不應該送出攻擊封包，packets=%d", len(packets))
	}
	if npc.HP != 100 {
		t.Fatalf("隔牆遠攻不應該傷害 NPC，HP=%d", npc.HP)
	}
}

func TestMeleeAttackSkipsPlayerBehindWall(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         101,
		Y:         100,
		MapID:     900,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     900,
	})
	pvp := &fakePvPManager{}
	s := newCombatLOSTestSystem(t, ws, pvp)

	s.processMeleeAttack(attacker.SessionID, target.CharID)

	if pvp.meleeCalls != 0 {
		t.Fatalf("隔牆近戰 PvP 不應該委派傷害，calls=%d", pvp.meleeCalls)
	}
}

func TestRangedAttackSkipsPlayerBehindWall(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         101,
		Y:         100,
		MapID:     900,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     900,
	})
	pvp := &fakePvPManager{}
	s := newCombatLOSTestSystem(t, ws, pvp)

	s.processRangedAttack(attacker.SessionID, target.CharID)

	if pvp.farCalls != 0 {
		t.Fatalf("隔牆遠攻 PvP 不應該委派傷害，calls=%d", pvp.farCalls)
	}
}
