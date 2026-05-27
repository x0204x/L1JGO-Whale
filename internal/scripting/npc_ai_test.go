package scripting

import (
	"testing"

	"go.uber.org/zap"
)

func TestRunNpcAIMobSkillTypeFiveReturnsAreaShockStun(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.random = function(n) return n end"); err != nil {
		t.Fatalf("設定 Lua deterministic random 失敗: %v", err)
	}

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("type 5 mob skill 應回傳一個 AI 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_shock_stun" {
		t.Fatalf("Java type 5 areashock_stun 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
}

func TestRunNpcAIMobSkillTypeFiveIgnoresMpConsumeLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.random = function(n) return n end"); err != nil {
		t.Fatalf("設定 Lua deterministic random 失敗: %v", err)
	}

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areashock_stun 沒有 MP 檢查，MP 不足仍應回傳 type 5 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_shock_stun" {
		t.Fatalf("Java type 5 MP 不足仍應走 area_shock_stun，got=%q", cmds[0].Type)
	}
}

func TestRunNpcAIMobSkillTypeFiveTriggerRandomOneAlwaysPassesLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	for i := 0; i < 20; i++ {
		cmds := engine.RunNpcAI(AIContext{
			NpcID:      231008,
			HP:         100,
			MaxHP:      100,
			MP:         100,
			MaxMP:      100,
			TargetID:   1001,
			TargetDist: 1,
			CanAttack:  true,
			CanMove:    true,
			Skills: []MobSkillEntry{
				{
					Type:          5,
					TriggerRandom: 1,
					TriggerRange:  -2,
				},
			},
		})

		if len(cmds) != 1 || cmds[0].Type != "area_shock_stun" {
			t.Fatalf("Java L1MobSkillUse 以 trigger_random=1 必定通過，round=%d cmds=%+v", i, cmds)
		}
	}
}

func TestRunNpcAIMobSkillWithoutTriggerConditionIsNotUsableLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("無可用 mobskill 時仍應回到一般攻擊決策，got=%d", len(cmds))
	}
	if cmds[0].Type != "attack" {
		t.Fatalf("yiwei isSkillUseble 沒有 HP/同族HP/距離/次數條件時回傳 false，不應施放 mobskill，got=%+v", cmds[0])
	}
}

func TestRunNpcAIMobSkillPositiveTriggerRangeRequiresFarDistanceLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.random = function(n) return n end"); err != nil {
		t.Fatalf("設定 Lua deterministic random 失敗: %v", err)
	}

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 3,
		CanAttack:  true,
		CanMove:    false,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  6,
				TargetDist:    3,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("正 trigger_range 未達距離時應退回一般 AI，got=%d", len(cmds))
	}
	if cmds[0].Type == "area_shock_stun" {
		t.Fatalf("yiwei isTriggerDistance(triggerRange>0) 需要 distance>=triggerRange，不應在距離 3 施放 trigger_range=6 的技能")
	}

	cmds = engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 6,
		CanAttack:  true,
		CanMove:    false,
		Skills: []MobSkillEntry{
			{
				Type:          5,
				TriggerRandom: 1,
				TriggerRange:  6,
				TargetDist:    6,
			},
		},
	})

	if len(cmds) != 1 || cmds[0].Type != "area_shock_stun" {
		t.Fatalf("yiwei isTriggerDistance(triggerRange>0) 在 distance>=triggerRange 時應可施放，cmds=%+v", cmds)
	}
}

