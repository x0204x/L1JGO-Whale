package data

// 副本宣告檔載入器（MISS-P0-003 Stage A）
//
// Java 對照：QuestMapTable + QuesttSpawnTable + 各副本 Java class 的資料合併。
//
// 本檔載入 server/data/yaml/quest_dungeons.yaml，提供 DungeonTable 給
// system/quest_world.go 查詢使用。本檔僅做資料載入與基本驗證，不含 runtime 行為。

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RoundTrigger 是 Round 觸發類型。
type RoundTrigger string

const (
	RoundTriggerOnEnter       RoundTrigger = "on_enter"        // 進場即出生（保留給 id != -1 的後續輪）
	RoundTriggerOnRoundClear  RoundTrigger = "on_round_clear"  // 上一輪 NPC 全清
	RoundTriggerOnTimer       RoundTrigger = "on_timer"        // 進場 N 秒後
	RoundTriggerOnLua         RoundTrigger = "on_lua"          // Lua hook 主動觸發
)

// DungeonDef 副本定義（單一副本的完整宣告）。
type DungeonDef struct {
	ID        int32  `yaml:"id"`         // 副本任務 ID（questid）
	Name      string `yaml:"name"`       // 中文顯示名稱
	MapID     int16  `yaml:"map_id"`     // 副本地圖 ID
	MaxUsers  int32  `yaml:"max_users"`  // 最大人數（0=不限）
	TimeLimit int32  `yaml:"time_limit"` // 時間限制秒數（-1=不限）
	OutStop   bool   `yaml:"out_stop"`   // 一人離開即整副本結束

	Entry  *DungeonEntrySpec `yaml:"entry,omitempty"`
	Exit   *DungeonExitSpec  `yaml:"exit,omitempty"`
	Rounds []DungeonRound    `yaml:"rounds,omitempty"`
	Hooks  *DungeonHookSpec  `yaml:"hooks,omitempty"`
}

// DungeonEntrySpec 進場條件（資料驅動 DSL）。
// 全部欄位皆為可選；空值代表不檢查。
type DungeonEntrySpec struct {
	MinLevel          int32                       `yaml:"min_level,omitempty"`
	MaxLevel          int32                       `yaml:"max_level,omitempty"`           // 0=不限
	ClassMask         int32                       `yaml:"class_mask,omitempty"`          // 0=不限；位元和上限 127（3.80C 無戰士）
	RequiredItems     []DungeonRequiredItem       `yaml:"required_items,omitempty"`      // 全部必須持有
	RequiredQuestStep []DungeonRequiredQuestStep  `yaml:"required_quest_step,omitempty"` // 全部必須符合
	ForbiddenBuffs    []int32                     `yaml:"forbidden_buffs,omitempty"`     // 持有任一即拒絕
	TeleportTo        *DungeonEntryDest           `yaml:"teleport_to,omitempty"`         // 進場座標（map_id 隱含為副本 map_id）
	RejectMessage     int32                       `yaml:"reject_message,omitempty"`      // 拒絕時送 S_ServerMessage(N)
}

// DungeonRequiredItem 進場需持有的物品。
type DungeonRequiredItem struct {
	ItemID  int32 `yaml:"item_id"`
	Count   int32 `yaml:"count"`
	Consume bool  `yaml:"consume,omitempty"` // true=進場消耗
}

// DungeonRequiredQuestStep 進場需符合的任務步驟。
type DungeonRequiredQuestStep struct {
	QuestID int32 `yaml:"quest_id"`
	Step    int32 `yaml:"step"`
}

// DungeonEntryDest 進場座標（map_id 由副本 MapID 決定）。
type DungeonEntryDest struct {
	X       int32 `yaml:"x"`
	Y       int32 `yaml:"y"`
	Heading int16 `yaml:"heading"`
}

// DungeonExitSpec 結束規則。
type DungeonExitSpec struct {
	TeleportTo    *DungeonExitDest   `yaml:"teleport_to,omitempty"`     // 倖存者傳送目的地
	CleanupItems  []int32            `yaml:"cleanup_items,omitempty"`   // 離場刪除的物品 ID
	RewardOnClear []DungeonReward    `yaml:"reward_on_clear,omitempty"` // 通關獎勵（最後一隻怪死亡觸發）
}

