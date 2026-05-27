package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaPolluteWaterAndHealTurnApplyVisiblePlayersLikeJava(t *testing.T) {
	for _, tc := range []struct {
		name    string
		typ     int
		command string
		skillID int32
		gfxID   int32
		actID   int
	}{
		{name: "pollute-water-wave", typ: 9, command: "area_pollute_water_wave", skillID: mobSkillPolluteWater, gfxID: 7782, actID: 48},
		{name: "heal-turn-to-damage", typ: 10, command: "area_heal_turn_to_damage", skillID: mobSkillHealTurnToDamage, gfxID: 7780, actID: 49},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
				Name:      "existing_4011",
				X:         101,
				Y:         100,
				MapID:     900,
				HP:        5000,
				MaxHP:     5000,
			})
			existingPotionTurn.AddBuff(&world.ActiveBuff{SkillID: mobSkillPotionTurnToDamage, TicksLeft: 25})
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
			existingDecay.AddBuff(&world.ActiveBuff{SkillID: mobSkillDecayPotion, TicksLeft: 80})
			npc := &world.NpcInfo{
				ID:            2001,
				NpcID:         45600,
				Impl:          "L1Monster",
				Name:          tc.command,
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
			withNpcAreaSpecialStatusMobSkill(t, s, tc.typ, 4, tc.actID, 0)

			s.tickMonsterAI(npc)
			packets := drainSkillTestPackets(target.Session)

			if !target.HasBuff(tc.skillID) {
				t.Fatalf("yiwei %s 應對可見玩家套 %d", tc.command, tc.skillID)
			}
			if got := target.GetBuff(tc.skillID).TicksLeft; got != 60 {
				t.Fatalf("yiwei %s 固定 12 秒，Go tick 應為 60，got=%d", tc.command, got)
			}
			if existingPotionTurn.HasBuff(tc.skillID) {
				t.Fatalf("yiwei %s 對已有 4011 的玩家不應再套 %d", tc.command, tc.skillID)
			}
			if !existingDecay.HasBuff(tc.skillID) {
				t.Fatalf("yiwei %s 只檢查 4011/4012/4013，不會因 71 藥霜跳過", tc.command)
			}
			if !hasSkillEffectPacket(packets, target.CharID, tc.gfxID) {
				t.Fatalf("yiwei %s 應對新套用目標播放 %d", tc.command, tc.gfxID)
			}
			if npc.AttackTimer != 5 {
				t.Fatalf("yiwei %s 成功後使用 sub_magic_speed 冷卻，got=%d want=5", tc.command, npc.AttackTimer)
			}
			if got := npc.MobSkillUseCounts[4]; got != 1 {
				t.Fatalf("yiwei %s 成功才累積 TriCount，got=%d want=1", tc.command, got)
			}
		})
	}
}

func withNpcAreaSpecialStatusMobSkill(t *testing.T, s *NpcAISystem, typ, actNo, actID, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45600
    skills:
      - act_no: %d
        name: area-special-status
        type: %d
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
        act_id: %d
        reuse_delay: %d
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: 0
`, actNo, typ, actID, reuseDelay))
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}

	s.deps.MobSkills = mobSkills
}
