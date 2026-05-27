package world

// MISS-P0-003 Stage B.4：ShowID AOI 隔離測試。
// 驗證：
//   - 主世界（ShowID=0）對主世界（ShowID=0）應互相可見 ← Java 對齊 L1World 預設可見性
//   - 主世界（ShowID=0）對副本（ShowID>0）應互不可見 ← Java set_showId 隔離核心語義
//   - 同副本（ShowID=N）對同副本（ShowID=N）應互相可見 ← 同實例共存
//   - 不同副本（ShowID=N vs M）應互不可見 ← 多實例隔離
//   - 既有 GetNearbyPlayers/GetNearbyNpcs 不過濾 ShowID（向下相容）

import (
	"testing"

	"github.com/l1jgo/server/internal/net"
)

// makePlayer 建立測試玩家 — 預設 ShowID=0（主世界）。
func makePlayer(sessID uint64, charID int32, x, y int32, mapID int16, showID int32) *PlayerInfo {
	return &PlayerInfo{
		SessionID: sessID,
		Session:   &net.Session{},
		CharID:    charID,
		X:         x,
		Y:         y,
		MapID:     mapID,
		ShowID:    showID,
	}
}

// makeNpc 建立測試 NPC — 預設 ShowID=0（主世界）。
func makeNpc(id, npcID int32, x, y int32, mapID int16, showID int32) *NpcInfo {
	return &NpcInfo{
		ID:     id,
		NpcID:  npcID,
		X:      x,
		Y:      y,
		MapID:  mapID,
		ShowID: showID,
	}
}

// TestGetNearbyPlayersInShowMainWorldVisibility 主世界（0）對主世界（0）應互相可見。
func TestGetNearbyPlayersInShowMainWorldVisibility(t *testing.T) {
	ws := NewState()
	viewer := makePlayer(1, 1001, 100, 100, 4, 0)
	other := makePlayer(2, 1002, 102, 100, 4, 0)
	ws.AddPlayer(viewer)
	ws.AddPlayer(other)

	got := ws.GetNearbyPlayersInShow(viewer.X, viewer.Y, viewer.MapID, viewer.SessionID, viewer.ShowID)
	if len(got) != 1 || got[0].CharID != 1002 {
		t.Fatalf("主世界（0）對主世界（0）應可見，got=%v", got)
	}
}

// TestGetNearbyPlayersInShowDungeonIsolation 主世界（0）看不到副本（100）玩家。
func TestGetNearbyPlayersInShowDungeonIsolation(t *testing.T) {
	ws := NewState()
	mainPlayer := makePlayer(1, 1001, 100, 100, 4, 0)
	dungeonPlayer := makePlayer(2, 1002, 102, 100, 4, 100)
	ws.AddPlayer(mainPlayer)
	ws.AddPlayer(dungeonPlayer)

	// 主世界視角
	mainView := ws.GetNearbyPlayersInShow(mainPlayer.X, mainPlayer.Y, mainPlayer.MapID, mainPlayer.SessionID, mainPlayer.ShowID)
	if len(mainView) != 0 {
		t.Fatalf("主世界視角不應看見副本玩家，got=%v", mainView)
	}

	// 副本視角
	dungeonView := ws.GetNearbyPlayersInShow(dungeonPlayer.X, dungeonPlayer.Y, dungeonPlayer.MapID, dungeonPlayer.SessionID, dungeonPlayer.ShowID)
	if len(dungeonView) != 0 {
		t.Fatalf("副本視角不應看見主世界玩家，got=%v", dungeonView)
	}
}

// TestGetNearbyPlayersInShowSameDungeonVisibility 同副本（100）的兩個玩家應互相可見。
func TestGetNearbyPlayersInShowSameDungeonVisibility(t *testing.T) {
	ws := NewState()
	a := makePlayer(1, 1001, 100, 100, 4, 100)
	b := makePlayer(2, 1002, 102, 100, 4, 100)
	ws.AddPlayer(a)
	ws.AddPlayer(b)

	got := ws.GetNearbyPlayersInShow(a.X, a.Y, a.MapID, a.SessionID, a.ShowID)
	if len(got) != 1 || got[0].CharID != 1002 {
		t.Fatalf("同副本（100）應互相可見，got=%v", got)
	}
}

// TestGetNearbyPlayersInShowDifferentDungeonIsolation 不同副本實例（100 vs 101）應互不可見。
func TestGetNearbyPlayersInShowDifferentDungeonIsolation(t *testing.T) {
	ws := NewState()
	d100 := makePlayer(1, 1001, 100, 100, 4, 100)
	d101 := makePlayer(2, 1002, 102, 100, 4, 101)
	ws.AddPlayer(d100)
	ws.AddPlayer(d101)

	got := ws.GetNearbyPlayersInShow(d100.X, d100.Y, d100.MapID, d100.SessionID, d100.ShowID)
	if len(got) != 0 {
		t.Fatalf("不同副本實例應互不可見，got=%v", got)
	}
}

