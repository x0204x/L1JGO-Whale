package system

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestAddHateLinksMobGroupMembersLikeJava(t *testing.T) {
	const attackerSID uint64 = 100
	const otherSID uint64 = 200

	leader := &world.NpcInfo{ID: 200_000_001, NpcID: 45001, ShowID: 7}
	minion := &world.NpcInfo{ID: 200_000_002, NpcID: 45002, ShowID: 7}
	busyMinion := &world.NpcInfo{
		ID:          200_000_003,
		NpcID:       45003,
		ShowID:      7,
		AggroTarget: otherSID,
		HateList:    map[uint64]int32{otherSID: 5},
	}
	group := &world.MobGroupInfo{
		Leader:  leader,
		Members: []*world.NpcInfo{leader, minion, busyMinion},
	}
	leader.GroupInfo = group
	minion.GroupInfo = group
	busyMinion.GroupInfo = group

	AddHate(leader, attackerSID, 30)

	if leader.HateList[attackerSID] != 30 || leader.AggroTarget != attackerSID {
		t.Fatalf("被攻擊的隊長應累加正常仇恨，hate=%v target=%d", leader.HateList, leader.AggroTarget)
	}
	if minion.HateList[attackerSID] != 0 || minion.AggroTarget != attackerSID {
		t.Fatalf("yiwei mob group setLink 應讓空仇恨隊員以 0 hate 鎖定攻擊者，hate=%v target=%d", minion.HateList, minion.AggroTarget)
	}
	if busyMinion.HateList[otherSID] != 5 || busyMinion.AggroTarget != otherSID {
		t.Fatalf("已有仇恨的隊員不應被 setLink 覆蓋，hate=%v target=%d", busyMinion.HateList, busyMinion.AggroTarget)
	}
}

func TestAddHateSkipsGyukiMobGroupLinkLikeJava(t *testing.T) {
	const attackerSID uint64 = 100

	leader := &world.NpcInfo{ID: 200_000_001, NpcID: 45001, ShowID: 0}
	gyuki := &world.NpcInfo{ID: 200_000_002, NpcID: 99007, ShowID: 0}
	group := &world.MobGroupInfo{
		Leader:  leader,
		Members: []*world.NpcInfo{leader, gyuki},
	}
	leader.GroupInfo = group
	gyuki.GroupInfo = group

	AddHate(leader, attackerSID, 30)

	if gyuki.AggroTarget != 0 || gyuki.HateList != nil {
		t.Fatalf("yiwei serchLink 會排除 99007 牛鬼，AggroTarget=%d HateList=%v", gyuki.AggroTarget, gyuki.HateList)
	}
}
