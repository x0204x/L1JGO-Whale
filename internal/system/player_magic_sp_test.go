package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestCalcMagicBonusLikeJava(t *testing.T) {
	cases := []struct {
		intel int
		want  int
	}{
		{5, -2},
		{8, -1},
		{11, 0},
		{14, 1},
		{17, 2},
		{18, 3},
		{24, 9},
		{25, 10},
		{35, 10},
		{36, 11},
		{42, 11},
		{43, 12},
		{49, 12},
		{50, 13},
		{60, 13},
	}

	for _, tc := range cases {
		if got := calcMagicBonusLikeJava(tc.intel); got != tc.want {
			t.Fatalf("INT %d magic bonus 應為 %d，實際 %d", tc.intel, tc.want, got)
		}
	}
}

func TestPlayerSpellPowerIncludesMagicLevelMagicBonusAndExtraSPLikeJava(t *testing.T) {
	player := &world.PlayerInfo{
		ClassType: 3, // 法師
		Level:     40,
		Intel:     18,
		SP:        2, // Go 欄位對應 Java _other_sp：裝備/buff 額外魔攻
	}

	if got, want := calcPlayerTrueSPLikeJava(player), 13; got != want {
		t.Fatalf("true SP 應為 magicLevel(10)+magicBonus(3)=%d，實際 %d", want, got)
	}
	if got, want := calcPlayerSPLikeJava(player), 15; got != want {
		t.Fatalf("Java getSp() 應為 trueSP+extraSP=%d，實際 %d", want, got)
	}
}

func TestMindBreakDamageUsesFullJavaSP(t *testing.T) {
	caster := &world.PlayerInfo{
		ClassType: 6,  // 幻術師
		Level:     60, // magic level 10
		Intel:     18, // magic bonus 3
		SP:        2,
	}

	if got, want := calcMindBreakDamage(caster), int32(57); got != want {
		t.Fatalf("心靈破壞應使用 Java getSp()*3.8，got=%d want=%d", got, want)
	}
}

func TestPlayerBaseStatsSubtractEquipmentAndBuffsLikeJava(t *testing.T) {
	player := &world.PlayerInfo{
		Str:   30,
		Dex:   31,
		Intel: 32,
		EquipBonuses: world.EquipStats{
			AddStr: 2,
			AddDex: 3,
			AddInt: 4,
		},
		ActiveBuffs: map[int32]*world.ActiveBuff{
			42: {
				DeltaStr:   5,
				DeltaDex:   6,
				DeltaIntel: 7,
			},
		},
	}

	if got, want := calcPlayerBaseStrLikeJava(player), 23; got != want {
		t.Fatalf("Base STR 應扣掉裝備與 buff STR delta，got=%d want=%d", got, want)
	}
	if got, want := calcPlayerBaseDexLikeJava(player), 22; got != want {
		t.Fatalf("Base DEX 應扣掉裝備與 buff DEX delta，got=%d want=%d", got, want)
	}
	if got, want := calcPlayerBaseIntLikeJava(player), 21; got != want {
		t.Fatalf("Base INT 應扣掉裝備與 buff INT delta，got=%d want=%d", got, want)
	}
}

func TestPlayerErUsesClassLevelDexAndDodgeLikeYiwei(t *testing.T) {
	player := &world.PlayerInfo{
		ClassType: 1, // knight: level / 4
		Level:     52,
		Dex:       18,
		Dodge:     2,
	}

	// knight level 52 => 13，DEX 18 => (18-8)/2+4 = 9，再加既有 Dodge 2。
	if got, want := calcPlayerErLikeYiwei(player), int16(24); got != want {
		t.Fatalf("yiwei getEr 應含職業等級、DEX 與 Dodge 加成，got=%d want=%d", got, want)
	}
}

func TestPlayerErStrikerGaleReturnsZeroLikeYiwei(t *testing.T) {
	player := &world.PlayerInfo{
		ClassType: 2, // elf: level / 6
		Level:     60,
		Dex:       30,
		Dodge:     50,
		ActiveBuffs: map[int32]*world.ActiveBuff{
			174: {SkillID: 174, TicksLeft: 10},
		},
	}

	if got := calcPlayerErLikeYiwei(player); got != 0 {
		t.Fatalf("yiwei getEr 在 STRIKER_GALE 狀態下應直接回 0，got=%d", got)
	}
}
