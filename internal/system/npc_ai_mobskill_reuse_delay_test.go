package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillReuseDelayBlocksImmediateReuseLikeJava(t *testing.T) {
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
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "self_healer",
		X:           101,
		Y:           100,
		MapID:       900,
		HP:          10,
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
	withNpcSelfHealReuseDelayMobSkill(t, s, 1000)

	s.tickMonsterAI(npc)
	firstHP := npc.HP
	if firstHP <= 10 {
		t.Fatalf("reuseDelay 測試前置失敗：第一次 mob skill 應成功自補，HP=%d", firstHP)
	}

	npc.AttackTimer = 0
	s.tickMonsterAI(npc)

	if npc.HP != firstHP {
		t.Fatalf("Yiwei reuseDelay 期間同 act_no 不應立刻重複施放，firstHP=%d afterSecond=%d", firstHP, npc.HP)
	}
}

func TestNpcMobSkillReuseDelayExpiresAfterTicksLikeJava(t *testing.T) {
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
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "self_healer",
		X:           101,
		Y:           100,
		MapID:       900,
		HP:          10,
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
	withNpcSelfHealReuseDelayMobSkill(t, s, 1000)

	s.tickMonsterAI(npc)
	firstHP := npc.HP
	if firstHP <= 10 {
		t.Fatalf("reuseDelay 測試前置失敗：第一次 mob skill 應成功自補，HP=%d", firstHP)
	}

	for i := 0; i < 4; i++ {
		npc.AttackTimer = 0
		s.tickMonsterAI(npc)
		if npc.HP != firstHP {
			t.Fatalf("reuseDelay 未到期前不應重複施放，tick=%d firstHP=%d currentHP=%d", i+1, firstHP, npc.HP)
		}
	}

	npc.AttackTimer = 0
	s.tickMonsterAI(npc)

	if npc.HP <= firstHP {
		t.Fatalf("reuseDelay 到期後應允許同 act_no 再次施放，firstHP=%d afterExpire=%d", firstHP, npc.HP)
	}
}

func TestNpcMobSkillTypeDelayBlocksOtherSummonLikeJava(t *testing.T) {
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
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "summoner",
		X:           101,
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
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcTwoSummonReuseDelayMobSkills(t, s)

	s.tickMonsterAI(npc)
	firstSummoned := len(ws.GetNearbyNpcs(npc.X, npc.Y, npc.MapID)) - 1
	if firstSummoned != 1 {
		t.Fatalf("reuseDelay 類型冷卻測試前置失敗：第一次應只召喚 1 隻，got=%d", firstSummoned)
	}

	npc.AttackTimer = 0
	s.tickMonsterAI(npc)
	secondSummoned := len(ws.GetNearbyNpcs(npc.X, npc.Y, npc.MapID)) - 1

	if secondSummoned != firstSummoned {
		t.Fatalf("Yiwei type 3 召喚冷卻期間應阻擋其他 act_no 的召喚，first=%d second=%d", firstSummoned, secondSummoned)
	}
}

func withNpcSelfHealReuseDelayMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入 skill_list.yaml 失敗: %v", err)
	}

	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: reuse-delay-self-heal
        type: 2
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 0
        change_target: 2
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
        reuse_delay: %d
`, reuseDelay))
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

func withNpcTwoSummonReuseDelayMobSkills(t *testing.T, s *NpcAISystem) {
	t.Helper()
	dir := t.TempDir()
	npcPath := filepath.Join(dir, "npc_list.yaml")
	if err := os.WriteFile(npcPath, []byte(`npcs:
  - npc_id: 45244
    name: summoned-one
    impl: L1Monster
    gfx_id: 1
    level: 1
    hp: 10
    mp: 0
    ac: 10
    str: 10
    dex: 10
    ranged: 1
  - npc_id: 45245
    name: summoned-two
    impl: L1Monster
    gfx_id: 1
    level: 1
    hp: 10
    mp: 0
    ac: 10
    str: 10
    dex: 10
    ranged: 1
`), 0o644); err != nil {
		t.Fatalf("寫入 NPC 測試資料失敗: %v", err)
	}
	npcs, err := data.LoadNpcTable(npcPath)
	if err != nil {
		t.Fatalf("載入 NPC 測試資料失敗: %v", err)
	}

	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: reuse-delay-summon-one
        type: 3
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
        act_id: 0
        summon_id: 45244
        summon_min: 1
        summon_max: 1
        poly_id: 0
        reuse_delay: 1000
      - act_no: 1
        name: reuse-delay-summon-two
        type: 3
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
        act_id: 0
        summon_id: 45245
        summon_min: 1
        summon_max: 1
        poly_id: 0
        reuse_delay: 1000
`)
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}

	s.deps.Npcs = npcs
	s.deps.MobSkills = mobSkills
}
