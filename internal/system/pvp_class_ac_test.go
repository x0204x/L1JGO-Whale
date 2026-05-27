package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestPvPMeleeContextUsesTargetClassACDefenseLikeYiwei(t *testing.T) {
	attacker := &world.PlayerInfo{
		Level:  50,
		Str:    25,
		HitMod: 10,
		DmgMod: 20,
	}
	target := &world.PlayerInfo{
		Level:     50,
		ClassType: 1,
		AC:        -60,
	}

	ctx := buildPvPMeleeCombatContext(attacker, target, 8)
	if ctx.TargetClassType != int(target.ClassType) {
		t.Fatalf("Yiwei PvP 物理傷害應依目標職業套用 AC 防禦，TargetClassType=%d want=%d", ctx.TargetClassType, target.ClassType)
	}
}

func TestPvPRangedContextUsesTargetClassACDefenseLikeYiwei(t *testing.T) {
	attacker := &world.PlayerInfo{
		Level:     50,
		Dex:       25,
		BowHitMod: 10,
		BowDmgMod: 20,
	}
	target := &world.PlayerInfo{
		Level:     50,
		ClassType: 3,
		AC:        -60,
	}

	ctx := buildPvPRangedCombatContext(attacker, target, 8, 1)
	if ctx.TargetClassType != int(target.ClassType) {
		t.Fatalf("Yiwei PvP 遠攻傷害應依目標職業套用 AC 防禦，TargetClassType=%d want=%d", ctx.TargetClassType, target.ClassType)
	}
}
