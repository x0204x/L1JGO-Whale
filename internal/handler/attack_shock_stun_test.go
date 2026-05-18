package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleAttackBlockedByShockStunBeforeCombatQueueLikeJava(t *testing.T) {
	tests := []struct {
		name   string
		handle func(*net.Session, *packet.Reader, *Deps)
		melee  bool
	}{
		{name: "melee", handle: HandleAttack, melee: true},
		{name: "ranged", handle: HandleFarAttack, melee: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := world.NewState()
			sess := newHandlerTestSession(t, 1)
			player := &world.PlayerInfo{
				SessionID: sess.ID,
				Session:   sess,
				CharID:    1001,
				Name:      "stunned",
				Paralyzed: true,
				Inv:       world.NewInventory(),
			}
			player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
			ws.AddPlayer(player)

			combat := &captureCombatQueue{}
			deps := &Deps{World: ws, Combat: combat, Log: zap.NewNop()}

			tt.handle(sess, attackReader(2002), deps)

			if len(combat.requests) != 0 {
				t.Fatalf("Java C_Attack/C_AttackBow 會以 isParalyzedX 擋下 SHOCK_STUN 中的攻擊，不應排入 CombatQueue，requests=%+v", combat.requests)
			}
		})
	}
}

type captureCombatQueue struct {
	requests []AttackRequest
}

func (q *captureCombatQueue) QueueAttack(req AttackRequest) {
	q.requests = append(q.requests, req)
}

func (q *captureCombatQueue) HandleNpcDeath(_ *world.NpcInfo, _ *world.PlayerInfo, _ []*world.PlayerInfo) *NpcKillResult {
	return nil
}

func (q *captureCombatQueue) AddExp(_ *world.PlayerInfo, _ int32) {
}

func attackReader(targetID int32) *packet.Reader {
	w := packet.NewWriterWithOpcode(packet.C_OPCODE_ATTACK)
	w.WriteD(targetID)
	w.WriteH(0)
	w.WriteH(0)
	return packet.NewReader(w.Bytes())
}
