package system

// MISS-P0-003 Stage C.5：QuestWorldSystem 測試。
//
// 測試策略：使用無 Entry.TeleportTo / Exit.TeleportTo 的副本定義，
// 避開 handler.TeleportPlayer 的重型依賴；只驗證系統內部的註冊表、
// ShowID 分配、時間限制、斷線清理等核心邏輯。

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

// newQuestWorldTestSystem 建立測試用 QuestWorldSystem。
// dungeonYAML 是 quest_dungeons.yaml 內容字串；若為空則用最小條目。
func newQuestWorldTestSystem(t *testing.T, dungeonYAML string) (*QuestWorldSystem, *world.State) {
	t.Helper()

	if dungeonYAML == "" {
		dungeonYAML = `
dungeons:
  - id: 999
    name: "測試副本（無傳送）"
    map_id: 999
    max_users: 4
    time_limit: -1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
`
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "quest_dungeons.yaml")
	if err := os.WriteFile(path, []byte(dungeonYAML), 0o644); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}
	tbl, err := data.LoadDungeonTable(path)
	if err != nil {
		t.Fatalf("載入測試 YAML 失敗: %v", err)
	}

	ws := world.NewState()
	deps := &handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
	}
	sys := NewQuestWorldSystem(ws, tbl, deps)
	return sys, ws
}

func newQuestWorldTestPlayer(charID int32, x, y int32, mapID int16) *world.PlayerInfo {
	return &world.PlayerInfo{
		CharID: charID,
		Name:   "測試玩家",
		X:      x,
		Y:      y,
		MapID:  mapID,
		Inv:    world.NewInventory(),
	}
}

// TestQuestWorldNextIDStartsFromHundred 流水號從 100 起（對齊 Java _nextId = 100）。
func TestQuestWorldNextIDStartsFromHundred(t *testing.T) {
	sys, _ := newQuestWorldTestSystem(t, "")
	if got := sys.NextID(); got != 100 {
		t.Fatalf("第一個 NextID 應為 100，實際 %d", got)
	}
	if got := sys.NextID(); got != 101 {
		t.Fatalf("第二個 NextID 應為 101，實際 %d", got)
	}
}

// TestQuestWorldEnterAssignsShowIDAndRegistersInstance Enter 應分配 ShowID 並註冊實例。
func TestQuestWorldEnterAssignsShowIDAndRegistersInstance(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 999)
	if inst == nil {
		t.Fatal("Enter 應返回實例")
	}
	if inst.ID != 100 {
		t.Fatalf("第一個 Enter 的 ShowID 應為 100，實際 %d", inst.ID)
	}
	if p.ShowID != 100 {
		t.Fatalf("玩家 ShowID 應設為 100，實際 %d", p.ShowID)
	}
	if inst.QuestID != 999 || inst.MapID != 999 {
		t.Fatalf("Instance 欄位錯誤: questID=%d mapID=%d", inst.QuestID, inst.MapID)
	}
	if inst.PlayerCount() != 1 {
		t.Fatalf("實例應有 1 玩家，實際 %d", inst.PlayerCount())
	}
	if sys.Count() != 1 {
		t.Fatalf("系統應有 1 實例，實際 %d", sys.Count())
	}
	if !sys.IsQuest(inst.ID) {
		t.Fatalf("IsQuest(%d) 應為 true", inst.ID)
	}
}

// TestQuestWorldEnterUnknownDungeonReturnsNil 不存在副本 ID 應回 nil。
func TestQuestWorldEnterUnknownDungeonReturnsNil(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	if inst := sys.Enter(p, 9999); inst != nil {
		t.Fatalf("不存在副本 ID 應回 nil，實際 %+v", inst)
	}
	if p.ShowID != 0 {
		t.Fatalf("Enter 失敗時 ShowID 應保持 0，實際 %d", p.ShowID)
	}
}

// TestQuestWorldJoinAddsToExistingInstance Join 加入既有實例不分配新 ShowID。
func TestQuestWorldJoinAddsToExistingInstance(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p1 := newQuestWorldTestPlayer(1001, 100, 100, 4)
	p2 := newQuestWorldTestPlayer(1002, 100, 100, 4)
	ws.AddPlayer(p1)
	ws.AddPlayer(p2)

	inst1 := sys.Enter(p1, 999)
	inst2 := sys.Join(p2, inst1.ID)
	if inst2 == nil || inst2.ID != inst1.ID {
		t.Fatalf("Join 應返回同一實例，inst1=%+v inst2=%+v", inst1, inst2)
	}
	if inst1.PlayerCount() != 2 {
		t.Fatalf("實例應有 2 玩家，實際 %d", inst1.PlayerCount())
	}
	if p2.ShowID != inst1.ID {
		t.Fatalf("p2 ShowID 應為 %d，實際 %d", inst1.ID, p2.ShowID)
	}
}

// TestQuestWorldJoinUnknownShowIDReturnsNil Join 不存在 showID 應回 nil。
func TestQuestWorldJoinUnknownShowIDReturnsNil(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	if inst := sys.Join(p, 99999); inst != nil {
		t.Fatalf("不存在 showID 應回 nil，實際 %+v", inst)
	}
}

