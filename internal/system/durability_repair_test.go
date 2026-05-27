package system

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestRepairWeaponSendsYiweiMessages(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "repair",
		Inv:       world.NewInventory(),
		Str:       18,
		Con:       18,
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	weapon.Durability = 3
	player.Inv.AddItemWithID(102, world.AdenaItemID, 1000, "adena", 0, 0, true, 1)
	ws.AddPlayer(player)

	s := NewNpcServiceSystem(&handler.Deps{World: ws, Items: items, Log: zap.NewNop()})
	if !s.RepairWeapon(sess, player, weapon, 600) {
		t.Fatal("金幣足夠時修復應成功")
	}

	packets := drainSkillTestPackets(sess)
	if weapon.Durability != 0 {
		t.Fatalf("修復後耐久應歸零，got=%d", weapon.Durability)
	}
	if player.Inv.GetAdena() != 400 {
		t.Fatalf("應扣除 600 金幣，got=%d", player.Inv.GetAdena())
	}
	if !hasServerMessage(packets, 464) {
		t.Fatal("修復完成應送出 yiwei 訊息 464")
	}
}

func TestRepairWeaponSendsNotEnoughAdenaMessage(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "repair",
		Inv:       world.NewInventory(),
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	weapon.Durability = 3
	player.Inv.AddItemWithID(102, world.AdenaItemID, 100, "adena", 0, 0, true, 1)
	ws.AddPlayer(player)

	s := NewNpcServiceSystem(&handler.Deps{World: ws, Items: items, Log: zap.NewNop()})
	if s.RepairWeapon(sess, player, weapon, 600) {
		t.Fatal("金幣不足時修復應失敗")
	}

	packets := drainSkillTestPackets(sess)
	if weapon.Durability != 3 {
		t.Fatalf("修復失敗不得改變耐久，got=%d", weapon.Durability)
	}
	if player.Inv.GetAdena() != 100 {
		t.Fatalf("修復失敗不得扣金幣，got=%d", player.Inv.GetAdena())
	}
	if !hasServerMessage(packets, 189) {
		t.Fatal("金幣不足應送出 yiwei 訊息 189")
	}
}

func TestUseWhetstoneRepairsOneDurabilityAndConsumesStone(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "hone",
		Inv:       world.NewInventory(),
		Str:       18,
		Con:       18,
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	weapon.Durability = 2
	stone := player.Inv.AddItemWithID(102, 40317, 2, "whetstone", 0, 0, true, 1)
	ws.AddPlayer(player)

	s := NewItemUseSystem(&handler.Deps{World: ws, Items: items, Log: zap.NewNop()})
	if !s.UseWhetstone(sess, player, stone, weapon.ObjectID) {
		t.Fatal("磨刀石對損壞武器應成功處理")
	}
	packets := drainSkillTestPackets(sess)

	if weapon.Durability != 1 {
		t.Fatalf("磨刀石一次應修復 1 點耐久，got=%d", weapon.Durability)
	}
	if stone.Count != 1 {
		t.Fatalf("磨刀石應消耗 1 個，got=%d", stone.Count)
	}
	if !hasServerMessage(packets, 463) {
		t.Fatal("尚未完全修復時應送出 yiwei 訊息 463")
	}
	if !hasDurabilityStatusPacket(packets, weapon.ObjectID, 1) {
		t.Fatal("磨刀石修復後應送出含耐久 tag=3 的 S_ItemStatus")
	}

	if !s.UseWhetstone(sess, player, stone, weapon.ObjectID) {
		t.Fatal("第二次磨刀石應成功處理")
	}
	packets = drainSkillTestPackets(sess)
	if weapon.Durability != 0 {
		t.Fatalf("第二次修復後耐久應歸零，got=%d", weapon.Durability)
	}
	if player.Inv.FindByObjectID(stone.ObjectID) != nil {
		t.Fatal("最後一個磨刀石應從背包移除")
	}
	if !hasServerMessage(packets, 464) {
		t.Fatal("完全修復時應送出 yiwei 訊息 464")
	}
	if !hasDurabilityStatusPacket(packets, weapon.ObjectID, 0) {
		t.Fatal("完全修復後應送出耐久歸零的 S_ItemStatus")
	}
}

