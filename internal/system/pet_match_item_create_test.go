package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestPetMatchGiveMedalUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := &PetMatchSystem{
		deps: &handler.Deps{
			Items:      testItemTable(t),
			ItemCreate: itemCreate,
			Log:        zap.NewNop(),
		},
	}
	player := &world.PlayerInfo{
		Name:    "pet-match-test",
		MapID:   petMatchMapIDs[0],
		Session: sess,
		Inv:     world.NewInventory(),
	}

	sys.giveMedal(player, 0, true)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != petMatchMedalID || itemCreate.count != 3 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=3", itemCreate.itemID, itemCreate.count, petMatchMedalID)
	}
}
