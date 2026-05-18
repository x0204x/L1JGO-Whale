package system

import (
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestQuestGiveItemUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewQuestSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "quest-test",
		Inv:  world.NewInventory(),
	}

	sys.giveQuestItem(sess, player, 1003, 2)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1003 || itemCreate.count != 2 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1003 count=2", itemCreate.itemID, itemCreate.count)
	}
	if !player.Dirty {
		t.Fatal("任務給物品後應標記角色 Dirty")
	}
}

func TestQuestGiveGoldUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	itemCreate := &shopItemCreateStub{}
	sys := NewQuestSystem(&handler.Deps{
		Items:      testItemTable(t),
		ItemCreate: itemCreate,
		Log:        zap.NewNop(),
	})
	player := &world.PlayerInfo{
		Name: "quest-test",
		Inv:  world.NewInventory(),
	}

	sys.giveQuestGold(sess, player, 77)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 77 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=77", itemCreate.itemID, itemCreate.count, world.AdenaItemID)
	}
	if !player.Dirty {
		t.Fatal("任務給金幣後應標記角色 Dirty")
	}
}
