package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func newSkillBuffTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	s := newSkillTestSystem(t, ws)

	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入技能資料失敗: %v", err)
	}
	buffIcons, err := data.LoadBuffIconTable("../../data/yaml/buff_icon_map.yaml")
	if err != nil {
		t.Fatalf("載入 buff 圖示資料失敗: %v", err)
	}

	s.deps.Skills = skills
	s.deps.BuffIcons = buffIcons
	return s
}

func TestSkillBuffCurePoisonSendsJavaSkillSound(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "cure-poison",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		KnownSpells: []int32{
			9,
		},
	})
	player.PoisonType = 1
	player.PoisonTicksLeft = 150
	player.PoisonDmgAmount = 5
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   9,
		TargetID:  player.CharID,
	})

	if player.PoisonType != 0 {
		t.Fatalf("解毒術應解除中毒，PoisonType=%d", player.PoisonType)
	}
	packets := drainSkillTestPackets(player.Session)
	if !hasSkillEffectPacket(packets, player.CharID, 871) {
		t.Fatal("Java CURE_POISON 會送目標 S_SkillSound 871，Go 也應送出")
	}
}

func TestSkillBuffCurePoisonSendsYiweiPostCastStatusRefresh(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "cure-poison-status",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		SP:        5,
		MR:        14,
		Dodge:     2,
		KnownSpells: []int32{
			9,
		},
	})
	player.PoisonType = 1
	player.PoisonTicksLeft = 150
	player.PoisonDmgAmount = 5
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   9,
		TargetID:  player.CharID,
	})

	packets := drainSkillTestPackets(player.Session)
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 解毒術後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 解毒術後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(player)) {
		t.Fatalf("yiwei sendGrfx 解毒術後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffHolyLightSendsYiweiEffectAndPostCastStatusRefresh(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "holy-light",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		SP:        6,
		MR:        21,
		Dodge:     4,
		KnownSpells: []int32{
			37,
		},
	})
	player.PoisonType = 1
	player.PoisonTicksLeft = 150
	player.PoisonDmgAmount = 5
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   37,
		TargetID:  player.CharID,
	})

	if player.PoisonType != 0 {
		t.Fatalf("聖潔之光應解除中毒，PoisonType=%d", player.PoisonType)
	}
	packets := drainSkillTestPackets(player.Session)
	if !hasSkillEffectPacket(packets, player.CharID, 227) {
		t.Fatalf("yiwei sendGrfx 聖潔之光會送目標 S_SkillSound 227，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 聖潔之光後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 聖潔之光後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(player)) {
		t.Fatalf("yiwei sendGrfx 聖潔之光後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffCurseBlindSendsYiweiEffectAndPostCastStatusRefresh(t *testing.T) {
	tests := []struct {
		name    string
		skillID int32
		gfxID   int32
	}{
		{name: "curse-blind", skillID: 20, gfxID: 746},
		{name: "darkness", skillID: 40, gfxID: 2175},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := world.NewState()
			player := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      tt.name,
				X:         100,
				Y:         100,
				MapID:     4,
				HP:        100,
				MaxHP:     100,
				MP:        100,
				MaxMP:     100,
				Level:     52,
				SP:        8,
				MR:        19,
				Dodge:     5,
				KnownSpells: []int32{
					tt.skillID,
				},
			})
			s := newSkillBuffTestSystem(t, ws)

			s.executeBuffSkill(player.Session, player, &data.SkillInfo{
				SkillID:      tt.skillID,
				Target:       "buff",
				BuffDuration: 16,
				ActionID:     19,
				CastGfx:      tt.gfxID,
			}, player.CharID)

			if !player.HasBuff(skillCurseBlindEffect) {
				t.Fatalf("%d 應套用致盲 buff", tt.skillID)
			}
			packets := drainSkillTestPackets(player.Session)
			if !hasSkillEffectPacket(packets, player.CharID, tt.gfxID) {
				t.Fatalf("yiwei sendGrfx 技能 %d 會送目標 S_SkillSound %d，packets=%v", tt.skillID, tt.gfxID, packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_SPMR 給 PC 目標，packets=%v", tt.skillID, packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_OwnCharStatus 給 PC 目標，packets=%v", tt.skillID, packets)
			}
			if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(player)) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", tt.skillID, packets)
			}
		})
	}
}

