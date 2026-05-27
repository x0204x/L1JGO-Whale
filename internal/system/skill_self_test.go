package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillStormWalkUsesJavaBraveConflictsWithoutRemovingStrengthBuff(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "storm-walk",
		X:          100,
		Y:          100,
		MapID:      4,
		Str:        23,
		BraveSpeed: 1,
	})
	for _, skillID := range []int32{
		52,  // HOLY_WALK
		101, // MOVING_ACCELERATION
		150, // WIND_WALK
		155, // FIRE_BLESS
		186, // BLOODLUST
		handler.SkillStatusBrave,
		handler.SkillStatusElfBrave,
	} {
		player.AddBuff(&world.ActiveBuff{
			SkillID:       skillID,
			TicksLeft:     100,
			SetBraveSpeed: 1,
		})
	}
	player.AddBuff(&world.ActiveBuff{
		SkillID:   42, // PHYSICAL_ENCHANT_STR, not HOLY_WALK.
		TicksLeft: 100,
		DeltaStr:  5,
	})

	s := newSkillTestSystem(t, ws)
	s.executeSelfSkill(player.Session, player, &data.SkillInfo{
		SkillID:  172,
		ActionID: 19,
	})

	for _, skillID := range []int32{
		52,
		101,
		150,
		155,
		186,
		handler.SkillStatusBrave,
		handler.SkillStatusElfBrave,
	} {
		if player.HasBuff(skillID) {
			t.Fatalf("STORM_WALK 應清除 Java 速度互斥 buff %d", skillID)
		}
	}
	if !player.HasBuff(42) {
		t.Fatal("STORM_WALK 不應移除 42 PHYSICAL_ENCHANT_STR")
	}
	if player.Str != 23 {
		t.Fatalf("STORM_WALK 不應回退 STR buff 數值，Str=%d", player.Str)
	}
	if !player.HasBuff(172) || player.BraveSpeed != 4 || player.BraveTicks != 300*5 {
		t.Fatalf("STORM_WALK 應套用 brave speed 4 與 300 秒 buff，buff172=%v BraveSpeed=%d BraveTicks=%d",
			player.GetBuff(172), player.BraveSpeed, player.BraveTicks)
	}
}

func TestSkillSelfBuffSendsYiweiPostCastStatusRefresh(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "body-to-mind",
		X:         100,
		Y:         100,
		MapID:     4,
		MP:        10,
		MaxMP:     100,
		SP:        7,
		MR:        22,
	})

	s := newSkillTestSystem(t, ws)
	s.executeSelfSkill(player.Session, player, &data.SkillInfo{
		SkillID:  130,
		ActionID: 19,
		CastGfx:  2179,
	})

	packets := drainSkillTestPackets(player.Session)
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 補助魔法後會送 S_SPMR，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 補助魔法後會送 S_OwnCharStatus，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(player)) {
		t.Fatalf("yiwei sendGrfx 補助魔法後會送 S_PacketBox.UPDATE_ER，packets=%v", packets)
	}
}

func TestSkillSelfInvisibilityBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    7,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-show",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    7,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-show",
		X:         101,
		Y:         101,
		MapID:     4,
		ShowID:    8,
	})

	newSkillTestSystem(t, ws).executeSelfSkill(player.Session, player, &data.SkillInfo{
		SkillID:  60,
		ActionID: 19,
	})

	samePackets := drainSkillTestPackets(sameShow.Session)
	otherPackets := drainSkillTestPackets(otherShow.Session)
	if !hasRemoveObjectPacket(samePackets, player.CharID) {
		t.Fatal("同 ShowID 玩家應收到隱身 remove object")
	}
	if hasRemoveObjectPacket(otherPackets, player.CharID) {
		t.Fatal("不同 ShowID 玩家不應收到隱身 remove object")
	}
	if !hasActionGfxPacket(samePackets, player.CharID, 19) {
		t.Fatal("同 ShowID 玩家應收到隱身施法動作")
	}
	if hasActionGfxPacket(otherPackets, player.CharID, 19) {
		t.Fatal("不同 ShowID 玩家不應收到隱身施法動作")
	}
}

func TestSkillDetectionAffectsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    11,
	})
	sameHidden := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-hidden",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    11,
		Invisible: true,
	})
	otherHidden := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-hidden",
		X:         101,
		Y:         101,
		MapID:     4,
		ShowID:    12,
		Invisible: true,
	})
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    11,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 5,
		Session:   newSkillTestSession(t, 5),
		CharID:    1005,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    12,
	})
	sameHidden.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})
	otherHidden.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})

	newSkillTestSystem(t, ws).executeSelfSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  13,
		ActionID: 19,
	})

	if sameHidden.HasBuff(60) {
		t.Fatal("同 ShowID 隱身玩家應被無所遁形揭示")
	}
	if !otherHidden.HasBuff(60) {
		t.Fatal("不同 ShowID 隱身玩家不應被無所遁形揭示")
	}
	if !hasPutObjectPacket(drainSkillTestPackets(sameViewer.Session), sameHidden.CharID) {
		t.Fatal("同 ShowID 觀眾應收到被揭示玩家 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherViewer.Session), sameHidden.CharID) {
		t.Fatal("不同 ShowID 觀眾不應收到被揭示玩家 put object")
	}
}

func hasYiweiUpdateERPacket(packets [][]byte, wantER int16) bool {
	for _, pkt := range packets {
		if len(pkt) >= 4 && pkt[0] == packet.S_OPCODE_EVENT && pkt[1] == 132 {
			got := int16(uint16(pkt[2]) | uint16(pkt[3])<<8)
			return got == wantER
		}
	}
	return false
}
