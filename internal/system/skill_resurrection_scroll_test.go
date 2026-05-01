package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newResurrectionScrollItemUseSystem(t *testing.T, ws *world.State) *ItemUseSystem {
	t.Helper()
	return NewItemUseSystem(&handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
	})
}

func TestSkillResurrectionScrollResurrectionScrollSetsHalfHpConsentForDeadPlayer(t *testing.T) {
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
		Name:      "dead",
		X:         101,
		Y:         100,
		MapID:     4,
		Dead:      true,
		HP:        0,
		MaxHP:     120,
	})
	scroll := caster.Inv.AddItemWithID(7001, 40089, 2, "復活卷軸", 471, 630, true, 1)
	s := newResurrectionScrollItemUseSystem(t, ws)

	if !s.UseResurrectionScroll(caster.Session, caster, scroll, target.CharID) {
		t.Fatal("復活卷軸對死亡玩家應視為已使用")
	}

	if !target.Dead || target.PendingResSkill != 61 || target.PendingResCaster != caster.CharID {
		t.Fatalf("普通復活卷軸應送半血復活同意，不立即復活，Dead=%v Pending=(%d,%d)",
			target.Dead, target.PendingResSkill, target.PendingResCaster)
	}
	if scroll.Count != 1 {
		t.Fatalf("有效復活卷軸應消耗 1 張，Count=%d", scroll.Count)
	}
}

func TestSkillResurrectionScrollBlessedResurrectionScrollSetsFullHpConsentForDeadPlayer(t *testing.T) {
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
		Name:      "dead",
		X:         101,
		Y:         100,
		MapID:     4,
		Dead:      true,
		HP:        0,
		MaxHP:     120,
	})
	scroll := caster.Inv.AddItemWithID(7001, 140089, 1, "祝福復活卷軸", 471, 630, true, 0)
	s := newResurrectionScrollItemUseSystem(t, ws)

	if !s.UseResurrectionScroll(caster.Session, caster, scroll, target.CharID) {
		t.Fatal("祝福復活卷軸對死亡玩家應視為已使用")
	}

	if !target.Dead || target.PendingResSkill != 75 || target.PendingResCaster != caster.CharID {
		t.Fatalf("祝福復活卷軸應送滿血復活同意，不立即復活，Dead=%v Pending=(%d,%d)",
			target.Dead, target.PendingResSkill, target.PendingResCaster)
	}
	if caster.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("最後一張祝福復活卷軸應從背包移除")
	}
}

func TestSkillResurrectionScrollResurrectionScrollResurrectsDeadNpcQuarterHP(t *testing.T) {
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
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		X:     101,
		Y:     100,
		MapID: 4,
		Dead:  true,
		HP:    0,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	ws.NpcDied(npc)
	scroll := caster.Inv.AddItemWithID(7001, 40089, 1, "復活卷軸", 471, 630, true, 1)
	s := newResurrectionScrollItemUseSystem(t, ws)

	if !s.UseResurrectionScroll(caster.Session, caster, scroll, npc.ID) {
		t.Fatal("復活卷軸對死亡 NPC 應視為已使用")
	}

	if npc.Dead || npc.HP != 25 {
		t.Fatalf("復活卷軸應以 1/4 HP 復活 NPC，Dead=%v HP=%d", npc.Dead, npc.HP)
	}
	if caster.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("復活 NPC 後卷軸應從背包移除")
	}
}

func TestSkillResurrectionScrollResurrectionScrollDoesNotConsumeForSelfTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Dead:      true,
		HP:        0,
		MaxHP:     100,
	})
	scroll := caster.Inv.AddItemWithID(7001, 40089, 2, "復活卷軸", 471, 630, true, 1)
	s := newResurrectionScrollItemUseSystem(t, ws)

	if s.UseResurrectionScroll(caster.Session, caster, scroll, caster.CharID) {
		t.Fatal("Java 復活卷軸對自己使用應直接返回，不應消耗")
	}
	if scroll.Count != 2 {
		t.Fatalf("對自己使用不應消耗復活卷軸，Count=%d", scroll.Count)
	}
}