// DungeonExitDest 結束傳送目的地（含 map_id）。
type DungeonExitDest struct {
	X       int32 `yaml:"x"`
	Y       int32 `yaml:"y"`
	MapID   int16 `yaml:"map_id"`
	Heading int16 `yaml:"heading"`
}

// DungeonReward 通關獎勵單元；三種類型擇一指定（item / exp / adena）。
type DungeonReward struct {
	ItemID int32 `yaml:"item_id,omitempty"`
	Count  int32 `yaml:"count,omitempty"`
	Exp    int32 `yaml:"exp,omitempty"`
	Adena  int32 `yaml:"adena,omitempty"`
}

// DungeonRound 單一輪次的出生規則。
type DungeonRound struct {
	ID      int32           `yaml:"id"`                // -1=入場、>=0=後續
	Trigger RoundTrigger    `yaml:"trigger,omitempty"` // 預設：id=-1 時 on_enter，其他需明確指定
	Timer   int32           `yaml:"timer,omitempty"`   // trigger=on_timer 時的秒數
	Spawns  []DungeonSpawn  `yaml:"spawns"`

	// RandomPick 若為 true，spawns 不全部執行；改為從中**隨機選擇 1 條** spawn 執行（用於 Boss 三選一）。
	// 對應火龍窟「最終 Boss 隨機出現一種」需求。其他 round 不應啟用此旗標。
	RandomPick bool `yaml:"random_pick,omitempty"`
}

// DungeonSpawn 單一 NPC 出生規則。
// 位置語義：
//   - Area 4 個元素都 != 0 → 區域出生
//   - 否則使用 Fixed 2 個元素 → 固定點
type DungeonSpawn struct {
	NpcID   int32   `yaml:"npc_id"`
	Count   int32   `yaml:"count"`
	Area    []int32 `yaml:"area,omitempty"`    // [x1, y1, x2, y2]
	Fixed   []int32 `yaml:"fixed,omitempty"`   // [x, y]
	Heading int16   `yaml:"heading,omitempty"`
	GroupID int32   `yaml:"group_id,omitempty"` // 0=無隊伍

	// Auxiliary 標記非戰鬥 NPC（商人 / 任務 / 對話）：仍生成於副本內並對玩家可見，
	// 但不計入 round clear 統計（避免不可擊殺的 NPC 永遠卡住下一輪出生）。
	// 對應火龍窟「入口死亡騎士」L1Merchant 角色：發放真死亡騎士烈炎之劍但不參戰。
	Auxiliary bool `yaml:"auxiliary,omitempty"`
}

// DungeonHookSpec Lua hook 路徑（格式 "scripts/dungeons/xxx.lua#func_name"）。
type DungeonHookSpec struct {
	OnEnter         string `yaml:"on_enter,omitempty"`
	OnNpcDeath      string `yaml:"on_npc_death,omitempty"`
	OnLastMobDeath  string `yaml:"on_last_mob_death,omitempty"`
	OnPlayerDeath   string `yaml:"on_player_death,omitempty"`
	OnTimeOut       string `yaml:"on_time_out,omitempty"`
	OnExit          string `yaml:"on_exit,omitempty"`
}

// dungeonFile YAML 根結構。
type dungeonFile struct {
	Dungeons []DungeonDef `yaml:"dungeons"`
}

// DungeonTable 副本資料索引表。
type DungeonTable struct {
	byID map[int32]*DungeonDef // dungeon_id → def
	all  []*DungeonDef         // 完整列表（給遍歷用，順序與 YAML 一致）
}

// LoadDungeonTable 從 YAML 載入副本宣告檔。
// 容許 dungeons 為空清單；會檢查 ID 重複、Round trigger 合法性、出生位置合法性。
func LoadDungeonTable(path string) (*DungeonTable, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取副本宣告檔: %w", err)
	}
	var f dungeonFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("解析副本宣告檔: %w", err)
	}

	t := &DungeonTable{
		byID: make(map[int32]*DungeonDef, len(f.Dungeons)),
		all:  make([]*DungeonDef, 0, len(f.Dungeons)),
	}

	for i := range f.Dungeons {
		d := &f.Dungeons[i]
		if err := validateDungeon(d); err != nil {
			return nil, fmt.Errorf("副本 id=%d 驗證失敗: %w", d.ID, err)
		}
		if _, dup := t.byID[d.ID]; dup {
			return nil, fmt.Errorf("副本 id=%d 重複定義", d.ID)
		}
		t.byID[d.ID] = d
		t.all = append(t.all, d)
	}

	return t, nil
}

