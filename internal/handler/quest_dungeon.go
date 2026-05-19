package handler

// 副本框架 demo NPC 動作（MISS-P0-003 Stage E）
//
// 提供 "enter_demo_dungeon" / "exit_demo_dungeon" 兩個 NPC 動作，
// 用於驗證 QuestWorldSystem 的入場與離場流程。實作層只做薄封裝，
// 真正的副本邏輯在 system/quest_world.go。
//
// 後續真實副本（火龍窟、屠龍等）會新增各自的 NPC 動作字串，
// 但都應該循這個樣板：解析參數 → 呼叫 QuestWorld.Enter/Exit/Join。

import (
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

// demoDungeonID 副本框架驗證 demo 的 dungeon ID（對應 quest_dungeons.yaml id=1001）。
const demoDungeonID int32 = 1001

// enterDemoDungeon 玩家觸發 NPC 動作 "enter_demo_dungeon"。
// 對齊 Java 風格：handler 只做委派，所有狀態變更在 system 層。
func enterDemoDungeon(_ *net.Session, player *world.PlayerInfo, deps *Deps) {
	if deps.QuestWorld == nil || player == nil {
		return
	}
	if player.ShowID > 0 {
		return // 已在副本中：避免一人多副本
	}
	deps.QuestWorld.Enter(player, demoDungeonID)
}

// exitDemoDungeon 玩家觸發 NPC 動作 "exit_demo_dungeon"。
func exitDemoDungeon(_ *net.Session, player *world.PlayerInfo, deps *Deps) {
	if deps.QuestWorld == nil || player == nil {
		return
	}
	if player.ShowID <= 0 {
		return
	}
	deps.QuestWorld.Exit(player)
}
