package dialog

import (
	"github.com/l1jgo/server/internal/world"
)

// Manager 全域對話註冊表，由 main.go 在啟動時建構並注入 handler/system 層。
// 一個 NPC 最多對應一個 Registry。
type Manager struct {
	byNpcID map[int32]*Registry
}

// NewManager 建構一個空 manager。
func NewManager() *Manager {
	return &Manager{byNpcID: make(map[int32]*Registry)}
}

// SetAll 一次設定所有 NPC 的對話定義（用於啟動載入或熱重載）。
func (m *Manager) SetAll(regs map[int32]*Registry) {
	if m == nil {
		return
	}
	m.byNpcID = regs
}

// Get 回傳指定 NPC 的對話定義；不存在回 nil。
func (m *Manager) Get(npcID int32) *Registry {
	if m == nil {
		return nil
	}
	return m.byNpcID[npcID]
}

// Count 已載入的 NPC 對話數。
func (m *Manager) Count() int {
	if m == nil {
		return 0
	}
	return len(m.byNpcID)
}

// PickTalkBranch 依條件挑出 on_talk 該送的分支。
// 第一個 when 為 nil（default）或滿足條件的分支獲選；都不符回 nil。
func (m *Manager) PickTalkBranch(reg *Registry, p *world.PlayerInfo) *TalkBranch {
	if reg == nil || p == nil {
		return nil
	}
	for i := range reg.OnTalk {
		b := &reg.OnTalk[i]
		if b.When == nil || EvalCondition(b.When, p) {
			return b
		}
	}
	return nil
}

// BuildRenderContext 從 PlayerInfo 抽取 .htm 模板可見的欄位。
// 模板能讀的就是這些欄位，不會洩漏整個 PlayerInfo 結構。
func BuildRenderContext(p *world.PlayerInfo, npcName string) *RenderContext {
	if p == nil {
		return &RenderContext{}
	}
	return &RenderContext{
		Level:         p.Level,
		ClassID:       p.ClassID,
		Lawful:        p.Lawful,
		Name:          p.Name,
		NextHansBagAt: p.NextHansBagAt,
		ShowID:        p.ShowID,
		MapID:         p.MapID,
		NpcName:       npcName,
	}
}
