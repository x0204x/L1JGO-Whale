package system

// LiveDialogSystem 每秒檢查所有玩家的 LiveDialog 狀態，到期就 re-render + 重送對話封包。
// Phase 3（PostUpdate）— 在遊戲邏輯後、輸出前執行，避免狀態不一致。
//
// 對應 world/live_dialog.go 與 handler/dynamic_dialog.go 的 LiveDialogRenderers 註冊表。

import (
	"time"

	coresys "github.com/l1jgo/server/internal/core/system"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

// LiveDialogSystem 動態對話刷新系統。
type LiveDialogSystem struct {
	world *world.State
}

// NewLiveDialogSystem 建構子（無刷新間隔參數：每 tick 都檢查、由各玩家的 NextRefreshAt 決定實際送封包頻率）。
func NewLiveDialogSystem(ws *world.State) *LiveDialogSystem {
	return &LiveDialogSystem{world: ws}
}

// Phase 在 PostUpdate phase 執行（戰鬥/技能/AI 之後、輸出之前）。
func (s *LiveDialogSystem) Phase() coresys.Phase { return coresys.PhasePostUpdate }

// Update 每 tick 觸發；逐玩家檢查 live dialog 並重送。
func (s *LiveDialogSystem) Update(_ time.Duration) {
	if s.world == nil {
		return
	}
	now := time.Now().Unix()
	s.world.AllPlayers(func(p *world.PlayerInfo) {
		if p == nil || p.LiveDialog == nil || p.Session == nil {
			return
		}
		ld := p.LiveDialog
		// 超過 ExpiresAt → 直接終止（不發終止訊息，玩家會看到對話停止刷新但仍可關閉）
		if now >= ld.ExpiresAt {
			p.LiveDialog = nil
			return
		}
		if now < ld.NextRefreshAt {
			return
		}
		// 查 renderer 並重新組對話封包
		body := handler.RenderLiveDialog(ld.RenderKey, p)
		if body == "" {
			// renderer 不存在或回空 → 視為終止
			p.LiveDialog = nil
			return
		}
		handler.SendDynamicHypertext(p.Session, ld.NpcObjID, body)
		ld.NextRefreshAt = now + ld.IntervalSec
	})
}
