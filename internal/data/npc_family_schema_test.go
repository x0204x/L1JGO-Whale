package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNpcTableFamilyAgroFamilyLikeJava(t *testing.T) {
	path := filepath.Join(t.TempDir(), "npc_list.yaml")
	yaml := []byte(`npcs:
  - npc_id: 45001
    name: orc
    family: orc
    agro_family: 1
  - npc_id: 45002
    name: support
    agro_family: 2
`)
	if err := os.WriteFile(path, yaml, 0o600); err != nil {
		t.Fatalf("寫入測試 YAML 失敗：%v", err)
	}

	table, err := LoadNpcTable(path)
	if err != nil {
		t.Fatalf("LoadNpcTable() error = %v", err)
	}

	sameFamily := table.Get(45001)
	if sameFamily == nil {
		t.Fatalf("找不到 npc 45001")
	}
	if sameFamily.Family != "orc" || sameFamily.AgroFamily != 1 {
		t.Fatalf("family/agro_family 載入錯誤：family=%q agro_family=%d", sameFamily.Family, sameFamily.AgroFamily)
	}
	globalFamily := table.Get(45002)
	if globalFamily == nil {
		t.Fatalf("找不到 npc 45002")
	}
	if globalFamily.Family != "" || globalFamily.AgroFamily != 2 {
		t.Fatalf("agro_family>1 不需 family 也應保留：family=%q agro_family=%d", globalFamily.Family, globalFamily.AgroFamily)
	}
}
