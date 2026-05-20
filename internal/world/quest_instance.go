package world

// 副本實例 runtime（MISS-P0-003 Stage A）
//
// Java 對照：L1QuestUser.java —— 單一執行中副本的玩家清單、NPC 清單、時間限制、
// 分數、章節物件等執行期狀態。本檔僅定義型別與基本訪問方法，
// 全域註冊器、生命週期、Round 引擎在 Stage C QuestWorldSystem 實作。

import (
	"sync"
)

// QuestInstance 單一執行中副本的 runtime 狀態。
//
// 對應 Java L1QuestUser 的欄位映射：
//   _id          → ID
//   _questid     → QuestID
//   _mapid       → MapID
//   _userList    → players (內部欄位，由 Add/Remove 方法操作)
//   _npcList     → npcs    (內部欄位，由 AddNpc/RemoveNpc 方法操作)
//   _time        → TimeLimit + StartTick
//   _outStop     → OutStop
//   _info        → Info
//   _score       → Score
//   _ResetProcess → ResetProcess
//   _SpawnedDragon → SpawnedDragon
//
// 並發策略：副本實例由 QuestWorldSystem（單線程遊戲迴圈）統一操作，
// 但 Lua hook 可能間接讀取，因此基本讀寫仍透過 mu 保護。
type QuestInstance struct {
	ID        int32 // 副本流水號（WorldQuest.nextId）
	QuestID   int32 // 副本任務 ID（DungeonDef.ID）
	MapID     int16 // 副本地圖 ID

	StartTick int64 // 入場 tick（時間限制計算基準）
	TimeLimit int32 // 時間限制（秒，-1=不限）

	OutStop bool // 一人離開即整副本結束
	Info    bool // 是否廣播剩餘怪物訊息（預設 true，對應 Java _info）

	Score int32 // 副本分數（章節副本用）

	ResetProcess  bool // 是否進行副本重置處理（對應 Java _ResetProcess，預設 true）
	SpawnedDragon bool // 是否已出生副本龍（對應 Java _SpawnedDragon）

	// 章節副本專屬狀態（對應 Java _hardinR / _orimR）—— 預留 interface{} 槽位，
	// Stage E 各副本實作時各自決定型別。
	ChapterState interface{}

	// 自訂變數（提供給 Lua hook 透過 set_dungeon_var/get_dungeon_var 使用）
	vars map[string]interface{}

	// 內部狀態（受 mu 保護）
	mu             sync.RWMutex
	players        []int32         // CharID 列表
	npcs           []int32         // NPC InstanceID 列表（戰鬥怪，計入 round clear）
	auxNpcs        []int32         // 輔助 NPC InstanceID 列表（商人/對話，不計入 round clear，副本結束時統一回收）
	spawnedRounds  map[int32]bool  // 已觸發過出生的 round ID 集合
}

// NewQuestInstance 建立新副本實例。
// timeLimit 從 DungeonDef.TimeLimit 帶入（-1=不限），startTick 從 QuestWorldSystem 當前 tick 帶入。
func NewQuestInstance(id, questID int32, mapID int16, timeLimit int32, outStop bool, startTick int64) *QuestInstance {
	return &QuestInstance{
		ID:            id,
		QuestID:       questID,
		MapID:         mapID,
		StartTick:     startTick,
		TimeLimit:     timeLimit,
		OutStop:       outStop,
		Info:          true,
		ResetProcess:  true,
		vars:          make(map[string]interface{}),
		players:       make([]int32, 0, 4),
		npcs:          make([]int32, 0, 16),
		spawnedRounds: make(map[int32]bool),
	}
}

// MarkRoundSpawned 標記 round 已觸發過出生（避免重複出生）。
// 回傳 true 表示這次標記成功（之前未被標記）；false 表示已被標記過。
func (q *QuestInstance) MarkRoundSpawned(roundID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.spawnedRounds[roundID] {
		return false
	}
	q.spawnedRounds[roundID] = true
	return true
}