func TestRunNpcAIMobSkillSelfTargetStillUsesTriggerRangeLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.random = function(n) return n end"); err != nil {
		t.Fatalf("設定 Lua deterministic random 失敗: %v", err)
	}

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 3,
		CanAttack:  true,
		CanMove:    false,
		Skills: []MobSkillEntry{
			{
				Type:          2,
				SkillID:       1,
				TriggerRandom: 1,
				TriggerRange:  6,
				TargetDist:    3,
				ChangeTarget:  2,
			},
		},
	})

	if len(cmds) != 1 || cmds[0].Type == "skill" {
		t.Fatalf("yiwei isSkillUseble 在 change_target=2 前先檢查 TriRange，距離未達不應自我施法，cmds=%+v", cmds)
	}

	cmds = engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         100,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 6,
		CanAttack:  true,
		CanMove:    false,
		Skills: []MobSkillEntry{
			{
				Type:          2,
				SkillID:       1,
				TriggerRandom: 1,
				TriggerRange:  6,
				TargetDist:    6,
				ChangeTarget:  2,
			},
		},
	})

	if len(cmds) != 1 || cmds[0].Type != "skill" || cmds[0].ChangeTarget != 2 {
		t.Fatalf("yiwei change_target=2 在 TriRange 通過後才改成自我施法，cmds=%+v", cmds)
	}
}

func TestRunNpcAIRandomlySelectsAmongUsableMobSkillsLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.randomseed(7)"); err != nil {
		t.Fatalf("設定 Lua randomseed 失敗: %v", err)
	}

	seen := map[string]bool{}
	for i := 0; i < 40; i++ {
		cmds := engine.RunNpcAI(AIContext{
			NpcID:      231008,
			HP:         100,
			MaxHP:      100,
			MP:         100,
			MaxMP:      100,
			TargetID:   1001,
			TargetDist: 1,
			CanAttack:  true,
			CanMove:    true,
			Skills: []MobSkillEntry{
				{
					ActNo:         1,
					Type:          5,
					TriggerRandom: 1,
					TriggerRange:  -2,
				},
				{
					ActNo:         2,
					Type:          6,
					TriggerRandom: 1,
					TriggerRange:  -2,
				},
			},
		})
		if len(cmds) != 1 {
			t.Fatalf("yiwei 有可用 mobskill 時應回傳一個指令，round=%d got=%d", i, len(cmds))
		}
		seen[cmds[0].Type] = true
	}

	if !seen["area_shock_stun"] || !seen["area_cancellation"] {
		t.Fatalf("yiwei L1MobSkillUse.skillUse 會在所有可用技能中隨機挑選，不應永遠使用第一筆，seen=%v", seen)
	}
}

func TestRunNpcAITriggerRandomIsCheckedAfterCandidateSelectionLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()
	if err := engine.vm.DoString("math.randomseed(11)"); err != nil {
		t.Fatalf("設定 Lua randomseed 失敗: %v", err)
	}

	seen := map[string]bool{}
	for i := 0; i < 40; i++ {
		cmds := engine.RunNpcAI(AIContext{
			NpcID:      231008,
			HP:         100,
			MaxHP:      100,
			MP:         100,
			MaxMP:      100,
			TargetID:   1001,
			TargetDist: 1,
			CanAttack:  true,
			CanMove:    true,
			Skills: []MobSkillEntry{
				{
					ActNo:         1,
					Type:          5,
					TriggerRandom: 0,
					TriggerRange:  -2,
				},
				{
					ActNo:         2,
					Type:          6,
					TriggerRandom: 1,
					TriggerRange:  -2,
				},
			},
		})
		if len(cmds) != 1 {
			t.Fatalf("yiwei 選中機率失敗技能時應退回一般攻擊決策，round=%d got=%d", i, len(cmds))
		}
		seen[cmds[0].Type] = true
	}

	if !seen["attack"] || !seen["area_cancellation"] {
		t.Fatalf("yiwei trigger_random 在候選隨機選中後才判定，機率失敗時不應 fallback 到其他技能，seen=%v", seen)
	}
}

func TestRunNpcAIMobSkillTypeElevenIgnoresMpAndReturnsAreaSilenceLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          11,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         51,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areasilence 沒有 MP 檢查，MP 不足仍應回傳 type 11 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_silence" {
		t.Fatalf("Java type 11 areasilence 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 51 {
		t.Fatalf("area_silence 應保留 mobskill act_id，got=%d want=51", cmds[0].ActID)
	}
}

