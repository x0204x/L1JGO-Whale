package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestApplyNpcInitialHideSpartoiSinksLikeJava(t *testing.T) {
	npc := &world.NpcInfo{NpcID: 45161}

	applyNpcInitialHideRoll(npc, 0)

	if npc.HiddenStatus != world.NpcHiddenSink {
		t.Fatalf("yiwei initHide roll=0 時 45161 史巴托應遁地，HiddenStatus=%d", npc.HiddenStatus)
	}
}

func TestApplyNpcInitialHideStoneGolemSinksWithStatusFourLikeJava(t *testing.T) {
	npc := &world.NpcInfo{NpcID: 45126}

	applyNpcInitialHideRoll(npc, 0)

	if npc.HiddenStatus != world.NpcHiddenSink {
		t.Fatalf("yiwei initHide roll=0 時 45126 高崙石頭怪應遁地，HiddenStatus=%d", npc.HiddenStatus)
	}
	if status := world.NpcActionStatus(npc); status != 4 {
		t.Fatalf("高崙石頭怪遁地 S_NPCPack status 應為 4，got=%d", status)
	}
}

func TestApplyNpcInitialHideStoneGolemRollOneStaysVisibleLikeJava(t *testing.T) {
	npc := &world.NpcInfo{NpcID: 45126}

	applyNpcInitialHideRoll(npc, 1)

	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("yiwei initHide roll!=0 時 45126 高崙石頭怪不應遁地，HiddenStatus=%d", npc.HiddenStatus)
	}
	if status := world.NpcActionStatus(npc); status != 0 {
		t.Fatalf("未遁地高崙石頭怪 S_NPCPack status 應為 0，got=%d", status)
	}
}

func TestApplyNpcInitialHideIceGolemFamilyUsesIceStatusLikeJava(t *testing.T) {
	npc := &world.NpcInfo{NpcID: 46125}

	applyNpcInitialHideRoll(npc, 2)

	if npc.HiddenStatus != world.NpcHiddenIce {
		t.Fatalf("yiwei initHide 46125 應為 ICE hidden，HiddenStatus=%d", npc.HiddenStatus)
	}
	if status := world.NpcActionStatus(npc); status != 4 {
		t.Fatalf("ICE hidden S_NPCPack status 應為 4，got=%d", status)
	}
}

func TestApplyNpcInitialHideDoesNotSinkDeathRuinsSpartoiLikeJava(t *testing.T) {
	npc := &world.NpcInfo{NpcID: 56119}

	applyNpcInitialHideRoll(npc, 0)

	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("yiwei hide/initHide switch 未包含 56119，Go 不應自行讓它遁地，HiddenStatus=%d", npc.HiddenStatus)
	}
}

func TestHiddenSpartoiApproachPlayerAppearsAndAggrosLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         102,
		Y:         100,
		MapID:     900,
		Known:     world.NewKnownEntities(),
	})
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45161,
		Impl:         "L1Monster",
		Name:         "spartoi",
		X:            100,
		Y:            100,
		MapID:        900,
		HP:           100,
		MaxHP:        100,
		HiddenStatus: world.NpcHiddenSink,
	}
	ws.AddNpc(npc)
	player.Known.Npcs[npc.ID] = world.KnownPos{X: npc.X, Y: npc.Y}
	s := NewVisibilitySystem(ws, &handler.Deps{World: ws})

	s.updateNpcVisibility(player)
	packets := drainSkillTestPackets(player.Session)

	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("玩家 2 格內接近滿血遁地 NPC 時應出土，HiddenStatus=%d", npc.HiddenStatus)
	}
	if npc.AggroTarget != player.SessionID {
		t.Fatalf("出土後應鎖定接近玩家，AggroTarget=%d", npc.AggroTarget)
	}
	if len(packets) == 0 {
		t.Fatal("出土時應廣播動作/NPC pack")
	}
}

func TestHiddenIceNpcAppearsWhenHpBelowMaxLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         110,
		Y:         100,
		MapID:     900,
		Known:     world.NewKnownEntities(),
	})
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        46125,
		Impl:         "L1Monster",
		Name:         "ice_golem",
		X:            100,
		Y:            100,
		MapID:        900,
		HP:           99,
		MaxHP:        100,
		HiddenStatus: world.NpcHiddenIce,
	}
	ws.AddNpc(npc)
	s := NewVisibilitySystem(ws, &handler.Deps{World: ws})

	if !s.approachHiddenNpc(player, npc) {
		t.Fatal("ICE hidden NPC 在 HP < MaxHP 時應依 Java appearOnGround 顯形")
	}
	packets := drainSkillTestPackets(player.Session)

	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("ICE hidden 顯形後 HiddenStatus 應清除，got=%d", npc.HiddenStatus)
	}
	if npc.AggroTarget != player.SessionID {
		t.Fatalf("ICE hidden 顯形後應鎖定觸發玩家，AggroTarget=%d", npc.AggroTarget)
	}
	if len(packets) == 0 {
		t.Fatal("ICE hidden 顯形應廣播 action/NPC pack")
	}
}

