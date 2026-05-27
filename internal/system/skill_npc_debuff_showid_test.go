package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillNpcDebuffRejectsDifferentShowNpcTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    7501,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    70,
	})
	npc := &world.NpcInfo{
		ID:       7601,
		NpcID:    45000,
		Impl:     "L1Monster",
		Name:     "other-show-npc",
		X:        101,
		Y:        100,
		MapID:    900,
		ShowID:   99,
		WeakAttr: weakElementalFire,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	s.executeNpcDebuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  23,
		ActionID: 19,
		Ranged:   10,
	}, npc)

	if npc.HateList != nil && npc.HateList[caster.SessionID] != 0 {
		t.Fatalf("yiwei isTarget 會以 showId 不同擋下 NPC debuff，不應加仇恨，hate=%v", npc.HateList)
	}
	if hasActionGfxPacket(drainSkillTestPackets(caster.Session), caster.CharID, 19) {
		t.Fatal("不同 ShowID NPC 目標應在施法動畫前被拒絕")
	}
}

func TestSkillNpcDebuffBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    7511,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    70,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    7512,
		Name:      "same",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    70,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    7513,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    99,
	})
	npc := &world.NpcInfo{
		ID:       7611,
		NpcID:    45000,
		Impl:     "L1Monster",
		Name:     "same-show-npc",
		X:        101,
		Y:        100,
		MapID:    900,
		ShowID:   70,
		WeakAttr: weakElementalFire,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	s.executeNpcDebuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  23,
		ActionID: 19,
		Ranged:   10,
	}, npc)

	samePackets := drainSkillTestPackets(sameShow.Session)
	if !hasActionGfxPacket(samePackets, caster.CharID, 19) {
		t.Fatal("同 ShowID 觀察者應收到 NPC debuff 施法動作")
	}
	if !hasSkillEffectPacket(samePackets, npc.ID, 2167) {
		t.Fatal("同 ShowID 觀察者應收到 NPC 弱火特效")
	}

	otherPackets := drainSkillTestPackets(otherShow.Session)
	if hasActionGfxPacket(otherPackets, caster.CharID, 19) {
		t.Fatal("不同 ShowID 觀察者不應收到 NPC debuff 施法動作")
	}
	if hasSkillEffectPacket(otherPackets, npc.ID, 2167) {
		t.Fatal("不同 ShowID 觀察者不應收到 NPC 弱點特效")
	}
}
