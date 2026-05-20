package system

// 任務副本世界管理器（MISS-P0-003 Stage C）
//
// Java 對照：com.lineage.server.world.WorldQuest（單例）+ L1QuestUser（單一副本實例）。
//
// 職責（Stage C 範圍）：
//   - 維護全域副本實例註冊表（key=showID）
//   - 流水號分配 NextID（從 100 起，對齊 Java WorldQuest._nextId 初值）
//   - Enter / Exit / Get / IsQuest API
//   - 時間限制 tick 計數 + 到期自動 endInstance
//   - 玩家斷線時 RemoveOnDisconnect
//
// Stage C 不做（留給 Stage D / Stage E）：
//   - Round 引擎與出生規則執行（DungeonSpawn）
//   - 進場條件 DSL 驗證（EntryRule）
//   - Lua hook 呼叫
//   - NPC MarkForDestruction（Stage D 需要先有出生才會有清理）
//   - 通關獎勵發放

import (
	"sync"
	"time"

	coresys "github.com/l1jgo/server/internal/core/system"
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// 副本流水號起始值（對應 Java WorldQuest._nextId = new AtomicInteger(100)）。
const questWorldFirstInstanceID int32 = 100

// 遊戲 tick 速率（5 ticks ≈ 1 秒，與 HauntedHouseSystem 同款）。
const questWorldTicksPerSecond int64 = 5

// QuestWorldSystem 任務副本世界管理器（Phase PostUpdate）。
// 實作 handler.QuestWorldManager 介面。
type QuestWorldSystem struct {
	deps     *handler.Deps
	ws       *world.State
	dungeons *data.DungeonTable

	mu        sync.RWMutex
	instances map[int32]*world.QuestInstance // showID → instance
	nextID    int32                          // 下一個分配的 showID（單調遞增）
	tick      int64                          // 系統累計 tick（時間限制基準）
}

// NewQuestWorldSystem 建立 QuestWorldSystem 實例。
func NewQuestWorldSystem(ws *world.State, dungeons *data.DungeonTable, deps *handler.Deps) *QuestWorldSystem {
	return &QuestWorldSystem{
		deps:      deps,
		ws:        ws,
		dungeons:  dungeons,
		instances: make(map[int32]*world.QuestInstance),
		nextID:    questWorldFirstInstanceID,
	}
}

// Phase 系統執行階段。
func (s *QuestWorldSystem) Phase() coresys.Phase { return coresys.PhasePostUpdate }

// Update 每 tick 增進系統時間並檢查：
//   1. on_timer 觸發類 round 是否到期
//   2. 時間限制副本是否過期
func (s *QuestWorldSystem) Update(_ time.Duration) {
	s.mu.Lock()
	s.tick++
	cur := s.tick

	// 收集到期副本與待觸發 on_timer round（避免在迴圈內修改 map）
	var expired []*world.QuestInstance
	type timerJob struct {
		inst  *world.QuestInstance
		round *data.DungeonRound
	}
	var timerJobs []timerJob
	for _, inst := range s.instances {
		elapsedTicks := cur - inst.StartTick

		// 時間限制檢查
		if inst.HasTimeLimit() {
			limitTicks := int64(inst.TimeLimit) * questWorldTicksPerSecond
			if elapsedTicks >= limitTicks {
				expired = append(expired, inst)
				continue
			}
		}

		// on_timer round 檢查
		if s.dungeons == nil {
			continue
		}
		def := s.dungeons.Get(inst.QuestID)
		if def == nil {
			continue
		}
		for i := range def.Rounds {
			r := &def.Rounds[i]
			if r.Trigger != data.RoundTriggerOnTimer {
				continue
			}
			fireTick := int64(r.Timer) * questWorldTicksPerSecond
			if elapsedTicks < fireTick {
				continue
			}
			if !inst.MarkRoundSpawned(r.ID) {
				continue
			}
			timerJobs = append(timerJobs, timerJob{inst: inst, round: r})
		}
	}
	s.mu.Unlock()

	for _, job := range timerJobs {
		s.spawnRound(job.inst, job.round)
	}
	for _, inst := range expired {
		s.endInstance(inst, "time_limit")
	}
}

// ─── 公開 API ────────────────────────────────────────────────────────

// NextID 取回下一個副本流水號（對齊 Java WorldQuest.nextId）。
func (s *QuestWorldSystem) NextID() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	return id
}

// Get 依 showID 取得副本實例。
func (s *QuestWorldSystem) Get(showID int32) *world.QuestInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.instances[showID]
}

// IsQuest 判斷指定 showID 是否屬於執行中副本。
// 對應 Java WorldQuest.isQuest。
func (s *QuestWorldSystem) IsQuest(showID int32) bool {
	if showID <= 0 {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.instances[showID]
	return ok
}

// Count 目前執行中副本實例數量。
func (s *QuestWorldSystem) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.instances)
}

