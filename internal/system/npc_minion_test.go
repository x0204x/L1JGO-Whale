package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestCreateMinionInheritsLeaderShowIDAndSinkHiddenLikeJava(t *testing.T) {
	sys := &NpcRespawnSystem{deps: &handler.Deps{}}
	leader := &world.NpcInfo{
		X:                  100,
		Y:                  100,
		MapID:              4,
		Heading:            5,
		ShowID:             77,
		HiddenStatus:       world.NpcHiddenSink,
		HiddenActionStatus: 13,
	}

	minion := sys.createMinion(&data.NpcTemplate{
		NpcID: 45161,
		Name:  "spartoi",
		HP:    30,
		MP:    5,
	}, 101, 100, leader)

	if minion.ShowID != leader.ShowID {
		t.Fatalf("yiwei L1MobGroupSpawn 會讓隊員繼承 leader showId，got=%d want=%d", minion.ShowID, leader.ShowID)
	}
	if minion.HiddenStatus != world.NpcHiddenSink || minion.HiddenActionStatus != 13 {
		t.Fatalf("leader sink hidden 時，史巴托系隊員應繼承 sink/status13，HiddenStatus=%d Action=%d", minion.HiddenStatus, minion.HiddenActionStatus)
	}
}

func TestCreateMinionDoesNotHideUnsupportedMinionWhenLeaderSinkLikeJava(t *testing.T) {
	sys := &NpcRespawnSystem{deps: &handler.Deps{}}
	leader := &world.NpcInfo{MapID: 4, HiddenStatus: world.NpcHiddenSink}

	minion := sys.createMinion(&data.NpcTemplate{
		NpcID: 45067,
		Name:  "harpy",
		HP:    30,
	}, 101, 100, leader)

	if minion.HiddenStatus != world.NpcHiddenNone || minion.HiddenActionStatus != 0 {
		t.Fatalf("leader sink hidden 時，飛天系隊員不在 Java sink minion 清單內，不應 hidden，HiddenStatus=%d Action=%d", minion.HiddenStatus, minion.HiddenActionStatus)
	}
}
