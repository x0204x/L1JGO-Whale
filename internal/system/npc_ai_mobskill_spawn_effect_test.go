package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillSpawnEffectSpawnsL1EffectAtCasterLikeJava(t *testing.T) {
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
		ID:            2001,
		NpcID:         230262,
		Impl:          "L1Monster",
		Name:          "spawn_effect",
		X:             101,
		Y:             100,
		MapID:         900,
		HP:            80,
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
	withNpcSpawnEffectMobSkill(t, s, 10000)
	withNpcSpawnEffectTemplate(t, s)

	s.tickMonsterAI(npc)

	effects := ws.GroundEffectList()
	if len(effects) != 1 {
		t.Fatalf("yiwei SpawnEffect 應在施法者座標生成一個 L1Effect，got=%d", len(effects))
	}
	effect := effects[0]
	if effect.NpcID != 230270 || effect.GfxID != 12750 || effect.Type != world.GroundEffectNpcEffect {
		t.Fatalf("SpawnEffect 應生成 230270/gfx12750 的一般 L1Effect，got npc=%d gfx=%d type=%d", effect.NpcID, effect.GfxID, effect.Type)
	}
	if effect.X != npc.X || effect.Y != npc.Y || effect.MapID != npc.MapID {
		t.Fatalf("Java SpawnEffect 使用 attacker 座標，got=(%d,%d,%d) want=(%d,%d,%d)", effect.X, effect.Y, effect.MapID, npc.X, npc.Y, npc.MapID)
	}
	if effect.TicksLeft != 25 {
		t.Fatalf("mobskill leverage=5 應換算為 5 秒、25 tick，got=%d", effect.TicksLeft)
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei SpawnEffect 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[16]; got != 1 {
		t.Fatalf("yiwei SpawnEffect 成功才累積 TriCount，got=%d want=1", got)
	}
}

func withNpcSpawnEffectMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 230262
    skills:
      - act_no: 16
        name: spawn-effect
        type: 16
        mp_consume: 25
        trigger_random: 1
        trigger_hp: 80
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 1
        change_target: 0
        range: 0
        area_width: 0
        area_height: 0
        leverage: 5
        skill_id: 0
        skill_area: 0
        gfx_id: 0
        act_id: 0
        reuse_delay: %d
        summon_id: 230270
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

func withNpcSpawnEffectTemplate(t *testing.T, s *NpcAISystem) {
	t.Helper()
	dir := t.TempDir()
	npcPath := filepath.Join(dir, "npc_list.yaml")
	raw := []byte(`npcs:
  - npc_id: 230270
    name: 火龍窟副本-特效動畫物件
    nameid: ''
    impl: L1Effect
    gfx_id: 12750
    level: 0
    hp: 0
    mp: 0
    ac: 0
    str: 0
    dex: 0
    con: 0
    wis: 0
    intel: 0
    mr: 0
    exp: 0
    lawful: 0
    size: ''
    ranged: 0
    atk_speed: 0
    sub_magic_speed: 0
    passive_speed: 0
`)
	if err := os.WriteFile(npcPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 npc 測試資料失敗: %v", err)
	}
	npcs, err := data.LoadNpcTable(npcPath)
	if err != nil {
		t.Fatalf("載入 npc 測試資料失敗: %v", err)
	}
	s.deps.Npcs = npcs
}