func TestUseWhetstoneKeepsEquippedWeaponMarkerWhenSlotIsEquipped(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "hone",
		Inv:       world.NewInventory(),
		Str:       18,
		Con:       18,
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	weapon.Durability = 1
	player.Equip.Set(world.SlotWeapon, weapon)
	stone := player.Inv.AddItemWithID(102, 40317, 1, "whetstone", 0, 0, true, 1)
	ws.AddPlayer(player)

	s := NewItemUseSystem(&handler.Deps{World: ws, Items: items, Log: zap.NewNop()})
	if !s.UseWhetstone(sess, player, stone, weapon.ObjectID) {
		t.Fatal("磨刀石對已裝備武器應成功處理")
	}

	packets := drainSkillTestPackets(sess)
	if player.Equip.Weapon() != weapon {
		t.Fatal("磨刀石不得清除武器裝備槽")
	}
	if !weapon.Equipped {
		t.Fatal("裝備槽指向武器時，耐久更新前應同步 Equipped flag")
	}
	if !hasItemStatusNameContaining(packets, weapon.ObjectID, "($9)") {
		t.Fatal("已裝備武器的 S_ItemStatus 名稱應保留 ($9) 裝備中標記")
	}
}

func TestUseWhetstoneConsumesStoneOnInvalidTargetButNotMissingTarget(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "hone",
		Inv:       world.NewInventory(),
		Str:       18,
		Con:       18,
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	stone := player.Inv.AddItemWithID(102, 40317, 2, "whetstone", 0, 0, true, 1)
	ws.AddPlayer(player)

	s := NewItemUseSystem(&handler.Deps{World: ws, Items: items, Log: zap.NewNop()})
	if !s.UseWhetstone(sess, player, stone, weapon.ObjectID) {
		t.Fatal("目標存在但未損壞時仍應消耗磨刀石")
	}
	packets := drainSkillTestPackets(sess)
	if stone.Count != 1 {
		t.Fatalf("無效目標仍應消耗 1 個磨刀石，got=%d", stone.Count)
	}
	if !hasServerMessage(packets, 79) {
		t.Fatal("無效目標應送出訊息 79")
	}

	if s.UseWhetstone(sess, player, stone, 999999) {
		t.Fatal("目標不存在時 Java 直接返回，不應消耗磨刀石")
	}
	if stone.Count != 1 {
		t.Fatalf("目標不存在不得消耗磨刀石，got=%d", stone.Count)
	}
}

func TestDamageWeaponDurabilitySkipsNonHardNpc(t *testing.T) {
	items := loadDurabilityTestItems(t)
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "attacker",
		Inv:       world.NewInventory(),
	}
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	player.Equip.Set(world.SlotWeapon, weapon)
	npc := &world.NpcInfo{ID: 2001, Hard: false}

	damageWeaponDurability(player, npc, &handler.Deps{Items: items, Log: zap.NewNop()})

	if weapon.Durability != 0 {
		t.Fatalf("非 hard NPC 不應損壞武器，got=%d", weapon.Durability)
	}
	if packets := drainSkillTestPackets(sess); len(packets) != 0 {
		t.Fatalf("非 hard NPC 不應送耐久封包，got=%d", len(packets))
	}
}

func TestDamageWeaponDurabilitySkipsWeaponCanBeDamagedFalse(t *testing.T) {
	items := loadDurabilityCanBeDamagedFalseTestItems(t)
	sess := newSkillTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "attacker",
		Inv:       world.NewInventory(),
	}
	weapon := player.Inv.AddItemWithID(101, 2, 1, "protected sword", 0, 0, false, 1)
	player.Equip.Set(world.SlotWeapon, weapon)
	npc := &world.NpcInfo{ID: 2001, Hard: true}

	damageWeaponDurability(player, npc, &handler.Deps{Items: items, Log: zap.NewNop()})

	if weapon.Durability != 0 {
		t.Fatalf("can_be_damaged=false 的武器不應損壞，got=%d", weapon.Durability)
	}
	if packets := drainSkillTestPackets(sess); len(packets) != 0 {
		t.Fatalf("不可損壞武器不應送耐久封包，got=%d", len(packets))
	}
}

