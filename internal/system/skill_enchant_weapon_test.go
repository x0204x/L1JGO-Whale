package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newEnchantWeaponTestSystem(t *testing.T, ws *world.State) *SkillSystem {
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
	return &SkillSystem{deps: &handler.Deps{
		World:     ws,
		Items:     items,
		Skills:    skills,
		Scripting: engine,
		Log:       zap.NewNop(),
	}}
}

func TestSkillEnchantWeaponStrengthHelmHalvesMpConsume(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          20,
		MaxMP:       20,
		KnownSpells: []int32{12},
	})
	weapon := player.Inv.AddItemWithID(7001, 31, 1, "長劍", 0, 1000, false, 1)
	helm := player.Inv.AddItemWithID(7002, 20015, 1, "力量魔法頭盔", 0, 1000, false, 1)
	helm.Equipped = true
	player.Equip.Set(world.SlotHelm, helm)
	s := newEnchantWeaponTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   12,
		TargetID:  weapon.ObjectID,
	})

	if player.MP != 10 {
		t.Fatalf("力量魔法頭盔施放擬似魔法武器應依 Java 消耗半數 MP，MP=%d", player.MP)
	}
	if weapon.DmgByMagic != 2 || weapon.DmgMagicExpiry != 1800*5 {
		t.Fatalf("擬似魔法武器仍應套用武器附魔，dmg=%d expiry=%d", weapon.DmgByMagic, weapon.DmgMagicExpiry)
	}
}

func TestSkillWeaponItemEnchantBroadcastsOnlySameShowLikeJava(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T, s *SkillSystem, player *world.PlayerInfo)
	}{
		{
			name: "blessed armor",
			run: func(t *testing.T, s *SkillSystem, player *world.PlayerInfo) {
				t.Helper()
				armor := player.Inv.AddItemWithID(7101, 20089, 1, "小藤甲", 0, 1000, false, 1)
				s.executeArmorEnchant(player.Session, player, &data.SkillInfo{
					SkillID:      21,
					BuffDuration: 1,
					ActionID:     19,
					CastGfx:      748,
				}, armor.ObjectID)
			},
		},
		{
			name: "create magical weapon",
			run: func(t *testing.T, s *SkillSystem, player *world.PlayerInfo) {
				t.Helper()
				weapon := player.Inv.AddItemWithID(7102, 31, 1, "長劍", 0, 1000, false, 1)
				s.executeCreateMagicalWeapon(player.Session, player, &data.SkillInfo{
					SkillID:      8,
					BuffDuration: 1,
					ActionID:     19,
					CastGfx:      748,
				}, weapon.ObjectID)
			},
		},
		{
			name: "bring stone",
			run: func(t *testing.T, s *SkillSystem, player *world.PlayerInfo) {
				t.Helper()
				stone := player.Inv.AddItemWithID(7103, 40320, 1, "黑魔石", 40320, 10, true, 1)
				player.Level = 200
				player.Wis = 50
				s.executeBringStone(player.Session, player, &data.SkillInfo{
					SkillID:      100,
					BuffDuration: 1,
					ActionID:     19,
					CastGfx:      748,
				}, stone.ObjectID)
			},
		},
		{
			name: "enchant weapon",
			run: func(t *testing.T, s *SkillSystem, player *world.PlayerInfo) {
				t.Helper()
				weapon := player.Inv.AddItemWithID(7104, 31, 1, "長劍", 0, 1000, false, 1)
				s.executeTargetedWeaponEnchant(player.Session, player, &data.SkillInfo{
					SkillID:      12,
					BuffDuration: 1800,
					ActionID:     19,
					CastGfx:      748,
				}, weapon.ObjectID)
			},
		},
		{
			name: "bless weapon",
			run: func(t *testing.T, s *SkillSystem, player *world.PlayerInfo) {
				t.Helper()
				weapon := player.Inv.AddItemWithID(7105, 31, 1, "長劍", 0, 1000, false, 1)
				weapon.Equipped = true
				player.Equip.Set(world.SlotWeapon, weapon)
				s.executeBlessWeaponEnchant(player.Session, player, &data.SkillInfo{
					SkillID:      48,
					BuffDuration: 1200,
					ActionID:     19,
					CastGfx:      748,
				}, player.CharID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := world.NewState()
			player := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "caster",
				X:         100,
				Y:         100,
				MapID:     4,
				ShowID:    9,
			})
			sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 2,
				Session:   newSkillTestSession(t, 2),
				CharID:    1002,
				Name:      "same-show",
				X:         101,
				Y:         100,
				MapID:     4,
				ShowID:    9,
			})
			otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 3,
				Session:   newSkillTestSession(t, 3),
				CharID:    1003,
				Name:      "other-show",
				X:         101,
				Y:         101,
				MapID:     4,
				ShowID:    10,
			})
			s := newEnchantWeaponTestSystem(t, ws)

			tc.run(t, s, player)

			samePackets := drainSkillTestPackets(sameShow.Session)
			otherPackets := drainSkillTestPackets(otherShow.Session)
			if !hasActionGfxPacket(samePackets, player.CharID, 19) {
				t.Fatalf("同 ShowID 玩家應收到 %s 施法動作", tc.name)
			}
			if !hasSkillEffectPacket(samePackets, player.CharID, 748) {
				t.Fatalf("同 ShowID 玩家應收到 %s 施法特效", tc.name)
			}
			if hasActionGfxPacket(otherPackets, player.CharID, 19) {
				t.Fatalf("不同 ShowID 玩家不應收到 %s 施法動作", tc.name)
			}
			if hasSkillEffectPacket(otherPackets, player.CharID, 748) {
				t.Fatalf("不同 ShowID 玩家不應收到 %s 施法特效", tc.name)
			}
		})
	}
}

func TestSkillBlessWeaponRejectsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
		ShowID:    20,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other-show-target",
		X:         101,
		Y:         100,
		MapID:     4,
		ShowID:    21,
	})
	weapon := target.Inv.AddItemWithID(7201, 31, 1, "長劍", 0, 1000, false, 1)
	weapon.Equipped = true
	target.Equip.Set(world.SlotWeapon, weapon)
	s := newEnchantWeaponTestSystem(t, ws)

	s.executeBlessWeaponEnchant(caster.Session, caster, &data.SkillInfo{
		SkillID:      48,
		BuffDuration: 1200,
		ActionID:     19,
		CastGfx:      748,
	}, target.CharID)

	if weapon.DmgByMagic != 0 || weapon.DmgMagicExpiry != 0 {
		t.Fatalf("不同 ShowID 目標不應被祝福魔法武器套用，dmg=%d expiry=%d", weapon.DmgByMagic, weapon.DmgMagicExpiry)
	}
}

func TestSkillDetectionStrengthHelmHalvesMpConsume(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          15,
		MaxMP:       15,
		KnownSpells: []int32{13},
	})
	helm := player.Inv.AddItemWithID(7002, 20015, 1, "力量魔法頭盔", 0, 1000, false, 1)
	helm.Equipped = true
	player.Equip.Set(world.SlotHelm, helm)
	s := newEnchantWeaponTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   13,
	})

	if player.MP != 8 {
		t.Fatalf("力量魔法頭盔施放無所遁形術應依 Java 15>>1 消耗 7 MP，MP=%d", player.MP)
	}
}
