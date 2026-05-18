package system

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
)

func attachShockStunItemTable(t *testing.T, s *SkillSystem) {
	t.Helper()
	items, err := data.LoadItemTable(
		filepath.Join("..", "..", "data", "yaml", "weapon_list.yaml"),
		filepath.Join("..", "..", "data", "yaml", "armor_list.yaml"),
		filepath.Join("..", "..", "data", "yaml", "etcitem_list.yaml"),
	)
	if err != nil {
		t.Fatalf("讀取物品資料失敗: %v", err)
	}
	s.deps.Items = items
}

func attachShockStunNpcTable(t *testing.T, s *SkillSystem) {
	t.Helper()
	npcs, err := data.LoadNpcTable(filepath.Join("..", "..", "data", "yaml", "npc_list.yaml"))
	if err != nil {
		t.Fatalf("讀取 NPC 資料失敗: %v", err)
	}
	s.deps.Npcs = npcs
}

func newShockStunSafetyMap(t *testing.T) *data.MapDataTable {
	t.Helper()
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "map_list.yaml")
	tileDir := filepath.Join(dir, "tiles")
	if err := os.Mkdir(tileDir, 0o755); err != nil {
		t.Fatalf("建立測試地圖目錄失敗: %v", err)
	}
	yaml := []byte(`maps:
  - map_id: 901
    name: shock_stun_safety_test
    start_x: 100
    end_x: 101
    start_y: 100
    end_y: 100
`)
	if err := os.WriteFile(yamlPath, yaml, 0o644); err != nil {
		t.Fatalf("寫入測試地圖清單失敗: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tileDir, "901.txt"), []byte("15,31\n"), 0o644); err != nil {
		t.Fatalf("寫入測試地圖格資料失敗: %v", err)
	}
	maps, err := data.LoadMapData(yamlPath, tileDir)
	if err != nil {
		t.Fatalf("載入測試地圖失敗: %v", err)
	}
	return maps
}

type shockStunPvPSpy struct {
	called   int
	attacker *world.PlayerInfo
	target   *world.PlayerInfo
}

func (s *shockStunPvPSpy) HandlePvPAttack(_, _ *world.PlayerInfo) {}

func (s *shockStunPvPSpy) HandlePvPFarAttack(_, _ *world.PlayerInfo) {}

func (s *shockStunPvPSpy) AddLawfulFromNpc(_ *world.PlayerInfo, _ int32) {}

func (s *shockStunPvPSpy) TriggerPinkName(attacker, target *world.PlayerInfo) {
	s.called++
	s.attacker = attacker
	s.target = target
}

type shockStunCombatSpy struct {
	requests []handler.AttackRequest
}

func (s *shockStunCombatSpy) QueueAttack(req handler.AttackRequest) {
	s.requests = append(s.requests, req)
}

func (s *shockStunCombatSpy) HandleNpcDeath(_ *world.NpcInfo, _ *world.PlayerInfo, _ []*world.PlayerInfo) *handler.NpcKillResult {
	return nil
}

func (s *shockStunCombatSpy) AddExp(_ *world.PlayerInfo, _ int32) {}

func hasGlobalSystemMessageText(packets [][]byte, text string) bool {
	want := append(packet.EncodeString(text), 0)
	for _, pkt := range packets {
		for i := 0; i+2+len(want) <= len(pkt); i++ {
			if pkt[i] == packet.S_OPCODE_MESSAGE && pkt[i+1] == 9 && string(pkt[i+2:i+2+len(want)]) == string(want) {
				return true
			}
		}
	}
	return false
}

func hasNormalChatText(packets [][]byte, text string) bool {
	want := append(packet.EncodeString(text), 0)
	for _, pkt := range packets {
		if len(pkt) >= 6+len(want) && pkt[0] == packet.S_OPCODE_SAY && pkt[1] == handler.ChatNormal &&
			string(pkt[6:6+len(want)]) == string(want) {
			return true
		}
	}
	return false
}

func hasShockStunGMDurationMessage(packets [][]byte) bool {
	for seconds := 1; seconds <= 5; seconds++ {
		if hasNormalChatText(packets, fmt.Sprintf("此次衝暈秒數為%d秒..只有GM看的到", seconds)) {
			return true
		}
	}
	return false
}

func hasActionGfxPacket(packets [][]byte, objectID int32, actionCode byte) bool {
	for _, pkt := range packets {
		if len(pkt) < 6 || pkt[0] != packet.S_OPCODE_ACTION {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID && pkt[5] == actionCode {
			return true
		}
	}
	return false
}

func countSkillEffectPackets(packets [][]byte, objectID int32, gfxID int32) int {
	count := 0
	for _, pkt := range packets {
		if len(pkt) < 7 || pkt[0] != packet.S_OPCODE_EFFECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID &&
			int32(binary.LittleEndian.Uint16(pkt[5:7])) == gfxID {
			count++
		}
	}
	return count
}

func TestSkillClanShockStunDurationMatchesJavaConfigRange(t *testing.T) {
	rand.Seed(1)

	for i := 0; i < 100; i++ {
		dur := shockStunDurationSeconds()
		if dur < 1 || dur > 5 {
			t.Fatalf("Java SHOCK_STUN_TIMER=1~5，暈眩秒數不可超出範圍，got=%d", dur)
		}
	}
}

func TestSkillClanShockStunGmReceivesJavaDurationMessage(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm",
		AccessLevel: 200,
		X:           100,
		Y:           100,
		MapID:       4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "target",
		X:          101,
		Y:          100,
		MapID:      4,
		RegistStun: -100,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if !hasShockStunGMDurationMessage(drainSkillTestPackets(caster.Session)) {
		t.Fatal("Java SHOCK_STUN 成功時 GM 會收到 S_ServerMessage(String) 顯示此次衝暈秒數")
	}
}

func TestSkillClanShockStunGmReceivesJavaDurationMessageForNpcTarget(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm",
		AccessLevel: 200,
		Level:       60,
		X:           100,
		Y:           100,
		MapID:       4,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	for seed := int64(1); seed <= 100; seed++ {
		rand.Seed(seed)
		npc.Paralyzed = false
		npc.ActiveDebuffs = nil
		_ = drainSkillTestPackets(caster.Session)

		s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)
		if !npc.HasDebuff(87) {
			continue
		}
		if !hasShockStunGMDurationMessage(drainSkillTestPackets(caster.Session)) {
			t.Fatal("Java SHOCK_STUN 玩家施放到 NPC 成功時，GM 也會收到 S_ServerMessage(String) 顯示此次衝暈秒數")
		}
		return
	}
	t.Fatal("測試種子 1..100 未觸發 SHOCK_STUN NPC 成功，無法驗證 GM 秒數訊息")
}

