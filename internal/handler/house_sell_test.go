package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	gonet "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/persist"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// MISS-P1-008 售屋（agsell）測試 — Java: C_NPCAction.sellHouse / S_SellHouse / C_Amount agsell
// 涵蓋 5 個關鍵分支：成功、非盟主、非君主、已上架、wrong keeper。

// auctionStub 記錄 AuctionManager 的呼叫，用於斷言上架資料。
type auctionStub struct {
	entries  map[int32]*persist.AuctionEntry
	createdN int
	lastSale *persist.AuctionEntry
	bids     map[int32]int64
}

func newAuctionStub() *auctionStub {
	return &auctionStub{entries: map[int32]*persist.AuctionEntry{}, bids: map[int32]int64{}}
}

func (a *auctionStub) GetEntriesForTown(_, _ int32) []*persist.AuctionEntry {
	var out []*persist.AuctionEntry
	for _, e := range a.entries {
		out = append(out, e)
	}
	return out
}
func (a *auctionStub) GetEntry(houseID int32) *persist.AuctionEntry { return a.entries[houseID] }
func (a *auctionStub) PlaceBid(_ *gonet.Session, _ *world.PlayerInfo, houseID int32, amount int64) bool {
	a.bids[houseID] = amount
	return true
}
func (a *auctionStub) IsAlreadyBidding(_ string) bool { return false }
func (a *auctionStub) CreateSale(e *persist.AuctionEntry) bool {
	if _, exists := a.entries[e.HouseID]; exists {
		return false
	}
	cp := *e
	a.entries[e.HouseID] = &cp
	a.lastSale = &cp
	a.createdN++
	return true
}

