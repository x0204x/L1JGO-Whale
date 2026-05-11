package system

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillSelectionPolymorphClosesMonlistAfterSuccessfulChoice(t *testing.T) {
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
	caster.Inv.AddItem(40318, 1, "magic gem", 0, 0, true, 0)
	s := newPolymorphTestSystem(t, ws)

	s.QueueSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 67, TargetID: caster.CharID})
	s.Update(200 * time.Millisecond)
	drainSkillTestPackets(caster.Session)

	s.deps.Polymorph.UsePolySkill(caster.Session, caster, "floating eye")

	if !hasCloseListPacket(drainSkillTestPackets(caster.Session), caster.CharID) {
		t.Fatal("變形術 monlist 成功選擇後應送出 S_CloseList 關閉對話")
	}
}

func TestSkillSelectionPolymorphKeepsPendingOnMonlistNavigation(t *testing.T) {
	ws := world.NewState()
	caster := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID:        1,
		Session:          newSkillTestSession(t, 1),
		CharID:           1001,
		Name:             "caster",
		X:                100,
		Y:                100,
		MapID:            4,
		PendingPolySkill: true,
	})
	s := newPolymorphTestSystem(t, ws)

	s.deps.Polymorph.UsePolySkill(caster.Session, caster, "monlist_75")

	if !caster.PendingPolySkill {
		t.Fatal("monlist 分類頁 action 不應清除 PendingPolySkill")
	}
	if caster.TempCharGfx != 0 || caster.PolyID != 0 {
		t.Fatalf("monlist 分類頁 action 不應執行變身，TempCharGfx=%d PolyID=%d", caster.TempCharGfx, caster.PolyID)
	}
	if !hasHypertextPacket(drainSkillTestPackets(caster.Session), caster.CharID, "monlist_75") {
		t.Fatal("monlist 分類頁 action 應回送同名 HTML 讓客戶端切換頁面")
	}
}

func TestSkillSelectionSummonRingClosesSummonlistAfterSuccessfulChoice(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "summoner",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     60,
		Cha:       18,
		Inv:       world.NewInventory(),
	})
	ring := &world.InvItem{ObjectID: 7001, ItemID: 20284, Equipped: true}
	caster.Inv.Items = append(caster.Inv.Items, ring)
	caster.Equip.Set(world.SlotRing1, ring)
	if !hasSummonRing(caster) {
		t.Fatal("測試角色應被判定為已裝備召喚控制戒指")
	}
	caster.Inv.AddItem(40318, 3, "magic gem", 0, 0, true, 0)
	s := newElementalSummonTestSystem(t, ws)
	skill := s.deps.Skills.Get(51)
	if skill == nil {
		t.Fatal("測試資料缺少召喚術 skill 51")
	}

	s.deps.Summon.ExecuteSummonMonster(caster.Session, caster, skill, 263)

	if !hasCloseListPacket(drainSkillTestPackets(caster.Session), caster.CharID) {
		t.Fatal("召喚控制戒指 summonlist 成功選擇後應送出 S_CloseList 關閉對話")
	}
}

func TestSkillSelectionSummonRingRequiresClientSummonID(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "summoner",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     60,
		Cha:       18,
		Inv:       world.NewInventory(),
	})
	ring := &world.InvItem{ObjectID: 7001, ItemID: 20284, Equipped: true}
	caster.Inv.Items = append(caster.Inv.Items, ring)
	caster.Equip.Set(world.SlotRing1, ring)
	caster.Inv.AddItem(40318, 3, "magic gem", 0, 0, true, 0)
	s := newElementalSummonTestSystem(t, ws)
	skill := s.deps.Skills.Get(51)
	if skill == nil {
		t.Fatal("測試資料缺少召喚術 skill 51")
	}

	s.deps.Summon.ExecuteSummonMonster(caster.Session, caster, skill, 0)

	if caster.SummonSelectionMode {
		t.Fatal("yiwei 主流程不應由伺服器設定 summonlist 待選狀態")
	}
	if len(ws.GetSummonsByOwner(caster.CharID)) != 0 {
		t.Fatal("缺少客戶端 summonId 時不應召喚任何召喚獸")
	}
	if hasHypertextPacket(drainSkillTestPackets(caster.Session), caster.CharID, "summonlist") {
		t.Fatal("yiwei 主流程不應由伺服器主動送 summonlist")
	}
}

func hasCloseListPacket(packets [][]byte, objID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_HYPERTEXT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) != objID {
			continue
		}
		if pkt[5] == 0 {
			return true
		}
	}
	return false
}

func hasHypertextPacket(packets [][]byte, objID int32, htmlID string) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_HYPERTEXT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) != objID {
			continue
		}
		end := 5
		for end < len(pkt) && pkt[end] != 0 {
			end++
		}
		if string(pkt[5:end]) == htmlID {
			return true
		}
	}
	return false
}
