package system

import (
	"net"
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestDeathTombKillPlayerDoesNotSpawnTombByDefaultFor38Client(t *testing.T) {
	ws := world.NewState()
	player := newDeathTombPlayer(t, 1)
	ws.AddPlayer(player)

	death := NewDeathSystem(newDeathTombDeps(t, ws, false))
	death.KillPlayer(player)

	if player.TombEffectID != 0 {
		t.Fatalf("3.80C 原生客戶端沒有墓碑圖檔，預設不應生成墓碑，TombEffectID=%d", player.TombEffectID)
	}
	if effects := ws.GetNearbyGroundEffects(player.X, player.Y, player.MapID); len(effects) != 0 {
		t.Fatalf("預設關閉墓碑時不應生成 GroundEffect，effects=%d", len(effects))
	}
}

func TestDeathTombKillPlayerSpawnsYiweiTombWhenAssetEnabled(t *testing.T) {
	ws := world.NewState()
	player := newDeathTombPlayer(t, 1)
	ws.AddPlayer(player)

	death := NewDeathSystem(newDeathTombDeps(t, ws, true))
	death.KillPlayer(player)

	if player.TombEffectID == 0 {
		t.Fatal("玩家死亡後應記錄對應墓碑物件 ID")
	}
	tomb := ws.GetGroundEffect(player.TombEffectID)
	if tomb == nil {
		t.Fatal("玩家死亡後應在世界狀態生成墓碑 GroundEffect")
	}
	if tomb.NpcID != world.TombEffectNpcID || tomb.GfxID != world.TombEffectGfxID || tomb.Type != world.GroundEffectTomb {
		t.Fatalf("墓碑模板不符 yiwei，NpcID=%d GfxID=%d Type=%d", tomb.NpcID, tomb.GfxID, tomb.Type)
	}
	if tomb.X != player.X || tomb.Y != player.Y || tomb.MapID != player.MapID {
		t.Fatalf("墓碑座標應等於死亡座標，got=(%d,%d,%d)", tomb.X, tomb.Y, tomb.MapID)
	}
	if tomb.OwnerCharID != player.CharID || tomb.OwnerName != player.Name || tomb.Lawful != player.Lawful {
		t.Fatalf("墓碑應記錄死亡玩家資訊，Owner=%d/%s Lawful=%d", tomb.OwnerCharID, tomb.OwnerName, tomb.Lawful)
	}
	if tomb.TicksLeft != 300*groundEffectTickSec {
		t.Fatalf("yiwei 墓碑存在時間應為 300 秒，TicksLeft=%d", tomb.TicksLeft)
	}
}

func TestDeathTombProcessRestartClearsTomb(t *testing.T) {
	ws := world.NewState()
	player := newDeathTombPlayer(t, 1)
	player.Dead = true
	player.HP = 0
	tomb := &world.GroundEffect{
		ID:          world.NextGroundEffectID(),
		NpcID:       world.TombEffectNpcID,
		GfxID:       world.TombEffectGfxID,
		Type:        world.GroundEffectTomb,
		X:           player.X,
		Y:           player.Y,
		MapID:       player.MapID,
		OwnerCharID: player.CharID,
		OwnerName:   player.Name,
		TicksLeft:   300 * groundEffectTickSec,
	}
	player.TombEffectID = tomb.ID
	ws.AddPlayer(player)
	ws.AddGroundEffect(tomb)

	death := NewDeathSystem(newDeathTombDeps(t, ws, true))
	death.ProcessRestart(player.Session, player)

	if player.TombEffectID != 0 {
		t.Fatalf("死亡回村後應清除墓碑指標，TombEffectID=%d", player.TombEffectID)
	}
	if ws.GetGroundEffect(tomb.ID) != nil {
		t.Fatal("死亡回村後應從世界狀態移除墓碑")
	}
}

func newDeathTombPlayer(t *testing.T, sessionID uint64) *world.PlayerInfo {
	t.Helper()
	return &world.PlayerInfo{
		SessionID: sessionID,
		Session:   newDeathTombSession(t, sessionID),
		CharID:    int32(1000 + sessionID),
		Name:      "dead",
		X:         32767,
		Y:         32768,
		MapID:     4,
		Level:     52,
		Lawful:    1234,
		HP:        500,
		MaxHP:     500,
		MP:        200,
		MaxMP:     200,
		Food:      40,
		Inv:       world.NewInventory(),
		Known:     world.NewKnownEntities(),
	}
}

func newDeathTombDeps(t *testing.T, ws *world.State, enableTombEffect bool) *handler.Deps {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	return &handler.Deps{
		World:     ws,
		Scripting: engine,
		Config: &config.Config{
			Gameplay: config.GameplayConfig{
				InitialFood:      40,
				EnableTombEffect: enableTombEffect,
			},
		},
		Log: zap.NewNop(),
	}
}

func newDeathTombSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 64, 64, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}
