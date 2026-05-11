package system

import (
	"math/rand"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newMagicScrollTestSystem(t *testing.T, ws *world.State) *ItemUseSystem {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入技能表失敗: %v", err)
	}
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	deps := &handler.Deps{
		World:     ws,
		Items:     items,
		Skills:    skills,
		Scripting: engine,
		Log:       zap.NewNop(),
	}
	deps.Equip = NewEquipSystem(deps)
	return NewItemUseSystem(deps)
}

func TestBlankMagicScrollWizardCreatesSpellScrollAndConsumesSkillResources(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		ClassType: 3,
		MP:        20,
		MaxMP:     20,
	})
	scroll := player.Inv.AddItemWithID(7001, 40090, 1, "blank", 859, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)

	if !s.UseBlankMagicScroll(player.Session, player, scroll, 0) {
		t.Fatal("法師使用一級空白魔法卷軸應成功")
	}

	if player.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("空白魔法卷軸應被消耗")
	}
	if made := player.Inv.FindByItemID(40859); made == nil || made.Count != 1 {
		t.Fatalf("應產生 40859 魔法卷軸，got=%v", made)
	}
	if player.MP != 16 {
		t.Fatalf("寫入 skill 1 應消耗 4 MP，got=%d want=16", player.MP)
	}
}

func TestBlankMagicScrollRejectsNonWizardWithoutConsuming(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		ClassType: 1,
		MP:        20,
		MaxMP:     20,
	})
	scroll := player.Inv.AddItemWithID(7001, 40090, 1, "blank", 859, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)

	if s.UseBlankMagicScroll(player.Session, player, scroll, 0) {
		t.Fatal("非法師不應可使用空白魔法卷軸")
	}
	if scroll.Count != 1 || player.Inv.FindByObjectID(scroll.ObjectID) == nil {
		t.Fatal("拒絕時不應消耗空白魔法卷軸")
	}
	if player.Inv.FindByItemID(40859) != nil {
		t.Fatal("拒絕時不應產生魔法卷軸")
	}
	if !hasServerMessage(drainSkillTestPackets(player.Session), 264) {
		t.Fatal("非法師使用空白魔法卷軸應送出訊息 264")
	}
}

func TestBlankMagicScrollRejectsSkillAboveScrollLevel(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		ClassType: 3,
		MP:        20,
		MaxMP:     20,
	})
	scroll := player.Inv.AddItemWithID(7001, 40090, 1, "blank", 859, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)

	if s.UseBlankMagicScroll(player.Session, player, scroll, 8) {
		t.Fatal("一級空白魔法卷軸不應可寫入第 9 格技能")
	}
	if scroll.Count != 1 || player.Inv.FindByObjectID(scroll.ObjectID) == nil {
		t.Fatal("等級不符時不應消耗空白魔法卷軸")
	}
	if player.Inv.FindByItemID(40867) != nil {
		t.Fatal("等級不符時不應產生魔法卷軸")
	}
	if !hasServerMessage(drainSkillTestPackets(player.Session), 591) {
		t.Fatal("等級不符應送出訊息 591")
	}
}

func TestMagicScrollCastsWithoutKnownSpellOrMpConsumption(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		MP:        0,
		MaxMP:     20,
		AC:        10,
	})
	scroll := player.Inv.AddItemWithID(7001, 49281, 1, "spell-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	if !s.UseMagicScroll(player.Session, player, scroll, info, player.CharID, 0, 0) {
		t.Fatal("魔法卷軸不應要求角色已學技能或足夠 MP")
	}
	if player.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("成功施放後魔法卷軸應被消耗")
	}
	if player.MP != 0 {
		t.Fatalf("魔法卷軸施放不應扣 MP，got=%d", player.MP)
	}
	if !player.HasBuff(42) {
		t.Fatal("49281 應施放 skill 42 並套用 buff")
	}
}

func TestMagicScrollRejectsLongSelfTargetWithoutConsuming(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		MP:        0,
		MaxMP:     20,
	})
	scroll := player.Inv.AddItemWithID(7001, 40862, 1, "spell-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	if s.UseMagicScroll(player.Session, player, scroll, info, player.CharID, int16(player.X), int16(player.Y)) {
		t.Fatal("遠距離魔法卷軸不應允許以自己為目標")
	}
	if scroll.Count != 1 || player.Inv.FindByObjectID(scroll.ObjectID) == nil {
		t.Fatal("目標不合法時不應消耗魔法卷軸")
	}
	if !hasServerMessage(drainSkillTestPackets(player.Session), 281) {
		t.Fatal("遠距離魔法卷軸目標不合法應送出訊息 281")
	}
}

func TestMagicScrollWeaponTargetConsumesScrollAndCasts(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		MP:        0,
		MaxMP:     20,
	})
	weapon := player.Inv.AddItemWithID(7001, 31, 1, "weapon", 0, 1000, false, 1)
	scroll := player.Inv.AddItemWithID(7002, 40870, 1, "weapon-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	if !s.UseMagicScroll(player.Session, player, scroll, info, weapon.ObjectID, 0, 0) {
		t.Fatal("40870 應以武器為目標施放 skill 12")
	}
	if player.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("成功施放後魔法卷軸應被消耗")
	}
	if player.HasBuff(12) {
		t.Fatal("40870 不應註冊成角色 buff 12")
	}
	if weapon.DmgByMagic != 2 || weapon.HitByMagic != 0 || weapon.DmgMagicExpiry != 1800*5 {
		t.Fatalf("40870 應套用 Java 擬似魔法武器附魔值，dmg=%d hit=%d expiry=%d", weapon.DmgByMagic, weapon.HitByMagic, weapon.DmgMagicExpiry)
	}
}

