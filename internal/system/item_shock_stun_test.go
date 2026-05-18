package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newShockStunItemUseSystem(t *testing.T, ws *world.State) *ItemUseSystem {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	t.Cleanup(engine.Close)
	return NewItemUseSystem(&handler.Deps{
		World:     ws,
		Scripting: engine,
		Log:       zap.NewNop(),
	})
}

func TestItemUseConsumableBlockedByShockStunLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "stunned",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        10,
		MaxHP:     100,
		Paralyzed: true,
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	potion := player.Inv.AddItemWithID(7001, 40010, 1, "治癒藥水", 189, 1000, true, 1)
	sys := newShockStunItemUseSystem(t, ws)

	used := sys.UseConsumable(player.Session, player, potion, nil)

	if used {
		t.Fatal("Java C_ItemUSe 會以 isParalyzedX 擋下 SHOCK_STUN 中的道具使用，不應消耗藥水")
	}
	if player.HP != 10 {
		t.Fatalf("SHOCK_STUN 中喝藥水不應恢復 HP，HP=%d", player.HP)
	}
	if item := player.Inv.FindByObjectID(potion.ObjectID); item == nil || item.Count != 1 {
		t.Fatalf("SHOCK_STUN 中喝藥水不應移除物品，item=%+v", item)
	}
}

func TestItemUseTeleportScrollBlockedByShockStunLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "stunned",
		X:         100,
		Y:         100,
		MapID:     4,
		Paralyzed: true,
	})
	player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	scroll := player.Inv.AddItemWithID(7002, 40100, 1, "瞬間移動卷軸", 0, 1000, true, 1)
	sys := newShockStunItemUseSystem(t, ws)

	sys.UseTeleportScroll(player.Session, teleportScrollUseReader(), player, scroll)

	if player.X != 100 || player.Y != 100 || player.MapID != 4 {
		t.Fatalf("SHOCK_STUN 中使用傳送卷軸不應移動，位置=%d,%d,%d", player.X, player.Y, player.MapID)
	}
	if item := player.Inv.FindByObjectID(scroll.ObjectID); item == nil || item.Count != 1 {
		t.Fatalf("SHOCK_STUN 中使用傳送卷軸不應移除物品，item=%+v", item)
	}
	if !hasTeleportUnlockPacket(drainSkillTestPackets(player.Session)) {
		t.Fatal("Java C_ItemUSe 在 isParalyzedX 擋下道具使用時會送 TeleportUnlock")
	}
}

func teleportScrollUseReader() *packet.Reader {
	w := packet.NewWriterWithOpcode(packet.C_OPCODE_USE_ITEM)
	w.WriteH(0)
	w.WriteD(0)
	return packet.NewReader(w.Bytes())
}
