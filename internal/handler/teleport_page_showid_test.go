package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestTeleportPageDepartureEffectBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	sameSess := newHandlerTestSession(t, 2)
	otherSess := newHandlerTestSession(t, 3)

	player := &world.PlayerInfo{
		SessionID:    sess.ID,
		Session:      sess,
		CharID:       7301,
		Name:         "tele_page",
		X:            100,
		Y:            100,
		MapID:        900,
		ShowID:       70,
		TeleCategory: "A",
		TelePage:     1,
		TeleNpcObjID: 33001,
	}
	sameShow := &world.PlayerInfo{
		SessionID: sameSess.ID,
		Session:   sameSess,
		CharID:    7302,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    70,
	}
	otherShow := &world.PlayerInfo{
		SessionID: otherSess.ID,
		Session:   otherSess,
		CharID:    7303,
		Name:      "other",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    99,
	}
	ws.AddPlayer(player)
	ws.AddPlayer(sameShow)
	ws.AddPlayer(otherShow)

	executeTeleportPage(sess, player, &data.TeleportPageDest{
		Name:  "target",
		X:     33000,
		Y:     33001,
		MapID: 900,
	}, &Deps{World: ws, Log: zap.NewNop()})

	if !hasTeleportPageSkillEffectPacket(drainHandlerTestPackets(sess), player.CharID, 169) {
		t.Fatal("yiwei sendPacketsAll 會讓傳送本人收到 169 離場特效")
	}
	if !hasTeleportPageSkillEffectPacket(drainHandlerTestPackets(sameSess), player.CharID, 169) {
		t.Fatal("同 ShowID 觀察者應收到傳送頁 169 離場特效")
	}
	if hasTeleportPageSkillEffectPacket(drainHandlerTestPackets(otherSess), player.CharID, 169) {
		t.Fatal("不同 ShowID 觀察者不應收到傳送頁 169 離場特效")
	}
	if player.ScrollTPTick != 2 || player.ScrollTPX != 33000 || player.ScrollTPY != 33001 || player.ScrollTPMap != 900 {
		t.Fatalf("傳送頁應保留 2 tick 延遲傳送狀態，got tick=%d x=%d y=%d map=%d",
			player.ScrollTPTick, player.ScrollTPX, player.ScrollTPY, player.ScrollTPMap)
	}
	if player.TeleCategory != "" || player.TelePage != 0 || player.TeleNpcObjID != 0 {
		t.Fatalf("傳送頁執行後應清除對話狀態，got category=%q page=%d npc=%d",
			player.TeleCategory, player.TelePage, player.TeleNpcObjID)
	}
}

func hasTeleportPageSkillEffectPacket(packets [][]byte, objectID int32, gfxID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_EFFECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID &&
			int32(binary.LittleEndian.Uint16(pkt[5:7])) == gfxID {
			return true
		}
	}
	return false
}
