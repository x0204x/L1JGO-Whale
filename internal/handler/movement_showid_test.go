package handler

import (
	"encoding/binary"
	stdnet "net"
	"testing"

	"github.com/l1jgo/server/internal/config"
	l1net "github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func newMovementShowIDTestSession(t *testing.T, id uint64) *l1net.Session {
	t.Helper()
	client, server := stdnet.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
	})
	sess := l1net.NewSession(server, id, 32, 32, 0, zap.NewNop())
	t.Cleanup(sess.Close)
	return sess
}

func drainMovementShowIDPackets(sess *l1net.Session) [][]byte {
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

func hasMovementShowIDPutObjectPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 9 || pkt[0] != packet.S_OPCODE_PUT_OBJECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[5:9])) == objectID {
			return true
		}
	}
	return false
}

func hasMovementShowIDMovePacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_MOVE_OBJECT {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}

func hasMovementShowIDChangeHeadingPacket(packets [][]byte, objectID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 || pkt[0] != packet.S_OPCODE_CHANGEHEADING {
			continue
		}
		if int32(binary.LittleEndian.Uint32(pkt[1:5])) == objectID {
			return true
		}
	}
	return false
}

func addMovementShowIDPlayer(ws *world.State, p *world.PlayerInfo) *world.PlayerInfo {
	if p.Known == nil {
		p.Known = world.NewKnownEntities()
	}
	ws.AddPlayer(p)
	return p
}

func TestMoveAndDirectionBroadcastOnlySameShowLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5001,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 1,
		Session:   newMovementShowIDTestSession(t, 1),
	})
	sameShow := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5002,
		Name:      "same",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 2,
		Session:   newMovementShowIDTestSession(t, 2),
	})
	otherShow := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5003,
		Name:      "other",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 3,
		Session:   newMovementShowIDTestSession(t, 3),
	})

	deps := &Deps{World: ws, Config: &config.Config{}}
	HandleMove(player.Session, packet.NewReader([]byte{0, 0, 0, 0, 0, 2}), deps)

	if !hasMovementShowIDMovePacket(drainMovementShowIDPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到移動封包")
	}
	if hasMovementShowIDMovePacket(drainMovementShowIDPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到移動封包")
	}

	HandleChangeDirection(player.Session, packet.NewReader([]byte{0, 4}), deps)

	if !hasMovementShowIDChangeHeadingPacket(drainMovementShowIDPackets(sameShow.Session), player.CharID) {
		t.Fatalf("同 ShowID 玩家應收到轉向封包")
	}
	if hasMovementShowIDChangeHeadingPacket(drainMovementShowIDPackets(otherShow.Session), player.CharID) {
		t.Fatalf("不同 ShowID 玩家不應收到轉向封包")
	}
}

