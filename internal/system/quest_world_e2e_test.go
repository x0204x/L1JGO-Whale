package system

// MISS-P0-003 Stage E：副本框架端對端驗證測試。
//
// 對應 dev doc Stage E item 25 的驗收條件：
//   - 兩個玩家進入同副本 → 應屬於同實例、互相可見（ShowID 相同）
//   - 第三個玩家進入 → 應建立新實例、與前兩人不互相可見
//
// 本測試直接驅動 QuestWorldSystem.Enter / Exit / Join，
// 跳過真實客戶端網路層；驗證 Stage A-D 框架（Round 引擎、
// ShowID AOI、NPC 生命週期）串成可用副本流程。

import (
	stdnet "net"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// newE2ETestSession 建立有效但無實際 I/O 的 Session（用 net.Pipe）。
func newE2ETestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() { _ = client.Close() })
	sess := l1net.NewSession(server, id, 8, 8, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

// newE2EDemoSystem 建立一個帶 demo 副本（id=1001）的 QuestWorldSystem。
func newE2EDemoSystem(t *testing.T) (*QuestWorldSystem, *world.State) {
	t.Helper()
	dir := t.TempDir()

	dungeonYAML := `
dungeons:
  - id: 1001
    name: "副本框架驗證 demo"
    map_id: 6666
    max_users: 4
    time_limit: -1
    rounds:
      - id: -1
        trigger: on_enter
        spawns:
          - { npc_id: 99001, count: 2, fixed: [100, 100] }
`
	dPath := filepath.Join(dir, "quest_dungeons.yaml")
	if err := os.WriteFile(dPath, []byte(dungeonYAML), 0o644); err != nil {
		t.Fatalf("寫副本 YAML: %v", err)
	}
	dt, err := data.LoadDungeonTable(dPath)
	if err != nil {
		t.Fatalf("載入副本 YAML: %v", err)
	}

	ws := world.NewState()
	deps := &handler.Deps{
		World: ws,
		Log:   zap.NewNop(),
		Npcs:  loadTestNpcs(t, dir),
	}
	sys := NewQuestWorldSystem(ws, dt, deps)
	deps.QuestWorld = sys
	return sys, ws
}

// TestQuestWorldE2ESamePartySeesSameInstance 兩玩家用 Join 進入同實例 → ShowID 一致、互相可見。
func TestQuestWorldE2ESamePartySeesSameInstance(t *testing.T) {
	sys, ws := newE2EDemoSystem(t)

	// 兩玩家直接放在副本地圖同一座標（避開「Enter 不會更新 AOI 位置」的測試環境差異）。
	pA := newQuestWorldTestPlayer(2001, 100, 100, 6666)
	pA.SessionID = 10001
	pA.Session = newE2ETestSession(t, pA.SessionID)
	ws.AddPlayer(pA)
	instA := sys.Enter(pA, 1001)
	if instA == nil {
		t.Fatal("玩家 A Enter 應成功")
	}
	if pA.ShowID == 0 {
		t.Fatal("玩家 A ShowID 應為非零（副本內）")
	}

	pB := newQuestWorldTestPlayer(2002, 100, 100, 6666)
	pB.SessionID = 10002
	pB.Session = newE2ETestSession(t, pB.SessionID)
	ws.AddPlayer(pB)
	instB := sys.Join(pB, instA.ID)
	if instB == nil {
		t.Fatal("玩家 B Join 應成功")
	}
	if instB.ID != instA.ID {
		t.Fatalf("B 應加入同實例 %d，實際 %d", instA.ID, instB.ID)
	}
	if pB.ShowID != pA.ShowID {
		t.Fatalf("B ShowID(%d) 應等於 A(%d)", pB.ShowID, pA.ShowID)
	}

	viewersForA := ws.GetNearbyPlayersInShow(pA.X, pA.Y, pA.MapID, pA.SessionID, pA.ShowID)
	if len(viewersForA) != 1 {
		t.Fatalf("玩家 A 視野內應有 1 個玩家（B），實際 %d", len(viewersForA))
	}
	if viewersForA[0].CharID != pB.CharID {
		t.Fatalf("A 應看見 B(%d)，實際看見 %d", pB.CharID, viewersForA[0].CharID)
	}

	if got := instA.PlayerCount(); got != 2 {
		t.Fatalf("實例應有 2 玩家，實際 %d", got)
	}
}

// TestQuestWorldE2EThirdPlayerSeparateInstance 第三玩家 Enter → 新實例、與前兩人不互相可見。
func TestQuestWorldE2EThirdPlayerSeparateInstance(t *testing.T) {
	sys, ws := newE2EDemoSystem(t)

	pA := newQuestWorldTestPlayer(3001, 100, 100, 6666)
	pA.SessionID = 20001
	pA.Session = newE2ETestSession(t, pA.SessionID)
	ws.AddPlayer(pA)
	instA := sys.Enter(pA, 1001)
	if instA == nil {
		t.Fatal("A Enter 失敗")
	}
	pB := newQuestWorldTestPlayer(3002, 100, 100, 6666)
	pB.SessionID = 20002
	pB.Session = newE2ETestSession(t, pB.SessionID)
	ws.AddPlayer(pB)
	_ = sys.Join(pB, instA.ID)

	// 第三玩家自行 Enter（不 Join）→ 應分到新實例
	pC := newQuestWorldTestPlayer(3003, 100, 100, 6666)
	pC.SessionID = 20003
	pC.Session = newE2ETestSession(t, pC.SessionID)
	ws.AddPlayer(pC)
	instC := sys.Enter(pC, 1001)
	if instC == nil {
		t.Fatal("C Enter 失敗")
	}
	if instC.ID == instA.ID {
		t.Fatalf("C 應有新 ShowID（不等於 %d），實際 %d", instA.ID, instC.ID)
	}
	if pC.ShowID == pA.ShowID {
		t.Fatalf("C ShowID(%d) 不應等於 A(%d)", pC.ShowID, pA.ShowID)
	}

	// AOI 隔離驗證：A 不應看見 C，C 不應看見 A/B
	viewersForA := ws.GetNearbyPlayersInShow(pA.X, pA.Y, pA.MapID, pA.SessionID, pA.ShowID)
	for _, v := range viewersForA {
		if v.CharID == pC.CharID {
			t.Fatalf("A 不應看見 C（不同副本實例），實際看見了")
		}
	}
	viewersForC := ws.GetNearbyPlayersInShow(pC.X, pC.Y, pC.MapID, pC.SessionID, pC.ShowID)
	for _, v := range viewersForC {
		if v.CharID == pA.CharID || v.CharID == pB.CharID {
			t.Fatalf("C 不應看見 A/B（不同副本實例），實際看見 %d", v.CharID)
		}
	}

	// 兩實例都有自己的 NPC（on_enter round 各自觸發 2 隻）
	if got := instA.NpcCount(); got != 2 {
		t.Fatalf("A 副本應 2 隻 NPC，實際 %d", got)
	}
	if got := instC.NpcCount(); got != 2 {
		t.Fatalf("C 副本應 2 隻 NPC，實際 %d", got)
	}

	// NPC 也應 ShowID 隔離：C 副本內看不到 A 副本的 NPC
	npcsForC := ws.GetNearbyNpcsInShow(pC.X, pC.Y, pC.MapID, pC.ShowID)
	if len(npcsForC) != 2 {
		t.Fatalf("C 副本內視野 NPC 應 2 隻，實際 %d", len(npcsForC))
	}
	for _, n := range npcsForC {
		if n.ShowID != pC.ShowID {
			t.Fatalf("C 視野內 NPC %d ShowID=%d，應為 %d", n.ID, n.ShowID, pC.ShowID)
		}
	}
}

// TestQuestWorldE2EFullCycle 完整生命週期：Enter → spawn → Exit → NPC 全清 → 玩家 ShowID 歸零。
func TestQuestWorldE2EFullCycle(t *testing.T) {
	sys, ws := newE2EDemoSystem(t)

	p := newQuestWorldTestPlayer(4001, 100, 100, 6666)
	p.SessionID = 30001
	p.Session = newE2ETestSession(t, p.SessionID)
	ws.AddPlayer(p)

	inst := sys.Enter(p, 1001)
	if inst == nil {
		t.Fatal("Enter 失敗")
	}
	npcIDs := inst.Npcs()
	if len(npcIDs) != 2 {
		t.Fatalf("Round -1 應 spawn 2 隻，實際 %d", len(npcIDs))
	}

	// 玩家退出（最後一人 → 觸發 endInstance）
	if !sys.Exit(p) {
		t.Fatal("Exit 應成功")
	}
	if p.ShowID != 0 {
		t.Fatalf("Exit 後 ShowID 應歸零，實際 %d", p.ShowID)
	}
	if sys.IsQuest(inst.ID) {
		t.Fatal("實例應已從註冊表移除")
	}
	// NPC 全清
	for _, id := range npcIDs {
		if ws.GetNpc(id) != nil {
			t.Fatalf("NPC %d 副本結束後仍存在", id)
		}
	}
}
