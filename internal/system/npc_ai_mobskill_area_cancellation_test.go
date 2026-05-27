package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaCancellationCancelsVisiblePlayerBuffsLikeJava(t *testing.T) {
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
	target.AddBuff(&world.ActiveBuff{SkillID: 29, TicksLeft: 80})
	target.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	target.Paralyzed = true
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45197,
		Impl:          "L1Monster",
		Name:          "area_canceller",
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
	withNpcAreaCancellationMobSkill(t, s, 0)

	s.tickMonsterAI(npc)

	if target.HasBuff(29) {
		t.Fatalf("yiwei areacancellation 應透過 CANCELLATION 移除可相消 buff 29")
	}
	if !target.HasBuff(87) {
		t.Fatalf("yiwei CANCELLATION 應保留不可相消的 SHOCK_STUN(87)")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei areacancellation 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[0]; got != 1 {
		t.Fatalf("yiwei areacancellation 成功才累積 TriCount，got=%d want=1", got)
	}
}

func withNpcAreaCancellationMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: area-cancellation
        type: 6
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
        leverage: 0
        skill_id: 0
        skill_area: 0
        gfx_id: 0
        act_id: 47
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
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}

	s.deps.MobSkills = mobSkills
}
