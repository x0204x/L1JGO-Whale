-- Default NPC AI script
-- Called once per tick per alive L1Monster NPC
-- Receives AIContext table, returns array of AICommand tables
--
-- Command types:
--   "attack"         - melee attack current target
--   "ranged_attack"  - ranged attack current target
--   "skill"          - use skill {skill_id, act_id} on target
--   "move_toward"    - move 1 tile toward target
--   "wander"         - move 1 tile in direction {dir} (-1 = continue current)
--   "lose_aggro"     - clear aggro target
--   "idle"           - do nothing

function npc_ai(ctx)
    -- Has aggro target
    if ctx.target_id > 0 then
        return ai_with_target(ctx)
    end

    -- No target: idle wander
    if ctx.can_move then
        local dir = pick_wander_dir(ctx)
        return {{ type = "wander", dir = dir }}
    end

    return {{ type = "idle" }}
end

-- AI logic when NPC has a target
function ai_with_target(ctx)
    -- 非戰鬥 NPC（鹿/兔等）：無攻擊力 + 有移動力 + 低等級 → 逃跑
    if ctx.atk_speed == 0 and ctx.move_speed > 0 and ctx.level <= 90 then
        if ctx.target_dist > 17 then return {{ type = "lose_aggro" }} end
        if ctx.can_move then return {{ type = "flee" }} end
        return {{ type = "idle" }}
    end

    -- Target too far: lose aggro
    if ctx.target_dist > 15 then
        return {{ type = "lose_aggro" }}
    end

    -- Determine effective attack range
    local atk_range = 1
    if ctx.ranged > 1 then
        atk_range = ctx.ranged
    end

    local in_range = ctx.target_dist <= atk_range

    -- In attack range: fight or wait for cooldown (NEVER move)
    if in_range then
        if ctx.can_attack then
            -- Try mob skill first
            local skill_cmd = try_use_skill(ctx)
            if skill_cmd then
                return { skill_cmd }
            end

            -- Ranged NPC and target is further than melee: use ranged attack
            if ctx.ranged > 1 and ctx.target_dist > 1 then
                return {{ type = "ranged_attack" }}
            else
                return {{ type = "attack" }}
            end
        end
        -- In range but attack on cooldown: stand still and wait
        return {{ type = "idle" }}
    end

    -- Out of range: try skill that can reach, otherwise chase
    if ctx.can_attack then
        local skill_cmd = try_use_skill(ctx)
        if skill_cmd then
            return { skill_cmd }
        end
    end

    if ctx.can_move then
        return {{ type = "move_toward" }}
    end

    return {{ type = "idle" }}
end

