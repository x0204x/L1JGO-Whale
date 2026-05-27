package data

// Stage A.5 驗證：loader 對 empty / 最小 / 完整三種 YAML 都能正確解析。

import (
	"os"
	"path/filepath"
	"testing"
)

// 寫一個臨時 YAML 檔並回傳路徑。
func writeYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "quest_dungeons.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}
	return path
}

// TestLoadDungeonTableEmpty 空清單必須被接受（loader 在無副本時不能爆）。
func TestLoadDungeonTableEmpty(t *testing.T) {
	path := writeYAML(t, "dungeons: []\n")
	tbl, err := LoadDungeonTable(path)
	if err != nil {
		t.Fatalf("空清單應可載入: %v", err)
	}
	if tbl.Count() != 0 {
		t.Fatalf("Count 應為 0，實際 %d", tbl.Count())
	}
	if got := tbl.Get(2600); got != nil {
		t.Fatalf("Get 不存在 ID 應回 nil，得到 %+v", got)
	}
	if tbl.IsDungeonMap(2600) {
		t.Fatalf("空清單 IsDungeonMap 應為 false")
	}
}

// TestLoadDungeonTableMinimal 最小宣告（只有必要欄位）。
func TestLoadDungeonTableMinimal(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 2600
    name: "火龍窟副本"
    map_id: 2600
    max_users: 6
    time_limit: 1800
    rounds:
      - id: -1
        spawns:
          - { npc_id: 81234, count: 10, area: [32700, 32700, 32800, 32800] }
`)
	tbl, err := LoadDungeonTable(path)
	if err != nil {
		t.Fatalf("最小宣告應可載入: %v", err)
	}
	if tbl.Count() != 1 {
		t.Fatalf("Count 應為 1，實際 %d", tbl.Count())
	}
	d := tbl.Get(2600)
	if d == nil {
		t.Fatal("Get(2600) 應返回副本定義")
	}
	if d.Name != "火龍窟副本" || d.MapID != 2600 || d.MaxUsers != 6 || d.TimeLimit != 1800 {
		t.Fatalf("欄位解析錯誤: %+v", d)
	}
	if !tbl.IsDungeonMap(2600) {
		t.Fatal("IsDungeonMap(2600) 應為 true")
	}
	if len(d.Rounds) != 1 || d.Rounds[0].Trigger != RoundTriggerOnEnter {
		t.Fatalf("Round 預設 trigger 應為 on_enter，實際: %+v", d.Rounds[0])
	}
}

// TestLoadDungeonTableFull 完整宣告（所有欄位都填）。
func TestLoadDungeonTableFull(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 100
    name: "完整測試副本"
    map_id: 999
    max_users: 4
    time_limit: 900
    out_stop: true
    entry:
      min_level: 45
      max_level: 60
      class_mask: 127
      required_items:
        - { item_id: 47010, count: 1, consume: true }
      required_quest_step:
        - { quest_id: 50, step: 100 }
      forbidden_buffs: [78, 4000]
      teleport_to: { x: 32700, y: 32700, heading: 5 }
      reject_message: 1413
    exit:
      teleport_to: { x: 33430, y: 32814, map_id: 4, heading: 4 }
      cleanup_items: [850]
      reward_on_clear:
        - { item_id: 41001, count: 1 }
        - { exp: 50000 }
        - { adena: 1000 }
    rounds:
      - id: -1
        spawns:
          - { npc_id: 81234, count: 5, area: [32700, 32700, 32760, 32760], heading: 5 }
          - { npc_id: 81235, count: 1, fixed: [32730, 32730], heading: 5 }
      - id: 1
        trigger: on_round_clear
        spawns:
          - { npc_id: 81236, count: 1, fixed: [32730, 32730] }
      - id: 2
        trigger: on_timer
        timer: 600
        spawns:
          - { npc_id: 81237, count: 3, area: [32700, 32700, 32760, 32760] }
    hooks:
      on_enter: "dungeons/test.lua#on_enter"
      on_npc_death: "dungeons/test.lua#on_npc_death"
      on_last_mob_death: "dungeons/test.lua#on_last_mob_death"
`)
	tbl, err := LoadDungeonTable(path)
	if err != nil {
		t.Fatalf("完整宣告應可載入: %v", err)
	}
	d := tbl.Get(100)
	if d == nil {
		t.Fatal("Get(100) 應返回副本定義")
	}

	// 進場
	if d.Entry == nil {
		t.Fatal("Entry 不應為 nil")
	}
	if d.Entry.MinLevel != 45 || d.Entry.MaxLevel != 60 || d.Entry.ClassMask != 127 {
		t.Fatalf("Entry 等級/職業欄位錯誤: %+v", d.Entry)
	}
	if len(d.Entry.RequiredItems) != 1 || !d.Entry.RequiredItems[0].Consume {
		t.Fatalf("Entry 物品欄位錯誤: %+v", d.Entry.RequiredItems)
	}
	if len(d.Entry.ForbiddenBuffs) != 2 || d.Entry.RejectMessage != 1413 {
		t.Fatalf("Entry buff/拒絕訊息欄位錯誤: %+v", d.Entry)
	}

	// 結束
	if d.Exit == nil || d.Exit.TeleportTo == nil {
		t.Fatal("Exit / TeleportTo 不應為 nil")
	}
	if d.Exit.TeleportTo.MapID != 4 || len(d.Exit.CleanupItems) != 1 || d.Exit.CleanupItems[0] != 850 {
		t.Fatalf("Exit 欄位錯誤: %+v", d.Exit)
	}
	if len(d.Exit.RewardOnClear) != 3 {
		t.Fatalf("RewardOnClear 應為 3 項，實際 %d", len(d.Exit.RewardOnClear))
	}

	// Round
	if len(d.Rounds) != 3 {
		t.Fatalf("Rounds 應為 3 個，實際 %d", len(d.Rounds))
	}
	if d.Rounds[0].Trigger != RoundTriggerOnEnter {
		t.Fatalf("Round[-1] 預設 trigger 應為 on_enter")
	}
	if d.Rounds[1].Trigger != RoundTriggerOnRoundClear {
		t.Fatalf("Round[1] trigger 應為 on_round_clear")
	}
	if d.Rounds[2].Trigger != RoundTriggerOnTimer || d.Rounds[2].Timer != 600 {
		t.Fatalf("Round[2] trigger/timer 錯誤: %+v", d.Rounds[2])
	}

	// Hooks
	if d.Hooks == nil || d.Hooks.OnEnter == "" || d.Hooks.OnNpcDeath == "" {
		t.Fatalf("Hooks 欄位錯誤: %+v", d.Hooks)
	}
}