func TestRejectMoveRebuildsOnlySameShowKnownObjectsLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5101,
		Name:      "player",
		X:         100,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 11,
		Session:   newMovementShowIDTestSession(t, 11),
	})
	samePlayer := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5102,
		Name:      "same_player",
		X:         101,
		Y:         100,
		MapID:     900,
		ShowID:    77,
		SessionID: 12,
		Session:   newMovementShowIDTestSession(t, 12),
	})
	otherPlayer := addMovementShowIDPlayer(ws, &world.PlayerInfo{
		CharID:    5103,
		Name:      "other_player",
		X:         102,
		Y:         100,
		MapID:     900,
		ShowID:    88,
		SessionID: 13,
		Session:   newMovementShowIDTestSession(t, 13),
	})

	sameNpc := &world.NpcInfo{ID: 5201, NpcID: 45000, Name: "same_npc", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherNpc := &world.NpcInfo{ID: 5202, NpcID: 45000, Name: "other_npc", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddNpc(sameNpc)
	ws.AddNpc(otherNpc)

	sameSummon := &world.SummonInfo{ID: 5301, OwnerCharID: player.CharID, Name: "same_summon", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherSummon := &world.SummonInfo{ID: 5302, OwnerCharID: player.CharID, Name: "other_summon", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddSummon(sameSummon)
	ws.AddSummon(otherSummon)

	sameDoll := &world.DollInfo{ID: 5401, OwnerCharID: player.CharID, Name: "same_doll", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherDoll := &world.DollInfo{ID: 5402, OwnerCharID: player.CharID, Name: "other_doll", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddDoll(sameDoll)
	ws.AddDoll(otherDoll)

	sameHierarch := &world.HierarchInfo{ID: 5501, OwnerCharID: player.CharID, Name: "same_hierarch", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherHierarch := &world.HierarchInfo{ID: 5502, OwnerCharID: player.CharID, Name: "other_hierarch", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddHierarch(sameHierarch)
	ws.AddHierarch(otherHierarch)

	sameFollower := &world.FollowerInfo{ID: 5601, OwnerCharID: player.CharID, Name: "same_follower", NameID: "$1", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherFollower := &world.FollowerInfo{ID: 5602, OwnerCharID: player.CharID, Name: "other_follower", NameID: "$1", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddFollower(sameFollower)
	ws.AddFollower(otherFollower)

	samePet := &world.PetInfo{ID: 5701, OwnerCharID: player.CharID, Name: "same_pet", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherPet := &world.PetInfo{ID: 5702, OwnerCharID: player.CharID, Name: "other_pet", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddPet(samePet)
	ws.AddPet(otherPet)

	sameGround := &world.GroundItem{ID: 5801, ItemID: 1001, Count: 1, Name: "same_ground", X: 101, Y: 100, MapID: 900, ShowID: 77}
	otherGround := &world.GroundItem{ID: 5802, ItemID: 1001, Count: 1, Name: "other_ground", X: 102, Y: 100, MapID: 900, ShowID: 88}
	ws.AddGroundItem(sameGround)
	ws.AddGroundItem(otherGround)

	rejectMove(player.Session, player, ws, &Deps{})

	if _, ok := player.Known.Players[samePlayer.CharID]; !ok {
		t.Fatalf("同 ShowID 玩家應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Players[otherPlayer.CharID]; ok {
		t.Fatalf("不同 ShowID 玩家不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Npcs[sameNpc.ID]; !ok {
		t.Fatalf("同 ShowID NPC 應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Npcs[otherNpc.ID]; ok {
		t.Fatalf("不同 ShowID NPC 不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Summons[sameSummon.ID]; !ok {
		t.Fatalf("同 ShowID 召喚物應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Summons[otherSummon.ID]; ok {
		t.Fatalf("不同 ShowID 召喚物不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Dolls[sameDoll.ID]; !ok {
		t.Fatalf("同 ShowID 娃娃應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Dolls[otherDoll.ID]; ok {
		t.Fatalf("不同 ShowID 娃娃不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Hierarchs[sameHierarch.ID]; !ok {
		t.Fatalf("同 ShowID 祭司應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Hierarchs[otherHierarch.ID]; ok {
		t.Fatalf("不同 ShowID 祭司不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Followers[sameFollower.ID]; !ok {
		t.Fatalf("同 ShowID 跟隨物應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Followers[otherFollower.ID]; ok {
		t.Fatalf("不同 ShowID 跟隨物不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Pets[samePet.ID]; !ok {
		t.Fatalf("同 ShowID 寵物應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.Pets[otherPet.ID]; ok {
		t.Fatalf("不同 ShowID 寵物不應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.GroundItems[sameGround.ID]; !ok {
		t.Fatalf("同 ShowID 地上物應在 rejectMove 後寫入 Known")
	}
	if _, ok := player.Known.GroundItems[otherGround.ID]; ok {
		t.Fatalf("不同 ShowID 地上物不應在 rejectMove 後寫入 Known")
	}

	packets := drainMovementShowIDPackets(player.Session)
	for _, objectID := range []int32{otherPlayer.CharID, otherNpc.ID, otherSummon.ID, otherDoll.ID, otherHierarch.ID, otherFollower.ID, otherPet.ID, otherGround.ID} {
		if hasMovementShowIDPutObjectPacket(packets, objectID) {
			t.Fatalf("不同 ShowID 物件 %d 不應在 rejectMove 後重送顯示封包", objectID)
		}
	}
}
