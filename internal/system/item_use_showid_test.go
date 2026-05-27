package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestItemUseHasteBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player, sameShow, otherShow := addItemUseShowIDPlayers(t, ws)
	sys := NewItemUseSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

	sys.ApplyHaste(player.Session, player, 120, 191)

	samePackets := drainSkillTestPackets(sameShow.Session)
	if !hasSpeedPacketForObject(samePackets, player.CharID) {
		t.Fatal("同 ShowID 玩家應收到加速封包")
	}
	if !hasSkillEffectPacket(samePackets, player.CharID, 191) {
		t.Fatal("同 ShowID 玩家應收到加速藥水特效")
	}
	otherPackets := drainSkillTestPackets(otherShow.Session)
	if hasSpeedPacketForObject(otherPackets, player.CharID) || hasSkillEffectPacket(otherPackets, player.CharID, 191) {
		t.Fatal("yiwei sendPacketsAll 會以 showId 隔離，不同 ShowID 玩家不應收到加速封包或特效")
	}
}

func TestItemUseFixedTeleportScrollEffectBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player, sameShow, otherShow := addItemUseShowIDPlayers(t, ws)
	scroll := player.Inv.AddItemWithID(9101, 40101, 1, "指定傳送卷軸", 0, 1000, true, 1)
	sys := NewItemUseSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

	sys.UseFixedTeleportScroll(player.Session, player, scroll, &data.ItemInfo{
		ItemID:   40101,
		Name:     "指定傳送卷軸",
		LocX:     200,
		LocY:     201,
		LocMapID: 4,
	})

	if !hasSkillEffectPacket(drainSkillTestPackets(sameShow.Session), player.CharID, 169) {
		t.Fatal("同 ShowID 玩家應收到傳送卷軸出發特效")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(otherShow.Session), player.CharID, 169) {
		t.Fatal("yiwei 玩家視覺封包會以 showId 隔離，不同 ShowID 玩家不應收到傳送卷軸出發特效")
	}
}

func addItemUseShowIDPlayers(t *testing.T, ws *world.State) (*world.PlayerInfo, *world.PlayerInfo, *world.PlayerInfo) {
	t.Helper()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9101,
		Name:      "item_user",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9102,
		Name:      "same_show_item_viewer",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9103,
		Name:      "other_show_item_viewer",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
	})
	return player, sameShow, otherShow
}

func hasSpeedPacketForObject(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_SPEED {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}
