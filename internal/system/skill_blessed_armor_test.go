package system

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/world"
)

func TestSkillBlessedArmorUnequippedArmorEnchantExpires(t *testing.T) {
	ws := world.NewState()
	player := addSkillTestPlayer(ws, &world.PlayerInfo{
		SessionID: 1,
		Session:   newSkillTestSession(t, 1),
		CharID:    1001,
		Name:      "caster",
		X:         100,
		Y:         100,
		MapID:     4,
	})
	armor := player.Inv.AddItemWithID(7001, 20089, 1, "小藤甲", 0, 1000, false, 1)
	s := newEnchantWeaponTestSystem(t, ws)

	s.executeArmorEnchant(player.Session, player, &data.SkillInfo{
		SkillID:      21,
		BuffDuration: 1,
		ActionID:     19,
		CastGfx:      748,
	}, armor.ObjectID)

	if armor.AcByMagic != 3 || armor.AcMagicExpiry != 5 {
		t.Fatalf("鎧甲護持應套用 1 秒防具附魔，ac=%d expiry=%d", armor.AcByMagic, armor.AcMagicExpiry)
	}
	for i := 0; i < 5; i++ {
		tickItemMagicEnchants(player, s.deps)
	}
	if armor.AcByMagic != 0 || armor.AcMagicExpiry != 0 {
		t.Fatalf("Java 物品附魔計時器即使未裝備也應到期清除，ac=%d expiry=%d", armor.AcByMagic, armor.AcMagicExpiry)
	}
}
