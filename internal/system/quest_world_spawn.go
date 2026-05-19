package system

// 副本 Round 出生模組（MISS-P0-003 Stage D）
//
// Java 對照：L1QuestUser.mobSpawn 系列方法 + 各副本的 spawnXxx() 函式。
//
// 本檔負責：
//   - buildDungeonNpc：從 NpcTemplate 建立 Transient + ShowID 標記的副本 NPC
//   - spawnRound：執行單一 DungeonRound 的所有 spawn 規則（area / fixed / group_id）
//   - 廣播給副本內玩家（透過 GetNearbyPlayersInShow 的 ShowID 過濾）
//
// 設計原則：
//   - Transient=true → 不進入持久化、副本結束時直接銷毀
//   - ShowID=副本實例 ID → 確保 AOI 過濾不會洩漏到主世界或其他副本
//   - 重複觸發保護：呼叫前由 caller 透過 inst.MarkRoundSpawned 防止重出生
//   - 出生失敗（找不到合法點、模板不存在、count<=0）不影響其他 spawn

import (
	"math/rand"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// buildDungeonNpc 依模板建立副本內 Transient NPC。
// 對應 Java L1Spawn.doSpawn → 設定 _showId + 直接加入世界。
//
// 與 createMinion / SpawnKeeper 不同處：
//   - Transient=true
//   - ShowID=showID（AOI 隔離標籤）
//   - RespawnDelay=0（副本 NPC 不重生，全由 Round 引擎控制）
//   - 不寫 SpawnX/SpawnY 等重生欄位（避免被 NpcRespawnSystem 誤判）
func buildDungeonNpc(deps *handler.Deps, tmpl *data.NpcTemplate, x, y int32, mapID, heading int16, showID int32) *world.NpcInfo {
	if tmpl == nil {
		return nil
	}

	atkSpeed := tmpl.AtkSpeed
	moveSpeed := tmpl.PassiveSpeed
	if deps != nil && deps.SprTable != nil {
		gfx := int(tmpl.GfxID)
		if tmpl.AtkSpeed != 0 {
			if v := deps.SprTable.GetAttackSpeed(gfx, data.ActAttack); v > 0 {
				atkSpeed = int16(v)
			}
		}
		if tmpl.PassiveSpeed != 0 {
			if v := deps.SprTable.GetMoveSpeed(gfx, data.ActWalk); v > 0 {
				moveSpeed = int16(v)
			}
		}
	}

	return &world.NpcInfo{
		ID:                world.NextNpcID(),
		NpcID:             tmpl.NpcID,
		Impl:              tmpl.Impl,
		GfxID:             tmpl.GfxID,
		LightSize:         byte(tmpl.LightSize),
		Name:              tmpl.Name,
		NameID:            tmpl.NameID,
		Level:             tmpl.Level,
		X:                 x,
		Y:                 y,
		MapID:             mapID,
		Heading:           heading,
		HP:                tmpl.HP,
		MaxHP:             tmpl.HP,
		MP:                tmpl.MP,
		MaxMP:             tmpl.MP,
		AC:                tmpl.AC,
		STR:               tmpl.STR,
		DEX:               tmpl.DEX,
		Exp:               tmpl.Exp,
		Lawful:            tmpl.Lawful,
		Size:              tmpl.Size,
		MR:                tmpl.MR,
		Undead:            tmpl.Undead,
		UndeadType:        tmpl.UndeadType,
		TurnUndeadable:    tmpl.EffectiveTurnUndeadable(),
		TurnUndeadableSet: true,
		Hard:              tmpl.Hard,
		CantResurrect:     tmpl.CantResurrect,
		Agro:              tmpl.Agro,
		AtkDmg:            int32(tmpl.Level) + int32(tmpl.STR)/3,
		Ranged:            tmpl.Ranged,
		AtkSpeed:          atkSpeed,
		SubMagicSpeed:     tmpl.SubMagicSpeed,
		MoveSpeed:         moveSpeed,
		PoisonAtk:         tmpl.PoisonAtk,
		FireRes:           tmpl.FireRes,
		WaterRes:          tmpl.WaterRes,
		WindRes:           tmpl.WindRes,
		EarthRes:          tmpl.EarthRes,
		WeakAttr:          tmpl.WeakAttr,
		RespawnDelay:      0, // 副本 NPC 不自動重生
		ShowID:            showID,
		Transient:         true,
	}
}

// spawnRound 執行單一 Round 的所有 spawn 規則。
// 回傳實際出生的 NPC 數量（含 group_id 帶出的隊員）。
//
// 流程：
//  1. 對每個 DungeonSpawn 抽 Count 次位置
//  2. 建 NPC → 加入世界 → 加入 inst.npcs → 封鎖地圖格子 → 廣播給副本內玩家
//  3. 若 group_id > 0：再依 MobGroup 定義生成隊員（同樣標記 Transient + ShowID）
//
// 呼叫前 caller 應透過 inst.MarkRoundSpawned(round.ID) 確保不重複觸發。
func (s *QuestWorldSystem) spawnRound(inst *world.QuestInstance, round *data.DungeonRound) int {
	if inst == nil || round == nil {
		return 0
	}
	if s.deps == nil || s.deps.Npcs == nil {
		return 0
	}

	spawned := 0
	for i := range round.Spawns {
		sp := &round.Spawns[i]
		tmpl := s.deps.Npcs.Get(sp.NpcID)
		if tmpl == nil {
			s.logf("副本 Round spawn：找不到 NPC 模板",
				zap.Int32("show_id", inst.ID),
				zap.Int32("round_id", round.ID),
				zap.Int32("npc_id", sp.NpcID),
			)
			continue
		}

		for j := int32(0); j < sp.Count; j++ {
			x, y, ok := s.pickSpawnPoint(sp, inst.MapID)
			if !ok {
				s.logf("副本 Round spawn：找不到合法生成點",
					zap.Int32("show_id", inst.ID),
					zap.Int32("round_id", round.ID),
					zap.Int32("npc_id", sp.NpcID),
				)
				continue
			}

			heading := sp.Heading
			if heading == 0 {
				heading = int16(rand.Intn(8))
			}

			npc := buildDungeonNpc(s.deps, tmpl, x, y, inst.MapID, heading, inst.ID)
			if npc == nil {
				continue
			}
			s.addDungeonNpc(inst, npc)
			spawned++

			// 群體隊員（leader 帶隊員出生）
			if sp.GroupID > 0 && s.deps.MobGroups != nil {
				if grp := s.deps.MobGroups.Get(sp.GroupID); grp != nil {
					spawned += s.spawnGroupMinions(inst, npc, grp)
				}
			}
		}
	}

	if spawned > 0 {
		s.logf("副本 Round 出生完成",
			zap.Int32("show_id", inst.ID),
			zap.Int32("round_id", round.ID),
			zap.Int("count", spawned),
		)
	}
	return spawned
}

// pickSpawnPoint 從 DungeonSpawn 抽合法生成點。
// 優先順序：area > fixed > rule.X/Y。
func (s *QuestWorldSystem) pickSpawnPoint(sp *data.DungeonSpawn, mapID int16) (int32, int32, bool) {
	rule := NpcSpawnRule{MapID: mapID}

	hasArea := len(sp.Area) == 4 && (sp.Area[0] != 0 || sp.Area[1] != 0 || sp.Area[2] != 0 || sp.Area[3] != 0)
	if hasArea {
		rule.LocX1 = sp.Area[0]
		rule.LocY1 = sp.Area[1]
		rule.LocX2 = sp.Area[2]
		rule.LocY2 = sp.Area[3]
		// FindNpcSpawnPoint 在無 X/Y 時也能透過 area 候選點工作
		rule.X = (sp.Area[0] + sp.Area[2]) / 2
		rule.Y = (sp.Area[1] + sp.Area[3]) / 2
	} else if len(sp.Fixed) == 2 {
		rule.X = sp.Fixed[0]
		rule.Y = sp.Fixed[1]
	} else {
		return 0, 0, false
	}

	var maps npcSpawnMap
	if s.deps != nil && s.deps.MapData != nil {
		maps = s.deps.MapData
	}
	return FindNpcSpawnPoint(rule, s.ws, maps, 0, nil)
}

// addDungeonNpc 將副本 NPC 加入世界 + 副本實例的 NPC 清單 + 廣播給副本內玩家。
func (s *QuestWorldSystem) addDungeonNpc(inst *world.QuestInstance, npc *world.NpcInfo) {
	if s.ws == nil || inst == nil || npc == nil {
		return
	}
	s.ws.AddNpc(npc)
	inst.AddNpc(npc.ID)

	if s.deps != nil && s.deps.MapData != nil {
		s.deps.MapData.SetImpassable(npc.MapID, npc.X, npc.Y, true)
	}

	// 廣播給同副本內 + 同地圖的玩家（ShowID 過濾自動排除主世界玩家）
	viewers := s.ws.GetNearbyPlayersInShow(npc.X, npc.Y, npc.MapID, 0, inst.ID)
	for _, viewer := range viewers {
		handler.SendNpcPack(viewer.Session, npc)
	}
}

// spawnGroupMinions 為已生成的 leader 帶出群體隊員。
// 對應 Java L1MobGroupSpawn.doSpawn 的副本場景變體（每個隊員也標 ShowID + Transient）。
func (s *QuestWorldSystem) spawnGroupMinions(inst *world.QuestInstance, leader *world.NpcInfo, grp *data.MobGroup) int {
	if leader == nil || grp == nil {
		return 0
	}

	groupInfo := &world.MobGroupInfo{
		Leader:             leader,
		Members:            []*world.NpcInfo{leader},
		RemoveGroupOnDeath: grp.RemoveGroupIfLeaderDie,
	}
	leader.GroupInfo = groupInfo

	count := 0
	for _, minion := range grp.Minions {
		if minion.NpcID == 0 || minion.Count == 0 {
			continue
		}
		mTmpl := s.deps.Npcs.Get(minion.NpcID)
		if mTmpl == nil {
			continue
		}
		for j := 0; j < minion.Count; j++ {
			rule := NpcSpawnRule{
				MapID: leader.MapID,
				X:     leader.X,
				Y:     leader.Y,
				LocX1: leader.X - 2,
				LocY1: leader.Y - 2,
				LocX2: leader.X + 3,
				LocY2: leader.Y + 3,
			}
			var maps npcSpawnMap
			if s.deps.MapData != nil {
				maps = s.deps.MapData
			}
			mx, my, ok := FindNpcSpawnPoint(rule, s.ws, maps, 0, nil)
			if !ok {
				continue
			}

			mob := buildDungeonNpc(s.deps, mTmpl, mx, my, leader.MapID, leader.Heading, inst.ID)
			if mob == nil {
				continue
			}
			mob.IsMinion = true

			s.addDungeonNpc(inst, mob)
			groupInfo.Members = append(groupInfo.Members, mob)
			count++
		}
	}
	return count
}
