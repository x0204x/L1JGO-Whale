package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestCreateMinionCopiesFamilyAgroFamilyLikeJava(t *testing.T) {
	sys := &NpcRespawnSystem{deps: &handler.Deps{}}
	leader := &world.NpcInfo{MapID: 4}

	minion := sys.createMinion(&data.NpcTemplate{
		NpcID:      45161,
		Name:       "spartoi",
		HP:         30,
		Family:     "orc",
		AgroFamily: 1,
	}, 101, 100, leader)

	if minion.Family != "orc" || minion.AgroFamily != 1 {
		t.Fatalf("minion 應繼承模板 family/agro_family：family=%q agro_family=%d", minion.Family, minion.AgroFamily)
	}
}