func TestSkillBuffCancellationSendsYiweiEffectAndPostCastStatusRefresh(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "cancellation",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		SP:        7,
		MR:        18,
		Dodge:     3,
		KnownSpells: []int32{
			44,
		},
	})
	s := newSkillBuffTestSystem(t, ws)
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 29, BuffDuration: 120})
	_ = drainSkillTestPackets(player.Session)

	s.executeBuffSkill(player.Session, player, &data.SkillInfo{
		SkillID:  44,
		Target:   "buff",
		ActionID: 19,
		CastGfx:  870,
	}, player.CharID)

	if player.HasBuff(29) {
		t.Fatal("Java CANCELLATION 應移除可相消 buff 29")
	}
	packets := drainSkillTestPackets(player.Session)
	if !hasSkillEffectPacket(packets, player.CharID, 870) {
		t.Fatalf("yiwei sendGrfx 魔法相消術會送目標 S_SkillSound 870，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 魔法相消術後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 魔法相消術後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(player)) {
		t.Fatalf("yiwei sendGrfx 魔法相消術後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffSleepAliasesSendYiweiEffectAndPostCastStatusRefresh(t *testing.T) {
	tests := []struct {
		name    string
		skillID int32
		gfxID   int32
	}{
		{name: "dark-blind", skillID: 103, gfxID: 2947},
		{name: "phantasm", skillID: 212, gfxID: 6530},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disablePlayerDebuffMRForStatusTest(t, tt.skillID)
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      tt.name + "-caster",
				X:         100,
				Y:         100,
				MapID:     4,
				HP:        100,
				MaxHP:     100,
				MP:        100,
				MaxMP:     100,
				Level:     52,
				KnownSpells: []int32{
					tt.skillID,
				},
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 2,
				Session:   newSkillTestSession(t, 2),
				CharID:    1002,
				Name:      tt.name + "-target",
				X:         101,
				Y:         100,
				MapID:     4,
				HP:        100,
				MaxHP:     100,
				MP:        100,
				MaxMP:     100,
				Level:     52,
				SP:        6,
				MR:        17,
				Dodge:     4,
			})
			s := newSkillBuffTestSystem(t, ws)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      tt.skillID,
				Target:       "buff",
				BuffDuration: 16,
				ActionID:     19,
				CastGfx:      tt.gfxID,
			}, target.CharID)

			if !target.HasBuff(66) {
				t.Fatalf("Java 技能 %d 應以 FOG_OF_SLEEPING(66) 作為實際睡眠效果", tt.skillID)
			}
			packets := drainSkillTestPackets(target.Session)
			if !hasSkillEffectPacket(packets, target.CharID, tt.gfxID) {
				t.Fatalf("yiwei sendGrfx 技能 %d 會送目標 S_SkillSound %d，packets=%v", tt.skillID, tt.gfxID, packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_SPMR 給 PC 目標，packets=%v", tt.skillID, packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_OwnCharStatus 給 PC 目標，packets=%v", tt.skillID, packets)
			}
			if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
				t.Fatalf("yiwei sendGrfx 技能 %d 後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", tt.skillID, packets)
			}
		})
	}
}