// 載入測試用 HouseTable，含一間 houseID=262145（奇岩）、keeper=50501。
func loadSellHouseTestHouses(t *testing.T) *data.HouseTable {
	t.Helper()
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "house_list.yaml")
	yamlContent := `houses:
  - house_id: 262145
    keeper_id: 50501
    map_id: 5000
    basement_map_id: 0
    home_x: 100
    home_y: 100
    x1: 95
    y1: 95
    x2: 105
    y2: 105
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("寫入 house_list.yaml 失敗: %v", err)
	}
	tbl, err := data.LoadHouseTable(yamlPath)
	if err != nil {
		t.Fatalf("LoadHouseTable: %v", err)
	}
	return tbl
}

func newSellHouseTestDeps(t *testing.T, auction AuctionManager) (*Deps, *world.State, *world.PlayerInfo, *world.ClanInfo) {
	t.Helper()
	ws := world.NewState()

	leader := &world.PlayerInfo{
		SessionID: 1,
		CharID:    9100,
		Name:      "leader",
		ClanID:    77,
		ClassType: 0, // Crown
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(leader)

	clan := &world.ClanInfo{
		ClanID:     77,
		ClanName:   "TestClan",
		LeaderID:   leader.CharID,
		LeaderName: leader.Name,
		HasHouse:   262145,
		Members: map[int32]*world.ClanMember{
			leader.CharID: {CharID: leader.CharID, CharName: leader.Name, Rank: 0},
		},
	}
	ws.Clans.AddClan(clan)

	deps := &Deps{
		World:   ws,
		Houses:  loadSellHouseTestHouses(t),
		Auction: auction,
		Log:     zap.NewNop(),
	}
	return deps, ws, leader, clan
}

func TestHandleHouseSellSendsS_SellHouseWhenValid(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess

	ok := handleHouseSell(sess, player, 999, 50501, deps)
	if !ok {
		t.Fatal("agsell 動作應該被處理（return true）")
	}
	if player.PendingSellHouseID != 262145 {
		t.Fatalf("PendingSellHouseID 應設為 262145，實際=%d", player.PendingSellHouseID)
	}
	pkts := drainHandlerTestPackets(sess)
	foundInputAmount := false
	for _, p := range pkts {
		if len(p) > 0 && p[0] == packet.S_OPCODE_INPUTAMOUNT {
			foundInputAmount = true
			if !containsBytes(p, []byte("agsell")) {
				t.Fatalf("S_SellHouse 必須帶 \"agsell\" htmlid，封包=%v", p)
			}
		}
	}
	if !foundInputAmount {
		t.Fatalf("應送出 S_OPCODE_INPUTAMOUNT 封包")
	}
}

func TestHandleHouseSellRejectsNonLeader(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess
	player.CharID = 9999 // 非盟主

	handleHouseSell(sess, player, 999, 50501, deps)

	if player.PendingSellHouseID != 0 {
		t.Fatalf("非盟主不應設 PendingSellHouseID，實際=%d", player.PendingSellHouseID)
	}
	pkts := drainHandlerTestPackets(sess)
	for _, p := range pkts {
		if len(p) > 0 && p[0] == packet.S_OPCODE_INPUTAMOUNT {
			t.Fatalf("非盟主不應收到 S_SellHouse")
		}
	}
}

func TestHandleHouseSellRejectsNonCrown(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess
	player.ClassType = 1 // Knight，非 Crown

	handleHouseSell(sess, player, 999, 50501, deps)

	if player.PendingSellHouseID != 0 {
		t.Fatalf("非君主不應設 PendingSellHouseID，實際=%d", player.PendingSellHouseID)
	}
	for _, p := range drainHandlerTestPackets(sess) {
		if len(p) > 0 && p[0] == packet.S_OPCODE_INPUTAMOUNT {
			t.Fatalf("非君主不應收到 S_SellHouse")
		}
	}
}

func TestHandleHouseSellWrongKeeperNoOp(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess

	// 用一個對不上自己小屋 keeperID 的 NPC（50531 不在 HouseTable 中）
	handleHouseSell(sess, player, 999, 50531, deps)

	if player.PendingSellHouseID != 0 {
		t.Fatalf("非自家 keeper 不應設 PendingSellHouseID")
	}
}

func TestHandleHouseSellAlreadyOnSaleSendsAgonsale(t *testing.T) {
	stub := newAuctionStub()
	stub.entries[262145] = &persist.AuctionEntry{HouseID: 262145, Price: 500_000}
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess

	handleHouseSell(sess, player, 999, 50501, deps)

	if player.PendingSellHouseID != 0 {
		t.Fatalf("已上架時不應設 PendingSellHouseID")
	}
	pkts := drainHandlerTestPackets(sess)
	hasHypertext := false
	for _, p := range pkts {
		if len(p) > 0 && p[0] == packet.S_OPCODE_HYPERTEXT && containsBytes(p, []byte("agonsale")) {
			hasHypertext = true
		}
		if len(p) > 0 && p[0] == packet.S_OPCODE_INPUTAMOUNT {
			t.Fatalf("已上架時不應發 S_SellHouse")
		}
	}
	if !hasHypertext {
		t.Fatalf("應送出 \"agonsale\" hypertext")
	}
}

// MISS-P1-008: C_Amount 回覆 "agsell 262145" + price → AuctionSystem.CreateSale 被呼叫。
func TestHandleSellHouseAmountCreatesSale(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess
	player.PendingSellHouseID = 262145

	// Java C_Amount: [D npcObjID][D amount][C unknown][S actionStr]
	body := buildSellHouseAmountPacket(t, 9999, 500_000, "agsell 262145")
	r := packet.NewReader(append([]byte{packet.S_OPCODE_INPUTAMOUNT}, body...))

	HandleSellHouseAmount(sess, r, player, deps)

	if stub.createdN != 1 {
		t.Fatalf("應建立 1 筆拍賣，實際=%d", stub.createdN)
	}
	if player.PendingSellHouseID != 0 {
		t.Fatalf("處理完應清掉 PendingSellHouseID")
	}
	if stub.lastSale.HouseID != 262145 || stub.lastSale.OldOwner != "leader" || stub.lastSale.Price != 500_000 {
		t.Fatalf("拍賣資料錯誤: %+v", stub.lastSale)
	}
	if stub.lastSale.Location != "奇岩" {
		t.Fatalf("Location 應從 houseID 區段推導為 奇岩，實際=%q", stub.lastSale.Location)
	}
	if stub.lastSale.HouseArea != 121 { // (105-95+1)*(105-95+1) = 121
		t.Fatalf("HouseArea 應由 X1..Y2 計算為 121，實際=%d", stub.lastSale.HouseArea)
	}
	// 截止日 5 天後（允許 ±1 分鐘誤差）
	expected := sellHouseDeadline(time.Now())
	if diff := stub.lastSale.Deadline.Sub(expected); diff < -time.Minute || diff > time.Minute {
		t.Fatalf("Deadline 應約 5 天後（hour-truncated），實際 diff=%v", diff)
	}
}

func TestHandleSellHouseAmountRejectsOutOfRange(t *testing.T) {
	stub := newAuctionStub()
	deps, _, player, _ := newSellHouseTestDeps(t, stub)
	sess := newHandlerTestSession(t, 1)
	player.Session = sess
	player.PendingSellHouseID = 262145

	body := buildSellHouseAmountPacket(t, 9999, 50_000, "agsell 262145") // 小於 100,000
	r := packet.NewReader(append([]byte{packet.S_OPCODE_INPUTAMOUNT}, body...))

	HandleSellHouseAmount(sess, r, player, deps)
	if stub.createdN != 0 {
		t.Fatalf("低於 100,000 不應建立拍賣")
	}
}

// 輔助函式：手動組合 C_Amount 封包 body。
func buildSellHouseAmountPacket(t *testing.T, npcObjID int32, amount int32, actionStr string) []byte {
	t.Helper()
	w := packet.NewWriterWithOpcode(packet.S_OPCODE_INPUTAMOUNT)
	w.WriteD(npcObjID)
	w.WriteD(amount)
	w.WriteC(0)
	w.WriteS(actionStr)
	bytes := w.Bytes()
	return bytes[1:] // 去掉 opcode（NewReader 會跳過第 0 byte）
}

func containsBytes(haystack, needle []byte) bool {
	return strings.Contains(string(haystack), string(needle))
}
