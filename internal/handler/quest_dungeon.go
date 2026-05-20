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
	"fmt"
	"time"

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

// ─── 火龍窟副本（MISS-P2-016）NPC 動作 ────────────────────────────────────

// fireDragonCaveDungeonID 火龍窟副本 dungeon ID（對應 quest_dungeons.yaml id=148）。
const fireDragonCaveDungeonID int32 = 148

// fireDragonFlameSwordItemID 真死亡騎士烈炎之劍 item ID。
const fireDragonFlameSwordItemID int32 = 850

// fireDragonHansBagItemID 漢的袋子 item ID（24h 冷卻、開箱型）。
const fireDragonHansBagItemID int32 = 80001

// hansBagCooldownSeconds 漢給袋子冷卻秒數（24 小時）。
const hansBagCooldownSeconds int64 = 86400

// giveHansBag NPC 動作：威頓村 NPC 漢(46180) 發放「漢的袋子」。
// 24 小時冷卻——使用玩家欄位 NextHansBagAt 記錄下次可領時間。
//
// Java 對照：義维版 Npc_Hamo.java 的「領取道具」對話分支。
func giveHansBag(sess *net.Session, player *world.PlayerInfo, deps *Deps) {
	if sess == nil || player == nil || deps == nil {
		return
	}
	if deps.ItemCreate == nil {
		return
	}
	nowUnix := time.Now().Unix()
	if player.NextHansBagAt > nowUnix {
		// 仍在冷卻：送提示訊息（包含剩餘小時數）
		remainingHours := (player.NextHansBagAt - nowUnix + 3599) / 3600
		SendSystemMessage(sess, fmt.Sprintf("尚需等候約 %d 小時才能再次領取漢的袋子。", remainingHours))
		return
	}
	// 給 1 個漢的袋子（透過共用 ItemCreate；自動處理背包/重量上限與訊息）
	if _, ok := deps.ItemCreate.GiveItem(sess, player, fireDragonHansBagItemID, 1); !ok {
		return
	}
	player.NextHansBagAt = nowUnix + hansBagCooldownSeconds
	player.Dirty = true
}

// enterFireDragonCave NPC 動作：愛德納斯(46181) 收「冷冽的氣息」傳送進火龍窟。
// 進場條件由 quest_dungeons.yaml 的 entry 區段定義（min_level=60、消耗 80020）。
func enterFireDragonCave(_ *net.Session, player *world.PlayerInfo, deps *Deps) {
	if deps == nil || deps.QuestWorld == nil || player == nil {
		return
	}
	if player.ShowID > 0 {
		return // 已在副本中
	}
	deps.QuestWorld.Enter(player, fireDragonCaveDungeonID)
}

// giveFlameSword NPC 動作：副本內死亡騎士(46164) 發放「真死亡騎士烈炎之劍」。
// 玩家只能在副本內取得本武器；離場時由 cleanup_items=[850] 自動刪除（避免帶出）。
func giveFlameSword(sess *net.Session, player *world.PlayerInfo, deps *Deps) {
	if sess == nil || player == nil || deps == nil || deps.ItemCreate == nil {
		return
	}
	if player.ShowID <= 0 {
		return // 不在副本內：拒絕發放
	}
	if player.Inv == nil {
		return
	}
	// 避免重複領取
	if player.Inv.FindByItemID(fireDragonFlameSwordItemID) != nil {
		return
	}
	deps.ItemCreate.GiveItem(sess, player, fireDragonFlameSwordItemID, 1)
}
