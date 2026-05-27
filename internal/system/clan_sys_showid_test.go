package system

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/config"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func TestClanSetTitleBroadcastsOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9001,
		Name:      "TitleOwner",
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		X:         33000,
		Y:         33000,
		MapID:     4,
		ShowID:    10,
		Level:     50,
	})
	sameShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9002,
		Name:      "SameShow",
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		X:         33001,
		Y:         33000,
		MapID:     4,
		ShowID:    10,
	})
	otherShow := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9003,
		Name:      "OtherShow",
		SessionID: 3,
		Session:   newSkillTestSession(t, 3),
		X:         33001,
		Y:         33000,
		MapID:     4,
		ShowID:    99,
	})

	sys := NewClanSystem(&handler.Deps{
		World:  ws,
		Config: &config.Config{},
		Log:    zap.NewNop(),
	})

	sys.SetTitle(player.Session, player, player.Name, "Captain")

	if !hasCharTitlePacket(drainSkillTestPackets(player.Session), player.CharID) {
		t.Fatalf("施放者自己應收到 S_CharTitle")
	}
	if !hasCharTitlePacket(drainSkillTestPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 觀眾應收到 S_CharTitle")
	}
	if hasCharTitlePacket(drainSkillTestPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 觀眾不應收到 S_CharTitle")
	}
}

func TestClanJoinRequestSkipsDifferentShowTargetLikeJava(t *testing.T) {
	ws := world.NewState()
	applicant := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9011,
		Name:      "Applicant",
		SessionID: 11,
		Session:   newSkillTestSession(t, 11),
		X:         33100,
		Y:         33100,
		MapID:     4,
		ShowID:    10,
	})
	otherShowLeader := addSkillTestPlayer(ws, &world.PlayerInfo{
		CharID:    9012,
		Name:      "OtherLeader",
		SessionID: 12,
		Session:   newSkillTestSession(t, 12),
		X:         33101,
		Y:         33100,
		MapID:     4,
		ShowID:    99,
		ClanID:    77,
		ClanName:  "OtherClan",
		ClanRank:  world.ClanRankPrince,
	})
	ws.Clans.AddClan(&world.ClanInfo{
		ClanID:     77,
		ClanName:   "OtherClan",
		LeaderID:   otherShowLeader.CharID,
		LeaderName: otherShowLeader.Name,
		Members: map[int32]*world.ClanMember{
			otherShowLeader.CharID: {
				CharID:   otherShowLeader.CharID,
				CharName: otherShowLeader.Name,
				Rank:     world.ClanRankPrince,
			},
		},
	})

	sys := NewClanSystem(&handler.Deps{
		World:  ws,
		Config: &config.Config{},
		Log:    zap.NewNop(),
	})

	sys.JoinRequest(applicant.Session, applicant)

	if otherShowLeader.PendingYesNoType == 97 || otherShowLeader.PendingYesNoData == applicant.CharID {
		t.Fatalf("不同 ShowID 盟主不應收到加入血盟 Y/N，Pending=(%d,%d)",
			otherShowLeader.PendingYesNoType, otherShowLeader.PendingYesNoData)
	}
}

func hasCharTitlePacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_CHARTITLE {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}
