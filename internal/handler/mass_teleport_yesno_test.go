package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestHandleYesNoResponseMassTeleportAcceptTeleportsToPendingDestination(t *testing.T) {
	ws := world.NewState()
	sess := newHandlerTestSession(t, 1)
	player := &world.PlayerInfo{
		SessionID:        1,
		Session:          sess,
		CharID:           1002,
		Name:             "ask-member",
		X:                101,
		Y:                100,
		MapID:            4,
		Heading:          5,
		ClanID:           10,
		PendingYesNoType: 748,
		TeleportX:        32710,
		TeleportY:        32820,
		TeleportMapID:    4,
		TeleportHeading:  5,
		Inv:              world.NewInventory(),
	}
	ws.AddPlayer(player)

	handleYesNoResponse(sess, player, true, &Deps{World: ws})

	if player.X != 32710 || player.Y != 32820 || player.MapID != 4 {
		t.Fatalf("接受集體傳送 748 後應移動到暫存座標，位置=%d,%d,%d", player.X, player.Y, player.MapID)
	}
	if player.PendingYesNoType != 0 {
		t.Fatalf("處理 748 回應後應清除 pending type，got=%d", player.PendingYesNoType)
	}
}

func TestHandleYesNoResponseCallClanBlockedByShockStunLikeJava(t *testing.T) {
	ws := world.NewState()
	callerSess := newHandlerTestSession(t, 1)
	caller := &world.PlayerInfo{
		SessionID: callerSess.ID,
		Session:   callerSess,
		CharID:    1001,
		Name:      "leader",
		X:         32710,
		Y:         32820,
		MapID:     4,
		ClanID:    10,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(caller)

	memberSess := newHandlerTestSession(t, 2)
	member := &world.PlayerInfo{
		SessionID:        memberSess.ID,
		Session:          memberSess,
		CharID:           1002,
		Name:             "stunned-member",
		X:                101,
		Y:                100,
		MapID:            4,
		ClanID:           10,
		Paralyzed:        true,
		PendingYesNoType: 729,
		PendingYesNoData: caller.CharID,
		Inv:              world.NewInventory(),
	}
	member.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	ws.AddPlayer(member)

	handleYesNoResponse(memberSess, member, true, &Deps{World: ws})

	if member.X != 101 || member.Y != 100 || member.MapID != 4 {
		t.Fatalf("Java callClan 會以 isParalyzedX 擋下 SHOCK_STUN 中接受呼喚盟友，不應傳送，位置=%d,%d,%d", member.X, member.Y, member.MapID)
	}
	if member.PendingYesNoType != 0 || member.PendingYesNoData != 0 {
		t.Fatalf("處理 729 回應後應清除 pending，type=%d data=%d", member.PendingYesNoType, member.PendingYesNoData)
	}
}

func TestHandleAttrCallClanResponseUsesJavaCAttrCase729(t *testing.T) {
	ws := world.NewState()
	callerSess := newHandlerTestSession(t, 1)
	caller := &world.PlayerInfo{
		SessionID: callerSess.ID,
		Session:   callerSess,
		CharID:    1001,
		Name:      "leader",
		X:         32710,
		Y:         32820,
		MapID:     4,
		ClanID:    10,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(caller)

	memberSess := newHandlerTestSession(t, 2)
	member := &world.PlayerInfo{
		SessionID:        memberSess.ID,
		Session:          memberSess,
		CharID:           1002,
		Name:             "member",
		X:                101,
		Y:                100,
		MapID:            4,
		ClanID:           10,
		PendingYesNoType: 729,
		PendingYesNoData: caller.CharID,
		Inv:              world.NewInventory(),
	}
	ws.AddPlayer(member)

	w := packet.NewWriterWithOpcode(packet.C_OPCODE_ATTR)
	w.WriteH(729)
	w.WriteC(1)

	HandleAttr(memberSess, packet.NewReader(w.RawBytes()), &Deps{World: ws, Log: zap.NewNop()})

	if member.X != caller.X || member.Y != caller.Y || member.MapID != caller.MapID {
		t.Fatalf("Java C_Attr case 729 接受呼喚盟友後應傳送到呼喚者位置，位置=%d,%d,%d", member.X, member.Y, member.MapID)
	}
	if member.PendingYesNoType != 0 || member.PendingYesNoData != 0 {
		t.Fatalf("C_Attr 729 回應後應清除 pending，type=%d data=%d", member.PendingYesNoType, member.PendingYesNoData)
	}
}

func TestHandleAttrAllianceCallClanResponseUsesJavaCAttrCase4976(t *testing.T) {
	ws := world.NewState()
	callerSess := newHandlerTestSession(t, 1)
	caller := &world.PlayerInfo{
		SessionID: callerSess.ID,
		Session:   callerSess,
		CharID:    1001,
		Name:      "alliance-leader",
		X:         32710,
		Y:         32820,
		MapID:     4,
		ClanID:    10,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(caller)

	alliances := NewAllianceManager()
	alliances.AddAlliance(&AllianceInfo{
		OrderID: 1,
		ClanIDs: [4]int32{
			10,
			20,
		},
	})

	memberSess := newHandlerTestSession(t, 2)
	member := &world.PlayerInfo{
		SessionID:        memberSess.ID,
		Session:          memberSess,
		CharID:           1002,
		Name:             "alliance-member",
		X:                101,
		Y:                100,
		MapID:            4,
		ClanID:           20,
		PendingYesNoType: 4976,
		PendingYesNoData: caller.CharID,
		Inv:              world.NewInventory(),
	}
	ws.AddPlayer(member)

	w := packet.NewWriterWithOpcode(packet.C_OPCODE_ATTR)
	w.WriteH(4976)
	w.WriteC(1)

	HandleAttr(memberSess, packet.NewReader(w.RawBytes()), &Deps{World: ws, Alliances: alliances, Log: zap.NewNop()})

	if member.X != caller.X || member.Y != caller.Y || member.MapID != caller.MapID {
		t.Fatalf("Java C_Attr case 4976 接受聯盟呼喚後應傳送到呼喚者位置，位置=%d,%d,%d", member.X, member.Y, member.MapID)
	}
}

func TestHandleAttrAllianceCallClanBlockedByShockStunLikeJava(t *testing.T) {
	ws := world.NewState()
	callerSess := newHandlerTestSession(t, 1)
	caller := &world.PlayerInfo{
		SessionID: callerSess.ID,
		Session:   callerSess,
		CharID:    1001,
		Name:      "alliance-leader",
		X:         32710,
		Y:         32820,
		MapID:     4,
		ClanID:    10,
		Inv:       world.NewInventory(),
	}
	ws.AddPlayer(caller)

	alliances := NewAllianceManager()
	alliances.AddAlliance(&AllianceInfo{
		OrderID: 1,
		ClanIDs: [4]int32{
			10,
			20,
		},
	})

	memberSess := newHandlerTestSession(t, 2)
	member := &world.PlayerInfo{
		SessionID:        memberSess.ID,
		Session:          memberSess,
		CharID:           1002,
		Name:             "stunned-alliance-member",
		X:                101,
		Y:                100,
		MapID:            4,
		ClanID:           20,
		Paralyzed:        true,
		PendingYesNoType: 4976,
		PendingYesNoData: caller.CharID,
		Inv:              world.NewInventory(),
	}
	member.AddBuff(&world.ActiveBuff{SkillID: 87, TicksLeft: 25, SetParalyzed: true})
	ws.AddPlayer(member)

	w := packet.NewWriterWithOpcode(packet.C_OPCODE_ATTR)
	w.WriteH(4976)
	w.WriteC(1)

	HandleAttr(memberSess, packet.NewReader(w.RawBytes()), &Deps{World: ws, Alliances: alliances, Log: zap.NewNop()})

	if member.X != 101 || member.Y != 100 || member.MapID != 4 {
		t.Fatalf("Java callClan1 會以 isParalyzedX 擋下 SHOCK_STUN 中接受聯盟呼喚，不應傳送，位置=%d,%d,%d", member.X, member.Y, member.MapID)
	}
}
