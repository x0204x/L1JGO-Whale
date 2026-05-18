package system

import "testing"

func TestSkillArmorBreakProbabilityUsesJavaConfiguredChances(t *testing.T) {
	tests := []struct {
		name         string
		attackLevel  int
		defenseLevel int
		baseInt      int
		want         int
	}{
		{name: "攻方等級較高", attackLevel: 51, defenseLevel: 50, baseInt: 18, want: 60},
		{name: "攻防等級相同", attackLevel: 50, defenseLevel: 50, baseInt: 18, want: 40},
		{name: "攻方等級較低", attackLevel: 49, defenseLevel: 50, baseInt: 18, want: 20},
		{name: "INT 25 追加 Java 加成", attackLevel: 51, defenseLevel: 50, baseInt: 25, want: 61},
		{name: "INT 45 追加 Java 上限加成", attackLevel: 51, defenseLevel: 50, baseInt: 45, want: 65},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := armorBreakProbabilityByLevel(tt.attackLevel, tt.defenseLevel, tt.baseInt)
			if got != tt.want {
				t.Fatalf("破壞盔甲成功率未對齊 Java：got=%d want=%d", got, tt.want)
			}
		})
	}
}
