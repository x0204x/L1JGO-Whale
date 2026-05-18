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