func TestMagicScrollBlessWeaponAppliesToEquippedWeapon(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		MP:        0,
		MaxMP:     20,
	})
	weapon := player.Inv.AddItemWithID(7001, 31, 1, "weapon", 0, 1000, false, 1)
	weapon.Equipped = true
	player.Equip.Set(world.SlotWeapon, weapon)
	scroll := player.Inv.AddItemWithID(7002, 49282, 1, "bless-weapon-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	if !s.UseMagicScroll(player.Session, player, scroll, info, player.CharID, 0, 0) {
		t.Fatal("49282 祝福魔法武器卷軸應可對自己已裝備武器施放")
	}
	if player.Inv.FindByObjectID(scroll.ObjectID) != nil {
		t.Fatal("祝福魔法武器卷軸成功施放後應消耗")
	}
	if player.HasBuff(48) {
		t.Fatal("祝福魔法武器應是武器附魔，不應註冊成角色 buff 48")
	}
	if weapon.DmgByMagic != 2 || weapon.HitByMagic != 2 || weapon.DmgMagicExpiry != 1200*5 {
		t.Fatalf("祝福魔法武器應套用 Java 武器附魔值，dmg=%d hit=%d expiry=%d", weapon.DmgByMagic, weapon.HitByMagic, weapon.DmgMagicExpiry)
	}
	if player.DmgMod != 2 {
		t.Fatalf("裝備能力應立即重算祝福魔法武器傷害，got=%d want=2", player.DmgMod)
	}
}

func TestMagicScrollNpcDebuffResistDoesNotSendCastFailMessage(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     1,
		Intel:     0,
		MP:        0,
		MaxMP:     20,
	})
	target := &world.NpcInfo{
		ID:    2001,
		Name:  "high-mr-target",
		X:     102,
		Y:     100,
		MapID: 4,
		Level: 100,
		HP:    100,
		MaxHP: 100,
		MR:    10000,
	}
	ws.AddNpc(target)

	scroll := player.Inv.AddItemWithID(7001, 40869, 20, "poison-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	rand.Seed(1)
	for i := 0; i < 20; i++ {
		player.SkillDelayUntil = time.Time{}
		if !s.UseMagicScroll(player.Session, player, scroll, info, target.ID, int16(target.X), int16(target.Y)) {
			t.Fatalf("第 %d 次毒咒魔法卷軸應可使用", i+1)
		}
	}

	packets := drainSkillTestPackets(player.Session)
	if hasServerMessage(packets, skillMsgCastFail) {
		t.Fatal("魔法卷軸 TYPE_SPELLSC 對 NPC debuff 被 MR 抵抗時不應送一般施咒失敗訊息")
	}
	if !hasOpcodePacket(packets, packet.S_OPCODE_ACTION) {
		t.Fatal("魔法卷軸即使 debuff 未命中，也應保留施法動作封包")
	}
}

func TestMagicScrollFreezingBreathUsesDirectedRangeSkillVisual(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		MP:        0,
		MaxMP:     20,
	})
	target := &world.NpcInfo{
		ID:    2001,
		Name:  "target",
		X:     104,
		Y:     100,
		MapID: 4,
		HP:    100,
		MaxHP: 100,
		MR:    0,
	}
	near := &world.NpcInfo{
		ID:    2002,
		Name:  "near",
		X:     105,
		Y:     100,
		MapID: 4,
		HP:    100,
		MaxHP: 100,
		MR:    0,
	}
	ws.AddNpc(target)
	ws.AddNpc(near)
	scroll := player.Inv.AddItemWithID(7001, 40880, 1, "freezing-breath-scroll", 863, 630, true, 1)
	s := newMagicScrollTestSystem(t, ws)
	info := s.deps.Items.Get(scroll.ItemID)

	if !s.UseMagicScroll(player.Session, player, scroll, info, target.ID, int16(target.X), int16(target.Y)) {
		t.Fatal("40880 寒冰氣息魔法卷軸應成功施放")
	}
	packets := drainSkillTestPackets(player.Session)

	if !hasOpcodePacket(packets, packet.S_OPCODE_RANGESKILLS) {
		t.Fatalf("寒冰氣息是有方向範圍攻擊，應送 S_RangeSkill opcode=%d", packet.S_OPCODE_RANGESKILLS)
	}
	if hasOpcodePacket(packets, packet.S_OPCODE_ATTACK) {
		t.Fatal("寒冰氣息不應用單體 S_UseAttackSkill 表現，客戶端會顯示施咒取消")
	}
	if target.HP >= 100 || near.HP >= 100 {
		t.Fatalf("寒冰氣息應傷害目標與範圍內 NPC，targetHP=%d nearHP=%d", target.HP, near.HP)
	}
}
