package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

type fakeSummonManager struct {
	summonTargetID int32
}

func (f *fakeSummonManager) ExecuteSummonMonster(_ *l1net.Session, _ *world.PlayerInfo, _ *data.SkillInfo, targetID int32) {
	f.summonTargetID = targetID
}

func (f *fakeSummonManager) ExecuteElementalSummon(_ *l1net.Session, _ *world.PlayerInfo, _ *data.SkillInfo) {
}

func (f *fakeSummonManager) ExecuteTamingMonster(_ *l1net.Session, _ *world.PlayerInfo, _ *data.SkillInfo, _ int32) {
}

func (f *fakeSummonManager) ExecuteCreateZombie(_ *l1net.Session, _ *world.PlayerInfo, _ *data.SkillInfo, _ int32) {
}

func (f *fakeSummonManager) ExecuteReturnToNature(_ *l1net.Session, _ *world.PlayerInfo, _ *data.SkillInfo) {
}

func (f *fakeSummonManager) DismissSummon(_ *world.SummonInfo, _ *world.PlayerInfo) {
}

func newTeleportSummonTestSystem(t *testing.T, ws *world.State) *SkillSystem {
	t.Helper()
	skills, err := data.LoadSkillTable("../../data/yaml/skill_list.yaml")
	if err != nil {
		t.Fatalf("載入技能資料失敗: %v", err)
	}
	return &SkillSystem{deps: &handler.Deps{
		World:  ws,
		Skills: skills,
		Log:    zap.NewNop(),
	}}
}

func TestSkillTeleportSummonSummonMonsterUsesParsedSummonID(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "summoner",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     60,
		Inv:       world.NewInventory(),
		KnownSpells: []int32{
			51,
		},
	})
	player.Inv.AddItem(40318, 3, "magic gem", 0, 0, true, 1)
	summon := &fakeSummonManager{}
	s := newTeleportSummonTestSystem(t, ws)
	s.deps.Summon = summon

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   51,
		SummonID:  263,
	})

	if summon.summonTargetID != 263 {
		t.Fatalf("召喚術應使用封包解析出的 SummonID，got=%d", summon.summonTargetID)
	}
}

func TestSkillTeleportSummonTeleportUsesParsedBookmarkID(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "teleporter",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     100,
		MP:        100,
		MaxMP:     100,
		Level:     52,
		Inv:       world.NewInventory(),
		KnownSpells: []int32{
			5,
		},
		Bookmarks: []world.Bookmark{
			{ID: 77, Name: "target", X: 32710, Y: 32820, MapID: 4},
		},
	})
	s := newTeleportSummonTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID:  player.SessionID,
		SkillID:    5,
		BookmarkID: 77,
		MapID:      4,
	})

	if player.X != 32710 || player.Y != 32820 || player.MapID != 4 {
		t.Fatalf("瞬間移動應使用封包解析出的 BookmarkID，位置=%d,%d,%d", player.X, player.Y, player.MapID)
	}
}
