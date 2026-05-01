package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillResurrectRestrictionNpcTemplateLoadsCantResurrect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "npc_list.yaml")
	err := os.WriteFile(path, []byte(`npcs:
  - npc_id: 45000
    name: no_rez
    impl: L1Monster
    hp: 100
    cant_resurrect: true
`), 0o600)
	if err != nil {
		t.Fatalf("建立 NPC 測試資料失敗: %v", err)
	}

	table, err := data.LoadNpcTable(path)
	if err != nil {
		t.Fatalf("讀取 NPC 測試資料失敗: %v", err)
	}
	tmpl := table.Get(45000)
	if tmpl == nil || !tmpl.CantResurrect {
		t.Fatalf("NPC YAML 應載入 cant_resurrect=true，tmpl=%+v", tmpl)
	}
}

func TestSkillResurrectRestrictionCallOfNatureRejectsCantResurrectNpc(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "elf",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45000,
		Impl:          "L1Monster",
		Name:          "no_rez",
		X:             101,
		Y:             100,
		MapID:         4,
		MaxHP:         150,
		CantResurrect: true,
	}
	ws.AddNpc(npc)
	npc.Dead = true
	npc.HP = 0
	ws.NpcDied(npc)
	s := newSkillTestSystem(t, ws)

	s.executeResurrection(caster.Session, caster, &data.SkillInfo{SkillID: 165, ActionID: 19, CastGfx: 2245}, npc.ID)

	if !npc.Dead || npc.HP != 0 {
		t.Fatalf("cant_resurrect NPC 不應被自然呼喚復活，Dead=%v HP=%d", npc.Dead, npc.HP)
	}
}
