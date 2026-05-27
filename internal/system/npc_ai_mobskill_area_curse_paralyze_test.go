package system

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func TestNpcMobSkillAreaCurseParalyzeAppliesVisiblePlayersLikeJava(t *testing.T) {
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
		NpcID:         55408,
		Impl:          "L1Monster",
		Name:          "area_curse_paralyze",
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
	withNpcAreaCurseParalyzeMobSkill(t, s, 0)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if target.CurseType != 1 || target.CurseTicksLeft != 25 {
		t.Fatalf("yiwei areams 應讓玩家進入詛咒麻痺延遲階段，CurseType=%d ticks=%d", target.CurseType, target.CurseTicksLeft)
	}
	if target.Paralyzed {
		t.Fatal("yiwei CURSE_PARALYZE 前 5 秒只是延遲階段，不應立刻 Paralyzed")
	}
	if !hasServerMessage(packets, 212) {
		t.Fatal("yiwei CURSE_PARALYZE 應送訊息 212：身體漸漸麻痺")
	}
	if !hasPoisonPacket(packets, target.CharID, 2) {
		t.Fatal("yiwei CURSE_PARALYZE 延遲階段應廣播灰色色調")
	}
	if npc.AttackTimer != 5 {
		t.Fatalf("yiwei areams 成功後使用 sub_magic_speed 冷卻，got=%d want=5", npc.AttackTimer)
	}
	if got := npc.MobSkillUseCounts[17]; got != 1 {
		t.Fatalf("yiwei areams 成功才累積 TriCount，got=%d want=1", got)
	}
}

func withNpcAreaCurseParalyzeMobSkill(t *testing.T, s *NpcAISystem, reuseDelay int) {
	t.Helper()
	dir := t.TempDir()
	mobSkillPath := filepath.Join(dir, "mob_skill_list.yaml")
	raw := []byte(fmt.Sprintf(`mob_skills:
  - mob_id: 55408
    skills:
      - act_no: 17
        name: area-curse-paralyze
        type: 17
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
        act_id: 19
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

func hasPoisonPacket(packets [][]byte, objectID int32, poisonType byte) bool {
	wantA, wantB := byte(0), byte(0)
	switch poisonType {
	case 1:
		wantA = 0x01
	case 2:
		wantB = 0x01
	}
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_POISON {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) != objectID {
			continue
		}
		if pkt[5] == wantA && pkt[6] == wantB {
			return true
		}
	}
	return false
}
