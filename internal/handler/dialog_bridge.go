package handler

// dialog_bridge.go ─ dialog 套件與 handler/system 層的橋接層。
//
//   1. EffectAdapter 實作：把 dialog.EffectAdapter interface 接到 deps（ItemCreate、QuestWorld 等）
//   2. CallHandler 註冊表：YAML 用 call_handler: name 呼叫的既有 Go function
//   3. TryDispatchTalk / TryDispatchAction：npctalk.go / npcaction.go 的入口
//      若 dialog 系統能處理 → 回 true（caller 不再走舊路徑）
//
// 不在 dialog 套件內直接 import handler 的原因：避免循環依賴。

import (
	"time"

	"github.com/l1jgo/server/internal/dialog"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// ─── EffectAdapter 實作 ──────────────────────────────────────────────────────

type dialogEffectAdapter struct {
	deps *Deps
}

func newDialogEffectAdapter(deps *Deps) *dialogEffectAdapter {
	return &dialogEffectAdapter{deps: deps}
}

func (a *dialogEffectAdapter) GiveItem(sess *net.Session, p *world.PlayerInfo, itemID, count int32) bool {
	if a.deps == nil || a.deps.ItemCreate == nil {
		return false
	}
	_, ok := a.deps.ItemCreate.GiveItem(sess, p, itemID, count)
	return ok
}

func (a *dialogEffectAdapter) TakeItem(sess *net.Session, p *world.PlayerInfo, itemID, count int32) bool {
	if p == nil || p.Inv == nil {
		return false
	}
	it := p.Inv.FindByItemID(itemID)
	if it == nil || it.Count < count {
		return false
	}
	it.Count -= count
	if it.Count <= 0 {
		for i := range p.Inv.Items {
			if p.Inv.Items[i] == it {
				p.Inv.Items = append(p.Inv.Items[:i], p.Inv.Items[i+1:]...)
				break
			}
		}
	}
	p.Dirty = true
	return true
}

func (a *dialogEffectAdapter) SendSystemMessage(sess *net.Session, msg string) {
	SendSystemMessage(sess, msg)
}

func (a *dialogEffectAdapter) Teleport(p *world.PlayerInfo, mapID int16, x, y int16, heading int8) {
	// TODO: 接既有 teleport service 完整流程（含視覺特效、AOI 重算）
	if p == nil {
		return
	}
	p.X, p.Y, p.MapID, p.Heading = int32(x), int32(y), mapID, int16(heading)
	p.Dirty = true
}

func (a *dialogEffectAdapter) EnterDungeon(p *world.PlayerInfo, dungeonID int32) {
	if a.deps == nil || a.deps.QuestWorld == nil || p == nil {
		return
	}
	// 不再擋 ShowID > 0：副本間過場（如火龍窟 Phase 1→2）需要在已有 instance
	// 的狀態下進新副本，QuestWorld.Enter 內部會 silentExit 舊 instance。
	a.deps.QuestWorld.Enter(p, dungeonID)
}

func (a *dialogEffectAdapter) ExitDungeon(p *world.PlayerInfo) {
	if a.deps == nil || a.deps.QuestWorld == nil || p == nil || p.ShowID <= 0 {
		return
	}
	a.deps.QuestWorld.Exit(p)
}

func (a *dialogEffectAdapter) CallHandler(name string, sess *net.Session, p *world.PlayerInfo) {
	if fn, ok := dialogCallHandlers[name]; ok {
		fn(sess, p, a.deps)
	}
}

// ─── call_handler 註冊表 ───────────────────────────────────────────────────

type callHandlerFunc func(sess *net.Session, p *world.PlayerInfo, deps *Deps)

// dialogCallHandlers YAML 用 call_handler: name 時呼叫此 map 的 function（escape hatch）。
var dialogCallHandlers = map[string]callHandlerFunc{}

// RegisterDialogCallHandler 註冊一個 call_handler 名稱。
func RegisterDialogCallHandler(name string, fn callHandlerFunc) {
	dialogCallHandlers[name] = fn
}

// ─── Talk / Action 入口 ────────────────────────────────────────────────────

// TryDispatchTalk 玩家點 NPC 時的入口。回 true → 已處理；false → 沒對話定義，caller fallback。
func TryDispatchTalk(sess *net.Session, player *world.PlayerInfo, objID int32, deps *Deps) bool {
	if deps == nil || deps.Dialogs == nil || sess == nil || player == nil {
		return false
	}
	npc := deps.World.GetNpc(objID)
	if npc == nil {
		return false
	}
	reg := deps.Dialogs.Get(npc.NpcID)
	if reg == nil {
		return false
	}
	branch := deps.Dialogs.PickTalkBranch(reg, player)
	if branch == nil {
		deps.Log.Debug("dialog: no on_talk branch matched", zap.Int32("npc_id", npc.NpcID))
		return false
	}
	sendDialogByKey(sess, player, objID, reg, branch.Send, branch.Refresh, branch.Duration, deps)
	return true
}

// TryDispatchAction 玩家按按鈕時的入口。回 true → 已處理。
func TryDispatchAction(sess *net.Session, player *world.PlayerInfo, objID int32, action string, deps *Deps) bool {
	if deps == nil || deps.Dialogs == nil || sess == nil || player == nil {
		return false
	}
	npc := deps.World.GetNpc(objID)
	if npc == nil {
		return false
	}
	reg := deps.Dialogs.Get(npc.NpcID)
	if reg == nil {
		return false
	}
	def, ok := reg.OnAction[action]
	if !ok {
		return false
	}
	adapter := newDialogEffectAdapter(deps)
	res := dialog.ExecuteAction(def, sess, player, adapter)
	if res.RejectedBy >= 0 {
		return true // require 失敗 → 直接關閉
	}
	if def.ThenSend != "" {
		// 動作完成後送的對話一律靜態（不啟用 live dialog；要動態的話下次玩家再點 NPC）
		sendDialogByKey(sess, player, objID, reg, def.ThenSend, 0, 0, deps)
	}
	return true
}

// ─── 對話送出（含 live dialog 整合）──────────────────────────────────────

// sendDialogByKey 渲染指定 dialog key 並送封包；若 refresh > 0 自動掛 live dialog。
func sendDialogByKey(
	sess *net.Session, player *world.PlayerInfo, objID int32,
	reg *dialog.Registry, key string, refresh, duration time.Duration, deps *Deps,
) {
	body := renderYamlDialog(reg, key, player, objID, deps)
	if body == "" {
		return
	}
	SendDynamicHypertext(sess, objID, body)

	if refresh > 0 {
		dur := duration
		if dur <= 0 {
			dur = dialog.LiveDialogDefaultDuration
		}
		// 掛 live dialog。renderer 用 "yaml_dialog"（init 中註冊），
		// 透過 LiveDialogState 新增的 NpcID/DialogKey 欄位定位要 render 哪個對話。
		player.LiveDialog = &world.LiveDialogState{
			NpcObjID:      objID,
			RenderKey:     yamlDialogRenderKey,
			YamlNpcID:     reg.NpcID,
			YamlDialogKey: key,
			ExpiresAt:     time.Now().Unix() + int64(dur.Seconds()),
			NextRefreshAt: time.Now().Unix() + int64(refresh.Seconds()),
			IntervalSec:   int64(refresh.Seconds()),
		}
	}
}

// renderYamlDialog 內部 helper：取 template、build context、render、wrap chrome。
// 回空字串 = render 失敗，caller 應 log 但不送封包。
func renderYamlDialog(reg *dialog.Registry, key string, player *world.PlayerInfo, objID int32, deps *Deps) string {
	if reg == nil {
		return ""
	}
	tpl := reg.Dialogs[key]
	if tpl == nil {
		deps.Log.Warn("dialog: missing template",
			zap.Int32("npc_id", reg.NpcID), zap.String("key", key))
		return ""
	}
	npcName := ""
	if deps.World != nil {
		if npc := deps.World.GetNpc(objID); npc != nil {
			npcName = npc.Name
		}
	}
	ctx := dialog.BuildRenderContext(player, npcName)
	rawBody, err := tpl.Render(ctx)
	if err != nil {
		deps.Log.Warn("dialog: render failed",
			zap.Int32("npc_id", reg.NpcID), zap.String("key", key), zap.Error(err))
		return ""
	}
	return dialog.WrapBodyWithChrome(rawBody, reg.Speaker, reg.SpeakerColor)
}

// yamlDialogRenderKey LiveDialog 用的單一全域 renderer key（所有 YAML 對話共用）。
const yamlDialogRenderKey = "yaml_dialog"

// SetDialogManagerForLive 由 main.go 注入 deps 後呼叫；註冊 live dialog 的 YAML renderer。
// 拆出來而非用 init()，是因為 renderer 內部要存取 deps（global state）。
func SetDialogManagerForLive(deps *Deps) {
	RegisterLiveDialogRenderer(yamlDialogRenderKey, func(p *world.PlayerInfo) string {
		if p == nil || p.LiveDialog == nil || deps.Dialogs == nil {
			return ""
		}
		reg := deps.Dialogs.Get(p.LiveDialog.YamlNpcID)
		if reg == nil {
			return ""
		}
		return renderYamlDialog(reg, p.LiveDialog.YamlDialogKey, p, p.LiveDialog.NpcObjID, deps)
	})
}
