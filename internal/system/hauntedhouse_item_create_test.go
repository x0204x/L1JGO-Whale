package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHauntedHouseGiveRewardUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := &HauntedHouseSystem{
		deps: &handler.Deps{
			Items:      testItemTable(t),
			ItemCreate: itemCreate,
			Log:        zap.NewNop(),
		},
	}
	player := &world.PlayerInfo{
		Name: "haunted-test",
		Inv:  world.NewInventory(),
	}

	sys.GiveReward(sess, player)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 41308 || itemCreate.count != 1 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=41308 count=1", itemCreate.itemID, itemCreate.count)
	}
}