func TestSkillBuffElementalFallDownSendsYiweiPostCastStatusRefresh(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 133)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elemental-fall-caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ElfAttr:   8,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "elemental-fall-target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		SP:        5,
		MR:        16,
		Dodge:     6,
		FireRes:   10,
		WaterRes:  11,
		WindRes:   12,
		EarthRes:  13,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      133,
		Target:       "buff",
		BuffDuration: 32,
		ActionID:     19,
		CastGfx:      4396,
	}, target.CharID)

	if !target.HasBuff(133) || target.WindRes != -38 {
		t.Fatalf("Java ELEMENTAL_FALL_DOWN 應依 caster ElfAttr 降低單一抗性，WindRes=%d buff=%v", target.WindRes, target.GetBuff(133))
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasSkillEffectPacket(packets, target.CharID, 4396) {
		t.Fatalf("yiwei sendGrfx 弱化屬性會送目標 S_SkillSound 4396，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 弱化屬性後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 弱化屬性後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
		t.Fatalf("yiwei sendGrfx 弱化屬性後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffEarthBindSendsYiweiPostCastStatusRefresh(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 157)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "earth-bind-caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "earth-bind-target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		SP:        4,
		MR:        15,
		Dodge:     7,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      157,
		Target:       "buff",
		BuffDuration: 16,
		ActionID:     19,
		CastGfx:      2251,
	}, target.CharID)

	buff := target.GetBuff(157)
	if buff == nil || !target.Paralyzed {
		t.Fatalf("Java EARTH_BIND 應凍結玩家目標，Paralyzed=%v buff=%v", target.Paralyzed, buff)
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasSkillEffectPacket(packets, target.CharID, 2251) {
		t.Fatalf("yiwei sendGrfx 大地屏障會送目標 S_SkillSound 2251，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 大地屏障後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 大地屏障後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
		t.Fatalf("yiwei sendGrfx 大地屏障後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffArmorBreakSendsYiweiPostCastStatusRefresh(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "armor-break-caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
		Intel:     127,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "armor-break-target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     1,
		SP:        5,
		MR:        14,
		Dodge:     8,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      112,
		Target:       "none",
		BuffDuration: 8,
		ActionID:     19,
	}, target.CharID)

	buff := target.GetBuff(112)
	if buff == nil || buff.TicksLeft != 40 {
		t.Fatalf("Java ARMOR_BREAK 成功時應套用 8 秒 buff，ticks=%v buff=%v", 40, buff)
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasSkillEffectPacket(packets, target.CharID, 3400) {
		t.Fatalf("Java ARMOR_BREAK 成功時會送目標 S_SkillSound 3400，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 破壞盔甲後會送 S_SPMR 給 PC 目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 破壞盔甲後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
		t.Fatalf("yiwei sendGrfx 破壞盔甲後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
	}
}

func TestSkillBuffOppositeMoveSpeedSendsYiweiPostCastStatusRefresh(t *testing.T) {
	tests := []struct {
		name          string
		setupSkillID  int32
		castSkillID   int32
		castTarget    string
		casterName    string
		targetName    string
		sp            int16
		mr            int16
		dodge         int16
		wantOldAbsent int32
		wantNewAbsent int32
	}{
		{
			name:          "haste-cancels-slow",
			setupSkillID:  29,
			castSkillID:   43,
			castTarget:    "buff",
			casterName:    "haste-caster",
			targetName:    "slowed-target",
			sp:            5,
			mr:            13,
			dodge:         7,
			wantOldAbsent: 29,
			wantNewAbsent: 43,
		},
		{
			name:          "slow-cancels-haste",
			setupSkillID:  43,
			castSkillID:   29,
			castTarget:    "attack",
			casterName:    "slow-caster",
			targetName:    "hasted-target",
			sp:            6,
			mr:            14,
			dodge:         8,
			wantOldAbsent: 43,
			wantNewAbsent: 29,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disablePlayerDebuffMRForStatusTest(t, tt.castSkillID)
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      tt.casterName,
				X:         100,
				Y:         100,
				MapID:     4,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 2,
				Session:   newSkillTestSession(t, 2),
				CharID:    1002,
				Name:      tt.targetName,
				X:         101,
				Y:         100,
				MapID:     4,
				SP:        tt.sp,
				MR:        tt.mr,
				Dodge:     tt.dodge,
			})
			s := newSkillBuffTestSystem(t, ws)
			s.applyBuffEffect(target, &data.SkillInfo{SkillID: tt.setupSkillID, BuffDuration: 120})
			_ = drainSkillTestPackets(target.Session)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      tt.castSkillID,
				Target:       tt.castTarget,
				BuffDuration: 120,
				ActionID:     18,
			}, target.CharID)

			if target.HasBuff(tt.wantOldAbsent) || target.HasBuff(tt.wantNewAbsent) || target.MoveSpeed != 0 {
				t.Fatalf("Java 相反速度技能只應解除既有效果，old=%v new=%v MoveSpeed=%d", target.GetBuff(tt.wantOldAbsent), target.GetBuff(tt.wantNewAbsent), target.MoveSpeed)
			}
			packets := drainSkillTestPackets(target.Session)
			if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
				t.Fatalf("yiwei sendGrfx 相反速度技能後會送 S_SPMR 給 PC 目標，packets=%v", packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
				t.Fatalf("yiwei sendGrfx 相反速度技能後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
			}
			if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
				t.Fatalf("yiwei sendGrfx 相反速度技能後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
			}
		})
	}
}

func TestSkillBuffSlowFamilySkipsBraveSpeedFiveTargetLikeJava(t *testing.T) {
	tests := []struct {
		name    string
		skillID int32
	}{
		{name: "slow", skillID: 29},
		{name: "mass-slow", skillID: 76},
		{name: "entangle", skillID: 152},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disablePlayerDebuffMRForStatusTest(t, tt.skillID)
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "slow-caster",
				X:         100,
				Y:         100,
				MapID:     4,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID:  2,
				Session:    newSkillTestSession(t, 2),
				CharID:     1002,
				Name:       "third-speed-target",
				X:          101,
				Y:          100,
				MapID:      4,
				BraveSpeed: 5,
			})
			s := newSkillBuffTestSystem(t, ws)
			s.applyBuffEffect(target, &data.SkillInfo{SkillID: 43, BuffDuration: 120})
			_ = drainSkillTestPackets(target.Session)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      tt.skillID,
				Target:       "attack",
				BuffDuration: 120,
				ActionID:     18,
			}, target.CharID)

			if target.BraveSpeed != 5 || target.MoveSpeed != 1 || !target.HasBuff(43) || target.HasBuff(tt.skillID) {
				t.Fatalf("yiwei 對 BraveSpeed=5 目標會跳過緩速系列，BraveSpeed=%d MoveSpeed=%d buff43=%v slowBuff=%v",
					target.BraveSpeed, target.MoveSpeed, target.GetBuff(43), target.GetBuff(tt.skillID))
			}
		})
	}
}

