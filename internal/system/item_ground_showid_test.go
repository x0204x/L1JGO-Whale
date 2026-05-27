package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestItemGroundDropBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 21),
		SessionID: 21,
		CharID:    2101,
		Name:      "dropper",
		X:         10,
		Y:         10,
		MapID:     4,
		ShowID:    77,
		Str:       20,
		Con:       20,
		Inv:       world.NewInventory(),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 22),
		SessionID: 22,
		CharID:    2102,
		Name:      "same_show",
		X:         11,
		Y:         10,
		MapID:     4,
		ShowID:    77,
		Inv:       world.NewInventory(),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 23),
		SessionID: 23,
		CharID:    2103,
		Name:      "other_show",
		X:         11,
		Y:         10,
		MapID:     4,
		ShowID:    88,
		Inv:       world.NewInventory(),
	})

	item := player.Inv.AddItem(1001, 1, "test potion", 101, 10, true, 1)
	sys := NewItemGroundSystem(&handler.Deps{
		World: ws,
		Items: testItemTable(t),
		Log:   zap.NewNop(),
	})

	sys.DropItem(player.Session, player, item.ObjectID, 1)

	ground := ws.GetNearbyGroundItems(player.X, player.Y, player.MapID)
	if len(ground) != 1 {
		t.Fatalf("應建立一個地上物，got=%d", len(ground))
	}
	if ground[0].ShowID != player.ShowID {
		t.Fatalf("Yiwei 掉落物應繼承玩家 ShowID，got=%d want=%d", ground[0].ShowID, player.ShowID)
	}
	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), ground[0].ID) {
		t.Fatalf("同 ShowID 玩家應收到掉落物顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), ground[0].ID) {
		t.Fatalf("不同 ShowID 玩家不應收到掉落物顯示封包")
	}
}

func TestGroundItemExpireBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 25),
		SessionID: 25,
		CharID:    2301,
		Name:      "same_show_expire_viewer",
		X:         10,
		Y:         10,
		MapID:     4,
		ShowID:    77,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 26),
		SessionID: 26,
		CharID:    2302,
		Name:      "other_show_expire_viewer",
		X:         10,
		Y:         10,
		MapID:     4,
		ShowID:    88,
	})
	ground := &world.GroundItem{
		ID:     2303,
		ItemID: 1001,
		Count:  1,
		Name:   "expired_item",
		GrdGfx: 201,
		X:      10,
		Y:      10,
		MapID:  4,
		ShowID: 77,
		TTL:    1,
	}
	ws.AddGroundItem(ground)

	NewGroundItemSystem(ws).Update(0)

	if !hasRemoveObjectPacket(drainSkillTestPackets(sameShow.Session), ground.ID) {
		t.Fatalf("同 ShowID 玩家應收到地上物到期移除封包")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(otherShow.Session), ground.ID) {
		t.Fatalf("不同 ShowID 玩家不應收到地上物到期移除封包")
	}
}

func TestItemGroundPickupRejectsDifferentShowLikeJava(t *testing.T) {
	ws := world.NewState()
	picker := addSkillTestPlayer(ws, &world.PlayerInfo{
		Session:   newSkillTestSession(t, 24),
		SessionID: 24,
		CharID:    2201,
		Name:      "picker",
		X:         10,
		Y:         10,
		MapID:     4,
		ShowID:    88,
		Str:       20,
		Con:       20,
		Inv:       world.NewInventory(),
	})
	ground := &world.GroundItem{
		ID:     2202,
		ItemID: 1001,
		Count:  1,
		Name:   "test potion",
		GrdGfx: 201,
		X:      10,
		Y:      10,
		MapID:  4,
		ShowID: 77,
	}
	ws.AddGroundItem(ground)
	sys := NewItemGroundSystem(&handler.Deps{
		World: ws,
		Items: testItemTable(t),
		Log:   zap.NewNop(),
	})

	sys.PickupItem(picker.Session, picker, ground.ID)

	if ws.GetGroundItem(ground.ID) == nil {
		t.Fatalf("不同 ShowID 玩家不應撿走地上物")
	}
	if picked := picker.Inv.FindByItemID(ground.ItemID); picked != nil {
		t.Fatalf("不同 ShowID 玩家不應取得地上物，got=%+v", picked)
	}
}
