package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaWeaponBreakDamagesVisiblePlayerWeaponLikeJava(t *testing.T) {
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
	weapon := target.Inv.AddItemWithID(7001, 1, 1, "sword", 0, 0, false, 0)
	weapon.Equipped = true
	target.Equip.Set(world.SlotWeapon, weapon)
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45670,
		Impl:          "L1Monster",
		Name:          "area_weapon_break",
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
	s.deps.Items = loadDurabilityTestItems(t)
	withNpcAreaWeaponBreakMobSkill(t, s, 0)
	withNpcAreaWeaponBreakNpcTemplate(t, s)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if weapon.Durability != 1 {
		t.Fatalf("yiwei weapon_break 應依怪物 INT/3 損壞玩家裝備武器，got=%d want=1", weapon.Durability)
	}
	if !hasSkillEffectPacket(packets, target.CharID, 172) {
		t.Fatal("yiwei weapon_break 應對每個有武器的玩家播放 172 特效")
	}
	if !hasServerMessage(packets, 268) {
		t.Fatal("yiwei weapon_break 應送出訊息 268 顯示武器受損")
	}
	if !hasDurabilityStatusPacket(packets, weapon.ObjectID, 1) {
		t.Fatal("yiwei weapon_break 應送 S_ItemStatus 更新武器損壞度")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei weapon_break 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[7]; got != 1 {
		t.Fatalf("yiwei weapon_break 成功才累積 TriCount，got=%d want=1", got)
	}
}

func withNpcAreaWeaponBreakMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45670
    skills:
      - act_no: 7
        name: area-weapon-break
        type: 7
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

func withNpcAreaWeaponBreakNpcTemplate(t *testing.T, s *NpcAISystem) {
	t.Helper()
	dir := t.TempDir()
	npcPath := filepath.Join(dir, "npc_list.yaml")
	raw := []byte(`npcs:
  - npc_id: 45670
    name: area_weapon_break
    nameid: area_weapon_break
    impl: L1Monster
    gfx_id: 1
    level: 50
    hp: 100
    mp: 100
    ac: 0
    str: 30
    dex: 30
    con: 30
    wis: 30
    intel: 3
    mr: 0
    exp: 0
    lawful: 0
    size: small
    ranged: 1
    atk_speed: 1000
    sub_magic_speed: 1000
    passive_speed: 1000
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