func TestSkillBuffSlowFamilySkipsHasteItemEquippedTargetLikeJava(t *testing.T) {
	tests := []struct {
		name    string
		skillID int32
	}{
		{name: "slow", skillID: 29},
		{name: "mass-slow", skillID: 76},
		{name: "entangle", skillID: 152},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disablePlayerDebuffMRForStatusTest(t, tt.skillID)
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "slow-caster",
				X:         100,
				Y:         100,
				MapID:     4,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID:         2,
				Session:           newSkillTestSession(t, 2),
				CharID:            1002,
				Name:              "haste-item-target",
				X:                 101,
				Y:                 100,
				MapID:             4,
				HasteItemEquipped: 1,
			})
			s := newSkillBuffTestSystem(t, ws)
			s.applyBuffEffect(target, &data.SkillInfo{SkillID: 43, BuffDuration: 120})
			_ = drainSkillTestPackets(target.Session)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      tt.skillID,
				Target:       "attack",
				BuffDuration: 120,
				ActionID:     18,
			}, target.CharID)

			if target.HasteItemEquipped != 1 || target.MoveSpeed != 1 || !target.HasBuff(43) || target.HasBuff(tt.skillID) {
				t.Fatalf("yiwei 對 HasteItemEquipped>0 目標會跳過緩速系列，HasteItemEquipped=%d MoveSpeed=%d buff43=%v slowBuff=%v",
					target.HasteItemEquipped, target.MoveSpeed, target.GetBuff(43), target.GetBuff(tt.skillID))
			}
		})
	}
}

func TestSkillBuffHasteFamilySkipsHasteItemEquippedTargetLikeJava(t *testing.T) {
	tests := []struct {
		name    string
		skillID int32
	}{
		{name: "haste", skillID: 43},
		{name: "greater-haste", skillID: 54},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "haste-caster",
				X:         100,
				Y:         100,
				MapID:     4,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID:         2,
				Session:           newSkillTestSession(t, 2),
				CharID:            1002,
				Name:              "haste-item-target",
				X:                 101,
				Y:                 100,
				MapID:             4,
				MoveSpeed:         1,
				HasteItemEquipped: 1,
			})
			s := newSkillBuffTestSystem(t, ws)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      tt.skillID,
				Target:       "buff",
				BuffDuration: 120,
				ActionID:     18,
			}, target.CharID)

			if target.HasteItemEquipped != 1 || target.MoveSpeed != 1 || target.HasBuff(tt.skillID) {
				t.Fatalf("yiwei 對 HasteItemEquipped>0 目標會跳過 haste 系列，HasteItemEquipped=%d MoveSpeed=%d buff=%v",
					target.HasteItemEquipped, target.MoveSpeed, target.GetBuff(tt.skillID))
			}
		})
	}
}

