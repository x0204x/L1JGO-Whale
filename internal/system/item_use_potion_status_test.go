package system

import (
	"math/rand"
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestItemUseHealPotionPolluteWaterWaveHalvesHealingLikeJava(t *testing.T) {
	ws := world.NewState()
	normal := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "normal",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     200,
	})
	polluted := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 2,
		Session:   newSkillTestSession(t, 2),
		CharID:    1002,
		Name:      "polluted",
		X:         101,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     200,
	})
	polluted.AddBuff(&world.ActiveBuff{SkillID: mobSkillPolluteWater, TicksLeft: 60})
	normalPotion := normal.Inv.AddItemWithID(7001, 40010, 1, "治癒藥水", 189, 1000, true, 1)
	pollutedPotion := polluted.Inv.AddItemWithID(7002, 40010, 1, "治癒藥水", 189, 1000, true, 1)
	sys := newShockStunItemUseSystem(t, ws)

	rand.Seed(7)
	if !sys.UseConsumable(normal.Session, normal, normalPotion, nil) {
		t.Fatal("普通治癒藥水應可使用")
	}
	normalGain := normal.HP - 100
	if normalGain <= 0 {
		t.Fatalf("普通治癒藥水應恢復 HP，gain=%d", normalGain)
	}

	rand.Seed(7)
	if !sys.UseConsumable(polluted.Session, polluted, pollutedPotion, nil) {
		t.Fatal("4012 狀態下治癒藥水仍應消耗並套用修正後效果")
	}
	pollutedGain := polluted.HP - 100
	if pollutedGain != normalGain>>1 {
		t.Fatalf("yiwei 4012 污濁的水流應讓藥水回復量減半，gain=%d want=%d", pollutedGain, normalGain>>1)
	}
}

func TestItemUseHealPotionPotionTurnToDamageLikeJava(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "potion-turn",
		X:         100,
		Y:         100,
		MapID:     4,
		HP:        100,
		MaxHP:     200,
	})
	player.AddBuff(&world.ActiveBuff{SkillID: mobSkillPotionTurnToDamage, TicksLeft: 60})
	potion := player.Inv.AddItemWithID(7001, 40010, 1, "治癒藥水", 189, 1000, true, 1)
	sys := newShockStunItemUseSystem(t, ws)

	rand.Seed(11)
	if !sys.UseConsumable(player.Session, player, potion, nil) {
		t.Fatal("4011 狀態下治癒藥水仍應消耗")
	}

	if player.HP >= 100 {
		t.Fatalf("yiwei 4011 藥水侵蝕術應把藥水治療轉成傷害，HP=%d", player.HP)
	}
}
