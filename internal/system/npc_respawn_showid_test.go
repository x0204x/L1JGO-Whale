package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestNpcRespawnBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addNpcTeleportHomeTestPlayer(t, ws, 1, 1001, "same_show", 100, 3)
	otherShow := addNpcTeleportHomeTestPlayer(t, ws, 2, 1002, "other_show", 100, 8)
	npc := &world.NpcInfo{
		ID:         2001,
		NpcID:      45040,
		Impl:       "L1Monster",
		Name:       "respawn_npc",
		X:          101,
		Y:          100,
		MapID:      900,
		ShowID:     3,
		SpawnX:     100,
		SpawnY:     100,
		SpawnMapID: 900,
		Dead:       true,
		HP:         0,
		MaxHP:      100,
		MP:         0,
		MaxMP:      10,
	}
	ws.AddNpc(npc)
	s := NewNpcRespawnSystem(ws, newSkillLOSTestMap(t), &handler.Deps{World: ws, Log: zap.NewNop()})

	s.respawnNpc(npc)

	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC 重生顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), npc.ID) {
		t.Fatal("yiwei NPC 顯示走同 ShowID 可見邊界，不同 ShowID 不應收到 NPC 重生顯示封包")
	}
}

func TestNpcCorpseCleanupBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addNpcTeleportHomeTestPlayer(t, ws, 1, 1001, "same_show", 100, 3)
	otherShow := addNpcTeleportHomeTestPlayer(t, ws, 2, 1002, "other_show", 100, 8)
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45040,
		Impl:         "L1Monster",
		Name:         "corpse_npc",
		X:            100,
		Y:            100,
		MapID:        900,
		ShowID:       3,
		SpawnX:       100,
		SpawnY:       100,
		SpawnMapID:   900,
		Dead:         true,
		DeleteTimer:  1,
		RespawnTimer: 10,
		MaxHP:        100,
		MaxMP:        10,
	}
	ws.AddNpc(npc)
	s := NewNpcRespawnSystem(ws, newSkillLOSTestMap(t), &handler.Deps{World: ws, Log: zap.NewNop()})

	s.Update(0)

	if !hasRemoveObjectPacket(drainSkillTestPackets(sameShow.Session), npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC 屍體移除封包")
	}
	if hasRemoveObjectPacket(drainSkillTestPackets(otherShow.Session), npc.ID) {
		t.Fatal("yiwei NPC remove 走同 ShowID 可見邊界，不同 ShowID 不應收到屍體移除封包")
	}
}

func TestNpcMobGroupRespawnBroadcastsMinionsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	sameShow := addNpcTeleportHomeTestPlayer(t, ws, 1, 1001, "same_show", 101, 3)
	otherShow := addNpcTeleportHomeTestPlayer(t, ws, 2, 1002, "other_show", 101, 8)
	leader := &world.NpcInfo{
		ID:         2001,
		NpcID:      45040,
		Impl:       "L1Monster",
		Name:       "leader",
		X:          100,
		Y:          100,
		MapID:      900,
		ShowID:     3,
		SpawnX:     100,
		SpawnY:     100,
		SpawnMapID: 900,
		HP:         100,
		MaxHP:      100,
	}
	ws.AddNpc(leader)
	npcs := loadNpcRespawnShowIDTestNpcTable(t)
	s := NewNpcRespawnSystem(ws, newSkillLOSTestMap(t), &handler.Deps{
		World: ws,
		Npcs:  npcs,
		Log:   zap.NewNop(),
	})
	group := &data.MobGroup{
		ID:       7,
		LeaderID: leader.NpcID,
		Minions:  []data.MinionEntry{{NpcID: 45041, Count: 1}},
	}

	s.respawnMobGroup(leader, group)

	minion := newestNpcExcept(t, ws, leader.ID)
	if minion.ShowID != leader.ShowID {
		t.Fatalf("yiwei mob group minion 應繼承 leader ShowID，got %d want %d", minion.ShowID, leader.ShowID)
	}
	if !hasPutObjectPacket(drainSkillTestPackets(sameShow.Session), minion.ID) {
		t.Fatal("同 ShowID 玩家應收到 mob group 隊員顯示封包")
	}
	if hasPutObjectPacket(drainSkillTestPackets(otherShow.Session), minion.ID) {
		t.Fatal("yiwei mob group 隊員顯示走同 ShowID 可見邊界，不同 ShowID 不應收到顯示封包")
	}
}

func loadNpcRespawnShowIDTestNpcTable(t *testing.T) *data.NpcTable {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "npc_list.yaml")
	raw := []byte(`npcs:
  - npc_id: 45041
    name: minion
    nameid: minion
    impl: L1Monster
    gfx_id: 1
    level: 10
    hp: 50
    mp: 10
    str: 10
    dex: 10
    con: 10
    wis: 10
    intel: 10
    mr: 0
    exp: 1
    lawful: 0
    size: small
    ranged: 1
    atk_speed: 1000
    sub_magic_speed: 1000
    passive_speed: 1000
`)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("寫入 NPC 測試資料失敗: %v", err)
	}
	npcs, err := data.LoadNpcTable(path)
	if err != nil {
		t.Fatalf("載入 NPC 測試資料失敗: %v", err)
	}
	return npcs
}

func newestNpcExcept(t *testing.T, ws *world.State, excludedID int32) *world.NpcInfo {
	t.Helper()
	for i := len(ws.NpcList()) - 1; i >= 0; i-- {
		npc := ws.NpcList()[i]
		if npc.ID != excludedID {
			return npc
		}
	}
	t.Fatal("找不到新生成的 NPC")
	return nil
}