func TestSkillBuffWindShackleSendsYiweiEffectAndPostCastStatusRefresh(t *testing.T) {
	tests := []struct {
		name            string
		existingTicks   int
		wantTicksUnkept bool
	}{
		{name: "fresh-target"},
		{name: "existing-target", existingTicks: 25, wantTicksUnkept: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disablePlayerDebuffMRForStatusTest(t, 167)
			ws := world.NewState()
			caster := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      tt.name + "-caster",
				X:         100,
				Y:         100,
				MapID:     4,
			})
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 2,
				Session:   newSkillTestSession(t, 2),
				CharID:    1002,
				Name:      tt.name + "-target",
				X:         101,
				Y:         100,
				MapID:     4,
				HP:        100,
				MaxHP:     100,
				MP:        100,
				MaxMP:     100,
				Level:     52,
				SP:        6,
				MR:        19,
				Dodge:     5,
			})
			if tt.existingTicks > 0 {
				target.AddBuff(&world.ActiveBuff{SkillID: 167, TicksLeft: tt.existingTicks})
			}
			s := newSkillBuffTestSystem(t, ws)

			s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
				SkillID:      167,
				Target:       "buff",
				BuffDuration: 16,
				ActionID:     19,
				CastGfx:      1799,
			}, target.CharID)

			buff := target.GetBuff(167)
			if buff == nil {
				t.Fatal("Java WIND_SHACKLE 應留下 167 buff")
			}
			if tt.wantTicksUnkept && buff.TicksLeft != tt.existingTicks {
				t.Fatalf("Java WIND_SHACKLE 對已有 167 目標不刷新時間，got=%d want=%d", buff.TicksLeft, tt.existingTicks)
			}
			packets := drainSkillTestPackets(target.Session)
			if !hasSkillEffectPacket(packets, target.CharID, 1799) {
				t.Fatalf("yiwei sendGrfx 風之枷鎖會送目標 S_SkillSound 1799，packets=%v", packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
				t.Fatalf("yiwei sendGrfx 風之枷鎖後會送 S_SPMR 給 PC 目標，packets=%v", packets)
			}
			if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
				t.Fatalf("yiwei sendGrfx 風之枷鎖後會送 S_OwnCharStatus 給 PC 目標，packets=%v", packets)
			}
			if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
				t.Fatalf("yiwei sendGrfx 風之枷鎖後會送 S_PacketBox.UPDATE_ER 給 PC 目標，packets=%v", packets)
			}
		})
	}
}

func TestSkillBuffSendsYiweiPostCastStatusRefreshToTarget(t *testing.T) {
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
		SP:        4,
		MR:        18,
		Dodge:     3,
	})

	s := newSkillBuffTestSystem(t, ws)
	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      68,
		ActionID:     19,
		CastGfx:      228,
		BuffDuration: 32,
	}, target.CharID)

	packets := drainSkillTestPackets(target.Session)
	if !hasOpcodePacket(packets, packet.S_OPCODE_MAGIC_STATUS) {
		t.Fatalf("yiwei sendGrfx 目標補助魔法後會送 S_SPMR 給目標，packets=%v", packets)
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_STATUS) {
		t.Fatalf("yiwei sendGrfx 目標補助魔法後會送 S_OwnCharStatus 給目標，packets=%v", packets)
	}
	if !hasYiweiUpdateERPacket(packets, calcPlayerErLikeYiwei(target)) {
		t.Fatalf("yiwei sendGrfx 目標補助魔法後會送 S_PacketBox.UPDATE_ER 給目標，packets=%v", packets)
	}
}

func TestSkillBuffWeakElementalBroadcastsNpcWeakAttrEffects(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "weak-elemental-caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{23},
	})
	viewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "weak-elemental-viewer",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:       world.NextNpcID(),
		NpcID:    45000,
		Impl:     "L1Monster",
		Name:     "weak target",
		X:        102,
		Y:        100,
		MapID:    4,
		HP:       100,
		MaxHP:    100,
		WeakAttr: 1 | 4,
	}
	ws.AddNpc(npc)
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: caster.SessionID,
		SkillID:   23,
		TargetID:  npc.ID,
	})

	packets := drainSkillTestPackets(viewer.Session)
	if !hasSkillEffectPacket(packets, npc.ID, 2169) {
		t.Fatal("Java WEAK_ELEMENTAL 的 weakAttr 地 bit 應廣播 S_SkillSound 2169")
	}
	if !hasSkillEffectPacket(packets, npc.ID, 2166) {
		t.Fatal("Java WEAK_ELEMENTAL 的 weakAttr 水 bit 應廣播 S_SkillSound 2166")
	}
}

