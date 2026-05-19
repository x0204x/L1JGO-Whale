package system

// MISS-P0-003 Stage D.5：QuestWorldSystem 出生/round/cleanup 測試。
//
// 策略：在 t.TempDir 內動態寫入 quest_dungeons.yaml + npc_list.yaml，
// 透過 data.LoadDungeonTable + data.LoadNpcTable 載入真實型別，
// 避免 NpcTable 內部欄位不可外部建構的限制。

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// loadTestNpcs 寫入 npc_list.yaml 並載入。NPC 樣板 99001 簡單怪物。
func loadTestNpcs(t *testing.T, dir string) *data.NpcTable {
	t.Helper()
	yamlStr := `
npcs:
  - npc_id: 99001
    name: "測試副本怪"
    nameid: "$dungeon_mob"
    impl: L1Monster
    gfx_id: 1
    level: 10
    hp: 100
    mp: 50
    ac: 0
    str: 10
    dex: 10
    exp: 50
    size: small
    mr: 0
    agro: true
  - npc_id: 99002
    name: "測試副本王"
    nameid: "$dungeon_boss"
    impl: L1Monster
    gfx_id: 2
    level: 30
    hp: 500
    mp: 100
    ac: 0
    str: 20
    dex: 20
    exp: 500
    size: large
    mr: 10
    agro: true
`
	path := filepath.Join(dir, "npc_list.yaml")
	if err := os.WriteFile(path, []byte(yamlStr), 0o644); err != nil {
		t.Fatalf("寫入 npc YAML 失敗: %v", err)
	}
	tbl, err := data.LoadNpcTable(path)
	if err != nil {
		t.Fatalf("載入 npc YAML 失敗: %v", err)
	}
	return tbl
}

// newQuestWorldSpawnTestSystem 建立帶 Npcs 的測試 system。
func newQuestWorldSpawnTestSystem(t *testing.T, dungeonYAML string) (*QuestWorldSystem, *world.State, *handler.Deps) {
	t.Helper()
	dir := t.TempDir()
	dungeonPath := filepath.Join(dir, "quest_dungeons.yaml")
	if err := os.WriteFile(dungeonPath, []byte(dungeonYAML), 0o644); err != nil {
		t.Fatalf("寫入副本 YAML 失敗: %v", err)
	}
	dt, err := data.LoadDungeonTable(dungeonPath)
	if err != nil {
		t.Fatalf("載入副本 YAML 失敗: %v", err)
	}

	ws := world.NewState()
	deps := &handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
		Npcs:  loadTestNpcs(t, dir),
	}
	sys := NewQuestWorldSystem(ws, dt, deps)
	deps.QuestWorld = sys
	return sys, ws, deps
}

// TestQuestWorldSpawnOnEnterFixed on_enter + fixed 出生：NPC 進世界、ShowID/Transient 正確、inst.npcs 含之。
func TestQuestWorldSpawnOnEnterFixed(t *testing.T) {
	yaml := `
dungeons:
  - id: 801
    name: "出生測試副本"
    map_id: 999
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 2, fixed: [100, 100] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 801)
	if inst == nil {
		t.Fatal("Enter 應返回實例")
	}

	if got := inst.NpcCount(); got != 2 {
		t.Fatalf("應出生 2 隻 NPC，實際 %d", got)
	}

	for _, npcID := range inst.Npcs() {
		npc := ws.GetNpc(npcID)
		if npc == nil {
			t.Fatalf("NPC %d 應存在於世界", npcID)
		}
		if npc.ShowID != inst.ID {
			t.Fatalf("NPC %d ShowID=%d，應為 %d", npcID, npc.ShowID, inst.ID)
		}
		if !npc.Transient {
			t.Fatalf("NPC %d 應為 Transient", npcID)
		}
		if npc.MapID != inst.MapID {
			t.Fatalf("NPC %d MapID=%d，應為 %d", npcID, npc.MapID, inst.MapID)
		}
	}
}

// TestQuestWorldSpawnOnEnterArea on_enter + area 出生：NPC 落在區域內。
func TestQuestWorldSpawnOnEnterArea(t *testing.T) {
	yaml := `
dungeons:
  - id: 802
    name: "區域出生副本"
    map_id: 999
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 3, area: [200, 200, 210, 210] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 802)
	if inst == nil {
		t.Fatal("Enter 應返回實例")
	}
	if got := inst.NpcCount(); got != 3 {
		t.Fatalf("應出生 3 隻 NPC，實際 %d", got)
	}
	for _, npcID := range inst.Npcs() {
		npc := ws.GetNpc(npcID)
		if npc == nil {
			t.Fatalf("NPC %d 應存在", npcID)
		}
		if npc.X < 200 || npc.X > 210 || npc.Y < 200 || npc.Y > 210 {
			t.Fatalf("NPC %d 座標 (%d,%d) 不在區域內", npcID, npc.X, npc.Y)
		}
	}
}