// Enter 玩家進入指定副本任務 ID 的新實例。
// 對應 Java WorldQuest.put 的「key 不存在時建立新副本」分支。
// 注意：Stage C 不做 Round 出生與進場條件 DSL；那些屬於 Stage D/E。
// 失敗回 nil（副本定義不存在）。
func (s *QuestWorldSystem) Enter(player *world.PlayerInfo, dungeonID int32) *world.QuestInstance {
	if player == nil || s.dungeons == nil {
		return nil
	}
	def := s.dungeons.Get(dungeonID)
	if def == nil {
		s.logf("QuestWorld.Enter 失敗：副本定義不存在", zap.Int32("dungeon_id", dungeonID))
		return nil
	}

	s.mu.Lock()
	id := s.nextID
	s.nextID++
	inst := world.NewQuestInstance(id, def.ID, def.MapID, def.TimeLimit, def.OutStop, s.tick)
	inst.AddPlayer(player.CharID)
	s.instances[id] = inst
	s.mu.Unlock()

	// 設定玩家 ShowID（AOI 隔離標籤）
	player.ShowID = id

	// 進場傳送（若 Entry.TeleportTo 有指定）
	if def.Entry != nil && def.Entry.TeleportTo != nil {
		dest := def.Entry.TeleportTo
		handler.TeleportPlayer(player.Session, player, dest.X, dest.Y, def.MapID, dest.Heading, s.deps)
	}

	// on_enter 觸發類 round 立即出生
	for i := range def.Rounds {
		r := &def.Rounds[i]
		if r.Trigger != data.RoundTriggerOnEnter {
			continue
		}
		if !inst.MarkRoundSpawned(r.ID) {
			continue
		}
		s.spawnRound(inst, r)
	}

	s.logf("QuestWorld.Enter",
		zap.Int32("show_id", id),
		zap.Int32("dungeon_id", def.ID),
		zap.String("name", def.Name),
		zap.Int32("char_id", player.CharID),
	)
	return inst
}

// Join 加入指定 showID 的既有副本實例（同隊伍/血盟成員追入）。
// 與 Enter 區別：不分配新 showID，直接加入現有實例。
// 對應 Java WorldQuest.put 的「key 已存在」分支。
func (s *QuestWorldSystem) Join(player *world.PlayerInfo, showID int32) *world.QuestInstance {
	if player == nil {
		return nil
	}

	s.mu.RLock()
	inst := s.instances[showID]
	s.mu.RUnlock()
	if inst == nil {
		return nil
	}

	if !inst.AddPlayer(player.CharID) {
		return inst // 已在副本內
	}
	player.ShowID = showID

	// 用副本定義的 Entry 傳送
	if s.dungeons != nil {
		if def := s.dungeons.Get(inst.QuestID); def != nil && def.Entry != nil && def.Entry.TeleportTo != nil {
			dest := def.Entry.TeleportTo
			handler.TeleportPlayer(player.Session, player, dest.X, dest.Y, inst.MapID, dest.Heading, s.deps)
		}
	}

	s.logf("QuestWorld.Join",
		zap.Int32("show_id", showID),
		zap.Int32("char_id", player.CharID),
	)
	return inst
}

// Exit 玩家退出當前所在副本。
// 對應 Java WorldQuest.remove(key, pc) 流程：
//  1. 從副本清單移除玩家、清 ShowID
//  2. 若 outStop=true 或 size<=0 → endInstance（傳出剩餘玩家 + 清 NPC）
//
// 回傳是否實際從副本移出（false = 玩家不在任何副本中）。
func (s *QuestWorldSystem) Exit(player *world.PlayerInfo) bool {
	if player == nil || player.ShowID <= 0 {
		return false
	}
	showID := player.ShowID

	s.mu.RLock()
	inst := s.instances[showID]
	s.mu.RUnlock()
	if inst == nil {
		// ShowID 指向不存在實例 → 清 ShowID 避免殘留標籤
		player.ShowID = 0
		return false
	}

	if !inst.RemovePlayer(player.CharID) {
		return false
	}
	player.ShowID = 0

	// 副本結束條件：outStop 或無玩家
	shouldEnd := inst.OutStop || inst.PlayerCount() <= 0

	s.logf("QuestWorld.Exit",
		zap.Int32("show_id", showID),
		zap.Int32("char_id", player.CharID),
		zap.Bool("end", shouldEnd),
	)

	if shouldEnd {
		s.endInstance(inst, "exit")
	}
	return true
}

// RemoveOnDisconnect 玩家斷線清理：實作 handler.QuestWorldManager 介面。
func (s *QuestWorldSystem) RemoveOnDisconnect(player *world.PlayerInfo) {
	s.Exit(player)
}