func TestSkillClanShockStunNpcExistingDebuffDoesNotRefreshLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm",
		AccessLevel: 200,
		Level:       60,
		X:           100,
		Y:           100,
		MapID:       4,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}
	npc.Paralyzed = true
	npc.AddDebuff(87, 999)

	for seed := int64(1); seed <= 100; seed++ {
		rand.Seed(seed)
		_ = drainSkillTestPackets(caster.Session)

		s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)
		if got := npc.ActiveDebuffs[87]; got != 999 {
			t.Fatalf("Java SHOCK_STUN 目標已有 87 效果時不刷新 NPC debuff，got ticks=%d", got)
		}
		packets := drainSkillTestPackets(caster.Session)
		if hasShockStunGMDurationMessage(packets) {
			t.Fatal("Java SHOCK_STUN 目標已有 87 效果時不會重送 GM 秒數訊息")
		}
		if !hasSkillEffectPacket(packets, npc.ID, 4434) {
			t.Fatal("Java SHOCK_STUN 目標已有 87 效果時仍會由 L1SkillUse 送目標 S_SkillSound(4434)")
		}
		if effects := ws.GetNearbyGroundEffects(npc.X, npc.Y, npc.MapID); len(effects) != 0 {
			t.Fatalf("Java SHOCK_STUN 目標已有 87 效果時不會再次 spawnEffect(81162)，got effects=%d", len(effects))
		}
	}
}

func TestSkillClanShockStunPlayerExistingBuffDoesNotRefreshButSendsJavaTargetGfx(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm",
		AccessLevel: 200,
		Level:       60,
		X:           100,
		Y:           100,
		MapID:       4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "target",
		Level:      50,
		X:          101,
		Y:          100,
		MapID:      4,
		RegistStun: -100,
		Paralyzed:  true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 999, SetParalyzed: true})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	_ = drainSkillTestPackets(target.Session)
	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if got := target.GetBuff(87).TicksLeft; got != 999 {
		t.Fatalf("Java SHOCK_STUN 目標已有 87 效果時不刷新玩家 buff，got ticks=%d", got)
	}
	packets := drainSkillTestPackets(target.Session)
	if hasShockStunGMDurationMessage(drainSkillTestPackets(caster.Session)) {
		t.Fatal("Java SHOCK_STUN 目標已有 87 效果時不會重送 GM 秒數訊息")
	}
	if hasActionGfxPacket(packets, caster.CharID, 19) {
		t.Fatal("Java SHOCK_STUN 目標已有 87 效果時只送目標 S_SkillSound(4434)，不送施法者 S_DoActionGFX")
	}
	if !hasSkillEffectPacket(packets, target.CharID, 4434) {
		t.Fatal("Java SHOCK_STUN 目標已有 87 效果時仍會由 L1SkillUse 送目標 S_SkillSound(4434)")
	}
	if effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID); len(effects) != 0 {
		t.Fatalf("Java SHOCK_STUN 目標已有 87 效果時不會再次 spawnEffect(81162)，got effects=%d", len(effects))
	}
}

func TestSkillClanShockStunPlayerCasterDoesNotParalyzeGuardNpcLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		Level:     60,
		X:         100,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 70500,
		Impl:  "L1Guard",
		Name:  "guard target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	for seed := int64(1); seed <= 100; seed++ {
		rand.Seed(seed)
		npc.Paralyzed = false
		npc.ActiveDebuffs = nil
		for _, effect := range ws.GroundEffectList() {
			ws.RemoveGroundEffect(effect.ID)
		}

		s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)
		if !npc.HasDebuff(87) {
			continue
		}
		if npc.Paralyzed {
			t.Fatal("Java 玩家施放 SHOCK_STUN 到 L1Guard 會套 87 效果與 81162，但不會 setParalyzed(true)")
		}
		return
	}
	t.Fatal("測試種子 1..100 未觸發 SHOCK_STUN L1Guard 成功，無法驗證玩家施放 NPC 類別邊界")
}

func TestSkillClanShockStunIntReducesMpConsumeLikeJava(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		Level:       60,
		Intel:       18,
		X:           100,
		Y:           100,
		MapID:       4,
		MP:          9,
		MaxMP:       9,
		KnownSpells: []int32{87},
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		Level:     1,
		X:         101,
		Y:         100,
		MapID:     4,
	})
	weapon := caster.Inv.AddItemWithID(7001, 52, 1, "雙手劍", 0, 1000, false, 1)
	weapon.Equipped = true
	caster.Equip.Set(world.SlotWeapon, weapon)

	s := newSkillBuffTestSystem(t, ws)
	attachShockStunItemTable(t, s)

	s.processSkill(handler.SkillRequest{
		SessionID: caster.SessionID,
		SkillID:   87,
		TargetID:  target.CharID,
	})

	if caster.MP != 0 {
		t.Fatalf("Java 對 SHOCK_STUN 會以 INT 18 將 MP 15 扣減 6，應消耗 9 MP，MP=%d", caster.MP)
	}
}

func TestSkillClanShockStunPlayerQueuesJavaOnActionMelee(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
	})
	s := newSkillTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if len(combat.requests) != 1 {
		t.Fatalf("Java L1SkillUse 會在 SHOCK_STUN 目標計算前呼叫 _target.onAction(_player)，got requests=%d", len(combat.requests))
	}
	req := combat.requests[0]
	if req.AttackerSessionID != caster.SessionID || req.TargetID != target.CharID || !req.IsMelee {
		t.Fatalf("SHOCK_STUN onAction 應排入近戰攻擊，request=%+v", req)
	}
}

func TestSkillClanShockStunInvisibleCasterBlockedLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "invis-knight",
		X:           100,
		Y:           100,
		MapID:       4,
		Level:       60,
		HP:          200,
		MaxHP:       200,
		MP:          100,
		MaxMP:       100,
		Invisible:   true,
		KnownSpells: []int32{87},
	})
	caster.AddBuff(&world.ActiveBuff{SkillID: 60, TicksLeft: 100, SetInvisible: true})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
	})
	s := newSkillBuffTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})

	s.processSkill(handler.SkillRequest{
		SessionID: caster.SessionID,
		SkillID:   87,
		TargetID:  target.CharID,
	})

	if !caster.Invisible || !caster.HasBuff(60) {
		t.Fatalf("Java C_UseSkill 會先以 isInvisUsableSkill 擋下 SHOCK_STUN，不應解除隱身，Invisible=%v buff60=%v", caster.Invisible, caster.GetBuff(60))
	}
	if len(combat.requests) != 0 {
		t.Fatalf("隱身狀態下 SHOCK_STUN 應被 C_UseSkill 拒絕，不應觸發 onAction，requests=%d", len(combat.requests))
	}
	if !hasServerMessage(drainSkillTestPackets(caster.Session), 1003) {
		t.Fatal("Java C_UseSkill 對隱身不可施放技能會回覆 S_ServerMessage(1003)")
	}
}

func TestSkillClanShockStunNpcQueuesJavaOnActionMelee(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, npc.ID)

	if len(combat.requests) != 1 {
		t.Fatalf("Java L1SkillUse 對 NPC SHOCK_STUN 也會呼叫 _target.onAction(_player)，got requests=%d", len(combat.requests))
	}
	req := combat.requests[0]
	if req.AttackerSessionID != caster.SessionID || req.TargetID != npc.ID || !req.IsMelee {
		t.Fatalf("SHOCK_STUN NPC onAction 應排入近戰攻擊，request=%+v", req)
	}
}

