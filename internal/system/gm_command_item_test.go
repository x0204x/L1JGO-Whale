package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

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

func newGMCommandItemTestSystem(t *testing.T) *GMCommandSystem {
	t.Helper()
	items, err := data.LoadItemTable("../../data/yaml/weapon_list.yaml", "../../data/yaml/armor_list.yaml", "../../data/yaml/etcitem_list.yaml")
	if err != nil {
		t.Fatalf("載入物品表失敗: %v", err)
	}
	return NewGMCommandSystem(&handler.Deps{
		Items: items,
		Log:   zap.NewNop(),
	})
}