func TestSinkHiddenNpcAppearsWhenRevealSkillHitsLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     900,
		Level:     60,
		Intel:     25,
	})
	npc := &world.NpcInfo{
		ID:                 2001,
		NpcID:              45126,
		Impl:               "L1Monster",
		Name:               "stone_golem",
		X:                  101,
		Y:                  100,
		MapID:              900,
		HP:                 100,
		MaxHP:              100,
		HiddenStatus:       world.NpcHiddenSink,
		HiddenActionStatus: 4,
	}
	ws.AddNpc(npc)
	combat := newCombatLOSTestSystem(t, ws, &fakePvPManager{})
	s := &SkillSystem{deps: combat.deps}

	s.executeAttackSkill(player.Session, player, &data.SkillInfo{
		SkillID:         194,
		Ranged:          5,
		DamageValue:     1,
		DamageDice:      1,
		DamageDiceCount: 1,
		ActionID:        19,
	}, npc.ID)
	packets := drainSkillTestPackets(player.Session)

	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("Java 194/213 類揭露技能命中 SINK hidden 時應 appearOnGround，HiddenStatus=%d", npc.HiddenStatus)
	}
	if npc.AggroTarget != player.SessionID {
		t.Fatalf("揭露 hidden NPC 後應鎖定施法者，AggroTarget=%d", npc.AggroTarget)
	}
	if len(packets) == 0 {
		t.Fatal("揭露 hidden NPC 應廣播 action/NPC pack")
	}
}

func TestHiddenNpcDoesNotAttackBeforeApproachLikeJava(t *testing.T) {
	ws := world.NewState()
	target := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "target",
		X:         101,
		Y:         100,
		MapID:     900,
		HP:        5000,
		MaxHP:     5000,
	})
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45161,
		Impl:         "L1Monster",
		Name:         "spartoi",
		X:            100,
		Y:            100,
		MapID:        900,
		HP:           100,
		MaxHP:        100,
		Level:        50,
		STR:          30,
		DEX:          30,
		AtkDmg:       20,
		AggroTarget:  target.SessionID,
		HiddenStatus: world.NpcHiddenSink,
	}
	ws.AddNpc(npc)
	s := newNpcAILOSTestSystem(t, ws)

	s.tickMonsterAI(npc)
	packets := drainSkillTestPackets(target.Session)

	if len(packets) != 0 {
		t.Fatalf("hidden status != NONE 時 Java NpcAI 不進 onTarget，不應送攻擊封包，packets=%d", len(packets))
	}
	if target.HP != target.MaxHP {
		t.Fatalf("遁地中不應造成傷害，HP=%d MaxHP=%d", target.HP, target.MaxHP)
	}
}

func TestHiddenSinkNpcRejectsMeleeDamageLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "attacker",
		X:         100,
		Y:         100,
		MapID:     900,
		Level:     80,
		Str:       35,
		Dex:       35,
		HitMod:    200,
		DmgMod:    10,
	})
	npc := &world.NpcInfo{
		ID:           2001,
		NpcID:        45126,
		Impl:         "L1Monster",
		Name:         "stone_golem",
		X:            101,
		Y:            100,
		MapID:        900,
		HP:           1000000,
		MaxHP:        1000000,
		HiddenStatus: world.NpcHiddenSink,
	}
	ws.AddNpc(npc)
	s := newCombatLOSTestSystem(t, ws, &fakePvPManager{})

	for i := 0; i < 40; i++ {
		s.processMeleeAttack(player.SessionID, npc.ID)
	}

	if npc.HP != npc.MaxHP {
		t.Fatalf("Java C_Attack/receiveDamage 會拒絕 SINK hidden 目標，Go 不應造成傷害，HP=%d", npc.HP)
	}
}

func TestTryNpcHideOnDamageSpartoiSinksBelowOneThirdLikeJava(t *testing.T) {
	ws := world.NewState()
	viewer := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "viewer",
		X:         100,
		Y:         100,
		MapID:     900,
	})
	npc := &world.NpcInfo{
		ID:          2001,
		NpcID:       45161,
		Impl:        "L1Monster",
		Name:        "spartoi",
		X:           100,
		Y:           100,
		MapID:       900,
		HP:          20,
		MaxHP:       90,
		AggroTarget: viewer.SessionID,
		HateList:    map[uint64]int32{viewer.SessionID: 30},
	}
	ws.AddNpc(npc)

	if !tryNpcHideOnDamageRoll(npc, ws, 0) {
		t.Fatal("HP < 1/3 且 roll=0 時史巴托應遁地")
	}
	packets := drainSkillTestPackets(viewer.Session)

	if npc.HiddenStatus != world.NpcHiddenSink {
		t.Fatalf("受傷遁地後 HiddenStatus 應為 SINK，got=%d", npc.HiddenStatus)
	}
	if npc.AggroTarget != 0 || npc.HateList != nil {
		t.Fatalf("受傷遁地應清除仇恨，AggroTarget=%d HateList=%v", npc.AggroTarget, npc.HateList)
	}
	if len(packets) == 0 {
		t.Fatal("受傷遁地應廣播動作/NPC pack")
	}
}

