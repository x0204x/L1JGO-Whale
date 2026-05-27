-- Skill damage formulas for PC -> NPC attack skills
-- Entry point: calc_skill_damage(ctx) routes to the correct sub-formula.
--
-- ctx.skill = {id, damage_value, damage_dice, damage_dice_count, skill_level, attr}
-- ctx.attacker = {level, str, dex, intel, wis, sp, dmg_mod, hit_mod, weapon_dmg, hp, max_hp}
-- ctx.target = {ac, level, mr, fire_res, water_res, wind_res, earth_res, mp}
--
-- Returns: {damage, drain_mp, hit_count}

-- Element constants (matching YAML attr values)
local ATTR_EARTH = 1
local ATTR_FIRE  = 2
local ATTR_WATER = 4
local ATTR_WIND  = 8

---------------------------------------------------------------------
-- Main routing function
---------------------------------------------------------------------
function calc_skill_damage(ctx)
    local sk = ctx.skill
    local sid = sk.id

    -- Special skills with unique formulas
    if sid == 207 then return calc_mind_break(ctx) end
    if sid == 218 then return calc_joy_of_pain(ctx) end

    -- Physical skills: damage_value == 0 AND damage_dice == 0
    if sk.damage_value == 0 and sk.damage_dice == 0 then
        return calc_physical_skill(ctx)
    end

    -- Magic damage: has damage_value or damage_dice
    return calc_magic_damage(ctx)
end

---------------------------------------------------------------------
-- Magic damage formula (wizard spells, DK/IL spells with dice)
-- Based on Java L1Magic.calcMagicDiceDamage + calcAttrResistance + calcMrDefense
---------------------------------------------------------------------
function calc_magic_damage(ctx)
    local sk = ctx.skill
    local atk = ctx.attacker
    local tgt = ctx.target

    -- Stage 1: Base damage + dice rolls
    local damage = sk.damage_value
    if sk.damage_dice > 0 and sk.damage_dice_count > 0 then
        for i = 1, sk.damage_dice_count do
            damage = damage + math.random(1, sk.damage_dice)
        end
    end

    -- Stage 2: INT + SP coefficient (Java: charaIntelligence)
    local base_int = atk.base_int or atk.intel
    if base_int <= 0 then base_int = atk.intel end
    local char_intel = atk.intel + atk.sp - 12
    if char_intel < 1 then char_intel = 1 end

    -- Stage 3: Elemental resistance
    local attr_defence = calc_attr_resistance(sk.attr, tgt)

    -- Stage 4: Apply coefficient = (1.0 - attrDefence + INT*3/32)
    local coefficient = 1.0 - attr_defence + char_intel * 3.0 / 32.0
    if coefficient < 0 then coefficient = 0 end
    damage = math.floor(damage * coefficient)

    -- Stage 5: Magic critical (Java: level 1-6 spells or DISINTEGRATE)
    if (sk.skill_level >= 1 and sk.skill_level <= 6) or sk.id == 77 then
        local crit_chance = table_lookup(INT_CRIT, atk.intel) + (atk.magic_critical or 0)
        if base_int >= 45 then crit_chance = crit_chance + 1 end
        if crit_chance > 0 and math.random(1, 100) < crit_chance then
            damage = math.floor(damage * 1.5)
        end
    end

    local pure_int = pure_stat_bonus(base_int)
    if pure_int > 0 then
        damage = damage + pure_int * 2
    end

    -- Stage 6: MR defense
    local magic_hit = table_lookup(INT_MAGIC_HIT, atk.intel) + (atk.original_magic_hit or 0)
    magic_hit = magic_hit + pure_stat_bonus(base_int)
    damage = apply_mr_defense(damage, tgt.mr, magic_hit)

    if damage < 1 then damage = 1 end
    return { damage = damage, drain_mp = 0, hit_count = 1 }
end

---------------------------------------------------------------------
-- Physical skill damage (108, 132, 208)
-- Uses STR/DEX/weapon tables from tables.lua, not INT/dice
-- 注意：203 SMASH 已改走 calc_magic_damage（Java 為 TYPE_ATTACK 資料驅動，
-- damage_value=12 + damage_dice=10 + attr=16 ATTR_RAY），不在此清單。
---------------------------------------------------------------------
function calc_physical_skill(ctx)
    local sk = ctx.skill
    local atk = ctx.attacker
    local tgt = ctx.target
    local sid = sk.id

    local str = atk.str
    local base_str = atk.base_str or str
    if base_str <= 0 then base_str = str end
    local level = atk.level
    local weapon_dmg = atk.weapon_dmg
    local hit_mod = atk.hit_mod or 0
    local dmg_mod = atk.dmg_mod or 0

    if weapon_dmg <= 0 then weapon_dmg = 4 end

    -- Java physical skill melee path uses STR_HIT plus equipment/buff hit modifiers.
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

    local damage = 0
    local hit_count = 1

    if is_hit then
        local base = math.random(1, weapon_dmg)
        local str_dmg = table_lookup(STR_DMG, str)
        damage = base + str_dmg + pure_str + dmg_mod

        -- Skill-specific bonuses
        -- 注：三重矢（132）已改走 combat.go 的 ExecuteRangedAttackOnNpc × 3
        -- + pvp.go 的 HandlePvPFarAttack × 3 完整弓箭路徑（含 DEX_DMG / 箭矢 / 武器 buff），
        -- 不再走本函式；亦不套用 Java ConfigSkill.TRIPLE_ARROW_DMG=5 倍率（Go 設計刻意捨棄該倍率）。
        if sid == 108 then      -- Critical Strike: guaranteed extra damage
            damage = damage + level + math.floor(str / 3)
        elseif sid == 208 then  -- Bone Break/Skull Destruction
            damage = damage + math.floor(level / 3)
        end

        if damage < 1 then damage = 1 end
    end

    return { damage = damage, drain_mp = 0, hit_count = hit_count }
