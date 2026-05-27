package system

import (
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newPvPShowIDTestSystem(t *testing.T, ws *world.State) *PvPSystem {
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
	cfg := &config.Config{}
	cfg.Rates.LawfulRate = 1
	return &PvPSystem{deps: &handler.Deps{
		Config:    cfg,
		World:     ws,
		Scripting: engine,
		Items:     items,
		Log:       zap.NewNop(),
	}}
}

func TestPvPMeleeRejectsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    200,
		HP:        1000,
		MaxHP:     1000,
	})
	s := newPvPShowIDTestSystem(t, ws)

	s.HandlePvPAttack(attacker, target)

	if hasOpcodePacket(drainSkillTestPackets(attacker.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("不同 ShowID PvP 近戰不應送攻擊封包")
	}
	if target.HP != 1000 {
		t.Fatalf("不同 ShowID PvP 近戰不應造成傷害，HP=%d", target.HP)
	}
}

func TestPvPMeleeBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    100,
		HP:        1000,
		MaxHP:     1000,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    200,
	})
	s := newPvPShowIDTestSystem(t, ws)

	s.HandlePvPAttack(attacker, target)

	if !hasOpcodePacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("同 ShowID 觀眾應收到 PvP 近戰封包")
	}
	if hasOpcodePacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("不同 ShowID 觀眾不應收到 PvP 近戰封包")
	}
}

func TestPvPFarAttackRejectsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	attacker.Inv.AddItemWithID(5001, 40743, 10, "arrow", 0, 0, true, 1)
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    200,
		HP:        1000,
		MaxHP:     1000,
	})
	s := newPvPShowIDTestSystem(t, ws)

	s.HandlePvPFarAttack(attacker, target)

	if hasOpcodePacket(drainSkillTestPackets(attacker.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("不同 ShowID PvP 遠攻不應送攻擊封包")
	}
	if target.HP != 1000 {
		t.Fatalf("不同 ShowID PvP 遠攻不應造成傷害，HP=%d", target.HP)
	}
}

func TestPvPFarAttackBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	attacker.Inv.AddItemWithID(5001, 40743, 10, "arrow", 0, 0, true, 1)
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    100,
		HP:        1000,
		MaxHP:     1000,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    200,
	})
	s := newPvPShowIDTestSystem(t, ws)

	s.HandlePvPFarAttack(attacker, target)

	if !hasOpcodePacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("同 ShowID 觀眾應收到 PvP 遠攻封包")
	}
	if hasOpcodePacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_ATTACK) {
		t.Fatalf("不同 ShowID 觀眾不應收到 PvP 遠攻封包")
	}
}

func TestPvPPinkNameBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	attacker := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	victim := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 2, Session: newSkillTestSession(t, 2), CharID: 1002, ShowID: 100})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 3, Session: newSkillTestSession(t, 3), CharID: 1003, X: 101, Y: 100, MapID: 900, ShowID: 100})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 4, Session: newSkillTestSession(t, 4), CharID: 1004, X: 101, Y: 100, MapID: 900, ShowID: 200})
	s := newPvPShowIDTestSystem(t, ws)

	s.TriggerPinkName(attacker, victim)

	if !hasOpcodePacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_PINKNAME) {
		t.Fatalf("同 ShowID 觀眾應收到粉名封包")
	}
	if hasOpcodePacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_PINKNAME) {
		t.Fatalf("不同 ShowID 觀眾不應收到粉名封包")
	}
}

func TestPvPAddLawfulFromNpcBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	killer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "killer",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 2, Session: newSkillTestSession(t, 2), CharID: 1002, X: 101, Y: 100, MapID: 900, ShowID: 100})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 3, Session: newSkillTestSession(t, 3), CharID: 1003, X: 101, Y: 100, MapID: 900, ShowID: 200})
	s := newPvPShowIDTestSystem(t, ws)

	s.AddLawfulFromNpc(killer, -100)

	if !hasOpcodePacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_LAWFUL) {
		t.Fatalf("同 ShowID 觀眾應收到正義值封包")
	}
	if hasOpcodePacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_LAWFUL) {
		t.Fatalf("不同 ShowID 觀眾不應收到正義值封包")
	}
}

func TestPvPPKKillBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	killer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "killer",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
		PinkName:  true,
		Lawful:    1000,
	})
	victim := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 2, Session: newSkillTestSession(t, 2), CharID: 1002, X: 101, Y: 100, MapID: 900, ShowID: 100, Lawful: 0})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 3, Session: newSkillTestSession(t, 3), CharID: 1003, X: 101, Y: 100, MapID: 900, ShowID: 100})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 4, Session: newSkillTestSession(t, 4), CharID: 1004, X: 101, Y: 100, MapID: 900, ShowID: 200})
	s := newPvPShowIDTestSystem(t, ws)

	s.processPKKill(killer, victim)

	samePackets := drainSkillTestPackets(sameShow.Session)
	if !hasOpcodePacket(samePackets, packet.S_OPCODE_PINKNAME) || !hasOpcodePacket(samePackets, packet.S_OPCODE_LAWFUL) {
		t.Fatalf("同 ShowID 觀眾應收到 PK 粉名解除與正義值封包")
	}
	otherPackets := drainSkillTestPackets(otherShow.Session)
	if hasOpcodePacket(otherPackets, packet.S_OPCODE_PINKNAME) || hasOpcodePacket(otherPackets, packet.S_OPCODE_LAWFUL) {
		t.Fatalf("不同 ShowID 觀眾不應收到 PK 粉名解除或正義值封包")
	}
}

func TestPvPDropOneItemBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	victim := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "victim",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	victim.Inv.AddItemWithID(7001, 40317, 1, "whetstone", 0, 0, true, 1)
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 2, Session: newSkillTestSession(t, 2), CharID: 1002, X: 101, Y: 100, MapID: 900, ShowID: 100})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{SessionID: 3, Session: newSkillTestSession(t, 3), CharID: 1003, X: 101, Y: 100, MapID: 900, ShowID: 200})
	s := newPvPShowIDTestSystem(t, ws)

	s.dropOneItem(victim)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), 7001) {
		t.Fatalf("同 ShowID 觀眾應收到 PK 掉落物件")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), 7001) {
		t.Fatalf("不同 ShowID 觀眾不應收到 PK 掉落物件")
	}
}
