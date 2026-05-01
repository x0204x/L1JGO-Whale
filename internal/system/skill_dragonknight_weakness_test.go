package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillDragonKnightWeaknessDragonKnightChainswordWeaknessProgressesAndResetsOnTargetChange(t *testing.T) {
	rolls := []int{0, 20, 20}
	oldRoll := dragonKnightWeaknessRoll
	dragonKnightWeaknessRoll = func(_ int) int {
		if len(rolls) == 0 {
			return 99
		}
		roll := rolls[0]
		rolls = rolls[1:]
		return roll
	}
	t.Cleanup(func() {
		dragonKnightWeaknessRoll = oldRoll
	})

	player := &world.PlayerInfo{ClassType: 5}

	applyDragonKnightWeaknessExposure(player, 2001, 1001, "chainsword")
	if player.WeaknessLevel != 1 || player.WeaknessTargetID != 2001 {
		t.Fatalf("第一次鎖鏈劍攻擊應建立弱點 Lv1，level=%d target=%d", player.WeaknessLevel, player.WeaknessTargetID)
	}
	applyDragonKnightWeaknessExposure(player, 2001, 1001, "chainsword")
	if player.WeaknessLevel != 2 {
		t.Fatalf("第二次鎖鏈劍攻擊應推進弱點 Lv2，level=%d", player.WeaknessLevel)
	}
	applyDragonKnightWeaknessExposure(player, 2001, 1001, "chainsword")
	if player.WeaknessLevel != 3 {
		t.Fatalf("第三次鎖鏈劍攻擊應推進弱點 Lv3，level=%d", player.WeaknessLevel)
	}

	applyDragonKnightWeaknessExposure(player, 3001, 1001, "chainsword")
	if player.WeaknessLevel != 0 || player.WeaknessTargetID != 3001 {
		t.Fatalf("切換目標應清除既有弱點，level=%d target=%d", player.WeaknessLevel, player.WeaknessTargetID)
	}
}

func TestSkillDragonKnightWeaknessFoeSlayerConsumesWeaknessBonusAndClearsWeakness(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:        1,
		Session:          newSkillTestSession(t, 1),
		CharID:           1001,
		Name:             "caster",
		X:                100,
		Y:                100,
		MapID:            4,
		Level:            70,
		Str:              35,
		DmgMod:           100,
		HP:               100,
		MaxHP:            100,
		WeaknessLevel:    3,
		WeaknessTargetID: 1002,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     1,
		HP:        1000,
		MaxHP:     1000,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:    187,
		Target:     "attack",
		Type:       64,
		Ranged:     2,
		DamageDice: 1,
		ActionID:   18,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if target.HP > 700 {
		t.Fatalf("屠宰者應吃弱點 Lv3 的三段加傷，HP=%d", target.HP)
	}
	if caster.WeaknessLevel != 0 || caster.WeaknessTargetID != 0 {
		t.Fatalf("屠宰者第三段後應清除弱點，level=%d target=%d", caster.WeaknessLevel, caster.WeaknessTargetID)
	}
}