func TestSkillClanShockStunPlayerTriggersPinkNameLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "target",
		X:          101,
		Y:          100,
		MapID:      4,
		RegistStun: -100,
	})
	s := newSkillTestSystem(t, ws)
	pvp := &shockStunPvPSpy{}
	s.deps.PvP = pvp
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if pvp.called != 1 || pvp.attacker != caster || pvp.target != target {
		t.Fatalf("Java SHOCK_STUN 成功時會 L1PinkName.onAction(target, caster)，called=%d attacker=%v target=%v", pvp.called, pvp.attacker, pvp.target)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaImpactHaloRate(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java SHOCK_STUN 高等玩家成功率基準為 IMPACT_HALO_1=40，不應使用 Go 一般公式 80%%，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
}

func TestSkillClanShockStunPlayerTargetSafetyZoneBlocksLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     901,
		Level:     80,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "safe-target",
		X:          101,
		Y:          100,
		MapID:      901,
		Level:      1,
		RegistStun: -100,
	})
	s := newSkillTestSystem(t, ws)
	s.deps.MapData = newShockStunSafetyMap(t)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java SHOCK_STUN 在安全區應被 checkZone 擋下，不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	packets := drainSkillTestPackets(caster.Session)
	if !hasGlobalSystemMessageText(packets, "在安全區域無法使用此技能。") {
		t.Fatal("Java SHOCK_STUN 安全區拒絕時應送 S_SystemMessage(\"在安全區域無法使用此技能。\")")
	}
	if countSkillEffectPackets(drainSkillTestPackets(target.Session), target.CharID, 4434) != 0 {
		t.Fatal("Java SHOCK_STUN 安全區拒絕時 targetList 為空，不應送目標 4434")
	}
}

func TestSkillClanShockStunPlayerTargetRangeMatchesJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "far-target",
		X:          102,
		Y:          100,
		MapID:      4,
		Level:      1,
		Sleeped:    true,
		RegistStun: -100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 66, TicksLeft: 100, SetSleeped: true})
	target.AddBuff(&world.ActiveBuff{SkillID: 153, TicksLeft: 100})
	s := newSkillTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434, Ranged: 1}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java L1SkillUse 會以 ranged=1 擋下遠距 SHOCK_STUN，不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if len(combat.requests) != 0 {
		t.Fatalf("射程外 SHOCK_STUN 不應進入 runSkill 觸發 onAction，requests=%d", len(combat.requests))
	}
	if !target.Sleeped || !target.HasBuff(66) || !target.HasBuff(153) {
		t.Fatalf("射程外 SHOCK_STUN 不應清除目標既有狀態，Sleeped=%v buff66=%v buff153=%v", target.Sleeped, target.HasBuff(66), target.HasBuff(153))
	}
	if countSkillEffectPackets(drainSkillTestPackets(target.Session), target.CharID, 4434) != 0 {
		t.Fatal("射程外 SHOCK_STUN 不應送目標 4434")
	}
}

func TestSkillClanShockStunNpcTargetRangeMatchesJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45000,
		Impl:          "L1Monster",
		Name:          "far-npc",
		X:             102,
		Y:             100,
		MapID:         4,
		Level:         1,
		Sleeped:       true,
		ActiveDebuffs: map[int32]int{66: 100, 153: 100},
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434, Ranged: 1}

	s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)

	if npc.HasDebuff(87) || npc.Paralyzed {
		t.Fatalf("Java L1SkillUse 會以 ranged=1 擋下遠距 NPC 目標 SHOCK_STUN，不應套用 87，debuff=%v Paralyzed=%v", npc.HasDebuff(87), npc.Paralyzed)
	}
	if len(combat.requests) != 0 {
		t.Fatalf("射程外 NPC 目標 SHOCK_STUN 不應進入 runSkill 觸發 onAction，requests=%d", len(combat.requests))
	}
	if !npc.Sleeped || !npc.HasDebuff(66) || !npc.HasDebuff(153) {
		t.Fatalf("射程外 NPC 目標 SHOCK_STUN 不應清除既有狀態，Sleeped=%v debuff66=%v debuff153=%v", npc.Sleeped, npc.HasDebuff(66), npc.HasDebuff(153))
	}
}

func TestSkillClanShockStunRangeFailureDoesNotConsumeMpLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "knight",
		X:           100,
		Y:           100,
		MapID:       4,
		Level:       80,
		MP:          15,
		MaxMP:       15,
		KnownSpells: []int32{87},
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "far-target",
		X:         102,
		Y:         100,
		MapID:     4,
		Level:     1,
	})
	weapon := caster.Inv.AddItemWithID(7001, 52, 1, "雙手劍", 0, 1000, false, 1)
	weapon.Equipped = true
	caster.Equip.Set(world.SlotWeapon, weapon)
	s := newSkillBuffTestSystem(t, ws)
	attachShockStunItemTable(t, s)

	s.processSkill(handler.SkillRequest{
		SessionID: caster.SessionID,
		SkillID:   87,
		TargetID:  target.CharID,
	})

	if caster.MP != 15 {
		t.Fatalf("Java checkUseSkill 會在 useConsume 前以 ranged=1 擋下射程外 SHOCK_STUN，不應消耗 MP，MP=%d", caster.MP)
	}
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("射程外 SHOCK_STUN 不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
}

func TestSkillClanShockStunPlayerTargetClearsSleepLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "sleep-target",
		X:          101,
		Y:          100,
		MapID:      4,
		Level:      1,
		RegistStun: -100,
		Sleeped:    true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 66, TicksLeft: 100, SetSleeped: true})
	target.AddBuff(&world.ActiveBuff{SkillID: 103, TicksLeft: 100, SetSleeped: true})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.Sleeped || target.HasBuff(66) || target.HasBuff(103) {
		t.Fatalf("Java L1SkillUse.runSkill() 會在 SHOCK_STUN 結果處理前移除睡眠效果，Sleeped=%v buff66=%v buff103=%v", target.Sleeped, target.HasBuff(66), target.HasBuff(103))
	}
	if !hasParalysisSubtype(drainSkillTestPackets(target.Session), handler.SleepRemove) {
		t.Fatal("SHOCK_STUN 清除玩家睡眠效果時應送 S_Paralysis SleepRemove")
	}
}

func TestSkillClanShockStunNpcTargetClearsSleepLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45000,
		Impl:          "L1Monster",
		Name:          "sleep-npc",
		X:             101,
		Y:             100,
		MapID:         4,
		Level:         1,
		Sleeped:       true,
		ActiveDebuffs: map[int32]int{66: 100, 103: 100},
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)

	if npc.Sleeped || npc.HasDebuff(66) || npc.HasDebuff(103) {
		t.Fatalf("Java L1SkillUse.runSkill() 會在 SHOCK_STUN NPC 目標結果處理前移除睡眠效果，Sleeped=%v debuff66=%v debuff103=%v", npc.Sleeped, npc.HasDebuff(66), npc.HasDebuff(103))
	}
}

