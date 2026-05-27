package handler

// 動態 HTML 對話封包（客戶端 patch 後啟用）
//
// 路徑分流：客戶端 S_HYPERTEXT(opcode 39) handler 在 0x00527A58 檢查 htmlName[0]：
//   - 'a'~'z' / 數字等：原本邏輯 → 查本地 lineage/html/<name>.html
//   - '@'：新邏輯 → 跳過檔案查找，直接吃封包第二個 string 當 HTML body，
//          交由既有 in-memory parser 0x004944E0 渲染
//
// 封包格式（@dynamic 路徑）：
//   opcode(B=0x27) + objID(D) + "@dynamic"(S) + htmlBody(S) + argc(H) + args...(N×S)
//
// htmlBody 內容限制（由 patch 確認）：
//   - 編碼：Big5/CP950（WriteS 已自動轉碼）
//   - tag 白名單：body/a/b/i/p/left/center/right/justify/scatter/br/font/img/var
//                 name/username/time/date/may/subHtml
//   - attr 白名單：align/valign/face/Size/fg/bg/src/link/action/top/middle/bottom
//                   lgap/agap/textwidth/tooltip/NumberInput
//   - 注意：font 用 fg="ff0000" 而非 color=；attr value 必須用雙引號
//   - 禁用：\0（封包 NUL 終止會錯位）、\n / \r\n（parser 不可靠，請用 <br>）
//   - HTML body buffer 上限：client 端 0x7000 bytes（28 KB），建議 < 4KB
//
// 本檔只提供低階 send 與 live dialog 註冊機制；實際對話內容由 internal/dialog 套件
// 的 YAML+HTM 引擎驅動（見 dialog_bridge.go）。

import (
	"time"

	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

// dynamicHtmlName 客戶端 patch 識別動態 HTML 的 magic name 前綴。
const dynamicHtmlName = "@dynamic"

// SendDynamicHypertext 送出伺服器內嵌 HTML 給客戶端渲染。Exported 給 system / dialog package 使用。
// htmlBody 不要含 \0 或非 <br> 換行；參數列表用於 %0/%1 替代（與舊路徑相同）。
func SendDynamicHypertext(sess *net.Session, objID int32, htmlBody string, args ...string) {
	if sess == nil {
		return
	}
	w := packet.NewWriterWithOpcode(packet.S_OPCODE_HYPERTEXT)
	w.WriteD(objID)
	w.WriteS(dynamicHtmlName)
	w.WriteS(htmlBody)
	w.WriteH(uint16(len(args)))
	for _, a := range args {
		w.WriteS(a)
	}
	sess.Send(w.Bytes())
}

// LiveDialogRenderer 將玩家狀態渲染成 HTML body 的函式。回空字串 = 終止 live dialog。
type LiveDialogRenderer func(player *world.PlayerInfo) string

// liveDialogRenderers live dialog 的 render key → renderer 註冊表。
// 不放 closure 到 PlayerInfo 是為了避免 GC 引用環、序列化困難、熱重載 renderer 困難。
var liveDialogRenderers = map[string]LiveDialogRenderer{}

// RegisterLiveDialogRenderer 註冊一個 render key。重複註冊會覆蓋舊的。
// 必須在系統啟動時呼叫（init 或 main 中）。
func RegisterLiveDialogRenderer(key string, renderer LiveDialogRenderer) {
	liveDialogRenderers[key] = renderer
}

// RenderLiveDialog 由 system 層呼叫；查表 + 執行 renderer。
// renderer 不存在或回空字串時回傳 ""，由呼叫端負責終止 live dialog。
func RenderLiveDialog(key string, player *world.PlayerInfo) string {
	if r, ok := liveDialogRenderers[key]; ok {
		return r(player)
	}
	return ""
}

// StartLiveDialog 在玩家身上設定 live dialog 狀態並立即送出第一次對話封包。
// 給尚未遷移到 YAML 引擎的 hardcoded 對話用；YAML 對話請走 dialog_bridge.go 的
// sendDialogByKey（含 refresh 參數）。
//
//   sess        玩家連線
//   player      玩家狀態
//   npcObjID    對話對象 NPC 的 world object ID
//   renderKey   liveDialogRenderers 註冊表 key
//   intervalSec 刷新間隔秒數（建議 1；過短會浪費頻寬）
//   durationSec 強制終止前的總時長（防止永久 re-render）
func StartLiveDialog(sess *net.Session, player *world.PlayerInfo, npcObjID int32,
	renderKey string, intervalSec, durationSec int64,
) {
	if sess == nil || player == nil {
		return
	}
	body := RenderLiveDialog(renderKey, player)
	if body == "" {
		return // renderer 不存在 → 不開啟
	}
	now := time.Now().Unix()
	player.LiveDialog = &world.LiveDialogState{
		NpcObjID:      npcObjID,
		RenderKey:     renderKey,
		ExpiresAt:     now + durationSec,
		NextRefreshAt: now + intervalSec,
		IntervalSec:   intervalSec,
	}
	SendDynamicHypertext(sess, npcObjID, body)
}
