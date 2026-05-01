package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newSkillTestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 8, 8, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func newSkillTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	return &SkillSystem{deps: &handler.Deps{
		World:     ws,
		Scripting: engine,
		Log:       zap.NewNop(),
	}}
}

func addSkillTestPlayer(ws *world.State, p *world.PlayerInfo) *world.PlayerInfo {
	if p.Level == 0 {
		p.Level = 50
	}
	if p.HP == 0 {
		p.HP = 100
	}
	if p.MaxHP == 0 {
		p.MaxHP = 100
	}
	if p.Intel == 0 {
		p.Intel = 18
	}
	if p.Inv == nil {
		p.Inv = world.NewInventory()
	}
	ws.AddPlayer(p)
	return p
}

func TestDeathTombttackSkillDamagesPlayerTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
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
		X:         103,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         4,
		SkillLevel:      1,
		Target:          "attack",
		Type:            64,
		DamageValue:     20,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          10,
		ActionID:        18,
		CastGfx:         167,
	}

	s.executeAttackSkill(caster.Session, caster, skill, target.CharID)

	if target.HP >= 100 {
		t.Fatalf("攻擊技能應傷害玩家目標，HP=%d", target.HP)
	}
	if !target.Dirty {
		t.Fatal("玩家目標受傷後應標記 Dirty")
	}
}

func TestSkillDamageHealSelfAreaAttackDamagesNearbyPlayersAndNpcs(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	nearPlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "near",
		X:         102,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	farPlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "far",
		X:         120,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		Name:  "npc",
		X:     101,
		Y:     100,
		MapID: 4,
		HP:    100,
		MaxHP: 100,
		MR:    0,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         53,
		SkillLevel:      7,
		Target:          "none",
		Type:            64,
		DamageValue:     20,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          0,
		Area:            4,
		ActionID:        18,
		CastGfx:         758,
	}

	s.executeSelfSkill(caster.Session, caster, skill)

	if nearPlayer.HP >= 100 {
		t.Fatalf("範圍攻擊應傷害附近玩家，HP=%d", nearPlayer.HP)
	}
	if farPlayer.HP != 100 {
		t.Fatalf("範圍外玩家不應受傷，HP=%d", farPlayer.HP)
	}
	if caster.HP != 100 {
		t.Fatalf("攻擊型範圍技能不應傷害施法者自己，HP=%d", caster.HP)
	}
	if npc.HP >= 100 {
		t.Fatalf("範圍攻擊仍應傷害附近 NPC，HP=%d", npc.HP)
	}
}

func TestSkillDamageHealSingleHealCapsAtMaxHP(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        95,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         57,
		SkillLevel:      8,
		Target:          "buff",
		Type:            16,
		DamageValue:     100,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          -1,
		ActionID:        19,
		CastGfx:         832,
	}

	s.executeBuffSkill(caster.Session, caster, skill, caster.CharID)

	if caster.HP != caster.MaxHP {
		t.Fatalf("治癒不可超過 MaxHP，HP=%d MaxHP=%d", caster.HP, caster.MaxHP)
	}
}

func TestSkillDamageHealHealAllHealsNearbyPlayers(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	nearPlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "near",
		X:         103,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	farPlayer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "far",
		X:         130,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         49,
		SkillLevel:      7,
		Target:          "none",
		Type:            16,
		DamageValue:     20,
		DamageDice:      1,
		DamageDiceCount: 1,
		Area:            -1,
		ActionID:        19,
		CastGfx:         759,
	}

	s.executeSelfSkill(caster.Session, caster, skill)

	if caster.HP <= 50 {
		t.Fatalf("全部治癒術應治癒施法者，HP=%d", caster.HP)
	}
	if nearPlayer.HP <= 50 {
		t.Fatalf("全部治癒術應治癒附近玩家，HP=%d", nearPlayer.HP)
	}
	if farPlayer.HP != 50 {
		t.Fatalf("全部治癒術不應治癒範圍外玩家，HP=%d", farPlayer.HP)
	}
}