func TestRunNpcAIMobSkillTypeSixIgnoresMpAndReturnsAreaCancellationLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          6,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         47,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areacancellation 沒有 MP 檢查，MP 不足仍應回傳 type 6 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_cancellation" {
		t.Fatalf("Java type 6 areacancellation 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 47 {
		t.Fatalf("area_cancellation 應保留 mobskill act_id，got=%d want=47", cmds[0].ActID)
	}
}

func TestRunNpcAIMobSkillTypeSevenIgnoresMpAndReturnsAreaWeaponBreakLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      45670,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          7,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         47,
				ReuseDelay:    16000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.weapon_break 沒有 MP 檢查，MP 不足仍應回傳 type 7 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_weapon_break" {
		t.Fatalf("Java type 7 weapon_break 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 47 {
		t.Fatalf("area_weapon_break 應保留 mobskill act_id，got=%d want=47", cmds[0].ActID)
	}
	if cmds[0].ReuseDelay != 16000 {
		t.Fatalf("area_weapon_break 應保留 mobskill reuse_delay，got=%d want=16000", cmds[0].ReuseDelay)
	}
}

func TestRunNpcAIMobSkillTypeEightIgnoresMpAndReturnsAreaPotionTurnToDamageLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      45600,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          8,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         47,
				ReuseDelay:    12000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.potionturntodmg 沒有 MP 檢查，MP 不足仍應回傳 type 8 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_potion_turn_to_damage" {
		t.Fatalf("Java type 8 potionturntodmg 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 47 {
		t.Fatalf("area_potion_turn_to_damage 應保留 mobskill act_id，got=%d want=47", cmds[0].ActID)
	}
	if cmds[0].ReuseDelay != 12000 {
		t.Fatalf("area_potion_turn_to_damage 應保留 mobskill reuse_delay，got=%d want=12000", cmds[0].ReuseDelay)
	}
}

func TestRunNpcAIMobSkillTypeNineAndTenIgnoreMpAndReturnSpecialAreaStatusLikeJava(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  int
		want string
		act  int
	}{
		{name: "pollute-water-wave", typ: 9, want: "area_pollute_water_wave", act: 48},
		{name: "heal-turn-to-damage", typ: 10, want: "area_heal_turn_to_damage", act: 49},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, err := NewEngine("../../scripts", zap.NewNop())
			if err != nil {
				t.Fatalf("建立 Lua engine 失敗: %v", err)
			}
			defer engine.Close()

			cmds := engine.RunNpcAI(AIContext{
				NpcID:      45600,
				HP:         100,
				MaxHP:      100,
				MP:         0,
				MaxMP:      100,
				TargetID:   1001,
				TargetDist: 1,
				CanAttack:  true,
				CanMove:    true,
				Skills: []MobSkillEntry{
					{
						Type:          tc.typ,
						TriggerRandom: 1,
						TriggerRange:  -2,
						MpConsume:     25,
						ActID:         tc.act,
						ReuseDelay:    12000,
					},
				},
			})

			if len(cmds) != 1 {
				t.Fatalf("Java L1MobSkillUse type %d 沒有 MP 檢查，MP 不足仍應回傳指令，got=%d", tc.typ, len(cmds))
			}
			if cmds[0].Type != tc.want {
				t.Fatalf("Java type %d 不應走一般 skill 指令，got=%q want=%q", tc.typ, cmds[0].Type, tc.want)
			}
			if cmds[0].ActID != tc.act {
				t.Fatalf("%s 應保留 mobskill act_id，got=%d want=%d", tc.want, cmds[0].ActID, tc.act)
			}
			if cmds[0].ReuseDelay != 12000 {
				t.Fatalf("%s 應保留 mobskill reuse_delay，got=%d want=12000", tc.want, cmds[0].ReuseDelay)
			}
		})
	}
}

