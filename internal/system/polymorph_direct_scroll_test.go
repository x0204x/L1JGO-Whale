package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestUseDirectPolyScrollTransformsOrcEmissaryAndConsumesScroll(t *testing.T) {
	ws := world.NewState()
	sess := &l1net.Session{ID: 1}
	player := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "orc-scroll",
		Level:     50,
		ClassType: 5,
		ClassID:   6658,
		Sex:       0,
	})
	scroll := player.Inv.AddItemWithID(7001, 49220, 1, "妖魔密使變形卷軸", 472, 1000, true, 1)
	poly := NewPolymorphSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

	poly.UseDirectPolyScroll(sess, player, scroll)

	if player.TempCharGfx != 6984 || player.PolyID != 6984 {
		t.Fatalf("妖魔密使變形卷軸應直接變成 GFX 6984，TempCharGfx=%d PolyID=%d", player.TempCharGfx, player.PolyID)
	}
	if player.Inv.FindByItemID(49220) != nil {
		t.Fatal("直接變身成功後應消耗妖魔密使變形卷軸")
	}
}

func TestUseDirectPolyScrollTransformsSharnaByClassAndSex(t *testing.T) {
	ws := world.NewState()
	sess := &l1net.Session{ID: 1}
	player := addPolymorphTestPlayer(ws, &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "sharna-scroll",
		Level:     70,
		ClassType: 1,
		ClassID:   48,
		Sex:       1,
	})
	scroll := player.Inv.AddItemWithID(7001, 49155, 1, "夏納的變身卷軸(等級70)", 2975, 630, true, 1)
	poly := NewPolymorphSystem(&handler.Deps{World: ws, Log: zap.NewNop()})

	poly.UseDirectPolyScroll(sess, player, scroll)

	if player.TempCharGfx != 6885 || player.PolyID != 6885 {
		t.Fatalf("女騎士使用等級70夏納卷軸應變成 GFX 6885，TempCharGfx=%d PolyID=%d", player.TempCharGfx, player.PolyID)
	}
	if player.Inv.FindByItemID(49155) != nil {
		t.Fatal("夏納直接變身成功後應消耗卷軸")
	}
}
