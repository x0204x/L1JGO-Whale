package handler

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

type captureSkillManager struct {
	reqs                       []SkillRequest
	cancelAbsoluteBarrierCalls int
}

func (m *captureSkillManager) QueueSkill(req SkillRequest) {
	m.reqs = append(m.reqs, req)
}

func (m *captureSkillManager) CancelAllBuffs(_ *world.PlayerInfo) {}

func (m *captureSkillManager) ClearAllBuffsOnDeath(_ *world.PlayerInfo) {}

func (m *captureSkillManager) GMClearAllStatuses(_ *world.PlayerInfo) {}

func (m *captureSkillManager) TickPlayerBuffs(_ *world.PlayerInfo) {}

func (m *captureSkillManager) RemoveBuffAndRevert(_ *world.PlayerInfo, _ int32) {}

func (m *captureSkillManager) ApplyNpcDebuff(_ *world.PlayerInfo, _ *data.SkillInfo) {}

func (m *captureSkillManager) CancelAbsoluteBarrier(_ *world.PlayerInfo) {
	m.cancelAbsoluteBarrierCalls++
}

func (m *captureSkillManager) CancelInvisibility(_ *world.PlayerInfo) {}

func (m *captureSkillManager) ApplyGMBuff(_ *world.PlayerInfo, _ int32) bool { return false }

func (m *captureSkillManager) RevertBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

func (m *captureSkillManager) ConsumeSkillResources(_ *net.Session, _ *world.PlayerInfo, _ *data.SkillInfo) {
}

func (m *captureSkillManager) ApplyBuffStats(_ *world.PlayerInfo, _ *world.ActiveBuff) {}

func TestHandleUseSpellBlockedByShockStunBeforeSkillQueueLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "stunned",
		Paralyzed: true,
		Inv:       world.NewInventory(),
	}
	player.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會以 isParalyzedX 擋下 SHOCK_STUN 中的施法，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
	if !hasHandlerServerMessage(drainHandlerTestPackets(sess), 285) {
		t.Fatal("Java C_UseSkill 在 isParalyzedX 擋下施法時會送 S_ServerMessage(285)")
	}
}

// Java `C_UseSkill.start()` 第 106-108 行 `if (pc.isTeleport()) return;` 在 isDead
// 與 isPrivateShop 之間，傳送待確認狀態（已預備好 C_TELEPORT 確認）下任何技能
// 都會被靜默阻擋（無訊息回饋，與 isSkillDelay 對齊）。鎖定 Go `HandleUseSpell`
// 在 `player.HasTeleport=true` 時不排入 SkillQueue，避免玩家在傳送預備中仍施放
// SHOCK_STUN。
func TestHandleUseSpellBlockedByPendingTeleportLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID:   sess.ID,
		Session:     sess,
		CharID:      1001,
		Name:        "teleporter",
		HasTeleport: true,
		Inv:         world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會在 pc.isTeleport() 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellBlockedByPrivateShopLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID:   sess.ID,
		Session:     sess,
		CharID:      1001,
		Name:        "seller",
		PrivateShop: true,
		Inv:         world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會在 pc.isPrivateShop() 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellBlockedByDeadPlayerLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "dead",
		Dead:      true,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會在 pc.isDead() 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellBlockedBySkillDelayLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID:       sess.ID,
		Session:         sess,
		CharID:          1001,
		Name:            "delayed",
		SkillDelayUntil: time.Now().Add(time.Second),
		Inv:             world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 會在 pc.isSkillDelay() 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellBlockedByOverweightLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "heavy",
		Str:       1,
		Con:       1,
		Inv:       world.NewInventory(),
	}
	player.Inv.AddItem(40001, 2500, "heavy item", 0, 1000, true, 0)
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 在 getWeight240 >= 197 時會拒絕施法，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
	if !hasHandlerServerMessage(drainHandlerTestPackets(sess), 316) {
		t.Fatal("Java C_UseSkill 負重過高時應送 S_ServerMessage(316)")
	}
}

