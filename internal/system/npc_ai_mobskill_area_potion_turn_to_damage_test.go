package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaPotionTurnToDamageAppliesVisiblePlayersLikeJava(t *testing.T) {
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
	existingPollute := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "existing_pollute",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	existingPollute.AddBuff(&world.ActiveBuff{SkillID: 4012, TicksLeft: 25})
	existingDecay := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "existing_decay",
		X:         102,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	existingDecay.AddBuff(&world.ActiveBuff{SkillID: 71, TicksLeft: 80})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45600,
		Impl:          "L1Monster",
		Name:          "area_potion_turn_to_damage",
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
	withNpcAreaPotionTurnToDamageMobSkill(t, s, 0)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if !target.HasBuff(4011) {
		t.Fatal("yiwei potionturntodmg 應對可見玩家套 4011 藥水侵蝕術")
	}
	if got := target.GetBuff(4011).TicksLeft; got != 60 {
		t.Fatalf("yiwei potionturntodmg 固定 12 秒，Go tick 應為 60，got=%d", got)
	}
	if existingPollute.HasBuff(4011) {
		t.Fatal("yiwei potionturntodmg 對已有 4012 的玩家不應再套 4011")
	}
	if existingDecay.HasBuff(4011) {
		t.Fatal("yiwei potionturntodmg 對已有 71 藥霜的玩家不應再套 4011")
	}
	if !hasSkillEffectPacket(packets, target.CharID, 7781) {
		t.Fatal("yiwei potionturntodmg 應對新套用目標播放 7781")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei potionturntodmg 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[4]; got != 1 {
		t.Fatalf("yiwei potionturntodmg 成功才累積 TriCount，got=%d want=1", got)
	}
}

func withNpcAreaPotionTurnToDamageMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45600
    skills:
      - act_no: 4
        name: area-potion-turn-to-damage
        type: 8
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