// TestGetNearbyNpcsInShowMainWorldVisibility 主世界 NPC 對主世界玩家可見。
func TestGetNearbyNpcsInShowMainWorldVisibility(t *testing.T) {
	ws := NewState()
	npc := makeNpc(1, 50, 100, 100, 4, 0)
	ws.AddNpc(npc)

	got := ws.GetNearbyNpcsInShow(102, 100, 4, 0)
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("主世界 NPC 對主世界玩家應可見，got=%v", got)
	}
}

// TestGetNearbyNpcsInShowDungeonIsolation 副本 NPC 不應出現在主世界。
func TestGetNearbyNpcsInShowDungeonIsolation(t *testing.T) {
	ws := NewState()
	dungeonNpc := makeNpc(1, 81234, 100, 100, 4, 100)
	ws.AddNpc(dungeonNpc)

	mainView := ws.GetNearbyNpcsInShow(102, 100, 4, 0)
	if len(mainView) != 0 {
		t.Fatalf("主世界視角不應看見副本 NPC，got=%v", mainView)
	}

	dungeonView := ws.GetNearbyNpcsInShow(102, 100, 4, 100)
	if len(dungeonView) != 1 || dungeonView[0].ID != 1 {
		t.Fatalf("副本視角應看見副本 NPC，got=%v", dungeonView)
	}
}

// TestGetNearbyNpcsInShowSkipsDead 死亡 NPC 仍應跳過（與原 GetNearbyNpcs 行為一致）。
func TestGetNearbyNpcsInShowSkipsDead(t *testing.T) {
	ws := NewState()
	npc := makeNpc(1, 50, 100, 100, 4, 0)
	npc.Dead = true
	ws.AddNpc(npc)

	got := ws.GetNearbyNpcsInShow(102, 100, 4, 0)
	if len(got) != 0 {
		t.Fatalf("死亡 NPC 不應被回傳，got=%v", got)
	}
}

// TestGetNearbyPlayersUnchanged 既有 GetNearbyPlayers 不過濾 ShowID（向下相容驗證）。
// 主世界（0）與副本（100）兩個玩家都應在結果中，因為舊方法忽略 ShowID。
func TestGetNearbyPlayersUnchangedIgnoresShowID(t *testing.T) {
	ws := NewState()
	main := makePlayer(1, 1001, 100, 100, 4, 0)
	dungeon := makePlayer(2, 1002, 102, 100, 4, 100)
	ws.AddPlayer(main)
	ws.AddPlayer(dungeon)

	// 從主世界視角查（excludeSession=1），舊方法不過濾 ShowID，應回傳副本玩家
	got := ws.GetNearbyPlayers(main.X, main.Y, main.MapID, main.SessionID)
	if len(got) != 1 || got[0].CharID != 1002 {
		t.Fatalf("既有 GetNearbyPlayers 應不過濾 ShowID（向下相容），got=%v", got)
	}
}

// TestGetNearbyNpcsUnchangedIgnoresShowID 既有 GetNearbyNpcs 不過濾 ShowID（向下相容驗證）。
func TestGetNearbyNpcsUnchangedIgnoresShowID(t *testing.T) {
	ws := NewState()
	mainNpc := makeNpc(1, 50, 100, 100, 4, 0)
	dungeonNpc := makeNpc(2, 81234, 100, 100, 4, 100)
	ws.AddNpc(mainNpc)
	ws.AddNpc(dungeonNpc)

	got := ws.GetNearbyNpcs(102, 100, 4)
	if len(got) != 2 {
		t.Fatalf("既有 GetNearbyNpcs 應回傳兩個 NPC（不過濾 ShowID，向下相容），got=%d", len(got))
	}
}

