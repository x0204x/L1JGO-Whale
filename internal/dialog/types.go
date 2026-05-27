// Package dialog 動態 HTML 對話資料層 + 引擎（搭配客戶端 @dynamic patch）。
//
// 設計目標：對話內容 = 100% 純 HTML 檔；路由邏輯 = 極簡 YAML。寫對話像寫網頁。
//
// 目錄結構：
//   server/data/dialogs/
//     46180_han/                NPC 編號 + 助記名（資料夾名取首段數字作 npc_id）
//       _routes.yaml            路由與動作（很短）
//       has_bag.htm             對話 1：可領取（純 HTML）
//       low_level.htm           對話 2：等級不足
//       cooldown.htm            對話 3：冷卻中（含 {{ remaining .NextHansBagAt }} 插值）
//       bag_given.htm           對話 4：給完之後
//
// _routes.yaml schema：
//   default_speaker: 漢                  # 自動套黃色 + ":" 前綴（可選）
//   default_speaker_color: ffcc00        # 預設黃，可省略
//   on_talk:                             # 玩家點 NPC 觸發；first-match
//     - when: { level_lt: 60 }       ;   send: low_level
//     - when: { cooldown_active: NextHansBagAt }
//       send: cooldown
//       refresh: 1s                      # 啟用 live dialog（每 1 秒重 render）
//       duration: 5m                     # live dialog 強制終止時間（預設 5m）
//     - send: has_bag                    # 無 when = default
//   on_action:                           # 玩家按按鈕觸發；key = action 字串
//     a:
//       require: [{ level_gte: 60 }, { cooldown_passed: NextHansBagAt }]
//       effects:
//         - give_item: { id: 80001, count: 1 }
//         - set_cooldown: { field: NextHansBagAt, seconds: 86400 }
//       then_send: bag_given             # 動作完後送的對話（省略 = 關閉）
//
// 對應引擎：condition.go（條件評估）、effect.go（動作執行）、render.go（HTML 渲染）、
// loader.go（檔案載入）、registry.go（in-memory 註冊表）。
package dialog

import (
	"text/template"
	"time"
)

// Registry 一個 NPC 的完整對話定義（路由 + 動作 + 對話內容）。
// 由 loader.go 從磁碟讀入後建構，全域可變註冊表存於 Manager 中。
type Registry struct {
	NpcID         int32                      // 從資料夾名首段數字解析
	FolderName    string                     // e.g. "46180_han"
	Speaker       string                     // 預設說話者（空字串 = NPC 不開口）
	SpeakerColor  string                     // 預設說話者顏色（hex，預設 "ffcc00"）
	OnTalk        []TalkBranch               // 玩家點 NPC 時的條件路由
	OnAction      map[string]*ActionDef      // 按鈕 action 字串 → 動作定義
	Dialogs       map[string]*DialogTemplate // 對話 key（.htm 檔名去副檔名）→ 編譯後模板
}

// TalkBranch 一條 on_talk 路由分支。
type TalkBranch struct {
	When     *Condition    // nil = default（無條件，總是匹配）
	Send     string        // 要送的對話 key（對應 Dialogs map）
	Refresh  time.Duration // 0 = 靜態；>0 = live dialog 重 render 間隔
	Duration time.Duration // live dialog 強制終止；0 = 用 LiveDialogDefaultDuration
}

// ActionDef 一個按鈕動作的完整定義。
type ActionDef struct {
	Require  []Condition // 動作執行前再次檢查；任一失敗 → 拒絕 + 關閉
	Effects  []Effect    // 依序執行
	ThenSend string      // 動作完成後送的對話 key（空字串 = 關閉對話）
}

// Condition 一個條件子句。每個 condition 物件只會啟用一個欄位（取決於 YAML 鍵）。
// 條件清單見 condition.go 的 evaluate。
type Condition struct {
	LevelLt          *int16          `yaml:"level_lt,omitempty"`
	LevelGte         *int16          `yaml:"level_gte,omitempty"`
	CooldownActive   string          `yaml:"cooldown_active,omitempty"` // 欄位名
	CooldownPassed   string          `yaml:"cooldown_passed,omitempty"`
	HasItem          *ItemRef        `yaml:"has_item,omitempty"`
	LacksItem        *ItemRef        `yaml:"lacks_item,omitempty"`
	InDungeon        *int32          `yaml:"in_dungeon,omitempty"`
	NotInDungeon     bool            `yaml:"not_in_dungeon,omitempty"`
	ClassMaskHas     *int32          `yaml:"class_mask_has,omitempty"`
}

// Effect 一個動作效果。每個 effect 物件只啟用一個欄位（YAML 鍵決定）。
type Effect struct {
	GiveItem      *ItemRef       `yaml:"give_item,omitempty"`
	TakeItem      *ItemRef       `yaml:"take_item,omitempty"`
	SetCooldown   *CooldownSet   `yaml:"set_cooldown,omitempty"` // {field, seconds}
	SystemMessage string         `yaml:"system_message,omitempty"`
	CallHandler   string         `yaml:"call_handler,omitempty"` // 呼叫既有 Go function（escape hatch）
	Teleport      *TeleportDest  `yaml:"teleport,omitempty"`
	EnterDungeon  *int32         `yaml:"enter_dungeon,omitempty"`
	ExitDungeon   bool           `yaml:"exit_dungeon,omitempty"`
}

// ItemRef 物品引用（id + 數量）。
type ItemRef struct {
	ID    int32 `yaml:"id"`
	Count int32 `yaml:"count,omitempty"` // 預設 1
}

// CooldownSet 設定玩家欄位為「現在 + seconds」。
type CooldownSet struct {
	Field   string `yaml:"field"`   // PlayerInfo 欄位名（白名單見 condition.go）
	Seconds int64  `yaml:"seconds"` // 從現在起 N 秒
}

// TeleportDest 傳送目的地。
type TeleportDest struct {
	MapID   int16 `yaml:"map"`
	X       int16 `yaml:"x"`
	Y       int16 `yaml:"y"`
	Heading int8  `yaml:"heading,omitempty"`
}

// DialogTemplate 一個對話模板（.htm 檔內容 + 編譯後的 Go template）。
type DialogTemplate struct {
	Key      string             // 檔名去副檔名（"has_bag"）
	RawHTML  string             // 原始 .htm 檔內容（normalize 後）
	Compiled *template.Template // 編譯後的 Go text/template
}

// 載入器層的中介結構（_routes.yaml 直接 unmarshal）— loader.go 用。

type rawRoute struct {
	DefaultSpeaker      string                    `yaml:"default_speaker,omitempty"`
	DefaultSpeakerColor string                    `yaml:"default_speaker_color,omitempty"`
	OnTalk              []rawTalkBranch           `yaml:"on_talk,omitempty"`
	OnAction            map[string]*rawActionDef  `yaml:"on_action,omitempty"`
}

type rawTalkBranch struct {
	When     *Condition `yaml:"when,omitempty"`
	Send     string     `yaml:"send"`
	Refresh  string     `yaml:"refresh,omitempty"`  // "1s" / "500ms"
	Duration string     `yaml:"duration,omitempty"` // "5m" / "30s"
}

type rawActionDef struct {
	Require  []Condition `yaml:"require,omitempty"`
	Effects  []Effect    `yaml:"effects,omitempty"`
	ThenSend string      `yaml:"then_send,omitempty"`
}

// LiveDialogDefaultDuration 預設 live dialog 終止時間（5 分鐘）。
const LiveDialogDefaultDuration = 5 * time.Minute