-- Try to use a mob skill. Returns a command table or nil.
function try_use_skill(ctx)
    local skills = ctx.skills
    if not skills or #skills == 0 then
        return nil
    end

    local hp_pct = 100
    if ctx.max_hp > 0 then
        hp_pct = math.floor(ctx.hp * 100 / ctx.max_hp)
    end

    local candidates = {}
    for _, sk in ipairs(skills) do
        local ok = true
        local usable = false

        -- HP threshold check (0 = no threshold, otherwise only use when HP% <= trigger_hp)
        if ok and (sk.trigger_hp or 0) > 0 then
            if hp_pct <= sk.trigger_hp then
                usable = true
            else
                ok = false
            end
        end

        -- Go preselects companion target only when Yiwei's companion HP trigger already has a valid target.
        if ok and (sk.trigger_companion_hp or 0) > 0 then
            usable = true
        end

        -- Range check mirrors Yiwei L1MobSkill.isTriggerDistance().
        if ok and (sk.trigger_range or 0) ~= 0 then
            usable = true
            local target_dist = sk.target_dist
            if target_dist == nil or ((sk.companion_target_id or 0) == 0 and target_dist == 0 and (ctx.target_dist or 0) > 0) then
                target_dist = ctx.target_dist
            end
            if not mob_skill_trigger_distance_like_java(sk.trigger_range or 0, target_dist) then
                ok = false
            end
        end

        -- Trigger count is pre-filtered by Go; if present and not exhausted it makes the skill usable.
        if ok and (sk.trigger_count or 0) > 0 then
            usable = true
        end

        if ok and not usable then
            ok = false
        end

        -- MP check. Java special area skills type 5/6/7/8/9/10/11/12/13/14/15/16/17 do not check or consume MP.
        if ok and sk.type ~= 5 and sk.type ~= 6 and sk.type ~= 7 and sk.type ~= 8 and sk.type ~= 9 and sk.type ~= 10 and sk.type ~= 11 and sk.type ~= 12 and sk.type ~= 13 and sk.type ~= 14 and sk.type ~= 15 and sk.type ~= 16 and sk.type ~= 17 and sk.mp_consume > 0 and sk.mp_consume > ctx.mp then
            ok = false
        end

        -- Passed all useble checks. Java applies trigger_random only after one candidate is selected.
        if ok then
            local cmd = build_mob_skill_command(sk)
            if cmd then
                candidates[#candidates + 1] = {
                    command = cmd,
                    trigger_random = sk.trigger_random or 0,
                }
            end
        end
    end

    if #candidates == 0 then
        return nil
    end
    local selected = candidates[math.random(#candidates)]
    if selected.trigger_random <= 0 or selected.trigger_random > math.random(100) then
        return nil
    end
    return selected.command
end

function mob_skill_trigger_distance_like_java(trigger_range, distance)
    if trigger_range < 0 then
        return distance <= math.abs(trigger_range)
    end
    if trigger_range > 0 then
        return distance >= trigger_range
    end
    return false
end

function build_mob_skill_command(sk)
    -- type 5 = 範圍衝暈（Java: areashock_stun）
    if sk.type == 5 then
        return {
            type = "area_shock_stun",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 6 = 範圍魔法相消（Java: areacancellation）
    if sk.type == 6 then
        return {
            type = "area_cancellation",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 7 = Java L1MobSkillUse.weapon_break
    if sk.type == 7 then
        return {
            type = "area_weapon_break",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 8 = 範圍藥水侵蝕術（Java: potionturntodmg）
    if sk.type == 8 then
        return {
            type = "area_potion_turn_to_damage",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 9 = 範圍污濁的水流（Java: pollutewaterwave）
    if sk.type == 9 then
        return {
            type = "area_pollute_water_wave",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 10 = 範圍治癒侵蝕術（Java: healturntodmg）
    if sk.type == 10 then
        return {
            type = "area_heal_turn_to_damage",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 11 = 範圍沉默（Java: areasilence）
    if sk.type == 11 then
        return {
            type = "area_silence",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 12 = 範圍藥水霜化（Java: areadecaypotion）
    if sk.type == 12 then
        return {
            type = "area_decay_potion",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 13 = 範圍風之枷鎖（Java: areawindshackle）
    if sk.type == 13 then
        return {
            type = "area_wind_shackle",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            gfx_id = sk.gfx_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 14 = Java L1MobSkillUse.areadebuff
    if sk.type == 14 and sk.skill_id > 0 then
        return {
            type = "area_debuff",
            act_no = sk.act_no,
            skill_type = sk.type,
            skill_id = sk.skill_id,
            act_id = sk.act_id,
            gfx_id = sk.gfx_id,
            leverage = sk.leverage,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 15 = Java L1MobSkillUse.area_poison
    if sk.type == 15 and sk.summon_id > 0 then
        return {
            type = "area_poison",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            leverage = sk.leverage,
            summon_id = sk.summon_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 16 = Java L1MobSkillUse.SpawnEffect
    if sk.type == 16 and sk.summon_id > 0 then
        return {
            type = "spawn_effect",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            leverage = sk.leverage,
            summon_id = sk.summon_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 17 = Java L1MobSkillUse.areams, fixed CURSE_PARALYZE(33)
    if sk.type == 17 then
        return {
            type = "area_curse_paralyze",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 3 = 召喚技能
    if sk.type == 3 and sk.summon_id > 0 then
        return {
            type = "summon",
            act_no = sk.act_no,
            skill_type = sk.type,
            skill_id = sk.skill_id,
            gfx_id = sk.gfx_id,
            summon_id = sk.summon_id,
            summon_min = sk.summon_min,
            summon_max = sk.summon_max,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    -- type 4 = 群體變形（Java: poly）
    if sk.type == 4 and sk.poly_id > 0 then
        return {
            type = "poly",
            act_no = sk.act_no,
            skill_type = sk.type,
            act_id = sk.act_id,
            poly_id = sk.poly_id,
            reuse_delay = sk.reuse_delay or 0,
        }
    end

    return {
        type = "skill",
        act_no = sk.act_no,
        skill_type = sk.type,
        skill_id = sk.skill_id,
        act_id = sk.act_id,
        gfx_id = sk.gfx_id,
        leverage = sk.leverage or 0,
        change_target = sk.change_target or 0,
        trigger_range = sk.trigger_range or 0,
        reuse_delay = sk.reuse_delay or 0,
        range = sk.range or 0,
        area_width = sk.area_width or 0,
        area_height = sk.area_height or 0,
        companion_target_id = sk.companion_target_id or 0,
    }
end

-- Pick a wander direction.
-- Returns heading 0-7 for a new direction, or -1 to continue current direction.
function pick_wander_dir(ctx)
    -- Still walking in current direction
    if ctx.wander_dist > 0 then
        return -1
    end

    -- Far from spawn: bias toward spawn (Go handles actual heading calculation)
    if ctx.spawn_dist > 20 then
        return -2  -- special: Go will calculate heading toward spawn
    end

    -- Java L1NpcInstance.noTarget(): random 0-39, 0-7 moves and 8-39 pauses.
    return math.random(0, 39)
end
