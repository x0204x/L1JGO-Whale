package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestPowerItemBuyUsesItemCreateWithOptions(t *testing.T) {
	ws := world.NewState()
	sess := newSkillTestSession(t, 1)
	itemCreate := &boxItemCreateStub{}
	sys := NewPowerItemSystem(&handler.Deps{
		World:      ws,
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   sess,
		CharID:    1001,
		Name:      "power-item-test",
		Inv:       world.NewInventory(),
	})
	player.Inv.AddItem(world.AdenaItemID, 1000, "金幣", 5, 0, true, 1)
	pItem := &data.PowerShopItem{
		ItemID:     1002,
		Price:      50,
		EnchantLvl: 7,
		Bless:      2,
		AttrKind:   1,
		AttrLevel:  3,
	}

	sys.BuyPowerItem(sess, player, pItem)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤: got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1002 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤: got item=%d count=%d want item=1002 count=1", itemCreate.itemID, itemCreate.count)
	}
	if itemCreate.opts.EnchantLvl != 7 || !itemCreate.opts.BlessSet || itemCreate.opts.Bless != 2 {
		t.Fatalf("ItemCreate 選項錯誤: enchant=%d blessSet=%t bless=%d",
			itemCreate.opts.EnchantLvl, itemCreate.opts.BlessSet, itemCreate.opts.Bless)
	}
	if itemCreate.item == nil || itemCreate.item.AttrEnchantKind != 1 || itemCreate.item.AttrEnchantLevel != 3 {
		t.Fatalf("屬性強化欄位錯誤: item=%+v", itemCreate.item)
	}
	if got := player.Inv.GetAdena(); got != 950 {
		t.Fatalf("金幣扣除錯誤: got %d want 950", got)
	}
}
