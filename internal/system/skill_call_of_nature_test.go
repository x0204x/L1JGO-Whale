package system

import (
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillCallOfNatureTeleportToMotherReturnsCasterToMotherTree(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
		X:         100,
		Y:         100,
		MapID:     4,
		Heading:   2,
	})
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 131, ActionID: 19, CastGfx: 169}, 0)

	if caster.X != 33047 || caster.Y != 32338 || caster.MapID != 4 || caster.Heading != 5 {
		t.Fatalf("世界樹的呼喚應回母樹座標，got=(%d,%d,%d,%d)", caster.X, caster.Y, caster.MapID, caster.Heading)
	}
}

func TestSkillCallOfNatureCallOfNatureRequestsDeadPlayerConsent(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
		MaxMP:     80,
	})
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, target.CharID)

	if !target.Dead {
		t.Fatalf("自然呼喚對玩家應等待同意，不應立即復活，Dead=%v", target.Dead)
	}
	if target.PendingResSkill != 165 || target.PendingResCaster != caster.CharID {
		t.Fatalf("自然呼喚應設定待同意復活資訊，Pending=(%d,%d)", target.PendingResSkill, target.PendingResCaster)
	}
}

func TestSkillCallOfNatureCallOfNatureRejectsCorpseTileOccupiedByAlivePlayer(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
		MaxMP:     80,
	})
	addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "blocker",
		X:         target.X,
		Y:         target.Y,
		MapID:     target.MapID,
	})
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, target.CharID)

	if target.PendingResSkill != 0 || target.PendingResCaster != 0 || !target.Dead {
		t.Fatalf("自然呼喚遇到屍體格有活人時應拒絕，Dead=%v Pending=(%d,%d)", target.Dead, target.PendingResSkill, target.PendingResCaster)
	}
}

func TestSkillCallOfNatureEarthBindPlayerDurationMatchesJavaRandomRange(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 157)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
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
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      157,
		BuffDuration: 16,
		Target:       "buff",
		ActionID:     19,
		CastGfx:      2251,
	}, target.CharID)

	buff := target.GetBuff(157)
	if buff == nil || !target.Paralyzed {
		t.Fatalf("大地屏障應凍結玩家目標，Paralyzed=%v buff=%v", target.Paralyzed, buff)
	}
	if buff.TicksLeft < 5 || buff.TicksLeft > 60 {
		t.Fatalf("大地屏障玩家持續時間應為 Java 1-12 秒，TicksLeft=%d", buff.TicksLeft)
	}
}

func TestSkillCallOfNatureElfRegenBuffsBypassOverweightHPMPBlock(t *testing.T) {
	engine := newSkillTestSystem(t, world.NewState()).deps.Scripting
	base := &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "elf",
		Level:      50,
		Str:        10,
		Con:        10,
		Wis:        18,
		Food:       100,
		HP:         50,
		MaxHP:      100,
		MP:         20,
		MaxMP:      100,
		MPR:        1,
		RegenHPAcc: 2,
		Inv:        world.NewInventory(),
	}
	base.Inv.AddItem(40308, 5000, "adena", 0, 1000, true, 1)

	for _, skillID := range []int32{169, 176} {
		ws := world.NewState()
		player := *base
		player.Inv = world.NewInventory()
		player.Inv.AddItem(40308, 5000, "adena", 0, 1000, true, 1)
		player.ActiveBuffs = map[int32]*world.ActiveBuff{skillID: {SkillID: skillID, TicksLeft: 100}}
		ws.AddPlayer(&player)
		regen := NewRegenSystem(ws, engine, nil, &config.Config{})

		regen.tickHPRegen(&player)
		regen.tickMPRegen(&player)

		if player.HP <= base.HP || player.MP <= base.MP {
			t.Fatalf("技能 %d 應在負重時允許 HP/MP 回復，HP=%d MP=%d", skillID, player.HP, player.MP)
		}
	}
}