// TestQuestWorldExitRemovesPlayerAndClearsShowID Exit 應移除玩家並清 ShowID。
func TestQuestWorldExitRemovesPlayerAndClearsShowID(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p1 := newQuestWorldTestPlayer(1001, 100, 100, 4)
	p2 := newQuestWorldTestPlayer(1002, 100, 100, 4)
	ws.AddPlayer(p1)
	ws.AddPlayer(p2)

	inst := sys.Enter(p1, 999)
	sys.Join(p2, inst.ID)

	if !sys.Exit(p1) {
		t.Fatal("Exit 應返回 true")
	}
	if p1.ShowID != 0 {
		t.Fatalf("Exit 後玩家 ShowID 應為 0，實際 %d", p1.ShowID)
	}
	if inst.PlayerCount() != 1 {
		t.Fatalf("實例應剩 1 玩家，實際 %d", inst.PlayerCount())
	}
	if sys.Count() != 1 {
		t.Fatal("仍有玩家在副本，實例應保留")
	}
}

// TestQuestWorldExitLastPlayerEndsInstance 最後一位玩家離開應結束副本。
func TestQuestWorldExitLastPlayerEndsInstance(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 999)
	showID := inst.ID
	sys.Exit(p)

	if sys.Count() != 0 {
		t.Fatalf("最後玩家離開後實例應清除，Count=%d", sys.Count())
	}
	if sys.IsQuest(showID) {
		t.Fatal("IsQuest 應為 false")
	}
}

// TestQuestWorldExitWithOutStopEndsInstance 開啟 out_stop 後任一玩家離開即結束。
func TestQuestWorldExitWithOutStopEndsInstance(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, `
dungeons:
  - id: 999
    name: "test"
    map_id: 999
    out_stop: true
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
`)
	p1 := newQuestWorldTestPlayer(1001, 100, 100, 4)
	p2 := newQuestWorldTestPlayer(1002, 100, 100, 4)
	ws.AddPlayer(p1)
	ws.AddPlayer(p2)

	inst := sys.Enter(p1, 999)
	sys.Join(p2, inst.ID)

	sys.Exit(p1)
	if sys.Count() != 0 {
		t.Fatalf("OutStop 模式下任一玩家離開應結束副本，Count=%d", sys.Count())
	}
	if p2.ShowID != 0 {
		t.Fatalf("剩餘玩家 ShowID 也應清空，實際 %d", p2.ShowID)
	}
}

// TestQuestWorldExitNotInDungeonReturnsFalse 不在副本中的玩家 Exit 應回 false。
func TestQuestWorldExitNotInDungeonReturnsFalse(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	if sys.Exit(p) {
		t.Fatal("不在副本中的玩家 Exit 應回 false")
	}
}

// TestQuestWorldRemoveOnDisconnect 斷線清理應同 Exit。
func TestQuestWorldRemoveOnDisconnect(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "")
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	sys.Enter(p, 999)
	sys.RemoveOnDisconnect(p)

	if p.ShowID != 0 {
		t.Fatalf("RemoveOnDisconnect 後 ShowID 應為 0，實際 %d", p.ShowID)
	}
	if sys.Count() != 0 {
		t.Fatalf("RemoveOnDisconnect 後實例應清除，Count=%d", sys.Count())
	}
}

// TestQuestWorldTimeLimitExpiresInstance 時間限制到期應自動結束副本。
func TestQuestWorldTimeLimitExpiresInstance(t *testing.T) {
	// 時間限制設 1 秒 = 5 ticks（questWorldTicksPerSecond）
	sys, ws := newQuestWorldTestSystem(t, `
dungeons:
  - id: 999
    name: "test"
    map_id: 999
    time_limit: 1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 99001, count: 1, fixed: [100, 100] }
`)
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 999)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}

	// 4 tick — 未到期
	for i := 0; i < 4; i++ {
		sys.Update(0)
	}
	if sys.Count() != 1 {
		t.Fatalf("4 tick 未到期，實例應保留，Count=%d", sys.Count())
	}
	// 第 5 tick — 達 1 秒，應到期
	sys.Update(0)
	if sys.Count() != 0 {
		t.Fatalf("5 tick 應觸發時間限制，Count=%d", sys.Count())
	}
	if p.ShowID != 0 {
		t.Fatalf("時間到期後玩家 ShowID 應為 0，實際 %d", p.ShowID)
	}
}

// TestQuestWorldTimeLimitDisabledKeepsInstance time_limit=-1 不過期。
func TestQuestWorldTimeLimitDisabledKeepsInstance(t *testing.T) {
	sys, ws := newQuestWorldTestSystem(t, "") // time_limit: -1
	p := newQuestWorldTestPlayer(1001, 100, 100, 4)
	ws.AddPlayer(p)

	sys.Enter(p, 999)
	for i := 0; i < 100; i++ {
		sys.Update(0)
	}
	if sys.Count() != 1 {
		t.Fatalf("無時間限制副本不應到期，Count=%d", sys.Count())
	}
}

// TestQuestWorldUpdateSetsTickMonotonic Update 應單調遞增 tick。
func TestQuestWorldUpdateSetsTickMonotonic(t *testing.T) {
	sys, _ := newQuestWorldTestSystem(t, "")
	initialTick := sys.tick
	sys.Update(time.Millisecond)
	if sys.tick != initialTick+1 {
		t.Fatalf("Update 應遞增 tick，before=%d after=%d", initialTick, sys.tick)
	}
}
