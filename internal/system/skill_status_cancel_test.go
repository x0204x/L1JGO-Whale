package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
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