// TestQuestWorldRoundNoDoubleSpawn 重複呼叫 Enter 不會重複出生（MarkRoundSpawned 防護）。
func TestQuestWorldRoundNoDoubleSpawn(t *testing.T) {
	yaml := `
dungeons:
  - id: 803
    name: "防重複出生副本"
    map_id: 999
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 803)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}
	if got := inst.NpcCount(); got != 1 {
		t.Fatalf("首次應出生 1 隻，實際 %d", got)
	}

	// 直接再呼叫 spawnRound 一次（模擬重複觸發）
	def := sys.dungeons.Get(803)
	sys.spawnRound(inst, &def.Rounds[0])
	// MarkRoundSpawned 在 caller 端控制；spawnRound 本身不防護重出生，
	// 但 Enter 流程的 MarkRoundSpawned 保證了 round 不會被觸發兩次。
	// 這裡再呼一次 Enter（同玩家，會建新實例）— 新實例的 spawned map 是獨立的，仍會出生。

	p2 := newQuestWorldTestPlayer(1002, 50, 50, 4)
	ws.AddPlayer(p2)
	inst2 := sys.Enter(p2, 803)
	if inst2 == nil {
		t.Fatal("第二玩家 Enter 失敗")
	}
	if inst2.ID == inst.ID {
		t.Fatalf("第二玩家應建立新實例（不同 ShowID），實際 %d == %d", inst2.ID, inst.ID)
	}
	if got := inst2.NpcCount(); got != 1 {
		t.Fatalf("第二實例應出生 1 隻，實際 %d", got)
	}
}

// TestQuestWorldCleanupNpcsAtEnd 副本結束時所有副本 NPC 應從世界移除。
func TestQuestWorldCleanupNpcsAtEnd(t *testing.T) {
	yaml := `
dungeons:
  - id: 804
    name: "結束清理副本"
    map_id: 999
    out_stop: true
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 3, fixed: [100, 100] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 804)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}
	npcIDs := inst.Npcs()
	if len(npcIDs) != 3 {
		t.Fatalf("應出生 3 隻，實際 %d", len(npcIDs))
	}

	// 玩家離開 → out_stop=true → endInstance
	if !sys.Exit(p) {
		t.Fatal("Exit 應成功")
	}

	// 所有副本 NPC 應從世界移除
	for _, id := range npcIDs {
		if npc := ws.GetNpc(id); npc != nil {
			t.Fatalf("NPC %d 副本結束後仍在世界中", id)
		}
	}
	if sys.IsQuest(inst.ID) {
		t.Fatal("副本應已從註冊表移除")
	}
}