// validateDungeon 對單一副本做欄位合法性檢查。
func validateDungeon(d *DungeonDef) error {
	if d.ID <= 0 {
		return fmt.Errorf("id 必須 > 0")
	}
	if d.MapID == 0 {
		return fmt.Errorf("map_id 必須非零")
	}
	if d.TimeLimit < -1 {
		return fmt.Errorf("time_limit 非法（允許 -1 或 > 0）")
	}
	if d.Entry != nil && d.Entry.ClassMask > 127 {
		return fmt.Errorf("class_mask=%d 超出 3.80C 七職業位元和上限 127", d.Entry.ClassMask)
	}

	roundIDs := make(map[int32]struct{}, len(d.Rounds))
	for i := range d.Rounds {
		r := &d.Rounds[i]
		if _, dup := roundIDs[r.ID]; dup {
			return fmt.Errorf("round id=%d 重複", r.ID)
		}
		roundIDs[r.ID] = struct{}{}

		// 預設 trigger：id=-1 → on_enter
		if r.Trigger == "" {
			if r.ID == -1 {
				r.Trigger = RoundTriggerOnEnter
			} else {
				return fmt.Errorf("round id=%d 缺少 trigger", r.ID)
			}
		}

		switch r.Trigger {
		case RoundTriggerOnEnter, RoundTriggerOnRoundClear, RoundTriggerOnTimer, RoundTriggerOnLua:
		default:
			return fmt.Errorf("round id=%d 未知 trigger=%q", r.ID, r.Trigger)
		}

		if r.Trigger == RoundTriggerOnTimer && r.Timer <= 0 {
			return fmt.Errorf("round id=%d trigger=on_timer 需要 timer > 0", r.ID)
		}

		for j := range r.Spawns {
			s := &r.Spawns[j]
			if s.NpcID <= 0 {
				return fmt.Errorf("round id=%d spawn[%d] npc_id 必須 > 0", r.ID, j)
			}
			if s.Count <= 0 {
				return fmt.Errorf("round id=%d spawn[%d] count 必須 > 0", r.ID, j)
			}
			hasArea := len(s.Area) == 4 && (s.Area[0] != 0 || s.Area[1] != 0 || s.Area[2] != 0 || s.Area[3] != 0)
			hasFixed := len(s.Fixed) == 2 && (s.Fixed[0] != 0 || s.Fixed[1] != 0)
			if !hasArea && !hasFixed {
				return fmt.Errorf("round id=%d spawn[%d] 必須指定 area 或 fixed", r.ID, j)
			}
			if hasArea && (s.Area[0] > s.Area[2] || s.Area[1] > s.Area[3]) {
				return fmt.Errorf("round id=%d spawn[%d] area 範圍非法（左下需 <= 右上）", r.ID, j)
			}
		}
	}

	if d.Hooks != nil {
		for name, ref := range map[string]string{
			"on_enter":          d.Hooks.OnEnter,
			"on_npc_death":      d.Hooks.OnNpcDeath,
			"on_last_mob_death": d.Hooks.OnLastMobDeath,
			"on_player_death":   d.Hooks.OnPlayerDeath,
			"on_time_out":       d.Hooks.OnTimeOut,
			"on_exit":           d.Hooks.OnExit,
		} {
			if ref == "" {
				continue
			}
			if !strings.Contains(ref, "#") {
				return fmt.Errorf("hook %s=%q 格式錯誤（需 path#func_name）", name, ref)
			}
		}
	}

	return nil
}

// Get 依副本 ID 取得定義；不存在回 nil。
func (t *DungeonTable) Get(id int32) *DungeonDef {
	if t == nil {
		return nil
	}
	return t.byID[id]
}

// All 取得所有副本定義（順序與 YAML 一致）。
func (t *DungeonTable) All() []*DungeonDef {
	if t == nil {
		return nil
	}
	return t.all
}

// Count 副本總數。
func (t *DungeonTable) Count() int {
	if t == nil {
		return 0
	}
	return len(t.all)
}

// IsDungeonMap 判斷指定 map_id 是否屬於某個副本（對應 Java QuestMapTable.isQuestMap）。
func (t *DungeonTable) IsDungeonMap(mapID int16) bool {
	if t == nil {
		return false
	}
	for _, d := range t.all {
		if d.MapID == mapID {
			return true
		}
	}
	return false
}
