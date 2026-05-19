package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type fishingNpcServiceStub struct {
	consumedObjectID int32
	consumedCount    int32
}

func (s *fishingNpcServiceStub) NpcFullHeal(*net.Session, *world.PlayerInfo, int32) {}
func (s *fishingNpcServiceStub) NpcWeaponEnchant(*net.Session, *world.PlayerInfo)   {}
func (s *fishingNpcServiceStub) NpcArmorEnchant(*net.Session, *world.PlayerInfo)    {}
func (s *fishingNpcServiceStub) NpcPoly(*net.Session, *world.PlayerInfo, int32)     {}
func (s *fishingNpcServiceStub) NpcTeleportWithCost(*net.Session, *world.PlayerInfo, *data.TeleportDest, int32) {
}
func (s *fishingNpcServiceStub) NpcUpgrade(*net.Session, *world.PlayerInfo, *data.ItemUpgrade) {}
func (s *fishingNpcServiceStub) ConsumeAdena(*net.Session, *world.PlayerInfo, int32) bool {
	return true
}
func (s *fishingNpcServiceStub) RepairWeapon(*net.Session, *world.PlayerInfo, *world.InvItem, int32) bool {
	return true
}
func (s *fishingNpcServiceStub) ConsumeItem(_ *net.Session, _ *world.PlayerInfo, objectID int32, count int32) bool {
	s.consumedObjectID = objectID
	s.consumedCount = count
	return true
}
func (s *fishingNpcServiceStub) Refine(*net.Session, *world.PlayerInfo, *world.InvItem, int32, int32) {
}
func (s *fishingNpcServiceStub) FireSmithCraft(*net.Session, *world.PlayerInfo, *data.FireSmithRecipe, int32, int32) {
}

type fishingItemCreateStub struct {
	itemID int32
	count  int32
}

func (s *fishingItemCreateStub) GiveItem(_ *net.Session, _ *world.PlayerInfo, itemID, count int32) (*world.InvItem, bool) {
	s.itemID = itemID
	s.count = count
	return &world.InvItem{ItemID: itemID, Count: count}, true
}

func TestFishingTickCreatesRewardItem(t *testing.T) {
	player := &world.PlayerInfo{
		Name:          "釣魚玩家",
		Inv:           world.NewInventory(),
		Fishing:       true,
		FishingTick:   fishingInterval - 1,
		FishingPoleID: fishingPoleA,
	}
	bait := player.Inv.AddItem(fishingBaitID, 3, "營養釣餌", 0, 1, true, 0)
	npcSvc := &fishingNpcServiceStub{}
	itemCreate := &fishingItemCreateStub{}
	sys := NewFishingSystem(&handler.Deps{
		NpcSvc:     npcSvc,
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})

	sys.Tick(player)

	if npcSvc.consumedObjectID != bait.ObjectID || npcSvc.consumedCount != 1 {
		t.Fatalf("釣魚應消耗餌料 object=%d count=1，got object=%d count=%d",
			bait.ObjectID, npcSvc.consumedObjectID, npcSvc.consumedCount)
	}
	if itemCreate.itemID == 0 || itemCreate.count != 1 {
		t.Fatalf("釣魚應給 1 個魚類獎勵，got item=%d count=%d", itemCreate.itemID, itemCreate.count)
	}
}