func TestSkillBuffMovingAccelerationReplacesBraveStatus(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "darkelf",
		X:          100,
		Y:          100,
		MapID:      4,
		BraveSpeed: 1,
	})
	player.AddBuff(&world.ActiveBuff{
		SkillID:       handler.SkillStatusBrave,
		TicksLeft:     100,
		SetBraveSpeed: 1,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 101, BuffDuration: 960})

	if player.HasBuff(handler.SkillStatusBrave) {
		t.Fatal("行走加速應依 Java 互斥組移除勇敢藥水狀態")
	}
	if !player.HasBuff(101) || player.BraveSpeed != 4 {
		t.Fatalf("行走加速應保留自身 buff 並設定 brave speed 4，buff101=%v BraveSpeed=%d",
			player.GetBuff(101), player.BraveSpeed)
	}
}

func TestSkillBuffDressMightyCancelUsesJavaStrupType(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
		Str:       12,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 109, BuffDuration: 960})
	_ = drainSkillTestPackets(player.Session)

	s.removeBuffAndRevert(player, 109)

	var strup []byte
	for _, pkt := range drainSkillTestPackets(player.Session) {
		if len(pkt) >= 6 && pkt[0] == 166 {
			strup = pkt
		}
	}
	if strup == nil {
		t.Fatal("力量提升移除時應送出 S_Strup 取消圖示")
	}
	if got := strup[5]; got != 3 {
		t.Fatalf("力量提升移除應依 Java L1SkillStop 送 S_Strup type 3，got=%d", got)
	}
}

func TestSkillBuffDressDexterityCancelUsesJavaDexupType(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		X:         100,
		Y:         100,
		MapID:     4,
		Dex:       12,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 110, BuffDuration: 960})
	_ = drainSkillTestPackets(player.Session)

	s.removeBuffAndRevert(player, 110)

	var dexup []byte
	for _, pkt := range drainSkillTestPackets(player.Session) {
		if len(pkt) >= 5 && pkt[0] == 188 {
			dexup = pkt
		}
	}
	if dexup == nil {
		t.Fatal("敏捷提升移除時應送出 S_Dexup 取消圖示")
	}
	if got := dexup[4]; got != 3 {
		t.Fatalf("敏捷提升移除應依 Java L1SkillStop 送 S_Dexup type 3，got=%d", got)
	}
}

func TestSkillBuffDressEvasionUsesJavaUpdateERPacket(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "darkelf",
		ClassType: 4,
		Level:     52,
		Dex:       18,
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 111, BuffDuration: 960})

	// dark elf level 52 => 13, DEX 18 => 9, DRESS_EVASION => 18.
	assertDressEvasionUpdateERPacket(t, drainSkillTestPackets(player.Session), 40)

	s.removeBuffAndRevert(player, 111)

	assertDressEvasionUpdateERPacket(t, drainSkillTestPackets(player.Session), 22)
}

func assertDressEvasionUpdateERPacket(t *testing.T, packets [][]byte, wantER int16) {
	t.Helper()
	for _, pkt := range packets {
		if len(pkt) >= 4 && pkt[0] == packet.S_OPCODE_EVENT && pkt[1] == 132 {
			got := int16(uint16(pkt[2]) | uint16(pkt[3])<<8)
			if got != wantER {
				t.Fatalf("DRESS_EVASION UPDATE_ER 值不符 Java：got=%d want=%d packet=%v", got, wantER, pkt)
			}
			return
		}
	}
	t.Fatalf("DRESS_EVASION 應送 Java S_PacketBox.UPDATE_ER(132)，packets=%v", packets)
}

func TestSkillBuffSilenceBuffSetsAndClearsCastingLock(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:      64,
		BuffDuration: 16,
	}

	s.applyBuffEffect(target, skill)

	if !target.Silenced {
		t.Fatal("沉默技能套用後應禁止玩家施法")
	}
	if !target.HasBuff(64) {
		t.Fatal("沉默技能應註冊 active buff")
	}

	s.removeBuffAndRevert(target, 64)

	if target.Silenced {
		t.Fatal("沉默 buff 移除後應解除施法禁止")
	}
}

