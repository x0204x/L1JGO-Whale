package world

import "testing"

func TestGroundEffectLifecycleAndVisibility(t *testing.T) {
	ws := NewState()
	effect := &GroundEffect{
		ID:        NextGroundEffectID(),
		SkillID:   63,
		NpcID:     81169,
		GfxID:     2231,
		Type:      GroundEffectLifeStream,
		X:         100,
		Y:         100,
		MapID:     4,
		TicksLeft: 2,
	}

	ws.AddGroundEffect(effect)

	nearby := ws.GetNearbyGroundEffects(102, 100, 4)
	if len(nearby) != 1 || nearby[0].ID != effect.ID {
		t.Fatalf("生命之泉地面效果應可被附近玩家看見，got=%d", len(nearby))
	}

	if expired := ws.TickGroundEffects(); len(expired) != 0 {
		t.Fatalf("第一個 tick 不應過期，expired=%d", len(expired))
	}
	expired := ws.TickGroundEffects()
	if len(expired) != 1 || expired[0].ID != effect.ID {
		t.Fatalf("第二個 tick 應回傳過期效果，expired=%d", len(expired))
	}
	if ws.GetGroundEffect(effect.ID) != nil {
		t.Fatal("過期地面效果應自世界狀態移除")
	}
}

func TestGroundEffectSamePointLookup(t *testing.T) {
	ws := NewState()
	effect := &GroundEffect{
		ID:      NextGroundEffectID(),
		SkillID: 58,
		NpcID:   81157,
		GfxID:   168,
		Type:    GroundEffectFireWall,
		X:       100,
		Y:       100,
		MapID:   4,
	}
	ws.AddGroundEffect(effect)

	if !ws.HasGroundEffectAt(100, 100, 4, 81157) {
		t.Fatal("同座標同 NPC ID 應視為已有地面效果")
	}
	if ws.HasGroundEffectAt(100, 100, 4, 81169) {
		t.Fatal("不同 NPC ID 不應被視為重複地面效果")
	}
}
