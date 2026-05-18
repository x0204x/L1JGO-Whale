package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type boxItemCreateStub struct {
	itemID int32
	count  int32
	opts   ItemCreateOptions
	item   *world.InvItem
	calls  int
}

func (s *boxItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.calls++
	return &world.InvItem{ItemID: itemID, Count: count}, true
}

func (s *boxItemCreateStub) GiveItemWithOptions(_ *net.Session, _ *world.PlayerInfo, itemID, count int32, opts ItemCreateOptions) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	s.opts = opts
	s.calls++
	item := &world.InvItem{ItemID: itemID, Count: count, EnchantLvl: opts.EnchantLvl, Bless: opts.Bless}
	if opts.BeforeSend != nil {
		opts.BeforeSend(item)
	}
	s.item = item
	return item, true
}

func TestGiveBoxRewardUsesItemCreateWithOptions(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "box-test",
	})
	itemCreate := &boxItemCreateStub{}
	sys := NewItemUseSystem(&handler.Deps{
		World:      ws,
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.GiveBoxReward(player.Session, player, 1002, 1, 1, 2, 7, false)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤: got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1002 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤: got item=%d count=%d want item=1002 count=1", itemCreate.itemID, itemCreate.count)
	}
	if itemCreate.opts.EnchantLvl != 7 || !itemCreate.opts.BlessSet || itemCreate.opts.Bless != 2 {
		t.Fatalf("ItemCreate 選項錯誤: got enchant=%d blessSet=%t bless=%d want enchant=7 blessSet=true bless=2",
			itemCreate.opts.EnchantLvl, itemCreate.opts.BlessSet, itemCreate.opts.Bless)
	}
}
