package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestGMLawfulParsesSignedDelta(t *testing.T) {
	capture := &captureGMCommandManager{}
	sess, player, deps := newGMItemCommandTestContext(t, capture)

	HandleGMCommand(sess, player, ".lawful +1000", deps)

	if !capture.lawfulCalled {
		t.Fatal(".lawful +數值 應委派給 GMCommandManager.AdjustLawful")
	}
	if capture.lawfulDelta != 1000 {
		t.Fatalf("正義值調整量錯誤: got=%d want=1000", capture.lawfulDelta)
	}

	capture.lawfulCalled = false
	HandleGMCommand(sess, player, ".lawful -500", deps)

	if !capture.lawfulCalled {
		t.Fatal(".lawful -數值 應委派給 GMCommandManager.AdjustLawful")
	}
	if capture.lawfulDelta != -500 {
		t.Fatalf("正義值調整量錯誤: got=%d want=-500", capture.lawfulDelta)
	}
}

func TestGMLawfulCanTargetOnlinePlayerByName(t *testing.T) {
	capture := &captureGMCommandManager{}
	sess, player, deps := newGMItemCommandTestContext(t, capture)
	targetSess := newHandlerTestSession(t, 2)
	target := &world.PlayerInfo{
		SessionID: targetSess.ID,
		Session:   targetSess,
		CharID:    2002,
		Name:      "target",
		Inv:       world.NewInventory(),
	}
	ws := world.NewState()
	ws.AddPlayer(player)
	ws.AddPlayer(target)
	deps.World = ws

	if !HandleGMCommand(sess, player, ".lawful +300 target", deps) {
		t.Fatal(".lawful 應該被 GM 指令處理器消耗")
	}
	if !capture.lawfulCalled || capture.lawfulDelta != 300 || capture.lawfulPlayer != target {
		t.Fatalf("指定目標解析錯誤: called=%v delta=%d target=%v", capture.lawfulCalled, capture.lawfulDelta, capture.lawfulPlayer)
	}
}
