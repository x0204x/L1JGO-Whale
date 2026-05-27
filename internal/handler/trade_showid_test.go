package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestFindFaceToFaceSkipsDifferentShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := &world.PlayerInfo{
		SessionID: 1,
		Session:   newHandlerTestSession(t, 1),
		CharID:    7101,
		Name:      "Trader",
		MapID:     4,
		X:         33000,
		Y:         33000,
		Heading:   0,
		ShowID:    10,
	}
	ws.AddPlayer(player)
	target := &world.PlayerInfo{
		SessionID: 2,
		Session:   newHandlerTestSession(t, 2),
		CharID:    7102,
		Name:      "OtherShow",
		MapID:     4,
		X:         33000,
		Y:         32999,
		Heading:   4,
		ShowID:    99,
	}
	ws.AddPlayer(target)

	deps := &Deps{World: ws, Log: zap.NewNop()}

	if got := findFaceToFace(player, deps); got != nil {
		t.Fatalf("不同 ShowID 的面對面玩家不應成為交易對象，got=%s", got.Name)
	}
}