func TestSkillClanShockStunPlayerTargetClearsEraseMagicLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "erase-target",
		X:          101,
		Y:          100,
		MapID:      4,
		Level:      1,
		RegistStun: -100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 153, TicksLeft: 100})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(153) {
		t.Fatal("Java L1SkillUse.runSkill() 會在 SHOCK_STUN 概率技能結果處理時移除目標 ERASE_MAGIC")
	}
}

func TestSkillClanShockStunNpcTargetClearsEraseMagicLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         45000,
		Impl:          "L1Monster",
		Name:          "erase-npc",
		X:             101,
		Y:             100,
		MapID:         4,
		Level:         1,
		ActiveDebuffs: map[int32]int{153: 100},
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)

	if npc.HasDebuff(153) {
		t.Fatal("Java L1SkillUse.runSkill() 會在 SHOCK_STUN NPC 目標概率技能結果處理時移除目標 ERASE_MAGIC")
	}
}

func TestSkillClanShockStunPlayerSuccessIgnoresMRWhenJavaImpactHaloMRZero(t *testing.T) {
	caster := &world.PlayerInfo{Level: 60, Intel: 18}
	target := &world.PlayerInfo{Level: 50, MR: 100}

	if got := shockStunPlayerProbability(caster, target); got != 40 {
		t.Fatalf("Java 預設 IMPACT_HALO_MR=0，SHOCK_STUN 不應扣 MR，got=%d want=40", got)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaStunLevel(t *testing.T) {
	caster := &world.PlayerInfo{Level: 49, StunLevel: 2}
	target := &world.PlayerInfo{Level: 50}

	if got := shockStunPlayerProbability(caster, target); got != 40 {
		t.Fatalf("Java SHOCK_STUN 會以攻方 Level + StunLevel 判定高低等成功率，got=%d want=40", got)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaOriginalMagicHit(t *testing.T) {
	caster := &world.PlayerInfo{Level: 50, OriginalMagicHit: 3}
	target := &world.PlayerInfo{Level: 50}
	npc := &world.NpcInfo{Level: 50}

	if got := shockStunPlayerProbability(caster, target); got != 33 {
		t.Fatalf("Java SHOCK_STUN 玩家目標成功率會加 getOriginalMagicHit，got=%d want=33", got)
	}
	if got := shockStunNpcProbability(caster, npc); got != 33 {
		t.Fatalf("Java SHOCK_STUN NPC 目標成功率會加 getOriginalMagicHit，got=%d want=33", got)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaIntMagicHitTable(t *testing.T) {
	caster := &world.PlayerInfo{Level: 50, Intel: 23}
	target := &world.PlayerInfo{Level: 50}
	npc := &world.NpcInfo{Level: 50}

	if got := shockStunPlayerProbability(caster, target); got != 31 {
		t.Fatalf("Java SHOCK_STUN 玩家目標成功率會加 L1AttackList.INTH 的 INT 魔法命中，got=%d want=31", got)
	}
	if got := shockStunNpcProbability(caster, npc); got != 31 {
		t.Fatalf("Java SHOCK_STUN NPC 目標成功率會加 L1AttackList.INTH 的 INT 魔法命中，got=%d want=31", got)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaBaseIntForPureIntBonus(t *testing.T) {
	caster := &world.PlayerInfo{
		Level: 50,
		Intel: 25,
		EquipBonuses: world.EquipStats{
			AddInt: 7,
		},
	}
	target := &world.PlayerInfo{Level: 50}
	npc := &world.NpcInfo{Level: 50}

	if got := shockStunPlayerProbability(caster, target); got != 31 {
		t.Fatalf("Java SHOCK_STUN 純 INT 加成使用 getBaseInt，不應把裝備 INT 算入純 INT 加成，got=%d want=31", got)
	}
	if got := shockStunNpcProbability(caster, npc); got != 31 {
		t.Fatalf("Java SHOCK_STUN NPC 目標純 INT 加成使用 getBaseInt，不應把裝備 INT 算入純 INT 加成，got=%d want=31", got)
	}
}

func TestSkillClanShockStunPlayerMissDoesNotSendTargetEffect(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
		Level:     50,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("測試前提錯誤：此 seed 應讓 SHOCK_STUN miss，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	casterPackets := drainSkillTestPackets(caster.Session)
	targetPackets := drainSkillTestPackets(target.Session)
	if hasServerMessage(casterPackets, skillMsgCastFail) {
		t.Fatal("Java SHOCK_STUN miss 時 _targetList 為空，不會送 S_ServerMessage(280)")
	}
	if hasActionGfxPacket(casterPackets, caster.CharID, 19) {
		t.Fatal("Java SHOCK_STUN miss 時 sendGrfx 會在 _targetList.size()==0 直接 return，不應送施法者 S_DoActionGFX")
	}
	if hasSkillEffectPacket(targetPackets, target.CharID, 4434) {
		t.Fatal("Java SHOCK_STUN miss 時 _targetList 為空，不應送出目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcSuccessUsesJavaImpactHaloRate(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, npc.ID)

	if npc.HasDebuff(87) || npc.Paralyzed {
		t.Fatalf("Java SHOCK_STUN 對 NPC 目標也以 IMPACT_HALO_1=40 為基準，不應使用 Go 一般 NPC 公式 95%%，debuff=%v Paralyzed=%v", npc.HasDebuff(87), npc.Paralyzed)
	}
	casterPackets := drainSkillTestPackets(caster.Session)
	if hasServerMessage(casterPackets, skillMsgCastFail) {
		t.Fatal("Java SHOCK_STUN NPC 目標 miss 時 _targetList 為空，不會送 S_ServerMessage(280)")
	}
	if hasActionGfxPacket(casterPackets, caster.CharID, 19) {
		t.Fatal("Java SHOCK_STUN NPC 目標 miss 時 sendGrfx 會在 _targetList.size()==0 直接 return，不應送施法者 S_DoActionGFX")
	}
}

func TestSkillClanShockStunNpcSuccessIgnoresMRWhenJavaImpactHaloMRZero(t *testing.T) {
	caster := &world.PlayerInfo{Level: 60, Intel: 18}
	npc := &world.NpcInfo{Level: 50, MR: 100}

	if got := shockStunNpcProbability(caster, npc); got != 40 {
		t.Fatalf("Java 預設 IMPACT_HALO_MR=0，SHOCK_STUN NPC 成功率不應扣 MR，got=%d want=40", got)
	}
}

func TestSkillClanShockStunPlayerCreatesJavaEffectNpc(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.applyShockStunToPlayer(caster.Session, caster, target, skill, ws.GetNearbyPlayersAt(target.X, target.Y, target.MapID))

	effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID)
	if len(effects) != 1 {
		t.Fatalf("Java SHOCK_STUN 成功時應 spawnEffect(81162) 到目標座標，got effects=%d", len(effects))
	}
	if effects[0].NpcID != 81162 || effects[0].GfxID != 4183 || effects[0].SkillID != 87 {
		t.Fatalf("SHOCK_STUN 地面效果資料錯誤: %+v", effects[0])
	}
	if effects[0].TicksLeft < 5 || effects[0].TicksLeft > 25 {
		t.Fatalf("SHOCK_STUN 81162 效果時間應等於 1~5 秒暈眩，ticks=%d", effects[0].TicksLeft)
	}
}

func TestSkillClanShockStunPlayerSuccessUsesJavaFixedTargetGfx(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	s.applyShockStunToPlayer(caster.Session, caster, target, skill, ws.GetNearbyPlayersAt(target.X, target.Y, target.MapID))

	casterPackets := drainSkillTestPackets(caster.Session)
	targetPackets := drainSkillTestPackets(target.Session)
	if hasActionGfxPacket(casterPackets, caster.CharID, 19) {
		t.Fatal("Java SHOCK_STUN 成功時 sendGrfx 特例只送目標 S_SkillSound(4434)，不送施法者 S_DoActionGFX")
	}
	if hasActionGfxPacket(targetPackets, caster.CharID, 19) {
		t.Fatal("Java SHOCK_STUN 成功時目標不應收到施法者 S_DoActionGFX")
	}
	if !hasSkillEffectPacket(targetPackets, target.CharID, 4434) {
		t.Fatal("Java SHOCK_STUN 成功時固定送目標 S_SkillSound(4434)，不依技能資料 CastGfx")
	}
}

func TestSkillClanShockStunPlayerTargetGfxBroadcastsFromTargetLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "target",
		X:          101,
		Y:          100,
		MapID:      4,
		RegistStun: -100,
	})
	observerNearTarget := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "observer",
		X:         121,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	packets := drainSkillTestPackets(observerNearTarget.Session)
	if !hasSkillEffectPacket(packets, target.CharID, 4434) {
		t.Fatal("Java SHOCK_STUN 玩家目標 4434 由目標 sendPacketsAll 廣播，看得到目標的玩家應收到")
	}
}

func TestSkillClanShockStunPlayerTargetEarthBindBlocksLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		Level:     80,
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  2,
		Session:    newSkillTestSession(t, 2),
		CharID:     1002,
		Name:       "target",
		Level:      1,
		X:          101,
		Y:          100,
		MapID:      4,
		RegistStun: -100,
		Paralyzed:  true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 157, TicksLeft: 100, SetParalyzed: true})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) {
		t.Fatal("Java 目標已有 EARTH_BIND 時 SHOCK_STUN 應在 calcProbabilityMagic 返回 false，不應套用 87")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(target.Session), target.CharID, 4434) {
		t.Fatal("Java 目標已有 EARTH_BIND 時 SHOCK_STUN targetList 為空，不應送目標 4434")
	}
}

func TestSkillClanShockStunPlayerTargetAbsoluteBarrierBlocksLikeJava(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 87)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:       2,
		Session:         newSkillTestSession(t, 2),
		CharID:          1002,
		Name:            "target",
		X:               101,
		Y:               100,
		MapID:           4,
		AbsoluteBarrier: true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 78, TicksLeft: 25, SetAbsoluteBarrier: true})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java L1SkillUse.checkTarget() 會排除 ABSOLUTE_BARRIER 目標，SHOCK_STUN 不應套用，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if hasSkillEffectPacket(drainSkillTestPackets(target.Session), target.CharID, 4434) {
		t.Fatal("Java SHOCK_STUN 對 ABSOLUTE_BARRIER 目標 targetList 為空，不應送目標 4434")
	}
}

