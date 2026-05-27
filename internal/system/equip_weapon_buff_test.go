package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestUnequipWeaponRemovesWeaponDependentBuffsLikeJava(t *testing.T) {
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("讀取道具 YAML 失敗: %v", err)
	}
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "weapon-buff-player",
		X:         100,
		Y:         100,
		MapID:     4,
		AC:        10,
	})
	weapon := &world.InvItem{ObjectID: 5001, ItemID: 1, Name: "dagger", Bless: 1, Equipped: true}
	player.Equip.Set(world.SlotWeapon, weapon)
	player.CurrentWeapon = world.WeaponVisualID("dagger")
	skill := newSkillBuffTestSystem(t, ws)
	equip := NewEquipSystem(&handler.Deps{
		World: ws,
		Items: items,
		Skill: skill,
		Log:   zap.NewNop(),
	})
	skill.applyBuffEffect(player, &data.SkillInfo{SkillID: 91, BuffDuration: 120})
	skill.applyBuffEffect(player, &data.SkillInfo{SkillID: 155, BuffDuration: 120})
	_ = drainSkillTestPackets(player.Session)

	equip.UnequipSlot(player.Session, player, world.SlotWeapon)

	if player.HasBuff(91) || player.HasBuff(155) {
		t.Fatalf("yiwei 卸下武器會解除 COUNTER_BARRIER/FIRE_BLESS，buff91=%v buff155=%v", player.GetBuff(91), player.GetBuff(155))
	}
	if player.BraveSpeed != 0 || player.AC != 10 {
		t.Fatalf("卸下武器解除 buff 後應還原速度與 AC，BraveSpeed=%d AC=%d", player.BraveSpeed, player.AC)
	}
	if !hasBravePacketWithDuration(drainSkillTestPackets(player.Session), player.CharID, 0, 0) {
		t.Fatal("解除 FIRE_BLESS 應送 S_SkillBrave(type=0,duration=0)")
	}
}

func hasBravePacketWithDuration(packets [][]byte, objectID int32, braveType byte, duration uint16) bool {
	for _, pkt := range packets {
		if len(pkt) < 10 || pkt[0] != packet.S_OPCODE_SKILLBRAVE {
			continue
		}
		gotID := int32(pkt[1]) | int32(pkt[2])<<8 | int32(pkt[3])<<16 | int32(pkt[4])<<24
		gotDuration := uint16(pkt[6]) | uint16(pkt[7])<<8
		if gotID == objectID && pkt[5] == braveType && gotDuration == duration {
			return true
		}
	}
	return false
}
