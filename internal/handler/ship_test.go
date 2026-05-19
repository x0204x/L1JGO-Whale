package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// MISS-P1-007: 對齊 Java `C_Ship`：消耗船票 → 送 S_OwnCharPack → L1Teleport.teleport(..., heading=0)。
// 重點驗證兩個 Java 行為差異：
//  1) 傳送前主動送一次 S_OwnCharPack（讓客戶端拿到 consumeItem 之後的狀態快照）。
//  2) 傳送 heading 固定 0（不是先前 Go 硬寫的 5）。
func TestHandleEnterShipMatchesJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    9001,
		Name:      "sailor",
		MapID:     5, // 說話之島船地圖
		X:         32733,
		Y:         32796,
		Heading:   3,
		Inv:       world.NewInventory(),
	}
	// 船票 40299（Java C_Ship 對 map=5 對應 40299）
	ticket := player.Inv.AddItem(40299, 1, "ticket", 0, 0, true, 0)
	ws.AddPlayer(player)

	stub := &cookingNpcServiceStub{}
	deps := &Deps{
		World:  ws,
		NpcSvc: stub,
		Config: &config.Config{},
		Log:    zap.NewNop(),
	}

	// 封包欄位順序對齊 Java：[H destMapId][H destX][H destY]
	body := []byte{
		0x04, 0x00, // destMapId = 4（古魯丁主大陸）
		0xC0, 0x7F, // destX = 32704
		0xC0, 0x7F, // destY = 32704
	}
	// packet.NewReader 內部會自動跳過 opcode（off=1），這裡只要在前面塞 opcode 即可。
	r := packet.NewReader(append([]byte{packet.C_OPCODE_ENTER_SHIP}, body...))

	HandleEnterShip(sess, r, deps)

	// 船票消耗：Java consumeItem(40299, 1)
	if stub.consumedObjectID != ticket.ObjectID {
		t.Fatalf("應消耗 ticket.ObjectID=%d，實際=%d", ticket.ObjectID, stub.consumedObjectID)
	}
	if stub.consumedCount != 1 {
		t.Fatalf("應消耗 1 張船票，實際=%d", stub.consumedCount)
	}

	// 傳送後玩家狀態：Java L1Teleport.teleport(..., heading=0)
	updated := ws.GetBySession(sess.ID)
	if updated == nil {
		t.Fatalf("玩家應仍存在於 world")
	}
	if updated.Heading != 0 {
		t.Fatalf("Java C_Ship 固定 heading=0，實際=%d", updated.Heading)
	}
	if updated.MapID != 4 {
		t.Fatalf("應傳送到 destMapID=4，實際=%d", updated.MapID)
	}
	if updated.X != 32704 || updated.Y != 32704 {
		t.Fatalf("應傳送到 (32704,32704)，實際=(%d,%d)", updated.X, updated.Y)
	}

	// 傳送前必須送 S_OwnCharPack（S_PUT_OBJECT, opcode 87）給玩家。
	// teleportPlayer 之後還會再送一次 — 所以總共至少 2 次 S_PUT_OBJECT 出現在玩家自己的 session 中。
	pkts := drainHandlerTestPackets(sess)
	ownCharPackCount := 0
	for _, p := range pkts {
		if len(p) > 0 && p[0] == packet.S_OPCODE_PUT_OBJECT {
			ownCharPackCount++
		}
	}
	if ownCharPackCount < 2 {
		t.Fatalf("Java C_Ship 在傳送前後各送一次 S_OwnCharPack，實際 S_PUT_OBJECT 封包數=%d / 總封包數=%d", ownCharPackCount, len(pkts))
	}
}

// MISS-P1-007: 未持船票時 Java 直接走 switch default（item_id=0）不消耗、不傳送、不送封包。
// Go 端在 ticket==nil 時靜默 return，玩家應留在原地。
func TestHandleEnterShipWithoutTicketSilentlyReturns(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 2)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    9002,
		Name:      "stowaway",
		MapID:     5,
		X:         32733,
		Y:         32796,
		Heading:   3,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	stub := &cookingNpcServiceStub{}
	deps := &Deps{World: ws, NpcSvc: stub, Config: &config.Config{}, Log: zap.NewNop()}

	body := []byte{0x04, 0x00, 0xC0, 0x7F, 0xC0, 0x7F}
	r := packet.NewReader(append([]byte{packet.C_OPCODE_ENTER_SHIP}, body...))

	HandleEnterShip(sess, r, deps)

	if stub.consumedCount != 0 {
		t.Fatalf("沒有船票時不應消耗物品，實際=%d", stub.consumedCount)
	}
	updated := ws.GetBySession(sess.ID)
	if updated.MapID != 5 || updated.X != 32733 || updated.Y != 32796 {
		t.Fatalf("沒有船票時應留在原地，實際=(%d,%d,%d)", updated.MapID, updated.X, updated.Y)
	}
}
