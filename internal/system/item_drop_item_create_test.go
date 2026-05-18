package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type dropItemCreateStub struct {
	itemID      int32
	count       int32
	opts        ItemCreateOptions
	item        *world.InvItem
	calls       int
	optionCalls int
}

func (s *dropItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.calls++
	item := &world.InvItem{ItemID: itemID, Count: count, Identified: true}
	s.item = item
	return item, true
}

func (s *dropItemCreateStub) GiveItemWithOptions(_ *net.Session, _ *world.PlayerInfo, itemID, count int32, opts ItemCreateOptions) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.opts = opts
	s.calls++
	s.optionCalls++
	item := &world.InvItem{ItemID: itemID, Count: count, Identified: true}
	if opts.BeforeSend != nil {
		opts.BeforeSend(item)
	}
	s.item = item
	return item, true
}

func TestGiveDropToPlayerUsesItemCreateForAdena(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "drop-test",
	})
	itemCreate := &dropItemCreateStub{}
	sys := NewItemUseSystem(&handler.Deps{
		World:      ws,
		Items:      loadDropItemCreateItems(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.giveDropToPlayer(player, data.DropItem{ItemID: world.AdenaItemID}, 123)

	if itemCreate.calls != 1 || itemCreate.optionCalls != 0 {
		t.Fatalf("金幣掉落應走基本 ItemCreate 一次: calls=%d optionCalls=%d", itemCreate.calls, itemCreate.optionCalls)
	}
	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 123 {
		t.Fatalf("ItemCreate 參數錯誤: item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
	_ = drainSkillTestPackets(player.Session)
}

func TestGiveDropToPlayerUsesItemCreateOptionsForEquipment(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "drop-test",
	})
	itemCreate := &dropItemCreateStub{}
	sys := NewItemUseSystem(&handler.Deps{
		World:      ws,
		Items:      loadDropItemCreateItems(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.giveDropToPlayer(player, data.DropItem{ItemID: 1, EnchantLevel: 3}, 1)

	if itemCreate.calls != 1 || itemCreate.optionCalls != 1 {
		t.Fatalf("裝備掉落應走 ItemCreateWithOptions 一次: calls=%d optionCalls=%d", itemCreate.calls, itemCreate.optionCalls)
	}
	if itemCreate.itemID != 1 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤: item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
	if itemCreate.opts.EnchantLvl != 3 {
		t.Fatalf("掉落強化值未傳入 ItemCreate: got %d want 3", itemCreate.opts.EnchantLvl)
	}
	if itemCreate.opts.BeforeSend == nil || itemCreate.item == nil || itemCreate.item.Identified {
		t.Fatalf("裝備掉落應在送封包前設為未鑑定: opts=%+v item=%+v", itemCreate.opts, itemCreate.item)
	}
	_ = drainSkillTestPackets(player.Session)
}

func loadDropItemCreateItems(t *testing.T) *data.ItemTable {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品資料失敗: %v", err)
	}
	return items
}
