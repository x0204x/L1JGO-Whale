package system

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newNpcAILOSTestSystem(t *testing.T, ws *world.State) *NpcAISystem {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	return NewNpcAISystem(ws, &handler.Deps{
		World:     ws,
		Scripting: engine,
		MapData:   newSkillLOSTestMap(t),
		Log:       zap.NewNop(),
	})
}

func withNpcSelfHealMobSkill(t *testing.T, s *NpcAISystem) {
	t.Helper()
	withNpcSelfHealMobSkillOptions(t, s, 1, 0)
}

func withNpcSelfHealMobSkillOptions(t *testing.T, s *NpcAISystem, triggerRandom, triggerCount int) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入技能表失敗: %v", err)
	}

	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45197
    skills:
      - act_no: 0
        name: los-self-heal
        type: 2
        mp_consume: 0
        trigger_random: %d
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: %d
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
`, triggerRandom, triggerCount))
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

func TestNpcMeleeAttackSkipsPlayerBehindWallLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45161,
		Impl:   "L1Monster",
		Name:   "spartoi",
		X:      101,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  50,
		STR:    30,
		DEX:    30,
		AtkDmg: 20,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.npcMeleeAttack(npc, target)
	packets := drainSkillTestPackets(target.Session)

	if len(packets) != 0 {
		t.Fatalf("Java isAttackPosition/glanceCheck 阻擋隔牆近戰，Go 不應送攻擊封包，packets=%d", len(packets))
	}
	if target.HP != target.MaxHP {
		t.Fatalf("隔牆近戰不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func TestNpcMeleeAttackSkipsTargetOutsideAttackRangeLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "far_target",
		X:         105,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45161,
		Impl:   "L1Monster",
		Name:   "short_range",
		X:      103,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  50,
		STR:    30,
		DEX:    30,
		AtkDmg: 20,
		Ranged: 1,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.npcMeleeAttack(npc, target)
	packets := drainSkillTestPackets(target.Session)

	if len(packets) != 0 {
		t.Fatalf("Java isAttackPosition 會先檢查距離，近戰距離外不應送攻擊封包，packets=%d", len(packets))
	}
	if target.HP != target.MaxHP {
		t.Fatalf("近戰距離外不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func TestNpcRangedAttackSkipsTargetOutsideAttackRangeLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "far_target",
		X:         105,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45161,
		Impl:   "L1Monster",
		Name:   "short_bow",
		X:      103,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  50,
		STR:    30,
		DEX:    30,
		AtkDmg: 20,
		Ranged: 1,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	s.deps.MapData = nil

	s.npcRangedAttack(npc, target)
	packets := drainSkillTestPackets(target.Session)

	if len(packets) != 0 {
		t.Fatalf("Java isAttackPosition 會先檢查距離，遠攻距離外不應送攻擊封包，packets=%d", len(packets))
	}
	if target.HP != target.MaxHP {
		t.Fatalf("遠攻距離外不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func TestNpcPhysicalSkillSkipsPlayerBehindWallLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         103,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45161,
		Impl:   "L1Monster",
		Name:   "spartoi",
		X:      101,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  50,
		STR:    30,
		DEX:    30,
		AtkDmg: 20,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.executeNpcPhysicalSkill(npc, target, 1, 0, 10, 10, 0, 0)
	packets := drainSkillTestPackets(target.Session)

	if len(packets) != 0 {
		t.Fatalf("Java 物理 mob skill 需受 glanceCheck 阻擋，Go 不應送攻擊封包，packets=%d", len(packets))
	}
	if target.HP != target.MaxHP {
		t.Fatalf("隔牆物理 mob skill 不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func TestNpcRangedAIChasesWallBlockedTargetInRangeLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "wall_target",
		X:         103,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45161,
		Impl:        "L1Monster",
		Name:        "ranged_wall_chaser",
		X:           101,
		Y:           100,
		MapID:       900,
		HP:          100,
		MaxHP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		AtkDmg:      20,
		Ranged:      5,
		MoveSpeed:   800,
		AggroTarget: target.SessionID,
		HateList:    map[uint64]int32{target.SessionID: 10},
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if npc.X == 101 && npc.Y == 100 {
		t.Fatalf("Java isAttackPosition 會因 glanceCheck=false 進入移動邏輯，NPC 不應留在原地")
	}
	if npc.AggroTarget != target.SessionID {
		t.Fatalf("隔牆追擊仍應保留仇恨目標，AggroTarget=%d want=%d", npc.AggroTarget, target.SessionID)
	}
	if npc.AttackTimer != 0 {
		t.Fatalf("隔牆未攻擊成功時不應吃攻擊冷卻，AttackTimer=%d", npc.AttackTimer)
	}
	if len(packets) == 0 {
		t.Fatalf("同 ShowID 玩家應收到 NPC 追擊移動封包")
	}
	if target.HP != target.MaxHP {
		t.Fatalf("隔牆追擊前不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}
