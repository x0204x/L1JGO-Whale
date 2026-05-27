package main

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/system"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

func spawnNpcsSafe(ws *world.State, npcTable *data.NpcTable, spawns []data.SpawnEntry, maps *data.MapDataTable, sprTable *data.SprTable, mobGroups *data.MobGroupTable, log *zap.Logger) int {
	total := 0
	for _, spawn := range spawns {
		tmpl := npcTable.Get(spawn.NpcID)
		if tmpl == nil {
			log.Warn("NPC生成略過：找不到NPC樣板", zap.Int32("npc_id", spawn.NpcID))
			continue
		}

		rule := npcSpawnRuleFromEntry(spawn)
		for i := 0; i < spawn.Count; i++ {
			x, y, ok := system.FindNpcSpawnPoint(rule, ws, maps, 0, nil)
			if !ok {
				log.Warn("NPC生成失敗：找不到可通行座標",
					zap.Int32("npc_id", spawn.NpcID),
					zap.Int16("map_id", spawn.MapID),
					zap.Int32("x", spawn.X),
					zap.Int32("y", spawn.Y))
				continue
			}

			leader := createNpcFromTemplate(tmpl, x, y, spawn.MapID, spawn.Heading, spawn.RespawnDelay, sprTable)
			system.ApplyNpcInitialHideLikeJava(leader)
			leader.MobGroupID = spawn.MobGroupID
			system.ApplyNpcSpawnRule(leader, rule)
			ws.AddNpc(leader)
			if maps != nil {
				maps.SetImpassable(leader.MapID, leader.X, leader.Y, true)
			}
			total++

			if spawn.MobGroupID > 0 && mobGroups != nil {
				group := mobGroups.Get(spawn.MobGroupID)
				if group != nil {
					total += spawnMobGroupSafe(ws, leader, group, npcTable, maps, sprTable, log)
				}
			}
		}
	}
	return total
}

func npcSpawnRuleFromEntry(spawn data.SpawnEntry) system.NpcSpawnRule {
	rule := system.NpcSpawnRule{
		MapID:         spawn.MapID,
		X:             spawn.X,
		Y:             spawn.Y,
		Count:         spawn.Count,
		RandomX:       spawn.RandomX,
		RandomY:       spawn.RandomY,
		LocX1:         spawn.LocX1,
		LocY1:         spawn.LocY1,
		LocX2:         spawn.LocX2,
		LocY2:         spawn.LocY2,
		AvoidPC:       spawn.AvoidPC,
		RespawnScreen: spawn.RespawnScreen,
	}
	if spawn.Spread == "point" {
		rule.Count = 1
		rule.RandomX = 0
		rule.RandomY = 0
		rule.LocX1 = 0
		rule.LocY1 = 0
		rule.LocX2 = 0
		rule.LocY2 = 0
	}
	return rule
}

func spawnMobGroupSafe(ws *world.State, leader *world.NpcInfo, group *data.MobGroup, npcTable *data.NpcTable, maps *data.MapDataTable, sprTable *data.SprTable, log *zap.Logger) int {
	groupInfo := &world.MobGroupInfo{
		Leader:             leader,
		Members:            []*world.NpcInfo{leader},
		RemoveGroupOnDeath: group.RemoveGroupIfLeaderDie,
	}
	leader.GroupInfo = groupInfo

	spawned := 0
	for _, minion := range group.Minions {
		if minion.NpcID == 0 || minion.Count == 0 {
			continue
		}
		mTmpl := npcTable.Get(minion.NpcID)
		if mTmpl == nil {
			continue
		}
		for j := 0; j < minion.Count; j++ {
			rule := system.NpcSpawnRule{
				MapID: leader.MapID,
				X:     leader.X,
				Y:     leader.Y,
				LocX1: leader.X - 2,
				LocY1: leader.Y - 2,
				LocX2: leader.X + 3,
				LocY2: leader.Y + 3,
			}
			mx, my, ok := system.FindNpcSpawnPoint(rule, ws, maps, 0, nil)
			if !ok {
				log.Warn("群怪生成失敗：找不到可通行座標",
					zap.Int32("leader_id", leader.ID),
					zap.Int32("npc_id", minion.NpcID),
					zap.Int16("map_id", leader.MapID),
					zap.Int32("x", leader.X),
					zap.Int32("y", leader.Y))
				continue
			}

			mob := createNpcFromTemplate(mTmpl, mx, my, leader.MapID, leader.Heading, 0, sprTable)
			mob.IsMinion = true
			mob.GroupInfo = groupInfo
			mob.SpawnX = mx
			mob.SpawnY = my
			mob.SpawnMapID = leader.SpawnMapID

			ws.AddNpc(mob)
			if maps != nil {
				maps.SetImpassable(mob.MapID, mob.X, mob.Y, true)
			}
			groupInfo.Members = append(groupInfo.Members, mob)
			spawned++
		}
	}
	return spawned
}