// IsRoundSpawned 查詢 round 是否已觸發過出生。
func (q *QuestInstance) IsRoundSpawned(roundID int32) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.spawnedRounds[roundID]
}

// ─── 玩家管理 ─────────────────────────────────────────────────────────

// AddPlayer 加入玩家（重複加入會略過）。
// 對應 Java L1QuestUser.add(L1PcInstance)。
func (q *QuestInstance) AddPlayer(charID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, id := range q.players {
		if id == charID {
			return false // 已存在
		}
	}
	q.players = append(q.players, charID)
	return true
}

// RemovePlayer 移除玩家；回傳是否實際移除。
// 對應 Java L1QuestUser.remove(L1PcInstance)。
func (q *QuestInstance) RemovePlayer(charID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, id := range q.players {
		if id == charID {
			q.players = append(q.players[:i], q.players[i+1:]...)
			return true
		}
	}
	return false
}

// Players 取得副本內所有玩家 CharID 的快照。
func (q *QuestInstance) Players() []int32 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]int32, len(q.players))
	copy(out, q.players)
	return out
}

// PlayerCount 副本內玩家數量。
// 對應 Java L1QuestUser.size()。
func (q *QuestInstance) PlayerCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.players)
}

// ─── NPC 管理 ────────────────────────────────────────────────────────

// AddNpc 加入 NPC（重複加入會略過）。
// 對應 Java L1QuestUser.addNpc(L1NpcInstance)。
func (q *QuestInstance) AddNpc(instanceID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, id := range q.npcs {
		if id == instanceID {
			return false
		}
	}
	q.npcs = append(q.npcs, instanceID)
	return true
}

// RemoveNpc 移除 NPC；回傳是否實際移除。
// 對應 Java L1QuestUser.removeMob(L1NpcInstance)。
func (q *QuestInstance) RemoveNpc(instanceID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, id := range q.npcs {
		if id == instanceID {
			q.npcs = append(q.npcs[:i], q.npcs[i+1:]...)
			return true
		}
	}
	return false
}

// Npcs 取得副本內所有 NPC InstanceID 的快照。
func (q *QuestInstance) Npcs() []int32 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]int32, len(q.npcs))
	copy(out, q.npcs)
	return out
}

// NpcCount 副本內 NPC 總數（含死亡未清的暫態）。
// 對應 Java L1QuestUser.npcSize()。
// 僅統計戰鬥怪（透過 AddNpc 加入的），輔助 NPC（透過 AddAuxiliaryNpc 加入的）不計入。
func (q *QuestInstance) NpcCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.npcs)
}

// AddAuxiliaryNpc 加入「不計入 round clear」的輔助 NPC（如商人/對話 NPC）。
// 同一 NPC 重複加入會略過。
func (q *QuestInstance) AddAuxiliaryNpc(instanceID int32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, id := range q.auxNpcs {
		if id == instanceID {
			return false
		}
	}
	q.auxNpcs = append(q.auxNpcs, instanceID)
	return true
}

// AuxiliaryNpcs 取得副本內所有輔助 NPC InstanceID 的快照（供副本結束時統一回收）。
func (q *QuestInstance) AuxiliaryNpcs() []int32 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]int32, len(q.auxNpcs))
	copy(out, q.auxNpcs)
	return out
}

// ─── 自訂變數（Lua hook 用） ──────────────────────────────────────

// SetVar 設定副本實例變數。
func (q *QuestInstance) SetVar(key string, value interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.vars[key] = value
}

// GetVar 讀取副本實例變數；不存在回 nil。
func (q *QuestInstance) GetVar(key string) interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.vars[key]
}

// ─── 工具方法 ────────────────────────────────────────────────────────

// HasTimeLimit 是否啟用時間限制。
// 對應 Java L1QuestUser.is_time()。
func (q *QuestInstance) HasTimeLimit() bool {
	return q.TimeLimit > 0
}