func TestHandleUseSpellBlockedByUnusableSkillMapLikeJava(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "noskillmap",
		MapID:     9900,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{
		World:   ws,
		Skill:   mgr,
		MapData: newHandlerSkillMapData(t, player.MapID, false),
	}

	HandleUseSpell(sess, useSpellReader(87), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 在 !pc.getMap().isUsableSkill() 時會拒絕施法，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
	if !hasHandlerServerMessage(drainHandlerTestPackets(sess), 563) {
		t.Fatal("Java C_UseSkill 在不可使用技能地圖應送 S_ServerMessage(563)")
	}
}

func TestHandleUseSpellRejectsSkillIDOverJavaMax(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID: sess.ID,
		Session:   sess,
		CharID:    1001,
		Name:      "invalidskill",
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(player)

	mgr := &captureSkillManager{}
	deps := &Deps{World: ws, Skill: mgr}

	HandleUseSpell(sess, useSpellRawReader(30, 0), deps)

	if len(mgr.reqs) != 0 {
		t.Fatalf("Java C_UseSkill 在 skillId > 239 時直接返回，不應排入 SkillQueue，reqs=%+v", mgr.reqs)
	}
}

func TestHandleUseSpellParsesTeleportBookmark(t *testing.T) {
	req := parseUseSpellForTest(69, func(w *packet.Writer) {
		w.WriteH(4)
		w.WriteD(98765)
	})

	if req.TargetID != 98765 {
		t.Fatalf("TargetID = %d, want bookmark id 98765", req.TargetID)
	}
	if req.BookmarkID != 98765 {
		t.Fatalf("BookmarkID = %d, want 98765", req.BookmarkID)
	}
	if req.MapID != 4 {
		t.Fatalf("MapID = %d, want 4", req.MapID)
	}
}

func TestHandleUseSpellParsesTrueTargetTextAndPosition(t *testing.T) {
	req := parseUseSpellForTest(113, func(w *packet.Writer) {
		w.WriteD(1234)
		w.WriteH(33010)
		w.WriteH(33011)
		w.WriteS("mark")
	})

	if req.TargetID != 1234 {
		t.Fatalf("TargetID = %d, want 1234", req.TargetID)
	}
	if req.TargetX != 33010 || req.TargetY != 33011 {
		t.Fatalf("Target position = (%d,%d), want (33010,33011)", req.TargetX, req.TargetY)
	}
	if req.Text != "mark" {
		t.Fatalf("Text = %q, want %q", req.Text, "mark")
	}
}

func TestHandleUseSpellParsesGroundTargetPosition(t *testing.T) {
	req := parseUseSpellForTest(58, func(w *packet.Writer) {
		w.WriteH(32768)
		w.WriteH(32769)
	})

	if req.TargetX != 32768 || req.TargetY != 32769 {
		t.Fatalf("Target position = (%d,%d), want (32768,32769)", req.TargetX, req.TargetY)
	}
}

func TestHandleUseSpellParsesClanTargetName(t *testing.T) {
	req := parseUseSpellForTest(116, func(w *packet.Writer) {
		w.WriteS("Alice[Clan]")
	})

	if req.TargetName != "Alice" {
		t.Fatalf("TargetName = %q, want %q", req.TargetName, "Alice")
	}
}

func TestHandleUseSpellPreservesSummonSelectionValue(t *testing.T) {
	req := parseUseSpellForTest(51, func(w *packet.Writer) {
		w.WriteD(812)
	})

	if req.TargetID != 812 {
		t.Fatalf("TargetID = %d, want 812", req.TargetID)
	}
	if req.SummonID != 812 {
		t.Fatalf("SummonID = %d, want 812", req.SummonID)
	}
}

func parseUseSpellForTest(skillID int32, writeRest func(*packet.Writer)) SkillRequest {
	row := byte((skillID - 1) / 8)
	column := byte((skillID - 1) % 8)

	w := packet.NewWriterWithOpcode(6)
	w.WriteC(row)
	w.WriteC(column)
	writeRest(w)

	mgr := &captureSkillManager{}
	deps := &Deps{Skill: mgr}
	sess := &net.Session{ID: 777}
	HandleUseSpell(sess, packet.NewReader(w.RawBytes()), deps)

	if len(mgr.reqs) != 1 {
		panic("expected exactly one queued skill request")
	}
	return mgr.reqs[0]
}

func useSpellReader(skillID int32) *packet.Reader {
	row := byte((skillID - 1) / 8)
	column := byte((skillID - 1) % 8)

	return useSpellRawReader(row, column)
}

func useSpellRawReader(row, column byte) *packet.Reader {
	w := packet.NewWriterWithOpcode(packet.C_OPCODE_USE_SPELL)
	w.WriteC(row)
	w.WriteC(column)
	w.WriteD(0)
	w.WriteH(0)
	w.WriteH(0)
	return packet.NewReader(w.RawBytes())
}

func newHandlerSkillMapData(t *testing.T, mapID int16, usableSkill bool) *data.MapDataTable {
	t.Helper()

	dir := t.TempDir()
	tileDir := filepath.Join(dir, "maps")
	if err := os.Mkdir(tileDir, 0o755); err != nil {
		t.Fatalf("建立測試地圖目錄失敗: %v", err)
	}

	yamlPath := filepath.Join(dir, "map_list.yaml")
	yaml := fmt.Sprintf(`maps:
  - map_id: %d
    name: test
    start_x: 0
    end_x: 0
    start_y: 0
    end_y: 0
    usable_skill: %t
`, mapID, usableSkill)
	if err := os.WriteFile(yamlPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("寫入測試 map_list.yaml 失敗: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tileDir, fmt.Sprintf("%d.txt", mapID)), []byte("0\n"), 0o644); err != nil {
		t.Fatalf("寫入測試地圖 tile 失敗: %v", err)
	}

	maps, err := data.LoadMapData(yamlPath, tileDir)
	if err != nil {
		t.Fatalf("載入測試地圖失敗: %v", err)
	}
	if maps.GetInfo(mapID) == nil {
		t.Fatalf("測試地圖 %d 未載入", mapID)
	}
	return maps
}

func hasHandlerServerMessage(packets [][]byte, msgID uint16) bool {
	for _, pkt := range packets {
		if len(pkt) >= 3 && pkt[0] == packet.S_OPCODE_MESSAGE_CODE &&
			binary.LittleEndian.Uint16(pkt[1:3]) == msgID {
			return true
		}
	}
	return false
}
