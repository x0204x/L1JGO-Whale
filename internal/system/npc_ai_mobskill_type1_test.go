package system

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestNpcMobSkillTypeOneUsesMobSkillRangeLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         102,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
		AC:        10,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45040,
		Impl:        "L1Monster",
		Name:        "type-one-range",
		X:           100,
		Y:           100,
		MapID:       900,
		HP:          100,
		MaxHP:       100,
		MP:          0,
		MaxMP:       100,
		Level:       50,
		STR:         30,
		DEX:         30,
		AtkDmg:      20,
		AtkSpeed:    1000,
		Ranged:      1,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)
	withNpcTypeOnePhysicalMobSkill(t, s, 1)

	s.tickMonsterAI(npc)

	if got := npc.MobSkillUseCounts[1]; got != 0 {
		t.Fatalf("yiwei physicalAttack 距離超過 mobskill range 時 return false，不應消耗 TriCount，got=%d", got)
	}
	if npc.AttackTimer != 0 {
		t.Fatalf("yiwei physicalAttack 距離超過 range 時不應進入攻擊冷卻，got=%d", npc.AttackTimer)
	}
	if target.HP != 5000 {
		t.Fatalf("yiwei physicalAttack 距離超過 range 不應造成傷害，got HP=%d", target.HP)
	}
}

func TestNpcMobSkillTypeOneAreaBoxDamagesFrontPlayersLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	front := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "front",
		X:         102,
		Y:         100,
		MapID:     900,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	side := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "side",
		X:         101,
		Y:         101,
		MapID:     900,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	behind := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 4,
		Session:   newSkillTestSession(t, 4),
		CharID:    1004,
		Name:      "behind",
		X:         99,
		Y:         100,
		MapID:     900,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 5,
		Session:   newSkillTestSession(t, 5),
		CharID:    1005,
		Name:      "other_show",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    7,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45040,
		Impl:   "L1Monster",
		Name:   "type-one-area",
		X:      100,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  100,
		STR:    127,
		DEX:    127,
		AtkDmg: 50,
		Ranged: 1,
	}
	ws.AddNpc(npc)
	s := newNpcAIPhysicalAreaTestSystem(t, ws)

	for i := 0; i < 20; i++ {
		if !s.executeNpcPhysicalSkill(npc, target, 30, 0, 10, 3, 0, 3) {
			t.Fatalf("yiwei physicalAttack 有前方 box 目標時應 return true")
		}
	}

	if front.HP >= front.MaxHP {
		t.Fatalf("yiwei physicalAttack area_height>0 應傷害前方 box 內玩家，front HP=%d", front.HP)
	}
	if side.HP != side.MaxHP {
		t.Fatalf("area_width=0 時側邊玩家不應被納入，side HP=%d", side.HP)
	}
	if behind.HP != behind.MaxHP {
		t.Fatalf("背後玩家不應被納入前方 box，behind HP=%d", behind.HP)
	}
	if otherShow.HP != otherShow.MaxHP {
		t.Fatalf("不同 ShowID 玩家不應被納入前方 box，otherShow HP=%d", otherShow.HP)
	}
}

func TestNpcMobSkillTypeOneBroadcastsOnlySameShowLikeJava(t *testing.T) {
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
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "other_show",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    8,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45040,
		Impl:   "L1Monster",
		Name:   "type-one-broadcast",
		X:      100,
		Y:      100,
		MapID:  900,
		ShowID: 3,
		HP:     100,
		MaxHP:  100,
		Level:  100,
		STR:    127,
		DEX:    127,
		AtkDmg: 50,
		Ranged: 1,
	}
	ws.AddNpc(npc)
	s := newNpcAIPhysicalAreaTestSystem(t, ws)

	if !s.executeNpcPhysicalSkill(npc, target, 30, 0, 10, 3, 0, 0) {
		t.Fatal("yiwei physicalAttack 單體目標有效時應 return true")
	}

	if !hasNpcAttackPacket(drainSkillTestPackets(target.Session), npc.ID) {
		t.Fatal("同 ShowID 玩家應收到 NPC type 1 物理技能攻擊封包")
	}
	if hasNpcAttackPacket(drainSkillTestPackets(otherShow.Session), npc.ID) {
		t.Fatal("yiwei L1AttackNpc.action 走 broadcastPacketAll，不同 ShowID 不應收到 NPC type 1 物理技能攻擊封包")
	}
}

func TestNpcMobSkillTypeOneUsesActIDOverrideLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        100000,
		MaxHP:     100000,
		AC:        10,
	})
	npc := &world.NpcInfo{
		ID:     2001,
		NpcID:  45061,
		Impl:   "L1Monster",
		Name:   "spartoi-act30",
		X:      100,
		Y:      100,
		MapID:  900,
		HP:     100,
		MaxHP:  100,
		Level:  100,
		STR:    127,
		DEX:    127,
		AtkDmg: 50,
		Ranged: 1,
	}
	ws.AddNpc(npc)
	s := newNpcAIPhysicalAreaTestSystem(t, ws)

	if !s.executeNpcPhysicalSkill(npc, target, 30, 0, 10, 1, 0, 0) {
		t.Fatal("yiwei physicalAttack act_id=30 應成功送出指定攻擊動作")
	}

	if !hasNpcAttackActionPacket(drainSkillTestPackets(target.Session), npc.ID, 30) {
		t.Fatal("Java L1MobSkillUse.physicalAttack 會 setActId(act_id)，史巴托/高崙 type 1 應送 action=30")
	}
}

func hasNpcAttackActionPacket(packets [][]byte, npcID int32, action byte) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ATTACK {
			continue
		}
		if pkt[1] == action && int32(binary.LittleEndian.Uint32(pkt[2:6])) == npcID {
			return true
		}
	}
	return false
}

func newNpcAIPhysicalAreaTestSystem(t *testing.T, ws *world.State) *NpcAISystem {
	t.Helper()
	engine, err := scripting.NewEngine("../../scripts", zap.NewNop())
	if err != nil {
		t.Fatalf("建立 Lua engine 失敗: %v", err)
	}
	return NewNpcAISystem(ws, &handler.Deps{
		World:     ws,
		Scripting: engine,
		Log:       zap.NewNop(),
	})
}

func withNpcTypeOnePhysicalMobSkill(t *testing.T, s *NpcAISystem, skillRange int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 45040
    skills:
      - act_no: 1
        name: type-one-physical
        type: 1
        mp_consume: 0
        trigger_random: 1
        trigger_hp: 0
        trigger_companion_hp: 0
        trigger_range: -14
        trigger_count: 1
        change_target: 0
        range: %d
        area_width: 0
        area_height: 0
        leverage: 10
        skill_id: 0
        skill_area: 0
        gfx_id: 0
        act_id: 30
        reuse_delay: 0
        summon_id: 0
        summon_min: 0
        summon_max: 0
        poly_id: 0
`, skillRange))
	if err := os.WriteFile(mobSkillPath, raw, 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("讀取 mob skill 測試資料失敗: %v", err)
	}
	s.deps.MobSkills = mobSkills
}
