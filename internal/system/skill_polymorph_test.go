package system

import (
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestSkillPolymorphShapeChangeSelfConsumesGemOnceAndSelectionTransforms(t *testing.T) {
	ws := world.NewState()
	caster := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          200,
		MaxMP:       200,
		KnownSpells: []int32{67},
	})
	caster.Inv.AddItem(40318, 1, "魔法寶石", 0, 0, true, 0)
	s := newPolymorphTestSystem(t, ws)

	s.QueueSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 67, TargetID: caster.CharID})
	s.Update(200 * time.Millisecond)

	if !caster.PendingPolySkill {
		t.Fatal("變形術自我施放後應開啟 monlist 並等待選擇")
	}
	if gem := caster.Inv.FindByItemID(40318); gem != nil {
		t.Fatalf("變形術施法時應只消耗 1 顆魔法寶石，monlist 選擇時不得再消耗，剩餘=%d", gem.Count)
	}

	s.deps.Polymorph.UsePolySkill(caster.Session, caster, "floating eye")

	if caster.TempCharGfx != 29 || caster.PolyID != 29 {
		t.Fatalf("monlist 選擇後應變成 floating eye(29)，TempCharGfx=%d PolyID=%d", caster.TempCharGfx, caster.PolyID)
	}
	if caster.PendingPolySkill {
		t.Fatal("monlist 選擇完成後應清除待選狀態")
	}
}

func TestSkillPolymorphShapeChangeOtherPlayerRandomPolymorphWithoutControlRing(t *testing.T) {
	ws := world.NewState()
	caster := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          200,
		MaxMP:       200,
		KnownSpells: []int32{67},
	})
	target := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
		MR:        0,
		ClassID:   61,
	})
	caster.Inv.AddItem(40318, 1, "魔法寶石", 0, 0, true, 0)
	s := newPolymorphTestSystem(t, ws)

	s.QueueSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 67, TargetID: target.CharID})
	s.Update(200 * time.Millisecond)

	if target.TempCharGfx == 0 || !isShapeChangeRandomPoly(target.TempCharGfx) {
		t.Fatalf("沒有控制戒指的其他玩家應被隨機變形為 Java 清單之一，TempCharGfx=%d", target.TempCharGfx)
	}
	if caster.PendingPolySkill || target.PendingPolySkill {
		t.Fatal("對無控制戒指玩家施放時不應開啟 monlist")
	}
}

func TestSkillPolymorphShapeChangeOtherPlayerWithControlRingOpensTargetList(t *testing.T) {
	ws := world.NewState()
	caster := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          200,
		MaxMP:       200,
		KnownSpells: []int32{67},
	})
	target := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
		ClassID:   61,
	})
	ring := &world.InvItem{ObjectID: 7001, ItemID: 20281, Equipped: true}
	target.Inv.Items = append(target.Inv.Items, ring)
	target.Equip.Set(world.SlotRing1, ring)
	caster.Inv.AddItem(40318, 1, "魔法寶石", 0, 0, true, 0)
	s := newPolymorphTestSystem(t, ws)

	s.QueueSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 67, TargetID: target.CharID})
	s.Update(200 * time.Millisecond)

	if !target.PendingPolySkill {
		t.Fatal("有變形控制戒指的目標玩家應由目標端開啟 monlist")
	}
	if target.TempCharGfx != 0 {
		t.Fatalf("有控制戒指時不應先被隨機變形，TempCharGfx=%d", target.TempCharGfx)
	}
}

func TestSkillPolymorphShapeChangePolymorphsLowLevelNpcAndRestoresOnExpire(t *testing.T) {
	ws := world.NewState()
	caster := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          200,
		MaxMP:       200,
		KnownSpells: []int32{67},
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		GfxID: 100,
		Level: 10,
		X:     101,
		Y:     100,
		MapID: 4,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	caster.Inv.AddItem(40318, 1, "魔法寶石", 0, 0, true, 0)
	s := newPolymorphTestSystem(t, ws)

	s.QueueSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 67, TargetID: npc.ID})
	s.Update(200 * time.Millisecond)

	if npc.GfxID == 100 || !isShapeChangeRandomPoly(npc.GfxID) {
		t.Fatalf("50 級以下 NPC 應被隨機變形，GfxID=%d", npc.GfxID)
	}
	if npc.PolyOriginalGfxID != 100 || !npc.HasDebuff(67) {
		t.Fatalf("NPC 變形時應保存原始 GFX 並掛上 debuff，Original=%d HasDebuff=%v", npc.PolyOriginalGfxID, npc.HasDebuff(67))
	}

	removeNpcDebuffEffect(npc, 67, ws)

	if npc.GfxID != 100 || npc.PolyOriginalGfxID != 0 {
		t.Fatalf("NPC 變形到期後應還原原始 GFX，GfxID=%d Original=%d", npc.GfxID, npc.PolyOriginalGfxID)
	}
}

func newPolymorphTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	polys, err := data.LoadPolymorphTable("../../data/yaml/polymorph_list.yaml")
	if err != nil {
		t.Fatalf("讀取變形資料失敗: %v", err)
	}
	deps := &handler.Deps{
		World:  ws,
		Skills: skills,
		Polys:  polys,
		Log:    zap.NewNop(),
	}
	s := NewSkillSystem(deps)
	deps.Skill = s
	deps.Polymorph = NewPolymorphSystem(deps)
	return s
}

func addPolymorphTestPlayer(ws *world.State, p *world.PlayerInfo) *world.PlayerInfo {
	if p.Inv == nil {
		p.Inv = world.NewInventory()
	}
	if p.Known == nil {
		p.Known = world.NewKnownEntities()
	}
	if p.Level == 0 {
		p.Level = 50
	}
	if p.HP == 0 {
		p.HP = 100
	}
	if p.MaxHP == 0 {
		p.MaxHP = 100
	}
	if p.Intel == 0 {
		p.Intel = 18
	}
	ws.AddPlayer(p)
	return p
}

func isShapeChangeRandomPoly(polyID int32) bool {
	for _, id := range shapeChangeRandomPolyIDs {
		if id == polyID {
			return true
		}
	}
	return false
}
