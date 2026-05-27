# 下次提交變更摘要

- 將伺服器版本修訂號提升至 `v0.3.31`，同步本次 yiwei 行為對齊推送。
- 對齊 yiwei 多個 buff/debuff 早退分支的 `sendGrfx()` 後置語義，補上 `CANCELLATION(44)`、`DARK_BLIND(103)`、`PHANTASM(212)`、`ARMOR_BREAK(112)`、`ELEMENTAL_FALL_DOWN(133)`、`EARTH_BIND(157)`、`WIND_SHACKLE(167)` 的目標特效與 PC 狀態刷新。
- 補齊 `POLLUTE_WATER(173)` / `STRIKER_GALE(174)` 對 NPC 目標的 debuff 套用、ShowID 廣播與 NPC/玩家治療倍率互動。
- 對齊 NPC mobskill 的 `trigger_random`、`isSkillUseble` gate 與可用技能隨機選擇流程。
- 補齊速度類互消後的 PC 狀態刷新，並讓 `SLOW(29)` / `MASS_SLOW(76)` / `ENTANGLE(152)` 跳過 `BraveSpeed=5` 目標。
- 對齊 yiwei 裝備型急速道具：YAML 標記 `20235 伊娃之盾` 為 `haste_item`，新增 `HasteItem` / `HasteItemEquipped`，裝備/卸裝/登入恢復/相消術/速度技能施放皆依 `getHasteItemEquipped()` 行為處理。
- 對齊 yiwei 卸下武器時解除武器依賴 buff：`COUNTER_BARRIER(91)` 與 `FIRE_BLESS(155)` 會在卸下最後一把武器時移除，並送出 `S_SkillBrave(type=0,duration=0)` 還原烈炎氣息速度。

- 補齊 yiwei PC→NPC 近戰打到持有 COUNTER_BARRIER(91) 的 NPC 時的反擊屏障路徑：命中後依 25 + npc.Level - player.Level + 33 判定，成功時反傷玩家、送 NPC 10710 與玩家 action=2，並取消原本對 NPC 的傷害。
- 補齊 yiwei NPC 持有 KIRTAS_BARRIER1(11060) 時的物攻反擊屏障：PC 近戰命中後無機率反傷玩家、送 NPC 10710 與玩家 action=2，並取消原本對 NPC 的物理傷害；91 COUNTER_BARRIER 原本機率與武器限制維持不變。
- 補齊 yiwei NPC type 2 mobskill 施放 KIRTAS_BARRIER1(11060) 時的自身屏障套用：NPC 會取得 11060 狀態、切換 Kirtas hidden/status20、廣播 NPC pack，並在 barrierTime > 15 或滿血顯形時清除 11060。
- 補齊 yiwei KIRTAS_BARRIER3(11058) / ABSOLUTE_BARRIER(78) 的 NPC 傷害歸零，以及 KIRTAS_BARRIER2(11059) 的 PC→NPC 魔法反射：11058 期間物攻與魔攻不傷 NPC，11059 會將 PC 魔法傷害反彈給施法者並取消原傷害。
