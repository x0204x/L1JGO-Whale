package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaDebuffAppliesLeverageDurationLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	existingPotionTurn := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "existing_potion_turn",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	existingPotionTurn.AddBuff(&world.ActiveBuff{SkillID: mobSkillPotionTurnToDamage, TicksLeft: 25})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         97257,
		Impl:          "L1Monster",
		Name:          "area_debuff",
		X:             101,
		Y:             100,
		MapID:         900,
		HP:            100,
		MaxHP:         100,
		MP:            0,
		MaxMP:         100,
		Level:         50,
		STR:           30,
		DEX:           30,
		AtkDmg:        20,
		SubMagicSpeed: 1000,
		Ranged:        1,
		AggroTarget:   target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	s.deps.Skill = newSkillTestSystem(t, ws)
	withNpcAreaDebuffMobSkill(t, s, 15000)

	s.tickMonsterAI(npc)

	if !target.HasBuff(71) {
		t.Fatalf("yiwei areadebuff 應對可見玩家套用 mobskill skill_id=71")
	}
	if got := target.GetBuff(71).TicksLeft; got != 25 {
		t.Fatalf("yiwei areadebuff 借用 leverage=5 作為 5 秒 buff，Go tick 應為 25，got=%d", got)
	}
	if existingPotionTurn.HasBuff(71) {
		t.Fatalf("yiwei areadebuff skill 71 不應與 4011 藥水侵蝕術共存")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei areadebuff buff 類技能成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[4]; got != 1 {
		t.Fatalf("yiwei areadebuff 成功後應消耗 TriCount，got=%d want=1", got)
	}
}

func withNpcAreaDebuffMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能表失敗: %v", err)
	}

	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 97257
    skills:
      - act_no: 4
        name: area-debuff-decay-potion
        type: 14
        mp_consume: 25
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 1
        change_target: 0
        range: 0
        area_width: 0
        area_height: 0
        leverage: 5
        skill_id: 71
        skill_area: 0
        gfx_id: 11030
        act_id: 18
        reuse_delay: %d
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: 0
`, reuseDelay))
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("讀取 mob skill 測試資料失敗: %v", err)
	}

	s.deps.Skills = skills
	s.deps.MobSkills = mobSkills
}
