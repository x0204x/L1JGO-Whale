package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestCraftProduceItemsUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	player := &world.PlayerInfo{
		SessionID: 1,
		CharID:    1,
		Name:      "craft-test",
		Inv:       world.NewInventory(),
	}
	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewCraftSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	recipe := &data.CraftRecipe{
		SuccessRate: 100,
		Items: []data.CraftOutput{
			{ItemID: 1001, Amount: 3},
		},
	}

	sys.ExecuteCraft(sess, player, nil, recipe, 2)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1001 || itemCreate.count != 6 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1001 count=6", itemCreate.itemID, itemCreate.count)
	}
}