func TestSkillClanShockStunNpcTargetEarthBindBlocksLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		Level:     80,
		X:         100,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 100,
		Impl:  "L1Monster",
		Name:  "frozen",
		Level: 1,
		X:     101,
		Y:     100,
		MapID: 4,
	}
	npc.Paralyzed = true
	npc.AddDebuff(157, 100)
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)

	if npc.HasDebuff(87) {
		t.Fatal("Java NPC 目標已有 EARTH_BIND 時 SHOCK_STUN 應在 calcProbabilityMagic 返回 false，不應套用 87")
	}
	if hasSkillEffectPacket(drainSkillTestPackets(caster.Session), npc.ID, 4434) {
		t.Fatal("Java NPC 目標已有 EARTH_BIND 時 SHOCK_STUN targetList 為空，不應送目標 4434")
	}
}

func TestSkillClanShockStunNpcCreatesJavaEffectNpc(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	for seed := int64(1); seed <= 100; seed++ {
		rand.Seed(seed)
		npc.Paralyzed = false
		npc.ActiveDebuffs = nil
		for _, effect := range ws.GroundEffectList() {
			ws.RemoveGroundEffect(effect.ID)
		}

		s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)
		if !npc.HasDebuff(87) {
			continue
		}

		effects := ws.GetNearbyGroundEffects(npc.X, npc.Y, npc.MapID)
		if len(effects) != 1 {
			t.Fatalf("Java SHOCK_STUN 對 NPC 成功時應 spawnEffect(81162) 到目標座標，got effects=%d seed=%d", len(effects), seed)
		}
		if effects[0].NpcID != 81162 || effects[0].GfxID != 4183 || effects[0].SkillID != 87 {
			t.Fatalf("SHOCK_STUN NPC 地面效果資料錯誤: %+v", effects[0])
		}
		if effects[0].TicksLeft < 5 || effects[0].TicksLeft > 25 {
			t.Fatalf("SHOCK_STUN NPC 81162 效果時間應等於 1~5 秒暈眩，ticks=%d", effects[0].TicksLeft)
		}
		return
	}
	t.Fatal("測試種子 1..100 未觸發 SHOCK_STUN NPC 成功，無法驗證 81162 效果")
}

func TestSkillClanShockStunNpcSuccessUsesJavaFixedTargetGfx(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     60,
	})
	npc := &world.NpcInfo{
		ID:    world.NextNpcID(),
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "stun target",
		X:     101,
		Y:     100,
		MapID: 4,
		Level: 50,
		HP:    100,
		MaxHP: 100,
	}
	ws.AddNpc(npc)
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	attachShockStunNpcTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 0}

	for seed := int64(1); seed <= 100; seed++ {
		rand.Seed(seed)
		npc.Paralyzed = false
		npc.ActiveDebuffs = nil
		_ = drainSkillTestPackets(caster.Session)

		s.executeNpcDebuffSkill(caster.Session, caster, skill, npc)
		if !npc.HasDebuff(87) {
			continue
		}
		if !hasSkillEffectPacket(drainSkillTestPackets(caster.Session), npc.ID, 4434) {
			t.Fatal("Java SHOCK_STUN 對 NPC 成功時固定送目標 S_SkillSound(4434)，不依技能資料 CastGfx")
		}
		return
	}
	t.Fatal("測試種子 1..100 未觸發 SHOCK_STUN NPC 成功，無法驗證固定 4434")
}

func TestSkillClanShockStunNpcCasterUsesJavaDurationAndEffectNpc(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	buff := target.GetBuff(87)
	if buff == nil || !target.Paralyzed {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 應套玩家暈眩，buff=%v Paralyzed=%v", buff, target.Paralyzed)
	}
	if buff.TicksLeft < 5 || buff.TicksLeft > 25 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 使用 1~5 秒，不應使用技能表 6 秒，ticks=%d", buff.TicksLeft)
	}
	effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID)
	if len(effects) != 1 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 應在目標座標 spawnEffect(81162)，got effects=%d", len(effects))
	}
	if effects[0].NpcID != 81162 || effects[0].SkillID != 87 || effects[0].OwnerCharID != npc.ID {
		t.Fatalf("NPC SHOCK_STUN 地面效果資料錯誤: %+v", effects[0])
	}
}

