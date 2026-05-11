package system

import (
	"fmt"

	"github.com/l1jgo/server/internal/world"
)

func syncEquippedFlagFromSlots(player *world.PlayerInfo, item *world.InvItem) {
	if player == nil || item == nil {
		return
	}
	for slot := world.EquipSlot(1); slot < world.SlotMax; slot++ {
		if player.Equip.Get(slot) == item {
			item.Equipped = true
			return
		}
	}
}

func itemLogName(item *world.InvItem) string {
	if item == nil {
		return ""
	}
	name := item.Name
	if item.EnchantLvl > 0 {
		name = fmt.Sprintf("+%d %s", item.EnchantLvl, name)
	} else if item.EnchantLvl < 0 {
		name = fmt.Sprintf("%d %s", item.EnchantLvl, name)
	}
	if item.Stackable && item.Count > 1 {
		name = fmt.Sprintf("%s (%d)", name, item.Count)
	}
	return name
}
