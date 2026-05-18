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
		X:         100,
		Y:         100,
		MapID:     4,
	})
	s := newSkillBuffTestSystem(t, ws)

	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 111, BuffDuration: 960})

	assertDressEvasionUpdateERPacket(t, drainSkillTestPackets(player.Session), 18)

	s.removeBuffAndRevert(player, 111)

	assertDressEvasionUpdateERPacket(t, drainSkillTestPackets(player.Session), 0)
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
