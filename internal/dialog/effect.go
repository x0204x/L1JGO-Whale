package dialog

import (
	"time"

	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

// EffectAdapter 由呼叫方（handler 層）注入的外部介面，供 effect 執行器呼叫遊戲邏輯。
// 把所有 game state 變更收斂在這個 interface，dialog 套件不直接依賴 handler / system，
// 避免循環 import。
type EffectAdapter interface {
	// GiveItem 給玩家物品（含背包/重量檢查、訊息）。回傳是否成功。
	GiveItem(sess *net.Session, p *world.PlayerInfo, itemID, count int32) bool

	// TakeItem 從玩家背包扣物品（不夠回 false，扣完發訊息）。
	TakeItem(sess *net.Session, p *world.PlayerInfo, itemID, count int32) bool

	// SendSystemMessage 發系統訊息給玩家。
	SendSystemMessage(sess *net.Session, msg string)

	// Teleport 傳送玩家到指定座標。
	Teleport(p *world.PlayerInfo, mapID int16, x, y int16, heading int8)

	// EnterDungeon 玩家進入副本（觸發 QuestWorld.Enter）。
	EnterDungeon(p *world.PlayerInfo, dungeonID int32)

	// ExitDungeon 玩家離開副本（觸發 QuestWorld.Exit）。
	ExitDungeon(p *world.PlayerInfo)

	// CallHandler 呼叫既有 Go function 作為 escape hatch。
	// handler 層維護一個 map[string]func(sess, player) 的註冊表。
	CallHandler(name string, sess *net.Session, p *world.PlayerInfo)
}

// ExecResult 一個 effect 鏈執行後的結果。
type ExecResult struct {
	// RejectedBy 若 require 條件不過，這裡帶失敗的 Condition 索引（< 0 = 沒被拒絕）。
	RejectedBy int

	// CloseDialog true = 全部執行完且應該關閉對話。
	// false = 還有後續對話要送（透過 then_send，由 caller 處理）。
	CloseDialog bool
}

// ExecuteAction 評估 require → 執行 effects → 設定 close flag。
// 不負責發 then_send 對話；caller 從 ActionDef.ThenSend 取值。
func ExecuteAction(
	def *ActionDef, sess *net.Session, p *world.PlayerInfo, adapter EffectAdapter,
) ExecResult {
	res := ExecResult{RejectedBy: -1}
	if def == nil || p == nil {
		res.CloseDialog = true
		return res
	}
	// require 檢查：第一個 fail 就 reject
	for i := range def.Require {
		if !EvalCondition(&def.Require[i], p) {
			res.RejectedBy = i
			res.CloseDialog = true
			return res
		}
	}
	// effects 依序執行
	for _, e := range def.Effects {
		applyEffect(&e, sess, p, adapter)
	}
	// CloseDialog 預設 true；若 ThenSend 不為空，caller 會送下一個對話而非關閉
	res.CloseDialog = (def.ThenSend == "")
	return res
}

// applyEffect 派發單一 effect 到對應的 adapter 方法。
func applyEffect(e *Effect, sess *net.Session, p *world.PlayerInfo, adapter EffectAdapter) {
	if e == nil || p == nil {
		return
	}
	switch {
	case e.GiveItem != nil:
		count := e.GiveItem.Count
		if count <= 0 {
			count = 1
		}
		if adapter != nil {
			adapter.GiveItem(sess, p, e.GiveItem.ID, count)
		}
	case e.TakeItem != nil:
		count := e.TakeItem.Count
		if count <= 0 {
			count = 1
		}
		if adapter != nil {
			adapter.TakeItem(sess, p, e.TakeItem.ID, count)
		}
	case e.SetCooldown != nil:
		writePlayerInt64Field(p, e.SetCooldown.Field, time.Now().Unix()+e.SetCooldown.Seconds)
		p.Dirty = true
	case e.SystemMessage != "":
		if adapter != nil {
			adapter.SendSystemMessage(sess, e.SystemMessage)
		}
	case e.CallHandler != "":
		if adapter != nil {
			adapter.CallHandler(e.CallHandler, sess, p)
		}
	case e.Teleport != nil:
		if adapter != nil {
			adapter.Teleport(p, e.Teleport.MapID, e.Teleport.X, e.Teleport.Y, e.Teleport.Heading)
		}
	case e.EnterDungeon != nil:
		if adapter != nil {
			adapter.EnterDungeon(p, *e.EnterDungeon)
		}
	case e.ExitDungeon:
		if adapter != nil {
			adapter.ExitDungeon(p)
		}
	}
}
