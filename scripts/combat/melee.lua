-- Melee combat formula for PC attacking NPC
-- Receives a context table, returns {is_hit, damage}

function calc_melee_attack(ctx)
    local atk = ctx.attacker
    local tgt = ctx.target

    local str = atk.str
    local base_str = atk.base_str or str
    if base_str <= 0 then base_str = str end
    local level = atk.level
    local weapon_dmg = atk.weapon_dmg
    local hit_mod = atk.hit_mod or 0
    local dmg_mod = atk.dmg_mod or 0

    -- Default fist damage
    if weapon_dmg <= 0 then
        weapon_dmg = 4
    end

    -- Java: melee hit uses STR_HIT plus equipment/buff hit modifiers.
    local str_hit = table_lookup(STR_HIT, str)
    local pure_str = pure_stat_bonus(base_str)
    local hit_rate = level + str_hit + pure_str + hit_mod

    local attack_roll = math.random(1, 20) + hit_rate - 10
    local defense = 10 - tgt.ac

    local fumble = hit_rate - 9
    local critical = hit_rate + 10

    local is_hit = false
    if attack_roll <= fumble then
        is_hit = false
    elseif attack_roll >= critical then
        is_hit = true
    elseif attack_roll > defense then
        is_hit = true
    end

    -- Damage calculation: STR for damage
    local damage = 0
    if is_hit then
        local base = math.random(1, weapon_dmg)
        local crit_chance = table_lookup(STR_CRIT, str) + (atk.critical or 0)
        if base_str >= 45 then crit_chance = crit_chance + 1 end
        if crit_chance > 0 and math.random(0, 99) < crit_chance then
            base = weapon_dmg * 2
        end
        local str_dmg = table_lookup(STR_DMG, str)
        damage = base + str_dmg + pure_str + dmg_mod

        -- 職業 AC 防禦減傷（Java: L1AttackMode.java — 僅對玩家目標生效）
        local target_class = tgt.class_type or -1
        if target_class >= 0 then
            local ac_def = calc_ac_defense(target_class, tgt.ac)
            damage = damage - ac_def
        end

        -- Minimum 1 damage on hit
        if damage < 1 then
            damage = 1
        end
    end

    return { is_hit = is_hit, damage = damage }
end

---------------------------------------------------------------------
-- Ranged (bow) combat formula for PC attacking NPC
-- Java: L1Attack — bow uses DEX for both hit and damage
--
-- ctx.attacker = {level, str, dex, bow_dmg, arrow_dmg, bow_hit_mod, bow_dmg_mod}
-- ctx.target = {ac, level, mr}
---------------------------------------------------------------------
function calc_ranged_attack(ctx)
    local atk = ctx.attacker
    local tgt = ctx.target

    local dex = atk.dex
    local base_dex = atk.base_dex or dex
    if base_dex <= 0 then base_dex = dex end
    local level = atk.level
    local bow_dmg = atk.bow_dmg or 1
    local arrow_dmg = atk.arrow_dmg or 0
    local bow_hit_mod = atk.bow_hit_mod or 0
    local bow_dmg_mod = atk.bow_dmg_mod or 0

    -- Hit calculation: DEX is primary for ranged hit (Java: calcBowHit)
    local dex_hit = table_lookup(DEX_HIT, dex)
    local pure_dex = pure_stat_bonus(base_dex)
    local hit_rate = level + dex_hit + pure_dex + bow_hit_mod

    local attack_roll = math.random(1, 20) + hit_rate - 10
    local defense = 10 - tgt.ac

    local fumble = hit_rate - 9
    local critical = hit_rate + 10

    local is_hit = false
    if attack_roll <= fumble then
        is_hit = false
    elseif attack_roll >= critical then
        is_hit = true
    elseif attack_roll > defense then
        is_hit = true
    end
    is_hit = apply_er_evasion(is_hit, tgt.dodge)

    -- Damage calculation: Java ranged path adds DEX_DMG, bow damage and arrow/sting damage.
    local damage = 0
    if is_hit then
        local base = math.random(1, bow_dmg)
        local crit_chance = table_lookup(DEX_CRIT, dex) + (atk.bow_critical or 0)
        if base_dex >= 45 then crit_chance = crit_chance + 1 end
        if crit_chance > 0 and math.random(0, 99) < crit_chance then
            base = bow_dmg
        end
        local dex_dmg = table_lookup(DEX_DMG, dex)
        damage = base + dex_dmg + pure_dex + arrow_dmg + bow_dmg_mod

        -- 職業 AC 防禦減傷（僅對玩家目標生效）
        local target_class = tgt.class_type or -1
        if target_class >= 0 then
            local ac_def = calc_ac_defense(target_class, tgt.ac)
            damage = damage - ac_def
        end

        if damage < 1 then
            damage = 1
        end
    end

    return { is_hit = is_hit, damage = damage }
end
