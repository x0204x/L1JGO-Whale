package system

import (
	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/handler"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/scripting"
	"github.com/l1jgo/server/internal/world"
)

// ========================================================================
//  自身技能
// ========================================================================

// executeSelfSkill 處理自身目標技能（護盾、光明、冥想等）。
func (s *SkillSystem) executeSelfSkill(sess *net.Session, player *world.PlayerInfo, skill *data.SkillInfo) {
	nearby := s.deps.World.GetNearbyPlayersAt(player.X, player.Y, player.MapID)

	switch skill.SkillID {
	case 2: // 日光術
		// 有持續時間但無屬性變化，由 applyBuffEffect 處理

	case 13, 72: // 無所遁形術 / 強力無所遁形術 — 揭示附近隱身玩家
		// Java 參考: L1SkillUse.detection() — 移除 buff 60（隱身術）和 97（暗影閃避）
		// 注意: GM 隱身免疫（isGmInvis），待 GM 系統實作後加入
		for _, tgt := range nearby {
			if tgt.CharID == player.CharID {
				continue
			}
			if tgt.HasBuff(60) || tgt.HasBuff(97) {
				s.removeBuffAndRevert(tgt, 60)
				s.removeBuffAndRevert(tgt, 97)
				// 通知被揭示者：你不再隱身
				handler.SendInvisible(tgt.Session, tgt.CharID, false)
				// 向附近玩家廣播角色出現（讓被揭示者重新顯示在其他人畫面上）
				nearbyOfTarget := s.deps.World.GetNearbyPlayersAt(tgt.X, tgt.Y, tgt.MapID)
				for _, viewer := range nearbyOfTarget {
					if viewer.CharID != tgt.CharID {
						handler.SendPutObject(viewer.Session, tgt)
					}
				}
			}
		}
		// 施法者自己若在隱身中也會被揭示
		if player.HasBuff(60) || player.HasBuff(97) {
			s.removeBuffAndRevert(player, 60)
			s.removeBuffAndRevert(player, 97)
			handler.SendInvisible(sess, player.CharID, false)
		}

	case 44: // 魔法相消術（自身）
		s.cancelAllBuffs(player)

	case 60: // 隱身術 — 施法者隱形，直到攻擊/施法/使用道具時解除
		player.Invisible = true
		invisBuff := &world.ActiveBuff{
			SkillID:      skill.SkillID,
			TicksLeft:    3600 * 5, // 永久直到行動解除（cancelInvisibility）
			SetInvisible: true,
		}
		old60 := player.AddBuff(invisBuff)
		if old60 != nil {
			s.revertBuffStats(player, old60)
		}
		// 通知施法者已隱身
		handler.SendInvisible(sess, player.CharID, true)
		// 從附近所有玩家畫面移除（Java: S_RemoveObject）
		removeData := handler.BuildRemoveObject(player.CharID)
		for _, viewer := range nearby {
			if viewer.CharID != player.CharID {
				viewer.Session.Send(removeData)
			}
		}

	case 78: // 絕對屏障 — 免疫所有傷害，停止 HP/MP 回復
		// Java: 攻擊/施法/使用道具/裝備武器時解除；移動時不解除
		player.AbsoluteBarrier = true
		dur := skill.BuffDuration
		if dur <= 0 {
			dur = 12
		}
		abBuff := &world.ActiveBuff{
			SkillID:            skill.SkillID,
			TicksLeft:          dur * 5,
			SetAbsoluteBarrier: true,
		}
		old78 := player.AddBuff(abBuff)
		if old78 != nil {
			s.revertBuffStats(player, old78)
		}
		// 廣播屏障特效（Java: castGfx 2234）
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
		}

	case 161: // 封印禁地 — 對附近玩家套用沉默，施法者不受影響
		handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))
		s.applyAreaOfSilence(player, skill, nearby)
		return

	case 90: // 堅固防護 — Java 需要盾牌/臂甲，效果為 ER +15
		if !s.validateSolidCarriage(player) {
			handler.SendServerMessage(sess, skillMsgCastFail)
			return
		}

	case 130: // 心靈轉換 — 恢復 2 MP（Java: BODY_TO_MIND +2）
		player.MP += 2
		if player.MP > player.MaxMP {
			player.MP = player.MaxMP
		}
		sendMpUpdate(sess, player)

	case 146: // 魂體轉換 — 增加當前 MP（Java: BLOODY_SOUL，使用 skill_level 匹配客戶端顯示）
		addMP := int32(skill.SkillLevel)
		player.MP += addMP
		if player.MP > player.MaxMP {
			player.MP = player.MaxMP
		}
		sendMpUpdate(sess, player)

	case 186: // 血之渴望 — 自身 buff + 勇敢速度（Java: BLOODLUST.java）
		// 與暴風疾走/聖潔之行等互斥（由 buffs.lua 無 exclusions，brave_speed 會覆蓋）
		s.applyBuffEffect(player, skill)
		if skill.CastGfx > 0 {
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
		}
		handler.SendPlayerStatus(player.Session, player)
		if skill.SysMsgHappen > 0 {
			handler.SendServerMessage(sess, uint16(skill.SysMsgHappen))
		}

	case 172: // 暴風疾走
		for _, conflictID := range []int32{
			handler.SkillStatusBrave, handler.SkillStatusElfBrave,
			42,  // HOLY_WALK
			101, // MOVING_ACCELERATION
			150, // WIND_WALK
			186, // BLOOD_LUST
		} {
			s.removeBuffAndRevert(player, conflictID)
		}
		stormBuff := &world.ActiveBuff{
			SkillID:       172,
			TicksLeft:     300 * 5,
			SetBraveSpeed: 4,
		}
		old172 := player.AddBuff(stormBuff)
		if old172 != nil {
			s.revertBuffStats(player, old172)
		}
		player.BraveSpeed = 4
		player.BraveTicks = stormBuff.TicksLeft
		s.sendSpeedToAll(player, 4, 300)
	}

	// 廣播施法動畫
	handler.BroadcastToPlayers(nearby, handler.BuildActionGfx(player.CharID, byte(skill.ActionID)))

	// 自身範圍治療
	if skill.Type == 16 && (skill.DamageValue > 0 || skill.DamageDice > 0) {
		casterINT := int(player.Intel)
		casterSP := int(player.SP)

		if skill.Area == -1 {
			heal := int32(s.deps.Scripting.CalcHeal(skill.DamageValue, skill.DamageDice, skill.DamageDiceCount, casterINT, casterSP))
			heal = s.applyElfWaterHealingModifiers(player, heal)
			if heal > 0 && player.HP < player.MaxHP {
				player.HP += heal
				if player.HP > player.MaxHP {
					player.HP = player.MaxHP
				}
				sendHpUpdate(sess, player)
			}
			for _, p := range nearby {
				if p.SessionID == sess.ID {
					continue
				}
				h := int32(s.deps.Scripting.CalcHeal(skill.DamageValue, skill.DamageDice, skill.DamageDiceCount, casterINT, casterSP))
				h = s.applyElfWaterHealingModifiers(p, h)
				if h > 0 && p.HP < p.MaxHP {
					p.HP += h
					if p.HP > p.MaxHP {
						p.HP = p.MaxHP
					}
					sendHpUpdate(p.Session, p)
				}
			}
		} else {
			heal := int32(s.deps.Scripting.CalcHeal(skill.DamageValue, skill.DamageDice, skill.DamageDiceCount, casterINT, casterSP))
			heal = s.applyElfWaterHealingModifiers(player, heal)
			if heal > 0 && player.HP < player.MaxHP {
				player.HP += heal
				if player.HP > player.MaxHP {
					player.HP = player.MaxHP
				}
				sendHpUpdate(sess, player)
			}
		}
	}

	// 自身範圍 AoE 傷害（龍捲風 53、震裂術 62、冰雪颶風 80 等）
	if skill.Type == 64 && skill.Area > 0 && (skill.DamageValue > 0 || skill.DamageDice > 0) {
		s.applySelfAreaSkillDamageToPlayers(sess, player, skill, nearby)
		nearbyNpcs := s.deps.World.GetNearbyNpcs(player.X, player.Y, player.MapID)
		for _, npc := range nearbyNpcs {
			if npc.Dead {
				continue
			}
			if chebyshevDist(player.X, player.Y, npc.X, npc.Y) > int32(skill.Area) {
				continue
			}
			ctx := scripting.SkillDamageContext{
				SkillID:            int(skill.SkillID),
				DamageValue:        skill.DamageValue,
				DamageDice:         skill.DamageDice,
				DamageDiceCount:    skill.DamageDiceCount,
				SkillLevel:         skill.SkillLevel,
				Attr:               skill.Attr,
				AttackerLevel:      int(player.Level),
				AttackerSTR:        int(player.Str),
				AttackerDEX:        int(player.Dex),
				AttackerINT:        int(player.Intel),
				AttackerWIS:        int(player.Wis),
				AttackerSP:         int(player.SP),
				AttackerDmgMod:     int(player.DmgMod),
				AttackerHitMod:     int(player.HitMod),
				AttackerMagicLevel: calcMagicLevel(int(player.ClassType), int(player.Level)),
				TargetAC:           int(npc.AC),
				TargetLevel:        int(npc.Level),
				TargetMR:           int(npc.MR),
				TargetFireRes:      int(npc.FireRes),
				TargetWaterRes:     int(npc.WaterRes),
				TargetWindRes:      int(npc.WindRes),
				TargetEarthRes:     int(npc.EarthRes),
			}
			res := s.deps.Scripting.CalcSkillDamage(ctx)
			dmg := int32(res.Damage)
			handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(npc.ID, skill.CastGfx))
			// 浮動傷害數字（自我範圍攻擊技能，魔防抵抗時顯示 MISS）
			if player.AttackView {
				handler.SendDamageNumbers(sess, npc.ID, dmg)
			}
			npc.HP -= dmg
			if npc.HP < 0 {
				npc.HP = 0
			}
			// 攻擊技能傷害累加仇恨
			AddHate(npc, sess.ID, dmg)
			hpRatio := int16(0)
			if npc.MaxHP > 0 {
				hpRatio = int16((npc.HP * 100) / npc.MaxHP)
			}
			handler.BroadcastToPlayers(nearby, handler.BuildHpMeter(npc.ID, hpRatio))
			if npc.HP <= 0 {
				handleNpcDeath(npc, player, nearby, s.deps)
				continue
			}

			// 冰雪颶風：傷害後凍結判定（Java: calcProbabilityMagic → setFrozen + S_Poison 灰色）
			if skill.SkillID == 80 && !npc.Paralyzed && !npc.HasDebuff(50) && !npc.HasDebuff(80) {
				if s.checkNpcMRResist(player, npc, skill.SkillID) {
					dur := skill.BuffDuration
					if dur <= 0 {
						dur = 16
					}
					npc.Paralyzed = true
					npc.AddDebuff(80, (dur+1)*5)
					handler.BroadcastToPlayers(nearby, handler.BuildPoison(npc.ID, 2))
				}
			}
		}
	}

	// Owner: skill_clan.go
	if s.applyRoyalAuraSkill(player, skill, nearby) {
		return
	}

	// 套用 buff 效果
	s.applyBuffEffect(player, skill)

	// 負重強化：套用時立即更新負重顯示
	if skill.SkillID == 14 || skill.SkillID == 218 {
		handler.SendWeightUpdate(sess, player)
	}

	// 效果 GFX
	if skill.CastGfx > 0 {
		handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
	}

	if skill.SysMsgHappen > 0 {
		handler.SendServerMessage(sess, uint16(skill.SysMsgHappen))
	}
}