func TestSkillClanShockStunNpcCasterBroadcastsJavaActionGfx(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if !hasActionGfxPacket(drainSkillTestPackets(target.Session), npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 應廣播施法者 S_DoActionGFX")
	}
}

func TestSkillClanShockStunNpcCasterExistingBuffSendsJavaPacketsOnly(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		Level:     1,
		X:         103,
		Y:         100,
		MapID:     4,
		Paralyzed: true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 999, SetParalyzed: true})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if got := target.GetBuff(87).TicksLeft; got != 999 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 到已有 87 目標不應刷新時間，ticks=%d", got)
	}
	if effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID); len(effects) != 0 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 到已有 87 目標不應再次 spawnEffect(81162)，got=%d", len(effects))
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 到已有 87 目標仍應廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) == 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN 到已有 87 目標仍應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcCasterExistingBuffStillUsesJavaProbability(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		Paralyzed:  true,
		RegistStun: 200,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 999, SetParalyzed: true})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if got := target.GetBuff(87).TicksLeft; got != 999 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN 到已有 87 目標即使命中判定失敗也不應刷新時間，ticks=%d", got)
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 命中判定失敗時仍只廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN 命中判定失敗時 targetList 為空，不應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcCasterUsesJavaStunResistProbability(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "resist-target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: 127,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 99,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	s.ApplyNpcShockStun(npc, target, &data.SkillInfo{SkillID: 87, BuffDuration: 6}, 0)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java L1MagicNpc 對 SHOCK_STUN 會以 70/50/30 成功率再扣暈眩抗性，高暈抗目標不應被套用，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID); len(effects) != 0 {
		t.Fatalf("NPC SHOCK_STUN 抵抗時不應產生 81162 效果，got=%d", len(effects))
	}
}

func TestSkillClanShockStunNpcCasterProbabilityCapsAtJavaNinety(t *testing.T) {
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 99,
	}
	target := &world.PlayerInfo{
		CharID:     1001,
		Name:       "target",
		Level:      1,
		RegistStun: -100,
	}

	if got := shockStunNpcCasterProbability(npc, target, 10); got != 90 {
		t.Fatalf("Java L1MagicNpc.calcProbabilityMagic 會將 SHOCK_STUN 成功率封頂為 90，got=%d", got)
	}
}

func TestSkillClanShockStunNpcCasterTargetEarthBindBlocksLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "earth-bind-target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		Paralyzed:  true,
		RegistStun: -100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 157, TicksLeft: 100, SetParalyzed: true})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 99,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	s.ApplyNpcShockStun(npc, target, &data.SkillInfo{SkillID: 87, BuffDuration: 6, ActionID: 19, CastGfx: 4434}, 0)

	if target.HasBuff(87) {
		t.Fatal("Java L1MagicNpc 目標已有 EARTH_BIND 時 SHOCK_STUN 應在 calcProbabilityMagic 返回 false，不應套用 87")
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 判定失敗時仍應廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN 對 EARTH_BIND 目標 targetList 為空，不應送目標 4434")
	}
}

func TestSkillClanShockStunNpcCasterLeverageAffectsJavaProbability(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 1)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java NPC SHOCK_STUN 會套用 mob_skill leverage 到命中率，低 leverage 時此 seed 應 miss，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID); len(effects) != 0 {
		t.Fatalf("Java NPC SHOCK_STUN leverage miss 不應生成 81162，got=%d", len(effects))
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java NPC SHOCK_STUN leverage miss 仍應只廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("Java NPC SHOCK_STUN leverage miss 不應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcCasterMissClearsSleepLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "sleep-target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		Sleeped:    true,
		RegistStun: 200,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 66, TicksLeft: 100, SetSleeped: true})
	target.AddBuff(&world.ActiveBuff{SkillID: 103, TicksLeft: 100, SetSleeped: true})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能表失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if target.Sleeped || target.HasBuff(66) || target.HasBuff(103) {
		t.Fatalf("Java L1SkillUse.runSkill() 會在概率判定前移除睡眠效果，Sleeped=%v buff66=%v buff103=%v", target.Sleeped, target.HasBuff(66), target.HasBuff(103))
	}
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("SHOCK_STUN miss 不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasParalysisSubtype(packets, handler.SleepRemove) {
		t.Fatalf("解除睡眠時應送 S_Paralysis SleepRemove，packets=%v", packets)
	}
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("SHOCK_STUN miss 仍應廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("SHOCK_STUN miss 不應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcCasterMissClearsEraseMagicLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "erase-target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: 200,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 153, TicksLeft: 100})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	s.ApplyNpcShockStun(npc, target, &data.SkillInfo{SkillID: 87, BuffDuration: 6, ActionID: 19, CastGfx: 4434}, 0)

	if target.HasBuff(153) {
		t.Fatal("Java L1SkillUse.runSkill() 會在 SHOCK_STUN NPC 施放概率判定後、命中結果前移除目標 ERASE_MAGIC")
	}
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("SHOCK_STUN miss 不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
}

func TestSkillClanShockStunNpcCasterSkipsGmInvisibleTargetLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "gm-invis-target",
		Level:       1,
		X:           103,
		Y:           100,
		MapID:       4,
		AccessLevel: 200,
		Invisible:   true,
		Sleeped:     true,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 66, TicksLeft: 100, SetSleeped: true})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
		MP:    100,
		MaxMP: 100,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能表失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if !target.Sleeped || !target.HasBuff(66) {
		t.Fatalf("Java checkTarget 會先排除 GM 隱身目標，不應進入 runSkill 清除睡眠，Sleeped=%v buff66=%v", target.Sleeped, target.HasBuff(66))
	}
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("GM 隱身目標不應被 NPC SHOCK_STUN 套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	packets := drainSkillTestPackets(target.Session)
	if hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java checkUseSkill=false 時不會送 NPC SHOCK_STUN 施法動作")
	}
	if hasParalysisSubtype(packets, handler.SleepRemove) {
		t.Fatal("GM 隱身目標被排除時不應送睡眠解除")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("GM 隱身目標被排除時不應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunPlayerCasterSkipsGmInvisibleTargetLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		Level:     80,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   2,
		Session:     newSkillTestSession(t, 2),
		CharID:      1002,
		Name:        "gm-invis-target",
		X:           101,
		Y:           100,
		MapID:       4,
		Level:       1,
		AccessLevel: 200,
		Invisible:   true,
		Sleeped:     true,
		RegistStun:  -100,
	})
	target.AddBuff(&world.ActiveBuff{SkillID: 66, TicksLeft: 100, SetSleeped: true})
	target.AddBuff(&world.ActiveBuff{SkillID: 153, TicksLeft: 100})
	s := newSkillTestSystem(t, ws)
	combat := &shockStunCombatSpy{}
	s.deps.Combat = combat
	attachShockStunItemTable(t, s)
	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java L1SkillUse.checkTarget() 會排除 GM 隱身玩家，SHOCK_STUN 不應套用，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if len(combat.requests) != 0 {
		t.Fatalf("Java checkUseSkill=false 時不會進入 runSkill，不應觸發 SHOCK_STUN onAction，requests=%d", len(combat.requests))
	}
	if !target.Sleeped || !target.HasBuff(66) || !target.HasBuff(153) {
		t.Fatalf("Java checkUseSkill=false 時不會進入 runSkill，不應清除 GM 隱身目標既有狀態，Sleeped=%v buff66=%v buff153=%v", target.Sleeped, target.HasBuff(66), target.HasBuff(153))
	}
	if countSkillEffectPackets(drainSkillTestPackets(target.Session), target.CharID, 4434) != 0 {
		t.Fatal("Java checkUseSkill=false 時不應送 GM 隱身目標 4434")
	}
}

