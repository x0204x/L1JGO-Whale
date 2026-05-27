package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func disablePlayerDebuffMRForStatusTest(t *testing.T, skillIDs ...int32) {
	t.Helper()
	previous := make(map[int32]bool, len(skillIDs))
	for _, skillID := range skillIDs {
		previous[skillID] = playerDebuffSkills[skillID]
		playerDebuffSkills[skillID] = false
	}
	t.Cleanup(func() {
		for _, skillID := range skillIDs {
			playerDebuffSkills[skillID] = previous[skillID]
		}
	})
}

func TestSkillStatusCancelCurseBlindRegistersBuffAndRemoveCurseClearsIt(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 20)
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
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      20,
		Target:       "attack",
		BuffDuration: 8,
		ActionID:     18,
	}, target.CharID)

	if !target.HasBuff(40) {
		t.Fatalf("闇盲咒術應依 Java 掛 40 致盲 buff，buff40=%v", target.GetBuff(40))
	}

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  37,
		Target:   "buff",
		ActionID: 18,
	}, target.CharID)

	if target.HasBuff(40) {
		t.Fatalf("解除詛咒應清除致盲 buff，buff40=%v", target.GetBuff(40))
	}
}

func TestSkillStatusCurseBlindUsesFloatingEyePacketType(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 20)
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
		X:         101,
		Y:         100,
		MapID:     4,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 1012, TicksLeft: 100})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      20,
		Target:       "attack",
		BuffDuration: 8,
		ActionID:     18,
	}, target.CharID)

	blindType, ok := findCurseBlindPacketType(drainSkillTestPackets(target.Session))
	if !ok {
		t.Fatalf("闇盲咒術應送出 S_CurseBlind 封包")
	}
	if blindType != 2 {
		t.Fatalf("有 STATUS_FLOATING_EYE(1012) 時 Java 送 S_CurseBlind(2)，got=%d", blindType)
	}
}

func findCurseBlindPacketType(packets [][]byte) (uint16, bool) {
	for _, pkt := range packets {
		if len(pkt) >= 3 && pkt[0] == packet.S_OPCODE_CURSEBLIND {
			return binary.LittleEndian.Uint16(pkt[1:3]), true
		}
	}
	return 0, false
}

func TestSkillStatusCancelCancellationWorksOnSelfAndKeepsNonCancellableBuff(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 43, BuffDuration: 120})
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 21, BuffDuration: 120})

	s.executeBuffSkill(player.Session, player, &data.SkillInfo{
		SkillID:  44,
		Target:   "buff",
		ActionID: 18,
	}, player.CharID)

	if player.HasBuff(43) || player.MoveSpeed != 0 {
		t.Fatalf("魔法相消對自己應移除加速術，buff43=%v MoveSpeed=%d", player.GetBuff(43), player.MoveSpeed)
	}
	if !player.HasBuff(21) {
		t.Fatalf("魔法相消應保留不可取消的鎧甲護持 buff")
	}
}

func TestSkillStatusRemovedWarriorSkillsAreNotCounterMagicExemptFor380C(t *testing.T) {
	for _, skillID := range []int32{226, 228, 230} {
		if counterMagicExempt[skillID] {
			t.Fatalf("3.80C 已剔除的戰士技能 %d 不應留在反魔法豁免表", skillID)
		}
	}
}

func TestSkillStatusCancelNpcKeepsShockStunDebuffLikeJava(t *testing.T) {
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
		Name:  "stunned-npc",
		X:     101,
		Y:     100,
		MapID: 4,
		HP:    100,
		MaxHP: 100,
		ActiveDebuffs: map[int32]int{
			87: 25,
			56: 25,
		},
		Paralyzed: true,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  44,
		Target:   "attack",
		ActionID: 18,
	}, npc.ID)

	if got := npc.ActiveDebuffs[87]; got != 25 {
		t.Fatalf("Java CANCELLATION 會略過 L1SkillMode.isNotCancelable 的 SHOCK_STUN，got ticks=%d", got)
	}
	if npc.HasDebuff(56) {
		t.Fatal("Java CANCELLATION 仍應移除可相消的 NPC debuff")
	}
	if npc.Paralyzed {
		t.Fatal("Java CANCELLATION 對 NPC 會執行 setParalyzed(false)，即使 87 效果本身保留")
	}
}

// Java `L1SkillMode.isNotCancelable()` 第 33 行明確列出 `SHOCK_STUN`，
// `CANCELLATION.java` 對玩家或 NPC 目標 buff 迴圈內若 `isNotCancelable(skillNum) && !cha.isDead()`
// 會略過該效果。NPC 目標已有 TestSkillStatusCancelNpcKeepsShockStunDebuffLikeJava；
// 本測試補上玩家目標的對等回歸，鎖定 Go `cancelAllBuffs` 透過 `IsNonCancellable(87)`
// 同樣略過 87 buff，但仍會解除其他可相消 buff。
func TestSkillStatusCancelPlayerKeepsShockStunBuffLikeJava(t *testing.T) {
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
		Name:      "stunned-player",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	target.Paralyzed = true
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 29, BuffDuration: 120})

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:  44,
		Target:   "buff",
		ActionID: 18,
	}, target.CharID)

	if !target.HasBuff(87) {
		t.Fatal("Java CANCELLATION 對玩家會略過 L1SkillMode.isNotCancelable 的 SHOCK_STUN，buff 87 不應被解除")
	}
	if target.HasBuff(29) {
		t.Fatal("Java CANCELLATION 仍應移除可相消的玩家 buff（緩速 29）")
	}
}

func TestSkillStatusCancelHasteOnSlowedTargetOnlyCancelsSlow(t *testing.T) {
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
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 29, BuffDuration: 120})

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      43,
		Target:       "buff",
		BuffDuration: 120,
		ActionID:     18,
	}, target.CharID)

	if target.HasBuff(29) || target.HasBuff(43) || target.MoveSpeed != 0 {
		t.Fatalf("加速術對緩速目標應只解除緩速，buff29=%v buff43=%v MoveSpeed=%d", target.GetBuff(29), target.GetBuff(43), target.MoveSpeed)
	}
}

func TestSkillStatusCancelSlowOnHastedTargetOnlyCancelsHaste(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 29)
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
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(target, &data.SkillInfo{SkillID: 43, BuffDuration: 120})

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      29,
		Target:       "attack",
		BuffDuration: 120,
		ActionID:     18,
	}, target.CharID)

	if target.HasBuff(43) || target.HasBuff(29) || target.MoveSpeed != 0 {
		t.Fatalf("緩速術對加速目標應只解除加速，buff43=%v buff29=%v MoveSpeed=%d", target.GetBuff(43), target.GetBuff(29), target.MoveSpeed)
	}
}
