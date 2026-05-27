package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillPolyChangesVisiblePlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		ClassID:   61,
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45197,
		Impl:          "L1Monster",
		Name:          "poly_mob",
		X:             101,
		Y:             100,
		MapID:         900,
		HP:            100,
		MaxHP:         100,
		MP:            100,
		MaxMP:         100,
		Level:         50,
		STR:           30,
		DEX:           30,
		AtkDmg:        20,
		AtkSpeed:      200,
		SubMagicSpeed: 1000,
		Ranged:        1,
		AggroTarget:   target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcPolyMobSkill(t, s, 52)

	s.tickMonsterAI(npc)

	if target.PolyID != 52 || target.TempCharGfx != 52 {
		t.Fatalf("Yiwei type 4 mobskill 應把可見玩家變形成 poly_id=52，PolyID=%d TempCharGfx=%d", target.PolyID, target.TempCharGfx)
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("Yiwei type 4 mobskill 成功後應使用 sub_magic_speed 冷卻，want=5 got=%d", npc.AttackTimer)
	}
}

func withNpcPolyMobSkill(t *testing.T, s *NpcAISystem, polyID int32) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: group-poly
        type: 4
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 0
        change_target: 0
        range: 0
        area_width: 0
        area_height: 0
        leverage: 0
        skill_id: 0
        skill_area: 0
        gfx_id: 0
        act_id: 19
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: %d
        reuse_delay: 1000
`, polyID))
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}
	polys, err := data.LoadPolymorphTable("../../data/yaml/polymorph_list.yaml")
	if err != nil {
		t.Fatalf("載入 polymorph 測試資料失敗: %v", err)
	}
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入 skill 測試資料失敗: %v", err)
	}
	s.deps.MobSkills = mobSkills
	s.deps.Polys = polys
	s.deps.Skills = skills
	s.deps.Polymorph = NewPolymorphSystem(s.deps)
}