func TestDamageEquippedWeaponDurabilityBroadcastsOnlySameShowLikeJava(t *testing.T) {
	items := loadDurabilityTestItems(t)
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "same_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    100,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    200,
	})
	weapon := player.Inv.AddItemWithID(101, 1, 1, "sword", 0, 0, false, 1)
	player.Equip.Set(world.SlotWeapon, weapon)

	deps := &handler.Deps{
		World: ws,
		Items: items,
		Log:   zap.NewNop(),
	}
	for i := 0; i < 1000 && weapon.Durability == 0; i++ {
		damageEquippedWeaponDurability(player, deps, false)
	}
	if weapon.Durability == 0 {
		t.Fatalf("測試前置失敗：1000 次耐久判定都未觸發損壞")
	}

	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), player.CharID, 10712) {
		t.Fatalf("同 ShowID 觀眾應收到武器損壞特效")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), player.CharID, 10712) {
		t.Fatalf("不同 ShowID 觀眾不應收到武器損壞特效")
	}
}

func hasServerMessage(packets [][]byte, msgID uint16) bool {
	for _, pkt := range packets {
		if len(pkt) >= 3 && pkt[0] == packet.S_OPCODE_MESSAGE_CODE &&
			binary.LittleEndian.Uint16(pkt[1:3]) == msgID {
			return true
		}
	}
	return false
}

func hasDurabilityStatusPacket(packets [][]byte, objectID int32, durability int8) bool {
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_CHANGE_ITEM_USE {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) != objectID {
			continue
		}
		off := 5
		for off < len(pkt) && pkt[off] != 0 {
			off++
		}
		if off >= len(pkt) {
			continue
		}
		off++
		if off+5 > len(pkt) {
			continue
		}
		off += 4
		statusLen := int(pkt[off])
		off++
		if off+statusLen > len(pkt) {
			continue
		}
		status := pkt[off : off+statusLen]
		for i := 0; i < len(status); i++ {
			if status[i] != 3 {
				continue
			}
			if durability == 0 {
				return true
			}
			if i+1 < len(status) && int8(status[i+1]) == durability {
				return true
			}
		}
		if durability == 0 && statusLen > 0 {
			return true
		}
	}
	return false
}

func hasItemStatusNameContaining(packets [][]byte, objectID int32, want string) bool {
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_CHANGE_ITEM_USE {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) != objectID {
			continue
		}
		off := 5
		for off < len(pkt) && pkt[off] != 0 {
			off++
		}
		if off >= len(pkt) {
			continue
		}
		name := string(pkt[5:off])
		if strings.Contains(name, want) {
			return true
		}
	}
	return false
}

func loadDurabilityTestItems(t *testing.T) *data.ItemTable {
	t.Helper()
	dir := t.TempDir()
	weaponPath := writeDurabilityTestYAML(t, dir, "weapon_list.yaml", `
weapons:
  - item_id: 1
    name: sword
    type: sword
    can_be_damaged: true
`)
	armorPath := writeDurabilityTestYAML(t, dir, "armor_list.yaml", `
armors:
  - item_id: 2
    name: armor
    type: armor
`)
	etcPath := writeDurabilityTestYAML(t, dir, "etcitem_list.yaml", `
items:
  - item_id: 40317
    name: whetstone
    item_type: other
    use_type: choice
    stackable: true
  - item_id: 40308
    name: adena
    item_type: other
    use_type: none
    stackable: true
`)
	table, err := data.LoadItemTable(weaponPath, armorPath, etcPath)
	if err != nil {
		t.Fatalf("載入測試物品失敗: %v", err)
	}
	return table
}

func loadDurabilityCanBeDamagedFalseTestItems(t *testing.T) *data.ItemTable {
	t.Helper()
	dir := t.TempDir()
	weaponPath := writeDurabilityTestYAML(t, dir, "weapon_list.yaml", `
weapons:
  - item_id: 2
    name: protected sword
    type: sword
    can_be_damaged: false
`)
	armorPath := writeDurabilityTestYAML(t, dir, "armor_list.yaml", "armors: []\n")
	etcPath := writeDurabilityTestYAML(t, dir, "etcitem_list.yaml", "items: []\n")
	table, err := data.LoadItemTable(weaponPath, armorPath, etcPath)
	if err != nil {
		t.Fatalf("載入測試物品失敗: %v", err)
	}
	return table
}

func writeDurabilityTestYAML(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}
	return path
}
