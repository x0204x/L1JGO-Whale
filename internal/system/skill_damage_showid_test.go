package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func hasUseAttackSkillPacket(packets [][]byte, casterID, targetID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 10 || pkt[0] != packet.S_OPCODE_ATTACK || pkt[1] != 18 {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[2:6])) == casterID &&
			int32(binary.LittleEndian.Uint32(pkt[6:10])) == targetID {
			return true
		}
	}
	return false
}

func finalBurnShowIDSkill() *data.SkillInfo {
	return &data.SkillInfo{
		SkillID:     skillFinalBurn,
		Target:      "attack",
		Type:        64,
		Ranged:      10,
		ActionID:    18,
		CastGfx:     167,
		DamageValue: 1,
	}
}

func TestSkillDamagePlayerAttackRejectsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    70,
		MP:        80,
		MaxMP:     100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other-show-target",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    71,
		HP:        200,
		MaxHP:     200,
	})

	newSkillTestSystem(t, ws).executeAttackSkill(caster.Session, caster, finalBurnShowIDSkill(), target.CharID)

	if target.HP != 200 {
		t.Fatalf("不同 ShowID 玩家目標不應受到攻擊技能傷害，HP=%d", target.HP)
	}
}

func TestSkillDamageNpcAttackRejectsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    72,
		MP:        80,
		MaxMP:     100,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45000,
		Impl:   "L1Monster",
		Name:   "other-show-npc",
		X:      101,
		Y:      100,
		MapID:  4,
		ShowID: 73,
		HP:     100000,
		MaxHP:  100000,
	}
	ws.AddNpc(npc)

	newSkillTestSystem(t, ws).executeAttackSkill(caster.Session, caster, finalBurnShowIDSkill(), npc.ID)

	if npc.HP != 100000 {
		t.Fatalf("不同 ShowID NPC 目標不應受到攻擊技能傷害，HP=%d", npc.HP)
	}
}

func TestSkillDamagePlayerAttackBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    74,
		MP:        80,
		MaxMP:     100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    74,
		HP:        200,
		MaxHP:     200,
	})
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    74,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    75,
	})

	newSkillTestSystem(t, ws).executeAttackSkill(caster.Session, caster, finalBurnShowIDSkill(), target.CharID)

	if !hasUseAttackSkillPacket(drainSkillTestPackets(sameViewer.Session), caster.CharID, target.CharID) {
		t.Fatal("同 ShowID 玩家應收到玩家攻擊技能封包")
	}
	if hasUseAttackSkillPacket(drainSkillTestPackets(otherViewer.Session), caster.CharID, target.CharID) {
		t.Fatal("不同 ShowID 玩家不應收到玩家攻擊技能封包")
	}
}

func TestSkillDamageNpcAttackBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    76,
		MP:        80,
		MaxMP:     100,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45000,
		Impl:   "L1Monster",
		Name:   "npc",
		X:      101,
		Y:      100,
		MapID:  4,
		ShowID: 76,
		HP:     100000,
		MaxHP:  100000,
	}
	ws.AddNpc(npc)
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    76,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    77,
	})

	newSkillTestSystem(t, ws).executeAttackSkill(caster.Session, caster, finalBurnShowIDSkill(), npc.ID)

	if !hasUseAttackSkillPacket(drainSkillTestPackets(sameViewer.Session), caster.CharID, npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC 攻擊技能封包")
	}
	if hasUseAttackSkillPacket(drainSkillTestPackets(otherViewer.Session), caster.CharID, npc.ID) {
		t.Fatal("不同 ShowID 玩家不應收到 NPC 攻擊技能封包")
	}
}

func TestSkillTurnUndeadBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    78,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45000,
		Impl:   "L1Monster",
		Name:   "npc",
		X:      101,
		Y:      100,
		MapID:  4,
		ShowID: 78,
		HP:     100000,
		MaxHP:  100000,
	}
	ws.AddNpc(npc)
	sameViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same-viewer",
		X:         102,
		Y:         100,
		MapID:     4,
		ShowID:    78,
	})
	otherViewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other-viewer",
		X:         102,
		Y:         101,
		MapID:     4,
		ShowID:    79,
	})

	newSkillTestSystem(t, ws).executeTurnUndead(caster.Session, caster, &data.SkillInfo{
		SkillID:  18,
		ActionID: 19,
		CastGfx:  754,
	}, npc)

	if !hasActionGfxPacket(drainSkillTestPackets(sameViewer.Session), caster.CharID, 19) {
		t.Fatal("同 ShowID 玩家應收到起死回生術施法動作")
	}
	if hasActionGfxPacket(drainSkillTestPackets(otherViewer.Session), caster.CharID, 19) {
		t.Fatal("不同 ShowID 玩家不應收到起死回生術施法動作")
	}
}

func TestSkillDamageNpcAreaAttackSkipsDifferentShowNpcsLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    80,
		Level:     90,
		Intel:     40,
		SP:        20,
	})
	primary := &world.NpcInfo{
		ID:     2001,
		NpcID:  45000,
		Impl:   "L1Monster",
		Name:   "primary",
		X:      101,
		Y:      100,
		MapID:  4,
		ShowID: 80,
		HP:     100000,
		MaxHP:  100000,
	}
	otherShowNpc := &world.NpcInfo{
		ID:     2002,
		NpcID:  45001,
		Impl:   "L1Monster",
		Name:   "other-show-npc",
		X:      102,
		Y:      100,
		MapID:  4,
		ShowID: 81,
		HP:     100000,
		MaxHP:  100000,
	}
	ws.AddNpc(primary)
	ws.AddNpc(otherShowNpc)
	skill := &data.SkillInfo{
		SkillID:         4,
		Target:          "attack",
		Type:            64,
		Ranged:          10,
		Area:            2,
		ActionID:        18,
		CastGfx:         167,
		DamageValue:     100,
		DamageDice:      1,
		DamageDiceCount: 1,
	}

	newSkillTestSystem(t, ws).executeAttackSkill(caster.Session, caster, skill, primary.ID)

	if otherShowNpc.HP != 100000 {
		t.Fatalf("不同 ShowID 的範圍額外 NPC 目標不應受傷，HP=%d", otherShowNpc.HP)
	}
}
