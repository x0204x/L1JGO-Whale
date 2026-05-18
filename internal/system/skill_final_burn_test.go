package system

import (
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func attachFinalBurnSkillTable(t *testing.T, s *SkillSystem) {
	t.Helper()
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
}

func TestSkillFinalBurnFinalBurnDamagesWithPreConsumeMP(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "darkelf",
		X:           100,
		Y:           100,
		MapID:       4,
		HP:          250,
		MaxHP:       300,
		MP:          80,
		MaxMP:       100,
		KnownSpells: []int32{108},
		AttackView:  false,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        200,
		MaxHP:     200,
	})
	s := newSkillTestSystem(t, ws)
	attachFinalBurnSkillTable(t, s)

	s.processSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 108, TargetID: target.CharID})

	if target.HP != 120 {
		t.Fatalf("會心一擊應依 Java 使用施法前 MP 80 作為傷害，targetHP=%d", target.HP)
	}
	if caster.HP != 100 || caster.MP != 1 {
		t.Fatalf("會心一擊造成傷害後才應將 HP/MP 扣到 100/1，caster HP/MP=%d/%d", caster.HP, caster.MP)
	}
}

func TestSkillFinalBurnFinalBurnRequiresHpAbove100(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "darkelf",
		X:           100,
		Y:           100,
		MapID:       4,
		HP:          100,
		MaxHP:       300,
		MP:          80,
		MaxMP:       100,
		KnownSpells: []int32{108},
		AttackView:  false,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        200,
		MaxHP:     200,
	})
	s := newSkillTestSystem(t, ws)
	attachFinalBurnSkillTable(t, s)

	s.processSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 108, TargetID: target.CharID})

	if caster.HP != 100 || caster.MP != 80 || target.HP != 200 {
		t.Fatalf("HP <= 100 時會心一擊應失敗且不消耗資源，caster HP/MP=%d/%d targetHP=%d", caster.HP, caster.MP, target.HP)
	}
}

func TestSkillFinalBurnFinalBurnConsumesHpAndMpToJavaFloor(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "darkelf",
		X:           100,
		Y:           100,
		MapID:       4,
		HP:          250,
		MaxHP:       300,
		MP:          80,
		MaxMP:       100,
		Str:         30,
		Dex:         30,
		Level:       60,
		KnownSpells: []int32{108},
		AttackView:  false,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        200,
		MaxHP:     200,
		AC:        10,
	})
	s := newSkillTestSystem(t, ws)
	attachFinalBurnSkillTable(t, s)

	s.processSkill(handler.SkillRequest{SessionID: caster.SessionID, SkillID: 108, TargetID: target.CharID})

	if caster.HP != 100 || caster.MP != 1 {
		t.Fatalf("會心一擊應依 Java 將 HP/MP 扣到 100/1，got HP/MP=%d/%d", caster.HP, caster.MP)
	}
}
