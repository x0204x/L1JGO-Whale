package handler

import (
	"encoding/binary"
	"testing"

	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// MISS-P1-009 求婚（C_Propose mode=0）對齊 Java 測試。
// 涵蓋四個 Java 行為差異：
//  1) 亡靈/死亡狀態靜默 return
//  2) 同性不可結婚 → S_ServerMessage(661)
//  3) 雙方等級總和 < 50 → S_SystemMessage 字串（非 661）
//  4) 教堂檢查置於最後，雙方任一不在教堂 → S_SystemMessage 字串

// newMarriageTestSetup 建立一對面對面、教堂內、互有戒指的玩家。
func newMarriageTestSetup(t *testing.T) (*Deps, *world.PlayerInfo, *world.PlayerInfo) {
	t.Helper()
	ws := world.NewState()

	a := &world.PlayerInfo{
		SessionID: 1,
		Session:   newHandlerTestSession(t, 1),
		CharID:    100,
		Name:      "alice",
		Sex:       1, // female
		Level:     30,
		MapID:     4,
		X:         33975,
		Y:         33363,
		Heading:   0, // dx=0, dy=1 → 對象在 (33975, 33364)
		Inv:       world.NewInventory(),
	}
	a.Inv.AddItem(40901, 1, "ring", 0, 0, false, 0)
	ws.AddPlayer(a)

	b := &world.PlayerInfo{
		SessionID: 2,
		Session:   newHandlerTestSession(t, 2),
		CharID:    200,
		Name:      "bob",
		Sex:       0, // male
		Level:     30,
		MapID:     4,
		X:         33975,
		Y:         33364, // 在 alice 朝向方向
		Inv:       world.NewInventory(),
	}
	b.Inv.AddItem(40902, 1, "ring", 0, 0, false, 0)
	ws.AddPlayer(b)

	deps := &Deps{World: ws, Log: zap.NewNop()}
	return deps, a, b
}

// 將 C_Propose 模式包成 packet.Reader。
func buildProposePacket(mode byte) *packet.Reader {
	body := []byte{mode}
	return packet.NewReader(append([]byte{0x00}, body...))
}

// MISS-P1-009: 成功路徑 → 目標收到 S_Message_YN(654)。
func TestHandleMarriageProposeSendsYesNoToTarget(t *testing.T) {
	deps, a, b := newMarriageTestSetup(t)

	HandleMarriage(a.Session, buildProposePacket(0), deps)

	if b.PendingYesNoType != 654 {
		t.Fatalf("目標應收到 PendingYesNoType=654，實際=%d", b.PendingYesNoType)
	}
	if b.PendingYesNoData != a.CharID {
		t.Fatalf("PendingYesNoData 應為求婚者 CharID=%d，實際=%d", a.CharID, b.PendingYesNoData)
	}
	if hasServerMessageInSession(a.Session, 661) {
		t.Fatalf("成功流程不應送 661 給求婚者")
	}
}

// MISS-P1-009: 同性求婚 → S_ServerMessage(661)，不送 Y/N 給對象。
func TestHandleMarriageProposeRejectsSameSex(t *testing.T) {
	deps, a, b := newMarriageTestSetup(t)
	b.Sex = a.Sex

	HandleMarriage(a.Session, buildProposePacket(0), deps)

	if !hasServerMessageInSession(a.Session, 661) {
		t.Fatalf("同性求婚應送 S_ServerMessage(661)")
	}
	if b.PendingYesNoType != 0 {
		t.Fatalf("同性求婚目標不應收到 Y/N，實際 PendingYesNoType=%d", b.PendingYesNoType)
	}
}

// MISS-P1-009: 等級總和 < 50 → S_SystemMessage 文字，而非 661。
func TestHandleMarriageProposeLevelSumUsesSystemMessage(t *testing.T) {
	deps, a, b := newMarriageTestSetup(t)
	a.Level = 10
	b.Level = 10

	HandleMarriage(a.Session, buildProposePacket(0), deps)

	pkts := drainHandlerTestPackets(a.Session)
	if !hasSystemMessage(pkts) {
		t.Fatalf("等級總和不足應送 S_SystemMessage（type=9），實際 packets=%d", len(pkts))
	}
	for _, p := range pkts {
		if isServerMessage(p, 661) {
			t.Fatalf("等級總和不足應送 S_SystemMessage 而非 S_ServerMessage 661")
		}
	}
}

// MISS-P1-009: 教堂外求婚 → S_SystemMessage 文字「必須在教堂中才能進行」。
// 玩家從一開始就放在教堂外（避免 AOI 索引與位置不同步）。
func TestHandleMarriageProposeOutOfChurchUsesSystemMessage(t *testing.T) {
	ws := world.NewState()
	a := &world.PlayerInfo{
		SessionID: 1, Session: newHandlerTestSession(t, 1),
		CharID: 100, Name: "alice", Sex: 1, Level: 30,
		MapID: 4, X: 32000, Y: 32000, Heading: 0,
		Inv: world.NewInventory(),
	}
	a.Inv.AddItem(40901, 1, "ring", 0, 0, false, 0)
	ws.AddPlayer(a)
	b := &world.PlayerInfo{
		SessionID: 2, Session: newHandlerTestSession(t, 2),
		CharID: 200, Name: "bob", Sex: 0, Level: 30,
		MapID: 4, X: 32000, Y: 32001,
		Inv: world.NewInventory(),
	}
	b.Inv.AddItem(40902, 1, "ring", 0, 0, false, 0)
	ws.AddPlayer(b)
	deps := &Deps{World: ws, Log: zap.NewNop()}

	HandleMarriage(a.Session, buildProposePacket(0), deps)

	pkts := drainHandlerTestPackets(a.Session)
	if !hasSystemMessage(pkts) {
		t.Fatalf("教堂外求婚應送 S_SystemMessage（type=9）")
	}
	for _, p := range pkts {
		if isServerMessage(p, 661) {
			t.Fatalf("教堂外應送 S_SystemMessage，不應送 S_ServerMessage 661")
		}
	}
}

// MISS-P1-009: 死亡時求婚應靜默 return（無 Y/N、無錯誤訊息）。
func TestHandleMarriageProposeWhenDeadSilent(t *testing.T) {
	deps, a, b := newMarriageTestSetup(t)
	a.Dead = true

	HandleMarriage(a.Session, buildProposePacket(0), deps)

	if b.PendingYesNoType != 0 {
		t.Fatalf("死亡狀態求婚不應觸發 Y/N")
	}
	pkts := drainHandlerTestPackets(a.Session)
	for _, p := range pkts {
		if isServerMessage(p, 661) || isServerMessage(p, 658) || isServerMessage(p, 659) || isServerMessage(p, 660) {
			t.Fatalf("死亡狀態求婚不應送出任何 S_ServerMessage")
		}
	}
}

// --- 封包工具 ---

// isServerMessage 檢測 S_ServerMessage 封包（opcode S_OPCODE_MESSAGE_CODE + writeH(msgID)）。
func isServerMessage(pkt []byte, msgID uint16) bool {
	if len(pkt) < 3 || pkt[0] != packet.S_OPCODE_MESSAGE_CODE {
		return false
	}
	return binary.LittleEndian.Uint16(pkt[1:3]) == msgID
}

// hasSystemMessage 檢測是否含 S_SystemMessage 封包（S_OPCODE_MESSAGE，第二 byte = type 9）。
// 字串本身已轉成客戶端編碼（Big5），無法用 UTF-8 needle 比對，僅檢查封包型別。
func hasSystemMessage(pkts [][]byte) bool {
	for _, p := range pkts {
		if len(p) >= 2 && p[0] == packet.S_OPCODE_MESSAGE && p[1] == 9 {
			return true
		}
	}
	return false
}

func hasServerMessageInSession(sess *l1net.Session, msgID uint16) bool {
	for _, p := range drainHandlerTestPackets(sess) {
		if isServerMessage(p, msgID) {
			return true
		}
	}
	return false
}

