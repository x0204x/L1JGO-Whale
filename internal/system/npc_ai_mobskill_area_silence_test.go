package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaSilenceSilencesVisiblePlayersLikeJava(t *testing.T) {
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
	alreadyAreaSilenced := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "area_silenced",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
		Silenced:  true,
	})
	alreadyAreaSilenced.AddBuff(&world.ActiveBuff{SkillID: 161, TicksLeft: 80, SetSilenced: true})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45197,
		Impl:          "L1Monster",
		Name:          "area_silencer",
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
	withNpcAreaSilenceMobSkill(t, s, 0)

	s.tickMonsterAI(npc)

	if !target.Silenced || !target.HasBuff(64) {
		t.Fatalf("yiwei areasilence 應對可見玩家套 64 SILENCE，silenced=%v buff64=%v", target.Silenced, target.GetBuff(64))
	}
	if got := target.GetBuff(64).TicksLeft; got != 80 {
		t.Fatalf("yiwei areasilence 固定 16 秒，Go tick 應為 80，got=%d", got)
	}
	if alreadyAreaSilenced.HasBuff(64) {
		t.Fatalf("yiwei areasilence 對已有 161 AREA_OF_SILENCE 的玩家不應再套 64")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei areasilence 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[0]; got != 1 {
		t.Fatalf("yiwei areasilence 成功才累積 TriCount，got=%d want=1", got)
	}
}

func TestNpcMobSkillAreaSilenceBroadcastsTargetEffectFromTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         120,
		Y:         100,
		MapID:     900,
		ShowID:    7,
		HP:        5000,
		MaxHP:     5000,
	})
	observerNearTarget := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "observer",
		X:         121,
		Y:         100,
		MapID:     900,
		ShowID:    7,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45197,
		Impl:   "L1Monster",
		Name:   "area_silencer",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 7,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.executeNpcAreaSilenceLikeJava(npc, 19)

	if !target.HasBuff(64) {
		t.Fatal("目標在怪物可見距離邊緣時仍應被 areasilence 套 64")
	}
	packets := drainSkillTestPackets(observerNearTarget.Session)
	if !hasSkillEffectPacket(packets, target.CharID, 10708) {
		t.Fatal("yiwei pc.sendPacketsAll 會從目標位置廣播 10708，目標旁觀察者應看得到")
	}
	if hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("觀察者不在怪物可見範圍內時，不應收到施法者動作封包")
	}
}

func withNpcAreaSilenceMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入技能資料失敗: %v", err)
	}

	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: area-silence
        type: 11
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
        act_id: 51
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

	s.deps.Skills = skills
	s.deps.MobSkills = mobSkills
}