// OnNpcDeath 副本內 NPC 死亡時觸發；實作 handler.QuestWorldManager 介面。
// 由 combat.handleNpcDeath 在 NPC 死亡末段呼叫（僅 ShowID > 0 的 NPC）。
//
// 流程：
//  1. 從 inst.npcs 移除（標示「該 NPC 不再算活著」）
//  2. 若該副本所有 NPC 已清光 → 觸發所有 on_round_clear round
func (s *QuestWorldSystem) OnNpcDeath(npc *world.NpcInfo) {
	if npc == nil || npc.ShowID <= 0 {
		return
	}

	s.mu.RLock()
	inst := s.instances[npc.ShowID]
	s.mu.RUnlock()
	if inst == nil {
		return
	}

	inst.RemoveNpc(npc.ID)

	// 全清檢查 → 觸發 on_round_clear
	if inst.NpcCount() > 0 {
		return
	}
	if s.dungeons == nil {
		return
	}
	def := s.dungeons.Get(inst.QuestID)
	if def == nil {
		return
	}

	// 只觸發「第一個」尚未出生的 on_round_clear round（依 YAML 順序）。
	// 避免多個 on_round_clear round 在單次全清時被一次全部觸發；
	// 取而代之的是讓玩家殺光第 N 區 → 第 N+1 區出生 → 殺光 → 第 N+2 區... 依序推進。
	spawned := false
	for i := range def.Rounds {
		r := &def.Rounds[i]
		if r.Trigger != data.RoundTriggerOnRoundClear {
			continue
		}
		if inst.IsRoundSpawned(r.ID) {
			continue
		}
		if !inst.MarkRoundSpawned(r.ID) {
			continue
		}
		s.spawnRound(inst, r)
		spawned = true
		break
	}

	// 若沒有任何新 round 被觸發 + 副本內已無怪 → 視為最終全清，結束副本。
	// 對應 Java L1QuestUser.endQuest 由 QuestMobExecutor.stopQuest 觸發的路徑。
	if !spawned && inst.NpcCount() == 0 {
		s.endInstance(inst, "last_mob_death")
	}
}

// ─── 內部邏輯 ────────────────────────────────────────────────────────

// endInstance 結束副本實例：清除副本 NPC + 傳出剩餘玩家 + 從註冊表移除。
// 對應 Java L1QuestUser.endQuest + removeMob 合併流程。
func (s *QuestWorldSystem) endInstance(inst *world.QuestInstance, reason string) {
	if inst == nil {
		return
	}

	// 取得副本定義以拿 Exit.TeleportTo
	var exitDest *data.DungeonExitDest
	if s.dungeons != nil {
		if def := s.dungeons.Get(inst.QuestID); def != nil && def.Exit != nil {
			exitDest = def.Exit.TeleportTo
		}
	}

	// 清除副本內所有 NPC（Transient 暫態 NPC 直接從世界移除 + 通知周圍玩家）
	s.cleanupDungeonNpcs(inst)

	// 傳出仍在副本地圖的玩家
	for _, charID := range inst.Players() {
		p := s.ws.GetByCharID(charID)
		if p == nil {
			continue
		}
		p.ShowID = 0
		if p.MapID == inst.MapID && exitDest != nil {
			handler.TeleportPlayer(p.Session, p, exitDest.X, exitDest.Y, exitDest.MapID, exitDest.Heading, s.deps)
		}
	}

	// 從註冊表移除
	s.mu.Lock()
	delete(s.instances, inst.ID)
	s.mu.Unlock()

	s.logf("QuestWorld.endInstance",
		zap.Int32("show_id", inst.ID),
		zap.Int32("dungeon_id", inst.QuestID),
		zap.String("reason", reason),
	)
}

// cleanupDungeonNpcs 移除副本內所有 NPC（含廣播 + 解除地圖封鎖）。
// 對應 Java L1QuestUser.removeAllMobs（forced cleanup）。
// 同時清理戰鬥怪 (inst.Npcs) 與輔助 NPC (inst.AuxiliaryNpcs)。
func (s *QuestWorldSystem) cleanupDungeonNpcs(inst *world.QuestInstance) {
	if inst == nil || s.ws == nil {
		return
	}
	allIDs := append([]int32{}, inst.Npcs()...)
	allIDs = append(allIDs, inst.AuxiliaryNpcs()...)
	for _, npcID := range allIDs {
		npc := s.ws.GetNpc(npcID)
		if npc == nil {
			continue
		}

		// 廣播給副本內玩家移除
		viewers := s.ws.GetNearbyPlayersInShow(npc.X, npc.Y, npc.MapID, 0, inst.ID)
		if len(viewers) > 0 {
			rmData := handler.BuildRemoveObject(npc.ID)
			handler.BroadcastToPlayers(viewers, rmData)
		}

		// 解除地圖格子封鎖
		if s.deps != nil && s.deps.MapData != nil {
			s.deps.MapData.SetImpassable(npc.MapID, npc.X, npc.Y, false)
		}

		// 從世界移除
		s.ws.RemoveNpc(npc.ID)
		inst.RemoveNpc(npc.ID)
	}
}

// logf 統一日誌輸出（避免 nil log 崩潰）。
func (s *QuestWorldSystem) logf(msg string, fields ...zap.Field) {
	if s.deps == nil || s.deps.Log == nil {
		return
	}
	s.deps.Log.Info(msg, fields...)
}