func TestNpcHiddenStateBroadcastsOnlySameShowLikeJava(t *testing.T) {
	t.Run("hide_on_damage", func(t *testing.T) {
		ws := world.NewState()
		sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
			SessionID: 1,
			Session:   newSkillTestSession(t, 1),
			CharID:    1001,
			Name:      "same_show",
			X:         100,
			Y:         100,
			MapID:     900,
			ShowID:    77,
		})
		otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
			SessionID: 2,
			Session:   newSkillTestSession(t, 2),
			CharID:    1002,
			Name:      "other_show",
			X:         100,
			Y:         100,
			MapID:     900,
			ShowID:    88,
		})
		npc := &world.NpcInfo{
			ID:          2001,
			NpcID:       45161,
			Impl:        "L1Monster",
			Name:        "spartoi",
			X:           100,
			Y:           100,
			MapID:       900,
			ShowID:      77,
			HP:          20,
			MaxHP:       90,
			AggroTarget: sameShow.SessionID,
			HateList:    map[uint64]int32{sameShow.SessionID: 30},
		}
		ws.AddNpc(npc)

		if !tryNpcHideOnDamageRoll(npc, ws, 0) {
			t.Fatal("HP < 1/3 且 roll=0 時史巴托應遁地")
		}

		samePackets := drainSkillTestPackets(sameShow.Session)
		if !hasActionGfxPacket(samePackets, npc.ID, 11) || !hasPutObjectPacket(samePackets, npc.ID) {
			t.Fatal("同 ShowID 玩家應收到 NPC 受傷遁地動作與 NPC pack")
		}
		otherPackets := drainSkillTestPackets(otherShow.Session)
		if hasActionGfxPacket(otherPackets, npc.ID, 11) || hasPutObjectPacket(otherPackets, npc.ID) {
			t.Fatal("不同 ShowID 玩家不應收到 NPC 受傷遁地動作或 NPC pack")
		}
	})

	t.Run("appear_on_ground", func(t *testing.T) {
		ws := world.NewState()
		trigger := addSkillTestPlayer(ws, &world.PlayerInfo{
			SessionID: 11,
			Session:   newSkillTestSession(t, 11),
			CharID:    1101,
			Name:      "trigger",
			X:         102,
			Y:         100,
			MapID:     900,
			ShowID:    77,
		})
		otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
			SessionID: 12,
			Session:   newSkillTestSession(t, 12),
			CharID:    1102,
			Name:      "other_show",
			X:         102,
			Y:         100,
			MapID:     900,
			ShowID:    88,
		})
		npc := &world.NpcInfo{
			ID:           2101,
			NpcID:        45161,
			Impl:         "L1Monster",
			Name:         "spartoi",
			X:            100,
			Y:            100,
			MapID:        900,
			ShowID:       77,
			HP:           100,
			MaxHP:        100,
			HiddenStatus: world.NpcHiddenSink,
		}
		ws.AddNpc(npc)

		if !npcAppearOnGroundLikeJava(npc, ws, trigger) {
			t.Fatal("接近遁地 NPC 應讓 NPC 出土")
		}

		samePackets := drainSkillTestPackets(trigger.Session)
		if !hasActionGfxPacket(samePackets, npc.ID, 4) || !hasPutObjectPacket(samePackets, npc.ID) {
			t.Fatal("同 ShowID 玩家應收到 NPC 出土動作與 NPC pack")
		}
		otherPackets := drainSkillTestPackets(otherShow.Session)
		if hasActionGfxPacket(otherPackets, npc.ID, 4) || hasPutObjectPacket(otherPackets, npc.ID) {
			t.Fatal("不同 ShowID 玩家不應收到 NPC 出土動作或 NPC pack")
		}
	})
}

func TestTryNpcHideOnDamageDeathRuinsSpartoiDoesNotSinkLikeJava(t *testing.T) {
	ws := world.NewState()
	npc := &world.NpcInfo{
		ID:    2001,
		NpcID: 56119,
		HP:    20,
		MaxHP: 90,
	}

	if tryNpcHideOnDamageRoll(npc, ws, 0) {
		t.Fatal("yiwei hide switch 未包含 56119，Go 不應自行讓死亡廢墟史巴托受傷遁地")
	}
	if npc.HiddenStatus != world.NpcHiddenNone {
		t.Fatalf("56119 不應受傷遁地，HiddenStatus=%d", npc.HiddenStatus)
	}
}
