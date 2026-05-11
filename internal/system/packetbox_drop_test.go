package system

import (
	"bytes"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestGiveDropToPlayerSendsShowDropForAdena(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "tester",
	})
	sys := newPacketBoxDropItemUseSystem(t, ws)

	sys.giveDropToPlayer(player, data.DropItem{ItemID: world.AdenaItemID}, 123)

	packets := drainSkillTestPackets(player.Session)
	want := handler.BuildShowDrop(handler.ShowDropAdena, 123)
	if !hasExactPacket(packets, want) {
		t.Fatalf("金幣掉落應發送 S_ShowDrop ADENA，packets=%v want=%v", packets, want)
	}
}

func TestGiveDropToPlayerSendsItemBoardForNormalItem(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "tester",
	})
	sys := newPacketBoxDropItemUseSystem(t, ws)
	itemInfo := sys.deps.Items.Get(40010)
	if itemInfo == nil {
		t.Fatal("測試資料缺少 item 40010")
	}

	sys.giveDropToPlayer(player, data.DropItem{ItemID: 40010}, 2)

	packets := drainSkillTestPackets(player.Session)
	want := handler.BuildItemBoard(uint16(itemInfo.InvGfx), "獲得 "+itemInfo.Name+" (2)")
	if !hasExactPacket(packets, want) {
		t.Fatalf("一般掉落應發送 S_ItemBoard，packets=%v want=%v", packets, want)
	}
}

func TestAddExpSendsShowDropForExpGain(t *testing.T) {
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "tester",
		Level:     1,
		HP:        100,
		MaxHP:     100,
		MP:        20,
		MaxMP:     20,
		Inv:       world.NewInventory(),
	}
	deps := &handler.Deps{Scripting: engine, Log: zap.NewNop()}

	addExp(player, 77, deps)

	packets := drainSkillTestPackets(player.Session)
	want := handler.BuildShowDrop(handler.ShowDropExp, 77)
	if !hasExactPacket(packets, want) {
		t.Fatalf("獲得經驗應發送 S_ShowDrop EXP，packets=%v want=%v", packets, want)
	}
}

func newPacketBoxDropItemUseSystem(t *testing.T, ws *world.State) *ItemUseSystem {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入道具資料失敗: %v", err)
	}
	return NewItemUseSystem(&handler.Deps{
		World: ws,
		Items: items,
		Log:   zap.NewNop(),
	})
}

func hasExactPacket(packets [][]byte, want []byte) bool {
	for _, pkt := range packets {
		if bytes.Equal(pkt, want) {
			return true
		}
	}
	return false
}
