package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillSummonMap93DoesNotConsumeTriggerCountLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     93,
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
		MapID:       93,
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
	s.deps.MapData = nil
	withNpcSingleSummonTriggerCountMobSkill(t, s)

	s.tickMonsterAI(npc)
	summonedOnMap93 := len(ws.GetNearbyNpcs(npc.X, npc.Y, npc.MapID)) - 1
	if summonedOnMap93 != 0 {
		t.Fatalf("Yiwei map 93 summon() 會 return false，不應召喚怪物，got=%d", summonedOnMap93)
	}

	ws.RemoveNpc(npc.ID)
	npc.X = 101
	npc.Y = 100
	npc.MapID = 900
	npc.AttackTimer = 0
	ws.AddNpc(npc)
	ws.UpdatePosition(target.SessionID, 100, 100, 900, 0)

	s.tickMonsterAI(npc)
	summonedOnNormalMap := len(ws.GetNearbyNpcs(npc.X, npc.Y, npc.MapID)) - 1
	if summonedOnNormalMap != 1 {
		t.Fatalf("map 93 召喚失敗不應消耗 TriCount，移到一般地圖後應仍可召喚 1 隻，got=%d", summonedOnNormalMap)
	}
}

func TestNpcMobSkillSummonUsesSubMagicSpeedCooldownLikeJava(t *testing.T) {
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
		NpcID:         45197,
		Impl:          "L1Monster",
		Name:          "summoner",
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
	withNpcSingleSummonTriggerCountMobSkill(t, s)

	s.tickMonsterAI(npc)

	if npc.AttackTimer != 5 {
		t.Fatalf("Yiwei 召喚成功後使用 sub_magic_speed 冷卻，want=5 got=%d", npc.AttackTimer)
	}
}

func TestNpcMobSkillSummonInheritsShowAndBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    3,
		HP:        5000,
		MaxHP:     5000,
	})
	otherShowObserver := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other_show",
		X:         101,
		Y:         101,
		MapID:     900,
		ShowID:    8,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45197,
		Impl:          "L1Monster",
		Name:          "summoner",
		X:             101,
		Y:             100,
		MapID:         900,
		ShowID:        3,
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
	withNpcSingleSummonTriggerCountMobSkill(t, s)

	s.tickMonsterAI(npc)

	var summoned *world.NpcInfo
	for _, candidate := range ws.GetNearbyNpcs(npc.X, npc.Y, npc.MapID) {
		if candidate.NpcID == 45244 {
			summoned = candidate
			break
		}
	}
	if summoned == nil {
		t.Fatal("yiwei mobspawn 應生成 summon_id 指定的小怪")
	}
	if summoned.ShowID != npc.ShowID {
		t.Fatalf("yiwei mobspawn 會讓召喚怪繼承 attacker showId，got=%d want=%d", summoned.ShowID, npc.ShowID)
	}
	if packets := drainSkillTestPackets(otherShowObserver.Session); len(packets) != 0 {
		t.Fatalf("yiwei broadcastPacketAll 只送同 ShowID，其他 ShowID 不應收到召喚封包，got=%d packets", len(packets))
	}
	if packets := drainSkillTestPackets(target.Session); len(packets) == 0 {
		t.Fatal("同 ShowID 玩家應收到召喚怪顯示封包")
	}
}

func withNpcSingleSummonTriggerCountMobSkill(t *testing.T, s *NpcAISystem) {
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
        name: summon-once
        type: 3
        mp_consume: 0
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
        act_id: 0
        summon_id: 45244
        summon_min: 1
        summon_max: 1
        poly_id: 0
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
