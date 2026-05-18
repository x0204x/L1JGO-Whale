package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleMoveBlockedByShockStunLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "stunned",
		X:         100,
		Y:         100,
		MapID:     4,
		Paralyzed: true,
	}
	player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	ws.AddPlayer(player)

	deps := &Deps{
		World:  ws,
		Config: &config.Config{Server: config.ServerConfig{Language: 5}},
		Log:    zap.NewNop(),
	}

	HandleMove(sess, moveReader(100, 100, 2), deps)

	if player.X != 100 || player.Y != 100 || player.MapID != 4 {
		t.Fatalf("Java C_MoveChar 會以 isParalyzedX 擋下 SHOCK_STUN 中的移動，不應改變座標，位置=%d,%d,%d", player.X, player.Y, player.MapID)
	}
	if player.LastMoveTime != 0 {
		t.Fatalf("SHOCK_STUN 中移動被拒絕時不應更新 LastMoveTime，got=%d", player.LastMoveTime)
	}
}

func moveReader(x, y uint16, heading byte) *packet.Reader {
	w := packet.NewWriterWithOpcode(packet.C_OPCODE_MOVE)
	w.WriteH(x)
	w.WriteH(y)
	w.WriteC(heading)
	return packet.NewReader(w.Bytes())
}
