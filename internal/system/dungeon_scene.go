package system

// 副本情境劇本（DungeonScene）執行邏輯
//
// 框架：data/yaml/quest_dungeons.yaml 的 `scenes:` 區段定義一系列 trigger → lines
// 對應關係，每行 line 帶 delay/speaker/text。QuestWorldSystem 在以下時機呼叫
// triggerDungeonScenes 把符合條件的 scene 排入 QuestInstance.sceneJobs 佇列：
//   - Enter()              → SceneTriggerOnEnter
//   - Update() on_timer    → SceneTriggerOnTimer
//   - OnNpcDeath round 全清 → SceneTriggerOnRoundClear
//
// QuestWorldSystem.Update 每 tick 對所有 instance 呼叫 tickDungeonScenes，
// 把到時的 job 透過 BuildSceneLineFromNpc / BuildSceneLineFromPlayer 廣播
// 給副本內玩家（type=0 浮空文字、不入聊天歷史）。
//
// 同一 scene.ID 在同個 instance 只播放一次（playedScenes 記錄）。

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

// triggerDungeonScenes 把符合 trigger（與選擇性的 roundID）條件的 scene 排入佇列。
//   - playerCharID：speaker=player 行的廣播主體（一般是觸發劇本的玩家）。
//     入場時是進場的玩家；on_round_clear / on_timer 時可選擇副本內第一個玩家或留 0。
func (s *QuestWorldSystem) triggerDungeonScenes(
	inst *world.QuestInstance,
	def *data.DungeonDef,
	trigger data.SceneTrigger,
	roundID int32,
	playerCharID int32,
) {
	if inst == nil || def == nil || len(def.Scenes) == 0 {
		return
	}
	// 取現在 tick 作為基準
	s.mu.RLock()
	now := s.tick
	s.mu.RUnlock()

	var allJobs []world.SceneJob
	for i := range def.Scenes {
		sc := &def.Scenes[i]
		if sc.Trigger != trigger {
			continue
		}
		// on_round_clear：若 scene 指定了 round，需 match
		if trigger == data.SceneTriggerOnRoundClear && sc.Round != 0 && sc.Round != roundID {
			continue
		}
		// 同 scene.ID 只播一次
		sceneKey := sc.ID
		if sceneKey == "" {
			// 沒 ID → 用 trigger+round 當 key（避免無限重播）
			sceneKey = scenePseudoKey(trigger, roundID, i)
		}
		if !inst.MarkScenePlayed(sceneKey) {
			continue
		}

		anchorPlayer := playerCharID
		if anchorPlayer == 0 {
			players := inst.Players()
			if len(players) > 0 {
				anchorPlayer = players[0]
			}
		}

		for j := range sc.Lines {
			ln := &sc.Lines[j]
			delayTicks := msToQuestWorldTicks(ln.Delay)
			allJobs = append(allJobs, world.SceneJob{
				FireTick:     now + delayTicks,
				PlayerCharID: anchorPlayer,
				AnchorNpcID:  sc.AnchorNpc,
				Speaker:      string(ln.Speaker),
				Text:         ln.Text,
			})
		}
	}

	inst.EnqueueSceneJobs(allJobs)
}

// tickDungeonScenes 取出到時 job 並廣播給副本內玩家。
func (s *QuestWorldSystem) tickDungeonScenes(inst *world.QuestInstance, now int64) {
	if inst == nil || !inst.HasPendingSceneJobs() {
		return
	}
	fired := inst.PopFiredSceneJobs(now)
	for _, j := range fired {
		s.broadcastSceneJob(inst, j)
	}
}

// broadcastSceneJob 廣播單一台詞給副本內所有玩家。
func (s *QuestWorldSystem) broadcastSceneJob(inst *world.QuestInstance, job world.SceneJob) {
	if s.ws == nil {
		return
	}

	var pkt []byte
	switch job.Speaker {
	case string(data.SceneSpeakerNpc):
		// 找副本內第一個符合 AnchorNpcID 的 NPC instance
		var anchor *world.NpcInfo
		for _, instanceID := range inst.Npcs() {
			n := s.ws.GetNpc(instanceID)
			if n != nil && n.NpcID == job.AnchorNpcID {
				anchor = n
				break
			}
		}
		if anchor == nil {
			// 戰鬥怪沒找到 → 找輔助 NPC
			for _, instanceID := range inst.AuxiliaryNpcs() {
				n := s.ws.GetNpc(instanceID)
				if n != nil && n.NpcID == job.AnchorNpcID {
					anchor = n
					break
				}
			}
		}
		if anchor == nil {
			s.logf("scene: anchor NPC 找不到，跳過此 line")
			return
		}
		prefix := anchor.NameID
		if prefix == "" {
			prefix = anchor.Name
		}
		pkt = handler.BuildSceneLineFromNpc(anchor.ID, prefix, job.Text)

	case string(data.SceneSpeakerPlayer):
		p := s.ws.GetByCharID(job.PlayerCharID)
		if p == nil {
			return
		}
		pkt = handler.BuildSceneLineFromPlayer(p.CharID, p.Name, job.Text)

	default:
		return
	}

	if pkt == nil {
		return
	}
	// 廣播給副本內所有玩家（用 ShowID 過濾，副本內可見）
	for _, charID := range inst.Players() {
		p := s.ws.GetByCharID(charID)
		if p == nil || p.Session == nil {
			continue
		}
		p.Session.Send(pkt)
	}
}

// scenePseudoKey 沒指定 scene.ID 時的退路 key。
func scenePseudoKey(trigger data.SceneTrigger, roundID int32, idx int) string {
	return string(trigger) + "_" + itoa(int(roundID)) + "_" + itoa(idx)
}

// msToQuestWorldTicks 把 ms 換算成 quest world 系統的 tick 數（5 ticks ≈ 1 秒）。
func msToQuestWorldTicks(ms int32) int64 {
	if ms <= 0 {
		return 0
	}
	// ticks = ms * 5 / 1000；至少 1 tick 避免 0 ms 被當成同一 tick 全部一起播
	t := int64(ms) * questWorldTicksPerSecond / 1000
	if t < 1 {
		t = 1
	}
	return t
}

// itoa 簡易 int → string（避免 fmt.Sprintf）。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
