package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type gmItemCreateStub struct {
	itemID      int32
	count       int32
	opts        ItemCreateOptions
	calls       int
	optionCalls int
}

func (s *gmItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count += count
	s.calls++
	return &world.InvItem{ItemID: itemID, Count: count}, true
}

func (s *gmItemCreateStub) GiveItemWithOptions(_ *net.Session, _ *world.PlayerInfo, itemID, count int32, opts ItemCreateOptions) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count += count
	s.opts = opts
	s.calls++
	s.optionCalls++
	return &world.InvItem{ItemID: itemID, Count: count, EnchantLvl: opts.EnchantLvl}, true
}

func TestGMCommandGiveItemCreatesSeparateEnchantedWeapons(t *testing.T) {
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "gm",
		Inv:       world.NewInventory(),
	}
	sys := newGMCommandItemTestSystem(t)

	sys.GiveItem(player.Session, player, 31, 3, 7)

	if len(player.Inv.Items) != 3 {
		t.Fatalf("不可堆疊武器 count=3 應建立 3 格，got=%d", len(player.Inv.Items))
	}
	for _, item := range player.Inv.Items {
		if item.ItemID != 31 || item.Count != 1 || item.EnchantLvl != 7 {
			t.Fatalf("武器實體錯誤: itemID=%d count=%d enchant=%d", item.ItemID, item.Count, item.EnchantLvl)
		}
	}
}

func TestGMCommandGiveItemStackableIgnoresEnchant(t *testing.T) {
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "gm",
		Inv:       world.NewInventory(),
	}
	sys := newGMCommandItemTestSystem(t)

	sys.GiveItem(player.Session, player, 40010, 5, 7)

	if len(player.Inv.Items) != 1 {
		t.Fatalf("可堆疊道具應建立 1 格，got=%d", len(player.Inv.Items))
	}
	item := player.Inv.Items[0]
	if item.ItemID != 40010 || item.Count != 5 || item.EnchantLvl != 0 {
		t.Fatalf("可堆疊道具錯誤: itemID=%d count=%d enchant=%d", item.ItemID, item.Count, item.EnchantLvl)
	}
}

func TestGMCommandGiveItemUsesItemCreateForStackable(t *testing.T) {
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "gm",
		Inv:       world.NewInventory(),
	}
	itemCreate := &gmItemCreateStub{}
	sys := newGMCommandItemTestSystemWithItemCreate(t, itemCreate)

	sys.GiveItem(player.Session, player, 40010, 5, 7)

	if itemCreate.calls != 1 || itemCreate.optionCalls != 0 {
		t.Fatalf("可堆疊 GM 給物品應走基本 ItemCreate 一次: calls=%d optionCalls=%d", itemCreate.calls, itemCreate.optionCalls)
	}
	if itemCreate.itemID != 40010 || itemCreate.count != 5 {
		t.Fatalf("ItemCreate 參數錯誤: item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
	if !player.Dirty {
		t.Fatal("GM 給物品成功後應標記玩家 Dirty")
	}
}

func TestGMCommandGiveItemUsesItemCreateOptionsForEquipment(t *testing.T) {
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "gm",
		Inv:       world.NewInventory(),
	}
	itemCreate := &gmItemCreateStub{}
	sys := newGMCommandItemTestSystemWithItemCreate(t, itemCreate)

	sys.GiveItem(player.Session, player, 31, 3, 7)

	if itemCreate.calls != 3 || itemCreate.optionCalls != 3 {
		t.Fatalf("不可堆疊 GM 給物品應逐件走 ItemCreateWithOptions: calls=%d optionCalls=%d", itemCreate.calls, itemCreate.optionCalls)
	}
	if itemCreate.itemID != 31 || itemCreate.count != 3 {
		t.Fatalf("ItemCreate 參數錯誤: item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
	if itemCreate.opts.EnchantLvl != 7 {
		t.Fatalf("GM 強化值未傳入 ItemCreate: got %d want 7", itemCreate.opts.EnchantLvl)
	}
	if !player.Dirty {
		t.Fatal("GM 給物品成功後應標記玩家 Dirty")
	}
}

func TestGMCommandGiveGoldUsesItemCreate(t *testing.T) {
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "gm",
		Inv:       world.NewInventory(),
	}
	itemCreate := &gmItemCreateStub{}
	sys := newGMCommandItemTestSystemWithItemCreate(t, itemCreate)

	sys.GiveGold(player.Session, player, 1234)

	if itemCreate.calls != 1 || itemCreate.itemID != world.AdenaItemID || itemCreate.count != 1234 {
		t.Fatalf("GM 金幣應走 ItemCreate: calls=%d item=%d count=%d", itemCreate.calls, itemCreate.itemID, itemCreate.count)
	}
}

func newGMCommandItemTestSystem(t *testing.T) *GMCommandSystem {
	t.Helper()
	return newGMCommandItemTestSystemWithItemCreate(t, nil)
}

func newGMCommandItemTestSystemWithItemCreate(t *testing.T, itemCreate handler.ItemCreateManager) *GMCommandSystem {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	return NewGMCommandSystem(&handler.Deps{
		Items:      items,
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
}
