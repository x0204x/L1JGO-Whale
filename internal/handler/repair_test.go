package handler

import (
	"encoding/binary"
	stdnet "net"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/data"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleFixWeaponListSendsEmptyListWhenNoDamagedWeapons(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "repair",
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	deps := &Deps{
		World:  ws,
		Items:  loadRepairTestItems(t),
		Config: &config.Config{Gameplay: config.GameplayConfig{RepairCostPerDurability: 200}},
		Log:    zap.NewNop(),
	}

	HandleFixWeaponList(sess, nil, deps)
	packets := drainHandlerTestPackets(sess)

	if len(packets) != 1 {
		t.Fatalf("沒有損壞武器時仍應送出空的 S_FixWeaponList，got packets=%d", len(packets))
	}
	pkt := packets[0]
	if len(pkt) < 8 || pkt[0] != packet.S_OPCODE_SELECTLIST {
		t.Fatalf("封包應為 S_OPCODE_SELECTLIST，packet=%v", pkt)
	}
	if price := int32(binary.LittleEndian.Uint32(pkt[1:5])); price != 200 {
		t.Fatalf("修復單價錯誤，got=%d", price)
	}
	if count := binary.LittleEndian.Uint16(pkt[5:7]); count != 0 {
		t.Fatalf("空清單 count 應為 0，got=%d", count)
	}
}

func newHandlerTestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 8, 8, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func drainHandlerTestPackets(sess *l1net.Session) [][]byte {
	sess.FlushOutput()
	var packets [][]byte
	for {
		select {
		case pkt := <-sess.OutQueue:
			packets = append(packets, pkt)
		default:
			return packets
		}
	}
}

func loadRepairTestItems(t *testing.T) *data.ItemTable {
	t.Helper()
	dir := t.TempDir()
	weaponPath := writeHandlerTestYAML(t, dir, "weapon_list.yaml", `
weapons:
  - item_id: 1
    name: sword
    type: sword
`)
	armorPath := writeHandlerTestYAML(t, dir, "armor_list.yaml", "armors: []\n")
	etcPath := writeHandlerTestYAML(t, dir, "etcitem_list.yaml", "items: []\n")
	table, err := data.LoadItemTable(weaponPath, armorPath, etcPath)
	if err != nil {
		t.Fatalf("載入測試物品失敗: %v", err)
	}
	return table
}

func writeHandlerTestYAML(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試 YAML 失敗: %v", err)
	}
	return path
}
