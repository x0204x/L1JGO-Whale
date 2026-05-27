package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestWeaponSkillDamageReductionUsesJavaMagicHitNotSP(t *testing.T) {
	caster := &world.PlayerInfo{
		Intel:            23,
		OriginalMagicHit: 20,
		SP:               0,
	}
	target := &world.NpcInfo{MR: 80}

	got := calcWeaponSkillDmgReduction(caster, target, 100, attrNone)
	if int(got) != 71 {
		t.Fatalf("武器魔法 MR 減傷應使用 INT 魔法命中 + OriginalMagicHit，而不是額外 SP；got=%.2f want=71", got)
	}
}
