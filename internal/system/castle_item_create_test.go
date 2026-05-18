package system

import (
	stdnet "net"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestCastleWarGiftUsesItemCreate(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	warGifts := testWarGiftTable(t)
	ws := world.NewState()
	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	player := &world.PlayerInfo{
		SessionID: 1,
		CharID:    1,
		Name:      "castle-test",
		ClanID:    7,
		Session:   sess,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	itemCreate := &shopItemCreateStub{}
	sys := &CastleSystem{
		deps: &handler.Deps{
			World:      ws,
			WarGifts:   warGifts,
			Items:      testItemTable(t),
			ItemCreate: itemCreate,
			Log:        zap.NewNop(),
		},
	}

	sys.distributeWarGifts(1, &handler.CastleInfo{OwnerClanID: 7})

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != 1001 || itemCreate.count != 2 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=1001 count=2", itemCreate.itemID, itemCreate.count)
	}
}

func TestCastleWithdrawUsesItemCreateForAdena(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	sess := l1net.NewSession(server, 1, 4, 16, 0, zap.NewNop())
	player := &world.PlayerInfo{
		SessionID: 1,
		CharID:    1,
		Name:      "castle-test",
		ClanID:    7,
		ClanRank:  1,
		Session:   sess,
		Inv:       world.NewInventory(),
	}
	itemCreate := &shopItemCreateStub{}
	sys := &CastleSystem{
		deps: &handler.Deps{
			Items:      testItemTable(t),
			ItemCreate: itemCreate,
			Log:        zap.NewNop(),
		},
		castles: map[int32]*handler.CastleInfo{
			1: {CastleID: 1, OwnerClanID: 7, PublicMoney: 500},
		},
	}

	sys.Withdraw(sess, player, 1, 120)

	if itemCreate.calls != 1 {
		t.Fatalf("ItemCreate 呼叫次數錯誤：got %d want 1", itemCreate.calls)
	}
	if itemCreate.itemID != world.AdenaItemID || itemCreate.count != 120 {
		t.Fatalf("ItemCreate 參數錯誤：got item=%d count=%d want item=%d count=120", itemCreate.itemID, itemCreate.count, world.AdenaItemID)
	}
	if sys.castles[1].PublicMoney != 380 {
		t.Fatalf("城堡資金扣除錯誤：got %d want 380", sys.castles[1].PublicMoney)
	}
	if !player.Dirty {
		t.Fatal("領出金幣成功後應標記玩家 Dirty")
	}
}

func testWarGiftTable(t *testing.T) *data.WarGiftTable {
	t.Helper()
	path := filepath.Join(t.TempDir(), "war_gift.yaml")
	raw := []byte("gifts:\n  - castle_id: 1\n    items:\n      - item_id: 1001\n        count: 2\n")
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("寫入攻城禮物測試資料失敗：%v", err)
	}
	table, err := data.LoadWarGiftTable(path)
	if err != nil {
		t.Fatalf("載入攻城禮物測試資料失敗：%v", err)
	}
	return table
}
