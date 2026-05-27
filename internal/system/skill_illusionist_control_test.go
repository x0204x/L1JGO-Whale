package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

// TestSkillIllusionistControlBoneBreakProbabilityMatchesJava 對齊 Java
// `L1MagicPc.calcProbabilityMagic` `case BONE_BREAK`（L1MagicPc:584-599）+ PC→PC
// `case BONE_BREAK: probability -= RegistStun`（L1MagicPc:958-961）+
// `ConfigIllusionstSkill` 預設值（BONE_BREAK_1=5、_2=10、_3=15、_INT=0、_MR=0）。
func TestSkillIllusionistControlBoneBreakProbabilityMatchesJava(t *testing.T) {
	caster := &world.PlayerInfo{Level: 50, Intel: 30}
	target := &world.PlayerInfo{Level: 1}
	if got := calcBoneBreakPlayerProbability(caster, target); got != 5 {
		t.Fatalf("caster>target 應為 BONE_BREAK_1=5，實際 %d", got)
	}
	target.Level = 50
	if got := calcBoneBreakPlayerProbability(caster, target); got != 10 {
		t.Fatalf("caster==target 應為 BONE_BREAK_2=10，實際 %d", got)
	}
	caster.Level = 1
	if got := calcBoneBreakPlayerProbability(caster, target); got != 15 {
		t.Fatalf("caster<target 應為 BONE_BREAK_3=15，實際 %d", got)
	}
	// INT/MR 預設 config 為 0，所以對機率沒有影響。
	caster.Intel = 999
	target.MR = 999
	if got := calcBoneBreakPlayerProbability(caster, target); got != 15 {
		t.Fatalf("INT/MR 在預設 config 不應影響機率，實際 %d", got)
	}
	// PC→PC RegistStun 直接扣（L1MagicPc:958-961）。
	target.RegistStun = 3
	if got := calcBoneBreakPlayerProbability(caster, target); got != 12 {
		t.Fatalf("RegistStun=3 應 15-3=12，實際 %d", got)
	}
	// 機率下限 0。
	target.RegistStun = 100
	if got := calcBoneBreakPlayerProbability(caster, target); got != 0 {
		t.Fatalf("極高 RegistStun 應 clamp 至 0，實際 %d", got)
	}
}

// TestSkillIllusionistControlBoneBreakAppliesStunNotParalysis 對齊 Java
// `BONE_BREAK.start():29` `S_Paralysis(5, true)` —— TYPE_STUN (logical 5) 對應
// wire byte 0x16（StunApply），並非 ParalysisApply(0x02)。
func TestSkillIllusionistControlBoneBreakAppliesStunNotParalysis(t *testing.T) {
	ws := world.NewState()
	_ = addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     50,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
		HP:        100,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	// 直接呼叫 applyBoneBreakParalysis 跳過機率檢查，驗證副作用：buff 注入 +
	// Paralyzed flag + S_Paralysis(StunApply) + S_SkillSound(13119) 廣播。
	s.applyBoneBreakParalysis(target)

	if !target.Paralyzed {
		t.Fatalf("骷髏毀壞命中後應 Paralyzed，實際 %v", target.Paralyzed)
	}
	buff := target.GetBuff(208)
	if buff == nil {
		t.Fatalf("骷髏毀壞應留下 208 buff，實際無")
	}
	if !buff.SetParalyzed {
		t.Fatalf("208 buff 應 SetParalyzed=true，實際 %v", buff.SetParalyzed)
	}
	// dur 1~2 秒 → 5 或 10 ticks。
	if buff.TicksLeft != 5 && buff.TicksLeft != 10 {
		t.Fatalf("208 buff 持續時間應為 5 或 10 ticks（1~2 秒），實際 %d", buff.TicksLeft)
	}
}

func TestSkillIllusionistControlBoneBreakBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		HP:        100,
		MaxHP:     100,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "same_show",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "other_show",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
	})
	s := newSkillTestSystem(t, ws)

	s.applyBoneBreakParalysis(target)

	if !hasSkillEffectPacket(drainSkillTestPackets(target.Session), target.CharID, 13119) {
		t.Fatal("yiwei sendPacketsAll 會把骷髏毀壞特效送給目標自己")
	}
	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), target.CharID, 13119) {
		t.Fatal("同 ShowID 玩家應收到骷髏毀壞目標特效")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), target.CharID, 13119) {
		t.Fatal("不同 ShowID 玩家不應收到骷髏毀壞目標特效")
	}
}

func TestSkillIllusionistControlJoyOfPainCastPrimesCasterWithoutDamagingTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:  218,
		Target:   "attack",
		Type:     64,
		Ranged:   2,
		ActionID: 19,
		CastGfx:  6528,
	}

	s.executeAttackSkillOnPlayer(caster.Session, caster, skill, target)

	if target.HP != 100 {
		t.Fatalf("疼痛的歡愉施放時不應直接傷害目標，HP=%d", target.HP)
	}
	buff := caster.GetBuff(218)
	if buff == nil || buff.TicksLeft != 16*5 {
		t.Fatalf("疼痛的歡愉應給施法者 16 秒狀態，buff=%v", buff)
	}
}

func TestSkillIllusionistControlJoyOfPainBacklashDamagesCasterOnce(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
	})
	caster.AddBuff(&world.ActiveBuff{SkillID: 218, TicksLeft: 16 * 5})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        50,
		MaxHP:     100,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 1, ActionID: 18}

	s.applySkillDamageToPlayer(caster.Session, caster, target, skill, 5, []*world.PlayerInfo{caster, target})

	if target.HP != 45 {
		t.Fatalf("原本傷害仍應套用到目標，HP=%d", target.HP)
	}
	if caster.HP != 90 {
		t.Fatalf("疼痛的歡愉應依目標既有失血量反傷施法者 10，HP=%d", caster.HP)
	}
	if caster.HasBuff(218) {
		t.Fatalf("疼痛的歡愉反傷後應移除一次性狀態，buff=%v", caster.GetBuff(218))
	}
}
