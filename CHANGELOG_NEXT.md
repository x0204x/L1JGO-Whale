# 下一版變更摘要
- 補齊 `CURE_POISON(9)` 早退路徑的 yiwei `sendGrfx()` 後置狀態刷新；解毒術成功後 now 會對 PC 目標送 `S_SPMR`、`S_OwnCharStatus` 與 `UPDATE_ER`，避免早退分支繞過補助/詛咒魔法通用刷新。
- 補齊 `CURSE_BLIND(20)` / `DARKNESS(40)` 早退路徑的 yiwei `sendGrfx()` 目標特效與後置狀態刷新；致盲類技能成功後會送各自 `cast_gfx`、`S_SPMR`、`S_OwnCharStatus` 與 `UPDATE_ER` 給 PC 目標。
- 補齊 `HOLY_LIGHT(37)` 早退路徑的 yiwei `sendGrfx()` 目標特效與後置狀態刷新；聖潔之光成功後會送 `S_SkillSound(227)`、`S_SPMR`、`S_OwnCharStatus` 與 `UPDATE_ER` 給 PC 目標。
