package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillChangeTargetRandomUsesHateCandidateLikeJava(t *testing.T) {
	ws := world.NewState()
	current := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "current",
		X:         104,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	candidate := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "candidate",
		X:         105,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "random_target_caster",
		X:           103,
		Y:           100,
		MapID:       900,
		HP:          100,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: current.SessionID,
		HateList:    map[uint64]int32{candidate.SessionID: 100},
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcRandomTargetMagicMobSkill(t, s)

	s.tickMonsterAI(npc)

	if candidate.HP == candidate.MaxHP {
		t.Fatalf("Yiwei change_target=3 應該選到仇恨候選目標施法，candidate HP=%d", candidate.HP)
	}
	if current.HP != current.MaxHP {
		t.Fatalf("Yiwei change_target=3 不應直接打原目標，current HP=%d MaxHP=%d", current.HP, current.MaxHP)
	}
}

func withNpcRandomTargetMagicMobSkill(t *testing.T, s *NpcAISystem) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入技能資料失敗: %v", err)
	}

	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: random-target-magic
        type: 2
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 0
        change_target: 3
        range: 0
        area_width: 0
        area_height: 0
        leverage: 0
        skill_id: 1
        skill_area: 0
        gfx_id: 0
        act_id: 18
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: 0
`)
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}

	s.deps.Skills = skills
	s.deps.MobSkills = mobSkills
}