// TestLoadDungeonTableRejectsDuplicateID 重複 ID 應被拒絕。
func TestLoadDungeonTableRejectsDuplicateID(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 100
    name: "A"
    map_id: 100
    rounds:
      - id: -1
        spawns:
          - { npc_id: 1, count: 1, fixed: [100, 100] }
  - id: 100
    name: "B"
    map_id: 101
    rounds:
      - id: -1
        spawns:
          - { npc_id: 1, count: 1, fixed: [100, 100] }
`)
	if _, err := LoadDungeonTable(path); err == nil {
		t.Fatal("重複 ID 應拒絕")
	}
}

// TestLoadDungeonTableRejectsBadClassMask class_mask > 127 應拒絕（3.80C 無戰士）。
func TestLoadDungeonTableRejectsBadClassMask(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 1
    name: "test"
    map_id: 1
    entry:
      class_mask: 255
`)
	if _, err := LoadDungeonTable(path); err == nil {
		t.Fatal("class_mask=255 應拒絕（3.80C 無戰士，上限 127）")
	}
}

// TestLoadDungeonTableRejectsOnTimerWithoutTimer trigger=on_timer 必須有 timer。
func TestLoadDungeonTableRejectsOnTimerWithoutTimer(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 1
    name: "test"
    map_id: 1
    rounds:
      - id: 1
        trigger: on_timer
        spawns:
          - { npc_id: 1, count: 1, fixed: [100, 100] }
`)
	if _, err := LoadDungeonTable(path); err == nil {
		t.Fatal("trigger=on_timer 無 timer 應拒絕")
	}
}

// TestLoadDungeonTableRejectsSpawnWithoutPosition spawn 沒指定 area/fixed 應拒絕。
func TestLoadDungeonTableRejectsSpawnWithoutPosition(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 1
    name: "test"
    map_id: 1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 1, count: 1 }
`)
	if _, err := LoadDungeonTable(path); err == nil {
		t.Fatal("spawn 無 area/fixed 應拒絕")
	}
}

// TestLoadDungeonTableRejectsBadHookFormat hook 格式錯誤應拒絕。
func TestLoadDungeonTableRejectsBadHookFormat(t *testing.T) {
	path := writeYAML(t, `
dungeons:
  - id: 1
    name: "test"
    map_id: 1
    rounds:
      - id: -1
        spawns:
          - { npc_id: 1, count: 1, fixed: [100, 100] }
    hooks:
      on_enter: "missing_hash"
`)
	if _, err := LoadDungeonTable(path); err == nil {
		t.Fatal("hook 格式錯誤應拒絕")
	}
}
