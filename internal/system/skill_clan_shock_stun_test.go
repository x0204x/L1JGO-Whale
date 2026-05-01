package system

import (
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func attachShockStunItemTable(t *testing.T, s *SkillSystem) {
	t.Helper()
	items, err := data.LoadItemTable(
		filepath.Join("..", "..", "data", "yaml", "weapon_list.yaml"),
		filepath.Join("..", "..", "data", "yaml", "armor_list.yaml"),
		filepath.Join("..", "..", "data", "yaml", "etcitem_list.yaml"),
	)
	if err != nil {
		t.Fatalf("讀取物品資料失敗: %v", err)
	}
	s.deps.Items = items
}

func TestSkillClanShockStunCallClanCanTargetSameClanAcrossMaps(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		MP:        100,
		MaxMP:     100,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         33000,
		Y:         33000,
		MapID:     304,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 116, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if member.PendingYesNoType != 729 || member.PendingYesNoData != caster.CharID {
		t.Fatalf("呼喚盟友應可跨地圖對同血盟成員送 729 確認，Pending=(%d,%d)", member.PendingYesNoType, member.PendingYesNoData)
	}
}

func TestSkillClanShockStunRunClanTeleportsToAllowedClanMap(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		MP:        100,
		MaxMP:     100,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         32710,
		Y:         32810,
		MapID:     304,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 118, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if caster.X != member.X || caster.Y != member.Y || caster.MapID != member.MapID {
		t.Fatalf("援護盟友應傳送到 0/4/304 的同血盟目標位置，got=(%d,%d,%d) want=(%d,%d,%d)",
			caster.X, caster.Y, caster.MapID, member.X, member.Y, member.MapID)
	}
}

func TestSkillClanShockStunRunClanRejectsDisallowedTargetMap(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     100,
		ClanID:    7,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     100,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 118, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if caster.X != 100 || caster.Y != 100 || caster.MapID != 100 {
		t.Fatalf("援護盟友不應傳送到 Java 禁止的目標地圖，got=(%d,%d,%d)", caster.X, caster.Y, caster.MapID)
	}
}

func TestSkillClanShockStunShockStunRequiresTwoHandSwordForPlayerTarget(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 87)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
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
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("未裝備雙手劍時衝擊之暈不應套用，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}

	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)
	if !target.HasBuff(87) || !target.Paralyzed {
		t.Fatalf("裝備雙手劍時衝擊之暈應套用暈眩，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
}

func TestSkillClanShockStunBounceAttackAddsHitOnlyOnce(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		HitMod:    2,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 89, BuffDuration: 64, Target: "none", ActionID: 19}

	s.executeSelfSkill(player.Session, player, skill)
	s.executeSelfSkill(player.Session, player, skill)

	if !player.HasBuff(89) || player.HitMod != 8 {
		t.Fatalf("尖刺盔甲應依 Java 給 Hit +6 且重放不疊加，buff=%v HitMod=%d", player.GetBuff(89), player.HitMod)
	}
}
