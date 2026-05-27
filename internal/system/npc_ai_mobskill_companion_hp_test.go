package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillTriggerCompanionHpTargetsSameFamilyNpcLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         104,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "companion_healer",
		Family:      "lizardman",
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
		Intel:       18,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	companion := &world.NpcInfo{
		ID:     2002,
		NpcID:  45200,
		Impl:   "L1Monster",
		Name:   "same_family",
		Family: "lizardman",
		X:      105,
		Y:      100,
		MapID:  900,
		HP:     50,
		MaxHP:  100,
	}
	otherFamily := &world.NpcInfo{
		ID:     2003,
		NpcID:  45201,
		Impl:   "L1Monster",
		Name:   "other_family",
		Family: "orc",
		X:      105,
		Y:      100,
		MapID:  900,
		HP:     1,
		MaxHP:  100,
	}
	ws.AddNpc(npc)
	ws.AddNpc(companion)
	ws.AddNpc(otherFamily)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcCompanionHealMobSkill(t, s)

	s.tickMonsterAI(npc)

	if companion.HP <= 50 {
		t.Fatalf("Yiwei TriCompanionHp 應改補同 family 低血 NPC，companion HP=%d", companion.HP)
	}
	if target.HP != target.MaxHP {
		t.Fatalf("TriCompanionHp 觸發後不應把治癒術打到玩家目標，target HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
	if otherFamily.HP != 1 {
		t.Fatalf("TriCompanionHp 不應選不同 family NPC，otherFamily HP=%d", otherFamily.HP)
	}
}

func TestNpcMobSkillTriggerCompanionHpWithoutCompanionFallsThroughOtherTriggersLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         104,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "mixed-trigger-caster",
		Family:      "lizardman",
		X:           100,
		Y:           100,
		MapID:       900,
		HP:          100,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		Intel:       18,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	s.deps.MapData = nil
	withNpcCompanionHpMixedTriggerAttackMobSkill(t, s)

	s.tickMonsterAI(npc)

	if target.HP >= target.MaxHP {
		t.Fatalf("yiwei 找不到同族 NPC 時不會直接 return false；TriRange 通過仍應對原目標施放，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func withNpcCompanionHealMobSkill(t *testing.T, s *NpcAISystem) {
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
        name: companion-heal
        type: 2
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 70
        trigger_range: -14
        trigger_count: 0
        change_target: 0
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

func withNpcCompanionHpMixedTriggerAttackMobSkill(t *testing.T, s *NpcAISystem) {
	t.Helper()
	dir := t.TempDir()

	skillPath := filepath.Join(dir, "skill_list.yaml")
	skillRaw := []byte(`skills:
  - skill_id: 910001
    name: fixed-damage-magic
    skill_level: 1
    skill_number: 1
    mp_consume: 0
    hp_consume: 0
    target: attack
    target_to: 0
    damage_value: 40
    damage_dice: 0
    damage_dice_count: 0
    attr: 0
    type: 32
    ranged: 10
    area: 0
    action_id: 18
    cast_gfx: 0
`)
	if err := os.WriteFile(skillPath, skillRaw, 0o644); err != nil {
		t.Fatalf("寫入技能測試資料失敗: %v", err)
	}
	skills, err := data.LoadSkillTable(skillPath)
	if err != nil {
		t.Fatalf("載入技能測試資料失敗: %v", err)
	}

	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: mixed-companion-range-magic
        type: 2
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 70
        trigger_range: -14
        trigger_count: 0
        change_target: 0
        range: 0
        area_width: 0
        area_height: 0
        leverage: 0
        skill_id: 910001
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