func TestSkillClanShockStunNpcCasterMissSendsOnlyJavaActionGfx(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "resist-target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: 127,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 99,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 0, 0)

	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("Java NPC 施放 SHOCK_STUN miss 時不應套用 87，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID); len(effects) != 0 {
		t.Fatalf("Java NPC 施放 SHOCK_STUN miss 時不應生成 81162，got=%d", len(effects))
	}
	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN miss 時仍應廣播施法者 S_DoActionGFX")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN miss 時 targetList 為空，不應送目標 S_SkillSound(4434)")
	}
}

func TestSkillClanShockStunNpcCasterUsesMobSkillActIDOverride(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能資料失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 7, 0, 0)

	packets := drainSkillTestPackets(target.Session)
	if !hasActionGfxPacket(packets, npc.ID, 7) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 應優先使用 mob_skill act_id 覆寫施法者動作")
	}
	if hasActionGfxPacket(packets, npc.ID, 19) {
		t.Fatal("mob_skill act_id 有值時不應回退使用技能表 action_id=19")
	}
}

func TestSkillClanShockStunNpcCasterUsesMobSkillGfxIDOverride(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          103,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能表失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 1234, 0)

	packets := drainSkillTestPackets(target.Session)
	if countSkillEffectPackets(packets, target.CharID, 1234) == 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN 應優先使用 mob_skill gfx_id 作為目標 S_SkillSound")
	}
	if countSkillEffectPackets(packets, target.CharID, 4434) != 0 {
		t.Fatal("mob_skill gfx_id 非 0 時不應回退技能表 cast_gfx=4434")
	}
}

func TestSkillClanShockStunNpcCasterBroadcastsFromCasterLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:  1,
		Session:    newSkillTestSession(t, 1),
		CharID:     1001,
		Name:       "target",
		Level:      1,
		X:          120,
		Y:          100,
		MapID:      4,
		RegistStun: -30,
	})
	observerNearCaster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "observer-near-caster",
		X:         80,
		Y:         100,
		MapID:     4,
	})
	observerNearTarget := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		CharID:    1003,
		Name:      "observer-near-target",
		X:         140,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 45000,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	skills, err := data.LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("讀取技能表失敗: %v", err)
	}
	s.deps.Skills = skills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcSkill(npc, target, 87, 19, 4434, 0)

	nearCasterPackets := drainSkillTestPackets(observerNearCaster.Session)
	if !hasActionGfxPacket(nearCasterPackets, npc.ID, 19) {
		t.Fatal("Java NPC 施放 SHOCK_STUN 應從施法 NPC 廣播 S_DoActionGFX")
	}
	if countSkillEffectPackets(nearCasterPackets, target.CharID, 4434) == 0 {
		t.Fatal("Java NPC 施放 SHOCK_STUN 的目標 S_SkillSound 也從施法 NPC 廣播")
	}

	nearTargetPackets := drainSkillTestPackets(observerNearTarget.Session)
	if hasActionGfxPacket(nearTargetPackets, npc.ID, 19) || countSkillEffectPackets(nearTargetPackets, target.CharID, 4434) != 0 {
		t.Fatal("不在施法 NPC 可見範圍內的玩家不應收到 NPC SHOCK_STUN 廣播封包")
	}
}

func TestSkillClanShockStunMobAreaShockStunMatchesJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	existing := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "existing",
		X:         102,
		Y:         100,
		MapID:     4,
		Paralyzed: true,
	})
	existing.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 999, SetParalyzed: true})
	gmInvis := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   3,
		Session:     newSkillTestSession(t, 3),
		CharID:      1003,
		Name:        "gm-invis",
		X:           103,
		Y:           100,
		MapID:       4,
		AccessLevel: 200,
		Invisible:   true,
	})
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 231008,
		Impl:  "L1Monster",
		Name:  "mob",
		Level: 50,
		X:     100,
		Y:     100,
		MapID: 4,
	}
	ws.AddNpc(npc)

	s := newSkillTestSystem(t, ws)
	attachShockStunNpcTable(t, s)
	s.deps.Skill = s
	ai := NewNpcAISystem(ws, s.deps)
	ai.executeNpcAreaShockStun(npc, 0)

	buff := target.GetBuff(87)
	if buff == nil || !target.Paralyzed {
		t.Fatalf("Java type 5 areashock_stun 應套玩家暈眩，buff=%v Paralyzed=%v", buff, target.Paralyzed)
	}
	if buff.TicksLeft < 10 || buff.TicksLeft > 25 {
		t.Fatalf("Java type 5 areashock_stun 使用 2~5 秒，ticks=%d", buff.TicksLeft)
	}
	if got := existing.GetBuff(87).TicksLeft; got != 999 {
		t.Fatalf("Java type 5 areashock_stun 不重套已有 SHOCK_STUN 的目標，ticks=%d", got)
	}
	if gmInvis.HasBuff(87) {
		t.Fatal("Java type 5 areashock_stun 會略過 GM 隱身目標")
	}
	effects := ws.GetNearbyGroundEffects(target.X, target.Y, target.MapID)
	if len(effects) != 1 {
		t.Fatalf("Java type 5 areashock_stun 應只在新命中的玩家座標 spawnEffect(81162)，got effects=%d", len(effects))
	}
	if effects[0].NpcID != 81162 || effects[0].SkillID != 87 || effects[0].OwnerCharID != npc.ID {
		t.Fatalf("type 5 areashock_stun 地面效果資料錯誤: %+v", effects[0])
	}
	if !hasActionGfxPacket(drainSkillTestPackets(target.Session), npc.ID, 1) {
		t.Fatal("Java type 5 areashock_stun act_id=0 時應廣播 S_DoActionGFX action=1")
	}
}

