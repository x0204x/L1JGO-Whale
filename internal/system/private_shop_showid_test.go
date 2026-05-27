package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestPrivateShopSetupBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	shopOwner, sameShow, otherShow := addPrivateShopShowIDPlayers(t, ws)
	sys := NewPrivateShopSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

	sys.SetupShop(shopOwner, nil, nil, []byte("tradezone0"))

	ownerPackets := drainSkillTestPackets(shopOwner.Session)
	if !hasActionGfxPacket(ownerPackets, shopOwner.CharID, 3) {
		t.Fatal("yiwei sendPacketsAll 會把開店前取消動作送給商店玩家自己")
	}
	if !hasShopActionPacket(ownerPackets, shopOwner.CharID) {
		t.Fatal("yiwei sendPacketsAll 會把個人商店動作送給商店玩家自己")
	}
	samePackets := drainSkillTestPackets(sameShow.Session)
	if !hasActionGfxPacket(samePackets, shopOwner.CharID, 3) {
		t.Fatal("同 ShowID 玩家應收到開店前取消動作")
	}
	if !hasShopActionPacket(samePackets, shopOwner.CharID) {
		t.Fatal("同 ShowID 玩家應收到個人商店動作")
	}
	otherPackets := drainSkillTestPackets(otherShow.Session)
	if hasActionGfxPacket(otherPackets, shopOwner.CharID, 3) || hasShopActionPacket(otherPackets, shopOwner.CharID) {
		t.Fatal("不同 ShowID 玩家不應收到個人商店開店動作")
	}
}

func TestPrivateShopCloseAndCancelBroadcastOnlySameShowLikeJava(t *testing.T) {
	cases := []struct {
		name string
		run  func(*PrivateShopSystem, *world.PlayerInfo)
	}{
		{name: "close", run: func(sys *PrivateShopSystem, p *world.PlayerInfo) { sys.CloseShop(p) }},
		{name: "not_tradable_cancel", run: func(sys *PrivateShopSystem, p *world.PlayerInfo) { sys.CancelShopNotTradable(p) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := world.NewState()
			shopOwner, sameShow, otherShow := addPrivateShopShowIDPlayers(t, ws)
			sys := NewPrivateShopSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

			tc.run(sys, shopOwner)

			if !hasActionGfxPacket(drainSkillTestPackets(shopOwner.Session), shopOwner.CharID, 3) {
				t.Fatalf("yiwei sendPacketsAll 會把 %s 動作送給商店玩家自己", tc.name)
			}
			if !hasActionGfxPacket(drainSkillTestPackets(sameShow.Session), shopOwner.CharID, 3) {
				t.Fatalf("同 ShowID 玩家應收到 %s 個人商店取消動作", tc.name)
			}
			if hasActionGfxPacket(drainSkillTestPackets(otherShow.Session), shopOwner.CharID, 3) {
				t.Fatalf("不同 ShowID 玩家不應收到 %s 個人商店取消動作", tc.name)
			}
		})
	}
}

func addPrivateShopShowIDPlayers(t *testing.T, ws *world.State) (*world.PlayerInfo, *world.PlayerInfo, *world.PlayerInfo) {
	t.Helper()
	shopOwner := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    8101,
		Name:      "shop_owner",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    8102,
		Name:      "same_show_shop_viewer",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    8103,
		Name:      "other_show_shop_viewer",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	return shopOwner, sameShow, otherShow
}

func hasShopActionPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ACTION {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID && pkt[5] == 70 {
			return true
		}
	}
	return false
}
