package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResolventTableLoadsCrystalCounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "resolvent_list.yaml")
	raw := []byte(`items:
  - item_id: 40010
    note: 治癒藥水
    crystal_count: 1
  - item_id: 31
    note: 長劍
    crystal_count: 60
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}

	table, err := LoadResolventTable(path)
	if err != nil {
		t.Fatalf("載入溶解表失敗: %v", err)
	}

	if table.Count() != 2 {
		t.Fatalf("溶解表筆數錯誤，got=%d want=2", table.Count())
	}
	if got := table.CrystalCount(31); got != 60 {
		t.Fatalf("長劍結晶數錯誤，got=%d want=60", got)
	}
	if got := table.CrystalCount(999999); got != 0 {
		t.Fatalf("未知物品應回傳 0，got=%d", got)
	}
}