func TestSkillClanShockStunMobAreaShockStunMap93DoesNotCooldownLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     93,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       990001,
		Impl:        "L1Monster",
		Name:        "mob",
		Level:       50,
		X:           100,
		Y:           100,
		MapID:       93,
		HP:          100,
		MaxHP:       100,
		MP:          100,
		MaxMP:       100,
		AtkSpeed:    1000,
		Agro:        true,
		AggroTarget: target.SessionID,
	}
	ws.AddNpc(npc)
	mobSkillPath := filepath.Join(t.TempDir(), "mob_skill_list.yaml")
	if err := os.WriteFile(mobSkillPath, []byte(`mob_skills:
  - mob_id: 990001
    skills:
      - act_no: 1
        type: 5
        mp_consume: 0
        trigger_random: 100
        trigger_hp: 0
        trigger_range: 2
        skill_id: 0
        act_id: 0
        leverage: 0
        gfx_id: 0
        skill_area: 1
        change_target: 0
        summon_id: 0
        summon_min: 0
        summon_max: 0
`), 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("讀取 mob skill 測試資料失敗: %v", err)
	}
	s := newSkillTestSystem(t, ws)
	s.deps.MobSkills = mobSkills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.Update(0)

	if npc.AttackTimer != 0 {
		t.Fatalf("Java areashock_stun 在 map 93 會 return false，不應設定技能延遲，AttackTimer=%d", npc.AttackTimer)
	}
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("map 93 不應套用 type 5 SHOCK_STUN，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	if hasActionGfxPacket(drainSkillTestPackets(target.Session), npc.ID, 1) {
		t.Fatal("map 93 不應廣播 type 5 SHOCK_STUN 施法動作")
	}
}

func TestSkillClanShockStunMobAreaShockStunUsesSubMagicSpeedCooldownLikeJava(t *testing.T) {
	rand.Seed(1)
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	npc := &world.NpcInfo{
		ID:            2001,
		NpcID:         990002,
		Impl:          "L1Monster",
		Name:          "mob",
		Level:         50,
		X:             100,
		Y:             100,
		MapID:         4,
		HP:            100,
		MaxHP:         100,
		MP:            100,
		MaxMP:         100,
		AtkSpeed:      1000,
		SubMagicSpeed: 1600,
		Agro:          true,
		AggroTarget:   target.SessionID,
	}
	ws.AddNpc(npc)
	mobSkillPath := filepath.Join(t.TempDir(), "mob_skill_list.yaml")
	if err := os.WriteFile(mobSkillPath, []byte(`mob_skills:
  - mob_id: 990002
    skills:
      - act_no: 1
        type: 5
        mp_consume: 0
        trigger_random: 100
        trigger_hp: 0
        trigger_range: 2
        skill_id: 0
        act_id: 0
        leverage: 0
        gfx_id: 0
        skill_area: 1
        change_target: 0
        summon_id: 0
        summon_min: 0
        summon_max: 0
`), 0o644); err != nil {
		t.Fatalf("寫入 mob skill 測試資料失敗: %v", err)
	}
	mobSkills, err := data.LoadMobSkillTable(mobSkillPath)
	if err != nil {
		t.Fatalf("讀取 mob skill 測試資料失敗: %v", err)
	}
	s := newSkillTestSystem(t, ws)
	s.deps.MobSkills = mobSkills
	s.deps.Skill = s

	ai := NewNpcAISystem(ws, s.deps)
	ai.Update(0)

	if npc.AttackTimer != 8 {
		t.Fatalf("Java areashock_stun 成功後使用 sub_magic_speed 設定延遲，AttackTimer=%d want=8", npc.AttackTimer)
	}
}

func TestSkillClanShockStunCallClanCanTargetSameClanAcrossMaps(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		MP:        100,
		MaxMP:     100,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         33000,
		Y:         33000,
		MapID:     304,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 116, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if member.PendingYesNoType != 729 || member.PendingYesNoData != caster.CharID {
		t.Fatalf("呼喚盟友應可跨地圖對同血盟成員送 729 確認，Pending=(%d,%d)", member.PendingYesNoType, member.PendingYesNoData)
	}
}

func TestSkillClanShockStunRunClanTeleportsToAllowedClanMap(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     4,
		ClanID:    7,
		MP:        100,
		MaxMP:     100,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         32710,
		Y:         32810,
		MapID:     304,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 118, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if caster.X != member.X || caster.Y != member.Y || caster.MapID != member.MapID {
		t.Fatalf("援護盟友應傳送到 0/4/304 的同血盟目標位置，got=(%d,%d,%d) want=(%d,%d,%d)",
			caster.X, caster.Y, caster.MapID, member.X, member.Y, member.MapID)
	}
}

func TestSkillClanShockStunRunClanRejectsDisallowedTargetMap(t *testing.T) {
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "leader",
		X:         100,
		Y:         100,
		MapID:     100,
		ClanID:    7,
	})
	member := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "member",
		X:         101,
		Y:         100,
		MapID:     100,
		ClanID:    7,
	})
	s := newSkillTestSystem(t, ws)

	s.executeBuffSkill(caster.Session, caster, &data.SkillInfo{SkillID: 118, BuffDuration: 0, Target: "buff", ActionID: 19}, member.CharID)

	if caster.X != 100 || caster.Y != 100 || caster.MapID != 100 {
		t.Fatalf("援護盟友不應傳送到 Java 禁止的目標地圖，got=(%d,%d,%d)", caster.X, caster.Y, caster.MapID)
	}
}

func TestSkillClanShockStunShockStunRequiresTwoHandSwordForPlayerTarget(t *testing.T) {
	disablePlayerDebuffMRForStatusTest(t, 87)
	ws := world.NewState()
	caster := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     4,
	})
	s := newSkillTestSystem(t, ws)
	attachShockStunItemTable(t, s)
	skill := &data.SkillInfo{SkillID: 87, BuffDuration: 6, Target: "buff", ActionID: 19, CastGfx: 4434}

	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)
	if target.HasBuff(87) || target.Paralyzed {
		t.Fatalf("未裝備雙手劍時衝擊之暈不應套用，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
	packets := drainSkillTestPackets(caster.Session)
	if !hasGlobalSystemMessageText(packets, "請使用雙手劍") {
		t.Fatal("Java SHOCK_STUN 未裝備雙手劍時應送 S_SystemMessage(\"請使用雙手劍\")，不可夾帶色碼或額外標點")
	}

	caster.Equip.Set(world.SlotWeapon, &world.InvItem{ObjectID: 5001, ItemID: 16, Equipped: true})
	s.executeBuffSkill(caster.Session, caster, skill, target.CharID)
	if !target.HasBuff(87) || !target.Paralyzed {
		t.Fatalf("裝備雙手劍時衝擊之暈應套用暈眩，buff=%v Paralyzed=%v", target.GetBuff(87), target.Paralyzed)
	}
}

func TestSkillClanShockStunBounceAttackAddsHitOnlyOnce(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "knight",
		X:         100,
		Y:         100,
		MapID:     4,
		HitMod:    2,
	})
	s := newSkillTestSystem(t, ws)
	skill := &data.SkillInfo{SkillID: 89, BuffDuration: 64, Target: "none", ActionID: 19}

	s.executeSelfSkill(player.Session, player, skill)
	s.executeSelfSkill(player.Session, player, skill)

	if !player.HasBuff(89) || player.HitMod != 8 {
		t.Fatalf("尖刺盔甲應依 Java 給 Hit +6 且重放不疊加，buff=%v HitMod=%d", player.GetBuff(89), player.HitMod)
	}
}
