package world

// 動態 HTML 對話即時更新狀態（搭配客戶端 @dynamic patch）。
//
// 使用情境：對話內容包含倒數計時、即時統計、副本剩餘時間等需要持續變動的資訊。
// 設計原則：
//   - 純資料結構：renderer 不放在這裡（避免 closure 引用導致 GC 問題），改用 RenderKey
//     字串到 handler 層的 renderer 註冊表查找
//   - 玩家最多 1 個 active live dialog（同時點兩個 NPC 會自動取代）
//   - 任何 NPC 動作都會清空（包含 cancel/close/實際按鈕）
//   - ExpiresAt 強制超時：避免玩家放著不管導致無限 re-render
//   - 不持久化：玩家斷線即清除（OnDisconnect 內清空）
//
// 對應 handler/dynamic_dialog.go 的 StartLiveDialog + LiveDialogRenderers 註冊表。

// LiveDialogState 單一玩家的動態對話狀態。
type LiveDialogState struct {
	// NpcObjID 對話對象 NPC 的 world object ID（C_HACTION 回送時的 objID 標的）。
	NpcObjID int32

	// RenderKey 在 handler 層 LiveDialogRenderers 註冊表中的 key。
	// - 程式碼 hardcoded renderer：用個別 key 如 "han_cooldown"
	// - YAML 對話：固定用 "yaml_dialog"，由 YamlNpcID + YamlDialogKey 定位實際對話
	RenderKey string

	// YamlNpcID / YamlDialogKey 僅在 RenderKey == "yaml_dialog" 時使用：
	// 指明要 render dialog.Manager 註冊表中哪個 NPC 的哪個對話。
	YamlNpcID     int32
	YamlDialogKey string

	// ExpiresAt 此 live dialog 強制終止的 Unix 秒數（防止永久 re-render）。
	ExpiresAt int64

	// NextRefreshAt 下次重新 render 並送封包的 Unix 秒數。
	NextRefreshAt int64

	// IntervalSec 刷新間隔秒數（通常 1）。
	IntervalSec int64
}
