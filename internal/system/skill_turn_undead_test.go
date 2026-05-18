package system

import (
	"testing"

	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillTurnUndeadRequiresJavaIsTUFlag(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID:   1,
		Session:     newSkillTestSession(t, 1),
		CharID:      1001,
		Name:        "caster",
		X:           100,
		Y:           100,
		MapID:       4,
		Level:       400,
		Intel:       12,
		MP:          100,
		MaxMP:       100,
		KnownSpells: []int32{18},
	})
	npc := &world.NpcInfo{
		ID:                2001,
		NpcID:             45000,
		Name:              "undead-but-tu-immune",
		X:                 101,
		Y:                 100,
		MapID:             4,
		HP:                30,
		MaxHP:             30,
		Undead:            true,
		UndeadType:        1,
		TurnUndeadable:    false,
		TurnUndeadableSet: true,
	}
	ws.AddNpc(npc)
	s := newEnchantWeaponTestSystem(t, ws)

	s.processSkill(handler.SkillRequest{
		SessionID: player.SessionID,
		SkillID:   18,
		TargetID:  npc.ID,
	})

	if npc.HP != 30 || npc.Dead {
		t.Fatalf("Java IsTU=false 的不死 NPC 不應被 TURN_UNDEAD 即死，HP=%d Dead=%v", npc.HP, npc.Dead)
	}
}