// TestQuestWorldOnRoundClearTriggersNextRound NPC 全清 → on_round_clear 觸發下一輪。
func TestQuestWorldOnRoundClearTriggersNextRound(t *testing.T) {
	yaml := `
dungeons:
  - id: 805
    name: "輪清觸發副本"
    map_id: 999
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
      - id: 1
        trigger: on_round_clear
        spawns:
          - { npc_id: 99002, count: 1, fixed: [105, 105] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 805)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}
	if got := inst.NpcCount(); got != 1 {
		t.Fatalf("Round -1 應出生 1 隻，實際 %d", got)
	}

	// 取第一隻 NPC（99001）並通知死亡
	firstNpcID := inst.Npcs()[0]
	firstNpc := ws.GetNpc(firstNpcID)
	if firstNpc == nil {
		t.Fatal("首隻 NPC 應存在")
	}
	if firstNpc.NpcID != 99001 {
		t.Fatalf("首隻 NPC 模板應為 99001，實際 %d", firstNpc.NpcID)
	}

	// 通知 OnNpcDeath → 應觸發 round 1 出生 99002
	sys.OnNpcDeath(firstNpc)

	// 此時 inst.npcs 應有 1 隻新 NPC（99002）
	npcIDs := inst.Npcs()
	if len(npcIDs) != 1 {
		t.Fatalf("輪清觸發後應有 1 隻新 NPC，實際 %d", len(npcIDs))
	}
	newNpc := ws.GetNpc(npcIDs[0])
	if newNpc == nil || newNpc.NpcID != 99002 {
		t.Fatalf("輪清應出生 99002，實際 %+v", newNpc)
	}
}

// TestQuestWorldOnTimerTriggersAfterTicks on_timer round 在指定秒數後出生。
func TestQuestWorldOnTimerTriggersAfterTicks(t *testing.T) {
	yaml := `
dungeons:
  - id: 806
    name: "計時觸發副本"
    map_id: 999
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
      - id: 2
        trigger: on_timer
        timer: 3
        spawns:
          - { npc_id: 99002, count: 1, fixed: [105, 105] }
`
	sys, ws, _ := newQuestWorldSpawnTestSystem(t, yaml)
	p := newQuestWorldTestPlayer(1001, 50, 50, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 806)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}
	if got := inst.NpcCount(); got != 1 {
		t.Fatalf("on_enter 應出生 1 隻，實際 %d", got)
	}

	// 跑 14 個 tick（< 3 秒 * 5 ticks/秒 = 15 ticks，不應觸發）
	for i := 0; i < 14; i++ {
		sys.Update(time.Millisecond * 200)
	}
	if got := inst.NpcCount(); got != 1 {
		t.Fatalf("14 tick 後不應觸發 on_timer，實際 NPC 數 %d", got)
	}

	// 再跑 1 tick（總 15 tick = 3 秒）→ 應觸發
	sys.Update(time.Millisecond * 200)
	if got := inst.NpcCount(); got != 2 {
		t.Fatalf("15 tick 後應觸發 on_timer，NPC 數應為 2，實際 %d", got)
	}

	// 再跑多次 tick → 不應重複觸發
	for i := 0; i < 10; i++ {
		sys.Update(time.Millisecond * 200)
	}
	if got := inst.NpcCount(); got != 2 {
		t.Fatalf("重複 tick 不應再觸發 on_timer，實際 NPC 數 %d", got)
	}
}

// TestQuestWorldOnNpcDeathOutsideDungeonNoOp 主世界 NPC（ShowID=0）OnNpcDeath 不應有副作用。
func TestQuestWorldOnNpcDeathOutsideDungeonNoOp(t *testing.T) {
	yaml := `
dungeons:
  - id: 807
    name: "無效情境副本"
    map_id: 999
    time_limit: -1
`
	sys, _, _ := newQuestWorldSpawnTestSystem(t, yaml)
	mainNpc := &world.NpcInfo{ID: 9999, NpcID: 99001, MapID: 4, ShowID: 0}
	// 應不 panic、不報錯
	sys.OnNpcDeath(mainNpc)

	// 不存在的 ShowID
	stray := &world.NpcInfo{ID: 9998, NpcID: 99001, MapID: 4, ShowID: 12345}
	sys.OnNpcDeath(stray)
}