func TestSkillBuffSilenceAllowsJavaWhitelistedSkill(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "silenced-knight",
		X:           100,
		Y:           100,
		MapID:       4,
		HP:          200,
		MaxHP:       200,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{88},
	})
	player.Silenced = true
	player.AddBuff(&world.ActiveBuff{SkillID: 64, TicksLeft: 80, SetSilenced: true})
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   88,
		TargetID:  player.CharID,
	})

	if !player.HasBuff(88) {
		t.Fatal("Java 魔封狀態下仍允許 REDUCTION_ARMOR，應成功套用 buff 88")
	}
	if !player.Silenced || !player.HasBuff(64) {
		t.Fatalf("施放白名單技能不應解除沉默狀態，Silenced=%v buff64=%v", player.Silenced, player.GetBuff(64))
	}
}

func TestSkillBuffImmuneToHarmHalvesPlayerMagicDamage(t *testing.T) {
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
		HP:        100,
		MaxHP:     100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 68, TicksLeft: 150})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:  4,
		ActionID: 18,
		CastGfx:  167,
	}

	s.applySkillDamageToPlayer(caster.Session, caster, target, skill, 20, []*world.PlayerInfo{caster, target})

	if target.HP != 90 {
		t.Fatalf("聖結界應將玩家魔法傷害 20 減半為 10，HP=%d", target.HP)
	}
}

func TestSkillBuffImmuneToHarmSendsJavaIconI2H(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "immune-target",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 68, BuffDuration: 32})

	assertImmuneToHarmIconPacket(t, drainSkillTestPackets(player.Session), 32)
}

func assertImmuneToHarmIconPacket(t *testing.T, packets [][]byte, wantDuration uint16) {
	t.Helper()
	for _, pkt := range packets {
		if len(pkt) >= 4 && pkt[0] == packet.S_OPCODE_EVENT && pkt[1] == 40 {
			got := uint16(pkt[2]) | uint16(pkt[3])<<8
			if got != wantDuration {
				t.Fatalf("IMMUNE_TO_HARM ICON_I2H duration=%d，want=%d packet=%v", got, wantDuration, pkt)
			}
			return
		}
	}
	t.Fatalf("IMMUNE_TO_HARM 應送 Java S_PacketBox.ICON_I2H(40)，packets=%v", packets)
}

func TestSkillBuffImmuneToHarmHelperHalvesPhysicalDamage(t *testing.T) {
	target := &world.PlayerInfo{}
	target.AddBuff(&world.ActiveBuff{SkillID: 68, TicksLeft: 150})

	damage := applyImmuneToHarmDamage(target, 21)

	if damage != 10 {
		t.Fatalf("聖結界共用 helper 應將物理/魔法傷害向下減半，got=%d", damage)
	}
}

