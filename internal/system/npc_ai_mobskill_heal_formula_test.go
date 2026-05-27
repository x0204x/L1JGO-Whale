package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillSelfHealUsesNpcMagicHealingFormulaLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
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
		X:           100,
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
		Lawful:      0,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcHealFormulaMobSkill(t, s, 990001, 2, 0, 10)

	s.tickMonsterAI(npc)

	if npc.HP != 15 {
		t.Fatalf("yiwei NPC 自補應以 damage_value + INT magicBonus 擲骰，HP=%d want=15", npc.HP)
	}
}

func TestNpcMobSkillCompanionHealUsesNpcLawfulAndLeverageLikeJava(t *testing.T) {
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
		Lawful:      32768,
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
		HP:     10,
		MaxHP:  100,
	}
	ws.AddNpc(npc)
	ws.AddNpc(companion)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcHealFormulaMobSkill(t, s, 990001, 0, 70, 20)

	s.tickMonsterAI(npc)

	if companion.HP != 30 {
		t.Fatalf("yiwei NPC 同族補血應套用 caster lawful 與 mobskill leverage，HP=%d want=30", companion.HP)
	}
}

func TestNpcMobSkillSelfHealPolluteWaterHalvesHealingLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "polluted_self_healer",
		X:           100,
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
		Lawful:      0,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	npc.AddDebuff(skillPolluteWater, 960)
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcHealFormulaMobSkill(t, s, 990001, 2, 0, 10)

	s.tickMonsterAI(npc)

	if npc.HP != 12 {
		t.Fatalf("yiwei NPC 自補目標有 POLLUTE_WATER 時回復量減半，HP=%d want=12", npc.HP)
	}
}

func TestNpcMobSkillSelfHealDragonWaterStatusesModifyHealingLikeJava(t *testing.T) {
	tests := []struct {
		name     string
		skillID  int32
		wantHP   int32
		wantText string
	}{
		{
			name:     "pollute-water-wave",
			skillID:  mobSkillPolluteWater,
			wantHP:   12,
			wantText: "4012 污濁的水流應讓 NPC 魔法治療減半",
		},
		{
			name:     "heal-turn-to-damage",
			skillID:  mobSkillHealTurnToDamage,
			wantHP:   5,
			wantText: "4013 治癒侵蝕術應把 NPC 魔法治療轉為傷害",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := world.NewState()
			target := addSkillTestPlayer(ws, &world.PlayerInfo{
				SessionID: 1,
				Session:   newSkillTestSession(t, 1),
				CharID:    1001,
				Name:      "target",
				X:         101,
				Y:         100,
				MapID:     900,
				HP:        5000,
				MaxHP:     5000,
			})
			npc := &world.NpcInfo{
				ID:          2001,
				NpcID:       45197,
				Impl:        "L1Monster",
				Name:        "dragon_status_self_healer",
				X:           100,
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
				Lawful:      0,
				AtkDmg:      20,
				Ranged:      1,
				AggroTarget: target.SessionID,
			}
			npc.AddDebuff(tt.skillID, 60)
			ws.AddNpc(npc)
			s := newNpcAILOSTestSystem(t, ws)
			withNpcHealFormulaMobSkill(t, s, 990001, 2, 0, 10)

			s.tickMonsterAI(npc)

			if npc.HP != tt.wantHP {
				t.Fatalf("%s，HP=%d want=%d", tt.wantText, npc.HP, tt.wantHP)
			}
		})
	}
}

func TestNpcMobSkillCompanionHealPolluteWaterHalvesTargetHealingLikeJava(t *testing.T) {
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
		Lawful:      32768,
		AtkDmg:      20,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	companion := &world.NpcInfo{
		ID:     2002,
		NpcID:  45200,
		Impl:   "L1Monster",
		Name:   "polluted_same_family",
		Family: "lizardman",
		X:      105,
		Y:      100,
		MapID:  900,
		HP:     10,
		MaxHP:  100,
	}
	companion.AddDebuff(skillPolluteWater, 960)
	ws.AddNpc(npc)
	ws.AddNpc(companion)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcHealFormulaMobSkill(t, s, 990001, 0, 70, 20)

	s.tickMonsterAI(npc)

	if companion.HP != 20 {
		t.Fatalf("yiwei NPC 同族補血看被補目標的 POLLUTE_WATER 減半，HP=%d want=20", companion.HP)
	}
}

func TestNpcMobSkillSelfHealBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
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
		X:         100,
		Y:         101,
		MapID:     900,
		ShowID:    8,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45197,
		Impl:        "L1Monster",
		Name:        "self_healer",
		X:           100,
		Y:           100,
		MapID:       900,
		ShowID:      3,
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
	withNpcHealFormulaMobSkill(t, s, 990001, 2, 0, 10)

	s.tickMonsterAI(npc)

	packets := drainSkillTestPackets(otherShowObserver.Session)
	if hasSkillEffectPacket(packets, npc.ID, 744) {
		t.Fatal("yiwei L1Character.broadcastPacketAll 只發同 ShowID，自補特效不應送給其他 ShowID 玩家")
	}
}

func TestNpcMobSkillCompanionHealBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         104,
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
		X:         103,
		Y:         101,
		MapID:     900,
		ShowID:    8,
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
		ShowID:      3,
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
		ShowID: 3,
		HP:     10,
		MaxHP:  100,
	}
	ws.AddNpc(npc)
	ws.AddNpc(companion)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcHealFormulaMobSkill(t, s, 990001, 0, 70, 10)

	s.tickMonsterAI(npc)

	packets := drainSkillTestPackets(otherShowObserver.Session)
	if hasSkillEffectPacket(packets, companion.ID, 744) {
		t.Fatal("yiwei L1Character.broadcastPacketAll 只發同 ShowID，同族補血特效不應送給其他 ShowID 玩家")
	}
}

func withNpcHealFormulaMobSkill(t *testing.T, s *NpcAISystem, skillID, changeTarget, triggerCompanionHP, leverage int) {
	t.Helper()
	dir := t.TempDir()
	skillPath := filepath.Join(dir, "skill_list.yaml")
	skillRaw := []byte(fmt.Sprintf(`skills:
  - skill_id: %d
    name: npc-heal-formula
    skill_level: 1
    skill_number: 0
    mp_consume: 0
    hp_consume: 0
    item_consume_id: 0
    item_consume_count: 0
    reuse_delay: 0
    buff_duration: 0
    target: buff
    target_to: 3
    damage_value: 2
    damage_dice: 1
    damage_dice_count: 99
    probability_value: 0
    probability_dice: 0
    attr: 4
    type: 16
    lawful: 0
    ranged: -1
    area: 0
    through: 0
    id: 1
    name_id: "$npc-heal-formula"
    action_id: 19
    cast_gfx: 744
    cast_gfx2: 0
    sys_msg_happen: 0
    sys_msg_stop: 0
    sys_msg_fail: 0
`, skillID))
	if err := os.WriteFile(skillPath, skillRaw, 0o644); err != nil {
		t.Fatalf("寫入技能測試資料失敗: %v", err)
	}
	skills, err := data.LoadSkillTable(skillPath)
	if err != nil {
		t.Fatalf("載入技能測試資料失敗: %v", err)
	}

	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	mobSkillRaw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: npc-heal-formula
        type: 2
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: %d
        trigger_range: -14
        trigger_count: 0
        change_target: %d
        range: 0
        area_width: 0
        area_height: 0
        leverage: %d
        skill_id: %d
        skill_area: 0
        gfx_id: 0
        act_id: 18
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: 0
`, triggerCompanionHP, changeTarget, leverage, skillID))
	if err := os.WriteFile(mobSkillPath, mobSkillRaw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("載入 mob skill 測試資料失敗: %v", err)
	}

	s.deps.Skills = skills
	s.deps.MobSkills = mobSkills
}
