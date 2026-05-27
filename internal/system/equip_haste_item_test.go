package system

import (
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestEquipArmorHasteItemAppliesPermanentHasteAndClearsSpeedBuffsLikeJava(t *testing.T) {
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品 YAML 失敗: %v", err)
	}
	itemInfo := items.Get(20235)
	if itemInfo == nil {
		t.Fatal("正式物品表缺少伊娃之盾 20235")
	}
	if !itemInfo.HasteItem {
		t.Fatal("伊娃之盾 20235 必須標記 haste_item，才能對齊 yiwei 裝備 haste item 行為")
	}

	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "eva-shield-player",
		X:         100,
		Y:         100,
		MapID:     4,
		ClassType: 1,
	})
	shield := &world.InvItem{ObjectID: 5001, ItemID: 20235, Name: "伊娃之盾", Bless: 1}
	skill := newSkillTestSystem(t, ws)
	equip := NewEquipSystem(&handler.Deps{
		World: ws,
		Items: items,
		Skill: skill,
		Log:   zap.NewNop(),
	})
	skill.applyBuffEffect(player, &data.SkillInfo{SkillID: 29, BuffDuration: 120})
	_ = drainSkillTestPackets(player.Session)

	equip.EquipArmor(player.Session, player, shield, itemInfo)

	if player.HasteItemEquipped != 1 || player.MoveSpeed != 1 || player.HasteTicks != 0 || player.HasBuff(29) {
		t.Fatalf("yiwei 裝備 haste item 會清速度技能並套永久 haste，HasteItemEquipped=%d MoveSpeed=%d HasteTicks=%d buff29=%v",
			player.HasteItemEquipped, player.MoveSpeed, player.HasteTicks, player.GetBuff(29))
	}
	if !hasSpeedPacketWithDuration(drainSkillTestPackets(player.Session), player.CharID, 1, 0xffff) {
		t.Fatal("yiwei 裝備 haste item 會送 S_SkillHaste(type=1,duration=-1) 給自己")
	}
}

func TestUnequipArmorHasteItemClearsPermanentHasteLikeJava(t *testing.T) {
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品 YAML 失敗: %v", err)
	}
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:         1,
		Session:           newSkillTestSession(t, 1),
		CharID:            1001,
		Name:              "eva-shield-player",
		X:                 100,
		Y:                 100,
		MapID:             4,
		ClassType:         1,
		MoveSpeed:         1,
		HasteItemEquipped: 1,
	})
	shield := &world.InvItem{ObjectID: 5001, ItemID: 20235, Name: "伊娃之盾", Bless: 1, Equipped: true}
	player.Equip.Set(world.SlotShield, shield)
	equip := NewEquipSystem(&handler.Deps{
		World: ws,
		Items: items,
		Log:   zap.NewNop(),
	})

	equip.UnequipSlot(player.Session, player, world.SlotShield)

	if player.HasteItemEquipped != 0 || player.MoveSpeed != 0 {
		t.Fatalf("yiwei 脫下最後一件 haste item 會清永久 haste，HasteItemEquipped=%d MoveSpeed=%d",
			player.HasteItemEquipped, player.MoveSpeed)
	}
	if !hasSpeedPacketWithDuration(drainSkillTestPackets(player.Session), player.CharID, 0, 0) {
		t.Fatal("yiwei 脫下最後一件 haste item 會送 S_SkillHaste(type=0,duration=0)")
	}
}

func TestInitEquipStatsRestoresHasteItemEquippedLikeJava(t *testing.T) {
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("讀取道具 YAML 失敗: %v", err)
	}
	player := &world.PlayerInfo{
		CharID:    1001,
		Name:      "eva-shield-player",
		ClassType: 1,
		AC:        10,
	}
	player.Equip.Set(world.SlotShield, &world.InvItem{ObjectID: 5001, ItemID: 20235, Name: "伊娃之盾", Bless: 1, Equipped: true})
	equip := NewEquipSystem(&handler.Deps{
		Items:  items,
		Config: &config.Config{Gameplay: config.GameplayConfig{BaseAC: 10}},
		Log:    zap.NewNop(),
	})

	equip.InitEquipStats(player)

	if player.HasteItemEquipped != 1 || player.MoveSpeed != 1 || player.HasteTicks != 0 {
		t.Fatalf("yiwei 進入世界時已裝備 haste item 應恢復永久 haste，HasteItemEquipped=%d MoveSpeed=%d HasteTicks=%d",
			player.HasteItemEquipped, player.MoveSpeed, player.HasteTicks)
	}
}

func hasSpeedPacketWithDuration(packets [][]byte, objectID int32, speedType byte, duration uint16) bool {
	for _, pkt := range packets {
		if len(pkt) < 8 || pkt[0] != packet.S_OPCODE_SPEED {
			continue
		}
		gotID := int32(pkt[1]) | int32(pkt[2])<<8 | int32(pkt[3])<<16 | int32(pkt[4])<<24
		gotDuration := uint16(pkt[6]) | uint16(pkt[7])<<8
		if gotID == objectID && pkt[5] == speedType && gotDuration == duration {
			return true
		}
	}
	return false
}