// Java REDUCTION_ARMOR(88, 增幅防禦) 在 5 條傷害路徑都做 flat 傷害減免：
// `L1AttackPc.java:1617-1620` (PvP physical) 為 `dmg -= (max(targetLvl,50)-50)/5 + 10`，
// `L1AttackNpc.java:437-440` (NPC→PC physical) / `L1MagicPc.java:1148,1296` / `L1MagicNpc.java:357`
// 為 `dmg -= (max(targetLvl,50)-50)/5 + 1`。Go 原本誤實作為 `ac = -4` Lua buff（完全不同機制），
// 本步補上 `applyReductionArmorDamage` helper 並先套用至 `npcMeleeAttack` (NPC→PC physical)。
// 測試覆蓋公式三個關鍵點：等級 50 邊界（floor）、+1/+10 路徑差、無 buff 不影響。
func TestSkillBuffReductionArmorDamageHelperMatchesJavaFormula(t *testing.T) {
	// 無 buff 88：不減免
	clean := &world.PlayerInfo{Level: 60}
	if got := applyReductionArmorDamage(clean, 100, false); got != 100 {
		t.Fatalf("無 buff 88 不應減免，got=%d", got)
	}

	// Level 50，NPC→PC physical：(50-50)/5 + 1 = 1 → 100-1=99
	lvl50 := &world.PlayerInfo{Level: 50}
	lvl50.AddBuff(&world.ActiveBuff{SkillID: 88, TicksLeft: 100})
	if got := applyReductionArmorDamage(lvl50, 100, false); got != 99 {
		t.Fatalf("Level 50 NPC→PC physical 應扣 1（(50-50)/5+1=1），got=%d", got)
	}

	// Level 50，PvP physical：(50-50)/5 + 10 = 10 → 100-10=90
	if got := applyReductionArmorDamage(lvl50, 100, true); got != 90 {
		t.Fatalf("Level 50 PvP physical 應扣 10（(50-50)/5+10=10），got=%d", got)
	}

	// Level 75，NPC→PC physical：(75-50)/5 + 1 = 6 → 100-6=94
	lvl75 := &world.PlayerInfo{Level: 75}
	lvl75.AddBuff(&world.ActiveBuff{SkillID: 88, TicksLeft: 100})
	if got := applyReductionArmorDamage(lvl75, 100, false); got != 94 {
		t.Fatalf("Level 75 NPC→PC physical 應扣 6（(75-50)/5+1=6），got=%d", got)
	}

	// Level 30（< 50）走 Java `Math.max(level, 50)` floor：等同 Level 50
	lvl30 := &world.PlayerInfo{Level: 30}
	lvl30.AddBuff(&world.ActiveBuff{SkillID: 88, TicksLeft: 100})
	if got := applyReductionArmorDamage(lvl30, 100, false); got != 99 {
		t.Fatalf("Level 30 應 floor 至 50（Java Math.max），扣 1，got=%d", got)
	}
}

func TestSkillBuffAdvanceSpiritUsesJavaBaseMax(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "advance-spirit",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        300,
		MaxHP:     560,
		MP:        120,
		MaxMP:     230,
		Level:     52,
		EquipBonuses: world.EquipStats{
			AddHP: 50,
			AddMP: 10,
		},
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 3001, TicksLeft: 100, DeltaMaxHP: 10, DeltaMaxMP: 20})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 79, BuffDuration: 120})

	buff := player.GetBuff(79)
	if buff == nil {
		t.Fatal("ADVANCE_SPIRIT 應建立 active buff")
	}
	if buff.DeltaMaxHP != 100 || buff.DeltaMaxMP != 40 {
		t.Fatalf("Java ADVANCE_SPIRIT 應以 baseMax/5 加成，DeltaMaxHP=%d DeltaMaxMP=%d", buff.DeltaMaxHP, buff.DeltaMaxMP)
	}
	if player.MaxHP != 660 || player.MaxMP != 270 {
		t.Fatalf("ADVANCE_SPIRIT 套用後 MaxHP/MaxMP 錯誤，MaxHP=%d MaxMP=%d", player.MaxHP, player.MaxMP)
	}
}

func TestSkillBuffCounterMagicConsumesBuffOnPlayerAttackSkill(t *testing.T) {
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
		HP:        100,
		MaxHP:     100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 31, TicksLeft: 150})
	s := newSkillBuffTestSystem(t, ws)
	skill := &data.SkillInfo{
		SkillID:         4,
		SkillLevel:      1,
		Target:          "attack",
		Type:            64,
		DamageValue:     20,
		DamageDice:      1,
		DamageDiceCount: 1,
		Ranged:          10,
		ActionID:        18,
		CastGfx:         167,
	}

	s.executeAttackSkill(caster.Session, caster, skill, target.CharID)

	if target.HP != 100 {
		t.Fatalf("魔法屏障抵消後不應受到攻擊魔法傷害，HP=%d", target.HP)
	}
	if target.HasBuff(31) {
		t.Fatal("魔法屏障抵消攻擊魔法後應被消耗")
	}
}

func TestSkillBuffCounterMagicDoesNotModifyAC(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "counter-magic",
		X:           100,
		Y:           100,
		MapID:       4,
		AC:          10,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{31},
	})
	player.Inv.AddItemWithID(6001, 40318, 1, "魔法寶石", 0, 0, true, 0)
	s := newSkillBuffTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   31,
		TargetID:  player.CharID,
	})

	if !player.HasBuff(31) {
		t.Fatal("魔法屏障應註冊 buff 31 供下一次可抵消魔法消耗")
	}
	if player.AC != 10 {
		t.Fatalf("Java COUNTER_MAGIC 不改 AC，got=%d", player.AC)
	}
}