func TestGetNearbyGroundItemsInShowDungeonIsolationLikeJava(t *testing.T) {
	ws := NewState()
	mainItem := &GroundItem{ID: 3001, ItemID: 1001, X: 100, Y: 100, MapID: 4, ShowID: 0}
	dungeonItem := &GroundItem{ID: 3002, ItemID: 1001, X: 101, Y: 100, MapID: 4, ShowID: 100}
	otherDungeonItem := &GroundItem{ID: 3003, ItemID: 1001, X: 102, Y: 100, MapID: 4, ShowID: 101}
	ws.AddGroundItem(mainItem)
	ws.AddGroundItem(dungeonItem)
	ws.AddGroundItem(otherDungeonItem)

	mainView := ws.GetNearbyGroundItemsInShow(100, 100, 4, 0)
	if len(mainView) != 1 || mainView[0].ID != mainItem.ID {
		t.Fatalf("主世界玩家只應看見主世界地上物，got=%v", mainView)
	}

	dungeonView := ws.GetNearbyGroundItemsInShow(100, 100, 4, 100)
	if len(dungeonView) != 1 || dungeonView[0].ID != dungeonItem.ID {
		t.Fatalf("副本玩家只應看見同 ShowID 地上物，got=%v", dungeonView)
	}

	otherDungeonView := ws.GetNearbyGroundItemsInShow(100, 100, 4, 101)
	if len(otherDungeonView) != 1 || otherDungeonView[0].ID != otherDungeonItem.ID {
		t.Fatalf("不同副本地上物不可互相可見，got=%v", otherDungeonView)
	}
}

func TestGroundItemShowIDDefaultZero(t *testing.T) {
	g := &GroundItem{}
	if g.ShowID != 0 {
		t.Fatalf("GroundItem 預設 ShowID 應為 0（主世界），實際 %d", g.ShowID)
	}
}

func TestGetNearbyGroundEffectsInShowDungeonIsolationLikeJava(t *testing.T) {
	ws := NewState()
	mainEffect := &GroundEffect{ID: 3101, NpcID: 81169, X: 100, Y: 100, MapID: 4, ShowID: 0}
	dungeonEffect := &GroundEffect{ID: 3102, NpcID: 81169, X: 101, Y: 100, MapID: 4, ShowID: 100}
	otherDungeonEffect := &GroundEffect{ID: 3103, NpcID: 81169, X: 102, Y: 100, MapID: 4, ShowID: 101}
	ws.AddGroundEffect(mainEffect)
	ws.AddGroundEffect(dungeonEffect)
	ws.AddGroundEffect(otherDungeonEffect)

	mainView := ws.GetNearbyGroundEffectsInShow(100, 100, 4, 0)
	if len(mainView) != 1 || mainView[0].ID != mainEffect.ID {
		t.Fatalf("主世界玩家只應看見主世界地面效果，got=%v", mainView)
	}

	dungeonView := ws.GetNearbyGroundEffectsInShow(100, 100, 4, 100)
	if len(dungeonView) != 1 || dungeonView[0].ID != dungeonEffect.ID {
		t.Fatalf("副本玩家只應看見同 ShowID 地面效果，got=%v", dungeonView)
	}

	otherDungeonView := ws.GetNearbyGroundEffectsInShow(100, 100, 4, 101)
	if len(otherDungeonView) != 1 || otherDungeonView[0].ID != otherDungeonEffect.ID {
		t.Fatalf("不同副本地面效果不可互相可見，got=%v", otherDungeonView)
	}
}

func TestHasGroundEffectAtInShowAllowsSameTileAcrossDungeonLikeJava(t *testing.T) {
	ws := NewState()
	ws.AddGroundEffect(&GroundEffect{ID: 3201, NpcID: 81157, X: 100, Y: 100, MapID: 4, ShowID: 100})

	if !ws.HasGroundEffectAtInShow(100, 100, 4, 81157, 100) {
		t.Fatalf("同 ShowID 應能找到同座標地面效果")
	}
	if ws.HasGroundEffectAtInShow(100, 100, 4, 81157, 101) {
		t.Fatalf("不同 ShowID 不應被同座標地面效果阻擋")
	}
}

func TestGroundEffectShowIDDefaultZero(t *testing.T) {
	e := &GroundEffect{}
	if e.ShowID != 0 {
		t.Fatalf("GroundEffect 預設 ShowID 應為 0（主世界），實際 %d", e.ShowID)
	}
}

// TestPlayerInfoShowIDDefaultZero PlayerInfo 零值建構後 ShowID 應為 0（主世界）。
func TestPlayerInfoShowIDDefaultZero(t *testing.T) {
	p := &PlayerInfo{}
	if p.ShowID != 0 {
		t.Fatalf("PlayerInfo 預設 ShowID 應為 0（主世界），實際 %d", p.ShowID)
	}
}

// TestNpcInfoShowIDDefaultZero NpcInfo 零值建構後 ShowID 應為 0。
func TestNpcInfoShowIDDefaultZero(t *testing.T) {
	n := &NpcInfo{}
	if n.ShowID != 0 {
		t.Fatalf("NpcInfo 預設 ShowID 應為 0（主世界），實際 %d", n.ShowID)
	}
	if n.Transient {
		t.Fatal("NpcInfo 預設 Transient 應為 false（主世界 NPC 走持久化）")
	}
}