end

---------------------------------------------------------------------
-- Mind Break (207): drains 5 MP from target, deals SP*3.8 damage
---------------------------------------------------------------------
function calc_mind_break(ctx)
    local atk = ctx.attacker
    local drain_mp = 5
    local true_sp = atk.true_sp or 0
    local full_sp = atk.full_sp or (true_sp + (atk.sp or 0))
    local damage = math.floor(full_sp * 3.8)

    if damage < 0 then damage = 0 end
    return { damage = damage, drain_mp = drain_mp, hit_count = 1 }
end

---------------------------------------------------------------------
-- Joy of Pain (218): Go primes the caster-side one-shot state.
---------------------------------------------------------------------
function calc_joy_of_pain(ctx)
    return { damage = 0, drain_mp = 0, hit_count = 1 }
end

---------------------------------------------------------------------
-- 魔法元素抗性計算
-- Java: L1MagicMode.calcAttrResistance() → abs(resist) / 10.0
-- 正負抗性都是減傷（取絕對值）
-- 用於 coefficient = max(1.0 - attrDefence + INT*3/32, 0)
---------------------------------------------------------------------
function calc_attr_resistance(attr, tgt)
    local resist = 0
    if attr == ATTR_EARTH then
        resist = tgt.earth_res or 0
    elseif attr == ATTR_FIRE then
        resist = tgt.fire_res or 0
    elseif attr == ATTR_WATER then
        resist = tgt.water_res or 0
    elseif attr == ATTR_WIND then
        resist = tgt.wind_res or 0
    else
        return 0
    end

    return math.abs(resist) / 10.0
end

---------------------------------------------------------------------
-- MR defense calculation (PC caster formula from Java L1MagicPc.calcMrDefense)
---------------------------------------------------------------------
function apply_mr_defense(damage, mr, magic_hit)
    if mr <= 0 then return damage end
    magic_hit = magic_hit or 0

    local mr_coefficient
    if mr < 100 then
        mr_coefficient = 1 - 0.01 * math.floor((mr - magic_hit) / 2)
    else
        mr_coefficient = 0.6 - 0.01 * math.floor((mr - magic_hit) / 10)
    end

    if mr_coefficient < 0 then mr_coefficient = 0 end
    return math.floor(damage * mr_coefficient)
end

---------------------------------------------------------------------
-- Heal amount calculation
-- calc_heal_amount(damage_value, damage_dice, damage_dice_count, intel, lawful, leverage)
-- Returns heal amount (int)
---------------------------------------------------------------------
function calc_heal_amount(damage_value, damage_dice, damage_dice_count, intel, lawful, leverage)
    lawful = lawful or 0
    leverage = leverage or 10
    if leverage <= 0 then leverage = 10 end

    local magic_bonus = 0
    if intel <= 5 then
        magic_bonus = -2
    elseif intel <= 8 then
        magic_bonus = -1
    elseif intel <= 11 then
        magic_bonus = 0
    elseif intel <= 14 then
        magic_bonus = 1
    elseif intel <= 17 then
        magic_bonus = 2
    elseif intel <= 24 then
        magic_bonus = intel - 15
    elseif intel <= 35 then
        magic_bonus = 10
    elseif intel <= 42 then
        magic_bonus = 11
    elseif intel <= 49 then
        magic_bonus = 12
    else
        magic_bonus = 13
    end
    if magic_bonus > 10 then magic_bonus = 10 end

    local heal = 0
    local dice_count = damage_value + magic_bonus
    if damage_dice > 0 and dice_count > 0 then
        for i = 1, dice_count do
            heal = heal + math.random(1, damage_dice)
        end
    end

    if lawful > 0 then
        heal = math.floor(heal * (1.0 + lawful / 32768.0))
    end

    heal = math.floor(heal * (leverage / 10.0))
    if heal < 0 then heal = 0 end
    return heal
end
