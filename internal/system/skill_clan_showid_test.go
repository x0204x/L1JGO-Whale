package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillClanCallClanCastActionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    10,
		ClanID:    7,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    10,
		ClanID:    7,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same_show",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other_show",
		X:         103,
		Y:         100,
		MapID:     4,
		ShowID:    99,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 116, Target: "buff", ActionID: 19}, member.CharID)

	if member.PendingYesNoType != 729 || member.PendingYesNoData != caster.CharID {
		t.Fatalf("呼喚盟友仍應送 729 確認給同血盟目標，Pending=(%d,%d)", member.PendingYesNoType, member.PendingYesNoData)
	}
	if !hasActionGfxPacket(drainSkillTestPackets(caster.Session), caster.CharID, 19) {
		t.Fatal("Java sendPacketsAll 會讓施法者自己看到 CALL_CLAN 施法動作")
	}
	if !hasActionGfxPacket(drainSkillTestPackets(sameShow.Session), caster.CharID, 19) {
		t.Fatal("同 ShowID 觀眾應看到 CALL_CLAN 施法動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherShow.Session), caster.CharID, 19) {
		t.Fatal("不同 ShowID 觀眾不應看到 CALL_CLAN 施法動作")
	}
}

func TestSkillClanRunClanCastActionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    10,
		ClanID:    7,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    10,
		ClanID:    7,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same_show",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other_show",
		X:         103,
		Y:         100,
		MapID:     4,
		ShowID:    99,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 118, Target: "buff", ActionID: 19}, member.CharID)

	if caster.X != member.X || caster.Y != member.Y || caster.MapID != member.MapID {
		t.Fatalf("援護盟友仍應傳送到 Java 允許的同血盟目標位置，got=(%d,%d,%d)", caster.X, caster.Y, caster.MapID)
	}
	if !hasActionGfxPacket(drainSkillTestPackets(caster.Session), caster.CharID, 19) {
		t.Fatal("Java sendPacketsAll 會讓施法者自己看到 RUN_CLAN 施法動作")
	}
	if !hasActionGfxPacket(drainSkillTestPackets(sameShow.Session), caster.CharID, 19) {
		t.Fatal("同 ShowID 觀眾應看到 RUN_CLAN 施法動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherShow.Session), caster.CharID, 19) {
		t.Fatal("不同 ShowID 觀眾不應看到 RUN_CLAN 施法動作")
	}
}
