package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillClanAuraSolidCarriageRequiresShieldOrGuarderAndAddsER(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		AC:        10,
		Dodge:     2,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 90, BuffDuration: 64, Target: "none", ActionID: 19}

	s.executeSelfSkill(player.Session, player, skill)
	if player.HasBuff(90) || player.Dodge != 2 || player.AC != 10 {
		t.Fatalf("未裝備盾/臂甲時堅固防護不應套用，buff90=%v Dodge=%d AC=%d", player.GetBuff(90), player.Dodge, player.AC)
	}

	player.Equip.Set(world.SlotShield, &world.InvItem{ObjectID: 5001, ItemID: 20230, Equipped: true})
	s.executeSelfSkill(player.Session, player, skill)
	if !player.HasBuff(90) || player.Dodge != 17 || player.AC != 10 {
		t.Fatalf("堅固防護應依 Java 給 ER/Dodge +15 且不改 AC，buff90=%v Dodge=%d AC=%d", player.GetBuff(90), player.Dodge, player.AC)
	}
}

func TestSkillClanAuraGlowingAuraAppliesToPartyWithJavaStats(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "prince",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	outsider := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "outsider",
		X:         102,
		Y:         100,
		MapID:     4,
	})
	ws.Parties.CreateParty(caster.CharID, member.CharID, world.PartyTypeNormal)
	s := newSkillTestSystem(t, ws)

	s.executeSelfSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      114,
		BuffDuration: 640,
		Target:       "none",
		TargetTo:     targetToParty,
		Area:         -1,
		ActionID:     19,
	})

	for _, p := range []*world.PlayerInfo{caster, member} {
		if !p.HasBuff(114) || p.HitMod != 5 || p.DmgMod != 5 || p.BowHitMod != 0 || p.MR != 0 {
			t.Fatalf("激勵士氣應套用隊伍並依 Java 給近戰命中/傷害 +5，player=%s buff=%v Hit=%d Dmg=%d BowHit=%d MR=%d",
				p.Name, p.GetBuff(114), p.HitMod, p.DmgMod, p.BowHitMod, p.MR)
		}
	}
	if outsider.HasBuff(114) {
		t.Fatalf("非隊伍成員不應取得激勵士氣，buff=%v", outsider.GetBuff(114))
	}
}

func TestSkillClanAuraShiningAuraAppliesToClanMembers(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "prince",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		AC:        10,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		AC:        10,
	})
	outsider := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "outsider",
		X:         102,
		Y:         100,
		MapID:     4,
		ClanID:    8,
		AC:        10,
	})
	s := newSkillTestSystem(t, ws)

	s.executeSelfSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      115,
		BuffDuration: 640,
		Target:       "none",
		TargetTo:     targetToClan | targetToParty,
		Area:         -1,
		ActionID:     19,
	})

	for _, p := range []*world.PlayerInfo{caster, member} {
		if !p.HasBuff(115) || p.AC != 2 {
			t.Fatalf("鋼鐵士氣應套用同血盟並 AC -8，player=%s buff=%v AC=%d", p.Name, p.GetBuff(115), p.AC)
		}
	}
	if outsider.HasBuff(115) || outsider.AC != 10 {
		t.Fatalf("非同血盟不應取得鋼鐵士氣，buff=%v AC=%d", outsider.GetBuff(115), outsider.AC)
	}
}

func TestSkillClanAuraBraveAuraIsProcFlagNotFlatDamage(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "prince",
		X:         100,
		Y:         100,
		MapID:     4,
		DmgMod:    3,
	})
	s := newSkillTestSystem(t, ws)
	s.applyBuffEffect(player, &data.SkillInfo{SkillID: 117, BuffDuration: 640})

	if !player.HasBuff(117) || player.DmgMod != 3 {
		t.Fatalf("衝擊士氣應只作為機率增傷旗標，不應固定 DmgMod +5，buff=%v DmgMod=%d", player.GetBuff(117), player.DmgMod)
	}
	if got := braveAuraDamageWithRoll(player, 100, 0); got != 150 {
		t.Fatalf("衝擊士氣命中機率內應造成 1.5 倍傷害，got=%d", got)
	}
	if got := braveAuraDamageWithRoll(player, 100, 99); got != 100 {
		t.Fatalf("衝擊士氣機率外不應增傷，got=%d", got)
	}
}

func TestSkillClanAuraTrueTargetRegistersBuffWithoutStatBonus(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "prince",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{
		SkillID:      113,
		BuffDuration: 5,
		Target:       "buff",
		ActionID:     19,
	}, target.CharID, "focus")

	if !target.HasBuff(113) || target.HitMod != 0 {
		t.Fatalf("精準目標應掛 113 狀態但不提供固定命中加成，buff=%v HitMod=%d", target.GetBuff(113), target.HitMod)
	}
}
