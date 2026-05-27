-- L1J stat lookup tables
-- Java: L1AttackList.java STRH/STRD/DEXH/DEXD/INTD/INTCRI, index 0-127.

STR_HIT = { _max_index = 127 }
for str = 0, 7 do
    STR_HIT[str] = 4
end
local str_hit = 4
for str = 8, 127 do
    if str % 3 == 0 or str % 3 == 2 then
        str_hit = str_hit + 1
    end
    STR_HIT[str] = str_hit
end

STR_DMG = { _max_index = 127 }
for str = 0, 9 do
    STR_DMG[str] = 2
end
local str_dmg = 2
for str = 10, 127 do
    if str % 2 == 0 then
        str_dmg = str_dmg + 1
    end
    STR_DMG[str] = str_dmg
end

STR_CRIT = { _max_index = 127 }
for str = 0, 39 do
    STR_CRIT[str] = 0
end
for str = 40, 127 do
    STR_CRIT[str] = math.floor((str - 30) / 10)
end

DEX_HIT = { _max_index = 127 }
for dex = 0, 7 do
    DEX_HIT[dex] = -3
end
local dex_hit = -3
for dex = 8, 127 do
    dex_hit = dex_hit + 1
    DEX_HIT[dex] = dex_hit
end

DEX_DMG = { _max_index = 127 }
for dex = 0, 8 do
    DEX_DMG[dex] = 2
end
local dex_dmg = 2
for dex = 9, 127 do
    if dex % 3 == 0 then
        dex_dmg = dex_dmg + 1
    end
    DEX_DMG[dex] = dex_dmg
end

DEX_CRIT = { _max_index = 127 }
for dex = 0, 39 do
    DEX_CRIT[dex] = 0
end
for dex = 40, 127 do
    DEX_CRIT[dex] = math.floor((dex - 30) / 10)
end

INT_DMG = { _max_index = 127 }
for intel = 0, 14 do
    INT_DMG[intel] = 0
end
local int_dmg = 0
for intel = 15, 127 do
    if intel % 5 == 0 then
        int_dmg = int_dmg + 1
    end
    INT_DMG[intel] = int_dmg
end

INT_CRIT = { _max_index = 127 }
for intel = 0, 34 do
    INT_CRIT[intel] = 0
end
local int_crit = 0
for intel = 35, 127 do
    int_crit = math.floor((intel - 30) / 5)
    INT_CRIT[intel] = int_crit
end

INT_MAGIC_HIT = { _max_index = 127 }
for intel = 0, 22 do
    INT_MAGIC_HIT[intel] = 0
end
local int_magic_hit = 0
for intel = 23, 127 do
    if intel % 3 == 2 then
        int_magic_hit = int_magic_hit + 1
    end
    INT_MAGIC_HIT[intel] = int_magic_hit
end

-- Experience table (cumulative exp for each level)
EXP_TABLE = {
    [1]  = 0,
    [2]  = 125,
    [3]  = 300,
    [4]  = 500,
    [5]  = 750,
    [6]  = 1296,
    [7]  = 2401,
    [8]  = 4096,
    [9]  = 6581,
    [10] = 10000,
    [11] = 14661,
    [12] = 20756,
    [13] = 28581,
    [14] = 38436,
    [15] = 50645,
    [16] = 65556,
    [17] = 83541,
    [18] = 104996,
    [19] = 130341,
    [20] = 160020,
    [21] = 194501,
    [22] = 234276,
    [23] = 279861,
    [24] = 331792,
    [25] = 390641,
    [26] = 456992,
    [27] = 531457,
    [28] = 614672,
    [29] = 707297,
    [30] = 810016,
    [31] = 923537,
    [32] = 1048592,
    [33] = 1185937,
    [34] = 1336352,
    [35] = 1500641,
    [36] = 1679632,
    [37] = 1874177,
    [38] = 2085152,
    [39] = 2313457,
    [40] = 2560016,
    [41] = 2825777,
    [42] = 3111713,
    [43] = 3418818,
    [44] = 3748113,
    [45] = 4100642,
    [46] = 4830002,
    [47] = 6338418,
    [48] = 9833681,
    [49] = 19745870,
    [50] = 55810962,
}

-- Helper: clamp value to table bounds.
function table_lookup(tbl, index)
    local max_index = tbl._max_index or #tbl
    if index < 0 then index = 0 end
    if index > max_index then index = max_index end
    local value = tbl[index]
    if value ~= nil then return value end
    return tbl[index + 1] or 0
end

function pure_stat_bonus(stat)
    if stat >= 25 and stat <= 44 then
        return math.floor((stat - 15) / 10)
    end
    if stat >= 45 then
        return 5
    end
    return 0
end

function apply_er_evasion(is_hit, dodge)
    if not is_hit then
        return false
    end
    local er = dodge or 0
    if er <= 0 then
        return true
    end
    return math.random(1, 3000) > er
end

-- Get level from cumulative exp
function level_from_exp(exp)
    for lv = 50, 2, -1 do
        if EXP_TABLE[lv] and exp >= EXP_TABLE[lv] then
            return lv
        end
    end
    return 1
end

-- Get cumulative exp required for a given level
function exp_for_level(level)
    if level <= 1 then return 0 end
    if level <= 50 then
        return EXP_TABLE[level] or 0
    end
    -- Beyond 50: linear extension
    return (EXP_TABLE[50] or 0) + (level - 50) * 10000000
end