func TestRunNpcAIMobSkillTypeThirteenIgnoresMpAndReturnsAreaWindShackleLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      231008,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          13,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         47,
				GfxID:         11733,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areawindshackle 沒有 MP 檢查，MP 不足仍應回傳 type 13 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_wind_shackle" {
		t.Fatalf("Java type 13 areawindshackle 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 47 {
		t.Fatalf("area_wind_shackle 應保留 mobskill act_id，got=%d want=47", cmds[0].ActID)
	}
	if cmds[0].GfxID != 11733 {
		t.Fatalf("area_wind_shackle 應保留 mobskill gfx_id，got=%d want=11733", cmds[0].GfxID)
	}
}

func TestRunNpcAIMobSkillTypeTwelveIgnoresMpAndReturnsAreaDecayPotionLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      81163,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          12,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         19,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areadecaypotion 沒有 MP 檢查，MP 不足仍應回傳 type 12 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_decay_potion" {
		t.Fatalf("Java type 12 areadecaypotion 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 19 {
		t.Fatalf("area_decay_potion 應保留 mobskill act_id，got=%d want=19", cmds[0].ActID)
	}
}

func TestRunNpcAIMobSkillTypeSeventeenIgnoresMpAndReturnsAreaCurseParalyzeLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      55408,
		HP:         100,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          17,
				TriggerRandom: 1,
				TriggerRange:  -2,
				MpConsume:     25,
				ActID:         19,
				ReuseDelay:    10000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.areams 沒有 MP 檢查，MP 不足仍應回傳 type 17 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "area_curse_paralyze" {
		t.Fatalf("Java type 17 areams 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].ActID != 19 {
		t.Fatalf("area_curse_paralyze 應保留 mobskill act_id，got=%d want=19", cmds[0].ActID)
	}
	if cmds[0].ReuseDelay != 10000 {
		t.Fatalf("area_curse_paralyze 應保留 mobskill reuse_delay，got=%d want=10000", cmds[0].ReuseDelay)
	}
}

func TestRunNpcAIMobSkillTypeSixteenIgnoresMpAndReturnsSpawnEffectLikeJava(t *testing.T) {
	engine, err := NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	defer engine.Close()

	cmds := engine.RunNpcAI(AIContext{
		NpcID:      230262,
		HP:         80,
		MaxHP:      100,
		MP:         0,
		MaxMP:      100,
		TargetID:   1001,
		TargetDist: 1,
		CanAttack:  true,
		CanMove:    true,
		Skills: []MobSkillEntry{
			{
				Type:          16,
				TriggerRandom: 1,
				TriggerHP:     80,
				TriggerRange:  -2,
				MpConsume:     25,
				Leverage:      5,
				SummonID:      230270,
				ReuseDelay:    10000,
			},
		},
	})

	if len(cmds) != 1 {
		t.Fatalf("Java L1MobSkillUse.SpawnEffect 沒有 MP 檢查，MP 不足仍應回傳 type 16 指令，got=%d", len(cmds))
	}
	if cmds[0].Type != "spawn_effect" {
		t.Fatalf("Java type 16 SpawnEffect 不應走一般 skill 指令，got=%q", cmds[0].Type)
	}
	if cmds[0].SummonID != 230270 {
		t.Fatalf("spawn_effect 應保留 mobskill summon_id，got=%d want=230270", cmds[0].SummonID)
	}
	if cmds[0].Leverage != 5 {
		t.Fatalf("spawn_effect 應保留 mobskill leverage 作為存在秒數，got=%d want=5", cmds[0].Leverage)
	}
	if cmds[0].ReuseDelay != 10000 {
		t.Fatalf("spawn_effect 應保留 mobskill reuse_delay，got=%d want=10000", cmds[0].ReuseDelay)
	}
}
