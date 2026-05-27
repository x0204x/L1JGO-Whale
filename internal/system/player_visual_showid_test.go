package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func hasVisualShowIDPacket(packets [][]byte, opcode byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != opcode {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}

func addVisualShowIDPlayer(ws *world.State, p *world.PlayerInfo) *world.PlayerInfo {
	if p.Known == nil {
		p.Known = world.NewKnownEntities()
	}
	return addSkillTestPlayer(ws, p)
}

func TestEquipVisualBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:        8101,
		Name:          "player",
		X:             100,
		Y:             100,
		MapID:         900,
		ShowID:        77,
		SessionID:     1,
		Session:       newSkillTestSession(t, 1),
		CurrentWeapon: 4,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8102,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8103,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	NewEquipSystem(&handler.Deps{World: ws}).broadcastVisualUpdate(player.Session, player)

	if !hasVisualShowIDPacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_CHANGE_DESC, player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到裝備視覺更新")
	}
	if hasVisualShowIDPacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_CHANGE_DESC, player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到裝備視覺更新")
	}
}

func TestInvisCloakBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8201,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8202,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8203,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	equip := NewEquipSystem(&handler.Deps{World: ws})
	equip.applyInvisCloak(player.Session, player, true)

	if !hasRemoveObjectPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到隱身斗篷 remove")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到隱身斗篷 remove")
	}

	equip.applyInvisCloak(player.Session, player, false)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到解除隱身 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到解除隱身 put object")
	}
}

func TestPolymorphVisualBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8301,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		ClassID:   61,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8302,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8303,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	poly := NewPolymorphSystem(&handler.Deps{World: ws, Log: zap.NewNop()})
	poly.applyPoly(player, &data.PolymorphInfo{PolyID: 945, Name: "orc", WeaponEquip: 2047, ArmorEquip: 8191}, 0)

	if !hasVisualShowIDPacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_POLY, player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到變身封包")
	}
	if hasVisualShowIDPacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_POLY, player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到變身封包")
	}

	poly.UndoPoly(player)

	if !hasVisualShowIDPacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_POLY, player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到解除變身封包")
	}
	if hasVisualShowIDPacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_POLY, player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到解除變身封包")
	}
}

func TestSkillCancelInvisibilityBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8401,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Invisible: true,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8402,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8403,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})

	newSkillTestSystem(t, ws).CancelInvisibility(player)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到隱身解除 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到隱身解除 put object")
	}
}

func TestSkillSpeedBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8501,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8502,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8503,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	skill := newSkillTestSystem(t, ws)
	skill.sendSpeedToAll(player, 1, 120)

	if !hasVisualShowIDPacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_SPEED, player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到速度封包")
	}
	if hasVisualShowIDPacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_SPEED, player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到速度封包")
	}

	skill.sendBraveToAll(player, 1, 120)

	if !hasVisualShowIDPacket(drainSkillTestPackets(sameShow.Session), packet.S_OPCODE_SKILLBRAVE, player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到勇敢封包")
	}
	if hasVisualShowIDPacket(drainSkillTestPackets(otherShow.Session), packet.S_OPCODE_SKILLBRAVE, player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到勇敢封包")
	}
}

func TestSkillCancelAllBuffsInvisibilityBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8601,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Invisible: true,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8602,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8603,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})

	newSkillTestSystem(t, ws).CancelAllBuffs(player)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到魔法相消後的 put object")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到魔法相消後的 put object")
	}
}

func TestSkillCounterMagicEffectBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8701,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8702,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8703,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 31, TicksLeft: 100})

	skill := newSkillTestSystem(t, ws)
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入技能表失敗: %v", err)
	}
	skill.deps.Skills = skills
	expectedGfx := int32(10702)
	if sk := skills.Get(31); sk != nil && sk.CastGfx2 > 0 {
		expectedGfx = sk.CastGfx2
	}
	if !skill.tryCounterMagic(target, 22) {
		t.Fatalf("目標有魔法屏障時應攔截非豁免技能")
	}

	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), target.CharID, expectedGfx) {
		t.Fatalf("同 ShowID 玩家應收到魔法屏障特效")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), target.CharID, expectedGfx) {
		t.Fatalf("不同 ShowID 玩家不應收到魔法屏障特效")
	}
}

func TestSkillCastActionBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8801,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		Level:     50,
		MaxHP:     100,
		MaxMP:     100,
		HP:        100,
		MP:        100,
	})
	sameShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8802,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addVisualShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    8803,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})

	newSkillTestSystem(t, ws).executeBuffSkill(player.Session, player, &data.SkillInfo{
		SkillID:      21,
		Target:       "buff",
		ActionID:     19,
		BuffDuration: 120,
	}, player.CharID)

	if !hasActionGfxPacket(drainSkillTestPackets(sameShow.Session), player.CharID, 19) {
		t.Fatalf("同 ShowID 玩家應收到施法動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherShow.Session), player.CharID, 19) {
		t.Fatalf("不同 ShowID 玩家不應收到施法動作")
	}
}
