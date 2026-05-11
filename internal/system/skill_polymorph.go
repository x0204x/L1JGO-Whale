package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/world"
)

var shapeChangeRandomPolyIDs = [...]int32{
	29, 945, 947, 979, 1037, 1039, 3860, 3861, 3862, 3863, 3864, 3865, 3904, 3906, 95,
	146, 2374, 2376, 2377, 2378, 3866, 3867, 3868, 3869, 3870, 3871, 3872, 3873, 3874, 3875, 3876,
}

var shapeChangeBlockedNpcIDs = map[int32]bool{
	45464: true,
	45473: true,
	45488: true,
	45497: true,
	45458: true,
	45752: true,
	45492: true,
	46035: true,
	99006: true,
}

func (s *SkillSystem) executeShapeChangePlayer(sess *net.Session, caster, target *world.PlayerInfo, skill *data.SkillInfo) {
	if caster == nil || target == nil {
		return
	}
	if caster.CharID == target.CharID {
		target.PendingPolySkill = true
		handler.SendShowPolyList(target.Session, target.CharID)
		return
	}
	if !sameClan(caster, target) && !shapeChangeProbability(caster, target.Level, target.MR) {
		s.sendCastFail(sess)
		return
	}
	if hasPolymorphControlRing(target) {
		target.PendingPolySkill = true
		handler.SendShowPolyList(target.Session, target.CharID)
		handler.SendServerMessage(target.Session, 966)
		return
	}
	if s.deps.Polymorph == nil {
		return
	}
	handler.SendServerMessageArgs(target.Session, 241, caster.Name)
	s.deps.Polymorph.DoPoly(target, randomShapeChangePolyID(), shapeChangeDuration(skill), data.PolyCauseMagic)
}

func (s *SkillSystem) executeShapeChangeNpc(sess *net.Session, caster *world.PlayerInfo, skill *data.SkillInfo, npc *world.NpcInfo, nearby []*world.PlayerInfo) {
	if caster == nil || npc == nil || npc.Dead {
		return
	}
	if npc.Level >= 50 || shapeChangeBlockedNpcIDs[npc.NpcID] {
		return
	}
	if !shapeChangeProbability(caster, npc.Level, npc.MR) {
		s.sendCastFail(sess)
		return
	}
	applyShapeChangeToNpc(npc, randomShapeChangePolyID(), shapeChangeDuration(skill))
	for _, viewer := range nearby {
		handler.SendChangeShape(viewer.Session, npc.ID, npc.GfxID, 0)
	}
}

func applyShapeChangeToNpc(npc *world.NpcInfo, polyID int32, durationSec int) {
	if npc == nil || polyID == 0 {
		return
	}
	if npc.PolyOriginalGfxID == 0 {
		npc.PolyOriginalGfxID = npc.GfxID
	}
	npc.GfxID = polyID
	if durationSec <= 0 {
		durationSec = 7200
	}
	npc.AddDebuff(handler.SkillShapeChange, durationSec*5)
}

func removeShapeChangeFromNpc(npc *world.NpcInfo) {
	if npc == nil || npc.PolyOriginalGfxID == 0 {
		return
	}
	npc.GfxID = npc.PolyOriginalGfxID
	npc.PolyOriginalGfxID = 0
}

func randomShapeChangePolyID() int32 {
	return shapeChangeRandomPolyIDs[world.RandInt(len(shapeChangeRandomPolyIDs))]
}

func shapeChangeDuration(skill *data.SkillInfo) int {
	if skill != nil && skill.BuffDuration > 0 {
		return skill.BuffDuration
	}
	return 7200
}

func shapeChangeProbability(caster *world.PlayerInfo, targetLevel int16, targetMR int16) bool {
	if caster == nil {
		return false
	}
	prob := 3*(int(caster.Level)-int(targetLevel)) + 200 - int(targetMR)
	if prob >= 100 {
		return true
	}
	if prob <= 0 {
		return false
	}
	return world.RandInt(100) < prob
}

func hasPolymorphControlRing(player *world.PlayerInfo) bool {
	if player == nil {
		return false
	}
	for _, slot := range []world.EquipSlot{world.SlotRing1, world.SlotRing2, world.SlotRing3, world.SlotRing4} {
		item := player.Equip.Get(slot)
		if item != nil && item.Equipped && (item.ItemID == 20281 || item.ItemID == 120281) {
			return true
		}
	}
	return false
}

func sameClan(a, b *world.PlayerInfo) bool {
	return a != nil && b != nil && a.ClanID != 0 && a.ClanID == b.ClanID
}
