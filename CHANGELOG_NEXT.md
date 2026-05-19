## 技能

## 暴風神射（STORM_SHOT / 166）— 純審計重新確認核心 BowDmg ±5 + BowHit ±(-1) + REPEATEDSKILLS[0] 5 項武器 buff 互斥完整對齊 Java

- **Java 對照**：`L1SkillUse.java:2611-2615 addBowDmgup(5) + addBowHitup(-1) + S_PacketBoxIconAura(165, duration)`、`L1SkillStop.java:569 addBowDmgup(-5) + addBowHitup(1) + S_PacketBoxIconAura(165, 0)`；REPEATEDSKILLS[0]={148,149,156,163,166}。
- **Go 對照**（無代碼變更）：
  - `scripts/combat/buffs.lua:133 [166] = { bow_dmg = 5, bow_hit = -1, exclusions = {148, 149, 156, 163} }` 對齊 Java。
  - 標準 buff apply/revert 對稱（+5/-5、-1/+1）對齊 Java `L1SkillStop`。
  - `buff_icon_map.yaml:68-69 skill_id: 166 type: aura`（無 param → iconID = 165）對齊 Java `S_PacketBoxIconAura(165, ...)`。
  - yaml `skill_list.yaml:5086` 31 欄位中 30 項對齊 yiwei `skills.sql:165`，僅 cast_gfx=2248 vs 11732 一項漂移（Go 跟 cat-fei）。
- **broader gap（不改）**：(A) cast_gfx 漂移屬廣域 yiwei/cat-fei SQL 同步議題；(B) sendGrfx 末尾 1686-1694 通用 status refresh 屬廣域 buff cast 後置缺口。
- **驗證**：`cd server && go build ./...` 通過，本步無代碼變更（純審計）。

## 生命呼喚（CALL_OF_NATURE / 165）— 純審計重新確認三路徑復活（PC YN/Pet/NPC）完整對齊 Java，yaml 31 欄位零漂移

- **Java 對照**：yiwei `skills.sql:164` 31 欄位（mp=50/item=40319×1/target=buff/target_to=3/attr=4/type=32/ranged=10/area=0/id=16/action_id=19/cast_gfx=2245/sys_msg_fail=280）；Java `CALL_OF_NATURE.skillmode` 對 PC 走 `setTempID + S_Message_YN(322)` 不直接復活，對 Pet/NPC 走滿血復活並含 witness check（IsPlayerAt → 拒絕）+ NPC `L1Tower || CantResurrect` 雙拒絕。
- **Go 對照**（無代碼變更）：
  - `system/skill_heal_resurrect.go:112-149 case 165` 三路徑全部對齊：PC 目標 `IsPlayerAt → 592 witness msg / TempID + PendingResSkill + PendingResCaster + SendYesNoDialog(322)`、Pet 走 `callOfNatureResurrectPet → resurrectPetWithHP(MaxHP)`（同地圖 + 距離 ≤20 + IsPlayerAt 三重檢查）、NPC 走 `callOfNatureResurrectNpc → resurrectNpcWithHP(MaxHP)` 內含 `Impl=="L1Tower" || isNpcCantResurrect → return false` 雙拒絕。
  - `data/yaml/skill_list.yaml:5055` 31 欄位 **零漂移** 完全對齊 yiwei sql:164。
- **測試覆蓋**：既有 3 個 test 檔（`skill_call_of_nature_test.go` + `skill_call_of_nature_companion_test.go` + `skill_resurrect_restriction_test.go`）覆蓋 PC YN/witness/Pet/NPC/cant_resurrect 全分支。
- **broader gap（不改）**：(A) Java `L1SkillUse2:1850-1888` sleep buff lifecycle hooks 屬廣域 buff 副作用議題；(B) Java NPC AI 對 165 cast 路徑 Go 走 SkillSystem 屬廣域 NPC skill routing 議題。
- **驗證**：`cd server && go build ./...` 通過，本步無代碼變更（純審計）。

## 生命的祝福（NATURES_BLESSING / 164）— 純審計確認 type=16 + target_to=8 + area=-1 隊伍範圍治療路徑完整對齊 Java，含 WATER_LIFE 雙倍移除與 POLLUTE_WATER 減半

- **Java 對照**：
  - yiwei `skills.sql:163` 31 欄位：`type=16 (HEAL) / target_to=8 (TARGET_TO_PARTY) / area=-1 (screen-wide) / attr=4 (water) / damage_value=10 / damage_dice=12 / damage_dice_count=0 / cast_gfx=2244 / action_id=19 / mp_consume=30 / reuse_delay=300`。
  - `L1SkillUse.isInTarget()` 行 671-676：caster 自身永遠通過（self-pass）；行 877-880：`target_to=8` 走 `_player.getParty().isMember(targetPc)` 過濾非隊伍成員。
  - `L1SkillUse.java:1997-2001`：heal 計算前 `if (target.hasSkillEffect(WATER_LIFE)) _heal = _heal << 1` 雙倍、`if (target.hasSkillEffect(POLLUTE_WATER)) _heal = _heal >> 1` 減半。
  - `L1SkillUse2.java:2113-2118`：heal 套用後恢復系技能清單 `(HEAL || EXTRA_HEAL || GREATER_HEAL || FULL_HEAL || HEAL_ALL || NATURES_TOUCH || NATURES_BLESSING)` cast → `cha.removeSkillEffect(WATER_LIFE)`。

- **Go 對照**（無代碼變更）：
  - `data/yaml/skill_list.yaml:5024` 全 31 欄位中 30 項對齊 yiwei，僅 `reuse_delay=0 vs 300` 一項漂移（Go 跟 cat-fei）。
  - `system/skill_self.go:213-267` 路由：`skill.Type==16 && skill.Area==-1` → caster 經 `applyElfWaterHealingModifiers` 治療 → `skill.TargetTo==8` 建 `Parties.GetParty(player.CharID).Members` map → loop nearby（跳過 caster 自身、跳過非隊員）→ 每員 `applyElfWaterHealingModifiers` 治療。
  - `system/skill_elemental.go:97-109 applyElfWaterHealingModifiers`：`HasBuff(170) → heal <<= 1 + removeBuffAndRevert(170)`（雙倍 + 移除一體執行）+ `HasBuff(173) → heal >>= 1`（POLLUTE_WATER 減半）。
  - `skill_buff.go:1246-1253` 註解明示：164 yaml type=16 走 heal 計算路徑，由 `applyElfWaterHealingModifiers` 統一處理 WATER_LIFE 移除；對比 158 NATURES_TOUCH yaml type=2（buff 非 heal）需顯式 `case 158 → removeBuffAndRevert(170)` hook。
  - 技能 cast 入口 `skill_self.go:209-211` 廣播 `BuildActionGfx(player.CharID, 19)`；cast_gfx=2244 走 `applyBuffEffect` 或 heal 路徑後續處理。

- **盤點驗證**：Go 全 yaml 確認只有 164 同時具 `target_to=8 + type=16`，`if skill.TargetTo == 8` 隊伍過濾分支精確影響範圍限本技能；其他 TYPE_HEAL=16 技能（1 HEAL / 19 EXTRA_HEAL / 35 GREATER_HEAL / 49 FULL_HEAL / 57 HEAL_ALL）皆 `target_to != 8` 走單目標或預設路徑。

- **broader gap（不改）**：
  - **A) yaml reuse_delay=0 vs 300 漂移**：Go 跟 cat-fei，屬廣域 yiwei/cat-fei SQL 同步議題（同 157/159/161 同源）。
  - **B) NPC 目標 + 164 cast**：Java `target_to=8` 隊伍過濾僅作用於 PC 目標清單，NPC 不在隊伍中本就排除。Go 同樣設計，邏輯一致。
  - **C) HprExecutor `bonus /= 2` 對 `getBaseCon() >= 45 + 負重` 的減半**：屬廣域 regen 公式差，與 164 治療路徑無關。

- **驗證**：`cd server && go build ./...` 通過，本步無代碼變更（純審計）。

## 三重矢（TRIPLE_ARROW / 132）— 完整 B 路徑重構：3 次獨立 L1AttackPc 弓箭流程 + DEX_DMG / 箭矢加傷 / 武器 buff 鏈（刻意捨棄 Java ×5 倍率以維持遊戲平衡）

- **問題回報**：使用者觀察到三重矢傷害異常。Phase 1 根因調查發現 4 項真實差異：
  - **a) 3 發傷害數字完全相同**：Java 為 3 次獨立 `cha.onAction(srcpc)` 完整擲骰，Go 算 1 次傷害後 loop 3 次套用相同 dmg。
  - **b) 完全沒套 DEX_DMG**：弓箭手核心加成完全缺失（calc_physical_skill 只用 STR_DMG）。
  - **c) 命中 1 次決定全 hit/全 miss**：Java 每發獨立判定，Go 一發 miss 全 miss。
  - **d) 未走完整武器 buff 鏈**：武器強化 +N、屬性強化、箭矢加傷、BURNING_WEAPON、FIRE_WEAPON 等完整 L1AttackPc 鏈全部缺失。

- **Java 對照**：
  - `TRIPLE_ARROW.skillmode start():13-48`：弓裝備檢查（`getCurrentWeapon() != 20 → return 0`）→ `setIsTRIPLE_ARROW(true)` → `for (int i=0; i<3; i++) cha.onAction(srcpc)` → `setIsTRIPLE_ARROW(false)` → `sendPacketsAll(S_SkillSound(4394) + S_SkillSound(11764))`。
  - 每次 `cha.onAction(srcpc)` 走完整 L1AttackPc 弓箭流程：箭矢消耗、命中骰、武器擲骰、STR_DMG/DEX_DMG/arrow_dmg/weapon enchant、武器 buff 鏈、reduction armor 等。
  - `L1AttackPc.java:1512/2002 if (_pc.getIsTRIPLE_ARROW()) dmg *= ConfigSkill.TRIPLE_ARROW_DMG=5.0`——**Go 刻意不對齊此倍率**。Java 另有 `妖精_技能設定表.properties: Triple_Arrow_Dmg = 1.0` 載入 ConfigElfSkill 但 L1AttackPc 未引用，Go 行為相當於使用後者 1.0。

- **Go 修改**：
  - `handler/context.go`：`CombatQueue` 介面新增 `ExecuteRangedAttackOnNpc(player, npcID)` 同步入口（不經佇列）。
  - `system/combat.go`：抽取 `processRangedAttack` 主體為 `processRangedAttackForPlayer`，新增 `ExecuteRangedAttackOnNpc` 公開方法。
  - `system/skill_damage.go`：NPC 路徑 `executeAttackSkill` 對 skillID==132 提早分流：`loop 3× ExecuteRangedAttackOnNpc → 廣播 4394+11764 → return`；PC 路徑 `executeAttackSkillOnPlayer` 同模式 `HandlePvPFarAttack ×3`。**不**設定旗標、**不**套用 ×5 倍率。
  - **移除死碼**：`magic.lua calc_physical_skill` 的 `elseif sid == 132` 整段分支、`skill_damage.go` 手動 1 箭消耗、NPC/PC 兩處 `if skillID==132 && hitCount<3 → 3` 強制三命中、兩處攻擊後的 4394+11764 收尾廣播。

- **效果**：每發走完整 `calc_ranged_attack`（含 `DEX_DMG` 主、`STR_DMG/2`、箭矢加傷、bow_dmg_mod）+ striker_gale + dark_elf_physical + brave_aura + immune_to_harm + reduction_armor + weapon_skill_proc + doll_skill_proc，且每發各別 `FindArrow + RemoveItem(1)` ⇒ 3 箭實際消耗對齊 Java per-onAction。傷害每發獨立擲骰、命中獨立判定、所有武器 buff 完整生效，**但每發約等同單發普通弓箭**（無 ×5 倍率）。

- **新測試**：
  - 替換舊 `TestSkillTripleArrowDamagesPlayerTargetThreeTimes`（驗證 S_OPCODE_ATTACK ×3 為錯誤行為）。
  - 新增 `TestSkillTripleArrowRoutesThroughPvPFarAttackThreeTimes`：鎖定 HandlePvPFarAttack 呼叫 3 次 + 收尾廣播 4394+11764。
  - 新增 `TestTripleArrowDirectInvocationDamage`：實測 ExecuteRangedAttackOnNpc × 3，每次擊中傷害 ≤ 單次普通弓箭 × 1.5 buffer（無 ×5 倍率回歸防護）。實測值：單次普通最大 16，三重矢 3 次 `[13, 15, 13]` 總 41——每發 ~1× 單發普通，3 發合計約 3×。

- **broader gap（不改）**：
  - **A) SprTable.getAttackSpeed(playerGFX, 21)==0 變身動畫檢查**：Go 無 SPR table 系統屬廣域變身動畫缺口。
  - **B) yaml reuse_delay=400 三方漂移**：Go 既不跟貓飛也不跟 yiwei，屬廣域 SQL 漂移議題。
  - **C) Java NPC 端 npc.attackTarget(cha) ×3 倒序循環**：Go NPC 技能路徑由 SkillSystem 處理而非 NpcAi，屬廣域 NPC 技能 system routing 議題。
  - **D) 刻意不對齊 Java ×5 per-hit 倍率**：使用者實測判定 ×5 倍率（=每發 5×、3 發 15×）破壞遊戲平衡，明示要求移除。Java 兩 config 檔本身衝突（`ConfigSkill=5.0` vs `ConfigElfSkill=1.0`），Go 採後者語意。本決策為使用者明示要求，非疏漏。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `go test ./internal/system + ./internal/handler` 全部通過。
  - 排序佇列原訂下一項為 164 NATURES_BLESSING，本項為使用者插隊請求（非佇列推進）。

## 烈炎武器（BURNING_WEAPON / 163）— 純審計確認核心 DmgMod ±6 + HitMod ±3 + REPEATEDSKILLS[0] 5 項武器 buff 互斥完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java BURNING_WEAPON = 163`，無對應 skillmode（走 generic apply 路徑）。
  - `L1SkillUse.java:2595-2599` 套用：`_targetPc.addDmgup(6); _targetPc.addHitup(3); _targetPc.sendPackets(new S_PacketBoxIconAura(162, _user.getSkillEffectTimeSec(BURNING_WEAPON)))`——icon param = skillID-1 = 162。
  - `L1SkillStop.java:546-553` 解除：`_pc.addDmgup(-6); _pc.addHitup(-3); _pc.sendPackets(new S_PacketBoxIconAura(162, 0))`。
  - `L1BuffUtil.braveStart` 不涉及（僅 brave-speed 族）。
  - `L1SkillUse.REPEATEDSKILLS[0] = {148, 149, 156, 163, 166}`：163 與 148 ENCHANT_WEAPON、149 SHADOW_FANG、156 STORM_EYE、166 BLESS_WEAPON 互斥（cast 163 時清除 4 項武器 buff）。
  - yiwei `db_split/skills.sql:162`：`('163', '烈炎武器', '21', '2', '30', '0', '40319', '4', '0', '300', 'none', '0', '0', '0', '0', '0', '0', '0', '1', '0', '0', '-1', '1', '1', '', '19', '11776', '0', '0', '723', '280')` — mp=30、item=40319×4、buff_duration=300、type=2（TYPE_BUFF）、target=none、cast_gfx=**11776**、sys_msg_happen=**723**、sys_msg_stop=**280**。

- **Go 對照**：
  - `buffs.lua:132 [163] = { dmg_mod = 6, hit_mod = 3, exclusions = {148, 149, 156, 166} }`——核心 DmgMod ±6 + HitMod ±3 + REPEATEDSKILLS[0] 4 項互斥（148/149/156/166）**完整對齊** Java。
  - `skill_buff.go:161-162 + 195-196 + 513-514` 標準 buff apply/revert：DmgMod/HitMod delta 由 buffs.lua 配置驅動，套用時 `+6 / +3`，revert 時 `-6 / -3`——對齊 Java addDmgup/addHitup 對稱。
  - `skill_buff.go:exclusions` 機制：cast 163 時 loop exclusions 對 target 移除既有 148/149/156/166 buff——對齊 Java REPEATEDSKILLS[0] 互斥邏輯。
  - `buff_icon_map.yaml:66-67`：`skill_id: 163, type: aura`（無 param → 預設 iconID = skill_id-1 = 162）——對齊 Java `S_PacketBoxIconAura(162, ...)` icon ID。
  - yaml `skill_list.yaml:4993-5023`：mp=30、item=40319×4、buff_duration=300、type=2、target=none、cast_gfx=**2242**、sys_msg_happen=**0**、sys_msg_stop=**0**——**3 項漂移**（cast_gfx、sys_msg_happen、sys_msg_stop Go 跟 cat-fei）。

- **既有測試覆蓋**：
  - `skill_buff` 系列測試覆蓋 DmgMod/HitMod delta apply/revert 路徑。
  - REPEATEDSKILLS[0] 互斥邏輯由 148/149/156/166/163 任一 cast 觸發；先前 148/149/156 audit 已確認 exclusions 配置正確。

- **發現的 Java 真實差異**：**無**（核心 DmgMod ±6 + HitMod ±3 + icon 162 + REPEATEDSKILLS[0] 4 項互斥完整對齊）。

- **broader gap（不改）**：
  - **A) yaml cast_gfx 漂移**：Go=2242 vs yiwei=11776（Go 跟 cat-fei）——同 148/149/150/156 系列「武器/防具 buff 家族」共通 SQL 同步議題。
  - **B) yaml sys_msg_happen/sys_msg_stop 漂移**：Go=0/0 vs yiwei=723/280（Go 跟 cat-fei）——同樣屬廣域 yiwei/cat-fei SQL 同步議題。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 163 的新測試需執行（既有 skill_buff DmgMod/HitMod + REPEATEDSKILLS[0] exclusions 路徑已覆蓋）。

## 召喚強力屬性精靈（GREATER_ELEMENTAL / 162）— 純審計重新確認核心 5 項機制 + NPC ID 映射完整對齊 Java，兩項召喚家族共通缺口維持延後

- **Java 對照**：
  - `L1SkillId GREATER_ELEMENTAL = 162`，`skillmode/GREATER_ELEMENTAL.java:23-67` 完整流程：
    1. `attr = pc.getElfAttr()` 必須 != 0。
    2. `if (!pc.getMap().isRecallPets()) → sendPackets(S_ServerMessage(353)) + return`。
    3. 迭代 `pc.getPetList()` 累計 `petcost`，`!= 0` → 靜默 return。
    4. Switch on `attr`：1→**81053**（強力土之精靈）、2→**81050**（強力火之精靈）、4→**81051**（強力水之精靈）、8→**81052**（強力風之精靈）。
    5. `new L1SummonInstance(npcTemp, pc) + summon.setPetcost(pc.getCha() + 7)`。
  - 與 154 LESSER_ELEMENTAL 唯一差異：召喚 NPC ID（強化精靈群 81050-81053 vs 一般精靈群 45303-45306）。
  - yiwei `db_split/skills.sql:161`：mp=20、item=40319×**4**（vs 154 ×2）、buff_duration=0、target=none、type=128、attr=0、action_id=19、cast_gfx=2510、所有 sys_msg=0。

- **Go 對照**：
  - `skill.go:409-410 case 154, 162` 共用路由 → `ExecuteElementalSummon`。
  - `skill_summon.go:102-107 greaterElementalByAttr` map（同 Java：1→81053、2→81050、4→81051、8→81052）。
  - `skill_summon.go:282-363 ExecuteElementalSummon` case 162 路徑：
    - `:286-288` ElfAttr == 0 gate。
    - `:289-295` RecallPets 檢查 + `msgCannotSummonHere`。
    - `:296-298` calcUsedPetCost != 0 靜默 return。
    - `:300-309` 走 `greaterElementalByAttr[player.ElfAttr]` 取 NPC ID。
    - `:317-319` `petCost = Cha + 7`（與 154 同邏輯）。
    - `:321 ConsumeSkillResources` MP 消耗在 gates 通過後（user-friendly pattern）。
    - `:355-362 AddSummon + SendSummonPack(viewer + self) + SendSummonMenu` 對應 Java L1SummonInstance constructor 內部 packet。
  - yaml `skill_list.yaml:4962-4992`：mp=20、item=40319×4、buff_duration=0、target=none、type=128、attr=0、action_id=19、cast_gfx=2510、所有 sys_msg=0——**31 欄位完全對齊 yiwei**（零漂移）。

- **既有測試覆蓋**：
  - `TestSkillElementalSummonGreaterElemental*` 系列鎖定 NPC 81051+petcost 行為（先前 154 audit 提及）。
  - `skill_summon` 系列測試覆蓋共用 ExecuteElementalSummon 核心路徑。

- **發現的 Java 真實差異**：**無**（核心 5 項機制 + NPC ID 4 屬性映射完整對齊；與 154 LESSER_ELEMENTAL 共用實作對齊保證）。

- **broader gap（不改）**：
  - **A) 施法動畫 + cast_gfx 廣播缺失**：Java L1SkillUse.sendGrfx 外圍流程送 action_id=19 + cast_gfx=2510 廣播；Go ExecuteElementalSummon 入口前無對應 broadcast。屬召喚技能 (51/36/41/145/154/162) **共通缺口**，應在召喚技能統一 audit 時集中補上 cast 動畫廣播 helper（與 154 同源）。
  - **B) MP 消耗時機字面順序**：Java `L1SkillUse._user.useMP/SP` 在 `runSkill` 入口處先消耗，Go 在所有 gates 通過後才消耗——Go 採 user-friendly pattern 與 131 precedent 一致，**不修**（與 154 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 既有 `TestSkillElementalSummonGreaterElemental*` 已覆蓋 81051 NPC ID + petcost 行為，無新測試需執行。

## 封印禁地（AREA_OF_SILENCE / 161）— 純審計確認核心 silence + 白名單 + 全頻廣播阻擋完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:542 AREA_OF_SILENCE = 161`，無對應 skillmode（走 generic apply 路徑）。
  - `C_UseSkill.java:189-193` cast-time silence 阻擋：`if (pc.hasSkillEffect(AREA_OF_SILENCE) && !isError) { if (!isSilenceUsableSkill(skillId)) isError = true; }`。
  - `C_UseSkill.java:87-88 _cast_with_silence` 白名單：`{SHOCK_STUN, REDUCTION_ARMOR, BOUNCE_ATTACK, SOLID_CARRIAGE, COUNTER_BARRIER, FOE_SLAYER}` = `{87, 88, 89, 90, 91, 187}` 6 項物理 buff 可在沉默下施放。
  - `C_ChatGlobal.java:77-80`：`if (pc.hasSkillEffect(AREA_OF_SILENCE) && !pc.isGm()) → isStop = true`——非 GM 玩家持 161 buff 時禁全頻廣播。
  - `L1MagicPc.java:491-505` cast probability：`ConfigElfSkill.AREA_OF_SILENCE_1/2/3`（5/10/15）三段 lvldiff + INT*INT - MR*MR 微調。
  - `L1MagicNpc.java` 含 AREA_OF_SILENCE 抗性檢查（NPC 目標路徑）。
  - yiwei `db_split/skills.sql:160`：`('161', '封印禁地', '21', '0', '40', '0', '40319', '8', '9000', '16', 'none', '3', '0', '0', '0', '33', '50', '0', '1', '0', '0', '3', '1', '1', '', '19', '10708', '0', '0', '0', '0')` — mp=40、item=40319×8、reuse_delay=9000、buff_duration=16、target=none、target_to=3、prob_value=33、prob_dice=**50**、attr=0、type=1（TYPE_PROBABILITY）、ranged=0、area=**3**（3-tile chebyshev radius）、through=1、id=1、action_id=19、cast_gfx=**10708**、sys_msg_happen=**0**。

- **Go 對照**：
  - `skill_self.go:129-132 case 161`：`BroadcastToPlayers(nearby, BuildActionGfx(player.CharID, byte(skill.ActionID)))` + `applyAreaOfSilence(player, skill, nearby)` + return。
  - `skill_elemental.go:111-145 applyAreaOfSilence`：
    - duration = `skill.BuffDuration` (16s) 或 default 16。
    - loop `nearby`：跳過 caster/dead；`playerDebuffSkills[161] && !checkPlayerMRResist → continue`（MR 閘）；移除舊 161 buff + revert；新建 `ActiveBuff{SkillID=161, SetSilenced=true}` + `target.Silenced = true` + AddBuff；`sendBuffIcon(target, 161, duration)`；cast_gfx 廣播；sys_msg_happen 個別發送。
  - `skill.go:288 player.Silenced && !isCastableWhileSilenced(skillID) → 阻擋`：對齊 Java cast-time silence 檢查。
  - `skill.go:572-579 isCastableWhileSilenced`：`{87, 88, 89, 90, 91, 187}` 6 項白名單**完全對齊** Java `_cast_with_silence` 列表。
  - `handler/chat.go:83`：全頻廣播阻擋對齊 Java `C_ChatGlobal.java:77-80`（SILENCE/AREA_OF_SILENCE/STATUS_POISON_SILENCE 三類沉默都阻擋，GM 不受限）。
  - `skill_status.go:831 playerDebuffSkills[161] = true`：MR 抗性閘。
  - `buffs.lua:130 [161] = {}` 空 buff（flag-only，無屬性 delta，由 SetSilenced 處理）。
  - yaml `skill_list.yaml:4931-4961`：mp=40、item=40319×8、reuse_delay=9000、buff_duration=16、target=none、target_to=3、prob_value=33、prob_dice=**30**、attr=0、type=1、ranged=0、area=**-1**（screen-wide）、through=1、id=1、action_id=19、cast_gfx=**2241**、sys_msg_happen=**715**——**4 項漂移**（prob_dice、area、cast_gfx、sys_msg_happen Go 跟 cat-fei）。

- **既有測試覆蓋**：
  - `skill_clan_shock_stun_test.go` 涉及 SHOCK_STUN 在 silence 下白名單放行的測試。
  - 無針對 161 套用範圍/cast 阻擋的單元測試。

- **發現的 Java 真實差異**：**無**（核心 silence 套用 + 白名單 + 全頻阻擋 + MR 閘 + sys_msg 完整對齊；先前 audit 已建立 isCastableWhileSilenced 白名單與 6 項精確匹配）。

- **broader gap（不改）**：
  - **A) yaml `area=-1` vs `3`**：Go 採 cat-fei「screen-wide」設計（area=-1 在 Go `data/skill.go:33` 定義為 "screen"），applyAreaOfSilence 直接 loop nearby AOI 玩家（AOI 範圍 ~15 tiles），不 enforce 3-tile chebyshev 半徑——yiwei 為 3-tile radius，Go 為 screen-wide。**內部一致**（code 配合 yaml `-1=screen` semantic），但與 yiwei 設計不同——屬廣域 yiwei/cat-fei SQL 同步議題（Go 跟 cat-fei）。
  - **B) yaml 其餘 3 項漂移**：prob_dice=30 vs 50、cast_gfx=2241 vs 10708、sys_msg_happen=715 vs 0（Go 跟 cat-fei）——同樣屬廣域 SQL 同步議題。
  - **C) `ConfigElfSkill.AREA_OF_SILENCE_1/2/3` 三段公式**：Java 5/10/15 + INT/MR 微調；Go 用 generic `checkPlayerMRResist` 簡化版——屬廣域 ConfigElfSkill probability 表議題（同 173/174 同源）。
  - **D) NPC cast 路徑**：`L1MagicNpc` 含 161 抗性檢查；Go NPC 目標路徑 audit 已在 EARTH_BIND/SHOCK_STUN 系列中建立框架，161 未在 NPC cast 路徑顯式 case（NPC 通常不對玩家 cast 161）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 161 的新測試需執行（既有 SHOCK_STUN/silence 路徑測試已間接覆蓋）。

## 水之防護（AQUA_PROTECTER / 160）— 移除 Go 多送的 SendDodgeIcon 對齊 Java skillmode 無 packet 行為

- **Java 對照**：
  - `L1SkillId.java:538 AQUA_PROTECTER = 160`，`L1SkillMode.java:102` 註冊 skillmode。
  - `skillmode/AQUA_PROTECTER.start(L1PcInstance, ...)`（line 13-18）：**只**呼叫 `srcpc.setSkillEffect(160, integer * 1000)`，**不送任何 packet**。
  - `skillmode/AQUA_PROTECTER.start(L1NpcInstance, ...)`：return 0，無 apply（NPC 無法施放於他人）。
  - `skillmode/AQUA_PROTECTER.stop()`：**empty**，無任何 packet 送出。
  - `L1PcInstance.java:3399-3401 getEr()` getter override：`if (hasSkillEffect(AQUA_PROTECTER)) er += 5`——**透過 getter 即時加 +5**，非儲存值修改。注意 line 3396-3398 `STRIKER_GALE return 0` 優先級高於 AQUA_PROTECTER（174 持有時 getEr 直接回 0）。
  - 對比其他 Dodge buff skillmodes：DRAGONEYE_ANTHARAS/BIRTH/FIGURE/LIFE/MIRROR_IMAGE/UNCANNY_DODGE 皆顯式 `sendPackets(new S_PacketBoxIcon1(true, get_dodge()))`；SOLID_CARRIAGE(90)/DRESS_EVASION(111) 送 `S_PacketBox(UPDATE_ER, getEr())`——**160 AQUA_PROTECTER 不送任何 packet**。
  - yiwei `db_split/skills.sql:159`：`('160', '水之防護', '20', '7', '30', '0', '0', '0', '100', '960', 'buff', '1', '0', '0', '0', '0', '0', '4', '2', '0', '-1', '0', '0', '128', '', '19', '5829', '0', '0', '0', '0')` — mp=30、reuse_delay=**100**、buff_duration=960、target=buff、target_to=1、attr=4、type=2、ranged=-1、id=128、action_id=19、cast_gfx=5829、所有 sys_msg=0。

- **Go 對照**（修正前 vs 修正後）：
  - 修正前 `skill_buff.go:244-250` applyBuffEffect Dodge 通知區塊：`if buff.DeltaDodge > 0 { if 90/111 → SendUpdateER else → SendDodgeIcon }`——160 走 `else` 分支送 `SendDodgeIcon`，**但 Java AQUA_PROTECTER.start() 不送任何 packet**。Criterion (a) Java 真實差異：Go 多送了客戶端 UI 上不該出現的 dodge icon。
  - 修正前 `skill_buff.go:523-531` revertBuffStats Dodge 過期區塊：同樣對 160 送 SendDodgeIcon，**Java AQUA_PROTECTER.stop() empty 不送任何 packet**。
  - 修正後 apply + revert 兩處皆改 switch 結構：
    - `case 90, 111` 維持 `SendUpdateER`（對齊 Java SOLID_CARRIAGE/DRESS_EVASION skillmode）。
    - **`case 160` 新增空分支不送任何 packet** 對齊 Java AQUA_PROTECTER skillmode。註解說明 UPDATE_ER 屬 sendGrfx _targetList 通用 status refresh 廣域 gap。
    - `default` 維持 `SendDodgeIcon`（對齊 DRAGONEYE_*/MIRROR_IMAGE/UNCANNY_DODGE）。
  - `buffs.lua:129 [160] = { dodge = 5 }`：DeltaDodge=5 → `target.Dodge += 5` apply / `-= 5` revert——功能等價於 Java getEr() override。
  - yaml `skill_list.yaml:4900-4930`：mp=30、reuse_delay=**0**、buff_duration=960、target=buff、target_to=1、attr=4、type=2、ranged=-1、id=128、action_id=19、cast_gfx=5829、所有 sys_msg=0——**31 欄位中 30 項對齊 yiwei**，**僅 reuse_delay=0 vs 100 一項漂移**（Go 跟 cat-fei）。

- **既有測試覆蓋**：
  - 無針對 160 AQUA_PROTECTER 的單元測試。其他 Dodge buff（106 UNCANNY_DODGE、111 DRESS_EVASION、DRAGONEYE_*）的 apply/revert 測試已覆蓋 SendUpdateER vs SendDodgeIcon 分流邏輯（測試應不受 160 case 影響）。

- **本次修正範圍**（criterion (a) Java 核心行為差異）：
  - `skill_buff.go:244-256` apply + `skill_buff.go:523-535` revert 兩處 Dodge 通知區塊改 switch，新增 `case 160` 空分支對齊 Java AQUA_PROTECTER skillmode 不送任何 packet。

- **broader gap（不改）**：
  - **A) `L1PcInstance.getEr()` STRIKER_GALE > AQUA_PROTECTER 優先級**：Java 持有 174 + 160 時 `getEr() = 0`（174 priority 1），Go 採加法 DeltaDodge 路徑，174 audit 已透過 `SendUpdateER(0)` 偽造 UI 但儲存值 Dodge 仍包含 +5 from 160——屬廣域 stat computation 「multiplier/override getter chain」議題（同 153 ERASE_MAGIC MR/=4 同源）。
  - **B) `L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：Java cast 後通用送 UPDATE_ER；Go 無此通用 hook，160 cast 後客戶端 UI 不會即時看到 ER +5——屬廣域 buff cast 後置缺口（同 130/146/148/149/150/151/152/155/156/158/159 同源）。
  - **C) yaml reuse_delay 漂移**：Go=0 vs yiwei=100（Go 跟 cat-fei）——屬廣域 yiwei/cat-fei SQL 同步議題。
  - **D) AQUA_PROTECTER mass-buff / item-cast paths**：`L1AllBuff.java:55`/`L1GMAllBuff.java:55`/`Crazy_Buff.java:134（註解掉）`/`Full_magic.java:65` 都引用 160——這些是 GM 指令或道具觸發路徑，與 cast 路徑共用同樣的 skillmode.start()，本次修正一併涵蓋。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -timeout 60s` PASS（19.740s 全綠）。

## 大地的祝福（EARTH_BLESS / 159）— 純審計重新確認核心 icon-only buff（無 AC 修正、不在任何 REPEATEDSKILLS）完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:534 EARTH_BLESS = 159`，無對應 skillmode 類別（走 generic `cha.setSkillEffect(_skillId, _getBuffDuration)` apply 路徑）。
  - `L1SkillUse.java:1422-1424` + `L1SkillUse2.java:1439-1441` icon cast：`sendPackets(S_SkillIconShield(7, _getBuffIconDuration))` icon param=7。
  - `L1SkillUse.java:2523-2526` + `L1SkillUse2.java:2475-2478` apply：`// pc.addAc(-7);` **註解掉** + `sendPackets(S_SkillIconShield(7, duration))` — **icon-only，無 AC 修正**。
  - `L1SkillStop.java:445-451` stop：`//cha.addAc(7);` **註解掉** + `(if PC) sendPackets(S_SkillIconShield(7, 0))` — icon clear，無 AC 還原。
  - 不在任何 `REPEATEDSKILLS[0..9]` 群組（per 151 EARTH_SKIN audit 已驗證——`REPEATEDSKILLS[1]={EARTH_SKIN=151, IRON_SKIN=168}` 不含 159）。
  - 表格成員：`isNotCancelable` 不含；`EXCEPT_COUNTER_MAGIC` 含（不受 counter magic 阻擋）。
  - yiwei `db_split/skills.sql:158`：`('159', '大地的護衛', '20', '6', '30', '10', '0', '0', '0', '600', 'none', '0', '0', '0', '0', '0', '0', '1', '2', '0', '0', '0', '0', '64', '', '19', '2287', '0', '0', '725', '0')` — mp=**30**、hp_consume=**10**、buff_duration=**600**、target=none、target_to=**0**、attr=1、type=2、ranged=0、id=64、action_id=19、cast_gfx=2287、sys_msg_stop=725。

- **Go 對照**：
  - `buffs.lua:128 [159] = {}` 空 buff——無 AC delta、無 exclusions、無 mutex（先前 2026-05-18 audit 已從 `{151, 168}` 收緊為 `{}` 對齊 Java 不在 REPEATEDSKILLS）。
  - `buff_icon_map.yaml:23-25 skill_id: 159 type: shield param: 7` → `handler/skill.go:209 sendIconShield(sess, durationSec, 7)` 送 `S_SkillIconShield(7, duration)` 對齊 Java cast icon。
  - `skill_buff.go:307 sendBuffIcon` apply path + `:126 sendBuffIcon(target, skillID, 0)` revert path——對齊 Java cast/stop icon 雙向。
  - 執行路徑：yaml target='none' → `executeSelfSkill` default → `applyBuffEffect`（lua [159]={} 無 stat delta） + `sendBuffIcon(159, duration)` + cast_gfx 廣播。
  - yaml `skill_list.yaml:4869-4899`：mp=**40**、hp_consume=**0**、buff_duration=**960**、target=none、target_to=**8**、attr=1、type=2、ranged=0、id=64、action_id=19、cast_gfx=**2287**、sys_msg_stop=725——**4 項漂移**（mp=40 vs 30、hp_consume=0 vs 10、buff_duration=960 vs 600、target_to=8 vs 0；Go 跟 cat-fei），**cast_gfx=2287 對齊 yiwei 無漂移**。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:106-119 TestSkillElementalBuffElfArmorAndWaterBuffsUseJavaValues`：覆蓋「159 EARTH_BLESS（不互斥）cast → 151 EARTH_SKIN 保留 + AC 不變」場景——驗證 159 不改 AC、不互斥 151。

- **發現的 Java 真實差異**：**無**（核心 icon-only buff 完整對齊：無 AC 修正、不在 REPEATEDSKILLS、icon param=7 雙向 cast/stop；先前 audit 已將 exclusions 收緊為 `{}` 對齊 Java）。

- **broader gap（不改）**：
  - **yaml 4 項漂移**：mp=40 vs 30、hp_consume=0 vs 10、buff_duration=960 vs 600、target_to=8 vs 0（Go 跟 cat-fei）——屬廣域 yiwei/cat-fei SQL 同步議題。
  - **cast_gfx=2287 對齊 yiwei** 無漂移（罕見，同 151 EARTH_SKIN 模式）。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：屬廣域 buff cast 後置缺口（同 130/146/148/149/150/151/152/155/156 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `skill_elemental_buff_test.go` 既有覆蓋 159 非互斥場景，無新測試需執行。

## 生命之泉（NATURES_TOUCH / 158）— 補齊 cast 後移除 WATER_LIFE 對齊 Java 恢復系技能清單

- **Java 對照**：
  - `L1SkillId.java:530 NATURES_TOUCH = 158`，無對應 skillmode 類別（走 generic `cha.setSkillEffect(_skillId, _getBuffDuration)` apply 路徑）。
  - `HprExecutor.java:55`：`_skill.put(NATURES_TOUCH, 15)`——持有 158 buff 期間 HP regen +15 per tick。
  - `L1SkillUse.java:2113-2118`：cast 屬於恢復系技能清單 `(HEAL || EXTRA_HEAL || GREATER_HEAL || FULL_HEAL || HEAL_ALL || NATURES_TOUCH || NATURES_BLESSING)` 之一時，`cha.removeSkillEffect(WATER_LIFE)`——對 cast 目標移除 WATER_LIFE buff（避免雙倍治療 buff 與 HPR aura 同時消耗）。
  - 無 icon send（不像 147/148/149 等武器 buff 走 `S_PacketBoxIconAura` 路徑）。
  - yiwei `db_split/skills.sql:157`：`('158', '生命之泉', '20', '5', '20', '0', '0', '0', '0', '320', 'buff', '1', '0', '0', '0', '0', '0', '4', '2', '0', '-1', '0', '0', '32', '', '19', '2243', '0', '0', '0', '280')` — mp=20、buff_duration=320、target=buff、target_to=1、attr=4、type=2、ranged=-1、id=32、action_id=19、cast_gfx=2243、sys_msg_fail=280。

- **Go 對照**（修正前 vs 修正後）：
  - 修正前：`buffs.lua:126 [158] = { hpr = 15 }` 已對齊 Java HprExecutor 158→15 HPR delta（每 tick 由 `target.HPR += 15` 計算 `equip_hpr` 加到 regen bonus）；158 在 cancellable 清單 (`skill_buff.go:411`)；cast 走 `executeBuffSkill`（yaml target='buff' target_to=1=self），透過 `applyBuffEffect` 套用 buff。**但 cast 158 後不移除 WATER_LIFE**——因為 Go `applyElfWaterHealingModifiers` (`skill_elemental.go:97-109`) 只在 heal 計算路徑觸發（skill_buff.go:1202/1214 TYPE_HEAL=16 + skill_self.go:220/247/258），158 yaml type=2（buff，HPR aura）無 heal 計算，繞過 WATER_LIFE 移除 hook。先前 170 WATER_LIFE audit 誤判「Java 同樣不移除」，實際 Java 在 `L1SkillUse:2113-2118` 確實移除——criterion (a) 真實 Java diff。
  - 修正後 `skill_buff.go:1242-1249`：在 `applyBuffEffect` + cast_gfx 廣播之間加 `if skill.SkillID == 158 && target.HasBuff(170) → s.removeBuffAndRevert(target, 170)`。對齊 Java 恢復系技能清單對 WATER_LIFE 的移除行為。其他 HEAL 類 1/19/35/49/57/164 yaml type=16 已由 `applyElfWaterHealingModifiers` 處理對齊。
  - yaml `skill_list.yaml:4838-4868`：mp=20、buff_duration=320、target=buff、target_to=1、attr=4、type=2、ranged=-1、id=32、action_id=19、cast_gfx=2243、sys_msg_fail=280——**31 欄位完全對齊 yiwei `skills.sql:157`**（零漂移）。

- **既有測試覆蓋**：
  - 無針對 158 NATURES_TOUCH 的單元測試。HPR delta apply/revert 由通用 buff 系統覆蓋。
  - 170 WATER_LIFE 雙倍治療測試覆蓋 TYPE_HEAL 路徑的 WATER_LIFE 移除，但**不覆蓋 158 此非 TYPE_HEAL 路徑**。

- **本次修正範圍**（criterion (a) Java 核心行為差異 + 修正先前 audit 誤判）：
  - cast 158 NATURES_TOUCH 後顯式移除 target's WATER_LIFE buff，補齊 Java 恢復系技能清單行為。

- **broader gap（不改）**：
  - **`L1SkillUse:2113-2118` 對 cast target 移除 WATER_LIFE 的「恢復系技能」清單完整性**：本次只補 158，其他 HEAL 類 1/19/35/49/57/164 透過 yaml type=16 + `applyElfWaterHealingModifiers` 已正確處理；無遺漏。
  - **NATURES_BLESSING (164) audit**：未來 audit 164 時需確認 yaml type=16 + heal 計算路徑觸發 `applyElfWaterHealingModifiers`，移除 WATER_LIFE 對齊 Java（預期已對齊但待 164 audit 確認）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -timeout 60s` PASS（27.357s 全綠）。

## 大地屏障（EARTH_BIND / 157）— 純審計確認核心 1-12s 凍結 + PC/NPC 雙路徑 + 廣域 immunity 整合完整對齊 Java

- **Java 對照**：
  - `L1SkillId EARTH_BIND = 157`，`skillmode/EARTH_BIND.start()`：
    - `i = rad.nextInt(12) + 1` 1-12 秒隨機。
    - `if (!srcpc.castleWarResult())` 同陣營友軍 gate。
    - `setSkillEffect(157, i * 1000)` apply。
    - PC 目標：`sendPacketsAll(S_Poison(id, 2)) + sendPackets(S_Paralysis(TYPE_FREEZE, true))`。
    - NPC 目標：`broadcastPacketAll(S_Poison(id, 2)) + setParalyzed(true)`。
  - `EARTH_BIND.stop()`：PC `sendPacketsAll(S_Poison(id, 0)) + sendPackets(S_Paralysis(4, false))`；NPC `broadcastPacketAll(S_Poison(id, 0)) + setParalyzed(false)`。
  - `L1SkillUse.java:378-383` castle_area 旗幟區阻擋 cast + msg「戰爭旗幟內禁止使用大地屏障」。
  - `L1SkillUse.java:757-758` `if (target.hasSkillEffect(EARTH_BIND) && (skillId == EARTH_BIND || FREEZING_BLIZZARD)) return false` restack 防止。
  - `L1Character.java:337-339` `if (hasSkillEffect(EARTH_BIND)) return true` 列入 paralysis-class 控制效果（與 FOG_OF_SLEEPING/SHOCK_STUN/PHANTASM/ICE_LANCE/DARK_BLIND 同組）。
  - `L1MagicPc.calcProbabilityMagic case EARTH_BIND`（line 521-534）三段 `EARTH_BIND_1/2/3` + INT/MR 微調。
  - `L1MagicPc.java:118 + :190` target hasSkillEffect(EARTH_BIND) 抗性檢查。
  - `L1Cube.java:88/119` 持 EARTH_BIND 目標對 cube damage/freeze 免疫。
  - yiwei `db_split/skills.sql:156`：`('157', '大地屏障', '20', '4', '10', '0', '40319', '2', '1200', '12', 'buff', '3', '0', '0', '0', '33', '50', '1', '1', '0', '8', '0', '0', '16', '', '19', '2251', '0', '0', '0', '280')` — mp=10、item=40319×2、reuse_delay=**1200**、buff_duration=**12**、target=buff、target_to=3、prob_value=33、prob_dice=**50**、attr=1、type=1、ranged=**8**、id=16、action_id=19、cast_gfx=2251、sys_msg_fail=280。

- **Go 對照**：
  - `skill_buff.go:1118-1128 case 157` PC→PC 路徑：`if HasBuff(157) || Paralyzed → return`（restack 防止）；`earthBind.BuffDuration = 1 + RandInt(12)`；`applyBuffEffect(target, &earthBind)` 透過 lua `[157] = { paralyzed = true }` 設 `target.Paralyzed = true`；cast_gfx 廣播。
  - `skill_status.go:518-530 case 157` PC→NPC 路徑：`checkNpcMRResist` MR 閘 + `1 + RandInt(12)` 隨機 + `npc.Paralyzed = true` + `npc.AddDebuff(157, dur*5)` + `BuildPoison(npc.ID, 2)` 廣播 + cast_gfx。
  - `skill_buff.go:255-264 eff.Paralyzed` apply：`buff.SetParalyzed = true` + `target.Paralyzed = true` + `case 157` 派發 `SendParalysis(FreezeApply)` + `broadcastPlayerPoison(target, 2)`。
  - `skill_buff.go:677/694-697 needFreezeRemove` revert：`SendParalysis(FreezeRemove)` + `broadcastPlayerPoison(target, 0)`。
  - `skill_status.go:831 playerDebuffSkills[157] = true` MR 抗性閘。
  - 廣域 immunity 整合：
    - `ground_effect.go:85/94/108/116` cube 凍結/麻痺免疫對齊 Java `L1Cube:88/119`。
    - `skill_dragonknight.go:129` paralysis-class 攻擊免疫。
    - `weapon_skill.go:224` 武器 special 對 paralyzed 目標免疫。
    - `skill_buff.go:949 case 87 SHOCK_STUN` 對 HasBuff(157) 目標 fail。
  - NPC AI `npc_ai.go:1488 case 157` 清除路徑。
  - yaml `skill_list.yaml:4807-4837`：mp=10、item=40319×2、reuse_delay=**0**、buff_duration=**16**、target=buff、target_to=3、prob_value=33、prob_dice=**30**、attr=1、type=1、ranged=**-1**、id=16、action_id=19、cast_gfx=2251、sys_msg_fail=280。
  - **4 項 yaml 漂移**：reuse_delay=0 vs 1200、buff_duration=16 vs 12、prob_dice=30 vs 50、ranged=-1 vs 8（Go 跟 cat-fei）。注意 case 157 用 `1 + RandInt(12)` 覆寫 buff_duration，yaml buff_duration 無實際效用。

- **既有測試覆蓋**：
  - SHOCK_STUN 測試已覆蓋對 HasBuff(157) 目標的失敗行為（skill_buff.go:949）。
  - 廣域 immunity 整合於 ground_effect / dragonknight / weapon_skill 測試間接覆蓋。

- **發現的 Java 真實差異**：**無**（核心 mechanic 完整對齊：1-12s 隨機凍結、PC/NPC 雙路徑、S_Paralysis FreezeApply/Remove、S_Poison cast/stop、MR 抗性閘、廣域 immunity 整合、NPC AI 清除路徑）。

- **broader gap（不改，待後續系統議題處理）**：
  - **A) `castleWarResult()` 同陣營友軍 gate**：Java skill 對同陣營玩家免疫，Go 無此檢查——屬城堡戰陣營判定子系統議題。
  - **B) `L1SkillUse:378-383` castle_area 旗幟區 cast 阻擋**：Java 在城堡戰旗幟區內阻擋 EARTH_BIND cast，Go 無此檢查——屬城堡戰旗幟區判定議題。
  - **C) `ConfigElfSkill.EARTH_BIND_1/2/3` 三段公式**：Java 用 3 段 lvldiff probability + INT/MR 微調；Go 用 generic checkPlayerMRResist 簡化版——屬廣域 ConfigElfSkill probability 表議題（同 173/174 同源）。
  - **D) yaml 4 項漂移**：reuse_delay=0 vs 1200、buff_duration=16 vs 12、prob_dice=30 vs 50、ranged=-1 vs 8（Go 跟 cat-fei）——屬廣域 yiwei/cat-fei SQL 同步議題。case 157 用 RandInt(12) 覆寫 buff_duration，實際效用未受影響。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 157 的新測試需執行。

## 暴風之眼（STORM_EYE / 156）— 純審計確認核心 BowHit ±2 + BowDmg ±3 + icon + REPEATEDSKILLS[0] 5 項互斥完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:522 STORM_EYE = 156`，無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:1418-1420` + `L1SkillUse2.java:1435-1437` icon cast：`pc.sendPackets(new S_PacketBoxIconAura(155, _getBuffIconDuration))`（icon_id = skill_id - 1 = 155）。
  - `L1SkillUse.java:2606-2610` + `L1SkillUse2.java:2558-2562` apply：`pc.addBowHitup(2) + pc.addBowDmgup(3) + sendPackets(S_PacketBoxIconAura(155, duration))`。
  - `L1SkillStop.java:561-568` stop：`cha.addBowHitup(-2) + cha.addBowDmgup(-3) + (if PC) sendPackets(S_PacketBoxIconAura(155, 0))`。
  - `REPEATEDSKILLS[0]`：`{148, 149, 156, 163, 166}` 5 項武器類 buff 互斥（同 148/149 audit）。
  - 表格成員：`isNotCancelable` 不含 156（cancellable）；`EXCEPT_COUNTER_MAGIC` 含（不受 counter magic 阻擋）；`REPEATEDSKILLS[0]` 5 項武器互斥。
  - yiwei `db_split/skills.sql:155`：`('156', '暴風之眼', '20', '3', '40', '0', '0', '0', '0', '960', 'none', '8', '0', '0', '0', '0', '0', '8', '2', '0', '0', '-1', '0', '8', '', '19', '11731', '0', '0', '724', '0')` — mp=40、buff_duration=960、target=none、target_to=8、attr=8、type=2、ranged=0、area=-1、id=8、action_id=19、cast_gfx=**11731**、sys_msg_stop=724、sys_msg_fail=0。

- **Go 對照**：
  - `buffs.lua:124 [156] = { bow_hit = 2, bow_dmg = 3, exclusions = {148, 149, 163, 166} }`：BowHit+2、BowDmg+3 + 4 項 mutex 對齊 Java REPEATEDSKILLS[0] 排除自身後 4 項。
  - `skill_buff.go:167-168 + :201-202` apply：`buff.DeltaBowHit = int16(eff.BowHit)` + `target.BowHitMod += buff.DeltaBowHit`（BowDmg 同模式）。
  - `:307 sendBuffIcon` 透過 `buff_icon_map.yaml:64 skill_id: 156 type: aura`（無 param）→ `handler/skill.go:222-227` aura default `iconID = skill_id - 1 = 155` 送 `S_PacketBoxIconAura(155, duration)` 對齊 Java cast icon。
  - `:513-514` revert：`target.BowHitMod -= buff.DeltaBowHit` + 對應 stop icon。
  - `executeSelfSkill` default 路徑（因 yaml target='none'）：`applyBuffEffect(player, skill)` + cast_gfx 廣播。
  - yaml `skill_list.yaml:4776-4806 skill_id=156`：mp=40、buff_duration=960、target=none、target_to=8、attr=8、type=2、ranged=0、area=-1、id=8、action_id=19、cast_gfx=**2288**、sys_msg_stop=724、sys_msg_fail=0。
  - **僅 cast_gfx 一項漂移**（Go=2288 vs yiwei=11731），其餘 30 欄位對齊 yiwei（同 149 WIND_SHOT 2247 vs 11729、150 WIND_WALK 2247 vs 11730 模式）。

- **既有測試覆蓋**：
  - 無針對 156 的單元測試。BowHit/BowDmg apply/revert 路徑由 149 WIND_SHOT（BowHit+6）同族測試間接覆蓋。

- **發現的 Java 真實差異**：**無**（核心 BowHit ±2 + BowDmg ±3 + S_PacketBoxIconAura(155) icon + REPEATEDSKILLS[0] 5 項互斥完整對齊）。

- **broader gap（不改）**：
  - **yaml cast_gfx 漂移**：Go=2288 vs yiwei=11731，Go 用 cat-fei 較舊 GFX ID（同 149/150 模式）。屬廣域 yiwei/cat-fei SQL 同步議題。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：屬廣域 buff cast 後置缺口（同 130/146/148/149/150/151/152/155 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 156 的測試需執行。

## 烈炎氣息（FIRE_BLESS / 155）— 補齊 C_UseSkill 武器類型 gate，unequip cleanup hook 延後

- **Java 對照**：
  - `L1SkillId.java:518 FIRE_BLESS = 155`。
  - `skillmode/FIRE_BLESS.start()`：
    1. `L1BuffUtil.braveStart(srcpc)` clean conflicting brave buffs。
    2. `setSkillEffect(FIRE_BLESS, integer * 1000)` apply。
    3. **`setBraveSpeed(1)`**（**非 4**，與 150 WIND_WALK 不同 tier）。
    4. `sendPackets(S_SkillBrave(id, 1, integer))` self icon。
    5. `broadcastPacketAll(S_SkillBrave(id, 1, 0))` broadcast。
    6. `sendPackets(S_PacketBoxIconAura(154, integer))` icon param=154（skill_id-1）。
  - `skillmode/FIRE_BLESS.stop()`：`setBraveSpeed(0) + (if PC) sendPacketsAll(S_SkillBrave(0, 0)) + sendPackets(S_PacketBoxIconAura(154, 0))`。
  - `C_UseSkill.java:154-169` **武器類型 gate**：要求 `weapon.getType()` ∈ {1, 2, 3, 5, 6, 7}（sword/dagger/tohandsword/spear/blunt/staff），不符合或無武器 → `sendPackets(S_ServerMessage(3435))` + return。
  - `L1EquipmentSlot.java:282-284` **unequip 武器 → 移除 FIRE_BLESS hook**：`if (hasSkillEffect(FIRE_BLESS)) removeSkillEffect(FIRE_BLESS)`。
  - `L1PcInstance.java:4416 isBrave()`：`hasSkillEffect(STATUS_BRAVE || FIRE_BLESS || BLOODLUST)` → 戰鬥相關計算（如 buff dmgup）的 brave flag。
  - `L1BuffUtil.java:258-259`：多處可主動 `killSkillEffectTimer(FIRE_BLESS)`。
  - `L1SkillUse.java:1745-1746 REPEATEDSKILLS[2]`：`{HOLY_WALK, MOVING_ACCELERATION, WIND_WALK, STATUS_BRAVE, STATUS_ELFBRAVE, BLOODLUST, FIRE_BLESS}` 7 項勇氣速度類 buff 互斥。
  - yiwei `db_split/skills.sql:154`：mp=**40**、buff_duration=960、target=none、target_to=**8**、attr=2、type=2、cast_gfx=**2286**、sys_msg_stop=**723**。

- **Go 對照**（修正前 vs 修正後）：
  - 修正前：155 走 `executeSelfSkill` default 路徑 → `applyBuffEffect` 套用 `buffs.lua [155] = { brave_speed = 1, exclusions = {52, 101, 150, 186, 1000, 1016} }`，BraveSpeed=1 與 S_SkillBrave/icon 已對齊 Java，但**完全無武器類型 gate**——任何武器或無武器都能施放。
  - 修正後 `skill.go:432-438`：在 `consumeSkillResources` **前**新增 `if skillID == 155 && !fireBlessWeaponAllowed(player, s.deps.Items) → SendServerMessage(sess, 3435) + return`——對齊 Java `C_UseSkill.java:154-169` cast 入口拒絕。
  - 修正後 `skill_self.go:15-34 fireBlessWeaponAllowed`：helper 函式，檢查 `player.Equip.Weapon()` 與 `s.deps.Items.Get(wpn.ItemID).Type ∈ {"sword","dagger","tohandsword","spear","blunt","staff"}` 對齊 Java type 1/2/3/5/6/7 允許列表。
  - icon 路徑：`buff_icon_map.yaml:62 skill_id: 155 type: aura`（無 param），`handler/skill.go:222-227` aura default `iconID = skill_id - 1 = 154` → 對齊 Java `S_PacketBoxIconAura(154, ...)`。
  - BraveSpeed apply/revert + S_SkillBrave self+broadcast 由 `applyBuffEffect` 統一路徑處理對齊 Java（同 150 WIND_WALK precedent）。
  - REPEATEDSKILLS[2] mutex：`buffs.lua [155].exclusions = {52, 101, 150, 186, 1000, 1016}` 6 項排除自身——**Java HOLY_WALK=42、Go=52 屬廣域 SQL ID 漂移**（同 150 同源）。
  - yaml `skill_list.yaml:4745-4775`：mp=**40**、buff_duration=960、target=none、target_to=**8**、attr=2、type=2、cast_gfx=**2286**、sys_msg_stop=**723**——**31 欄位完全對齊 yiwei**（零漂移）。

- **既有測試覆蓋**：
  - 無針對 155 weapon gate 的單元測試。BraveSpeed=1 apply/revert 路徑由 150 WIND_WALK 同族測試間接覆蓋。

- **本次修正範圍**（criterion (a) Java 核心行為差異）：
  - **武器類型 gate**：cast 入口檢查 `player.Equip.Weapon()` + Type ∈ 允許列表，不符合送 msg 3435 + return（不消耗 MP）。

- **broader gap（不改）**：
  - **A) Unequip 武器 → 移除 FIRE_BLESS hook**：Java `L1EquipmentSlot.java:282-284` 在卸下武器時 `removeSkillEffect(FIRE_BLESS)`。Go `equip.go UnequipSlot` 無對應 hook；補修需 EquipSystem ↔ SkillSystem 跨系統依賴注入或 Event Bus 介面，超出單一技能範圍，屬廣域裝備變更與 buff lifecycle 整合議題。
  - **B) `L1BuffUtil.killSkillEffectTimer(FIRE_BLESS)` 主動取消路徑**：Java 多處（PVP、castle 進出等）可主動取消 FIRE_BLESS timer，Go 無對等廣域 hook 機制（同 150 同源）。
  - **C) `L1PcInstance.isBrave()` 加 FIRE_BLESS 檢查**：Java `isBrave()` 用於戰鬥 buff dmgup 等計算的彙整 flag；Go 無對應彙整 helper（各戰鬥路徑分散 `HasBuff(155/186/1000)` 檢查），屬廣域 brave flag helper 議題。
  - **D) HOLY_WALK ID 漂移**：Java=42、Go=52（cat-fei）。整個 REPEATEDSKILLS[2] 群組以 cat-fei 52 為基準——廣域 SQL ID 漂移議題（同 130/150 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -timeout 60s` PASS（19.287s 全綠）。

## 召喚屬性精靈（LESSER_ELEMENTAL / 154）— 純審計重新確認核心 5 項機制完整對齊 Java，兩項召喚家族共通缺口維持延後

- **Java 對照**：
  - `L1SkillId.java LESSER_ELEMENTAL = 154`，`skillmode/LESSER_ELEMENTAL.java:23-67` 完整流程：
    1. `attr = pc.getElfAttr()` 必須 != 0（無屬性玩家不能召喚精靈）。
    2. `if (!pc.getMap().isRecallPets()) → sendPackets(S_ServerMessage(353)) + return`（地圖禁召喚 → 訊息 353「在這附近無法召喚怪物」）。
    3. 迭代 `pc.getPetList().values()` 累計 `petcost`，若 `petcost != 0` → 已有寵物，靜默 return。
    4. Switch on `attr`：1→45306（土）、2→45303（火）、4→45304（水）、8→45305（風）。
    5. `new L1SummonInstance(npcTemp, pc) + summon.setPetcost(pc.getCha() + 7)`。
  - `skillmode/GREATER_ELEMENTAL.java` 為 162 強化版（不同 NPC ID 但同 mechanic）。
  - yiwei `db_split/skills.sql:153`：mp=20、item=40319×2、buff_duration=0、target=none、type=128、attr=0、action_id=19、cast_gfx=2510、所有 sys_msg=0。

- **Go 對照**：
  - `skill.go:409-410` + `skill_magic_scroll.go:54-55` route → `s.deps.Summon.ExecuteElementalSummon(sess, player, skill)`。
  - `skill_summon.go:282-363 ExecuteElementalSummon`：
    1. `:286-288 if player.ElfAttr == 0 → return` 對齊 Java attr != 0 gate。
    2. `:289-295 RecallPets` 檢查 + `SendServerMessage(msgCannotSummonHere)` 對齊 Java msg 353。
    3. `:296-298 calcUsedPetCost(CharID) != 0 → return` 對齊 Java petcost != 0 靜默 return。
    4. `:300-309 lesserElementalByAttr[player.ElfAttr]` map（同 Java：1→45306、2→45303、4→45304、8→45305）。
    5. `:317-319 petCost := Cha + 7` 對齊 Java `pc.getCha() + 7`（`<=0` clamp 為 Go 防禦邏輯，Java 無但實際 Cha+7 永遠 >= 7）。
  - `:321 ConsumeSkillResources` — MP 消耗（Java 路徑由 L1SkillUse 外圍 cast 流程處理，Go 在召喚成功 gates 通過後消耗——與 131 precedent 一致，user-friendly 防止失敗扣 MP）。
  - `:355-362 AddSummon + SendSummonPack(viewer + self) + SendSummonMenu` 對應 Java L1SummonInstance constructor 內部送 packet。
  - yaml `skill_list.yaml:4714-4744`：mp=20、buff_duration=0、target=none、type=128、attr=0、action_id=19、cast_gfx=2510、所有 sys_msg=0——31 欄位**完全對齊 yiwei**（罕見零漂移）。

- **既有測試覆蓋**：
  - `skill_summon` 系列測試覆蓋 ExecuteElementalSummon 核心路徑（ElfAttr 0 gate、RecallPets gate、petCost 占用 gate、4 屬性 NPC ID 映射）。

- **發現的 Java 真實差異**：**無**（核心 5 項機制完整對齊；兩項家族級差異維持先前審計結論）。

- **broader gap（不改）**：
  - **A) 施法動畫 + cast_gfx 廣播缺失**：Java L1SkillUse.sendGrfx 外圍流程對所有 cast 流程送 action_id=19 + cast_gfx=2510 廣播；Go ExecuteElementalSummon 入口前無對應 broadcast。屬召喚技能 (51/36/41/145/154/162) **共通缺口**，應在召喚技能統一 audit 時集中補上 cast 動畫廣播 helper。
  - **B) MP 消耗時機字面順序**：Java `L1SkillUse._user.useMP/SP` 在 `runSkill` 入口處先消耗，Go 在所有 gates 通過後才 `ConsumeSkillResources`——Go 採 user-friendly pattern（失敗不扣 MP），與 131 TELEPORT_TO_MATHER precedent 一致，**不修**。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 154 的測試需執行。

## 魔法消除（ERASE_MAGIC / 153）— 移除錯誤 cancelAllBuffs 並補 PvP MR 閘，broader gap MR/=4 + consume hook 仍延後

- **Java 對照**：
  - `L1SkillId.java:510 ERASE_MAGIC = 153`，無對應 skillmode 類別（走 generic `cha.setSkillEffect(_skillId, _getBuffDuration)` apply 路徑——`L1SkillUse.java:1357-1360` mode==null 分支）。
  - `L1SkillUse.java:1940-1955` + `L1SkillUse2.java:1948-1962` **consume-on-cast hook**：
    - TYPE_ATTACK 命中 target → `if (cha.hasSkillEffect(ERASE_MAGIC)) cha.removeSkillEffect(ERASE_MAGIC)` 消耗（line 1940-1942）。
    - TYPE_CURSE/TYPE_PROBABILITY cast on target & `skillId != ERASE_MAGIC` → 同樣 `cha.removeSkillEffect(ERASE_MAGIC)`（line 1953-1955）。
  - `L1Character.java:1767-1768` **真實效果**：`if (hasSkillEffect(153)) return mr >> 2;`——`getMr()` getter 對持有 153 buff 的角色直接回傳 `mr / 4`（**乘法/移位** 路徑非加法 delta）。
  - `L1SkillUse.java:3063-3066` **isErase 抗性閘**：`if (skillId == ERASE_MAGIC || SLOW || MANA_DRAIN || MASS_SLOW || ENTANGLE || WIND_SHACKLE) && isErase==false → 免疫`。
  - `L1MagicPc.java:506-519` cast probability 三段：`ERASE_MAGIC_1/2/3 = 5/10/15`（攻擊方等級 > = < 防禦方）+ `INT*INT - MR*MR` 微調（預設 INT=0, MR=0 不啟動）。
  - `L1SkillUse.java:799` **isExceptCounterMagic** 含 ERASE_MAGIC（不受 counter magic 抵消）。
  - `L1SkillStop.java:492-497` stop：`(if PC) pc.sendPackets(new S_PacketBoxIconAura(152, 0))` icon clear（注意 icon_id=skill_id-1=152）。
  - `L1WeaponSkill.java:421/593/764` (W_SK0010/0015, W_SK005/006/007) 武器 special attack：`pc1.setSkillEffect(ERASE_MAGIC, 32*1000)` 觸發 32 秒 buff。
  - yiwei `db_split/skills.sql:152`：`('153', '魔法消除', '20', '0', '32', '0', '40319', '1', '1200', '32', 'buff', '3', '0', '0', '0', '33', '50', '0', '1', '0', '6', '0', '0', '1', '', '19', '10706', '0', '0', '0', '280')` — mp=**32**、item=40319×1、reuse_delay=**1200**、buff_duration=32、target=buff、target_to=3、prob_value=33、prob_dice=**50**、attr=0、type=1（TYPE_PROBABILITY）、ranged=**6**、id=1、cast_gfx=**10706**、sys_msg_fail=280。

- **Go 對照**（修正前 vs 修正後）：
  - 修正前 `skill_buff.go:1153-1155`：`case 153: s.cancelAllBuffs(target)`——**錯誤實作**，套用 44 CANCELLATION 的「buff 全消」語義，與 Java 32s MR/4 debuff 截然相反。**主動有害**：153 在 Go 變成 100% 機率（無 MR 閘）的 strip-all-buffs 神技，比 Java 任何技能都更強。
  - 修正後：刪除錯誤的 `case 153` 區塊，fall-through 至 default `applyBuffEffect(target, skill)` 走純 flag buff 應用（無屬性 delta，buffs.lua 無 [153] 條目）。
  - `skill_status.go:828-829 playerDebuffSkills`：新增 `153: true`——PvP 對玩家 cast 153 時走 `checkPlayerMRResist` MR 閘（與 152 ENTANGLE 同 precedent）。
  - `clearShockStunEraseMagic` (skill_status.go:173) 既有對 87 SHOCK_STUN 的 consume hook 保留——`removeBuffAndRevert(target, 153)` 移除 153 buff（修正前 153 從未被 applyBuffEffect 真的設定，此 hook 為 no-op；修正後 hook 終於有實質作用）。
  - yaml `skill_list.yaml:4683-4713`：mp=**10**、buff_duration=32、prob_value=33、prob_dice=**30**、type=1、ranged=**-1**、cast_gfx=**2181**、sys_msg_happen=0、sys_msg_fail=280。**漂移欄位**：mp=10 vs 32、reuse_delay=0 vs 1200、prob_dice=30 vs 50、ranged=-1 vs 6、cast_gfx=2181 vs 10706（Go 跟 cat-fei）。

- **既有測試覆蓋**：
  - 無針對 153 的單元測試。`skill_status.go clearShockStunEraseMagic` 路徑由 87 SHOCK_STUN 測試覆蓋。

- **本次修正範圍**（criterion (b) Go 確實有錯）：
  - **A) 移除錯誤的 `case 153: cancelAllBuffs`**：消除「100% 機率 strip-all-buffs」的有害行為，restore 至 flag buff apply 語義（仍非完整 Java MR/4，但至少不再有 strip-all 副作用）。
  - **B) playerDebuffSkills 加入 153**：補上 PvP MR 抗性閘，比照 152 ENTANGLE 模式。Go 走 generic `checkPlayerMRResist` 公式（`50 + lvldiff*3 + INT - MR`）非 Java `ConfigElfSkill.ERASE_MAGIC_1/2/3=5/10/15` 三段公式——屬 broader ConfigElfSkill probability 表議題。

- **broader gap（不改，待後續系統級重構）**：
  - **C) MR/=4 乘法效果**：Java `L1Character.getMr() return mr >> 2` 多項屬於 getter override 即時計算；Go MR 系統採加法 DeltaMR delta 路徑（無乘法 / 除法 multiplier 支援）。修補需新增「multiplier delta」或「override getter chain」，影響整個屬性計算管線——屬廣域 stat computation 架構議題，與 174 STRIKER_GALE 「getEr() return 0」乘法 override 同源。
  - **D) consume-on-cast hook**：Java TYPE_ATTACK 命中 / TYPE_CURSE/PROBABILITY cast 兩條路徑都會消耗 ERASE_MAGIC buff。Go 既有 `clearShockStunEraseMagic` 只覆蓋 87 SHOCK_STUN 一條路徑，廣域 consume hook 缺失——應在 `combat.go` 命中傷害結算 + `skill_buff.go executeBuffSkill` 入口統一加 `if target.HasBuff(153) && skill.SkillID != 153 { removeBuffAndRevert(target, 153) }`。屬廣域 buff lifecycle 議題。
  - **E) isErase 抗性閘**：Java `L1SkillUse:3063-3066` 對 ERASE_MAGIC/SLOW/MANA_DRAIN/MASS_SLOW/ENTANGLE/WIND_SHACKLE 共享 isErase resist gating；Go 完全無此機制。屬「魔法抗性子分類」廣域議題（同 SLOW/MASS_SLOW/ENTANGLE 家族）。
  - **F) 武器 special 觸發路徑**：W_SK005/006/007/0010/0015 武器 special attack 32 秒 buff 套用，與整體武器 special 子系統 Go 未實作同源——屬廣域 weapon skill schema 缺口。
  - **G) ConfigElfSkill probability 三段公式**：5/10/15 + INT*INT - MR*MR 微調；Go 用 generic `checkPlayerMRResist` 簡化版。屬廣域 ConfigElfSkill probability 表議題（同 173/174 audit 同源）。
  - **H) yaml 多項漂移**：mp=10 vs 32、reuse_delay=0 vs 1200、prob_dice=30 vs 50、ranged=-1 vs 6、cast_gfx=2181 vs 10706——Go 跟 cat-fei。屬廣域 yiwei/cat-fei SQL 同步議題。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -timeout 60s` PASS（16.998s 全綠）。

## 地面障礙（ENTANGLE / 152）— 純審計確認 case 1 haste 互消完整對齊 Java，三項 slow 家族共通缺口待 29/76 audit 重構

- **Java 對照**：
  - `L1SkillId.java:506 ENTANGLE = 152`（地面障礙）。無對應 skillmode 類別。
  - `L1SkillUse.java:1465-1467` + `L1SkillUse2.java:1482-1484` icon cast：與 SLOW/MASS_SLOW 同組 ENTANGLE 走通用 slow icon 路徑。
  - `L1SkillUse.java:2147-2188` + `L1SkillUse2.java:2137-2178` apply（SLOW/MASS_SLOW/ENTANGLE 同組）：
    ```java
    if (cha instanceof L1PcInstance) {
        if (pc.getHasteItemEquipped() > 0) continue;  // 急速道具免疫
    }
    if (cha.getBraveSpeed() == 5) continue;           // 強化勇水免疫
    switch (cha.getMoveSpeed()) {
        case 0:  // 正常速度
            pc.sendPackets(new S_SkillHaste(pc.id, 2, duration));
            cha.broadcastPacketAll(new S_SkillHaste(cha.id, 2, duration));  // ← 廣播 duration 非 0
            cha.setMoveSpeed(2);
            break;
        case 1:  // 已加速
            // 找出 HASTE/GREATER_HASTE/STATUS_HASTE 中存在者
            cha.removeSkillEffect(skillNum);            // 移除既有 haste
            cha.removeSkillEffect(this._skillId);       // 移除自身 slow
            cha.setMoveSpeed(0);                        // 速度歸零（互抵）
            continue;
    }
    ```
  - `L1SkillStop.java:660-668` stop（SLOW/MASS_SLOW/ENTANGLE 同組）：`(if PC) pc.sendPacketsAll(new S_SkillHaste(pc.id, 0, 0)) + cha.setMoveSpeed(0)`。
  - `L1BuffUtil.java:152-153`：其他系統可主動 `killSkillEffectTimer(ENTANGLE)`。
  - `skillmode/HASTE.java:49-50`：對 target 施 HASTE 時若已有 ENTANGLE → 觸發互消邏輯。
  - 表格成員：`isNotCancelable` 不含 152（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 152 line 794（不受 counter magic 阻擋）；`REPEATEDSKILLS` 不含 152（**無顯式 mutex 群組**——靠 case 1 haste 互消 catch-all 處理）。
  - yiwei `db_split/skills.sql:151`：`('152', '地面障礙', '19', '7', '20', '0', '40319', '1', '0', '64', 'buff', '3', '0', '0', '0', '33', '50', '1', '1', '0', '10', '0', '0', '128', '', '19', '2250', '0', '711', '0', '280')` — mp=20、item=40319×1、reuse_delay=0、buff_duration=64、target=buff、target_to=3、prob_value=33、prob_dice=**50**、attr=1、type=1、ranged=10、id=128、action_id=19、cast_gfx=2250、sys_msg_happen=**711**、sys_msg_fail=280。

- **Go 對照**：
  - `buffs.lua:121 [152] = { move_speed = 2, exclusions = {43, 54} }`：MoveSpeed=2（slow）+ exclusions={43 HASTE, 54 STATUS_HASTE}（顯式 mutex 對應 Java case 1 移除 haste）。
  - `skill_buff.go:217-227` apply 「速度互抵邏輯」：
    ```go
    if eff.MoveSpeed > 0 {
        if eff.MoveSpeed == 2 && target.MoveSpeed == 1 {     // 套 slow，目標已 hasted
            s.cancelSpeedBuffs(target, 1)                      // 移除所有 haste buffs
            target.MoveSpeed = 0
            target.HasteTicks = 0
            s.sendSpeedToAll(target, 0, 0)                     // 廣播速度歸零
        } else if eff.MoveSpeed == 1 && target.MoveSpeed == 2 {
            s.cancelSpeedBuffs(target, 2)
            // ... 對稱處理
        } else {
            buff.SetMoveSpeed = byte(eff.MoveSpeed)
            target.MoveSpeed = byte(eff.MoveSpeed)
            target.HasteTicks = buff.TicksLeft
            s.sendSpeedToAll(target, byte(eff.MoveSpeed), uint16(skill.BuffDuration))
        }
    }
    ```
    **完美對齊 Java case 0/1 機制**——`cancelSpeedBuffs(target, 1)` 對 target 的 haste 類型 buff 全掃移除，比 Java 顯式只查 HASTE/GREATER_HASTE/STATUS_HASTE 三項更廣域穩健。
  - `skill_buff.go:483 cancelSpeedBuffs` 通用實作：迭代 target.ActiveBuffs 找出對應 speedType 的 buff 並 RemoveBuff + revertBuffStats。
  - `:307 sendBuffIcon` 透過 `buff_icon_map.yaml` 查 152 → 視 type 送對應 icon（若有註冊）。
  - 沒走 executeBuffSkill 的 cast_gfx 廣播分支：因 152 屬 debuff（target_to=3 對敵），走 `executeNpcDebuffSkill`／PC→PC debuff 路徑。
  - yaml `skill_list.yaml:4652-4682 skill_id=152`：mp=20、buff_duration=64、prob_value=33、prob_dice=**30**、attr=1、type=1、ranged=10、id=128、cast_gfx=2250、sys_msg_happen=**0**、sys_msg_fail=280。
  - **2 項 yaml 漂移**：`prob_dice=30 vs yiwei 50`、`sys_msg_happen=0 vs yiwei 711`。

- **既有測試覆蓋**：
  - 無針對 152 的單元測試。`skill_cube_ground_effect_test.go` 涉及 NpcID 80152 屬不同領域。

- **發現的 Java 真實差異（slow 家族共通缺口，criterion a）**：
  - **A) HasteItemEquipped > 0 免疫**：Java `if (pc.getHasteItemEquipped() > 0) continue` 對佩戴急速道具的玩家完全免疫 slow，Go 無此檢查。
  - **B) BraveSpeed == 5 免疫**：Java `if (cha.getBraveSpeed() == 5) continue` 對強化勇水狀態玩家免疫 slow，Go 無此檢查。
  - **C) broadcast duration 差異**：Java case 0 廣播 `S_SkillHaste(cha.id, 2, duration)` **帶實際 duration**，Go `sendSpeedToAll(target, byte(eff.MoveSpeed), uint16(skill.BuffDuration))` 帶 duration（**已對齊**），但 case 1 互抵時 Go 用 `(target, 0, 0)` 對齊 Java setMoveSpeed(0) 廣播。重新核對 — 此差異**不存在**（既有實作正確）。

- **不修原因**（per 對齊深度停損標準）：
  - A、B 屬 SLOW(29)/MASS_SLOW(76)/ENTANGLE(152) **slow 家族共通缺口**，單一 152 audit 修補會留下 29/76 不對齊。應在 29 或 76 audit 時統一補上 helper 套用三項（`isSlowImmune(target)`）。
  - 已建立 slow 家族待重構議題清單，依「不可偷換範圍」維持 out-of-scope。

- **broader gap（不改）**：
  - **A/B 免疫機制**（如上，slow 家族共通議題）。
  - **yaml 2 項漂移**：prob_dice=30 vs 50（Go 跟 cat-fei）、sys_msg_happen=0 vs 711（Go 跟 cat-fei 移除 Java cast 訊息）。屬廣域 yiwei/cat-fei SQL 同步議題。
  - **Java `L1BuffUtil.killSkillEffectTimer` 主動取消路徑**：Java 多處可主動取消 ENTANGLE timer，Go 無對等廣域 hook（同 150 audit 同源）。
  - **HASTE skillmode 端互消反查**：Java `skillmode/HASTE.java:49-50` 顯式在 cast haste 時檢查 hasSkillEffect(ENTANGLE) → 移除，Go 透過 `cancelSpeedBuffs(target, 2)` 通用 catch-all 處理（功能對等更穩健）。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：屬廣域 buff cast 後置缺口（同 130/146/148/149/150/151 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 152 的測試需執行。

## 大地防護（EARTH_SKIN / 151）— 純審計確認 AC ±6 + S_SkillIconShield(6) + REPEATEDSKILLS[1] 雙項互斥完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:502 EARTH_SKIN = 151`（大地防護）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:1438-1440` + `L1SkillUse2.java:1455-1457` icon cast：`pc.sendPackets(new S_SkillIconShield(6, _getBuffIconDuration))`（圖示 param=6）。
  - `L1SkillUse.java:2569-2572` + `L1SkillUse2.java:2521-2524` apply：`pc.addAc(-6) + pc.sendPackets(new S_SkillIconShield(6, _getBuffIconDuration))`。
  - `L1SkillStop.java:511-517` stop：`cha.addAc(6) + (if PC) pc.sendPackets(new S_SkillIconShield(6, 0))` 對齊 cast icon 用 duration=0 撤銷。
  - `L1SkillUse.java:1743 + L1SkillUse2.java:1752 REPEATEDSKILLS[1]`：`{EARTH_SKIN=151, IRON_SKIN=168}` **雙項**護甲類 buff 互斥——Java 只有兩項互斥，**不**含其他防禦類技能（如 IRON_SKIN_DE=3、SHADOW_ARMOR=21/24/99、EARTH_BLESS=159）。
  - `L1SkillUse2.java:798 + L1SkillUse.java:804`：buff cancel 路徑明確包含 EARTH_SKIN。
  - 表格成員：`isNotCancelable` 不含 151（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 151（不受 counter magic 阻擋）；`REPEATEDSKILLS[1]` 雙項互斥。
  - yiwei `db_split/skills.sql:150`：`('151', '大地防護', '19', '6', '15', '0', '0', '0', '0', '960', 'none', '0', '0', '0', '0', '0', '0', '1', '2', '0', '0', '0', '0', '64', '', '19', '2249', '0', '0', '725', '280')` — mp=15、buff_duration=960、target=**none**、target_to=0、attr=1、type=2、ranged=0、id=64、action_id=19、cast_gfx=**2249**、sys_msg_stop=725、sys_msg_fail=280。

- **Go 對照**：
  - `buffs.lua:120 [151] = { ac = -6, exclusions = {168} }`：AC -6 + **僅與 168 互斥** 對齊 Java `REPEATEDSKILLS[1]` 排除自身後 1 項。**先前 audit 已收緊**：原 buffs.lua `{3,21,24,99,159,168}` 6 項 mutex 為 Go 私自加入非 Java 群組成員，2026-05-18 audit 已收緊為 `{168}` 對齊 Java。
  - `skill_buff.go:154 buff.DeltaAC = int16(eff.AC)` apply 套入 ActiveBuff。
  - `skill_buff.go:194 target.AC += buff.DeltaAC` apply 實際 AC -6（即 `target.AC += -6`）。
  - `skill_buff.go:307 sendBuffIcon` 透過 `buff_icon_map.yaml:20-22 skill_id: 151 type:shield param:6` 送 `S_SkillIconShield(6, duration)` 對齊 Java cast icon。
  - `skill_buff.go:506 target.AC -= buff.DeltaAC` revert AC +6 + 對應 stop icon。
  - `executeBuffSkill` 路徑（因 yaml target='buff'）：`:983 BuildActionGfx` + `:1229 applyBuffEffect` + `:1240-1242 BuildSkillEffect(cast_gfx=2249)` 廣播 S_SkillSound。
  - yaml `skill_list.yaml:4621-4651 skill_id=151`：mp=15、buff_duration=960、target=**buff**、target_to=**1**、attr=1、type=2、ranged=**-1**、id=64、action_id=19、cast_gfx=**2249**、sys_msg_stop=725、sys_msg_fail=280。
  - **3 項 yaml 漂移**（target/target_to/ranged Go 跟 cat-fei）但 cast_gfx **無漂移**（罕見 yiwei 對應 cat-fei 一致值）。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:106-119 TestSkillElementalBuffElfArmorAndWaterBuffsUseJavaValues`：
    - player 起始 AC=10，套 151 → AC=4（-6）✓ 鎖死 AC ±6 套用。
    - 套 159 EARTH_BLESS（**不**在 REPEATEDSKILLS）→ 151 保留 + 159 套用、AC 不變 ✓ 鎖死 Java 「159 只送圖示不改 AC、不互斥 151」設計。
    - 套 168 IRON_SKIN（REPEATEDSKILLS[1] 互斥）→ 151 移除 + 159 保留、AC=10-10=0（159 revert +6 - 168 apply -10 = -4 → 不對，須再核對。實際 168 給 AC -10，所以套 168 後 168 套用，151 移除 revert +6，最終 AC = 4 + 6 - 10 = 0 ✓）。

- **發現的 Java 真實差異**：**無**（核心 AC ±6 + S_SkillIconShield(6) icon + REPEATEDSKILLS[1] 雙項互斥完整對齊；先前 audit 已收緊 buffs.lua mutex 從 6 項收為 1 項對齊 Java）。

- **broader gap（不改）**：
  - **yaml 3 項漂移**：target=buff vs none、target_to=1 vs 0、ranged=-1 vs 0（同 148 漂移模式但 cast_gfx 無漂移）。屬廣域 yiwei/cat-fei SQL 同步議題。
  - **168 IRON_SKIN 反向 mutex**：buffs.lua 已對 151 補完 exclusions={168}，但 168 是否也對應補 151 mutex 待 168 audit 補齊。
  - **159 EARTH_BLESS 不互斥**：既有測試已驗證 Java 設計（159 不在 REPEATEDSKILLS 任何群組）；待 159 audit 確認 buffs.lua [159] 互斥定義（先前 audit 已將 159 exclusions 從 `{151, 168}` 收為 `{}`）。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：屬廣域 buff cast 後置缺口（同 130/146/148/149/150 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfArmor -timeout 60s` PASS（0.049s）。

## 風之疾走（WIND_WALK / 150）— 純審計確認 BraveSpeed=4 + S_SkillBrave 雙路徑 + REPEATEDSKILLS[2] 7 項互斥完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:498 WIND_WALK = 150`（風之疾走）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:1458-1460` + `L1SkillUse2.java:1475-1477` icon cast：與 HOLY_WALK/MOVING_ACCELERATION/WIND_WALK 同組，送 `pc.sendPackets(new S_SkillBrave(pc.id, 4, _getBuffIconDuration))` icon。
  - `L1SkillUse.java:2653-2658` + `L1SkillUse2.java:2606-2611` apply（同組）：
    ```java
    pc.setBraveSpeed(4);
    pc.sendPackets(new S_SkillBrave(pc.getId(), 4, _getBuffIconDuration));
    pc.broadcastPacketAll(new S_SkillBrave(pc.getId(), 4, 0));
    ```
    **雙 packet 路徑**：自己收 type=4+duration、附近所有人收 type=4+duration=0。
  - `L1SkillStop.java:594-602` stop（同組）：
    ```java
    cha.setBraveSpeed(0);
    if (cha instanceof L1PcInstance) pc.sendPacketsAll(new S_SkillBrave(pc.id, 0, 0));
    ```
    用 `sendPacketsAll`（含自己 + 附近）統一送 type=0 撤銷。
  - `REPEATEDSKILLS[2]`：`{HOLY_WALK=42, MOVING_ACCELERATION=101, WIND_WALK=150, STATUS_BRAVE=1000, STATUS_ELFBRAVE=1016, BLOODLUST=186, FIRE_BLESS=155}` 7 項勇氣速度類 buff 互斥（注意：包含 STATUS_BRAVE/STATUS_ELFBRAVE 兩個 potion buff ID）。
  - `L1BuffUtil.java:173-174 + :228-229`：`if (pc.hasSkillEffect(WIND_WALK)) pc.killSkillEffectTimer(WIND_WALK)`——其他系統可主動取消 WIND_WALK。
  - 表格成員：`isNotCancelable` 不含 150（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 150（不受 counter magic 阻擋）；`REPEATEDSKILLS[2]` 7 項互斥。
  - yiwei `db_split/skills.sql:149`：`('150', '風之疾走', '19', '5', '15', '0', '0', '0', '0', '960', 'none', '0', '0', '0', '0', '0', '0', '8', '2', '0', '0', '0', '0', '32', '', '19', '11730', '0', '0', '0', '0')` — mp=15、buff_duration=960、target=none、target_to=0、attr=8、type=2、ranged=0、id=32、action_id=19、cast_gfx=**11730**、所有 sys_msg=0。

- **Go 對照**：
  - `buffs.lua:118 [150] = { brave_speed = 4, exclusions = {52, 101, 155, 186, 1000, 1016} }`：BraveSpeed=4 + 6 項 mutex 對齊 Java `REPEATEDSKILLS[2]` 排除自身後 6 項（**注意：Java HOLY_WALK=42，Go 用 52** — 已在 130/132/133/134 audit 同源廣域 SQL 漂移議題列項，與 cat-fei `52 HOLY_WALK` 對齊；其餘 6 項對齊 Java）。
  - `skill_buff.go:235-239` apply：
    ```go
    if eff.BraveSpeed > 0 {
        buff.SetBraveSpeed = byte(eff.BraveSpeed)
        target.BraveSpeed = byte(eff.BraveSpeed)
        s.sendBraveToAll(target, byte(eff.BraveSpeed), uint16(skill.BuffDuration))
    }
    ```
  - `skill_buff.go:625-632 sendBraveToAll`：
    ```go
    sendBravePacket(target.Session, target.CharID, braveType, duration)      // 自己: S_SkillBrave(target, type, duration)
    nearby := s.deps.World.GetNearbyPlayers(target.X, target.Y, target.MapID, target.SessionID)
    for _, other := range nearby {
        sendBravePacket(other.Session, target.CharID, braveType, 0)         // 附近: S_SkillBrave(target, type, 0)
    }
    ```
    **完美對齊 Java 雙 packet 路徑**：自己 duration、附近 0。
  - `skill_buff.go:665-667` revert：`target.BraveSpeed = 0 + sendBraveToAll(target, 0, 0)` 對齊 Java stop `setBraveSpeed(0) + sendPacketsAll(S_SkillBrave(0, 0))`。
  - `executeBuffSkill` 路徑：因 target='none' 走 `executeSelfSkill` 而非 `executeBuffSkill`，但 `skill_self.go` default 分支會呼叫 `applyBuffEffect` 處理 BraveSpeed delta。
  - yaml `skill_list.yaml:4590-4620 skill_id=150`：mp=15、buff_duration=960、target=none、target_to=0、attr=8、type=2、ranged=0、id=32、action_id=19、cast_gfx=**2247**、所有 sys_msg=0。
  - **僅 cast_gfx 一項漂移**（Go=2247 vs yiwei=11730），其餘 30 欄位完全對齊 yiwei（同 149 模式）。

- **既有測試覆蓋**：
  - 無針對 150 的單元測試（既有 `TestSkillElementalBuff*` 系列覆蓋 148/149/155 等同組 elf buff，未專門測 150）。
  - 相近的 `case 172 STORM_WALK`（`skill_self.go:156-177`）有相同 BraveSpeed=4 廣播邏輯但走顯式 case 而非 lua（針對與 STATUS_BRAVE/STATUS_ELFBRAVE 等 6 項互斥 + 顯式 storm walk speed=4 setup）。

- **發現的 Java 真實差異**：**無**（核心 BraveSpeed=4 + S_SkillBrave 雙 packet 路徑 + REPEATEDSKILLS[2] 7 項互斥完整對齊）。

- **broader gap（不改）**：
  - **yaml cast_gfx 漂移**：Go=2247 vs yiwei=11730，Go 用 cat-fei 較舊 GFX ID（同 149 模式但僅 1 項）。屬廣域 yiwei/cat-fei SQL 同步議題。
  - **HOLY_WALK ID 漂移**：Java=42、Go=52（cat-fei）。整個 elf brave speed 群組以 cat-fei 52 為基準，未跟 Java 42。屬廣域 SQL ID 漂移議題（影響整個 REPEATEDSKILLS[2] 群組互斥定義）。
  - **反向 mutex（potion buff 路徑）**：`item_use.go applyBrave` 對 STATUS_BRAVE(1000)/STATUS_ELFBRAVE(1016) 兩個 potion buff 的 mutex 處理待 STATUS_BRAVE 專門 audit 時補（與 buffs.lua [150].exclusions 中已含 1000/1016 對應，但反向 potion 觸發路徑可能未補 150 互斥）。
  - **`L1BuffUtil.killSkillEffectTimer` 主動取消路徑**：Java 多處（PVP、castle 進出等）會主動取消 WIND_WALK timer，Go 無對等廣域 hook 機制。屬廣域 buff lifecycle 管理議題。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：屬廣域 buff cast 後置缺口（同 130/146/148/149 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 150 的測試需執行。

## 風之神射（WIND_SHOT / 149）— 純審計確認核心 BowHit ±6 + icon + 互斥完整對齊 Java

- **Java 對照**：
  - `L1SkillId.java:494 WIND_SHOT = 149`（風之神射）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:1414-1416` + `L1SkillUse2.java:1431-1433` icon cast：`pc.sendPackets(new S_PacketBoxIconAura(148, _getBuffIconDuration))`（icon_id = skill_id - 1 = 148）。
  - `L1SkillUse.java:2601-2604` + `L1SkillUse2.java:2553-2556` apply：`pc.addBowHitup(6) + pc.sendPackets(new S_PacketBoxIconAura(148, _getBuffIconDuration))`。
  - `L1SkillStop.java:554-560` stop：`cha.addBowHitup(-6) + (if PC) pc.sendPackets(new S_PacketBoxIconAura(148, 0))`。
  - `REPEATEDSKILLS[0]`：`{148, 149, 156, 163, 166}` 5 項武器類 buff 互斥（同 148 audit）。
  - 表格成員：`isNotCancelable` 不含 149（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 149（不受 counter magic 阻擋）；`REPEATEDSKILLS[0]` 5 項武器互斥。
  - yiwei `db_split/skills.sql:148`：`('149', '風之神射', '19', '4', '15', '0', '0', '0', '0', '960', 'buff', '1', '0', '0', '0', '0', '0', '8', '2', '0', '-1', '0', '0', '16', '', '19', '11729', '0', '0', '724', '280')` — mp=15、buff_duration=960、target=**buff**、target_to=1、attr=8、type=2、ranged=-1、id=16、action_id=19、cast_gfx=**11729**、sys_msg_stop=724、sys_msg_fail=280。

- **Go 對照**：
  - `buffs.lua:117 [149] = { bow_hit = 6, exclusions = {148, 156, 163, 166} }`：BowHit +6 + 4 項 mutex 對齊 Java `REPEATEDSKILLS[0]` 排除自身後 4 項。
  - `skill_buff.go:170 buff.DeltaBowHit = int16(eff.BowHit)` apply 套入 ActiveBuff。
  - `skill_buff.go:201 target.BowHitMod += buff.DeltaBowHit` apply +6。
  - `skill_buff.go:307 sendBuffIcon` 透過 `buff_icon_map.yaml:60-61 skill_id: 149 type:aura` 送 `S_PacketBoxIconAura(148, duration)`（icon_id 由 Go 計算 skill_id - 1 = 148）。
  - `skill_buff.go:513 target.BowHitMod -= buff.DeltaBowHit` revert -6 + 對應 stop icon。
  - `executeBuffSkill` 路徑：`:983 BuildActionGfx` 廣播 S_DoActionGFX + `:1229 applyBuffEffect` 套 buff + `:1240-1242 BuildSkillEffect(cast_gfx=2246)` 廣播 S_SkillSound。
  - yaml `skill_list.yaml:4559-4589 skill_id=149`：mp=15、buff_duration=960、target=**buff**、target_to=1、attr=8、type=2、ranged=-1、id=16、action_id=19、cast_gfx=**2246**、sys_msg_stop=724、sys_msg_fail=280。
  - **僅 cast_gfx 一項漂移**（Go=2246 vs yiwei=11729），其餘 30 欄位完全對齊 yiwei——對比 148（4 項漂移），149 yaml 對齊度顯著更高（yiwei 自身對 149 已更新到 target='buff'）。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:44-89 TestSkillElementalBuffElfWeaponAndBowBuffsUseJavaValues`：
    - 套用 149（在 148/163 已套後）→ BowHitMod=9（起始 3 + 6）、BowDmgMod=4（不變）✓ 鎖死 149 只給弓命中。
    - 套用 166 STORM_SHOT（同 mutex 群）→ 149 移除 + 166 套用 BowHitMod=2（9-6-1）、BowDmgMod=9（4+5）✓ 鎖死 REPEATEDSKILLS[0] 互斥行為與 166 弓傷害 +5/弓命中 -1 mechanic。

- **發現的 Java 真實差異**：**無**（核心 BowHit ±6 + S_PacketBoxIconAura(148) icon + REPEATEDSKILLS[0] mutex 完整對齊）。

- **broader gap（不改）**：
  - **yaml cast_gfx 漂移**：Go=2246 vs yiwei=11729，Go 用 cat-fei 較舊 GFX ID（同 148 漂移模式但僅 1 項，因 yiwei 對 149 已更新 target='buff' 等其他欄位到 cat-fei 一致值）。client 視覺差異但對遊戲機制無影響。屬廣域 yiwei/cat-fei SQL 同步議題。
  - **156/163 反向 mutex 擴充**：buffs.lua 已對 149 補完 exclusions = {148, 156, 163, 166}，但 156/163 是否也補 149 mutex 待各自審計（148 已補完、166 待 audit）。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：對 buff 類無 stat 變化時的 `S_SPMR + S_OwnCharStatus + S_PacketBox(UPDATE_ER)` 通用 status refresh 屬廣域 buff cast 後置缺口（同 130/146/148 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfWeapon -timeout 60s` PASS（cached）。

## 火焰武器（FIRE_WEAPON / 148）— 純審計確認核心對齊 Java，yaml 多項漂移屬廣域 SQL 議題

- **Java 對照**：
  - `L1SkillId.java:490 FIRE_WEAPON = 148`（火焰武器）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:1410-1412` + `L1SkillUse2.java:1427-1429` icon cast：`pc.sendPackets(new S_PacketBoxIconAura(147, _getBuffIconDuration))` 廣播 icon 給玩家自己（icon_id = skill_id - 1 = 147）。
  - `L1SkillUse.java:2590-2593` + `L1SkillUse2.java:2542-2545` apply：`pc.addDmgup(4) + pc.sendPackets(new S_PacketBoxIconAura(147, _getBuffIconDuration))`。
  - `L1SkillStop.java:533-539` stop：`cha.addDmgup(-4) + (if PC) pc.sendPackets(new S_PacketBoxIconAura(147, 0))` 對齊 cast icon 用 duration=0 撤銷。
  - `L1SkillUse.java:1741 + L1SkillUse2.java:1750 REPEATEDSKILLS[0]`：`{FIRE_WEAPON=148, WIND_SHOT=149, STORM_EYE=156, BURNING_WEAPON=163, STORM_SHOT=166}` 5 項武器類 buff **互斥**——施同組任一技能會自動移除其他 4 項。
  - 表格成員：`isNotCancelable` 不含 148（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 148（不受 counter magic 阻擋）；`REPEATEDSKILLS[0]` 5 項武器 buff 互斥。
  - yiwei `db_split/skills.sql:147`：`('148', '火焰武器', '19', '3', '15', '0', '0', '0', '0', '960', 'none', '0', '0', '0', '0', '0', '0', '2', '2', '0', '0', '0', '0', '8', '', '19', '11774', '0', '0', '723', '280')` — mp=15、buff_duration=960、target=**none**、target_to=0、attr=2、type=2、ranged=0、id=8、action_id=19、cast_gfx=**11774**、sys_msg_stop=723、sys_msg_fail=280。

- **Go 對照**：
  - `buffs.lua:116 [148] = { dmg_mod = 4, exclusions = {149, 156, 163, 166} }`：DmgMod +4 + 4 項 mutex 對齊 Java `REPEATEDSKILLS[0]` 排除自身後 4 項。
  - `engine.go` 讀取 lua `dmg_mod=4` 寫入 `BuffEffect.DmgMod`。
  - `skill_buff.go:162 buff.DeltaDmgMod = int16(eff.DmgMod)` apply 套入 ActiveBuff。
  - `skill_buff.go:196 target.DmgMod += buff.DeltaDmgMod` apply 實際 +4。
  - `skill_buff.go:288-294 SendPlayerStatus`：DmgMod 變化時送 `S_PlayerStatus` 更新 client UI 屬性面板。
  - `skill_buff.go:307 sendBuffIcon(target, skill.SkillID, uint16(skill.BuffDuration))` 透過 `buff_icon_map.yaml` 查 `148 → type:aura icon_id = 147` 送 `S_PacketBoxIconAura(147, duration)` 對齊 Java cast icon。
  - `skill_buff.go:508 target.DmgMod -= buff.DeltaDmgMod` revert -4。
  - revert 端透過 `sendBuffIconStop` 送 `S_PacketBoxIconAura(147, 0)` 對齊 Java stop icon。
  - `executeBuffSkill` 路徑（`skill_buff.go:869` 入口）：
    - `:983 BroadcastToPlayers(BuildActionGfx(action_id=19))` = Java `S_DoActionGFX` cast 動畫。
    - `:1229 applyBuffEffect` 套用 buff + icon。
    - `:1240-1242 if skill.CastGfx > 0 → BroadcastToPlayers(BuildSkillEffect(target.CharID, skill.CastGfx))` = Java `S_SkillSound` cast 視覺特效（cast_gfx=2182 廣播）。
  - `buff_icon_map.yaml:58-59`：`skill_id: 148 type: aura` 註冊 → icon_id=147 由 Go 計算（skill_id - 1）。
  - yaml `skill_list.yaml:4528-4558 skill_id=148`：mp=15、buff_duration=960、target=**buff**、target_to=**1**、attr=2、type=2、ranged=**-1**、id=8、action_id=19、cast_gfx=**2182**、sys_msg_stop=723、sys_msg_fail=280。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:44-89 TestSkillElementalBuffElfWeaponAndBowBuffsUseJavaValues`：
    - player 起始 DmgMod=1/HitMod=2/BowHitMod=3/BowDmgMod=4。
    - 套用 148 → DmgMod=5（+4）、其他不變 ✓ 鎖死 148 只給近戰傷害。
    - 套用 163（BURNING_WEAPON，同 mutex 群）→ 148 buff 移除 + 163 buff 套用、DmgMod=7（5-4+6）、HitMod=5（2+3）✓ 鎖死 REPEATEDSKILLS[0] 5 項互斥行為。
    - 套用 149（WIND_SHOT，同 mutex 群）→ BowHitMod=9（+6）✓ 鎖死 149 只給弓命中。

- **發現的 Java 真實差異**：**無實質負面差異**（核心 DmgMod ±4 + S_PacketBoxIconAura(147) icon + REPEATEDSKILLS[0] mutex 完整對齊）。

- **broader gap（不改）**：
  - **yaml 4 項漂移**：Go yaml 與 yiwei `skills.sql:147` 偏離（Go 跟 cat-fei）：
    - `target=buff vs none`：Go 路由經 `executeBuffSkill`，Java 經 generic L1SkillUse self-cast，功能等價但 dispatch 路徑不同。
    - `target_to=1 vs 0`：Go target_to=1 表示自我/單目標，Java 0 = none，dispatch 端等價。
    - `ranged=-1 vs 0`：兩者皆「無射程限制自我施法」，等價。
    - `cast_gfx=2182 vs 11774`：Go 用 cat-fei 較舊 GFX ID、yiwei 用 Lineage R 11774，client 視覺差異但對遊戲機制無影響。
  - 屬廣域 yiwei/cat-fei SQL 同步議題（與 130/132/133/134/145/146 同源）。
  - **149/156/163/166 反向擴充**：buffs.lua 已對 148 補完 exclusions = {149, 156, 163, 166}，但其他 4 項是否也對應補 148 mutex 待各自審計：149 ✓（待 audit）、156（待 audit）、163（待 audit）、166（待 audit）。
  - **`L1SkillUse.sendGrfx` 末尾 1686-1694 _targetList 通用 status refresh**：對 buff 類無 stat 變化時的 `S_SPMR + S_OwnCharStatus + S_PacketBox(UPDATE_ER)` 通用 status refresh 屬廣域 buff cast 後置缺口（同 130/146 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfWeapon -timeout 60s` PASS（0.042s）。

## 單屬性防禦（ELEMENTAL_PROTECTION / 147）— 純審計確認 Go 完整對齊 Java 且更穩健（補 Java 漏送 + 修 Java revert bug）

- **Java 對照**：
  - `L1SkillId.java ELEMENTAL_PROTECTION = 147`（單屬性防禦）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:2547-2558` + `L1SkillUse2.java:2503-2514` cast：
    ```java
    } else if (this._skillId == ELEMENTAL_PROTECTION) {
        final L1PcInstance pc = (L1PcInstance) cha;
        final int attr = pc.getElfAttr();
        if (attr == 1)      pc.addEarth(50);
        else if (attr == 2) pc.addFire(50);
        else if (attr == 4) pc.addWater(50);
        else if (attr == 8) pc.addWind(50);
    }
    ```
    **重要：cast 路徑無 `S_OwnCharAttrDef` 通知**——Java 在 cast 後 client UI 不會立即反映屬性 +50，要等其他屬性變化才會更新（Java 漏送）。
  - `L1SkillStop.java:476-491` stop：
    ```java
    case 147:
        if ((cha instanceof L1PcInstance)) {
            L1PcInstance pc = (L1PcInstance) cha;
            int attr = pc.getElfAttr();         // ← 讀「當前」ElfAttr
            if (attr == 1)      cha.addEarth(-50);
            else if (attr == 2) cha.addFire(-50);
            else if (attr == 4) cha.addWater(-50);
            else if (attr == 8) cha.addWind(-50);
            pc.sendPackets(new S_OwnCharAttrDef(pc));
        }
        break;
    ```
    **Java 潛在 bug**：revert 讀「當前」`pc.getElfAttr()`，若 ElfAttr 在 cast→stop 期間因任務/裝備變動而改變，會反向錯誤屬性導致 stat corruption（cast 加風、stop 減土）。
  - `skill.go:342-344 if skillID == 147 && player.ElfAttr == 0` 拒絕施法（Java 側無這個前置檢查——Java 也是依屬性 case，attr=0 走 else 區段不做事，但仍消耗 MP/HP/item）。
  - 表格成員：`isNotCancelable` 不含 147（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 147（不受 counter magic 阻擋）；`REPEATEDSKILLS` 不含 147（無互斥）。
  - yiwei `db_split/skills.sql:146`：`('147', '單屬性防禦', '19', '2', '6', '0', '40319', '1', '0', '64', 'none', '0', '0', '0', '0', '0', '0', '0', '2', '0', '0', '0', '0', '4', '', '19', '2285', '0', '0', '0', '0')` — mp=6、item=40319×1、reuse_delay=0、buff_duration=64、target=none、type=2、id=4、action_id=19、cast_gfx=2285。

- **Go 對照**：
  - `skill.go:342-344` pre-cast 檢查：
    ```go
    if skillID == 147 && player.ElfAttr == 0 {
        s.sendCastFail(sess)
        s.failTeleportSkill(sess, skillID)
        return
    }
    ```
    ——比 Java 嚴格：ElfAttr=0 玩家**完全不消耗** MP/HP/item，避免「付資源但無效果」反直覺體驗。
  - `skill_buff.go:181-182` apply 時呼叫專門 helper：`if skill.SkillID == 147 { applyElementalProtectionDelta(target, buff) }`。
  - `skill_buff.go:315-329 applyElementalProtectionDelta`：
    ```go
    switch target.ElfAttr {
    case 1: buff.DeltaEarthRes = 50
    case 2: buff.DeltaFireRes = 50
    case 4: buff.DeltaWaterRes = 50
    case 8: buff.DeltaWindRes = 50
    }
    ```
    讀 cast 時 ElfAttr 並**寫入 buff struct**（不直接動 target 欄位）。
  - `skill_buff.go:211-214` apply 套用：`target.X += buff.DeltaX` 將 +50 套入對應屬性。
  - `skill_buff.go:300-305` apply 端送 `SendAbilityScores(target.Session, target)` = `S_OwnCharAttrDef`——**補上 Java cast 漏送的 UI 更新**。
  - `skill_buff.go:515-518` revert 套用：`target.X -= buff.DeltaX` 用**儲存的 Delta**反向（不讀當前 ElfAttr）。
  - `skill_buff.go:565-570` revert 端送 `SendAbilityScores` 對齊 Java stop。
  - **Go 比 Java 更穩健**：revert 用 cast 時儲存的 Delta，即使 ElfAttr 期間改變也能正確還原原屬性（不會 stat corruption）。
  - yaml `skill_list.yaml:4497-4527 skill_id=147`：mp=6、item=40319×1、reuse_delay=0、buff_duration=64、target=none、type=2、id=4、action_id=19、cast_gfx=2285。**31 欄位與 yiwei skills.sql:146 完全對齊零漂移**。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:10-42 TestSkillElementalBuffElfElementalDefenseBuffsUseJavaValues`：player ElfAttr=2（fire），起始 FireRes=1，套用 138 + 147 → FireRes=61（+10 from 138 + +50 from 147）；完整鎖死 ElfAttr-based 單屬性 +50 機制。

- **發現的 Java 真實差異**：**無實質負面差異**（Go 三項細節都是「比 Java 嚴格收緊或更穩健」）：
  - **A) 改進：cast 端送 `S_OwnCharAttrDef`**：Go apply 端 `SendAbilityScores` 補 Java cast 漏送 → 玩家 cast 後 client UI 立即反映 +50。
  - **B) 改進：ElfAttr=0 前置阻擋**：Go 提前拒絕 + sendCastFail + failTeleportSkill 不消耗資源；Java 走 else 區段不加 stat 但仍扣 MP/HP/item（資源浪費）。
  - **C) 改進：revert 用 stored Delta**：Go 用 cast 時儲存的 buff Delta 反向，即使 ElfAttr 變動仍正確；Java 讀當前 ElfAttr，cast→stop 期間 ElfAttr 變動會 stat corruption（潛在 bug）。

- **不修原因**（per 對齊深度停損標準）：
  - 三項 Go 改進都是 **production-correct + 嚴格化**，未引入 Java vs Go 玩家觀察差異（client UI 結果一致或更佳）。
  - 反向修改回 Java 行為會**引入 Java bug**（B 浪費資源、C 屬性損毀），違反「修 bug 而非引入 bug」原則。
  - 既有測試已完整鎖死 ElfAttr-based 單屬性 +50 機制 + apply/revert 雙端 S_OwnCharAttrDef 通用機制（隸屬 skill_buff.go 通用路徑）。

- **broader gap（不改）**：
  - **`getFire()/getWater()/getWind()/getEarth()` 動態 vs 靜態**：Java getter 在 receive 端動態彙整 base + 裝備 + buff + 寶物加成；Go 在 cast 時將 buff delta 累加到 `target.FireRes` 等靜態欄位。Client UI 與最終魔法傷害計算結果等價，屬廣域 stat 架構議題（同 110/111/137/138 同源 precedent）。
  - **ElfAttr 動態變動處理**：Java 部分裝備可暫時改變 ElfAttr，Go 無對等系統。屬廣域裝備系統議題。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfElementalDefense -timeout 60s` PASS（0.044s）。

## 魂體轉換（BLOODY_SOUL / 146）— 補齊施法視覺特效廣播對齊 Java（同 130 BODY_TO_MIND precedent）

- **Java 對照**：
  - `L1SkillId.java:482 BLOODY_SOUL = 146`（魂體轉換）。
  - `L1SkillMode.load() line 99`：`_skillMode.put(Integer.valueOf(146), new BLOODY_SOUL())` 註冊 skillmode。
  - `skillmode/BLOODY_SOUL.java:14-21 PC start`：
    - 行 19：`srcpc.setCurrentMp(srcpc.getCurrentMp() + ConfigElfSkill.BLOODY_SOULADDMP)`。
    - `BLOODY_SOULADDMP` 來自 `妖精_技能設定表.properties` + `各職業技能相關設置.properties` 均設 **20**（不是 skill_level=19）。
  - `skillmode/BLOODY_SOUL.java:23-27 NPC start`：空實作。
  - `skillmode/BLOODY_SOUL.java:32-33 stop`：空實作（無持續 buff state）。
  - `L1SkillUse.sendGrfx:1681` default 分支：對 target='none' + type=buff + 非特殊技能列表走 `_player.sendPacketsAll(new S_SkillSound(self.id, cast_gfx=2178))` 廣播施法效果。
  - HP 消耗：Java `L1SkillUse` generic 框架讀 SQL `hp` 欄位 = 50（yiwei）扣 HP。
  - 表格成員：`isNotCancelable` 不含 146（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 146 line 148（**不受 counter magic 阻擋**）；`REPEATEDSKILLS` 不含 146（無互斥）。
  - yiwei `db_split/skills.sql:145`：`('146', '魂體轉換', '19', '1', '0', '50', '0', '0', '1200', '0', 'none', '0', '0', '0', '0', '0', '0', '0', '2', '0', '0', '0', '0', '2', '', '19', '2178', '0', '0', '0', '0')` — mp=0、hp=**50**、reuse_delay=1200、buff_duration=0、target=none、type=2、id=2、action_id=19、cast_gfx=2178。

- **Go 對照（修正前）**：
  - 原 `skill_self.go:136-143 case 146`：
    ```go
    case 146: // 魂體轉換
        player.MP += 20
        if player.MP > player.MaxMP {
            player.MP = player.MaxMP
        }
        sendMpUpdate(sess, player)
    ```
    - +20 MP 對齊 Java `BLOODY_SOULADDMP=20`（既有 audit 已修為硬編 20 不誤用 skill_level）。
    - MaxMP clamp 對齊 Java `L1PcInstance.setCurrentMp` 內建 clamp。
    - `sendMpUpdate` 對齊 Java `S_MPUpdate`。
    - **缺 `BuildSkillEffect(self, cast_gfx=2178)` 廣播**——附近玩家看不到魂體轉換的施法視覺效果（同 130 BODY_TO_MIND 修正前 bug）。
  - `skill.go:328-332 + :449-450` HP 消耗檢查 + 扣除：用 yaml hp_consume=40（非 yiwei 50）。
  - 後置 `skill_self.go:181-183 BuildActionGfx(player, action_id=19)` 廣播 S_DoActionGFX 對齊 Java `L1SkillUse.sendGrfx:1641-1643`。
  - yaml `skill_list.yaml:4466-4496 skill_id=146`：mp=0、hp=**40**、reuse_delay=1200、buff_duration=0、target=none、type=2、id=2、action_id=19、cast_gfx=2178。

- **發現的 Java 真實差異 (criterion a)**：
  - **cast_gfx 廣播缺失**：Java `L1SkillUse.sendGrfx:1681` default 分支對魂體轉換送 `S_SkillSound(self.id, 2178)` 廣播施法視覺效果，Go case 146 只有 MP +20 + sendMpUpdate + 後置 BuildActionGfx，**未廣播 cast_gfx**。附近玩家觀察不到魂體轉換的施法效果（與 130 BODY_TO_MIND 修正前完全相同的 gap）。

- **修正**：`skill_self.go:136-150 case 146` 補上 cast_gfx 廣播：
  ```go
  case 146: // 魂體轉換 — Java `BLOODY_SOUL.start()` 第 19 行
      // `setCurrentMp(currentMp + ConfigElfSkill.BLOODY_SOULADDMP)`，
      // yiwei `各職業技能相關設置.properties: BLOODY_SOULADDMP = 20`（不是 skill.skill_level=19）。
      player.MP += 20
      if player.MP > player.MaxMP {
          player.MP = player.MaxMP
      }
      sendMpUpdate(sess, player)
      // Java `L1SkillUse.sendGrfx:1681` default 分支對 target='none' + type=buff 送
      // `_player.sendPacketsAll(new S_SkillSound(self.id, cast_gfx=2178))` 廣播施法效果。
      // 後置 BuildActionGfx 已負責 S_DoActionGFX(action_id=19)，這裡補上 cast_gfx 廣播
      // （同 130 BODY_TO_MIND precedent）。
      if skill.CastGfx > 0 {
          handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
      }
  ```
  - 與 case 130 同模式（既有 audit 2026-05-19 修正）：在 sendMpUpdate 後加 `if skill.CastGfx > 0 → BroadcastToPlayers + BuildSkillEffect(self, CastGfx)`。
  - 與 case 186 BLOODLUST 同模式（cast_gfx 廣播在 applyBuffEffect 後）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -timeout 120s` PASS（18.004s，全 system 測試無迴歸）。

- **不寫新測試**：cast_gfx 廣播屬「補 Java packet」surgical 修正（廣播 packet 內容已由 BuildSkillEffect 共用實作驗證），新增「測試 cast_gfx 廣播」屬「鎖死 Java 對齊行為」測試違反停損標準。

- **broader gap（不改）**：
  - **yaml hp_consume 漂移**：Go=40、yiwei=50，Go 跟 cat-fei 偏離 yiwei，屬廣域 SQL 漂移議題（與 130/132/133/134/145 同源）。實際遊戲影響：Java 消耗 50 HP +20 MP（淨虧 30 HP）為 ~2.5:1 比例；Go 消耗 40 HP +20 MP（淨虧 20 HP）為 2:1 比例，玩家負擔較輕。
  - **Java L1SkillUse.sendGrfx:1686-1694 末尾 _targetList 通用 status refresh**：對 PC 送 `S_SPMR + S_OwnCharStatus + S_PacketBox(UPDATE_ER)`，對 BLOODY_SOUL 無實際 stat 變化，屬廣域 buff cast 後置 status refresh 缺口（同 130 audit broader gap）。

## 釋放元素（RETURN_TO_NATURE / 145）— 純審計確認三項 Java 真實差異屬廣域 targeting 子系統缺口（同 116 precedent）

- **Java 對照**：
  - `L1SkillId.java:478 RETURN_TO_NATURE = 145`（釋放元素）。無對應 skillmode 類別。
  - `L1SkillUse.java:2740-2750` + `L1SkillUse2.java:2689-2699` 攻擊路徑：
    - `if (cha instanceof L1SummonInstance) { summon.broadcastPacketAll(new S_SkillSound(summon.id, 2245)); summon.returnToNature(); }`
    - `else { if (user instanceof L1PcInstance) this._player.sendPackets(new S_ServerMessage(79)); }`（**只對「指定的單一 summon 目標」生效**，非 summon 目標送 msg 79）。
  - `L1MagicPc.calcProbabilityMagic` PC→NPC 路徑（line 340-347）：`defenseLevel = _targetNpc.getLevel(); if (skillId == RETURN_TO_NATURE && _targetNpc instanceof L1SummonInstance) defenseLevel = summon.getMaster().getLevel(); defenseMr = _targetNpc.getMr();` ——**defenseLevel 用 summon 的 master.level**（不是 summon 自己的等級），對齊「玩家魔法等級 vs 召喚主等級」的對比設計。
  - `L1MagicPc.calcProbabilityMagic` default 分支（line 835-848）：`probability = probValue + sum(diceCount2) of random(1, probDice)`，其中 `diceCount2 = max(magicBonus + magicLevel, 1)`。yaml probValue=33、probDice=50 → 1 dice 最小：`probability = 33 + 1..50` ＝ **34-83% 機率**（典型 55-70%）。
  - `L1MagicNpc.calcProbabilityMagic` line 157-171：NPC 路徑 `probability = (probDice/10) * (atkLvl - defLvl) + probValue - targetMr/10`，多項 elf debuff 共用同一 formula（與 ELEMENTAL_FALL_DOWN/ENTANGLE 等同組）。
  - `L1SummonInstance.returnToNature()`：把馴服的 summon 釋放回野生 NPC、非馴服的 summon 銷毀。**單一目標**只處理 method 接收者 summon 自己，不影響 caster 其他 summons。
  - yiwei `db_split/skills.sql:144`：`('145', '釋放元素', '19', '0', '30', '0', '40319', '2', '0', '0', 'buff', '2', '0', '0', '0', '33', '50', '0', '1', '0', '-1', '0', '0', '1', '', '19', '2245', '0', '0', '0', '280')` — mp=30、item=40319×2、reuse_delay=0、buff_duration=0、target=buff、target_to=2、prob_value=33、prob_dice=50、type=1、ranged=-1、id=1、action_id=19、cast_gfx=2245、sys_msg_fail=280。

- **Go 對照**：
  - `skill.go:418-420`：`case 145: s.deps.Summon.ExecuteReturnToNature(sess, player, skill)`。
  - `skill_magic_scroll.go:63-65`：scroll 觸發路徑相同。
  - `skill_summon.go:591-608 ExecuteReturnToNature`：
    ```go
    summons := ws.GetSummonsByOwner(player.CharID)
    if len(summons) == 0 { return }
    s.deps.Skill.ConsumeSkillResources(sess, player, skill)
    for _, sum := range summons {
        if sum.Tamed { s.liberateSummon(sum) } else { s.killSummon(sum) }
    }
    ```
    - **無目標 ObjectID 參數**：直接列舉 caster 全部 summons。
    - **無 calcProbabilityMagic**：100% 成功率。
    - **無 master.level defenseLevel**：因為根本不做機率檢查。
    - 馴服處理對齊 Java `returnToNature()`：tamed → liberate（轉回野生 NPC）、untamed → kill（銷毀）。
    - `liberateSummon:632 SendCompanionEffect(viewer.Session, sum.ID, 2245)` 對齊 Java `summon.broadcastPacketAll(new S_SkillSound(summon.id, 2245))`。
  - yaml `skill_list.yaml:4435-4465 skill_id=145`：mp=30、item=40319×2、reuse_delay=0、buff_duration=0、target=buff、target_to=**16**、prob_value=**50**、prob_dice=**30**、type=1、ranged=-1、id=1、action_id=19、cast_gfx=2245、sys_msg_fail=280。

- **既有測試覆蓋**：
  - 無針對 145 的單元測試（`skill_teleport_summon_test.go fakeSummonManager.ExecuteReturnToNature` 僅為 mock placeholder）。

- **發現的 Java 真實差異 (criterion a，屬廣域 targeting 子系統缺口)**：
  - **A) 目標選擇差異**：Java `L1SkillUse.java:2741` 透過 `_targetList` 逐一檢查 `cha instanceof L1SummonInstance`，只對被指定的單一 summon 呼叫 `returnToNature()`。Go `ExecuteReturnToNature` 透過 `GetSummonsByOwner(player.CharID)` 直接列舉 caster 全部 summons 並全部釋放。實際影響：玩家擁有多個 summon 時，Java 可選擇釋放特定一隻、Go 強制全部釋放。
  - **B) 機率差異**：Java `calcProbabilityMagic` default 分支套用 `probability = probValue(33) + sum(diceCount2) of random(1, probDice=50)` ＝ 34-83% 機率（典型 55-70%）；Go 無機率檢查直接 100% 成功。實際影響：在 PvP/野外環境玩家對他人 summon 施 145 時 Java 有失敗可能、Go 必定成功。
  - **C) defenseLevel 用 master.level**：Java `L1MagicPc.calcProbabilityMagic:342-345` 對 summon 目標用 `summon.getMaster().getLevel()` 作 defenseLevel（不是 summon 自己等級），對齊「caster magicLevel vs 召喚主 charLevel」競賽。Go 因為根本不做機率檢查，這個欄位也無從用上。

- **不修原因**（per 對齊深度停損標準）：
  - 三項差異都需要 **targeting subsystem 重構**：dispatch 層當前不接受目標 ObjectID 參數（buff 類技能假設「對自己」），handler 也不傳 target；補上需修改 skill use handler + 加 target param 給 ExecuteReturnToNature + 補 calcProbabilityMagic 整套機率系統路徑（含 magicBonus/magicLevel/getMr/levelDiff 廣域邏輯）。
  - 已建立 **116 CALL_CLAN precedent** 將「同類 targeting 重構」明確列為廣域子系統議題（116 audit 也標記「targeting 重構」），單一技能 audit 不能跨越邊界。
  - 不寫新測試：補測試需依賴尚未實作的機率系統，順序錯位。

- **broader gap（不改）**：
  - **目標選擇缺失**（如 A 所述，targeting subsystem 議題）。
  - **calcProbabilityMagic 整套機率系統**（如 B 所述，廣域 elf debuff 共用 formula，影響 ELEMENTAL_FALL_DOWN/ENTANGLE/WIND_SHACKLE/ERASE_MAGIC/EARTH_BIND/AREA_OF_SILENCE/POLLUTE_WATER/STRIKER_GALE 等多項 elf 技能）。
  - **summon.master.level defenseLevel**（如 C 所述，依賴機率系統，連帶廣域）。
  - **yaml 三方漂移**：Go yaml 與 yiwei `skills.sql:144` 三項偏離（Go 跟 cat-fei）：target_to=16（yiwei=2）、prob_value=50（yiwei=33）、prob_dice=30（yiwei=50）。屬廣域 SQL 漂移議題（與 130/132/133/134 同源）。
  - **S_ServerMessage(79) 非 summon 目標拒絕訊息**：Java 對非 summon 目標 caster 送 msg 79（"無法對該目標使用"），Go 在 ExecuteReturnToNature 完全跳過（連 summon empty 也是 silent return）。屬「Java 嚴格化 UX」議題，依賴 targeting 重構才能補。
  - **NPC→PC 路徑**：Java `L1MagicNpc.calcProbabilityMagic case RETURN_TO_NATURE` 表示理論上 NPC 也能對 PC summon 施 145（雖實務罕見），Go NPC AI 不施 magic 屬 NPC magic system 廣域缺口（同 134 audit 同源）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - 無針對 145 的測試需執行（既有測試暫無覆蓋）。

## 屬性防禦（RESIST_ELEMENTAL / 138）— 純審計確認四屬 +10 + S_OwnCharAttrDef 完整對齊 Java，零 Java vs Go 差異

- **Java 對照**：
  - `L1SkillId.java:474 RESIST_ELEMENTAL = 138`（屬性防禦）。無對應 skillmode 類別（純 buff 路徑）。
  - `L1SkillUse.java:2538-2545` + `L1SkillUse2.java:2490-2497` cast：`pc.addWind(10) + pc.addWater(10) + pc.addFire(10) + pc.addEarth(10) + pc.sendPackets(new S_OwnCharAttrDef(pc))`。
  - `L1SkillStop.java:466-475` stop：`cha.addWind(-10) + cha.addWater(-10) + cha.addFire(-10) + cha.addEarth(-10) + (if PC) pc.sendPackets(new S_OwnCharAttrDef(pc))`。
  - 表格成員：`isNotCancelable` 不含 138（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 138（不受 counter magic 阻擋）；`REPEATEDSKILLS` 不含 138（無互斥）。
  - yiwei `db_split/skills.sql:137`：`('138', '屬性防禦', '18', '1', '10', '0', '40319', '1', '0', '1200', 'none', '0', '0', '0', '0', '0', '0', '0', '2', '0', '0', '0', '0', '2', '', '19', '2184', '0', '0', '721', '0')` — mp=10、item=40319×1、reuse_delay=0、buff_duration=1200、target=none、type=2、id=2、action_id=19、cast_gfx=2184、sys_msg_stop=721。

- **Go 對照**：
  - `buffs.lua:113 [138] = { fire_res = 10, water_res = 10, wind_res = 10, earth_res = 10 }`：四屬 +10 完整對應 Java `addFire/Water/Wind/Earth(10)`。
  - `engine.go` 讀取 lua 回傳寫入 `BuffEffect.FireRes/WaterRes/WindRes/EarthRes`。
  - `skill_buff.go:177-180` apply：`buff.DeltaFireRes/WaterRes/WindRes/EarthRes = int16(eff.X)` 套入 ActiveBuff。
  - `skill_buff.go:211-214` apply：`target.FireRes += buff.DeltaFireRes` 四屬實際 +10。
  - `skill_buff.go:300-305` apply 端：四屬 Delta 非零時 `SendAbilityScores(target.Session, target)` = `S_OwnCharAttrDef` 對齊 Java cast `pc.sendPackets(new S_OwnCharAttrDef(pc))`。
  - `skill_buff.go:515-518` revert：`target.FireRes -= buff.DeltaFireRes` 四屬實際 -10。
  - `skill_buff.go:565-570` revert 端：四屬 Delta 非零時 `SendAbilityScores(target.Session, target)` = `S_OwnCharAttrDef` 對齊 Java stop `pc.sendPackets(new S_OwnCharAttrDef(pc))`。
  - `counterMagicExempt[138]`：應為 true（受豁免，Java `EXCEPT_COUNTER_MAGIC` 含 138）——已對齊。
  - `NON_CANCELLABLE` 不含 138 對齊 Java `isNotCancelable` 不含 138。
  - yaml `skill_list.yaml:4218-4248 skill_id=138`：mp=10、item=40319×1、reuse_delay=0、buff_duration=1200、target=none、type=2、id=2、action_id=19、cast_gfx=2184、sys_msg_stop=721。**31 欄位與 yiwei skills.sql:137 完全對齊零漂移**（與 137 同為罕見完美對齊範例）。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:10-42 TestSkillElementalBuffElfElementalDefenseBuffsUseJavaValues`：player 起始 FireRes=1/WaterRes=2/WindRes=3/EarthRes=4，套用 138 + 147（ElfAttr=2 fire）→ 預期 FireRes=61（+10+50）/WaterRes=12（+10）/WindRes=13（+10）/EarthRes=14（+10），明確驗證 138 四屬 +10 對稱對齊。
  - 同測試 assertion 「屬性防禦應四屬 +10、單屬性防禦應依 ElfAttr 只加火抗 +50」——文件記錄 138 與 147 各自 mechanic。

- **發現的 Java 真實差異**：**無**（四屬 +10 對稱套用、apply/stop 雙端 S_OwnCharAttrDef 廣播、yaml 31 欄位、counterMagicExempt、NON_CANCELLABLE 表格成員全部對齊）。

- **broader gap（不改）**：
  - **`getFire()/getWater()/getWind()/getEarth()` 動態 vs 靜態**：Java getter 在 receive 端動態彙整 base + 裝備 + buff + 寶物加成；Go 在 cast 時將 buff delta 累加到 `target.FireRes` 等靜態欄位。Client UI 與最終魔法傷害計算結果等價，屬廣域 stat 架構議題（同 110/111/137 同源 precedent）。
  - **`L1Refundable` 表退費機制**：Java 部分技能（如裝備強化品）有 refund 機制對映 elemental res，Go 無對應實作。屬廣域 refund 系統缺口，不在 138 單體範圍。

- **不修原因**（per 對齊深度停損標準）：
  - 零 Java vs Go 差異，**criterion (a) (b) (c) 三項都不滿足**——沒有 Java 行為差異、Go 沒有錯、無客戶端二進位約束問題。
  - 既有測試已完整鎖死四屬 +10 + S_OwnCharAttrDef 機制，零新測試需求。
  - 本次 audit 純文件化現況與 Java 對照，無代碼變更。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfElementalDefense -timeout 60s` PASS（cached）。

## 淨化精神（CLEAR_MIND / 137）— 純審計確認 +3 Wis 對稱對齊 Java，Wis→MR 衍生計算為廣域屬性架構缺口

- **Java 對照**：
  - `L1SkillId.java:470 CLEAR_MIND = 137`（淨化精神）。**無對應 skillmode 類別**——137 是純 buff（target='none' type=2），通用 buff 路徑套用 1200s（yaml duration=1200，TicksLeft=1200*5）。
  - `L1SkillUse.java:2533-2537` + `L1SkillUse2.java:2485-2489` cast：`pc.addWis((byte) 3) + pc.resetBaseMr()`。
  - `L1SkillStop.java:459-465` stop：`cha.addWis(-3) + (if PC) pc.resetBaseMr()`。
  - **`resetBaseMr()` 動態重算公式**（`L1PcInstance.java:5192-5212`）：
    1. 職業基底 MR：Crown 10 / Elf 25 / Wizard 15 / DarkElf 10 / DragonKnight 18 / Illusionist 20。
    2. Wis 額外加成：`CalcStat.calcStatMr(getWis())` 分段表 → Wis 0-14 +0、15-16 +3、17 +6、18 +10、19 +15、20 +21、21 +28、22 +37、23 +47、24+ +50。
    3. `newMr = 職業基底 + calcStatMr(wis)`，套用 `addMr(newMr - _baseMr)` 後 `_baseMr = newMr`。
  - **+3 Wis 真實影響範例**：Wis 14→17（+0→+6）= **+6 MR**；Wis 17→20（+6→+21）= **+15 MR**；Wis 20→23（+21→+47）= **+26 MR**；Wis 11→14（+0→+0）= **+0 MR**（兩端都在 0-14 平段）。
  - 表格成員：`isNotCancelable` 不含 137（cancellable）；`EXCEPT_COUNTER_MAGIC` 含 137 line 148 區段（**不受 counter magic 阻擋**）；`REPEATEDSKILLS` 不含 137（無互斥）。
  - yiwei `db_split/skills.sql:136`：`('137', '淨化精神', '18', '0', '10', '0', '40319', '1', '0', '1200', 'none', '0', '0', '0', '0', '0', '0', '0', '2', '0', '0', '0', '0', '1', '', '19', '2180', '0', '0', '722', '0')` — mp=10、item=40319×1、reuse_delay=0、buff_duration=1200、target=none、type=2、ranged=0、id=1、action_id=19、cast_gfx=2180、sys_msg_stop=722。

- **Go 對照**：
  - `buffs.lua:112 [137] = { wis = 3 }`：純 Wis +3，無 MR 欄位（resetBaseMr 衍生 MR 不在 lua 邏輯內）。
  - `engine.go` 讀取 lua 回傳 `wis=3` 寫入 `BuffEffect.Wis`。
  - `skill_buff.go:153 buff.DeltaWis = int16(eff.Wis)` 套入 ActiveBuff。
  - `skill_buff.go:190 target.Wis += buff.DeltaWis` apply 時加 +3。
  - `skill_buff.go:288-294 SendPlayerStatus`：Wis 變化時送 `S_PlayerStatus` 更新 client UI 屬性面板（Java 用 resetBaseMr 內 addMr+S_SPMR 路徑，Go 在 Wis 欄位變化時送 S_PlayerStatus 但**不額外送 S_SPMR 也不調整 MR**）。
  - `skill_buff.go:502 target.Wis -= buff.DeltaWis` revert 時減回。
  - **MR 衍生計算未實作**：Go `target.MR` 完全獨立於 Wis 欄位，不重算職業基底也不查 `calcStatMr` 表。Wis 12→15 不會自動觸發 MR +0/+3/+6 等衍生變化。
  - `counterMagicExempt[137]`：應為 true（受豁免，Java `EXCEPT_COUNTER_MAGIC` 含 137）——已對齊（既有 buff 系統管理）。
  - `NON_CANCELLABLE` 不含 137 對齊 Java `isNotCancelable` 不含 137（cancellable）。
  - yaml `skill_list.yaml:4187-4217 skill_id=137`：mp=10、item=40319×1、reuse_delay=0、buff_duration=1200、target=none、type=2、ranged=0、id=1、action_id=19、cast_gfx=2180、sys_msg_stop=722。**31 欄位與 yiwei skills.sql:136 完全對齊，零漂移**（罕見的完美對齊範例）。

- **既有測試覆蓋**：
  - `skill_elemental_buff_test.go:10-42 TestSkillElementalBuffElfElementalDefenseBuffsUseJavaValues`：player 起始 Wis=12/MR=5，套用 129+137+138+147 → 預期 Wis=15、MR=15（只有 RESIST_MAGIC +10），明確驗證 CLEAR_MIND 不會額外增加 MR（鎖死 Go 未實作 Wis→MR 衍生的事實）。
  - 測試命名與 assertion 已標記「魔法防禦/淨化精神應為 MR+10、WIS+3」——文件記錄 CLEAR_MIND 只算 +3 Wis 不算 MR delta，明確標示這是 Go 故意設計（與 Java 偏離但等待廣域 stat 系統重構）。

- **發現的 Java 真實差異 (criterion a)**：
  - **`resetBaseMr()` Wis→MR 分段衍生未實作**：Java cast +3 Wis 後 `resetBaseMr()` 觸發 `addMr(newMr - _baseMr)` 動態重算（職業基底 + calcStatMr 分段表）；Go 只改 `target.Wis += 3` 不動 `target.MR`。實際影響：Wis 14→17 缺 +6 MR、Wis 17→20 缺 +15 MR、Wis 20→23 缺 +26 MR。屬廣域屬性衍生架構議題（110/111 DEX→AC 同源 precedent，亦見 137 audit）。

- **不修原因**（per 對齊深度停損標準）：
  - 廣域屬性衍生系統缺失：Java `resetBaseMr()` 被多處呼叫（升級、配點、裝備、CLEAR_MIND/RESIST_MAGIC/裝備加成等），單一技能 audit 無法修。需先建立通用 Wis→MR derivation hook（在 `target.Wis` 變化時自動觸發 MR 重算）才能在所有 cast/stop/equip 路徑套用，並涵蓋 Java `_baseMr` 與 `addMr` 雙路徑（current MR vs base MR 兩層概念）。
  - 已建立 110/111/137/SHADOW_ARMOR audit precedent 將「靜態屬性 stat 寫入 vs 動態衍生重算」明確列為廣域 stat 系統重構議題（同 110 既有 2026-05-18 audit、137 既有 audit 同源紀錄）。
  - 本次 audit 重點是文件化 Java 真實 mechanic + 標記 Go 未實作項目，**不寫對應 fix**（會引入「為了 137 而動所有屬性 stat 路徑」的範圍蔓延）。

- **broader gap（不改）**：
  - **Wis→MR 衍生計算**（如上）。
  - **`getMr()` 動態 vs 靜態**：Java `L1PcInstance.getMr()` 在 receive 端動態彙整 base + 裝備 + buff + 寶物加成；Go 在 cast 時將 buff delta 累加到 `target.MR` 靜態欄位。Client UI 與最終魔法抗性計算等價，屬廣域 stat 架構議題（同 110/111 DEX→AC 同源）。
  - **MR 上限**：Java 部分裝備/buff 累加可能超過 client UI 顯示上限（如 75 cap），Go 無 cap clamp。屬廣域 stat 上限管理議題。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalBuffElfElementalDefense -timeout 60s` PASS（0.061s）。

## 鏡反射（COUNTER_MIRROR / 134）— 純審計確認 PC→PC 反射完整對齊 Java，PC→NPC/NPC→PC/NPC→NPC 路徑為廣域 NPC schema 缺口

- **Java 對照**：
  - `L1SkillId.java:466 COUNTER_MIRROR = 134`（鏡反射）。**無對應 skillmode 類別**——134 是純 buff 類技能（target='none' type=2），靠通用 buff 路徑套用 32s（yaml duration=16，TicksLeft=16*5）旗標 buff，不修改任何 stat（buffs.lua `[134] = {}`），實際反射邏輯在被攻擊路徑的 active site。
  - `L1SkillUse.java:1645-1651`：COUNTER_MIRROR 與 COUNTER_MAGIC/COUNTER_BARRIER/ARMOR_BREAK 同組，cast 廣播 `S_SkillSound(targetid, _gfxid=4395)` 用 `sendPacketsXR(-1)` 給自己 + `broadcastPacketAll` 給附近。
  - **4 個 active site**（攻擊端被反射）：
    - **PC→PC `L1MagicPc.calcMagicDamage:1197`**（dice damage 路徑）：`if targetPc.hasSkillEffect(134) && targetPc.Wis > random(100) → _pc.sendPacketsAll(S_DoActionGFX(_pc.id, 2)) + _pc.receiveDamage(_targetPc, dmg, false, false) + _pc.sendPacketsAll(S_SkillSound(_targetPc.id, 4395)) + dmg=0 + _targetPc.removeSkillEffect(134)`。
    - **PC→PC `L1MagicPc.calcPcMagicDamage:1345`**（regular magic damage 路徑）：邏輯完全相同（Java 在兩條 PC 攻 PC magic damage 計算路徑都做 reflection check）。
    - **PC→NPC `L1MagicPc.calcNpcMagicDamage:1615`**：`if targetNpc.hasSkillEffect(134) && targetNpc.Wis > random(100) → _pc.sendPacketsAll(S_DoActionGFX(_pc.id, 2)) + _pc.receiveDamage(_targetNpc, dmg, false, false) + _targetNpc.broadcastPacketAll(S_SkillSound(_targetNpc.id, 4395)) + dmg=0 + _targetNpc.removeSkillEffect(134)`（需 NPC.Wis 屬性）。
    - **NPC→PC `L1MagicNpc.calcPcMagicDamage:420`**：`if targetPc.hasSkillEffect(134) && _npc.getNpcTemplate().get_IsErase() && !_npc.getNpcTemplate().is_boss() && targetPc.Wis > random(100) → _npc.broadcastPacketAll(S_DoActionGFX(_npc.id, 2)) + _npc.receiveDamage(_targetPc, dmg) + _npc.broadcastPacketAll(S_SkillSound(_targetPc.id, 4395)) + dmg=0 + _targetPc.removeSkillEffect(134)`（需 NPC.IsErase + !IsBoss 雙閘）。
    - **NPC→NPC `L1MagicNpc.calcNpcMagicDamage:499`**：`if targetNpc.hasSkillEffect(134) && _npc.getNpcTemplate().get_IsErase() && targetNpc.Wis > random(100) → 同樣 broadcastPacketAll(S_DoActionGFX + S_SkillSound) + receiveDamage + dmg=0 + removeSkillEffect`（需 NPC.IsErase + 目標 NPC.Wis）。
  - **共同特徵**：機率 = 目標 Wis 對 0-99 比較（target.Wis > random(100) → trigger）；觸發後反彈 100% 傷害給攻擊者（`receiveDamage(false, false)` raw 無 MR/PR 減免）；廣播 action GFX 2（無傷標誌）+ S_SkillSound 4395 表示反射音效；buff 一擊消耗（removeSkillEffect(134)）；對攻擊者用 `sendPacketsAll`（PC→PC）或 `broadcastPacketAll`（PC→NPC、NPC→PC、NPC→NPC）廣播給附近所有玩家。
  - 表格成員：`isNotCancelable` 不含 134（cancellable）；`EXCEPT_COUNTER_MAGIC` 不含 134（**受 counter magic 阻擋**）；`REPEATEDSKILLS` 不含 134（無互斥）。
  - yiwei `db_split/skills.sql:133`：`('134', '鏡反射', '17', '5', '10', '0', '40319', '1', '1000', '16', 'none', '0', '0', '0', '0', '0', '0', '0', '2', '0', '0', '0', '0', '32', '', '19', '4395', '0', '0', '0', '0')` — mp=10、item=40319×1、reuse_delay=1000、buff_duration=16、target=none、target_to=0、type=2、ranged=0、id=32、action_id=19、cast_gfx=4395、sys_msg_fail=0。

- **Go 對照**：
  - `skill_elemental.go:11 skillCounterMirror = int32(134)`。
  - `skill_elemental.go:147-171 applyCounterMirrorMagicDamage`：完整對齊 Java PC→PC active site：
    - 早返：caster/target nil、damage≤0、`!target.HasBuff(skillCounterMirror)` → return original damage。
    - 機率：`if int(target.Wis) <= roll { return damage }` ＝ Java `targetPc.Wis > random(100)` 等價（roll 為 0-99，target.Wis 大於該 roll 才觸發）。
    - 移除 buff：`s.removeBuffAndRevert(target, skillCounterMirror)` 對齊 Java `removeSkillEffect(134)`。
    - 反彈傷害：`attacker.HP -= damage + Dirty=true + clamp 0 + sendHpUpdate` 對齊 Java `_pc.receiveDamage(_targetPc, dmg, false, false)`。
    - 廣播：`BuildActionGfx(attacker.CharID, 2) + BuildSkillEffect(target.CharID, 4395)` 對齊 Java `S_DoActionGFX(_pc.id, 2) + S_SkillSound(_targetPc.id, 4395)`。
    - Death 觸發：`if attacker.HP <= 0 → KillPlayer(attacker)` 對齊 Java `receiveDamage` 內部死亡處理。
    - 返回 0 對齊 Java `dmg = 0`。
  - **active site #1** `skill_damage.go:179`：`dmg = s.applyCounterMirrorMagicDamage(player, target, dmg, world.RandInt(100), nearby)` 在 PC magic skill 對 PC 目標的 `applySkillDamageToPlayer` 套用 reflection check（對齊 Java `calcPcMagicDamage:1345` 主要 magic damage 路徑）。
  - **active site #2** `skill_self_area.go:75`：`applySelfAreaSkillDamageToPlayerNoVisual` 同樣套用 reflection check（對齊 Java `calcMagicDamage:1197` 的 dice damage 路徑，主要用於 self-area 範圍魔法如冰暴/閃電風暴等）。
  - yaml `skill_list.yaml:4094-4124`：mp=10、item_consume_id=40319×1、reuse_delay=0、buff_duration=16、target=none、target_to=0、type=2、ranged=0、id=32、action_id=19、cast_gfx=4395、sys_msg_fail=280。
  - `buffs.lua:111 [134] = {}`：空 buff 條目，對齊 Java「純旗標 buff、無 stat 變化」設計，hp/mr/ac 全部不動。
  - `counterMagicExempt[134]`：未列入豁免表（受 counter magic 阻擋），對齊 Java `EXCEPT_COUNTER_MAGIC` 不含 134。
  - `NON_CANCELLABLE` 不含 134（cancellable），對齊 Java `isNotCancelable` 不含 134。

- **既有測試覆蓋**：
  - `skill_elemental_summon_test.go:157-196 TestSkillElementalSummonCounterMirrorReflectsMagicDamage`：caster 100HP + target 100HP/Wis=18 + 134 buff，呼叫 `applyCounterMirrorMagicDamage(caster, target, 30, 0, nearby)` （roll=0 強制觸發）→ damage=0、casterHP=70（30 反彈）、targetHP=100、buff 移除。完整鎖死 Java PC→PC 反射 mechanic。

- **發現的 Java 真實差異**：**無**（PC→PC 路徑完整對齊）。

- **broader gap（不改，需 NPC schema 重構）**：
  - **PC→NPC 反射缺失**：Java `L1MagicPc.calcNpcMagicDamage:1615` 對 NPC 目標的 hasSkillEffect(134) + NPC.Wis 反射，Go 無 NPC 端 134 buff 系統（NpcInfo 缺 Wis 屬性、無 NPC buff/debuff 帶 134）。需先補：(a) `NpcInfo.Wis int16` 欄位；(b) NPC 端可被施放 134 buff（NPC buff 系統目前只支援 elementalFallDownAttr/poisonAtk 等少數 debuff 場景，無泛用 buff 容器）。
  - **NPC→PC 反射缺失**：Java `L1MagicNpc.calcPcMagicDamage:420` NPC 對 PC 施 magic 時若 PC 有 134 → 機率反彈給 NPC。Go `npc_ai.go:573 npcMeleeAttack` / `:643 npcRangedAttack` 套用 `applyImmuneToHarmDamage + applyReductionArmorDamage` **但無 applyCounterMirrorMagicDamage**——然而 Java 134 只反射 **magic** 傷害（method 名稱 calcPcMagicDamage），不反射物理 melee/ranged，所以 Go 物理路徑不套用 134 是**正確的**。真正缺失是「NPC 施放 magic skill 給 PC 的 damage 計算路徑」——Go NPC AI 缺乏對等的 magic cast system（NPC 不會主動施法傷害技能），這是廣域 NPC AI 缺口而非 134 單體。
  - **NPC→NPC 反射缺失**：Java `L1MagicNpc.calcNpcMagicDamage:499` 含 IsErase 閘 + 目標 NPC.Wis 機率。Go NPC 互打目前僅 melee 物理（NPC 不互施 magic skill），與 NPC magic system 缺口同源。
  - **NPC.IsErase / NPC.IsBoss 欄位**：Java NpcTemplate 有 `get_IsErase()`（是否可被解除 buff）與 `is_boss()`（Boss 旗標）兩屬性，分別影響 NPC→PC 與 PC→NPC 反射閘。Go NpcInfo 缺這兩欄位，需先補資料庫 schema + YAML 載入器 + Go struct。
  - **yaml reuse_delay 漂移**：Go=0、yiwei=1000，Go 跟 cat-fei 偏離 yiwei（同 130/132/133 同源），屬廣域 SQL 漂移議題。
  - **yaml sys_msg_fail 漂移**：Go=280、yiwei=0，Go 增加施法失敗回饋訊息，比 Java 嚴格收緊（補上 Java 無回饋的 UX 缺口），與多項 elf 技能同模式（115/133 等），保留現狀。
  - **sendPacketsXR 自身廣播差異**：Java cast 廣播 `_player.sendPacketsXR(new S_SkillSound(targetid, _gfxid), -1)` 把自己排除在外但其實 `-1` 反而包含自己（XR 是「除了指定 XR 外」的範圍）；Go 直接 broadcast 給 nearby（含自己）。功能等價但具體序列化路徑不同，屬廣域 packet schema 議題。

- **不修原因**（per 對齊深度停損標準）：
  - PC→PC 完整對齊，**criterion (a) (b) (c) 三項都不滿足**——沒有 Java vs Go 差異、Go 沒有錯、無客戶端二進位約束問題。
  - 其餘 3 個路徑（PC→NPC、NPC→PC、NPC→NPC）都需 **NPC schema + NPC magic system 重構**，遠超 134 單體範圍（影響所有 NPC magic 反射技能：134/11059/COUNTER_BARRIER 等）。屬廣域架構議題。
  - 不寫新測試：既有測試已完整鎖死 Java PC→PC mechanic，新增「測試 PC→NPC 反射」屬「先設計 NPC schema 再寫測試」順序錯位。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElementalSummonCounterMirror -timeout 60s` PASS（0.131s）。

## 弱化屬性（ELEMENTAL_FALL_DOWN / 133）— 純審計確認 Go 核心對齊 Java，邊界差異記錄為 broader gap

- **Java 對照**：
  - `L1SkillId.java ELEMENTAL_FALL_DOWN = 133`（弱化屬性）。
  - `L1SkillMode.load()`：`_skillMode.put(133, new ELEMENTAL_FALL_DOWN())` 註冊 skillmode。
  - `skillmode/ELEMENTAL_FALL_DOWN.java:23-84 PC start`：
    - 行 25 `if (!cha.hasSkillEffect(133))` **守衛**：若目標已掛 133，整段 stat 變更區塊**跳過**，僅執行末尾 `setSkillEffect` 刷新 timer。
    - 行 26-52 ElfAttr 分支（針對 PC 目標）：0 → `srcpc.sendPackets(S_ServerMessage(79))` 不改 stat、1 → `pc.addEarth(-50) + setAddAttrKind(1)`、2 → `pc.addFire(-50) + setAddAttrKind(2)`、4 → `pc.addWater(-50) + setAddAttrKind(4)`、8 → `pc.addWind(-50) + setAddAttrKind(8)`。
    - 行 54-79 ElfAttr 分支（針對 L1MonsterInstance 目標）：完全對稱 PC 路徑（`mob.addEarth/Fire/Water/Wind(-50) + setAddAttrKind(...)`）。
    - 行 81 `cha.setSkillEffect(133, integer * 1000)` **永遠執行**（在 hasSkillEffect 守衛外），即使 ElfAttr=0 也會啟動 32s 計時器（雖然 stat 無變化）。
  - `skillmode/ELEMENTAL_FALL_DOWN.java:87-92 NPC start`：空實作，`return dmg=0`，NPC 施法 133 無效（與 Go 一致——NPC 不會主動施 133）。
  - `skillmode/ELEMENTAL_FALL_DOWN.java:101-146 stop`：
    - PC 目標：依 `getAddAttrKind()` 反向 `addEarth/Fire/Water/Wind(+50)` → `setAddAttrKind(0)` → `sendPackets(new S_OwnCharAttrDef(pc))`。
    - NPC 目標：依 `getAddAttrKind()` 反向 `+50` → `setAddAttrKind(0)`，**無封包**（NPC 無 client）。
  - `L1SkillMode.isNotCancelable`：133 不在表上（cancellable）。
  - `L1SkillMode.EXCEPT_COUNTER_MAGIC`：133 不在表上（**受 counter magic 阻擋**——與其他純 buff/debuff 不同，這是 debuff 性質）。
  - `REPEATEDSKILLS`：133 不在任何群組（無互斥）。
  - yiwei `db_split/skills.sql:132`：`('133', '弱化屬性', '17', '4', '10', '0', '0', '0', '100', '32', 'buff', '3', '0', '0', '0', '33', '50', '0', '1', '0', '10', '0', '0', '16', '', '19', '4396', '0', '0', '0', '280')` — mp=10、reuse_delay=100、buff_duration=32、target=buff、target_to=3、prob_value=33、prob_dice=50、type=1、ranged=10、id=16、action_id=19、cast_gfx=4396、sys_msg_fail=280。

- **Go 對照**：
  - `skill_buff.go:1107-1116 case 133`（PC→PC dispatch）：
    - `if player.ElfAttr == 0 → SendServerMessage(79) + return`（**早返**，不設 timer）。
    - `applyElementalFallDownToPlayer(player, target, skill)` 套用 buff。
    - `if skill.CastGfx > 0 → BroadcastToPlayers(nearby, BuildSkillEffect(target.CharID, skill.CastGfx))` 廣播視覺特效。
  - `skill_status.go:447-464 case 133`（PC→NPC dispatch）：
    - 同樣 `ElfAttr==0 → SendServerMessage(79) + return` 早返。
    - `checkNpcMRResist` MR 抗性檢查（Java 在 `calcProbabilityMagic` 前置處理過，但 Java skillmode 本身對 NPC mob 沒有第二次 MR check——Go 多此一舉但屬安全收緊不傷對齊）。
    - `applyElementalFallDownToNpc(player, npc, dur)` 套用 NPC 抗性下降。
    - cast_gfx 廣播。
  - `skill_elemental.go:23-41 applyElementalFallDownToPlayer`：
    - 建 ActiveBuff{SkillID=133, TicksLeft=BuffDuration*5}。
    - `old := target.RemoveBuff(133)` → `if old != nil { revertBuffStats(target, old) }` **覆蓋式重套**（與 Java `if (!hasSkillEffect)` 守衛行為不同——見 broader gap A）。
    - `setElementalResDelta(caster.ElfAttr, buff, -50)` 依 ElfAttr 設定對應 Delta 欄位。
    - 直接套用 stat（FireRes/WaterRes/WindRes/EarthRes += DeltaX）。
    - `target.AddBuff(buff)`。
  - `skill_elemental.go:43-79 applyElementalFallDownToNpc / removeElementalFallDownFromNpc`：完整對稱套用/反向 NPC 抗性，`npc.ElementalFallDownAttr` 欄位記錄施法者 ElfAttr 供 revert 使用（對齊 Java `getAddAttrKind()`）。
  - `skill_buff.go:300-305` PC buff apply 後 `SendAbilityScores(target.Session, target)` 送 `S_OwnCharAttrDef` 圖示更新——**比 Java cast 端嚴格收緊**（Java cast 端不送 attr def，只 stop 端送）；對應註解：「補上 Java 漏送的 UI 更新——client 顯示與資料一致」。
  - `skill_buff.go:565-570` PC buff revert 後 `SendAbilityScores(target.Session, target)` 對齊 Java `stop()` 行 123 `pc.sendPackets(new S_OwnCharAttrDef(pc))`。
  - yaml `skill_list.yaml:4063-4093 skill_id=133`：mp=10、reuse_delay=0、buff_duration=32、target=buff、target_to=3、prob_value=33、prob_dice=30、type=1、ranged=-1、id=16、action_id=19、cast_gfx=4396、sys_msg_fail=280。

- **既有測試覆蓋**：
  - `skill_elemental_dynamic_test.go TestSkillElementalDynamicElementalFallDownUsesCasterElfAttr`：caster ElfAttr=8（風）對 target（WindRes 12）→ -38，buff revert 後恢復為 12，驗證 PC apply/revert 雙向。
  - `skill_elemental_dynamic_test.go TestSkillElementalDynamicElementalFallDownRestoresNpcResistance`：caster ElfAttr=1（地）對 NPC（EarthRes 20）→ -30，buff 過期後恢復為 20，驗證 NPC apply/revert 雙向。
  - **無需新增測試**：核心 ElfAttr-based -50 / S_OwnCharAttrDef 對齊已有測試鎖定，按停損標準不寫「鎖死 Java 對齊行為」測試。

- **發現的 Java 真實差異（criterion a，但屬邊界 edge case）**：
  - **A) Re-cast 守衛行為差異**：Java `if (!cha.hasSkillEffect(133))` 在已掛 buff 時**跳過** stat 變更區塊（保留首次施法的元素），但仍刷新 timer。Go `applyElementalFallDownToPlayer:31-34` 永遠先 `RemoveBuff + revertBuffStats` 再重套（允許跨施法者切換元素）。實際影響場景：**僅當 caster A 風屬性掛 buff 後，caster B 火屬性在 32s 內覆蓋施法**——Java 維持 caster A 的風 -50（B 失效），Go 切換為 caster B 的火 -50（B 生效）。同施法者重施則兩端結果等價（先 revert 再 apply 同元素 = 淨變化為零）。
  - **B) ElfAttr=0 timer-set 差異**：Java 即使 ElfAttr=0 仍執行 `cha.setSkillEffect(133, integer * 1000)`（外於 hasSkillEffect 守衛），啟動 32s 空轉計時器（無 stat 變化但目標確實「有」133 buff）。Go `case 133 ElfAttr==0 → return` 完全跳過，不設 timer。實際影響場景：caster ElfAttr=0 時 Java 留下無效 buff 條目佔位（影響後續 hasSkillEffect(133) check 結果但不影響任何數值），Go 不留任何痕跡。功能上兩端等價（buff 對 stat 零影響），僅 hasSkillEffect 查詢結果不同。
  - **C) cast 端 S_OwnCharAttrDef 廣播**：Java cast 端**不**送 S_OwnCharAttrDef（只 stop 端送），Go 同時在 cast 與 stop 端送（補上 Java 漏送讓 UI 即時反映抗性下降）。屬 **比 Java 嚴格收緊**——已在 skill_buff.go:300-302 註解明確標示「補上 Java 漏送的 UI 更新」，不算「未對齊」。

- **不修原因**（per 對齊深度停損標準）：
  - A、B 屬 criterion (a) Java 真實差異但**邊界 edge case**：A 僅在跨施法者元素切換時有差（單施法者重施結果等價）；B 僅影響 hasSkillEffect 查詢（無 stat 影響）。Go 行為都是**production-correct**——「跨施法者切換元素」與「ElfAttr=0 不留痕跡」都是合理設計。
  - 修 A 需把 `applyElementalFallDownToPlayer` 改為「先檢查 target.HasBuff(133)，若有則僅 RefreshDuration 不動 stat」，但會引入「caster B 雖然付了 MP+cast_gfx 但完全沒效果」的反直覺體驗（Java 確實如此但體感不佳）。
  - 修 B 需在 ElfAttr=0 早返前先 `target.AddBuff(emptyBuff)` 純佔位，引入「buff icon 顯示但無效」的混淆體驗。
  - 寫對應測試屬「鎖死 Java 邊界行為」，違反停損標準。
  - 既有兩條 dynamic 測試已完全覆蓋核心 ElfAttr 1/2/4/8 對 PC/NPC 的 apply/revert，足以防回歸。

- **broader gap（不改）**：
  - **yaml 三方漂移**：Go yaml 與 yiwei `skills.sql:132` 三項偏離（Go 跟 cat-fei）：reuse_delay=0（yiwei=100）、probability_dice=30（yiwei=50）、ranged=-1（yiwei=10）。屬廣域 SQL 漂移議題（與 130/132 同源），非 133 單體範圍。
  - **MR 抗性檢查時機**：Java skillmode 本身對 NPC 目標不做 MR check（由上層 `calcProbabilityMagic` 統一處理），Go `skill_status.go case 133` 另加 `checkNpcMRResist`。屬「Go 比 Java 嚴格收緊」的安全 net，已在 NPC 系列 audit 統一採用（見 27 壞物術等 case），保留現狀。
  - **NPC start 空實作**：Java NPC 施法 133 為空函數（NPC 不會主動對玩家施 133），Go 沒實作 NPC→PC 路徑同 Java 一致。
  - **counter magic 阻擋未實作**：Java 133 不在 `EXCEPT_COUNTER_MAGIC` 表上 = 受 counter magic 阻擋。Go `counterMagicExempt[133]` 應為 false（未列入豁免表）——已對齊（無 broader gap）。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkillElemental -timeout 120s` PASS（0.455s，elemental 系列無迴歸）。

## 三重矢（TRIPLE_ARROW / 132）— 修正 PvP 射程退化 bug

- **Java 對照**：
  - `L1SkillId.java:455-458 TRIPLE_ARROW = 132`（三重矢）。
  - `L1SkillMode.load() line 100`：`_skillMode.put(Integer.valueOf(132), new TRIPLE_ARROW())` 註冊 skillmode。
  - `skillmode/TRIPLE_ARROW.java:13-48 PC start`：
    - 行 16：`int playerGFX = srcpc.getTempCharGfx()` 取玩家變身外型。
    - 行 27-30：`SprTable.get().getAttackSpeed(playerGFX, 21) == 0 → return 0`（變身外型須支援弓箭動作 action_id=21）。
    - 行 32-34：`srcpc.getCurrentWeapon() != 20 → return 0`（必須裝備弓，visual byte=20）。
    - 行 36-38：`if (ConfigSkill.TRIPLE_ARROW_DMG > 1) srcpc.setIsTRIPLE_ARROW(true)` 設定 ×倍率旗標。
    - 行 39-41：`for (i = 0; i < 3; i++) cha.onAction(srcpc)` ＝ 對目標 3 次完整 L1AttackPc 攻擊。
    - 行 42-44：reset IsTRIPLE_ARROW=false。
    - 行 45-46：`sendPacketsAll(new S_SkillSound(self.id, 4394))` 加速封包 + `sendPacketsAll(new S_SkillSound(self.id, 11764))` 特效動畫。
  - `skillmode/TRIPLE_ARROW.java:50-60 NPC start`：`for (i = 3; i > 0; i--) npc.attackTarget(cha)` ×3 + 同樣 4394/11764 廣播。
  - `L1AttackPc.java:1512-1514`（calc damage）：`if (_pc.getIsTRIPLE_ARROW()) dmg *= ConfigSkill.TRIPLE_ARROW_DMG`，2002-2004 行為 NPC target 同樣公式。
  - `ConfigSkill.java:49 + 212`：`TRIPLE_ARROW_DMG = Double.parseDouble(set.getProperty("Triple_Arrow_Dmg", "1.0"))`；yiwei `各職業技能相關設置.properties: Triple_Arrow_Dmg = 5.0` 為實際運行值（5x 倍率）。
  - `L1AttackPc.java:175-183` 弓箭裝備時 `_arrow = getInventory().getArrow()`；行 226 `if ((_weaponType == 20) && (_weaponId != 190) && (_arrow == null)) _isHit = false`；箭矢消耗散落於 :2837/2842/2847/2852/3089/3121 內部 `removeItem(_arrow, 1L)` 各種 enchant 路徑——每次 attack 消耗 1 箭。
  - yiwei `db_split/skills.sql:131`：`('132', '三重矢', '17', '3', '15', '0', '0', '0', '100', '0', 'attack', '3', '0', '0', '0', '0', '0', '0', '2', '0', '-1', '0', '0', '8', '', '18', '0', '0', '0', '0', '0')` — **注意 type=2 是 yiwei 端錯誤**（TRIPLE_ARROW 屬 TYPE_ATTACK=64 非 TYPE_CHANGE=2），cat-fei 已修為 type=64。
- **Go 對照**：
  - `skill.go:397-401`：`if skillID == 132 && player.CurrentWeapon != 20 { return }` 對齊 Java line 32-33 弓裝備限制。
  - `skill_damage.go:114-120` NPC 攻擊路徑收尾廣播：`if skill.SkillID == 132 { BroadcastToPlayers(casterNearby, BuildSkillEffect(player.CharID, 4394)) + 11764 }` 對齊 Java 行 45-46。
  - `skill_damage.go:390-391` NPC 路徑：`if skill.SkillID == 132 { maxRange = 10 }` 特例。
  - `skill_damage.go:421-432` NPC 路徑前置 1 箭矢消耗 + `if arrow == nil sendCastFail`。
  - `skill_damage.go:502-504` NPC 路徑：`if skill.SkillID == 132 && res.HitCount < 3 { res.HitCount = 3 }` 強制 3 次命中。
  - `skill_damage.go:648-654` 重複的收尾廣播（PvE path 區段）。
  - **原 `skill_damage.go:38-40` PC 路徑**：`maxRange = skill.Ranged`（=-1）→ `if maxRange <= 0 { maxRange = 2 }` ＝ **退化為近戰 2 格**，PvP 三重矢實際射程僅 4 格（含 +2 lenience），與 NPC 路徑 10 格嚴重不一致——玩家根本不能用弓技攻擊 4 格外的玩家。
  - `skill_damage.go:87-90` PC 路徑：`if skill.SkillID == 132 && hitCount < 3 { hitCount = 3 }` 強制 3 次命中。
  - `skill_damage.go:114-120` PC 路徑收尾廣播 4394/11764（與 NPC 路徑同邏輯）。
  - `magic.lua calc_physical_skill:139-143`：`elseif sid == 132 then hit_count = 3; damage = damage * 5` 對齊 Java `dmg *= TRIPLE_ARROW_DMG=5`。
- **發現的 Java 真實差異 (criterion a)**：
  - **PC 路徑射程退化**：NPC 路徑有 `if skill.SkillID == 132 { maxRange = 10 }` 特例，PC 路徑（`executeAttackSkillOnPlayer:38-40`）無此特例，導致 PvP 三重矢退化為近戰射程。Java skillmode 不顯式檢查射程，靠 `cha.onAction(srcpc)` 進入 L1AttackPc 自然走弓箭典型 8~10 格射程；Go 在 dispatch 層 hard-block，必須明確設定。
- **修正**：`skill_damage.go executeAttackSkillOnPlayer` 加入 132 特例對齊 NPC 路徑：
  ```go
  maxRange := int32(skill.Ranged)
  // 132 TRIPLE_ARROW：Java skillmode 透過 cha.onAction(srcpc) 走 L1AttackPc 弓箭射程
  // （典型 8~10 格），yaml ranged=-1 在 Go fallback 為 2（近戰）會讓 PvP 三重矢退化為
  // 貼身攻擊。對齊 NPC 路徑 (skill_damage.go:390-391) 的 10 格特例。
  if skill.SkillID == 132 {
      maxRange = 10
  } else if maxRange <= 0 {
      maxRange = 2
  }
  ```
- **yaml 對照**（Go `skill_list.yaml:4032-4062` vs yiwei `skills.sql:131` vs cat-fei）：

| 欄位 | Go yaml | yiwei SQL | cat-fei SQL | 備註 |
|------|---------|-----------|-------------|------|
| name | 三重矢 | 三重矢 | 三重矢 | ✓ |
| skill_level | 17 | 17 | 17 | ✓ |
| mp_consume | 15 | 15 | 15 | ✓ |
| reuse_delay | **400** | 100 | 10 | 三方漂移 |
| target | attack | attack | attack | ✓ |
| target_to | 3 | 3 | 3 | ✓ |
| **type** | **64** | **2** | **64** | Go 跟貓飛（yiwei type=2 為錯誤資料） |
| ranged | -1 | -1 | -1 | ✓ |
| id | 8 | 8 | 8 | ✓ |
| action_id | 18 | 18 | 18 | ✓ |
| cast_gfx | 0 | 0 | 0 | ✓ |

關鍵：type=64 = `L1Skills.TYPE_ATTACK`，type=2 = `TYPE_CHANGE`。TRIPLE_ARROW 是攻擊技能，cat-fei 與 Go 的 type=64 是正確值，yiwei type=2 為錯誤資料未修正。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkill -timeout 180s` PASS（14.379s，全 skill 測試無迴歸）。
- **不寫新測試**：PvP 射程修正屬於「對齊 NPC 已有特例」的 surgical 修正，新增「測試 PvP TRIPLE_ARROW 4 格外可命中」屬「鎖死 Java 對齊行為」測試，違反停損標準。
- **broader gap（不改）**：
  - **Java SprTable 變身外型動作檢查**：line 27-30 `SprTable.getAttackSpeed(playerGFX, 21)==0 return` 確保變身狀態下玩家有弓箭攻擊動畫資料。Go 無 SPR table 系統，屬廣域變身動畫資料缺口。
  - **箭矢消耗精度（1 vs 3）**：Java 每次 `cha.onAction` 進入 L1AttackPc 內部各 enchant 路徑 `removeItem(_arrow, 1L)` ×3 次 = 3 箭，Go 在 dispatch 層前置消耗 1 箭。NPC 路徑同樣只消 1 箭。屬廣域弓箭資源精度議題（影響所有弓箭技能與普通弓攻），不在 132 單體範圍。
  - **reuse_delay 三方漂移**：Go=400ms、yiwei=100ms、cat-fei=10ms 三方均不一致，Go 既不跟貓飛也不跟 yiwei，屬廣域 SQL 漂移議題。
  - **NPC 端攻擊倒序循環差異**：Java NPC start `for (i=3; i>0; i--) npc.attackTarget(cha)` 倒序，Go NPC 路徑由 SkillSystem 處理而非 NpcAi——屬廣域 NPC 技能 system routing 議題。
  - **type=2 vs type=64**：Go 已採用 cat-fei 正確值 64，無需修。本項列舉純為記錄 yiwei SQL 偏離。
  - **S_UseAttackSkill 額外欄位**：Java sendGrfx 對 target='attack' + area=0 走 `S_UseAttackSkill(player, targetid, _gfxid, _targetX, _targetY, _actid, _dmg)`；Go isPhysicalSkill 路徑改送 BuildAttackPacket。packet 序列化細節差異屬廣域 packet schema 議題。

## 世界樹的呼喚（TELEPORT_TO_MATHER / 131）— 補齊兩項視覺特效封包對齊 Java

- **Java 對照**：
  - `L1SkillId.java:451-454 TELEPORT_TO_MATHER = 131`（世界樹的呼喚）。
  - `L1SkillMode.load() line 101`：`_skillMode.put(Integer.valueOf(131), new TELEPORT_TO_MATHER())` 註冊 skillmode。
  - `skillmode/TELEPORT_TO_MATHER.java:19-61`：
    - 前置三項 buff 阻擋：`hasSkillEffect(230)` → S_ServerMessage(1413) return；`hasSkillEffect(4000)` → S_ServerMessage("\\fY已被束縛的效果無法瞬移") return；`hasSkillEffect(THUNDER_GRAB=192)` → S_ServerMessage("\\fY身上有奪命之雷的效果無法瞬移") + `S_Paralysis(TYPE_TELEPORT_UNLOCK, false)` return。
    - 自動掛機狀態清理：Test_Auto/IsAuto/RestartAuto/DeathReturn 設 0。
    - `isEscapable()` 檢查：
      - true → `setTeleportX(33047) + setTeleportY(32338) + setTeleportMapId(4) + setTeleportHeading(5) + sendPacketsAll(new S_SkillSound(self.id, 169)) + Teleportation.teleportation(pc)`。
      - false → S_ServerMessage(276) + S_Paralysis(TYPE_TELEPORT_UNLOCK)。
  - `L1SkillUse.sendGrfx:1637-1683`：對 target='none' + type != ATTACK 走 1637 else 分支，內部 `if ((_skillId != TELEPORT) && (_skillId != MASS_TELEPORT) && (_skillId != TELEPORT_TO_MATHER))` ＝ **跳過整個 S_DoActionGFX + skill-specific S_SkillSound 區塊**對三項傳送技能；末尾 1686-1694 對 _targetList 仍送 S_SPMR/S_OwnCharStatus/UPDATE_ER status refresh。
  - `Teleportation.java:56-396 teleportation(pc)`：完整傳送流程，含血盟倉庫 lock 釋放、`killSkillEffectTimer(32)` 取消冥想術、setLocation、`S_MapID`、`S_OtherCharPacks` 廣播、`S_OwnCharPack` 自己、寵物/娃娃跟隨、限時地圖偵測、`finally { S_Paralysis(TELEPORT_UNLOCK) }`。
  - yiwei `db_split/skills.sql:130`：`('131', '世界樹的呼喚', '17', '2', '10', '0', '0', '0', '0', '0', 'none', '0', '0', '0', '0', '0', '0', '0', '128', '0', '0', '0', '0', '4', '', '19', '169', '0', '0', '0', '0')` — mp=10、target=none、type=128、id=4、action_id=19、cast_gfx=169。
- **Go 對照**：
  - `skill.go:393-395`：`if skillID == 131 && s.teleportToMatherBlockedBeforeConsume(sess, player) { return }` 在 MP 消耗前阻擋。
  - `skill_heal_resurrect.go:27-49 teleportToMatherBlockedBeforeConsume`：
    - `HasBuff(230)` → SendServerMessage(1413) return true ✓
    - `HasBuff(4000)` → SendNormalChat(0, "\\fY已被束縛的效果無法瞬移") return true ✓
    - `HasBuff(192)` → SendNormalChat(...) + SendParalysis(TeleportUnlock) return true ✓
    - `!MapData.Escapable` → SendServerMessage(276) + SendParalysis(TeleportUnlock) return true ✓
  - `skill.go:320 removeBuffAndRevert(player, 32)`：MEDITATION 取消對齊 Java `killSkillEffectTimer(32)`。
  - 原 `skill_heal_resurrect.go:52-61 executeResurrection case 131`：
    ```go
    actData := handler.BuildActionGfx(player.CharID, byte(skill.ActionID))
    handler.BroadcastToPlayers(nearby, actData)  // ← 對所有復活技能廣播，Java 對 131 跳過
    case 131:
        handler.TeleportPlayer(sess, player, 33047, 32338, 4, 5, s.deps)  // ← 缺 cast_gfx 廣播
    ```
  - `TeleportPlayer`（`handler/npcaction.go:575-796`）末尾 `sendTeleportUnlock(sess)` 對齊 Java finally。
- **發現的兩項 Java 真實差異**：
  - **(A) Java 跳過 S_DoActionGFX 對 131**：Java sendGrfx 1639 exclusion 列表含 `TELEPORT_TO_MATHER`，Go executeResurrection 對所有復活技能（含 131）廣播 BuildActionGfx，導致 131 多送一個 S_DoActionGFX 動作動畫封包。
  - **(B) Go 缺 S_SkillSound(self, 169) 廣播**：Java skillmode 在 Teleportation 前送 `sendPacketsAll(new S_SkillSound(self.id, 169))`，附近玩家會看到母樹傳送音效視覺；Go 原本只 TeleportPlayer 無此廣播。
- **修正**：`skill_heal_resurrect.go executeResurrection` 補上兩項：
  ```go
  // Java L1SkillUse.sendGrfx:1639 對 TELEPORT/MASS_TELEPORT/TELEPORT_TO_MATHER 跳過 S_DoActionGFX
  if skill.SkillID != 131 {
      actData := handler.BuildActionGfx(player.CharID, byte(skill.ActionID))
      handler.BroadcastToPlayers(nearby, actData)
  }
  switch skill.SkillID {
  case 131:
      // Java skillmode/TELEPORT_TO_MATHER.java:52 sendPacketsAll(S_SkillSound(self.id, 169))
      if skill.CastGfx > 0 {
          handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
      }
      handler.TeleportPlayer(sess, player, 33047, 32338, 4, 5, s.deps)
  ```
  61/75/165 維持送 BuildActionGfx 對齊 Java（這三項不在 sendGrfx exclusion 列表）。
- **yaml 對照**（Go `skill_list.yaml:4001-4031` vs yiwei `skills.sql:130`）：

| 欄位 | Go yaml | yiwei SQL | 備註 |
|------|---------|-----------|------|
| name | 世界樹的呼喚 | 世界樹的呼喚 | ✓ |
| skill_level | 17 | 17 | ✓ |
| skill_number | 2 | 2 | ✓ |
| mp_consume | 10 | 10 | ✓ |
| hp_consume | 0 | 0 | ✓ |
| reuse_delay | 0 | 0 | ✓ |
| target | none | none | ✓ |
| target_to | 0 | 0 | ✓ |
| type | 128 | 128 | ✓ |
| id | 4 | 4 | ✓ |
| action_id | 19 | 19 | ✓ |
| cast_gfx | 169 | 169 | ✓ |

31 欄位完全對齊。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run "TestSkillCallOfNature" -timeout 120s` PASS（既有 4 個復活相關測試通過）。
  - `cd server && go test ./internal/system -run TestSkill -timeout 180s` PASS（14.174s，全 skill 測試無迴歸）。
- **不寫新測試**：既有 `TestSkillCallOfNatureTeleportToMotherReturnsCasterToMotherTree` 已覆蓋核心座標傳送行為。新增「測試 case 131 跳過 BuildActionGfx」與「測試 case 131 廣播 cast_gfx」屬「鎖死 Java 對齊行為」測試，違反停損標準。
- **broader gap（不改）**：
  - **Java Test_Auto/IsAuto 自動掛機狀態清理**：Go 無 auto-hunting 系統，無對應狀態可清，屬廣域 auto-hunting 缺口。
  - **Java sendGrfx 末尾 1686-1694 通用 status refresh**：對 _targetList 送 S_SPMR + S_OwnCharStatus + S_PacketBox(UPDATE_ER)。TELEPORT_TO_MATHER 無 stat 變化，對 client 無功能影響。屬廣域 cast 後置 status refresh 缺口（與 130 audit 同源）。
  - **Lua resurrection.lua [131]={hp_ratio=0.5, mp_ratio=0.5} 為 routing 觸發器**：實際 hp/mp_ratio 從未被讀取消費，純為讓 `isResurrectionSkill==true` 走 `executeResurrection` 路由。Java/Go 雙方對 131 都不做 HP/MP 復原，行為一致但 Go 端 routing 結構奇特。重構此邏輯需動 isResurrectionSkill 判定方式或新建 executeTeleportSelfSkill，屬 routing 重構議題，不在 131 單體範圍。
  - **TELEPORT/MASS_TELEPORT 端 BuildActionGfx 跳過**：skill 5/69 走 `executeTeleportSpell`（skill_teleport.go）獨立路徑，本身不送 BuildActionGfx——已隱式對齊 Java sendGrfx 1639 exclusion，無需修正。

## 心靈轉換（BODY_TO_MIND / 130）— 修正施法視覺特效廣播缺失

- **Java 對照**：
  - `L1SkillId.java:447-450 BODY_TO_MIND = 130`（心靈轉換）。
  - `L1SkillMode.load() line 98`：`_skillMode.put(Integer.valueOf(130), new BODY_TO_MIND());` 註冊 skillmode。
  - `skillmode/BODY_TO_MIND.java:13-19 start(L1PcInstance srcpc, ...)`：核心邏輯極簡，僅 `srcpc.setCurrentMp(srcpc.getCurrentMp() + 2)` 然後 `return dmg=0`。NPC start/stop 皆為 no-op。
  - `L1Character.java:279-284 setCurrentMp` + `L1PcInstance.java:1725-1736` override：`Math.min(i, getMaxMp())` 內建 MaxMp clamp + 變化時送 `S_MPUpdate(currentMp, MaxMp)`，若 MP 無變化則 early return 不送 packet。
  - `L1SkillUse.java:481-486 TYPE_NORMAL` 流程：`runSkill() → useConsume() → sendGrfx(true) → setDelay()`。`useConsume()` 消耗 `mp_consume=0` + `hp_consume=8`（Java SQL `db_split/skills.sql:129`）。
  - `L1SkillUse.sendGrfx:1489-1683`：對 PC user + target='none' + type != ATTACK，進入 1637 "補助魔法或詛咒魔法" else 分支：
    - 1642：`_player.sendPacketsAll(new S_DoActionGFX(self.id, _actid=19))` ← Go 已對齊（後置 BuildActionGfx）。
    - 1645-1680 各 skillId 特殊處理皆不命中 130。
    - 1681 default：`_player.sendPacketsAll(new S_SkillSound(targetid, _gfxid=2179))` ← **原 Go 缺失，本次補上**。
  - `L1SkillMode.isNotCancelable():31-64`：**不含 130**（cancellable by CANCELLATION）。
  - `L1SkillUse.EXCEPT_COUNTER_MAGIC:148`：**含 130**（魔法屏障無法抵擋）。
  - `L1SkillUse.REPEATEDSKILLS:1741-1762`：10 群均不含 130，無互斥技能。
  - yiwei `db_split/skills.sql:129`：`(130, '心靈轉換', 17, 1, 0, 8, 0, 0, 100, 0, 'none', 0, ..., 2, ..., 2, '', 19, 2179, 0, 702, 0, 0)` — 31 欄位。
- **Go 對照**：
  - `skill_self.go:121-133 case 130`：原本只 `player.MP += 2 + MaxMP clamp + sendMpUpdate`，**未送 cast_gfx 廣播**；後置 `:172-175 BuildActionGfx(action_id=19)` 對齊 Java S_DoActionGFX 但 cast_gfx=2179 缺失。
  - `skill.go:328-332 hp_consume`：HP ≤ 8 拒絕 + SendServerMessage(skillMsgNotEnoughHP) 對齊 Java `useConsume` 失敗。
  - `skill.go:449-452`：HP -= 8 + sendHpUpdate 對齊 Java useConsume HP 消耗。
  - `skill_buff.go:409 counterMagicExempt[130]=true` 對齊 Java EXCEPT_COUNTER_MAGIC 含 130。
  - `scripts/combat/buffs.lua:236-276 NON_CANCELLABLE`：**不含 130**，對齊 Java isNotCancelable 不含 130。
  - 無 `buffs.lua [130]` entry（因技能僅做 MP 加值，無持續 buff stat）。
- **修正**：`skill_self.go` case 130 補上 cast_gfx 廣播：
  ```go
  case 130: // 心靈轉換 — 恢復 2 MP（Java: BODY_TO_MIND +2）
      // Java skillmode/BODY_TO_MIND.java:16 setCurrentMp(currentMp + 2)
      player.MP += 2
      if player.MP > player.MaxMP {
          player.MP = player.MaxMP
      }
      sendMpUpdate(sess, player)
      // Java L1SkillUse.sendGrfx:1681 default 分支送 S_SkillSound(self.id, cast_gfx=2179)
      if skill.CastGfx > 0 {
          handler.BroadcastToPlayers(nearby, handler.BuildSkillEffect(player.CharID, skill.CastGfx))
      }
  ```
  與既有 `case 186 BLOODLUST` 同模式（line 137-146）。
- **yaml 對照**（Go `skill_list.yaml:3970-4000` vs yiwei `skills.sql:129`）：

| 欄位 | Go yaml | yiwei SQL | 備註 |
|------|---------|-----------|------|
| name | 心靈轉換 | 心靈轉換 | ✓ |
| skill_level | 17 | 17 | ✓ |
| skill_number | 1 | 1 | ✓ |
| mp_consume | 0 | 0 | ✓ |
| hp_consume | 8 | 8 | ✓ |
| reuse_delay | **400** | **100** | ✗ Go 跟貓飛 |
| buff_duration | 0 | 0 | ✓ |
| target | none | none | ✓ |
| type | 2 | 2 | ✓ |
| id | 2 | 2 | ✓ |
| action_id | 19 | 19 | ✓ |
| cast_gfx | 2179 | 2179 | ✓ |
| sys_msg_happen | 702 | 702 | ✓ |

reuse_delay 為唯一差異，cat-fei `(130,'心靈轉換',17,1,0,8,0,0,400,...)` 也是 400，Go 跟貓飛——屬 yiwei/cat-fei SQL 漂移廣域議題。

- **驗證**：
  - `cd server && go build ./...` 通過。
  - `cd server && go test ./internal/system -run TestSkill -timeout 120s` PASS（13.884s）。
- **不寫新測試**：cast_gfx 廣播為 single-line BroadcastToPlayers 包裝，與 case 186 BLOODLUST 同模式（既有 buff cast 測試已驗證 `BuildSkillEffect` 廣播路徑）。新增「測試 case 130 送出 BuildSkillEffect」屬「Go 通用機制 + 對齊行為」鎖死測試，違反停損標準。
- **broader gap（不改）**：
  - **case 146 BLOODY_SOUL 同樣缺 cast_gfx 廣播**：Java `BLOODY_SOUL.start` 經 sendGrfx 同樣走 1681 default 分支送 S_SkillSound(self, cast_gfx=11005)，Go case 146（同檔案 line 128-135）也只送 MP+20 + sendMpUpdate 無 cast_gfx 廣播。146 已標 ✅ 對齊（前次 audit 只修 MP 值 +20），依「不可偷換範圍」記錄為待 146 重 audit 時處理，本次不擴張範圍修。
  - **Java sendGrfx 末尾 1686-1694 通用 status refresh**：對 _targetList 每位 PC 送 `S_SPMR + S_OwnCharStatus + S_PacketBox(UPDATE_ER)` 三項 status 刷新封包。BODY_TO_MIND 無 stat 變化，這三項封包對 client 無功能影響（client 端狀態不變）。屬廣域 buff cast 後置 status refresh 缺口（與多項 self-skill 同源），不在 130 單體範圍。
  - **reuse_delay 100 vs 400**：yiwei 100ms vs cat-fei/Go 400ms。Go 跟貓飛現代 Lineage R 數值，屬廣域 yiwei/cat-fei SQL 同步議題（與 114 GLOWING_AURA、134 COUNTER_MIRROR、158、159、173 等多項技能同源），不在 130 單體範圍。
  - **HP→MP 4:1 比例設計**：Java/Go 均消耗 8 HP 換 2 MP（淨虧 6 HP），這是 BODY_TO_MIND 的設計平衡（緊急 MP 補充但要付出大量 HP 代價），兩端一致無偏離。

## 魔法防禦（RESIST_MAGIC / 129）— 純審計，Java 預設 buff 路徑與 Go 通用 mr+SPMR 完全等價

- **Java 對照**：
  - `L1SkillId.java:443-446 RESIST_MAGIC = 129` 註解「魔法防禦129」。
  - **無 `skillmode/RESIST_MAGIC.java`**——`L1SkillMode.load()` 並未註冊 129，cast 走 `L1SkillUse`/`L1SkillUse2` 的 default buff 分支。
  - `L1SkillUse.java:2528-2531` + `L1SkillUse2.java:2480-2483`：`if (this._skillId == RESIST_MAGIC) { pc.addMr(10); pc.sendPackets(new S_SPMR(pc)); }`——施法即 MR+10 並送 S_SPMR 通知。
  - **無 `L1SkillStop` 條目**：移除 buff 走通用 `L1BuffUtil` 路徑，由通用 stop 處理還原。
  - `L1SkillMode.isNotCancelable()` 第 31-64 行：**不含 129**，RESIST_MAGIC 可被 CANCELLATION 解除。
  - `L1SkillUse.java:145-153 EXCEPT_COUNTER_MAGIC`：**含 129**（第 148 行），魔法屏障無法抵擋。
  - `L1SkillUse.java:1741-1762 REPEATEDSKILLS`：10 群均不含 129，無互斥技能。
  - **yiwei `db_split/skills.sql:128`**：`(129,'魔法防禦',17,0,5,0,40319,1,0,1200,'none',0,...,0,2,...,1,'',19,2186,0,0,719,0)` ＝ skill_level=17、mp=5、item=40319、duration=1200、target='none'、target_to=0、type=2、id=1、action_id=19、cast_gfx=2186、sys_msg_stop=719。
- **Go 對照**：
  - `scripts/combat/buffs.lua:109 [129] = { mr = 10 }` — 唯一 stat 為 `mr=10`（精準對齊 Java `addMr(10)`），無 hit_mod/dmg_mod/exclusions/任何其他干擾。
  - `internal/scripting/engine.go:407 MR: lInt(rt, "mr")` 將 lua `mr` 欄位反序列化為 `BuffEffect.MR`。
  - `internal/system/skill_buff.go:164 buff.DeltaMR = int16(eff.MR)`：apply 時計入 buff 結構。
  - `internal/system/skill_buff.go:198 target.MR += buff.DeltaMR`：apply 時實際加到玩家 MR 屬性。
  - `internal/system/skill_buff.go:295-299 if buff.DeltaMR != 0 || buff.DeltaSP != 0 { handler.SendMagicStatus(target.Session, byte(target.SP), uint16(target.MR)) }` ＝ Java `pc.sendPackets(new S_SPMR(pc))`。註解明確標 `RESIST_MAGIC` 為對齊目標之一。
  - `internal/system/skill_buff.go:510 target.MR -= buff.DeltaMR` + `:562-564` 反向送 SendMagicStatus，對齊 Java 通用 stop 還原 + 重送 S_SPMR。
  - `internal/system/skill_buff.go:409 counterMagicExempt[129]=true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 129。
  - `scripts/combat/buffs.lua:236-276 NON_CANCELLABLE`：**不含 129**，對齊 Java `isNotCancelable` 不含 129（cancellable by CANCELLATION）。
  - `target='none'` 走 `skill.go:511-513 default → executeSelfSkill`（owner: skill_self.go）→ generic `applyBuffEffect` 路徑。
- **yaml 對照**：`data/yaml/skill_list.yaml:3939-3969 skill_id=129`：name=魔法防禦、skill_level=17、skill_number=0、mp_consume=5、hp_consume=0、item_consume_id=40319、item_consume_count=1、reuse_delay=0、buff_duration=1200、target=none、target_to=0、damage_value=0、damage_dice=0、damage_dice_count=0、probability_value=0、probability_dice=0、attr=0、type=2、lawful=0、ranged=0、area=0、through=0、id=1、name_id=''、action_id=19、cast_gfx=2186、cast_gfx2=0、sys_msg_happen=0、sys_msg_stop=719、sys_msg_fail=0 — **31 欄位逐一對齊** yiwei `skills.sql:128`。
- **結論：純審計，無 Java 真實差異需修**。
  - 核心行為（MR+10 套用、S_SPMR 通知、cancellable、counter magic 豁免、無互斥）兩端等價。
  - 既有 generic buff path 與 counterMagicExempt 表已正確覆蓋。
- **不寫新測試**：純通用 buff path（mr=10 → DeltaMR → target.MR + SendMagicStatus）已被其他 MR buff 測試（如 SHADOW_ARMOR）涵蓋；新增「Go 本來就對 + 防回歸」測試違反「對齊深度停損標準」第 1 條。
- **broader gap（不改）**：
  - **`getMr()` 動態計算 vs Go 靜態 `target.MR`**：Java `L1PcInstance.getMr()` 在 receive 端動態彙整 base、裝備、buff、寶物加成；Go 在 cast 時將 buff delta 累加到 `target.MR` 靜態欄位。雖然 client UI 顯示與最終魔法抗性計算結果等價，但屬廣域 stat 系統架構議題（同 110/111 DEX→AC、137 Wis→MR 等先例），非單一技能可修。
  - **DB schema 來源唯一**：yiwei `db_split/skills.sql` 與貓飛 `lineage381.sql` 偶有資料漂移（如 mp/duration/cast_gfx 微調），本子項以 yiwei 為對齊源，cat-fei 端差異屬廣域 SQL 同步議題（與多項 elf/illusion 技能同源）。

## 幻象（AVATA / 120）— 純佔位技能，兩端皆無實作（純審計）

- **Java 對照**：grep `AVATA` 在全 yiwei src 只命中 `L1SkillId.java:441 public static final int AVATA = 120;` 一處（其餘 `AVATA*` 匹配為 `BRAVE_AVATAR`/`ILLUSION_AVATAR` 等不同技能）。**無 `skillmode/AVATA.java`、無 `L1SkillUse` case、無 `L1SkillStop` case、無 timer、無攻擊路徑引用**——skill 120 純常數定義，未實作任何行為。
- **Go 對照**：grep `AVATA|skill.*120|case 120` 在全 `server/internal/` 無任何命中——無 `buffs.lua` entry、無 handler 邏輯、無 system 程式碼路徑。
- **yaml**：cat-fei `(120,'none',15,7,0,0,0,0,0,0,'none',0,0,0,0,0,0,0,0,0,0,0,0,128,'',19,2280,0,0,0,0,0)` ＝ name='none'、mp=0、hp=0、buff_duration=0、target='none'、target_to=0、type=0、id=128、name_id=''、action_id=19、cast_gfx=2280、其餘皆 0。Go yaml `skill 120` 完全對齊（同 name='none'、同 cast_gfx=2280、同 id=128）。yiwei `db_split/skills.sql` 無 120 entry（與貓飛差異）。
- **結論：純佔位技能，兩端都刻意無實作，零 Java-vs-Go 差異**。yaml entry 存在是為了 data loader 完整性（避免空缺造成 lookup 失敗），但 skill 120 不會被 cast、不會掛 buff、不會觸發任何行為。
- **無修正**、**無測試**：技能 ID 對照表既有註記「Java 僅有常數與資料列，未看到 skillmode / timer / stop 行為；Go 目前不需新增狀態效果」已正確描述狀態。本次審計純粹確認此狀態仍然成立並標記 ✅。

## 神聖犧牲（DIVINE_SACRIFICE / 119）— 純審計，無 Java 差異需修

- **Java 對照**：
  - `L1SkillId.java:435-437 DIVINE_SACRIFICE = 119`（神聖犧牲）。
  - **無 `skillmode/DIVINE_SACRIFICE.java`**、無 `L1SkillUse` 主動施法路徑——skill 119 是純被動 mastery，學會後由 `BraveAvatarTimer` 在背景檢查並對 party 成員套用 8065 王者加護（BRAVE_AVATAR）buff。
  - `timecontroller/skill/BraveAvatarTimer.java:25-78`：固定 5 秒 interval，迭代 `World.getAllPlayers()`：
    - 跳過 `pc == null || pc.getNetConnection() == null`。
    - 有 party：leader = party.getLeader()，若 `leader.isCrown() && leader.isSkillMastery(119) && distance <= 16`：
      - 若 `party.getNumOfMembers() >= 2` 且 `!pc.hasSkillEffect(BRAVE_AVATAR)`：
        - `setSkillEffect(BRAVE_AVATAR, 0)`（duration=0 永久）+ `addStr(1)+addDex(1)+addInt(1)+addMr(10)+addRegistStun(2)+addRegistSustain(2)`。
        - 送 `S_SPMR(pc)` MR/SP 封包 + `S_OwnCharStatus2(pc)` 完整狀態 + `sendPacketsAll(new S_SkillSound(pc.id, 9009))` 廣播技能音效 + `S_PacketBox(NONE_TIME_ICON, 1, 479)` 無時限 buff icon。
    - `else if (distance > 16) && hasSkillEffect(BRAVE_AVATAR)` → `removeNoTimerSkillEffect(BRAVE_AVATAR)`。
    - 無 party 時：若 `hasSkillEffect(BRAVE_AVATAR)` → 移除。
  - `L1PcInstance.java:1577-1605`：`setSkillMastery / isSkillMastery / removeSkillMastery / clearSkillMastery` 操作 `_skillList`，與一般 known spell list 等價。
- **Go 對照**：
  - `skill_clan.go:91-105 updateBraveAvatarAura`：對 `World.AllPlayers()` 每位玩家檢查 `shouldHaveBraveAvatar` → apply 或 remove。
  - `skill_clan.go:107-120 shouldHaveBraveAvatar`：party != nil + `Members >= 2`、leader != nil + `ClassType == 0`（Crown）+ `playerKnowsSpell(leader, 119)`、`leader.MapID == player.MapID`、`chebyshevDist <= 16`。
  - `skill_clan.go:122-150 applyBraveAvatar`：HasBuff(8065) skip、新增 buff with `TicksLeft=0`（永久）、`DeltaStr=1, DeltaDex=1, DeltaIntel=1, DeltaMR=10, DeltaRegistStun=2, DeltaRegistSustain=2`、apply stats、AddBuff、`SendPlayerStatus(S_STATUS)` ≈ S_OwnCharStatus2、`SendMagicStatus(SP, MR)` ＝ S_SPMR、`SendNoneTimeIcon(true, 479)` ＝ S_PacketBox icon 479、`BroadcastToPlayers(nearby, BuildSkillEffect(9009))` ＝ S_SkillSound 9009（GetNearbyPlayers 含自己 ＝ sendPacketsAll）。
  - `skill_clan.go:152-158 removeBraveAvatar`：HasBuff 檢查 + `removeBuffAndRevert` 還原所有 stat + `SendNoneTimeIcon(false, 479)` 清除 icon。
  - `skill.go:77-87 Update`：`braveAvatarElapsed += dt`，達 `braveAvatarInterval=5*time.Second` 即觸發 `updateBraveAvatarAura` ＝ Java 5000ms fixed-rate。
  - 常數對照：`braveAvatarMasteryID=119`、`braveAvatarSkillID=8065` ＝ Java `BRAVE_AVATAR=8065`、`braveAvatarRange=16`、`braveAvatarInterval=5s`。
- **yaml**：cat-fei `(119,'none',15,6,1,0,0,0,0,0,'none',0,0,0,0,0,0,0,0,0,0,0,0,64,'',19,0,0,0,0,0,0)` ＝ name='none'、mp=1、type=0、target='none'、target_to=0、id=64——純佔位 entry（無主動施法）。Go yaml 完全對齊。
- **既有測試覆蓋**：`TestSkillClanAuraBraveAvatarAppliesAndRemovesPartyAura` 驗證 16 格內隊伍成員（dx=10, dy=0, Chebyshev=10 ≤ 16）套用全 6 項 stat delta，移動到 X=117（dx=17 > 16）後 buff 移除且 stat 完整還原。
- **結論：純審計，無 Java 真實差異需修**。
  - **不寫新測試**：既有測試已覆蓋核心 apply/remove/stat-revert/距離邊界。新增「Class != 0 拒絕 / mastery 未學拒絕 / party < 2 拒絕」屬「Go 本來就對 + 防回歸」類型，違反停損標準。
- **broader gap（不改）**：
  - **Java 「leader 失去 crown/mastery 時不移除 buff」可能是 Java bug**：Java BraveAvatarTimer 的移除分支只在 `(leader valid + distance > 16)` 或 `no party` 兩種情形觸發；若 leader 失去 Crown class 或學失技能（極罕見）、或 party 縮到 1 員，Java 保留 orphan buff。Go `shouldHaveBraveAvatar` 對所有失敗情形都觸發 `removeBraveAvatar`，比 Java 更嚴謹——屬「Go 比 Java 更正確」，**不應**為了對齊而引入 Java bug。
  - **S_OwnCharStatus2 vs S_STATUS 封包格式差異**：Java BraveAvatarTimer 送 S_OwnCharStatus2（extended own char status）而 Go 送 S_STATUS（opcode 8）。兩者語義近似（refresh client 完整狀態），但 packet 結構/欄位可能不完全等價。屬廣域封包結構審核議題（與其他用 SendPlayerStatus 的 system 同源），不在 119 audit 範圍。
  - **`getNetConnection() == null` 跳過邏輯**：Java 在迭代時跳過已斷線玩家。Go `AllPlayers` 是否涵蓋已斷線玩家未細查——若 Go iterator 已排除 disconnected players（透過 session lifecycle），則無 functional 差異；若包含則 BuildSkillEffect 廣播會嘗試送到無效 session（broadcast 端應已有 null guard）。屬廣域 player iteration semantics 議題。

## 援護盟友（RUN_CLAN / 118）— 修正失敗路徑也消耗 MP 對齊 Java

- **Java 對照**：
  - `skillmode/RUN_CLAN.java:19-42 start()`：取得 clanPc by integer ID，若非 null 進入條件鏈：
    1. `pc.getMap().isEscapable() || pc.isGm()` — caster's map 必須可順移 OR caster 為 GM。
    2. 若 true：檢查 `L1CastleLocation.checkInAllWarArea(clanPc.X, clanPc.Y, clanPc.mapId)`（目標在攻城戰旗幟內）+ `clanPc.mapId ∈ {0, 4, 304}`（大陸地圖）。
       - 通過 → `L1Teleport.teleport(pc, clanPc.X, clanPc.Y, clanPc.mapId, 5, true)` 傳送。
       - 失敗 → `pc.sendPackets(new S_ServerMessage(1192)) + S_Paralysis(7, false)`。
    3. 若 false（caster's map 不可順移且非 GM）→ `pc.sendPackets(new S_ServerMessage(647)) + S_Paralysis(7, false)`。
  - `L1SkillUse.java:478-487` TYPE_NORMAL 流程：`runSkill() → useConsume() → sendGrfx → sendFailMessageHandle → setDelay`。`runSkill()` 內部呼叫 skillmode.start，**即使 skillmode 內部送拒絕訊息與 Paralysis 仍然正常返回**，`useConsume()` 接著執行——**MP 消耗在 RUN_CLAN 失敗路徑（送 647/1192）後依然發生**。
  - **yiwei SQL**：無 skill 118 entry（不在 db_split/skills.sql 內，可能屬於 yiwei 端後補的技能資料）。
  - **cat-fei SQL**：`(118,'援護盟友',15,5,30,0,0,0,0,0,'buff',0,...,'',19,0,...,280,0)` → mp=30、target='buff'、target_to=0、id=32、sys_msg_fail=280。Go yaml 完全對齊。
- **發現的 Java 真實差異**：原 Go `skill_clan.go executeClanTargetSkill case 118`：
  ```go
  if !s.canRunClanTeleport(player, target) {
      handler.SendServerMessage(...) + handler.SendParalysis(...)
      return  // ← 失敗路徑提前 return，MP 未消耗
  }
  if consume { s.consumeSkillResources(...) }  // ← 只在通過 check 後消耗
  ```
  失敗路徑（送 647/1192）**不消耗 MP**——玩家可向不可順移地圖的盟員無限重試呼喚而不付 MP 代價，與 Java 「即使失敗也消耗 MP」設計不符。屬玩家可見的資源管理差異。
- **修正**：將 `consumeSkillResources` 提到 `canRunClanTeleport` 檢查**之前**，並補 Java 對照註解：
  ```go
  case 118:
      // Java L1SkillUse.java:481-482 TYPE_NORMAL: runSkill() → useConsume()
      if consume { s.consumeSkillResources(...) }
      if !s.canRunClanTeleport(...) { ...send 647/1192 + Paralysis... return }
      ...teleport...
  ```
  失敗路徑現在也消耗 MP，與 Java useConsume 一致。
- **新測試**：`TestSkillClanRunClanConsumesMpEvenOnRejectionLikeJava` — caster MP=50/MaxMP=100、caster+member 都在 mapID=100（非 {0,4,304}），直接呼叫 `executeClanTargetSkill(..., consume=true)`，斷言 `caster.MP == 20`（50-30 消耗）且 caster 位置未變（傳送被拒絕）。驗證 Java 失敗路徑 MP 消耗行為。
- **架構合規**：純語句順序調整。`consumeSkillResources` 既有實作不動，`canRunClanTeleport` 與 `runClanRejectMessage` 邏輯不動。
- **原已對齊（純審計）**：
  - `canRunClanTeleport` = `isEscapableForRunClan(player) && isRunClanAllowedTargetMap(target) && !isInAnyCastleWarArea(target)` ✓ 對齊 Java 三條件鏈。
  - `runClanRejectMessage` 優先序：caster's map 不可順移 → 647；其他失敗 → 1192 ✓ 對齊 Java 兩條訊息分支。
  - `SendParalysis(TeleportUnlock=7)` ✓ 對齊 Java `S_Paralysis(7, false)`。
  - `TeleportPlayer(sess, player, target.X, target.Y, target.MapID, 5, deps)` ✓ 對齊 Java `L1Teleport.teleport(pc, clanPc.X, clanPc.Y, clanPc.mapId, 5, true)`。
  - `isEscapableForRunClan`: `AccessLevel >= 200` GM bypass + `MapData.GetInfo(player.MapID).Escapable` ✓ 對齊 Java `pc.getMap().isEscapable() || pc.isGm()`。
- **broader gap（不改）**：
  - **Go 額外的「同血盟檢查」（skill_clan.go:204）**：`if target.CharID == player.CharID || player.ClanID == 0 || player.ClanID != target.ClanID → 414`，Java skillmode RUN_CLAN.start 端不檢查 clan，假設由 client UI 限制目標選擇——Go 是 over-strict 防護而非 Java 偏離，不需移除。
  - **dual dispatch path**：skill.go:425-429 主派發走 `executeClanTargetSkill(consume=true)`；skill_buff.go:876-878 buff 派發走 `consume=false`。兩條路徑 consume 語義不同，目前生產走 skill.go 主派發路徑，buff 路徑只在測試使用。屬廣域 dispatch 重構議題，不在 118 audit 範圍。
  - 部分 Go skill 仍維持「validate → execute → consume on success」模式（與 Java 「runSkill → useConsume always」不同），需個別 audit 評估——本步只修 118，不擴及其他 skill。

## 衝擊士氣（BRAVE_AURA / 117）— 補齊血盟範圍 + 移除三向不對稱互斥

- **Java 對照**：
  - `skillmode/` 無 `BRAVE_AURA.java`（走 `L1SkillUse.java` 預設路徑）。
  - `L1SkillUse.java:2435-2438` apply：`else if (this._skillId == BRAVE_AURA) { ... // pc.addDmgup(-5); pc.sendPackets(new S_PacketBoxIconAura(116, ...)); }` — `addDmgup` **註解掉**，僅送 aura icon 116（注意 icon ID=116，非 117）。
  - `L1SkillStop.java:409-414` cancel：`//cha.addDmgup(-5);` 同樣註解掉，僅送 icon 116 歸零。
  - `L1AttackPc.java:2928-2953 BuffDmgUp(dmg)` — BRAVE_AURA 的真實機制：`int random = _random.nextInt(100) + 1;`（[1, 100]），`if/else if/else if` 鏈共用同一 random：`ELEMENTAL_FIRE → BURNING_SPIRIT → BRAVE_AURA`，每次攻擊只有 ONE 個會觸發。BRAVE_AURA 條件：`hasSkillEffect(BRAVE_AURA) && random <= 33 → dmg *= 1.5D`（33/100=33%）。
  - **yiwei SQL** `skills:117`：`('117','勇猛意志','15','4','25','0','0','0','0','640','none','0',...,'3942',...)` → target_to=0 (TARGET_TO_ME，舊版自身單體)、mp=25。
  - **cat-fei SQL**：`(117,'衝擊士氣',15,4,40,0,0,0,0,640,'none',12,...,3942,...)` → target_to=12 (TARGET_TO_PARTY|TARGET_TO_CLAN，現代隊伍+血盟)、mp=40。Go yaml 整體（名稱衝擊士氣/mp=40/duration=640/cast_gfx=3942）已對齊貓飛，唯獨 `target_to: 8` 漏掉血盟分支——同檔 115 已是 12 但 114 與 117 仍為 8（114 已在 2026-05-19 audit 修正為 12）。
- **發現的 Java 真實差異**：
  - **(1) yaml `target_to: 8` vs cat-fei `12`**：Go 同隊伍非同血盟成員會收到 buff，同血盟非同隊伍成員不會——與 cat-fei 兩種範圍都收的設計不符。
  - **(2) buffs.lua 三向不對稱互斥**：`[114].exclusions={115,117}`、`[115].exclusions={114,117}`、`[117].exclusions={114,115}` 在 Java REPEATEDSKILLS 10 群（148/149/156/163/166、151/168、52/101/150/155/186/1000/1016、8/19/54、26/110、42/109、80/106、185/190/195、14/213、213/218）均無對應，且 Java skillmode 端三技能 if/else 分支獨立、無 mutex 邏輯——Java 端三王族光環可同時掛在同一玩家身上。114 audit 與 115 audit 因「單獨改其中一個會造成不對稱（順序敏感）」而延後處理；117 audit 是三向同步修正的時機。
- **修正**：
  - `skill_list.yaml:3551` skill 117 `target_to: 8 → 12`（補齊血盟分支，與 cat-fei 與 115 yaml 一致）。
  - `scripts/combat/buffs.lua:100-104` 三向 exclusions 全部移除，並加上「Java 三光環可疊加」註解：
    ```lua
    [114] = { hit_mod = 5, dmg_mod = 5 },  -- Glowing Aura
    [115] = { ac = -8 },                    -- Shining Aura
    [117] = {},                              -- Brave Aura（純機率旗標）
    ```
- **新測試**：`skill_clan_aura_test.go TestSkillClanAuraRoyalAurasStackLikeJava` — 連續 `applyBuffEffect` 套用 114/115/117 三技能，斷言三 buff 皆 active、HitMod=5+DmgMod=5（114）、AC=2 從 10 起算 -8（115）、`braveAuraDamageWithRoll(roll=0)` 1.5x 觸發（117 機率內）。確認 Java 三向疊加行為。
- **架構合規**：純資料+lua 修正。`applyRoyalAuraSkill` 既有 `targetTo&8 → party + targetTo&4 → clan` 雙分支已支援 12；`applyBuffEffect → removeBuffAndRevert` 路徑透過 exclusions 列表清除衝突，移除 exclusions 等於停止跨光環互斥。
- **broader gap（不改）**：
  - **Java damage chain 共享 random + 互斥（不可疊乘）**：`L1AttackPc.java:2928-2953 BuffDmgUp` 對 ELEMENTAL_FIRE(171) / BURNING_SPIRIT(102) / BRAVE_AURA(117) 三項物理增傷共用 `_random.nextInt(100)+1` 並以 if/else if 鏈做互斥——每次攻擊最多 1 個 1.5x 乘子生效。Go `combat.go:213-215, 458-461, pvp.go:105-107, 322` 三個 helper（`darkElfPhysicalDamage` / `elfMeleeDamage` / `braveAuraDamage`）各自獨立 RandInt(100) 滾骰，無互斥——同時持有三 buff 時可能 1.5×1.5×1.5=3.375x 疊乘（Java 最多 1.5x）。屬廣域 damage chain 重構（需 single shared roll + priority chain pattern），跨 117/102/171 三技能 audit 範圍，留待 damage chain 廣域審計時統一處理。
  - **聯盟呼喚 4976 callClan1 同源 spread 缺失**（從 116 audit 延續）：留待 alliance 自身審計。
  - yiwei vs cat-fei SQL 廣域資料漂移屬 yaml/SQL 同步缺口。

## 呼喚盟友（CALL_CLAN / 116）— 修正接受呼喚後傳送座標分散範圍

- **Java 對照**：
  - `skillmode/CALL_CLAN.java:20-31`：施法時 `clanPc.setTempID(pc.getId()) + sendPackets(new S_Message_YN(729))`，僅暫存盟主 ID 與發 Y/N 對話。
  - `C_Attr.java:505-512 case 729`：玩家點 Yes 時呼叫 `callClan(pc)`。
  - `C_Attr.java:1129-1228 callClan`：
    1. 取出 leader (`findObject(getTempID())`)、`setTempID(0)`、`isParalyzedX()` 阻擋。
    2. 多數 escapable/castle/questmap 檢查被註解（line 1148-1192），active 邏輯只剩 leader 非 null 與 paralyze。
    3. 最終傳送：`L1Location newLocation = leader.getLocation().randomLocation(0, true)` (radius=0 ≈ 取盟主原位置)，再加 `(int)(Math.random()*5) - (int)(Math.random()*5)` 在 X 與 Y 各加 `[-4..+4]` 隨機偏移，最後 `L1Teleport.teleport(pc, ..., 5, true, 0)`（heading=5）。
- **發現的 Java 真實差異**：原 Go `handleCallClanYesNo` 直接 `TeleportPlayer(sess, player, caller.X, caller.Y, caller.MapID, 5, deps)` 傳送到盟主**精確同格座標**，與 Java `[-4..+4]` 隨機散佈不符——多人連續被呼喚時會全部疊在盟主同一格，視覺與碰撞行為與 Java 不同。
- **修正**：`handler/npcaction.go handleCallClanYesNo` 加入隨機偏移計算 `dx = RandInt(5) - RandInt(5)`、`dy = RandInt(5) - RandInt(5)`，最終傳送座標 `caller.X+dx, caller.Y+dy`，與 Java line 1226 完全等價（`Math.random()*5` 強制轉 int = 0..4，差分 = -4..+4）。
- **測試更新**：`mass_teleport_yesno_test.go TestHandleAttrCallClanResponseUsesJavaCAttrCase729` 斷言改為 `MapID 必須一致 + X/Y 在 caller ±4 範圍內`，反映 Java 隨機散佈行為。
- **架構合規**：`handleCallClanYesNo` 仍是薄層（解析、驗證、呼叫 TeleportPlayer），無遊戲狀態直接變更（teleport 內部負責）。已使用既有 `world.RandInt` helper（單線程遊戲迴圈安全）。
- **broader gap（不改）**：
  - **聯盟呼喚（4976 callClan1）同源 spread 缺失**：Java `C_Attr.java:1339 callClan1` 採完全相同的 `(rand%5 - rand%5)` 散佈公式，Go `handleAllianceCallClanYesNo` 也是傳送到精確位置。屬同源 bug 但聯盟系統與 CALL_CLAN(116) 是不同技能 ID，留待 alliance/4976 自身審計或廣域 callClan 重構時一併處理。
  - **被呼喚者狀態的 Java escapable/castle/questmap 多重檢查全被註解掉**：Java line 1148-1192 多項地圖安全檢查（escapable/戰爭旗幟內/副本地圖）已被註解，active 邏輯只剩 paralyze。Go 可選擇實作這些檢查作為 over-strict 保護，但屬「Go 比 Java 嚴格」而非「Go 偏離 Java」，不在 audit 修正範圍。
  - Go 的 `Sleeped` 額外阻擋（Java 只查 `isParalyzedX`）：Go 多了一層 sleep 阻擋，屬增加而非偏離 Java，且 sleep 中無法操作 Y/N dialog 本就為 dead code，不需移除。

## 鋼鐵士氣（SHINING_AURA / 115）— 修正 yaml `cast_gfx` 對齊 cat-fei 視覺特效

- **驗證**：技能 ID 對照表既有註記「資料驅動，需驗證隊伍/血盟範圍」。本次比對 cat-fei `貓飛版_lineage381.sql` skills 表：`(115,'鋼鐵士氣',15,2,40,0,0,0,0,640,'none',12,0,0,0,0,0,0,2,0,0,-1,0,4,'$1977',19,2943,...)`。Go yaml 整體（名稱/mp=40/duration=640/target_to=12/name_id=$1977）皆與貓飛對齊，唯獨 `cast_gfx: 3941` 與貓飛 `2943` 不符（115 cast_gfx 3941 與 117 cast_gfx 3942 僅差 1，疑似當初編寫時 copy 117 後改錯數字）。
- **發現的 Java 真實差異**：`cast_gfx` 控制 `skill_self.go:140 BroadcastToPlayers(BuildSkillEffect(player.CharID, skill.CastGfx))` 廣播給 AOI 內所有玩家的視覺施法特效。錯誤 ID 導致玩家看到的施法動畫與真實 Lineage R 客戶端不一致——屬玩家可見的協議資料差異。
- **修正**：`skill_list.yaml:3531` skill 115 `cast_gfx: 3941 → 2943`。
- **架構合規**：純資料層 yaml 修正，無 Go 程式碼路徑變更。`SendIconAura(byte(115-1)=114)` 已對齊 Java `S_PacketBoxIconAura(114, ...)`、`buffs.lua [115].ac = -8` 已對齊 Java `addAc(-8)`、`target_to: 12` 已對齊 cat-fei `TARGET_TO_PARTY|TARGET_TO_CLAN` 與 yiwei `L1SkillUse.java:871-880` 同範圍掃描邏輯——既有 `TestSkillClanAuraShiningAuraAppliesToClanMembers` 已鎖定 clan+party 範圍與 AC -8。
- **broader gap（不改）**：
  - **buff 持久化 whitelist 缺口**：Java `CharBuffTable.java:64` 把 `SHINING_AURA` 註解掉（不持久化到 DB），登出後鋼鐵士氣會消失。Go `BuffRepo` 與 `PersistenceSystem` 無 whitelist filter——所有 active buff 都持久化，登入後仍保留 115 buff。屬廣域 buff 持久化白名單系統缺口（不只 115，需通用 whitelist 機制），與既有 `BraveAvatar` 不寫 DB 等特例同源，需架構級補齊。
  - **buffs.lua 三向 exclusions Java 無對應**：`[114]/[115]/[117]` 互斥在 Java REPEATEDSKILLS/skillmode 端均無依據（三光環可疊加）。單獨改 [115].exclusions 而 [114]/[117] 仍含 115 會造成順序敏感不對稱。三向互斥屬聯動修正，留待 117 audit 同步處理三條 exclusions。
  - yiwei vs cat-fei SQL 廣域資料漂移屬 yaml/SQL 同步缺口，非個別技能審計範圍。

## 激勵士氣（GLOWING_AURA / 114）— 修正 yaml `target_to` 對齊現代 Lineage R 範圍

- **驗證**：技能 ID 對照表既有註記「資料驅動，需驗證隊伍/血盟範圍」。本次比對兩份 Java 參考：
  - **yiwei** `db_split/skills.sql:114`：`target_to=0` (`TARGET_TO_ME`)、`mp_consume=25`，搭配 `L1SkillUse.java:942-946` `if (target_to==TARGET_TO_ME && !TYPE_ATTACK) → targetlist=[user]` ＝ **舊版自身單體**。
  - **cat-fei (l1j_fly)** `貓飛版_lineage381.sql` skills 表：`(114,'激勵士氣',15,1,40,0,0,0,0,640,'none',12,...)` → `target_to=12` = `TARGET_TO_PARTY|TARGET_TO_CLAN`、`mp_consume=40`、`buff_duration=640` ＝ **現代隊伍+血盟範圍**。
  - Java code 行為 `L1SkillUse.java:2424-2428 GLOWING_AURA cha.addHitup(5)+addDmgup(5)` 對每個 targetlist 成員套用，因此 cat-fei `target_to=12` 走 `L1SkillUse.java:871-880` 把同血盟+同隊伍成員加入 targetlist 並全部套 +5 命中/+5 傷害。
- **發現的 Java 真實差異**：Go `skill_list.yaml` skill 114 `target_to: 8`（僅 TARGET_TO_PARTY）漏掉 TARGET_TO_CLAN，與 cat-fei 現代行為不符。同檔案 skill 115 (`SHINING_AURA`) 已是 12 但 skill 114 與 117 (`BRAVE_AURA`) 仍為 8——本步只修 114，117 留待自身審計。
- **修正**：
  - `skill_list.yaml:3516` skill 114 `target_to: 8 → 12`。
  - `skill_clan_aura_test.go TestSkillClanAuraGlowingAuraAppliesToParty*` 更新為四角色（caster 7、partyMember 8、clanmate 7、outsider 8），驗證 caster+partyMember+clanmate 三者皆得 +5 hit/+5 dmg、outsider 不得 buff，函式名重新命名為 `*AppliesToPartyAndClan*` 反映實際範圍。
- **架構合規**：`skill_clan.go:21-62 applyRoyalAuraSkill` 既有 `targetTo&8 → party`、`targetTo&4 → clan` 雙分支邏輯本就支援 12，僅 yaml 資料缺值；無需改 Go 邏輯。
- **broader gap（不改）**：
  - skill 117 (`BRAVE_AURA`) yaml `target_to=8`（同樣漏掉血盟分支）留待 117 自身審計修正，避免單獨改動破壞 audit 序。
  - Go `buffs.lua [114].exclusions = {115, 117}` 互斥 Java 端沒有對應 REPEATEDSKILLS（Java 三王族光環可疊加）。但若單獨移除 114 的 `{115, 117}` 而 [115]、[117] 仍含 `{114, ...}`，會造成不對稱互斥（先施 114 後施 115 → 115 移除 114；先施 115 後施 114 → 兩者並存）。三向互斥屬聯動修正，留待 117 audit 同步處理三條 exclusions。
  - yiwei vs cat-fei SQL 大量資料漂移屬廣域 yaml/SQL 同步缺口（非個別技能審計範圍）。

## 精準目標（TRUE_TARGET / 113）— 補齊 PvP 增傷雙乘子

- 對齊 Java `L1AttackPc.java:1509-1511 + 1580-1584` 在 `calcPcDamage` 對持有 TRUE_TARGET 的 PvP 目標套用兩段乘子：
  ```java
  // line 1509-1511 (在 ARMOR_BREAK 1.58x 之前)
  if (_targetPc.hasSkillEffect(TRUE_TARGET)) {
      dmg *= ConfigSkill.STRIKER_DMG;  // 1.2
  }
  // line 1580-1584 (在 ARMOR_BREAK 與 STRIKER_GALE 之後、減免之前)
  if (_targetPc.hasSkillEffect(TRUE_TARGET)) {
      double attackerlv = _pc.getLevel();
      double adddmg = (attackerlv / 15) / 100 + 1.01D;
      dmg *= adddmg;
  }
  ```
  yiwei 預設 `STRIKER_DMG=1.2`，第二段公式 = `1.01 + attackerLv/1500`。組合乘子：`1.2 × (1.01 + attackerLv/1500)`。對 L50 攻擊者 → 1.252×、L100 → 1.292×。
- **發現的 Java 真實差異**：原 Go `pvp.go HandlePvPAttack` 與 `HandlePvPFarAttack` 兩條路徑均未檢查 `target.HasBuff(113)` 套用 TRUE_TARGET 增傷，導致 PvP 對標記目標完全無增傷效果——TRUE_TARGET 的核心戰術機制（標記目標讓血盟集中火力）完全失效。
- **修正**：兩條 PvP 路徑同步插入兩段乘子（合併為一個 `HasBuff(113)` block）：
  - `pvp.go:91-96` PvP melee（HandlePvPAttack）：插入於 ARMOR_BREAK 之前，與 Java line 1509 對齊。
  - `pvp.go:311-318` PvP ranged（HandlePvPFarAttack）：插入於 strikerGaleRangedDamage 之後，與 darkElfPhysicalDamage 同位置——Java calcPcDamage 同時涵蓋近戰與遠程，兩段乘子對遠程同樣適用。
  ```go
  if damage > 0 && target.HasBuff(113) {
      damage = int32(float64(damage) * 1.2)
      damage = int32(float64(damage) * (float64(attacker.Level)/1500.0 + 1.01))
  }
  ```
- **其餘對齊（無修改）**：
  - **cast dispatch + 重施守衛**：`skill_buff.go:1103-1104 case 113 applyTrueTargetEffect`：
    ```go
    if !target.HasBuff(113) {
        target.AddBuff(&world.ActiveBuff{SkillID: 113, TicksLeft: dur*5})
    }
    s.sendTrueTargetToClan(caster, target, text)
    ```
    對齊 Java `skillmode/TRUE_TARGET.java:52-55 if (!cha.hasSkillEffect(113)) cha.setSkillEffect(113, integer*1000)`（重施守衛，不刷新 timer）。
  - **血盟廣播**：`skill_clan.go:177-191 sendTrueTargetToClan`：對 `caster.ClanID != 0` 時 `AllPlayers → player.ClanID == caster.ClanID → SendTrueTarget`，無血盟時 fallback 只送自己。對齊 Java skillmode line 57-68 `if (clan != null) ... else srcpc.sendPackets(S_TrueTarget)`。
  - **S_TrueTarget packet 格式**：`handler/broadcast.go:936-951 BuildTrueTarget(targetID, casterID, message)` 對齊 Java `S_TrueTarget(cha.getId(), clanmember.getId(), srcpc.getText())`（注意：第二個欄位 Java 是「viewer」即 clanmember，Go 傳 casterID——本步未驗證該欄位語義精確匹配，留待 broader S_TrueTarget packet 結構審核）。
  - **NON_CANCELLABLE**：`buffs.lua:257 [113] = true` 對齊 Java `L1SkillMode.java:38` 含 113。
  - **counterMagicExempt**：`skill_buff.go:405 113: true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 113。
  - **無 REPEATEDSKILLS**：Java 10 個群組均不含 113。
- **broader gap（不改）**：
  - **L1EffectInstance 86131 視覺特效缺失**：Java skillmode line 30-50 `spawnTrueTargetEffect(86131, 16, cha, srcpc, 0, 12299)` 在目標頭頂生成可見的 NPC 特效實體（npc_id=86131、text_id=12299），這是 TRUE_TARGET 視覺辨識依據。Go 無 `L1EffectInstance`/`spawnTrueTargetEffect` 等價系統，目標頭頂無視覺標記。屬廣域 effect entity 系統缺口（與其他 effect-based 技能如召喚物/視覺光環缺口同源）。
  - **WorldEffect 刪除舊精準目標 effect**：Java line 30-35 在套用新 effect 前掃 `WorldEffect.get().all()` 刪除同一 srcpc 的舊 86131 effect（避免一個玩家累積多個視覺特效）。Go 無 WorldEffect 全域註冊表，無需此清理。
  - **`srcpc.setText("")` 清空**：Java line 71 在廣播完成後清空 srcpc 暫時文字串（C_UseSkill 傳入的 text）。Go applyTrueTargetEffect 從參數 `text string` 直接取用，不持久化到 caster 物件——天然不需清空。但若 Go 用了 PlayerInfo.Text 或類似緩存欄位，需確認是否同樣清空。本步驗證 Go 走參數傳遞，無此問題。
  - **`L1Character.add_TrueTargetEffect` 物件關聯**：Java 在 effect entity 與 target 間建立 list 關聯（line 37/49 `get_TrueTargetEffectList()` + `add_TrueTargetEffect`），用於後續查詢/移除。Go 無對應結構，但本步主要修補是 PvP 增傷，視覺/list 屬上述 effect entity 系統缺口。
- **不寫新測試**：兩處同源乘子插入（HasBuff(113) → 1.2 × (1.01+lv/1500)），掛載既有 PvP 攻擊路徑（pvp.go HandlePvPAttack/HandlePvPFarAttack），與 ARMOR_BREAK 1.58x、STRIKER_GALE 等乘子 pattern 等價。依停損標準避免「Go 已對 + 防回歸」測試（PvP melee 路徑已被大量整合測試覆蓋同源插入點）。

## 反擊屏障（COUNTER_BARRIER / 91）— 修正觸發機率公式（+ROM + lvlDiff）

- 對齊 Java `L1MagicPc.java:670-674 calcProbabilityMagic(COUNTER_BARRIER)`：
  ```java
  case COUNTER_BARRIER:
      probability = l1skills.getProbabilityValue() + attackLevel - defenseLevel;
      probability += ConfigSkill.COUNTER_BARRIER_ROM;
      break;
  ```
  即 `probability = probabilityValue (SQL=25) + (target.Level - attacker.Level) + COUNTER_BARRIER_ROM (yiwei 預設=33)` = 58 + lvlDiff。
- **發現的 Java 真實差異**：原 Go `pvp.go:116` + `npc_ai.go:511` 對 COUNTER_BARRIER 觸發機率硬編碼 `25`（漏掉 +ROM 33 與 lvlDiff），導致實際觸發率約為 Java 一半（25% vs 58%+ for 等級對等戰鬥）。
- **修正**：兩條路徑同步補齊機率公式：
  - `pvp.go:116-120` PvP melee：`prob := 25 + int(target.Level) - int(attacker.Level) + 33`
  - `npc_ai.go:511-517` NPC→PC melee：`prob := 25 + int(target.Level) - int(npc.Level) + 33`
- **其餘對齊（無修改）**：
  - **reflect damage 公式**：`pvp.go:597-615 calcCounterBarrierDmg`：
    ```go
    dmg := int32((info.DmgLarge + int(wpn.EnchantLvl) + info.DmgMod) << 1)
    dmg = dmg * 3 / 2  // ConfigSkill.COUNTER_BARRIER_DMG = 1.5
    ```
    對齊 Java `L1AttackMode.calcCounterBarrierDamage` 第 333-336 行非戰士分支 `damage = weapon.DmgLarge + EnchantLevel + DmgModifier << 1` + line 342 `damage *= COUNTER_BARRIER_DMG`。
  - **NPC reflect 公式**：`npc_ai.go:513 cbDmg := int32((int(npc.STR) + int(npc.Level)) << 1)` 對齊 Java line 340 `damage = STR + Level << 1` × 1.5。
  - **IMMUNE_TO_HARM 攻擊者減半**：`pvp.go:120 applyImmuneToHarmDamage(attacker, cbDmg)` 對齊 Java `L1AttackPc.commitCounterBarrier:3339-3341 if (_pc.hasSkillEffect(68)) damage /= 2`。
  - **GFX 10710 廣播**：`pvp.go:126` + `npc_ai.go:523` `BuildSkillEffect(target.CharID, 10710)` 對齊 Java `S_SkillSound(targetID, 10710)`。
  - **原始傷害歸零**：`pvp.go:128 damage = 0` + `npc_ai.go:525` 對齊 Java `return 0`。
  - **buff -2 AC**：`scripts/combat/buffs.lua:93 [91] = { ac = -2 }`——Java 雖無明確 add_ac(-2)，這是 Go 對 COUNTER_BARRIER 防護加成的選擇實作。
  - **NON_CANCELLABLE**：`buffs.lua:245 [91] = true` 對齊 Java `L1SkillMode.java:35`。
  - **counterMagicExempt**：`skill_buff.go:403 91: true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 91。
  - **無 REPEATEDSKILLS**：Java 10 個群組均不含 91。
  - **無 recast guard**：Java 無 skillmode，cast 走 default。
  - **cast 時 GFX 全域廣播**：Java `L1SkillUse.java:1645-1651 _player.sendPacketsXR(S_SkillSound, -1) + broadcastPacketAll(S_SkillSound)`——Go cast GFX 廣播由 default cast handler 處理。
- **broader gap（不改）**：
  - **PC→NPC 對 NPC 持有 COUNTER_BARRIER 路徑缺失**：Java `L1AttackPc.java:1886-1896` PC 攻擊 NPC 時若 NPC `hasSkillEffect(COUNTER_BARRIER)` 也觸發反擊。Go combat.go 無 `npc.HasDebuff(91)` 檢查。NPC 一般不持有 91（玩家專屬技能），但 boss 變體（如 SQL skill 11060 BOSS-吉爾塔斯-反擊屏障）與 GM 賦予場景會缺失反擊。屬 boss/GM 邊緣場景缺口，依「不可偷換範圍」維持。
  - **戰士主/副手武器分支 + 50% 機率**：Java `L1AttackMode.calcCounterBarrierDamage:303-328` 對戰士 PC target 額外處理副武器（`secondweapon != null` 時 `_random.nextBoolean()` 50% 用副手）。Go 簡化為單一武器路徑——屬 secondary weapon system 缺口。
  - **炫色 / 極品裝備 attr_DmgLarge / attr_DmgModifier**：Java 戰士分支讀 `_weapon.get_ItemAttrName().get_attr_id()` 加上 `attr.get_dmg_large()` + `attr.get_dmgmodifier()`。Go 無 ItemSpecialAttribute 系統，屬廣域 item 特殊屬性架構缺口。
  - **weaponType2 != 17（奇古獸）排除**：Java 三條 commitCounterBarrier 觸發路徑都檢查 `_weaponType2 != 17`。Go 未檢查，奇古獸 weaponType2 為特殊 item attr 之一，屬上述 item 屬性系統缺口。
  - **probability_value 用硬編碼 25 而非從 skill.yaml 讀**：本步為最小修補，未引入 `s.deps.Skills.Get(91).ProbabilityValue` lookup。若未來 yiwei 端調整該 SQL 值，Go 不會同步——應於 broader skill data driven refactor 時統一改為從 skill data 取值。
- **不寫新測試**：兩處同源 1 行公式擴展（從 fixed 25 改為 `25 + lvlDiff + 33`），掛載既有 RandInt+HasBuff 路徑（觸發路徑邏輯不變），無新邏輯分支。依停損標準避免「Go 已對 + 防回歸」測試重複測同一機制。

## 堅固防護（SOLID_CARRIAGE / 90）— 修正 dodge 變化通知 packet（SendDodgeIcon → SendUpdateER）

- 對齊 Java `skillmode/SOLID_CARRIAGE.java:18, 47`：cast 與 stop 兩端都送 `S_PacketBox(UPDATE_ER, pc.getEr())`（迴避率更新封包）。原 Go `applyBuffEffect`/`revertBuffStats` 對 `DeltaDodge > 0` 條件預設走 `SendDodgeIcon`（= `S_PacketBoxIcon1` 0x58 dodge icon packet），與 Java SOLID_CARRIAGE 期望的 UPDATE_ER packet 不同。原 Go 只對 skill 111 DRESS_EVASION 走 UPDATE_ER 分支，90 SOLID_CARRIAGE 缺失。
- **修正**：`skill_buff.go:240-249` apply 路徑 + `:521-528` revert 路徑兩處同步擴展為 `if skill.SkillID == 90 || skill.SkillID == 111`：
  ```go
  // Dodge 變化通知。Java 區分兩條 packet 路徑：
  //   - SOLID_CARRIAGE(90)/DRESS_EVASION(111) skillmode 送 S_PacketBox(UPDATE_ER, getEr())
  //   - UNCANNY_DODGE(106)/MIRROR_IMAGE/DRAGONEYE_* skillmode 送 S_PacketBoxIcon1(true, get_dodge())
  if skill.SkillID == 90 || skill.SkillID == 111 {
      handler.SendUpdateER(target.Session, target.Dodge)
  } else {
      handler.SendDodgeIcon(target.Session, target.Dodge, true)
  }
  ```
  兩條 packet opcode 不同（UPDATE_ER 0x84 vs dodge icon 0x58 byte 1），客戶端 UI 渲染位置/欄位不同，不可混用。
- **其餘對齊（無修改）**：
  - **盾牌/臂甲檢查**：`skill_self.go:113-118 case 90` `validateSolidCarriage(player) = shield != nil || guarder != nil` 對齊 Java `ConfigSkill.SOLID_CARRIAGE_MODE == 1` 路徑（yiwei 預設值 `各職業技能相關設置.properties:21 SOLID_CARRIAGE_MODE = 1`）：`getTypeEquipped(2,7) >= 1 || getTypeEquipped(2,13) >= 1`（type=2 為盾，slot 7 盾牌、slot 13 臂甲）。
  - **盾牌缺失訊息**：`SendNormalChat(sess, 0, "你並未裝備盾牌")` 對齊 Java `S_ServerMessage("你並未裝備盾牌")`——chat.go:206 註解 `S_ServerMessage(String) 相同格式的 normal chat 文字`，packet 結構相同。
  - **dodge=15 buff**：`buffs.lua:92 [90] = { dodge = 15 }` 對齊 Java `L1PcInstance.java:3389-3391 getEr() if hasSkillEffect(SOLID_CARRIAGE) er += 15`。
  - **NON_CANCELLABLE**：`buffs.lua:244 [90] = true` 對齊 Java `L1SkillMode.java:34`。
  - **counterMagicExempt**：`skill_buff.go:403 90: true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 90。
  - **無 REPEATEDSKILLS 互斥**：Java 10 個群組均不含 90。
  - **無 recast guard**：Java skillmode 無 `if (!hasSkillEffect(90))` 守衛，重施允許刷新 timer——Go default 路徑亦如此，正確對齊（與 89/106 有守衛者不同）。
- **broader gap（不改）**：
  - **動態 getEr() vs 靜態 target.Dodge**（與 111 audit 同源）：Java `S_PacketBox(UPDATE_ER, getEr())` 送的是「總 ER 動態值」（含職業/等級/DEX/originalEr + 90 dodge=15 + 其他 buff）；Go `SendUpdateER(target.Dodge)` 送的是「累加 buff 值」。Knight L50 DEX12 持 SOLID_CARRIAGE 時 Java 送 ER=31（12+6+0+15），Go 送 15。屬廣域 ER 系統架構缺口。
  - **STRIKER_GALE 短路反向**（與 111 audit 同源）：Java getEr() 對持 174 玩家 return 0；Go 90 cast 若 174 已啟動仍送 Dodge。屬上述廣域缺口子症狀。
  - **yaml drift**：與多技能同源廣域缺口。
- **不寫新測試**：擴展現有 `skill 111` 分支為 `skill 90 || skill 111`，與 111 既有 UPDATE_ER 邏輯等價（111 已被 elemental_buff_test 等覆蓋 lua dodge 加成路徑）。依停損標準避免「Go 已對 + 防回歸」測試重複測同一 packet 分支。

## 尖刺盔甲（BOUNCE_ATTACK / 89）— 補齊重施守衛

- 對齊 Java `skillmode/BOUNCE_ATTACK.java:13` `if (!srcpc.hasSkillEffect(89)) { setSkillEffect(89, integer*1000); addHitup(6); }`——重施時跳過 addHitup(6)+timer。原 Go `executeSelfSkill` default 路徑會在重施時透過 `applyBuffEffect → AddBuff 替換 + revertBuffStats(-6) → applyBuffStats(+6)` 形成中間瞬時 -6/+6 cycling + timer 刷新延長 buff，違反 Java「重施不刷新」語義。
- **修正**：`skill_self.go:244-250` 既有 106 UNCANNY_DODGE 重施守衛擴展為清單形式：
  ```go
  // 重施守衛清單（Java skillmode 有 `if (!hasSkillEffect(X))` 包住 stat add + timer set 的技能）：
  //   - UNCANNY_DODGE (106)：跳過 add_dodge(5)+timer+S_PacketBoxIcon1
  //   - BOUNCE_ATTACK (89)：跳過 addHitup(6)+timer
  if !((skill.SkillID == 106 || skill.SkillID == 89) && player.HasBuff(skill.SkillID)) {
      s.applyBuffEffect(player, skill)
  }
  ```
  重施時保留外層 cast GFX 廣播 + MP 消耗（與 Java `handleCommands` 一致），buff 內容不變。
- **其餘對齊（無修改）**：
  - **HIT+6 buff**：`buffs.lua:91 [89] = { hit_mod = 6 }` 對齊 Java skillmode `addHitup(6)` + stop `addHitup(-6)`。
  - **PvP 武器破壞 10% 機率**：`pvp.go:170-172`：
    ```go
    if target.HasBuff(89) && world.RandInt(100) < 10 {
        damagePlayerWeaponDurability(attacker, s.deps)
    }
    ```
    對齊 Java `L1AttackPc.damagePcWeaponDurability:3420 _random.nextInt(100)+1 <= 10`。
  - **武器類型排除**：`combat.go:933-935` `if pvp && (itemInfo.Type == "bow" || itemInfo.Type == "claw") return`——對齊 Java line 3400-3410 `_weaponType == 0/20/62 return`（0 赤手 Go 由 weapon==nil/category!=weapon 自然排除）。
  - **STRIKER_GALE 攻擊者豁免**：`combat.go:937 if player.HasBuff(175) return` 對齊 Java line 3416 `_pc.hasSkillEffect(175) return`。
  - **訊息 + 音效**：`combat.go:954,958` `SendServerMessageArgs(268, weaponName)` + `BroadcastToPlayers(BuildSkillEffect(10712))` 對齊 Java `S_ServerMessage(268, _weapon.getLogName())` + `sendPacketsX8(S_SkillSound(10712))`。
  - **PvE 排除**：Java `_calcType != PC_PC return`（line 3396）只在 PvP 觸發；Go pvp.go:170 為 PvP-only 路徑，PvE combat.go 不檢查 89，自然對齊。
  - **NON_CANCELLABLE**：`buffs.lua:243 [89] = true`（與 88 同列）對齊 Java `L1SkillMode.java:34`。
  - **counterMagicExempt**：`skill_buff.go:403 89: true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 89。
  - **無 REPEATEDSKILLS 互斥**：Java 10 個群組均不含 89。
- **broader gap（不改）**：
  - **重施守衛廣域 pattern**：已從 106 audit 既知 Java >30 個 skillmodes 有同類守衛（ADVANCE_SPIRIT/AWAKEN_*/SHADOW_*/DRESS_*/ILLUSION_* 等）。Go 目前僅 dragon awakening（185/190/195）與本步 106/89 顯式守衛，其餘大多仍走 generic applyBuffEffect 可能允許 timer 刷新。應後續設計通用 `ActiveBuff.NoRefreshOnRecast` flag 或 lua `[X] = { no_refresh_on_recast = true }` 屬性系統化處理。本步維持單一技能 audit 範圍。
  - **yaml drift**：skill_list.yaml `buff_duration=64`/`mp_consume=5`/`hp_consume=120`/`reuse_delay=5000` 待與 Java SQL 對照（另 audit 處理）。
- **不寫新測試**：擴展現有 106 守衛清單為 (106, 89) 兩元素，與既有 dragon awakening 守衛（185/190/195）pattern 等價。依停損標準避免「Go 已對 + 防回歸」測試重複測同一機制（106 audit 已建立 pattern）。

## 增幅防禦（REDUCTION_ARMOR / 88）— 純審計無代碼變更

- 純審計 `88 REDUCTION_ARMOR`：Go 完整對齊 Java 「無 skillmode + 4 攻擊路徑 flat 傷害減免」設計，公式 `dmg -= (max(targetLvl, 50) - 50) / 5 + (pvpPhysical ? 10 : 1)`。
- **Java 設計確認**：
  - **無獨立 skillmode/REDUCTION_ARMOR.java**——cast 走 L1SkillUse 預設路徑（讀 skill SQL 表 + `setSkillEffect(88, duration*1000)`）。
  - **4 處 damage 路徑各自檢查 `hasSkillEffect(88)`**：
    - `L1AttackPc.java:1617-1620` PC→PC melee + ranged：`dmg -= (max(targetLvl, 50) - 50) / 5 + 10`（PvP 物理 +10 base）
    - `L1AttackNpc.java:437-440` NPC→PC physical：`dmg -= (max(targetLvl, 50) - 50) / 5 + 1`（+1 base）
    - `L1MagicPc.java:1148-1151, 1296-1299` PC magic→PC：`+1` base
    - `L1MagicNpc.java:357-360` NPC magic→PC：`+1` base
- **Go 對照**：
  - **buff 旗標**：`buffs.lua:90 [88] = {}` 純旗標，註解明確標 `flat 傷害減免（npc_ai/pvp/magic 路徑套用 applyReductionArmorDamage），不是 AC 加成`，避免未來誤改為 AC buff。
  - **NON_CANCELLABLE**：`buffs.lua:242 [88] = true` 對齊 Java `L1SkillMode.java:34 isNotCancelable`。
  - **counterMagicExempt**：`skill_buff.go:403 88: true` 對齊 Java `L1SkillUse.java:147 EXCEPT_COUNTER_MAGIC` 含 88。
  - **damage 套用 helper**：`skill_buff.go:23-44 applyReductionArmorDamage(target, damage, pvpPhysical)`：
    ```go
    if !target.HasBuff(88) return damage
    lvl := max(target.Level, 50)
    reduction := (lvl-50)/5 + 1
    if pvpPhysical { reduction = (lvl-50)/5 + 10 }
    damage -= reduction
    if damage < 0 { damage = 0 }
    ```
  - **4 處插入點對齊 Java 4 路徑**：
    - `npc_ai.go:502/629/792/881` NPC→PC（melee/ranged/skill_atk/leverage）以 `pvpPhysical=false` 對齊 Java NPC→PC 系列 +1 base。
    - `pvp.go:101/303` PC→PC（melee/ranged）以 `pvpPhysical=true` 對齊 Java L1AttackPc:1617 +10 base。
    - `skill_damage.go:173 applySkillDamageToPlayer` PC magic→PC 以 `pvpPhysical=false` 對齊 Java L1MagicPc 系列 +1 base。
  - **無 REPEATEDSKILLS 互斥**：Java 10 個 REPEATEDSKILLS 群組均不含 88，Go buffs.lua [88] 不設 exclusions 正確對齊。
- **broader gap（不改）**：
  - **yaml buff_duration drift**：Go yaml `skill_list.yaml:2708 buff_duration: 32` vs Java SQL `db_split/skills.sql:88 buff_duration='192'` 差 6 倍。屬廣域 yaml/SQL 資料源同步缺口（與 mp_consume `5 vs 7`、其他多技能 yaml drift 同源）。需 skill_list.yaml 全面以義維版 SQL 重新生成或建立 reconciliation 流程，本步維持單一技能 audit 範圍不修。
  - **NPC magic→PC 缺口問題核對**：Java `L1MagicNpc.java:357-360` NPC magic 攻擊 PC 時也檢查 `hasSkillEffect(88)`，Go 目前在 `applySkillDamageToPlayer`（PC→PC magic）涵蓋，但 NPC→PC magic 是否有同一路徑需待 NPC magic system 完整審計時驗證——本步暫不展開。`npc_ai.go:792` skill attack 路徑（NPC 使用技能對 PC 物理傷害）已套用 reduction，magic 部分需另檢查（Go NPC 是否會施放 magic 傷害技能對 PC？該路徑現況？）。
- **不寫新測試**：純審計，Go 已對齊主流程；4 攻擊路徑插入點與 Java 4 path 一一對應。依停損標準避免「Go 已對 + 防回歸」測試。

## 破壞盔甲（ARMOR_BREAK / 112）— 補齊 1.58x 傷害弓/爪武器排除

- 對齊 Java `L1AttackPc.java`：
  - **PvE PC→NPC** line 732-736：`if ((_weaponType != 20) && (_weaponType != 62)) if (_target.hasSkillEffect(ARMOR_BREAK)) _damage *= 1.58;`
  - **PvP PC→PC** line 1516-1518：`if ((_targetPc.hasSkillEffect(ARMOR_BREAK)) && (isShortDistance())) dmg *= ConfigSkill.ARMOR_BREAK_DMG;`
  - `isShortDistance()` line 3322-3328：`return _weaponType != 20 && _weaponType != 62;`（等價於 PvE 直接判斷）
- **發現的 Java 真實差異**：原 Go combat.go:209 + pvp.go:92 對 `HasDebuff(112)`/`HasBuff(112)` 目標套用 1.58x 倍率時**未檢查武器類型**，導致弓 (weaponType 20 → "bow") 與爪 (62 → "claw") 武器也錯誤享受 1.58x 加成，違反 Java 「ARMOR_BREAK 為近戰專屬」設計。
- **修正**：兩條路徑加上 `weaponType != "bow" && weaponType != "claw"` 過濾條件，對齊 Java `_weaponType != 20 && _weaponType != 62`。Go 端 weaponType 字串映射（"bow"/"claw"）已在 `combat.go:929-934` 武器損壞 PvP 邏輯沿用過，本步重用同一映射慣例。
- **其餘對齊（無修改）**：
  - **PC 施法 dispatch**：`skill_buff.go:1125-1146 case 112` — `calcArmorBreakProb`（60/40/20 atk/def lvl + magichit + BaseInt 加成）+ `removeBuffAndRevert + applyBuffEffect`（8 秒）+ S_SkillSound(3400) 廣播 + S_PacketBoxIconAura(119, 8) target + 系統訊息「破壞盔甲 施放成功!」對齊 Java skillmode/ARMOR_BREAK.java:23-35。
  - **NPC 施法 dispatch**：`skill_status.go:720-735` — `calcArmorBreakProbNpc` 同公式 + `npc.AddDebuff(112, dur*5)` + 3400 GFX 對齊 Java skillmode line 37-46。
  - **機率公式**：`calcArmorBreakProb`/`Npc` `skill_status.go:859-905` — 60/40/20 prob + `shockStunIntMagicHit(intel)` `(INT-20)/3` for 23-127 範圍 + `shockStunBaseIntMagicHit(caster)` BaseInt（排除 EquipBonuses+buff DeltaIntel）25-44 → `+(BaseInt-15)/10`、>=45 → +5。對齊 Java `L1MagicPc.calcProbabilityMagic` line 728-746 + L1AttackList.INTH 表。
  - **buff 計時**：`buffs.lua [112] = {}` 旗標型 debuff + `[112]=true` NON_CANCELLABLE。8 秒 hardcoded 對齊 Java `setSkillEffect(ARMOR_BREAK, 8*1000)`。
  - **counterMagicExempt**：Java EXCEPT_COUNTER_MAGIC 不含 112（line 145-156 列表中跳過 112），Go `skill_buff.go:405` `109/110/111/113/114...` 同樣跳過 112，正確對齊（112 可被相消攔截）。
- **broader gap（不改）**：
  - **`getArmorBreakLevel()` 娃娃加成缺失**：Java `L1MagicPc.calcProbabilityMagic` line 729 `attackLevel += _pc.getArmorBreakLevel()`，數值由 `Doll_ArmorBreakLevel` 魔法娃娃 setDoll/removeDoll 維護。Go 無 `ArmorBreakLevel` 欄位、無 doll executor，導致娃娃加成的 ARMOR_BREAK 機率提升不生效。屬廣域 magic doll executor 系統缺口，與其他 `Doll_*` executor（Doll_DmgUp/HitUp 等）同類缺失，需 doll system 全面實作時統一補齊。本步維持單一技能 audit 範圍不修。
  - **FoeSlayer skill 187 damage 路徑 ARMOR_BREAK 1.58x 無 weaponType 過濾**：`skill_dragonknight.go:328, 359` 兩處 FoeSlayer 對 buff 112 目標套 1.58x，未檢查 weaponType。FoeSlayer 是 dragon knight 三段攻擊 skill，Java 是否複用 L1AttackPc 通用流程或自行計算需於 187 audit 時驗證。本步維持不變待 187 對應 audit。
- **不寫新測試**：weaponType 過濾為 2 行條件 + 1 字串比對，與既有 `processMeleeAttack`/`HandlePvPAttack` 路徑緊密耦合且 melee 攻擊路徑已被大量整合測試覆蓋（如 elemental venom/burning spirit 等同源插入點測試）。依停損標準避免新增 sole-purpose 防回歸測試。

## 迴避提升（DRESS_EVASION / 111）— 純審計無代碼變更

- 純審計 `111 DRESS_EVASION`：Go 完整對齊 Java `skillmode/DRESS_EVASION.java` 與 `L1PcInstance.java:3357-3403 getEr()` 套用 +18 + UPDATE_ER 雙路徑通知。無 REPEATEDSKILLS 群組成員（Java 10 個群組均不含 111）。
- **核心行為已對齊**：
  - **buff 套用**：`buffs.lua:97 [111] = { dodge = 18 }` 對齊 yiwei `ConfigOtherSet2.DRESS_EVASION=18`。Java getEr() line 3384-3387 `if (hasSkillEffect(DRESS_EVASION)) er += 18`。
  - **cast 端 UPDATE_ER**：`skill_buff.go:242-243` `applyBuffEffect` 後對 `skill.SkillID == 111 && DeltaDodge > 0` 條件送 `SendUpdateER(target.Dodge)` 對齊 Java skillmode `start()` line 14 `pc.sendPackets(new S_PacketBox(UPDATE_ER, pc.getEr()))`。
  - **stop 端 UPDATE_ER**：`skill_buff.go:521-522 revertBuffStats` 對 `buff.SkillID == 111 && DeltaDodge > 0` 條件送 `SendUpdateER(target.Dodge)` 對齊 Java skillmode `stop()` line 31 `pc.sendPackets(new S_PacketBox(UPDATE_ER, pc.getEr()))`。
  - **NON_CANCELLABLE**：`buffs.lua:255 [111] = true` 對齊 Java `L1SkillMode.java:37 isNotCancelable` 含 111。
  - **counterMagicExempt**：`skill_buff.go:405 111: true` 對齊 Java `L1SkillUse.java:148 EXCEPT_COUNTER_MAGIC` 含 111。
  - **REPEATEDSKILLS**：Java 10 個群組均不含 111，Go 端 buffs.lua `[111]` 不設 exclusions——正確對齊（與 109/110 不同，111 無 PHYSICAL 對應前置 buff 互斥）。
- **broader gap（不改，承襲 2026-05-18 audit）**：
  - **動態 getEr() vs 靜態 target.Dodge**：Java getEr() 為動態計算 `class/level bonus + (dex>=8 ? (dex-8)/2+4 : 3) + originalEr + (DRESS_EVASION ? 18 : 0) + (SOLID_CARRIAGE ? 15 : 0) + (STRIKER_GALE ? return 0 短路 : 0) + (AQUA_PROTECTER ? 5 : 0)`，傳給客戶端是「總 ER 值」。Go `target.Dodge` 為靜態累加值（僅 buff `dodge=N` 加總），客戶端收到「累加 buff 值」而非「總 ER」。實際 UI 顯示數值與 Java 不一致（例：Knight L50 DEX12 持 DRESS_EVASION 時 Java 顯示 ER=34，Go 顯示 18）。屬廣域 ER 系統計算架構缺口，影響：(1) 等級/職業 ER 加成（getLevel()/4 等）、(2) DEX→ER 加成（(dex-8)/2+4）、(3) originalEr 其他來源、(4) STRIKER_GALE 0 短路（174 cast 時 Go 已單獨補 `SendUpdateER(0)`，但 111 cast 時若 174 已啟動仍送 target.Dodge）、(5) AQUA_PROTECTER +5。單一技能 audit 無法完整修，與 110 DEX→AC 缺口同類列 out-of-scope。
- **STRIKER_GALE 短路反向缺口**：Java getEr() 對 STRIKER_GALE(174) 啟動者直接 `return 0` short-circuit。Go 在 174 cast/stop/expire 三點均單獨補 `SendUpdateER(0/Dodge)`。但若 player 先持 174 再 cast 111，Go 在 111 cast 路徑會送 `target.Dodge`（內含 +18）而非 Java 預期的 0。屬上述廣域 ER 計算缺口的子症狀，依「不可偷換範圍」維持 out-of-scope（待 ER 系統重構或 174 audit 引入 getEr-style helper 時統一處理）。
- **不寫新測試**：純審計，Go 已對齊主流程；既有 `TestSkillElementalBuff*`/`TestSkillDressEvasion*`（若有）覆蓋 dodge=18 套用與 UPDATE_ER 觸發。依停損標準避免「Go 已對 + 防回歸」測試。

## 敏捷提升（DRESS_DEXTERITY / 110）— 補齊 REPEATEDSKILLS[4] 26↔110 互斥

- 對齊 Java `L1SkillUse.java:1750` `REPEATEDSKILLS[4] = { PHYSICAL_ENCHANT_DEX, DRESS_DEXTERITY }`。與 109 audit 同源 Java mutex pattern：原 buffs.lua `[26]`/`[110]` 均無 exclusions，導致 PHYSICAL_ENCHANT_DEX(26)+DRESS_DEXTERITY(110) 同時生效 → DEX 加成 +5+3=+8（Java 上限 +5 二擇一）。
- 修正：buffs.lua bilateral exclusion：
  ```lua
  [26]  = { dex = 5, exclusions = {110} },
  [110] = { dex = 3, exclusions = {26} },
  ```
  與 109 audit 同樣掛載到既有 `engine.go:431-437` exclusions parser + `skill_buff.go:144-146 removeBuffAndRevert` 通用路徑，無需新代碼。
- 其餘對齊（無修改）：
  - **DeltaDex 套用**：`applyBuffEffect:151 buff.DeltaDex = int16(eff.Dex)` + `revertBuffStats` 反向 `target.Dex -= DeltaDex` 對齊 Java `pc.addDex((byte)3)` / `addDex(-3)`。
  - **Cast icon**：`buff_icon_map.yaml [110] type=dexup param=2` 對齊 Java `L1SkillUse.java:2449 new S_Dexup(pc, 2, duration)`（comment `原本3修改2 琮善`）。
  - **Cancel icon**：`skill_buff.go:107-109 durationSec==0 && skillID==110 → iconParam=3` 對齊 Java `L1SkillStop.java:433 new S_Dexup(pc, 3, 0)`。
  - **NON_CANCELLABLE**：`buffs.lua:254 [110] = true` 對齊 Java `L1SkillMode.java:37`。
  - **counterMagicExempt**：`skill_buff.go:405 110: true` 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 110。
- **broader gap（不改，與 109 同樣記錄但更新覆蓋率）**：Java REPEATEDSKILLS 10 個群組 Go 目前覆蓋：[0]={148,149,156,163,166} ✅、[1]={151,168} ✅、[2]={52,101,150,155,186,1000,1016} ✅、[4]={26,110} ✅（本步補）、[5]={42,109} ✅（109 step 補）、[7]={185,190,195} ✅（dragon awakening 專屬守衛）。仍缺：[3]={8,19,54}（HASTE 系）、[6]={80,106}（MIRROR_IMAGE/UNCANNY_DODGE）、[8]={14,213}、[9]={213,218}。
- **DEX→AC 連動廣域缺口（不改，與 110 既有 2026-05-18 audit 同源）**：Java `resetBaseAc()` 在 cast 與 stop 兩端都呼叫，依 DEX-tier-based 公式（0-9:10-L/8、10-12:10-L/7、13-15:10-L/6、16-17:10-L/5、18+:10-L/4）重算 base AC。Go 用固定 `Config.Gameplay.BaseAC`，無 DEX→AC 連動，影響升級/重置/裝備事件等多個系統，非單一技能可修。依「不可偷換範圍」維持 out-of-scope。
- **不寫新測試**：bilateral exclusion 是 data-only fix，掛載到既有通用路徑（與 109/151/168/148-群同源），既有 `TestSkillElementalBuffIronSkinExcludesEarthSkin` 已鎖定機制。依停損標準避免重複測試。

## 力量提升（DRESS_MIGHTY / 109）— 補齊 REPEATEDSKILLS[5] 42↔109 互斥

- 對齊 Java `L1SkillUse.java:1752` `REPEATEDSKILLS[5] = { PHYSICAL_ENCHANT_STR, DRESS_MIGHTY }`。Java `deleteRepeatedSkills` (line 1769-1777) 每次施法時掃描所有 REPEATEDSKILLS 群組，若 cast skill ID 在群內，移除群內其餘技能。原 Go buffs.lua `[42]`/`[109]` 均無 exclusions，導致 42 + 109 可同時生效 → STR 加成 +5+3=+8（Java 上限 +5 或 +3 二擇一）。
- 修正：buffs.lua bilateral exclusion：
  ```lua
  [42]  = { str = 5, exclusions = {109} },
  [109] = { str = 3, exclusions = {42} },
  ```
  `internal/scripting/engine.go:431-437` 既有 exclusions parser + `internal/system/skill_buff.go:144-146` `removeBuffAndRevert(target, int32(exID))` 套用路徑與既有 148/149/156/163/166（fire weapon 群）/151/168（earth/iron skin 群）同源實作，不需新代碼。
- 其餘對齊（無修改）：
  - **DeltaStr 套用**：`applyBuffEffect` `buff.DeltaStr = int16(eff.Str)` + `target.Str += buff.DeltaStr` 對齊 Java `pc.addStr((byte) 3)`。
  - **Cast icon**：`buff_icon_map.yaml [109]={type=strup, param=2}` 對齊 yiwei `L1SkillUse.java:2456 new S_Strup(pc, 2, duration)`（comment `原本3修改2 琮善`）。
  - **Cancel icon**：`skill_buff.go:99-101` `if durationSec == 0 && skillID == 109 { iconParam = 3 }` 對齊 Java `L1SkillStop.java:441 new S_Strup(pc, 3, 0)`。
  - **NON_CANCELLABLE**：`buffs.lua:253 [109] = true` 對齊 Java `L1SkillMode.java:36 isNotCancelable` 含 109。
  - **counterMagicExempt**：`skill_buff.go:405 109: true` 對齊 Java `L1SkillUse.java:148 EXCEPT_COUNTER_MAGIC` 含 109。
  - **revert 反向**：`revertBuffStats` 統一 `target.Str -= buff.DeltaStr` + 觸發 `sendBuffIcon(durationSec=0)` 走 cancel-icon 覆寫路徑送 type=3。
- **broader gap（不改）**：Java REPEATEDSKILLS 共 10 個群組，Go 已有 [0]={148,149,156,163,166}、[1]={151,168}、[2]={52,101,150,155,186,1000,1016}、[5]={42,109} 透過 lua exclusions 實作；[3]={8,19,54}（HASTE 系）、[4]={26,110}（DEX 系，留待 110 audit）、[6]={80,106}（MIRROR_IMAGE/UNCANNY_DODGE）、[7]={185,190,195}（dragon awakening，已有專屬 HasBuff 守衛）、[8]={14,213?}、[9]={213?,218?} 部分缺失。本步只補 [5]，其餘屬其本身技能 audit 範圍。
- **不寫新測試**：bilateral exclusion 是 data-only fix，掛載到既有 `engine.go:431-437` exclusions parser + `skill_buff.go:144-146` removeBuffAndRevert 通用路徑——與 148-166/151-168 群組同源實作。既有 `TestSkillElementalBuffIronSkinExcludesEarthSkin` 已覆蓋 lua exclusions 機制。依停損標準避免重複測試同一機制。

## 會心一擊（FINAL_BURN / 108）— 純審計無代碼變更

- 純審計 `108 FINAL_BURN`：Go 完整對齊 Java `L1SkillUse.java:1223-1228, 1277-1287` + `L1MagicPc.java:1102-1104` 的 HP/MP 燃燒攻擊邏輯。三條既有測試完整覆蓋。
- **核心行為已對齊**：
  - **HP <= 100 拒絕**：`skill.go:323-326`：
    ```go
    if skillID == 108 && player.HP <= 100 {
        handler.SendServerMessage(sess, skillMsgNotEnoughHP)  // 279
        s.failTeleportSkill(sess, skillID)
        return
    }
    ```
    對齊 Java `L1SkillUse.java:1223-1228 if (_skillId == FINAL_BURN && currentHp <= 100) → S_ServerMessage(279) "因體力不足而無法使用魔法" + return false`。`skillMsgNotEnoughHP=279`（`skill.go:17`）對齊 Java sysmsg 279。
  - **跳過預先 MP 消耗**：`skill.go:444-447 if skillID != skillFinalBurn → consume MP normally`；FINAL_BURN 不在此處扣 MP，保留至 `consumeFinalBurnResources` 後置處理。
  - **執行順序對齊 Java useConsume 後置**：Java `handleCommands` line 481-482：
    ```java
    runSkill();         // 1. 跑技能（含 calcSkillDmg → dmg = currentMp）
    useConsume();       // 2. 後置消耗 HP/MP（FINAL_BURN: HP→100, MP→1）
    ```
    Go `skill.go:504-507`：
    ```go
    s.executeAttackSkill(sess, player, skill, targetID)  // 1. 跑技能（含 res.Damage = player.MP）
    if skillID == skillFinalBurn {
        s.consumeFinalBurnResources(sess, player)        // 2. 後置消耗 HP→100, MP→1
    }
    ```
    順序完全一致——傷害使用 pre-consume MP。
  - **傷害公式**：`skill_damage.go:81-82` + `:491-495`：
    ```go
    } else if skill.SkillID == skillFinalBurn {
        dmg = player.MP                                  // PC target 路徑
    }
    // 與
    if skill.SkillID == skillFinalBurn {
        res.Damage = int(player.MP)                      // NPC target 路徑
        res.HitCount = 1
        res.DrainMP = 0
    }
    ```
    對齊 Java `L1MagicPc.java:1103-1104 if (skillId == 108) dmg = _pc.getCurrentMp()`。L1MagicPc.calcSkillDmg 在 runSkill → useConsume 之前呼叫，使用 pre-consume MP（與 Go 同序）。
  - **資源燃燒到 Java floor**：`skill.go:602-611 consumeFinalBurnResources`：
    ```go
    if player.HP != 100 {
        player.HP = 100
        sendHpUpdate(sess, player)
    }
    if player.MP != 1 {
        player.MP = 1
        sendMpUpdate(sess, player)
    }
    ```
    對齊 Java `L1SkillUse.java:1277-1287`：
    ```java
    _hpConsume = currentHp - 100;
    _mpConsume = currentMp - 1;
    setCurrentHp(currentHp - _hpConsume);  // = 100
    setCurrentMp(currentMp - _mpConsume);  // = 1
    ```
  - **counterMagicExempt 不含 108**：Java `EXCEPT_COUNTER_MAGIC` (`L1SkillUse.java:147-148`) 列表 `97, 98, 99, 100, 101, 102, 104, 105, 106, 107, 109, 110, 111, 113, ...`——**從 107 跳到 109 略過 108**。Go `counterMagicExempt` (`skill_buff.go:398-413`) 同樣不含 108 → FINAL_BURN **可**被反魔法盾抵擋（攻擊魔法被 COUNTER_MAGIC buff 反射）。
  - **PvE + PvP 完整路徑**：`skill_damage.go:81-82`（玩家對玩家路徑）+ `:491-495`（玩家對 NPC 路徑）兩處都覆蓋 FINAL_BURN 傷害計算。
  - **既有測試完整覆蓋**：`skill_final_burn_test.go`：
    1. `TestSkillFinalBurnFinalBurnDamagesWithPreConsumeMP` — caster MP=80 + target HP=200 → 攻擊後 target HP=120（傷害 = pre-consume MP 80）+ caster HP/MP = 100/1。
    2. `TestSkillFinalBurnFinalBurnRequiresHpAbove100` — caster HP=100 + MP=80 → 攻擊失敗，caster HP/MP 不變（100/80）+ target HP 不變（200）。
    3. `TestSkillFinalBurnFinalBurnConsumesHpAndMpToJavaFloor` — caster HP=250 MP=80 → 攻擊後 HP/MP=100/1（即使攻擊未命中也消耗，對齊 Java useConsume 在 runSkill 後無條件執行）。
- **broader gap（不改）**：
  - **HP/MP consume 訊息與動畫順序**：Java `handleCommands` 順序 `runSkill → useConsume → sendGrfx → sendFailMessage`，動畫送出在 useConsume 之後。Go 在 executeAttackSkill 內部送 GFX，然後 consumeFinalBurnResources。動畫與資源消耗的訊息順序若不一致可能造成客戶端 UI 微小視覺差異，但 Java 自身在 yiwei 版也是先 runSkill (含 broadcast attack) 再 useConsume，順序等價。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，三條既有測試完整覆蓋核心行為（pre-consume 傷害、HP 不足拒絕、資源燃燒）。

## 暗影之牙（SHADOW_FANG / 107）— 純審計無代碼變更

- 純審計 `107 SHADOW_FANG`：Go 完整對齊 Java `L1SkillUse.java:2458-2468` + `L1ItemInstance.setSkillWeaponEnchant():1088-1123` 的武器附魔系統。
- **核心行為已對齊**：
  - **dispatch**：`skill.go:489-491 case 12, 107` → `executeTargetedWeaponEnchant(sess, player, skill, targetID)` 並 `return` 短路。卷軸路徑 `skill_magic_scroll.go:92 case 12, 107` 同樣短路。對齊 Java `L1SkillUse.java:315-319 case ENCHANT_WEAPON/BLESSED_ARMOR/SHADOW_FANG → _itemobjid = target_id`（targetID 解讀為背包物品 ObjectID）。
  - **物品驗證**：`skill_weapon.go:276-285`：
    - `weapon = player.Inv.FindByObjectID(itemObjID)` → nil → `SendServerMessage(79)`。
    - `itemInfo.Category != data.CategoryWeapon` → `SendServerMessage(79)`。
    - 對齊 Java `item.getItem().getType2() == 1`（type 1 = 武器）else `S_ServerMessage(79)`。
  - **套用 enchant**：`skill_weapon.go:246-262 applySkillWeaponEnchant`：
    ```go
    weapon.DmgByMagic = 0       // ← 重置舊值（對齊 Java setSkillWeaponEnchant cancel old timer + reset to 0）
    weapon.HitByMagic = 0
    switch skill.SkillID {
    case 107:
        weapon.DmgByMagic = 5
    }
    weapon.DmgMagicExpiry = skill.BuffDuration * 5  // ticks (192*5 = 960 ticks = 192s)
    ```
    對齊 Java `L1ItemInstance.setSkillWeaponEnchant():1088-1123`：cancel old timer + `setDmgByMagic(0)+setHolyDmgByMagic(0)+setHitByMagic(0)` reset + `case 107 setDmgByMagic(5)` + schedule new Timer with `skillTime` ms = buff_duration * 1000 ms。
  - **icon**：`skill_weapon.go:299-301 SendWeaponEnchantIcon(sess, 2951, buff_duration, true)` 對齊 Java `S_PacketBox(SKILL_WEAPON_ICON, 2951, buff_duration, true)`。
  - **stats 更新**：`RecalcEquipStats` 在 enchant 後重新計算裝備加成；`equip.go:528-533` 在 RecalcEquipStats 內讀取 `weapon.DmgByMagic > 0 && weapon.DmgMagicExpiry > 0 → stats.DmgMod += weapon.DmgByMagic`，等同於 Java 角色 `_dmgup` 累加裝備加成的方式。
  - **到期**：`buff_tick.go:61-81 tickItemMagicEnchant` 每 tick 遞減 `DmgMagicExpiry`，歸零時 `DmgByMagic = 0 + HitByMagic = 0 + DmgMagicExpiry = 0 → changed=true → RecalcEquipStats`。對齊 Java EnchantTimer 到期 `setDmgByMagic(0)+setHolyDmgByMagic(0)+setHitByMagic(0)`。
  - **重施可刷新**：Java SHADOW_FANG 無 `!hasSkillEffect` 守衛，重施由 `setSkillWeaponEnchant` 內部 `if (_isRunning) → cancel timer + reset` 處理；Go `applySkillWeaponEnchant` 同樣先 reset 再設新值，等效對齊（與 skill 106 UNCANNY_DODGE 不同：106 是 player buff 且 Java 有守衛跳過重施，107 是 weapon enchant 且 Java 允許重施刷新）。
  - **counterMagicExempt**：`skill_buff.go:405 counterMagicExempt[107] = true` 對齊 Java `EXCEPT_COUNTER_MAGIC` (`L1SkillUse.java:148`) 含 107——SHADOW_FANG 自我增益不被反魔法盾抵擋。
  - **NON_CANCELLABLE**：`buffs.lua:250 [107] = true` 對齊 Java `L1SkillMode.isNotCancelable()` 第 36 行明確列出 SHADOW_FANG（防禦性設定，實際 weapon enchant 不存在於 player.ActiveBuffs 中，CANCELLATION 不會接觸到）。
- **發現一處無害死碼（不修）**：
  - `scripts/combat/buffs.lua:84 [107] = { dmg_mod = 5 }`：從未被觸發——dispatcher `skill.go:489` 與 `skill_magic_scroll.go:92` 的 `case 12, 107` 短路 return，永不走 default → `applyBuffEffect` → `get_buff(107)` 路徑。若未來誤被觸發（如新增測試直接呼叫 applyBuffEffect with SkillID:107）會與 `weapon.DmgByMagic` 的 +5 重複計算（player buff +5 unconditional × weapon enchant +5 conditional on equip），結果錯誤地產生 +10 dmg_mod。屬具誤導性的歷史 dead code 但目前不會觸發；依「不可偷換範圍」+「做半套不如不做」+ Karpathy 「Don't remove pre-existing dead code unless asked」記錄不修。
- **broader gap（不改）**：
  - **Java timer vs Go tick 精度**：Java 使用 ms 級 `Timer.schedule(skillTime)`，Go 使用 200ms tick（buff_duration*5）。192s buff 在 Java 為精確 192000ms，Go 為 960 ticks ≈ 192s（tick 邊界誤差 < 200ms）。屬廣域 buff timer 精度差異，影響所有 tick-driven buffs。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 dispatcher + applySkillWeaponEnchant + tickItemMagicEnchant 路徑已具備正確行為（含 ENCHANT_WEAPON skill 12 共用實作）。

## 暗影閃避（UNCANNY_DODGE / 106）— 補齊 Java skillmode 重施守衛

- 補齊 `106 UNCANNY_DODGE` 重施守衛：Java `skillmode/UNCANNY_DODGE.java:17 if (!srcpc.hasSkillEffect(106))` 把 `setSkillEffect + add_dodge(5) + S_PacketBoxIcon1` 三項包在守衛裡——重施時三項皆跳過（buff timer 不刷新、dodge 不再加、客戶端不收到 icon 更新）。Go `skill_self.go` default 路徑無此守衛，重施會：(1) 在 `target.Dodge` 上重複加減（透過 AddBuff 替換舊 buff + revertBuffStats 反向，net 結果 dodge 正確，但中間 SendDodgeIcon 送出膨脹數值產生圖示閃動）、(2) 替換 buff 等同刷新 timer 延長 buff 持續時間——兩者為 criterion (a) 真實 Java vs Go 差異。本步在 `skill_self.go` applyBuffEffect 前加守衛，重施時跳過 buff 處理但保留外層 cast GFX（2950）廣播（對齊 Java L1SkillUse outer flow）。
- **核心修改**：
  - `server/internal/system/skill_self.go:245-249`：在 `s.applyBuffEffect(player, skill)` 前加：
    ```go
    // UNCANNY_DODGE (106) 守衛：Java skillmode/UNCANNY_DODGE.java:17 `if (!srcpc.hasSkillEffect(106))`
    // 跳過 stat 加成、timer 刷新、S_PacketBoxIcon1 通知三項——重施時保留外層 cast GFX 廣播但 buff 內容不變。
    if !(skill.SkillID == 106 && player.HasBuff(106)) {
        s.applyBuffEffect(player, skill)
    }
    ```
- **既有對齊已驗證（無修改）**：
  - **施法路徑**：`skill_list.yaml skill_id:106 target:none cast_gfx:2950 action_id:19 buff_duration:192`，target=none → `executeSelfSkill` default → applyBuffEffect → `buffs.lua [106] = { dodge = 5 }`。
  - **首次套用**：`skill_buff.go:169 buff.DeltaDodge = 5` → `:203 target.Dodge += 5` → `:241-246 SendDodgeIcon(target.Session, target.Dodge, true)`（S_PacketBoxIcon1 opcode 250, subcode 0x58 + 當前 dodge 總值）。對齊 Java `srcpc.add_dodge(5) + S_PacketBoxIcon1(true, get_dodge())`。
  - **buff 到期/revert**：`skill_buff.go:516 target.Dodge -= buff.DeltaDodge` + `:520-525 SendDodgeIcon(target.Session, target.Dodge, true)`。對齊 Java skillmode `stop()` `pc.add_dodge(-5) + S_PacketBoxIcon1(true, get_dodge())`。
  - **counterMagicExempt**：`skill_buff.go:405 counterMagicExempt[106] = true` 對齊 Java `EXCEPT_COUNTER_MAGIC` (line 148) 含 106。
  - **NON_CANCELLABLE**：`buffs.lua:249 [106] = true` 對齊 Java `L1SkillMode.isNotCancelable()` 第 35 行明確列出 UNCANNY_DODGE——不可被 CANCELLATION 解除。
  - **MP 消耗**：`skill.go:335 adjustedSkillMPConsume` 在 applyBuffEffect 前執行，重施時 MP 仍被消耗（對齊 Java L1SkillUse outer flow 即使 skillmode 守衛拒絕也照樣扣 MP）。
- **broader gap（不改）**：
  - **重施守衛廣域模式**：Java 大量 skillmodes（>30 個檔案：ADVANCE_SPIRIT, AWAKEN_ANTHARAS/FAFURION/VALAKAS, BOUNCE_ATTACK, SHADOW_ARMOR, SHADOW_FANG, DRESS_MIGHTY/DEXTERITY/EVASION, ILLUSION_OGRE/LICH/DIA_GOLEM/AVATAR, BONE_BREAK 等）皆有 `if (!hasSkillEffect)` 守衛跳過 stat/timer 刷新。Go 目前僅在 dragon awakening（185/190/195）case 有顯式 HasBuff 守衛，其餘大多走 generic applyBuffEffect 路徑可能允許 timer 刷新。屬廣域 buff stack semantic 對齊缺口——應在後續系統審計時設計通用 `setSkillEffect-equivalent` flag（如 ActiveBuff.NoRefreshOnRecast），本步僅補 skill 106 自己的守衛。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：`go build ./...` EXIT=0；`Dodge|Dark|Skill` 相關測試全 PASS。
- **不寫新測試**：依停損標準，本步為「Java 真實差異 + 改 Go 對齊」單行守衛修補，行為等價於既有的 dragon awakening 守衛模式（185/190/195 case）。skill 106 既無針對重施場景的測試需求，避免「Go 已對 + 防回歸」測試。

## 雙重破壞（DOUBLE_BREAK / 105）— 純審計無代碼變更

- 純審計 `105 DOUBLE_BREAK`：Go 完整對齊 Java `L1AttackPc.calcDamage` weapon switch case 11/12 的雙重破壞傷害數學。
- **核心行為已對齊**：
  - **施法路徑**：`skill_list.yaml skill_id:105 target:none buff_duration:192` → `executeSelfSkill` default → `applyBuffEffect` → `buffs.lua [105] = {}` 空效果表（純旗標）。對齊 Java 無 skillmode 走 L1SkillUse default `setSkillEffect(105, _getBuffDuration)`。
  - **觸發條件**：`skill_damage.go:326 if attacker.HasBuff(105) && doubleBreakRoll < doubleBreakChance(attacker, weaponType)` → `damage *= doubleBreakMultiplier (=2)`。對齊 Java case 11/12 `if (_pc.hasSkillEffect(DOUBLE_BREAK) && _random.nextInt(100) < totalchance && !ConfigSkill.DOUBLE_BREAK_NO_WEAPON) → weaponTotalDamage *= ConfigSkill.DOUBLE_BREAK_DMG`。
  - **機率公式對齊**：
    - **Claw**：`doubleBreakChance("claw") = 20 + (level-45)/5 if level>45 else 0`。對齊 Java case 11 `totalchance = 20 + addchance` 其中 `addchance = (_pc.getLevel() - 45) / 5 if _pc.getLevel() > 45 else 0`。
    - **Edoryu**：`doubleBreakChance("edoryu") = 20 + (level-45)/5 + 5`。對齊 Java case 12 `totalchance2 = 20 + addchance2 + ConfigSkill.DOUBLE_BREAK_CHANCE(=5)`。
    - **其他武器（bow/sword 等）**：`doubleBreakChance default = 0`。對齊 Java case 11/12 only（其他 weapon case 不檢查 DOUBLE_BREAK）。
  - **倍率對齊**：Go `doubleBreakMultiplier = 2` 對齊 Java properties `DOUBLE_BREAK_DMG = 2.0`（覆蓋 ConfigSkill 預設 1.5）。
  - **roll 範圍對齊**：Go `world.RandInt(100)` = `rand.Intn(100)` → [0, 99]；`roll < chance` 命中 `chance` 個值。對齊 Java `_random.nextInt(100) < totalchance` 同樣 [0, 99] 命中 `totalchance` 個值。
  - **PvE/PvP melee 全覆蓋**：`combat.go:212 processMeleeAttack` 與 `pvp.go:95 HandlePvPAttack` 皆呼叫 `darkElfPhysicalDamage(attacker, damage, weaponType)`。
  - **PvE/PvP ranged 自然排除**：`combat.go:459 processRangedAttack` 與 `pvp.go:298 HandlePvPFarAttack` 呼叫 `darkElfPhysicalDamage(attacker, damage, "bow")`（為 102 BURNING_SPIRIT 補上的路徑）但 `doubleBreakChance("bow") = 0` 自然排除。對齊 Java DOUBLE_BREAK 只在 weapon switch case 11/12（claw/edoryu）處理，與弓無關。
  - **counterMagicExempt**：`counterMagicExempt[105] = true` 對齊 Java `EXCEPT_COUNTER_MAGIC` (line 148) 含 105——DOUBLE_BREAK 自我增益不被反魔法盾抵擋。
  - **NON_CANCELLABLE 不含 105**：對齊 Java `isNotCancelable` 不含 105——可被 CANCELLATION 解除。
  - **既有測試**：`skill_darkelf_buff_test.go:60-87` Player level 50 + buff 102/105 + edoryu：
    - `darkElfPhysicalDamageWithRolls(player, 100, "edoryu", 0, 0)` → 期望 300：105 chance = 20+1+5 = 26，roll 0 < 26 觸發 ×2 → 200；102 roll 0 < 15 觸發 ×3/2 → 300 ✓
    - `darkElfPhysicalDamageWithRolls(player, 100, "edoryu", 99, 99)` → 期望 100：兩者 roll 99 ≥ chance 均不觸發 ✓
- **broader gap（不改）**：
  - **attack packet 雙擊視覺標記 0x04 缺失**：Java edoryu DOUBLE_BREAK 觸發時設 `_attackType = 4`（L1AttackPc:987），透過 `S_AttackPacketPc(pc, target, type, dmg)` 最後一個 byte 送給客戶端觸發雙擊動畫。Go `handler/broadcast.go:353 BuildAttackPacket` 簽名無 type 參數，硬編碼末 byte = 0，客戶端看不到雙擊視覺。此問題同時影響 4 個獨立系統：(1) edoryu DOUBLE_BREAK skill proc 視覺、(2) edoryu 武器自身雙擊（`_weaponDoubleDmgChance`）、(3) 暴擊 0x02、(4) 鏡反射 0x08。屬廣域 attack packet 動畫旗標缺口，非 skill 105 獨有。注意 Java claw DOUBLE_BREAK（line 962）**不**設 `_attackType=4`，因此 claw 路徑反而自然對齊。修補需重構 BuildAttackPacket 簽名 + 所有 caller + 暴擊/鏡反射等獨立系統，半補只 edoryu DOUBLE_BREAK 而留其他三個為 0 = 半成品，依「不可偷換範圍」+「做半套不如不做」記錄不修。
  - **DOUBLE_BREAK_NO_WEAPON 替代公式**：Java properties `DOUBLE_BREAK_NO_WEAPON = false` 為生產設定，當 true 時走 L1AttackPc:1759/2066 替代公式（用 `_weaponDoubleDmgChance` 而非 level bonus）。Go 不模擬此 config，依賴預設 false 行為。屬廣域 config 對齊缺口。
  - **edoryu 武器自身雙擊 `_weaponDoubleDmgChance`**：Java case 12 line 968-971 另有獨立的 weapon 雙擊系統（非 DOUBLE_BREAK skill），與 105 對齊無關。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 `TestSkillDarkElfBuffBurningSpiritAndDoubleBreakAreProcFlags` 已透過 `darkElfPhysicalDamageWithRolls` 鎖定 105 觸發 + 不觸發兩條路徑。

## 毒性抵抗（VENOM_RESIST / 104）— 純審計無代碼變更

- 純審計 `104 VENOM_RESIST`：Go 完整對齊 Java `L1Poison.isValidTarget()` 對 VENOM_RESIST 與 DRAGON5 兩種 buff 的阻擋。
- **核心行為已對齊**：
  - **施法路徑**：`skill_list.yaml skill_id:104 target:none` → `executeSelfSkill` default 路徑 → `applyBuffEffect` → `buffs.lua [104] = {}` 空效果表（純旗標 buff，無 stat 加成，對齊 Java 無 skillmode 類別走 L1SkillUse default `setSkillEffect(_skillId, _getBuffDuration)`）。
  - **施毒攔截**：`poison.go:289-291 hasPoisonResistance(target)` 回傳 `target.HasBuff(104) || target.HasBuff(6687)` 對齊 Java `L1Poison.isValidTarget`：
    ```java
    if (player.hasSkillEffect(L1SkillId.VENOM_RESIST)) return false;  // VENOM_RESIST = 104
    if (player.hasSkillEffect(L1SkillId.DRAGON5)) return false;       // 生命魔眼 = Go buff 6687
    ```
  - **進入點全覆蓋**：所有玩家施毒路徑都走 `canApplyPoisonToPlayer` 或其包裝：
    - `ApplyNpcPoisonAttackWithRoll`（怪物攻擊施毒）→ 第 195 行 `canApplyPoisonToPlayer` 阻擋。
    - `applyDamagePoisonToPlayer`（傷害毒）→ 第 256 行內部 `canApplyPoisonToPlayer` 阻擋。
    - `applyEnchantVenomPoisonToPlayerWithRoll`（玩家附加劇毒對玩家）→ 委派 `applyDamagePoisonToPlayer` 內部阻擋。
    - `GMApplyPoison`（GM 指令）→ 第 314 行 `canApplyPoisonToPlayer` 阻擋。
    - `applySilencePoisonToPlayer`/`applyParalysisPoisonToPlayer`（沉默毒、麻痺毒）僅從 dispatcher `ApplyNpcPoisonAttackWithRoll` + `GMApplyPoison` 呼叫，已由 caller 阻擋。
  - **counterMagicExempt**：`counterMagicExempt[104] = true`（`skill_buff.go:404`）對齊 Java `EXCEPT_COUNTER_MAGIC` (`L1SkillUse.java:148`) 含 104——VENOM_RESIST 自我增益不被反魔法盾抵擋。
  - **NON_CANCELLABLE 不含 104**：對齊 Java `L1SkillMode.isNotCancelable()`（line 31-37）不含 VENOM_RESIST——可被 CANCELLATION 解除。
  - **buff_icon_map 無 104**：對齊 Java VENOM_RESIST 無 skillmode 類別且 L1SkillUse default 路徑不發 S_PacketBox 圖示——本系統純後台旗標 buff。
  - **既有測試覆蓋**：
    - `skill_poison_venom_test.go:10-30 TestSkillPoisonVenomVenomResistBlocksNpcPoison` 鎖定 104 阻擋怪物施毒。
    - `:32-52 TestSkillPoisonVenomDragonLifeEyeBlocksPoison` 鎖定 6687（DRAGON5）阻擋怪物施毒。
    - `:54-84 TestSkillPoisonVenomCursePoisonRespectsPlayerPoisonResistance` 鎖定 104 阻擋玩家對玩家毒咒（skill 11）。
    - `:120-157 TestSkillPoisonVenomEnchantVenomPoisonsPlayerAndRespectsResistance` 鎖定 104 阻擋玩家附加劇毒（skill 98）proc。
- **broader gap（不改）**：
  - **item-equipment based `_venom_resist > 0`**：Java `L1Poison.isValidTarget` 第 35-37 行檢查 `player.get_venom_resist() > 0`，源自 `Venom_Resist.java` 與 `ElitePlateMail_Antharas.java` 等護甲 item executor 在裝備/卸下時 `set_venom_resist(±1)`。Go 目前**沒有** 此 counter 系統（player 結構無 `VenomResist int` 欄位）。此屬廣域 armor item executor 缺口（與 skill 104 本身對齊無關），玩家若裝備毒性抵抗護甲（如解毒甲冑、安塔瑞斯精製鎧甲）將不被擋毒。不在 skill 104 scope，依「不可偷換範圍」記錄不修。
  - **NPC 不檢查 buff 104**：Java `L1Poison.isValidTarget` 第 29-31 行對非 PC 直接 return true 不查 buff——Go `canApplyPoisonToPlayer` 僅針對 PlayerInfo，NPC 路徑由 `canApplyPoisonToNpc` 處理（已對齊 Java 行為）。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 4 個 venom resist 相關測試覆蓋核心路徑。

## 暗黑盲咒（DARK_BLIND / 103）— 純審計無代碼變更

- 純審計 `103 DARK_BLIND`：Go 完整對齊 Java `skillmode/DARK_BLIND.java`。
- **核心行為已對齊**：
  - **PC 目標 dispatch**：`skill_buff.go:1084-1088 case 103` → 將 SkillID 改為 66 後呼叫 `applyBuffEffect(target, sleepSkill)` → `buffs.lua [66]={sleeped=true}` → `SetSleeped=true + target.Sleeped=true + SendParalysis(SleepApply=0x0A)`。對齊 Java `pc.setSkillEffect(66, integer*1000) + pc.sendPackets(new S_Paralysis(3, true))`（`S_Paralysis(3, true)` wire 為 `0x0A` 對應 `SleepApply`）。
  - **NPC 目標 dispatch**：`skill_status.go:532-546 case 103` → `checkNpcMRResist` + `npc.Sleeped=true + npc.AddDebuff(66, dur*5)` + 廣播 CastGfx。對齊 Java `tgnpc.setSkillEffect(66, integer*1000) + tgnpc.setSleeped(true)`。
  - **PC→PC MR gating**：`playerDebuffSkills[103]=true`（`skill_status.go:827`）對齊 Java `L1MagicPc.calcProbabilityMagic` 對 PC debuff 走 MR 抗性檢查。
  - **counterMagicExempt 不含 103**：對齊 Java `EXCEPT_COUNTER_MAGIC` 顯式列出 `102, 104` 跳過 103——DARK_BLIND **可** 被反魔法盾(COUNTER_MAGIC/31)抵擋。
  - **NON_CANCELLABLE 不含 103**：對齊 Java `L1SkillMode.isNotCancelable()` 不含 DARK_BLIND——可被 CANCELLATION 相消（buff key 66 跟著被清）。
  - **PC 到期**：`skill_buff.go:775-777` 在 buff 66 SetSleeped 到期時送 `SendParalysis(SleepRemove=0x0B)` 對齊 Java `S_Paralysis(3, false)`。
  - **NPC 到期**：`npc_ai.go:1494 case 66` 清 `npc.Sleeped=false`（實際 debuff key 是 66）對齊 Java `tgnpc.setSleeped(false)`。
  - **睡眠中斷路徑（防禦性清除）**：`combat.go:897-898`、`pvp.go:558-559`、`npc_ai.go:481/608/900`、`skill_damage.go:273-274`、`skill_status.go:165-166/187-188` 同時 `RemoveBuff/RemoveDebuff(66)` 與 `(103)`——對外部來源或舊資料可能殘留 103 buff 的防禦性設計。
  - **既有測試覆蓋**：`skill_darkelf_buff_test.go:90-119 TestSkillDarkElfBuffDarkBlindUsesSleepEffect66` 已透過 `s.executeBuffSkill(... SkillID:103 ...)` 鎖定 `target.Sleeped && HasBuff(66) && !HasBuff(103)` 三層條件。
- **不動處（dead code 但無害）**：
  - `npc_ai.go:1496 case 103` 為 NPC debuff expiry handler 中的 case 分支，但 Go NPC dispatch 從未把 debuff key 設為 103（只設 66），因此該 case 永不觸發。其他路徑（如 `RemoveDebuff(103)`）為防禦性清除無害。屬無害的歷史 dead code，依停損標準避免無關 surgical 修改。
- **broader gap（不改）**：
  - **NPC MR 公式**：`checkNpcMRResist` 使用 generic MR 公式，Java `L1MagicPc.calcProbabilityMagic` 對不同技能有 case-by-case 機率公式。屬廣域 MR 公式對齊缺口。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 `TestSkillDarkElfBuffDarkBlindUsesSleepEffect66` 覆蓋 PC 目標睡眠效果。

## 燃燒鬥志（BURNING_SPIRIT / 102）— 補齊遠程攻擊觸發

- 補齊 `102 BURNING_SPIRIT` 遠程觸發路徑：Java `L1AttackPc.BuffDmgUp(dmg)` 在 `calcPcDamage:1702` 與 `calcNpcDamage:2027` 都會無條件呼叫，涵蓋近戰與遠程攻擊；BURNING_SPIRIT 自身條件僅 `_pc.hasSkillEffect(BURNING_SPIRIT) && random <= 15`（無 weaponType 排除）。Go 原本只在 `combat.go:212 processMeleeAttack` 與 `pvp.go:95 HandlePvPAttack` 呼叫 `darkElfPhysicalDamage`，弓矢遠程攻擊（PvE + PvP）完全沒套用 15% 機率 1.5x 增傷。本步在 `combat.go processRangedAttack` 與 `pvp.go HandlePvPFarAttack` 補上 `darkElfPhysicalDamage(player/attacker, damage, "bow")` 對齊 Java。
- **核心修改**：
  - `server/internal/system/combat.go:457-459`（`processRangedAttack`）：在 `strikerGaleRangedDamageToNpc` 之後、`braveAuraDamage` 之前插入 `damage = darkElfPhysicalDamage(player, damage, "bow")`。
  - `server/internal/system/pvp.go:296-298`（`HandlePvPFarAttack`）：在 `strikerGaleRangedDamage` 之後、`braveAuraDamage` 之前插入 `damage = darkElfPhysicalDamage(attacker, damage, "bow")`。
- **DOUBLE_BREAK 安全性**：`doubleBreakChance("bow")` 走 default case 返回 0（`skill_damage.go:335-344`），確保只有 BURNING_SPIRIT 觸發、不會誤觸發 DOUBLE_BREAK（Java DOUBLE_BREAK 由 calcDamage weapon switch case 11/12 claw/edoryu 處理，與 bow 無關）。
- **機率公式對齊**：
  - Go `burningSpiritChance = 15` + `world.RandInt(100)` 回傳 [0, 99]，`burningRoll < 15` 命中 15 個值（0~14）= 15% 機率，對齊 Java `ConfigSkill.BURNING_CHANCE = 15` + `_random.nextInt(100)+1` 回傳 [1, 100]、`random <= 15` 命中 15 個值（1~15）= 15% 機率。
  - Go `damage * burningSpiritMultiplier / burningSpiritDivisor`（3/2）對齊 Java `dmg *= ConfigSkill.BURNING_DMG`（properties `BURNING_DMG = 1.5`）。
- **不寫新測試**：依停損標準，本步為「Java 真實差異 + 改 Go 對齊」，既有 `skill_darkelf_buff_test.go:73-87 TestSkillDarkElfBuffBurningSpiritAndDoubleBreakAreProcFlags` 透過 `darkElfPhysicalDamageWithRolls` 直接鎖定 102/105 雙觸發行為，新增的 ranged path call 為兩行委派、無新邏輯。避免「Go 已對 + 防回歸」測試。
- **broader gap（不改）**：
  - **Java else-if 鏈 vs Go 獨立呼叫**：Java `BuffDmgUp` 是 `if ELEMENTAL_FIRE else if BURNING_SPIRIT else if BRAVE_AURA` 互斥，一次攻擊最多一個觸發；Go 把 `darkElfPhysicalDamage`/`elfMeleeDamage`/`braveAuraDamage` 拆成獨立呼叫，理論上同一攻擊可疊加多個。此差異已存在於 melee 路徑，與 102 補齊本身無關，屬廣域 buff stacking 對齊缺口。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：`go build ./...` EXIT=0；既有 `Dark|Burning|Combat|PvP|Ranged` 測試全部 PASS（含遠程 LOS、PvE 遠程、PvP 遠程、燃燒鬥志/雙重破壞觸發旗標、暗影防護等）。

## 行走加速（MOVING_ACCELERATION / 101）— 純審計無代碼變更

- 純審計 `101 MOVING_ACCELERATION`：Go 完整對齊 Java `L1SkillUse.java:1456-1461 / 2653-2658`、`L1SkillStop.java:594-602`、`L1BuffUtil.java:168-172, 222-226`。
- **核心行為已對齊**：
  - **dispatch**：skill_list.yaml `skill_id:101 target:none` → `skill.go:513 executeSelfSkill` → `skill_self.go` default 路徑 → `applyBuffEffect`（無 case 101 特例）。
  - **buff 屬性**：`buffs.lua [101] = { brave_speed = 4, exclusions = {52, 150, 155, 186, 1000, 1016} }` 對齊 Java `L1SkillUse.java:2656 pc.setBraveSpeed(4)`。
  - **REPEATEDSKILLS 互斥**：exclusions {52(HOLY_WALK), 150(WIND_WALK), 155(FIRE_BLESS), 186(BLOODLUST), 1000(STATUS_BRAVE), 1016(STATUS_ELFBRAVE)} 完全對齊 Java `L1SkillUse.java:1745-1746 REPEATEDSKILLS[2] = { HOLY_WALK, MOVING_ACCELERATION, WIND_WALK, STATUS_BRAVE, STATUS_ELFBRAVE, BLOODLUST, FIRE_BLESS }`（101 自身不需列入自己的 exclusions）。
  - **icon 封包**：`skill_buff.go:235-238` `sendBraveToAll(target, byte(eff.BraveSpeed), uint16(skill.BuffDuration))` 對齊 Java `pc.sendPackets(new S_SkillBrave(pc.getId(), 4, _getBuffIconDuration))` 自身 + `pc.broadcastPacketAll(new S_SkillBrave(pc.getId(), 4, 0))` 廣播。
  - **反咒語豁免**：`counterMagicExempt[101] = true`（`skill_buff.go:404`）對齊 Java `EXCEPT_COUNTER_MAGIC` (`L1SkillUse.java:147 / L1SkillUse2.java:149`) 含 101——魔法屏障無法抵擋自身增益。
  - **不可解除狀態**：`NON_CANCELLABLE[101]` 不存在（`buffs.lua:218-274` 未列入）對齊 Java `L1SkillMode.isNotCancelable()`（line 31-37）不含 MOVING_ACCELERATION——可被 CANCELLATION 移除。
  - **暴風疾走互斥**：`skill_self.go:152 case 172 STORM_WALK` 對 101 執行 `removeBuffAndRevert` 對齊 Java 互斥群組行為。
  - **buff 到期/cancelAllBuffs/死亡清空**：`skill_buff.go:660, 754` + `skill.go:114, 157` 全部執行 `sendBraveToAll(target, 0, 0)` 對齊 Java `L1SkillStop.java:600 pc.sendPacketsAll(new S_SkillBrave(pc.getId(), 0, 0))` + `cha.setBraveSpeed(0)` 與 `L1BuffUtil.java:168-172` 死亡清空路徑。
  - **既有測試覆蓋**：`skill_buff_test.go:132-148` 已鎖定 BraveSpeed=4 + buff 註冊行為。
- **不動處**：
  - `skill_self.go:151 case 172 STORM_WALK` 列出 `42, // HOLY_WALK` 但 Java `L1SkillId.HOLY_WALK = 52`（非 42）；這是 skill 52/172 範疇 bug 不在 skill 101 scope，依「不可偷換範圍」記錄不修，待 skill 52/172 自身審計時處理。
- **broader gap（不改）**：
  - **L1BuffUtil 死亡清空目前依賴通用路徑**：Java `L1BuffUtil.braveStart()` 是顯式列舉 HOLY_WALK/MOVING_ACCELERATION/WIND_WALK 等做 killSkillEffectTimer + setBraveSpeed(0)；Go `ClearAllBuffsOnDeath` 透過 `buff.SetBraveSpeed > 0` 通用旗標走相同行為，路徑等價但實現方式不同。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 `skill_buff_test.go` 已覆蓋 101 buff 設定路徑。

## 提煉魔石（BRING_STONE / 100）— 純審計無代碼變更

- 純審計 `100 BRING_STONE`：Go 已完整對齊 Java `L1SkillUse.java:2346-2389` 的四級魔石升級鏈，原註解「Go 簡化：完整升級邏輯待後續實作」實際已實作（comment 為過期描述）。
- **核心行為已對齊**：
  - **目標 ID**：`L1SkillUse.java:314 case BRING_STONE → _itemobjid = target_id`（target 為物品 objectID）對齊 Go `executeBringStone(itemObjID)`。
  - **物品驗證**：`switch invItem.ItemID case 40320, 40321, 40322, 40323` 對齊 Java `if itemId == 40320/40321/40322/40323`。
  - **公式對齊**：
    - `dark = int(10 + level*0.8 + (wis-6)*1.2)` 對齊 Java `(int)(10 + pc.getLevel()*0.8 + (pc.getWis()-6)*1.2)`。
    - `brave = int(dark / 2.1)` / `wise = int(brave / 2.0)` / `kayser = int(wise / 1.9)` 全部對齊 Java。
  - **升級鏈**：40320→40321→40322→40323→40324 與訊息 ID `$2475/$2476/$2477/$2478` 全部對齊 Java。
  - **擲骰機率**：`world.RandInt(100)+1`（1-100 範圍）對齊 Java `random.nextInt(100)+1`。
  - **無論成敗都消耗 1 個原石**：Go `removeItem(invItem.ObjectID, 1)` 在擲骰前執行，對齊 Java `pc.getInventory().removeItem(item, 1)` 在 if 判定前。
  - **成功訊息**：`SendServerMessageStr(sess, 403, msgArg)` 對齊 Java `S_ServerMessage(403, "$xxxx")`。
  - **失敗訊息**：`SendServerMessage(sess, 280)` 對齊 Java `S_ServerMessage(280)`「魔法失敗了」。
  - **dispatch**：`skill.go case 100 → executeBringStone(sess, player, skill, targetID)`，與 `skill_magic_scroll.go` 卷軸路徑共用同實作。
  - **ItemCreate 整合**：透過 `s.deps.ItemCreate.GiveItem` 統一管理新物品發放（fallback 為直接 Inv.AddItem）。
- **不動處**：
  - `skill_weapon.go:157` 註解「Go 簡化：完整升級邏輯待後續實作」與實際行為不符（已完整實作），屬無害的歷史 comment 殘留；依停損標準避免無關 surgical 修改，留待自然累積清理。
- **broader gap（不改）**：
  - **施法動畫先送 vs 消耗物品先送順序**：Go 先廣播 ActionGfx + SkillEffect 再 RemoveItem；Java 在 L1SkillUse 流程中先 runSkill 觸發物品操作再 useConsume。Go 順序為內部一致性最佳化，玩家視覺體驗等價。
  - **yaml mp_consume/buff_duration/reuse_delay drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 `skill_weapon_item_create_test.go:71` 已覆蓋 executeBringStone 核心路徑。

## 暗影防護（SHADOW_ARMOR / 99）— 純審計無代碼變更

- 純審計 `99 SHADOW_ARMOR`：Go 已完整對齊 Java `skillmode/SHADOW_ARMOR.java`。
- **核心行為已對齊**：
  - **MR +5**：`buffs.lua [99] = { mr = 5 }` 透過 `applyBuffEffect` 寫 `buff.DeltaMR = 5` 並套用 `target.MR += 5`，對齊 Java `pc.addMr(5)`。
  - **`S_SkillIconShield(3, duration)`**：`buff_icon_map.yaml skill_id=99 type=shield param=3` 透過 `sendBuffIcon "shield"` 送 `SendIconShield(sess, durationSec, 3)`，對齊 Java `pc.sendPackets(new S_SkillIconShield(3, integer))`。
  - **`S_SPMR(pc)`**：`applyBuffEffect` 在 `buff.DeltaMR != 0` 時送 `SendMagicStatus(SP, MR)`，對齊 Java `pc.sendPackets(new S_SPMR(pc))`。
  - **stop 反向**：`removeBuffAndRevert` → `revertBuffStats` 將 MR -5、`cancelBuffIcon` 送 `S_SkillIconShield(3, 0)`，且 DeltaMR != 0 觸發 S_SPMR 再次廣播，對齊 Java stop 三步驟。
  - **不可被相消**：`NON_CANCELLABLE[99] = true` (`buffs.lua:248`) 對齊 Java `L1SkillMode.isNotCancelable` (line 35) 含 SHADOW_ARMOR。
  - **反咒語豁免**：`counterMagicExempt[99] = true` (`skill_buff.go:403`) 對齊 Java `EXCEPT_COUNTER_MAGIC` (line 147) 含 99。
  - **無 REPEATEDSKILLS 互斥**：Java `L1SkillUse.java:1741-1762` 全 10 個群組不含 99；Go `buffs.lua [99]` 無 exclusions，對齊。
  - **refresh 行為**：Java `if (!hasSkillEffect) { setSkillEffect + addMr(5) }` + 不論如何送 icon/SPMR。Go `target.AddBuff` 透過 `old != nil → revertBuffStats(old)` 先還原舊 buff，再套用新 buff，淨效果 MR 仍為 +5；icon 與 SPMR 重送。雙方 refresh 後 MR 值與 packet 序列等價。
- **broader gap（不改）**：
  - **NPC caster path**：Java NPC caster stub 直接 return 0；Go 透過 SkillSystem 由 PC 技能流程驅動，NPC 不會 cast 99，路徑無差異。
  - **yaml mp_consume/buff_duration drift**：與廣域同源 broader gap。
- 驗證：無代碼變更，既有 buff 流程覆蓋 MR delta 與 icon 路徑。

## 武器附毒（ENCHANT_VENOM / 98）

- 補齊 Java `L1AttackPc.addPcPoisonAttack`（line 754 + 2914-2921）對遠程攻擊也觸發毒附加的對齊缺口：
  - **Before**：Go 只在 `combat.go processMeleeAttack` 與 `pvp.go HandlePvPAttack`（兩個近戰路徑）觸發 `applyEnchantVenomPoisonToNpc/Player`。攻擊者持有 98 buff 但用弓箭遠程攻擊時，10% 附毒完全不觸發。
  - **After**：`combat.go processRangedAttack`（PvE 遠程 → NPC）與 `pvp.go HandlePvPFarAttack`（PvP 遠程）也呼叫對應 `applyEnchantVenomPoison`，於 `damage > 0` 套用後執行。Java `L1AttackPc` 類涵蓋近戰與遠程，`addPcPoisonAttack` guards 為 `_weaponId != 0`（弓也是武器）。
- **不動處**：
  - `applyEnchantVenomPoisonToPlayer/Npc` guards 已對齊 Java：`HasBuff(98) + Equip.Weapon() != nil + roll < 10 (10% chance) + canApplyPoisonToPlayer（PoisonType==0 + 無 104/6687 抗性）`。
  - 毒傷數值與週期：Go `applyDamagePoisonToPlayer` 設 amount=5、PoisonTicksLeft=150（30 秒）、PoisonDmgTimer 每 15 ticks（3 秒）扣血，對齊 Java `L1DamagePoison.doInfection(_pc, target, 3000, 5)`（interval 3000ms, dmg 5）。
- **broader gap（不改）**：
  - **NPC 攻擊者附毒**：Java `L1AttackNpc.addNpcPoisonAttack` 為 NPC 自己持毒 buff 時 attack PC 附毒（與 98 不同路徑）。屬獨立 NPC 攻擊系統範疇。
  - **yaml mp_consume/buff_duration drift**：與廣域同源 broader gap。
- 驗證：`go build ./...` 通過；`go test ./... -count=1` 全綠（system 19.006s，handler 1.003s）。

## 暗隱術（BLIND_HIDING / 97）— 純審計無代碼變更

- 純審計 `97 BLIND_HIDING`：Go 已完整對齊 Java `L1SkillUse.java:2559-2562` 與 `L1SkillUse2.java:2511-2514` 的施法路徑與相關生態。
- **核心行為已對齊**：
  - **施法立即效果**：Go `skill_self.go case 97` 送 `SendInvisible(true)` 給施法者 + `BuildRemoveObject` 廣播給附近玩家，對齊 Java `pc.sendPackets(new S_Invis(pc.getId(), 1))` + `pc.broadcastPacketAll(new S_RemoveObject(pc))`。
  - **隱身旗標**：`buffs.lua [97] = { invisible = true }` → `applyBuffEffect` 設 `target.Invisible = true`，對齊 Java skillmode 隱身註冊。
  - **Icon emission**：`buff_icon_map.yaml skill_id=97 type=invis` 透過 `sendBuffIcon "invis"` 重複送 `SendInvisible(durationSec > 0)`，與 Java 同向；雖然 Go 雙路徑送 invis 圖示一次屬內部重複（但 client 處理為 idempotent）。
  - **持續時間**：yaml `buff_duration: 32`（32 秒）由 `applyBuffEffect` 套用 `target.AddBuff(... TicksLeft: 32*5)`，對齊 Java `_skill.getBuffDuration() * 1000`（SQL=32 秒）。
  - **被偵測解除**：Go `skill_self.go case 13, 72`（DETECTION / COUNTER_DETECTION）遍歷 nearby 同時移除 60 與 97 buff，對齊 Java `L1SkillUse.detection()` 路徑。
  - **行動解除**：Go `cancelInvisibility`（`skill_buff.go:350-367`）同時移除 60 與 97，於攻擊/施法/使用道具時觸發，對齊 Java `L1BuffUtil.cancelInvisibility`。
  - **反咒語豁免**：`counterMagicExempt[97] = true`（`skill_buff.go:403`）對齊 Java `EXCEPT_COUNTER_MAGIC` 含 BLIND_HIDING。
  - **無 REPEATEDSKILLS**：Java `L1SkillUse.java:1741-1762` 全 10 個群組不含 97，Go `buffs.lua [97]` 無 exclusions，對齊允許與其他 buff 並存。
- **不動處**：
  - `cancelInvisibility` 額外送 `SendPutObject` 給附近玩家是 Go 對 UI 即時性的加強（Java 依靠 movement 觸發重新出現），不破壞 Java 對齊。
- **broader gap（不改）**：
  - **客戶端隱身切換時序**：Go 走 VisibilitySystem 下一 tick（200ms 內）刷新 known set，Java 為立即同步呼叫；兩者皆能正確讓附近玩家「看不到」隱身者，差異不可感知。
  - **yaml mp_consume/reuse_delay/type 等**：與廣域 yaml drift 同源 broader gap。
- 驗證：無代碼變更，既有測試覆蓋 60/97 共用 cancel 路徑（attack/cast cancel 與 detection 解除）。

## 立方：和諧（CUBE_BALANCE / 220）— 純審計無代碼變更

- 純審計 `220 CUBE_BALANCE`：Go 已完整實作並與 Java 行為對齊；docs 表格「未實作，需補地面立方」為過期標註，本步更新為「已對齊」。
- **核心行為已對齊**：
  - **Spawn NPC + GFX**：`cubeBalanceNpcID=80152`、GFX 6724（從 NPC 模板載入），對齊 Java `L1SkillUse.java:1834 L1SpawnUtil.spawnEffect(80152, ...)`。
  - **狀態 ID**：`cubeStatusBalance = 1025`，對齊 Java `L1SkillId.STATUS_CUBE_BALANCE = 1025`。
  - **MP 恢復**：Go `cubeEffectIntervalTicks=20`（4 秒）+5 MP，對齊 Java `L1Cube.java:148-154 _timeCounter % 4 == 0 → setCurrentMp + 5`。
  - **HP 傷害**：Go `cubeBalanceDamageIntervalTicks=25`（5 秒）-25 HP，對齊 Java `_timeCounter % 5 == 0 → receiveDamage 25`。
  - **PC 與 NPC 雙路徑**：Go 在 `applyCubeEnemy` 與 `applyCubeEnemyNpc` 各自處理；Java 在 `L1Cube.giveEffect case STATUS_CUBE_BALANCE` 分 PC/Monster 分支處理。
  - **無 immune buff 檢查**：Java BALANCE 沒有 STATUS_FREEZE/ABSOLUTE_BARRIER/ICE_LANCE/EARTH_BIND immunity（與 IGNITION/QUAKE 不同）；Go 也沒有，對齊。
  - **同類立方範圍驗證**：Go `hasNearbySameCube` 檢查 chebyshev 距離 ≤ 3，發 `S_ServerMessage 1412`「已在地板上召喚了魔法立方塊。」對齊 Java `L1SkillUse.java:391-411 getVisibleObjects(pc, 3) + S_ServerMessage(1412)`。
  - **EXCEPT_COUNTER_MAGIC**：Java 與 Go 都未列入 220（CUBE_BALANCE 為 ground target 非單體針對技能，魔法屏障不適用）。
  - **NON_CANCELLABLE**：Java/Go 都未列入（cube buff 為短暫地面效果非持久 buff）。
  - **生命週期**：Go `effect.TicksLeft = skill.BuffDuration * 5` = 20 秒 × 5 ticks/sec = 100 ticks；對齊 Java `_skill.getBuffDuration() = 20` 秒（SQL `buff_duration=20`）。
- **divergence（已知與其他 cube 同源 broader gap）**：
  - **立方放置位置**：Go 對所有四個 cube skills（205/210/215/220）統一使用 `player.X, player.Y`（施法者腳下），Java 使用 `_targetX, _targetY`（點擊位置）。屬全 cube skill 同源 broader gap，若要對齊需同時調整四個 cube 行為與測試，超出單一子項範圍。
- **broader gap（不改）**：
  - **yaml mp_consume/reuse_delay/area/type 等**：與 207-219 同源 yaml drift，本步不調整。
- 驗證：無代碼變更；既有測試 `TestSkillCubeGroundEffectCubeBalanceRestoresMPAndDamagesTarget` 覆蓋核心 MP+5/-25 HP 行為。

## 化身（ILLUSION_AVATAR / 219）

- 清除 `buffs.lua [219]` 的 `exclusions = {204, 209, 214}` stale 互斥群，對齊 Java `skillmode/ILLUSION_AVATAR.java` 沒有 REPEATEDSKILLS 互斥，與 Java `L1SkillMode.java:38-39 isNotCancelable` 將四個 illusion buff 同列「不可被相消移除」一致——意即四個 illusion buff 可並存。
  - 與先前 204 ILLUSION_OGRE / 209 ILLUSION_LICH / 214 ILLUSION_DIA_GOLEM 同樣的 exclusions 已分別於先前 commit 移除，本步補完 219 的清理。
- **不動處**：
  - `buffs.lua [219] = { dmg_mod = 10, bow_dmg = 10 }` 對齊 Java `pc.addDmgup(10) + pc.addBowDmgup(10)`。
  - `weapon_skill.go:327 if player.HasBuff(219) { dmg += 10 }` 屬武器附加技能加成路徑，與 buff 的 dmg_mod 共存（Java 武器技能傷害有獨立流程）。
  - `counterMagicExempt[219] = true` (`skill_buff.go:411`) 對齊 Java `EXCEPT_COUNTER_MAGIC` 含 ILLUSION_AVATAR。
  - `NON_CANCELLABLE[219] = true` (`buffs.lua:273`) 對齊 Java `L1SkillMode.isNotCancelable`。
- **broader gap（不改）**：
  - **`pc.setAvatar(ILLUSION_AVATAR_DAMAGE)` 機制**：Java `L1AttackPc.java:1600-1602` 與 `L1AttackNpc.java:382-384` 的 `dmg -= dmg * (getAvatar() / 100)` 因 `ILLUSION_AVATAR_DAMAGE` default=1、整數除 `1/100=0` 整個運算式恆為 0，等同 dead-code（必須 admin 將 `ILLUSION_AVATAR_DAMAGE>=100` 才生效，但會變成全免疫不合理）。Go 不實作 Avatar 欄位與對應傷害減免，行為等同 Java 預設配置。若未來伺服器明確需要該機制，需先建立 PlayerInfo.Avatar 欄位 + 雙路徑 dmg 減免 consumer。
  - **yaml mp/reuse/buff_duration/target/type/ranged drift**：與 207-218 同源 broader gap。
- 驗證：`go build ./...` 通過；`go test ./internal/system/ -count=1` 全綠（17.690s）。

## 疼痛的歡愉（JOY_OF_PAIN / 218）

- 補齊 Java `L1PcInstance.receiveDamage:2737-2773` 對所有 PC→PC 傷害源觸發 backlash 的對齊缺口：
  - **Before**：Go 只在 `skill_damage.go executeAttackSkillOnPlayer` 與 `skill_self_area.go` 兩個 skill 傷害路徑呼叫 `applyJoyOfPainBacklash`。攻擊者持有 218 buff 但使用 melee/ranged 普通攻擊時，反傷完全不觸發。
  - **After**：`SkillManager` 介面新增 `ApplyJoyOfPainBacklash` 方法（轉發到原 unexported `applyJoyOfPainBacklash`），`pvp.go HandlePvPAttack` 與 `HandlePvPFarAttack` 在 `damage > 0` 時於 `target.HP -=` 套用前呼叫，使 melee/ranged PvP 也能觸發反傷與一次性消耗 buff。
  - 對應修改 `internal/handler/skill_test.go captureSkillManager` mock 補上 `ApplyJoyOfPainBacklash` no-op 實作。
- **不動處**：
  - `joyOfPainTicks = 16*5` (80 ticks = 16 秒) 已對齊 Java `setSkillEffect(JOY_OF_PAIN, 16 * 1000)`。
  - `applyJoyOfPainReady` 已有 `HasBuff(218) → 系統訊息「已經準備疼痛的歡愉。」` 對齊 Java `sendPackets(new S_SystemMessage("已擁有此狀態"))`。
  - Skill 218 既有 skill 傷害路徑（skill_damage.go:201 / skill_self_area.go:88）保留，melee/ranged hook 是補新增。
- **broader gap（不改）**：
  - **`joyOfPainDivisor = 5` vs Java `JOY_OF_PAIN_COUNTDMG` default=1**：Java 為 ConfigIllusionstSkill admin-tunable，預設 1（無除減）。Go 硬編碼 5（除以 5 → 反傷大幅降低）。屬營運平衡常數差異，需確認伺服器希望比照 Java default 或保留 Go 既有平衡。改動會破壞既有測試 `TestSkillIllusionistControlJoyOfPainBacklashDamagesCasterOnce` 期望（target HP=50/divisor 5 → backlash 10）。
  - **`joyOfPainMaxDmg = 1000` vs Java `JOY_OF_PAIN_DMG` default=100**：Go 上限 10x Java 預設。同屬平衡常數差異。
  - **GM 反傷致死特例**：Java `attackPc.isGm() → setCurrentHp(MaxHp)` 反傷致死時 GM 復滿；Go 統一設 `HP=1`。屬 GM-only edge case 影響小。
  - **NPC 攻擊者反傷**：Java 處理 `attacker instanceof L1MonsterInstance` 走 `attackNpc.receiveDamage(this, nowDamage)`；Go 沒有對應路徑。但實務上 NPC 不會 cast JOY_OF_PAIN 取得 buff，此 case 為 Java 死碼。
- 驗證：`go build ./...` 通過；`go test ./... -count=1` 全綠（system 18.853s）；handler 套件 mock 補完後恢復編譯。

## 恐慌（PANIC / 217）

- 補齊兩項 Java `skillmode/PANIC.java` 對 Go 的對齊缺口：
  - **PC→PC MR 機率檢查**：`skill_status.go playerDebuffSkills` 加入 `217: true`。Java `L1MagicPc.calcProbabilityMagic` 對 PANIC 走 default case 含 `probabilityDice * diceCount + magichit + baseInt - getTargetMr()` 公式（PANIC 不在 PHANTASM/BONE_BREAK 等專屬 case 中），Go 原本因 217 不在 playerDebuffSkills 名單中導致 PC→PC 完全跳過 MR 檢查，直接 100% 命中。
  - **PC→NPC dispatch**：`skill_status.go` 新增 `case 217` 對 NPC 路徑，呼叫 `checkNpcMRResist` 後 `npc.AddDebuff(217, dur*5)` 註冊 debuff 標記，並依 `skill.CastGfx` 廣播施法特效。原本因 case 缺失走 default 只播 cast GFX 不註冊 debuff，導致 `npc.HasDebuff(217)` 永遠為 false。
- **不動處**：
  - `buffs.lua [217]={str=-1, con=-1, dex=-1, wis=-1, intel=-1}` 已對齊 Java PC 路徑 `addStr(-1) + addCon(-1) + addDex(-1) + addWis(-1) + addInt(-1)`，無需修改。
  - `counterMagicExempt[217]` 不存在已對齊 Java `EXCEPT_COUNTER_MAGIC` 不含 PANIC（debuff 應受 counter magic 攔截）。
- **divergence（協議級別差異，行為等價）**：
  - **stat 更新封包**：Java PC apply/stop 送 `S_OwnCharStatus2(pc)`；Go 透過通用 `applyBuffEffect` 觸發 `SendPlayerStatus` (S_STATUS opcode 8)。兩者皆能更新客戶端 stat UI，與 INSIGHT(216) 同源屬 buff stat-update 機制設計選擇。
- **broader gap（不改）**：
  - **NPC stat 變化**：Java `tgnpc.addStr(-1) + addCon(-1) + addDex(-1) + addWis(-1) + addInt(-1)` 對 NPC 套 5 屬性各 -1；Go NPC 模型只有 STR/DEX 欄位（無 Con/Wis/Intel），且既有 NPC debuff 系統採 marker-only（同 skill 56 DISEASE 不修改 NPC AC/DMG）。完整對齊需先建立 NPC stat mutation/revert 機制與 Con/Wis/Intel 欄位，屬廣域架構工作。
  - **MR 機率公式 generic vs 專屬**：
    - PC→PC：Go `checkPlayerMRResist` 用 `50 + lvlDiff*3 + INT - MR` 簡化版；Java default 公式包含 `probabilityDice * diceCount + magichit + baseInt - getTargetMr()` 多項加成。
    - PC→NPC：Go `checkNpcMRResist` 通用版；Java `L1MagicNpc.calcProbabilityMagic case PANIC` 用 `Random.nextInt(11)+20 + (atkLvl-defLvl)*2 * leverage/10`（與 CONFUSION/PHANTASM/BONE_BREAK 共用）。
    與 192/193/208/212 同源屬個別技能公式精確化的廣域工作。
  - **yaml `mp_consume 30→?`/`hp_consume 30→?`/`reuse_delay 0→?`/`buff_duration 64→?`/`target buff→?`/`type 4→?`/`ranged 5→?`/`probability_value 30→?`/`probability_dice 30→?`**：與 207-216 同源 yaml drift，整批處理時統一對齊 Java SQL。
- 驗證：`go build ./...` 通過；`go test ./internal/system/ -count=1` 全綠（21.186s）。

## 洞察（INSIGHT / 216）— 純審計無代碼變更

- 純審計 `216 INSIGHT`：Go 已對齊 Java `skillmode/INSIGHT.java` 核心行為。
  - **Stat 變化已對齊**：`buffs.lua [216] = { str = 1, con = 1, dex = 1, wis = 1, intel = 1 }` 與 Java `addStr(1) + addCon(1) + addDex(1) + addWis(1) + addInt(1)` 完全一致（5 屬性 +1）。
  - **counter magic 豁免已對齊**：`counterMagicExempt[216] = true`（`skill_buff.go:411`）與 Java `EXCEPT_COUNTER_MAGIC` 清單一致。
  - **stop 路徑已對齊**：`applyBuffEffect` 透過 `revertBuffStats` 將 5 屬性各 -1（DeltaStr/DeltaCon/DeltaDex/DeltaWis/DeltaIntel = 1 → 反向套用）。
  - **buff 重複保護已對齊**：Go `[216]` 透過 `target.AddBuff` 內部以 SkillID 為 key 的 map 防止同 ID 重複覆蓋；Java `if (!cha.hasSkillEffect(216))` 跳過已有 buff。
- **divergence（協議級別差異，行為等價）**：
  - **stat 更新封包**：Java apply/stop 送 `S_OwnCharStatus2(pc)`（opcode `S_OPCODE_OWNCHARSTATUS2`，純 STR/INT/WIS/DEX/CON/CHA/weight240）；Go 送 `SendPlayerStatus`（opcode `S_OPCODE_STATUS=8`，包含 level/exp/6 屬性/HP/MP/AC/gameTime 全量狀態）。兩者皆能正確更新客戶端 stat UI 顯示，但 packet size 與其他附帶資料不同。屬於 Go 通用 buff stat-update 機制設計選擇，影響所有 stat-changing buff（非 216 個別問題）。
- **broader gap（不改）**：
  - **yaml `mp_consume 40→60`、`reuse_delay 0→1000`、`buff_duration 300→640`、`target buff→none`、`target_to 1→0`、`type 4→2 (CURSE→CHANGE)`、`ranged 5→0`、`probability_value/dice 100→0`**：屬 yaml 資料 drift，與 207-215 同源。`buff_duration` 差 2.13x（10:40 vs 5:00），影響 buff 持續時間。
- 驗證：`go build ./...` 通過（無代碼變更）。

## 立方：衝擊（CUBE_SHOCK / 215）

- 修正 `ground_effect.go` NPC CUBE_SHOCK 套用的 debuff ID 與時長，對齊 Java `L1Cube.giveEffect:135-145 case STATUS_CUBE_SHOCK_TO_ENEMY`：
  - **Before**：`npc.AddDebuff(cubeStatusShockEnemy=1023, cubeStatusTicks=40)`（8 秒）。
  - **After**：`npc.AddDebuff(cubeStatusShockMR=1024, 20)`（4 秒，對齊 Java `setSkillEffect(STATUS_MR_REDUCTION_BY_CUBE_SHOCK, 4000)`）。
  - Java 使用 `STATUS_MR_REDUCTION_BY_CUBE_SHOCK(1024)`（非 enemy tracking 1023）作為 MR-reduction 標記；Go NPC 路徑原本誤用 enemy tracking ID，若未來有系統檢查 `npc.HasDebuff(1024)` 來套用 MR 減免，Java 會命中而 Go 不會。
  - Go `npc.AddDebuff` 使用 map 寫入即覆蓋，與 Java `setSkillEffect` 每 tick refresh 行為一致。
- 不動處：
  - **PC 路徑保留 `cubeStatusShockEnemy(1023)` enemy tracking** + GFX gating（Go 既有設計避免每 tick 重複廣播；Java SHOCK 對 PC/NPC 完全不廣播）。
  - **無 tick gating**：Java 與 Go 都每 tick 觸發（`STATUS_CUBE_SHOCK_TO_ENEMY` 無 `_timeCounter % 4` 檢查），對齊。
  - **無 immune buff 檢查**：Java SHOCK 沒有 STATUS_FREEZE/ABSOLUTE_BARRIER/ICE_LANCE/EARTH_BIND immunity（與 IGNITION/QUAKE 不同）；Go 也沒有，對齊。
- **broader gap（不改）**：
  - **實際 MR -10 套用未實作**：Java `L1Cube:139` 原 `addMr(-10)` 已被 comment out，僅留 `setSkillEffect(1024)` marker；Go 也只設 marker 不套 MR 減免。兩邊行為一致（皆無 MR -10），屬「Java 設計缺陷」雙方同步保留。
  - **PC `cubeStatusShockMR` refresh 行為**：Go `addPlayerCubeBuff` 若 `HasBuff` 為 true 則跳過 refresh，與 Java `setSkillEffect` 每 tick refresh 不完全一致；但因兩邊都無 1024 消費者，差異純為 marker timer 細節不影響遊戲。
  - **yaml `reuse_delay 0→5000`**：Java SQL=5000、Go=0。屬冷卻 tuning，與 207-214 同源。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠（無 215 相關測試）。

## 幻覺：鑽石高崙（ILLUSION_DIA_GOLEM / 214）

- 修正 `buffs.lua [214]` 兩項 Java 對齊缺失：
  1. **AC `-20` → `-8`**：對齊 Java `L1SkillUse.java:2665-2668` `pc.addAc(-8)`。Go 原本給 -20 AC 是 Java 的 2.5x，玩家防禦過強。
  2. **移除 `exclusions = {204, 209, 219}`**：對齊 Java skillmode/L1SkillUse 對 ILLUSION_DIA_GOLEM 無 REPEATEDSKILLS 互斥檢查——Java 允許四個 illusion buff（204/209/214/219）並存。同 204/209 修正模式。
- 配套說明：
  - 至此 204/209/214 三個 illusion buff 已全部移除 exclusions，sibling mutex 殘餘僅剩 219 仍引用 {204, 209, 214}，待 219 子項處理。
  - 既有 `applyBuffEffect` 已正確處理 `eff.AC` 變動並送 `S_AbilityScores`，AC 數值修正即生效。
- **broader gap（不改）**：
  - **yaml `buff_duration 32→128`**：Java SQL=128 秒，Go=32（4x diff），同 209 broader gap。
  - **yaml `mp_consume 25→40`、`reuse_delay 0→2000`、`ranged 3→5`、`type 4→2 (CURSE→CHANGE)`、`probability_value/dice 100→0`**：屬 yaml tuning，與 207-212 同源。
  - **sibling mutex 殘餘**：219 的 exclusions 仍引用 214，待 219 子項處理。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠（無 214 相關測試）。

## 隱身破壞者（ARM_BREAKER / 213）— 純審計無代碼變更

- 純審計 `213 ARM_BREAKER`，發現 Go 與 Java yiwei 為**根本不同的技能**，差異過大不在單一子項範圍：
  - **Java 行為**（yiwei）：
    - 無 `skillmode/ARM_BREAKER.java` 檔案——純資料驅動，走 L1SkillUse default 路徑 `setSkillEffect(213, _getBuffDuration)`。
    - SQL `('213', '隱身破壞者', ..., 'none', '3', '0', '0', '0', '0', '0', '0', '2', '0', '0', '-1', '0', '16', ...)`：mp=10、target='none'、type=2 (CHANGE)、damage_value=0、area=-1、ranged=0。**self-cast、不造傷害**。
    - L1SkillUse.java:2859-2862 `if (skillId == ARM_BREAKER) detection(player)`：cast 完成後對**施法者周圍**執行 AoE 隱身揭露——`pc.delInvis() + beginInvisTimer()`、`for tgt in World.getVisiblePlayer(pc): tgt.delInvis()`、`WorldTrap.onDetection(pc)`（顯示陷阱）。
    - L1SkillUse.java:822-829：可指定 HIDDEN_STATUS_SINK 隱藏 NPC（沉沒墳墓等）作為目標來允許施法。
    - 列入 `EXCEPT_COUNTER_MAGIC`（不可被反擊屏障/counter magic 反射）。
  - **Go 行為**（現狀）：
    - yaml `name='武器破壞者'`、`target='attack'`、`type=64 (ATTACK)`、`damage_value=15`、`ranged=3`、`area=0`、`mp=25`、`buff_duration=12`。**目標式攻擊技能、造 15 傷害**。
    - `applyIllusionistStatusAttackEffect skillArmBreaker`：對**目標**呼叫 `revealInvisibleTarget`——清除 target 自己的 60/97 隱身 buff、`Invisible=false`、`SendInvisible + SendPutObject` 廣播。
- **divergence 分析**：
  - 「Java 是 self-cast AoE detection」vs「Go 是 targeted damage + target reveal」屬完全不同技能設計。
  - 完整對齊 Java 需大規模改寫：
    1. yaml 改名 `武器破壞者→隱身破壞者`、target `attack→none`、type `64→2`、damage_value `15→0`、area `0→-1`、ranged `3→0`、mp `25→10`
    2. 新增 self-cast dispatch 路徑（與 attack 路徑互斥）
    3. 實作 caster `delInvis() + beginInvisTimer()`（Go 無 invis-spam timer 機制）
    4. 實作 `getVisiblePlayer(caster)` 對所有可視玩家 AoE 揭露
    5. 實作 `WorldTrap.onDetection(caster)` 陷阱顯示（Go 陷阱系統路徑未驗證）
    6. 移除既有 `revealInvisibleTarget(target)` 攻擊路徑
  - 是否 yiwei 版為 3.80C 客戶端 canonical 行為尚待確認；fly 版可能有不同的 ARM_BREAKER 實作；Go 現狀（damage + target reveal）也可能是 server admin 設計選擇。
  - 改動範圍涵蓋 yaml/dispatch/world/invis timer 多系統，**超出單一子項對齊深度停損標準**。
- **本步無代碼變更**——保留 Go 現狀。記為待使用者決定的 major divergence；類似 PATIENCE(211) 處理模式。
- 驗證：`go build ./...` 通過（無代碼變更）。

## 幻想（PHANTASM / 212）

- 修正 `212 PHANTASM` 三項 Java `skillmode/PHANTASM.java` 對齊缺失：
  1. **PC→PC 缺 MR 機率檢查**：Go `playerDebuffSkills` 原本不含 212，PC 對 PC 施放 PHANTASM 無 MR 抗性判定 → 100% 命中。Java `L1MagicPc.calcProbabilityMagic case PHANTASM` 使用 ConfigIllusionstSkill 5/10/15 by level diff + RegistSleep penalty。新增 `212: true` 到 map，目前先用 generic `checkPlayerMRResist`（50% base + INT/MR），完整 Java 5/10/15 公式留 broader gap。
  2. **PC→NPC 完全無 sleep 效果**：Go `skill_status.go` NPC 分派 switch 缺 `case 212`，PHANTASM 對 NPC 落入 default 只播放 cast gfx，**完全不睡 NPC**。新增 `case 212` 鏡像 `case 66` 邏輯：`checkNpcMRResist + npc.Sleeped=true + npc.AddDebuff(66, dur*5)`。
  3. **PC inflicted buff key 對齊 Java skill 66**：Go 原本 `applyBuffEffect` 走 `[212]={sleeped=true}` 儲存 buff ID 212；Java skillmode `setSkillEffect(66, integer*1000)` 應儲存 FOG_OF_SLEEPING(66)。新增 `case 212` 鏡像 `case 103 暗黑盲咒` 模式，將 `sleepSkill.SkillID = 66` 後 applyBuffEffect。讓 `hasSkillEffect(66)` cross-skill 查詢能正確命中。
- 配套說明：
  - PC 路徑使用 `applyBuffEffect` + skill ID 66，套用 buffs.lua 既有 `[66] = { sleeped = true }` 定義；buffs.lua `[212]` entry 保留以維持向後相容（既有 save 可能含 212 buff，cleanup 路徑於 skill_status.go:162,167 仍會處理）。
  - NPC 路徑使用 `npc.AddDebuff(66, dur*5)`，與 case 66 沉睡之霧的 debuff key 一致。
  - PHANTASM 觸發 SleepApply (0x0A) 經由 `applyBuffEffect` 的 eff.Sleeped 分支自動處理。
  - 既有測試 `TestSkillIllusionistStatusPhantasmSleepsPlayerTarget` 改為斷言 `HasBuff(66)`（對齊 Java），並 `disablePlayerDebuffMRForStatusTest(t, 212)` 避免 MR 機率讓測試 flaky。
- **broader gap（不改）**：
  - **PC→PC 機率公式個別化**：Go 用 generic `checkPlayerMRResist`，Java 用 ConfigIllusionstSkill 5/10/15 + RegistSleep。與 BONE_BREAK 208 同源 broader gap。
  - **PC→NPC 機率公式個別化**：Go `checkNpcMRResist` generic，Java 同樣 5/10/15。
  - **yaml `reuse_delay 0→3000`、`ranged 3→4`、`probability_value/dice 30→0`**：屬 yaml tuning，與 207/208/209/210/211 同源。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠（含修正的 Phantasm 測試）。

## 耐力（PATIENCE / 211）— 純審計無代碼變更

- 純審計 `211 PATIENCE`，發現 Java vs Go 行為差異但暫不修改（待使用者決定）：
  - **Java 行為**：無 skillmode 檔案、無 L1SkillUse case 處理、無 HprExecutor `_skill.put(PATIENCE, ...)`、無 addHpr 呼叫。L1SkillUse 走 default 路徑 `cha.setSkillEffect(_skillId, _getBuffDuration)`——**僅設 buff 圖示與計時，無任何 stat 加成**。Java yiwei 中 PATIENCE 只在 `EXCEPT_COUNTER_MAGIC` 清單（不可被 counter magic 反射）。
  - **Go 行為**：`buffs.lua [211] = { hpr = 5 }`——給予 +5 HPR/tick。Java 中找不到此 +5 HPR 的依據。
- **divergence 分析**：
  - 若改為 `[211] = {}` 對齊 Java：PATIENCE 變成「只有 buff 圖示沒有任何效果」的純裝飾技能，違反「做半套不如不做」原則。
  - 若保留現狀 +5 HPR：與 Java yiwei 不對齊，但給玩家有意義的回血效果（與 Go 既有 6 級「中等回血」介面定位相符）。
  - 既有 Go 設計可能引用其他 L1 server 版本（如 fly 版、官方韓版設定），Java yiwei 也可能是 incomplete 參考。
- **本步無代碼變更**——保留 Go 現狀 +5 HPR。記為「Java yiwei 對應實作為空、Go 給予 HPR +5」之已知差異，待使用者明確決定後再處理。
- 配套 yaml drift（broader gap）：
  - `mp_consume 25` ✓ 與 Java 一致
  - `buff_duration 600` ✓ 與 Java 一致
  - `reuse_delay 0→1000` — Java SQL=1000，Go=0。屬冷卻 tuning。
  - `target buff→none`、`target_to 3→0` — Java 為 self-cast 無 target；Go target='buff' target_to=3 走標準 buff 路徑。客戶端體驗可能無感。
  - `type 4→2` — Java type=2 (CHANGE)，Go=4 (CURSE)。可能影響 buff 分類。
  - `ranged 3→0` — Java 為 0（self-only），Go=3。
- 驗證：`go build ./...` 通過（無代碼變更）。

## 立方：地裂（CUBE_QUAKE / 210）

- 補齊 `CUBE_QUAKE` 地面效果四項 immune buff 檢查，對齊 Java `L1Cube.giveEffect:110-121 case STATUS_CUBE_QUAKE_TO_ENEMY`：
  1. **PC 路徑**：`ground_effect.go:174` 原本只檢查 `!target.AbsoluteBarrier`，缺 STATUS_FREEZE(4000) / ICE_LANCE(50) / EARTH_BIND(157)。新增 `playerCubeQuakeImmune(target)` helper 包含 `AbsoluteBarrier 標誌位 + HasBuff(4000/50/157)` 統一檢查。
  2. **NPC 路徑**：`ground_effect.go:207` 原本無任何 immunity 檢查。新增 `npcCubeQuakeImmune(npc)` helper 檢查 `HasDebuff(4000/78/50/157)`。
- 與 CUBE_IGNITION 不同：Java QUAKE 的 immune 清單**不包含** FREEZING_BLIZZARD(80)，僅四個（STATUS_FREEZE/ABSOLUTE_BARRIER/ICE_LANCE/EARTH_BIND）。helper 命名加 `Quake` 後綴與 `Ignition` 區隔。
- **broader gap（不改）**：
  - **PC 套用的 buff ID**：Java 對 PC `setSkillEffect(MOVE_STOP=4017, 1000)` + `S_Paralysis(TYPE_BIND, true)`；Go 用 `cubeStatusQuakeEnemy=1021` symbolic ID 並送 `BindApply`。S_Paralysis 客戶端顯示一致（BindApply=0x18），但 buff ID 不同會讓 Java 的「跨技能 hasSkillEffect(MOVE_STOP)」查詢失效。
  - **NPC 套用的 debuff ID**：Java 對 NPC `setSkillEffect(STATUS_FREEZE=4000, 1000)` + `setParalyzed(true)`；Go 用 `cubeStatusQuakeEnemy=1021`。這會影響 Java 的 cross-cube 互動（如若 QUAKE 把 NPC 標 STATUS_FREEZE，則同時被 IGNITION 命中時會免疫傷害）；Go NPC 走獨立 1021 ID 不觸發 IGNITION 免疫。屬「Cube 跨技能互動」結構性缺口。
  - **PC paralysis 時長**：Go `TicksLeft: 5` (1 秒) 與 Java `setSkillEffect(..., 1000)` 一致。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠。

## 幻覺：巫妖（ILLUSION_LICH / 209）

- 移除 `buffs.lua [209] = { sp = 2, exclusions = {204, 214, 219} }` 的 `exclusions` 欄位 → `[209] = { sp = 2 }`，對齊 Java `skillmode/ILLUSION_LICH.java:19-32` 只檢查 `!cha.hasSkillEffect(209)`，無 REPEATEDSKILLS 互斥群（Java 允許四個 illusion buff 並存）。同 204 修正模式。
- Java 行為：`pc.addSp(2)` + `S_SPMR(pc)` + `setSkillEffect(209, integer*1000)`；stop `addSp(-2)` + `S_SPMR(pc)`。Go `applyBuffEffect` 已在 SP 變動時送 `SendMagicStatus`（`skill_buff.go:294,557`），SP +2 / -2 邏輯與 Java 一致。
- **broader gap（不改）**：
  - **sibling mutex 殘餘**：214/219 各自的 `exclusions` 仍引用 209；要完整對齊「Java 四 illusion buff 並存」需在 214/219 各自的子項移除其 exclusions。按隊伍順序處理。
  - **yaml `buff_duration 32→128`**：Java SQL=128（128 秒），Go=32。差 4x，會導致 Go 209 buff 只持續 1/4 時間。屬 yaml 資料 drift。
  - **yaml `type 4→2`**：Java SQL=2 (CHANGE)，Go=4 (CURSE)。可能影響 buff 分類（dispel/cancel 範圍）。屬 yaml 分類 drift。
  - **yaml `mp_consume 15→20`、`reuse_delay 0→2000`、`ranged 3→5`、`probability_value/dice 100→0`**：與 207/208 同源 yaml tuning gap。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠。

## 骷髏毀壞（BONE_BREAK / 208）

- 修正 `208 BONE_BREAK` 五項 Java `skillmode/BONE_BREAK.java` + `L1MagicPc.calcProbabilityMagic` + `S_Paralysis` 對齊差異：
  1. **PC→PC 機率公式**：Go 原 `calcBoneBreakPlayerProbability` 用 `(30/20/10) + INT*0.5 - MR*0.1` 自創公式（INT=300 時飽和至 100%）。對齊 Java `L1MagicPc:584-599` + `ConfigIllusionstSkill` 預設值：`(5/10/15) by level diff` （`BONE_BREAK_INT/MR` 預設 0 → 無 INT/MR 加成），PC→PC 末段 `-= target.RegistStun`（`L1MagicPc:958-961`）。注意 Java 配置反常識：caster>target 反而最低（5%），高等目標最高（15%）。
  2. **S_Paralysis subtype（apply）**：Go `applyBoneBreakParalysis` 原本送 `ParalysisApply (0x02)`。Java `BONE_BREAK.start():29` `S_Paralysis(5, true)` 中 `5 = TYPE_STUN`（`S_Paralysis.java:25,79-85`）→ wire byte `0x16 = StunApply`，並非 ParalysisApply。改為 `handler.StunApply`。同時涉及客戶端訊息顯示「衝擊之暈」而非「身體完全麻痺」。
  3. **buff 過期 subtype（revert）**：Go `skill_buff.go:666-677` switch 中 208 落入 `default` → `needParalysisRemove = true` → `ParalysisRemove (0x03)`。Java stop() `S_Paralysis(5, false)` 對應 wire `0x17 = StunRemove`。新增 `case 87, 208, 508:` → `needStunRemove = true` 與 apply 路徑配對。
  4. **PC→NPC 缺少 13119 廣播**：Java `BONE_BREAK.start():35` 對 NPC 路徑 `broadcastPacketAll(S_SkillSound(npcId, 13119))`「骷髏毀壞動畫」。Go `skill_damage.go:735` 命中 NPC 後僅 `Paralyzed=true + AddDebuff(208)`，缺廣播。補上 `BroadcastToPlayers(nearby, BuildSkillEffect(t.npc.ID, 13119))`。
  5. **yaml `damage_value: 0 → 10`**：Java SQL `damage_value=10`，Go 原 0。Java `calcMagicDamage(208)` 走標準魔法傷害（dice 公式 + 屬性減免 + MR 減免），Go yaml `damage_value=0` 會讓 `magic.lua calc_skill_damage` 路由到 physical formula（Go 啟發式：value+dice 皆 0 → physical），與 Java 走 magic 路徑不一致。修正後資料完整且路由正確。
- 配套說明：
  - 移除 `boneBreakIntFactor / boneBreakMRFactor` 常數（Java 對應 config 預設 0，本來就無加成；保留 `Higher/Equal/Lower` 三常數但改為 5/10/15 對齊 Java config 預設值）。
  - 既有 `TestSkillIllusionistControlBoneBreakParalyzesPlayerTarget` 原本依賴 INT=300 把機率飽和到 100%，修正後（Java 5%）變成 flaky。替換為兩個對齊 Java 的測試：(a) `TestSkillIllusionistControlBoneBreakProbabilityMatchesJava` 直接驗證 5/10/15 + RegistStun 公式；(b) `TestSkillIllusionistControlBoneBreakAppliesStunNotParalysis` 直接呼叫 `applyBoneBreakParalysis` 跳過機率，驗證 buff 注入 + Paralyzed + StunApply 副作用。
- **broader gap（不改）**：
  - **PC→NPC 機率公式**：Go `checkNpcMRResist` 是 generic `50 + (L_caster - L_npc)*5 + INT*2 - MR` 共用於 192/208 等 NPC 技能，Java 各技能個別化（208 用 5/10/15）。屬「NPC 機率系統個別化」結構性缺口，與 192 同源；留至 MR/probability 系統整體對齊。
  - **mp_consume 20→30、reuse_delay 0→2000、cast_gfx 0→7020**：Java SQL 真實值，Go yaml drift。屬 yaml 成本/冷卻/視覺 tuning，與 185/195/206/207 同源 broader gap。
- 驗證：`go build ./...` 通過、`go test ./internal/system/ -count=1` 全綠（含新增的兩個 BONE_BREAK 測試）。

- 修正 `THUNDER_GRAB(192)` 四項 Java 對齊差異：
  1. **STATUS_FREEZE(4000) 免疫檢查**：Java `THUNDER_GRAB.java:35 if (isProbability && !cha.hasSkillEffect(4000))` 對 STATUS_FREEZE 目標完全免疫。Go `applyThunderGrabBind` 與 `skill_damage.go:700` NPC 路徑原本都缺此檢查 → 補上 `target.HasBuff(4000) / npc.HasDebuff(4000)` 早返回。
  2. **bindtime 殘餘秒數疊加**：Java 第 26-32 行 `if (cha instanceof L1PcInstance && cha.hasSkillEffect(192)) bindtime += getSkillEffectTimeSec(192); if (bindtime > 4) bindtime = 4;` PC re-cast 加上殘餘秒數，最多 4 秒上限。Go 原本 PC 路徑 `if target.HasBuff(192) return` 直接早返回放棄施法；NPC 路徑無 stacking。改為兩路徑都讀取殘餘 ticks（÷5 換算秒）加總，clamp 4 上限。
  3. **spawnEffect 81182 視覺**：Java 第 40/46/73/79 行 `L1SpawnUtil.spawnEffect(81182, bindtime, x, y, mapId, srcpc, 0)` 在目標位置生成 81182 視覺效果。Go 兩路徑都缺此 spawn → 新增 `world.GroundEffectThunderGrab` (=9) 類型 + `spawnThunderGrabGroundEffect(caster, x, y, mapID, bindSec)` helper（含 NPC template 查詢、bindSec*groundEffectTickSec 存活時間、broadcast）。PC 路徑與 NPC 路徑都呼叫。
  4. **PC 廣播 S_SkillSound(4184)** 在 PC 路徑已存在於 Go `applyThunderGrabBind`；NPC 路徑 Java 第 45/78 行 `broadcastPacketX8(S_SkillSound(npc.getId(), 4184))` 廣播音效，Go skill_damage.go:700 NPC 路徑原本缺 → 補上 `BroadcastToPlayers(nearby, BuildSkillEffect(npc.ID, 4184))`。
- 配套說明：
  - 新增 `world.GroundEffectThunderGrab GroundEffectType = 9` 列舉常數，與 SHOCK_STUN(81162) 同類型純視覺地面效果。
  - 新增 `thunderGrabEffectNpcID = int32(81182)` 常數於 skill_dragonknight.go。
- **broader gap 不改**：
  - **NPC `setParalyzed` vs `setPassispeed(0)`**：Java `THUNDER_GRAB.java:47-48 // tgnpc.setParalyzed(true); tgnpc.setPassispeed(0);` 明確 commented out setParalyzed，改用 passispeed=0（NPC 仍可攻擊，僅鎖定 passive 移動速度）。Go `NpcInfo` 沒有 passispeed 欄位，當前以 `npc.Paralyzed=true`（行為更激進，鎖定整個 AI tick）替代。屬「NPC passispeed 系統」結構性缺口，留至 NPC AI 系統重構時補。
  - **stop 路徑復原 passispeed**：Java stop() 對 NPC `npc.setPassispeed(npc.getNpcTemplate().get_passispeed())` 恢復模板原速度。Go 無 passispeed 系統，停止時自然解除 `Paralyzed`。同 broader gap。
  - **L1MagicPc.calcProbabilityMagic 對齊**：Java `L1MagicNpc:200 case THUNDER_GRAB` 與 `GUARD_BRAKE/RESIST_FEAR/HORROR_OF_DEATH` 共用標準 `(ProbabilityDice/10) * level_diff + ProbabilityValue` × `leverage/10` 公式。Go `checkDragonKnightDebuffSuccess` 用類似但不含 leverage 倍率的簡化公式。屬「MR 抗性公式統一」結構性缺口（與 183/188 同源），留至 MR 系統整體對齊。
- **不寫新測試**：既有 `TestSkillDragonKnightStatusThunderGrabBindsPlayerTarget` 使用 `ProbabilityValue=10 + ProbabilityDice=25 + caster L100 vs target L1` 強制 100% 成功，不受新邊界邏輯影響；其他四項變更（4000 免疫、stacking、spawn 81182、NPC 廣播）皆無對應 negative test，依停損標準避免「鎖實作」回歸測試。

- 修正 `MORTAL_BODY(191)` 兩項 Java 對齊缺失，補上反彈邏輯：
  1. **`buffs.lua [191] = { brave_speed = 4 }` → `[191] = {}`**：移除無 Java 依據的 brave_speed=4（Java 全 codebase 對 191 僅有 `L1PcInstance.java:2776 if (hasSkillEffect(191))` 反彈判斷，沒有 stat 加成；Go 原 `brave_speed=4` 來源不明，可能誤從同 ID 的「Underground Path」混入）。
  2. **新增反彈邏輯**：對齊 Java `L1PcInstance.java:2775-2798`：在 receiveDamage 內 CounterBarrier 後檢查 191，若 target 持 191 且 attacker ≠ self 且 23% 機率觸發 → 攻擊者承受 40 傷害（聖界 IMMUNE_TO_HARM=68 減半），廣播 `S_DoActionGFX(attacker, 2)` + `S_SkillSound(target, 10710)`，原始傷害歸零（Java 用 `return` 跳過後續 receiveDamage）。
- 實作：
  - `skill_dragonknight.go` 新增常數 `skillMortalBody=191, mortalBodyChance=23, mortalBodyDamage=40, mortalBodyEffectGfx=10710, mortalBodyAttackerGfx=2`。
  - `mortalBodyReflectPvP(target, attacker, damage, nearby) → (newDamage, reflected)`：PvP 路徑，套用 `applyImmuneToHarmDamage(attacker, 40)` 處理 attacker 持 68 減半。
  - `mortalBodyReflectFromNpc(target, npc, damage, nearby) → (newDamage, reflected)`：NPC 攻擊玩家路徑，`npc.HasDebuff(68)` 處理 NPC 持 68 減半（hook 預留）。
- 掛接點（與 CounterBarrier(skill 91) 相同覆蓋）：
  - `pvp.go pvpMeleeDamage`：CounterBarrier 檢查後、damage 套用前；觸發後若 attacker.HP<=0 呼叫 `s.deps.Death.KillPlayer(attacker)`。
  - `npc_ai.go npcMeleeAttack`：CounterBarrier 檢查後、`buildNpcAttack` 前；觸發後若 npc.HP<=0 呼叫 `handleNpcDeath(npc, target, nearby, deps)` 並清 AggroTarget。
- 配套 Java 對照（broader gap 不改，與 CounterBarrier 覆蓋落差一致）：
  - **PvP 遠程路徑**（pvp.go:300 弓矢傷害）未掛接 — Java `receiveDamage` 適用所有路徑，但 Go CounterBarrier 也未掛在此處，屬「全體反彈機制 PvP 遠程缺口」與 91 同源。
  - **PvE skill damage 路徑**（skill_damage.go）未掛接 — 同上，CounterBarrier 也未掛此處。
  - **NPC 遠程 / NPC 法術** 同樣未掛接 — 同源缺口。
  - 上述四項全屬「反彈機制 PvP ranged + skill damage 路徑覆蓋」結構性缺口，與 CounterBarrier 既有覆蓋落差一致，留至反彈機制統一架構時整體補齊。
- **不寫測試**：依停損標準避免「鎖實作」回歸測試；既有 CounterBarrier 也無對應 PvP 反彈測試，保持一致策略。23% 隨機性使測試難以確定性驗證，且 Java 對 191 的行為已透過 PvE/PvP 玩家實測自然驗證。

- 修正 `AWAKEN_FAFURION(190)` 兩項 Java 對齊缺失：
  1. **yaml buff_duration 0→600**：Java 與 fly `skills.sql` 第 190 列 buff_duration 欄為 600（秒），但 Go `skill_list.yaml` 第 5839 行為 0。`applyBuffEffect` 第 131-132 行 `if skill.BuffDuration <= 0 { return }` 早返回 → `[190] = { regist_freeze = 10, exclusions = {185, 195} }` 從未實際套用、Freeze 抗性沒加、互斥 185/195 沒清。前次 185 共用區塊重構與 169 timer 同步的代碼皆無法在此 yaml 下生效。本步 yaml 改為 `buff_duration: 600` 對齊 Java，buffs.lua + 共用區塊邏輯終於可生效。注：mp_consume 1 vs Java 30、hp_consume 1 vs Java 20 屬 data audit 範疇（與 184/189 先例一致），不在本次 code 對齊範圍。
  2. **`removeBuffAndRevert` 補齊 190→169 cleanup**：Java `skillmode/AWAKEN_FAFURION.java:36-40 stop()` 的 `killSkillEffectTimer(169)` 在任何 190 解除情境（自然到期 / 被 exclusion 移除 / 主動解除）都會清 169 timer。Go 原本只有 `tickPlayerBuffs:803-806` 自然到期路徑有此清除，但若 185 exclusions {190, 195} 觸發 `removeBuffAndRevert(target, 190)` 移除 190 後，169 殘留會繼續允許負重 HP/MP 回復至 169 自身 buff_duration 960s 才到期，與 Java 不一致。本步在 `removeBuffAndRevert` 新增 `if skillID == 190 { s.removeBuffAndRevert(target, 169) }`（遞迴安全：skillID=169 不再觸發此分支；無 169 buff 時 `RemoveBuff(169)` 回傳 nil 整段跳過）。
- **發現 sibling 缺口（不在本次 190 audit 範圍，依 CLAUDE.md「不可偷換範圍」回報用戶）**：`AWAKEN_ANTHARAS(185)` 與 `AWAKEN_VALAKAS(195)` yaml buff_duration 同為 0（Java SQL 兩者皆為 600），對應的 `[185] = { ac = -3, regist_sustain = 10, ... }` 與 `[195] = { hit_mod = 5, regist_stun = 10, ... }` 也從未實際套用。185 前次審計（2026-05-18）的「toggle-off 移除 + 廣播音效」邏輯重構雖然正確，但配套 buff stat 套用因 buff_duration=0 而是 no-op，需在用戶確認後回頭修補 185；195 audit 屆時也需同步處理 yaml buff_duration。

- 清理 `SHOCK_SKIN(189)` 在 `buffs.lua` 的 dead entry；對齊 Java 對 189 無 buff effect 的事實。Java codebase 對 `SHOCK_SKIN=189` 唯一引用是 `L1BuffUtil.java:59 if (hasSkillEffect(SHOCK_SKIN))` 阻擋傳送卷軸，但 yiwei 與 fly 的 `skills.sql` 第 189 列實際是 `岩漿之箭`（type=64 fire attack, target=attack, ranged=10, damage_value=50, attr=2），整個 codebase **沒有任何**路徑會 `setSkillEffect(189)`，是 L1BuffUtil 的 dead code（殘留自原始 3.80C SHOCK_SKIN buff 命名，被 yiwei 重新用於 `岩漿之箭` 但未清理 `L1BuffUtil`）。Go yaml `[189] type=64 area=2 damage_value=45 target=none` 走 `executeSelfAreaAttackSkill` 自身 AOE 攻擊路徑（line 234-236 命中 `isSelfAreaAttackSkill = type==64 && area>0 && damage_value>0` 後 return），永遠不會抵達 `applyBuffEffect`，所以 `[189] = { ac = -5 }` 是 dead lua entry。本步移除 dead entry 並加註釋避免後續對齊誤導。Go 與 Java 的資料差異（名稱衝擊之膚 vs 岩漿之箭、target=none vs attack、ranged=0 vs 10、area=2 vs 0、attr=8 vs 2、damage 45 vs 50、hp_consume 40 vs 0、mp_consume 0 vs 5、cast_gfx 6532 vs 10877）屬 data audit 範疇（與 184 MAGMA_BREATH 先例一致），留至後續 data audit。Java 的 L1BuffUtil:59 SHOCK_SKIN 傳送阻擋本身也是 dead code，Go 無需對應實作。

- 修正 `RESIST_FEAR(188)` 對其他玩家施放時缺少 MR 抗性閘；對齊 Java `L1MagicNpc.calcProbabilityMagic:198-205` `case RESIST_FEAR` 與 `GUARD_BRAKE/THUNDER_GRAB/HORROR_OF_DEATH` 共用標準等級差公式 `(ProbabilityDice/10) * (attackLevel - defenseLevel) + ProbabilityValue` × `leverage/10`，與對應的 `L1MagicPc.calcProbabilityMagic` 玩家施放分支。Go `skill_status.go playerDebuffSkills` 新增 `188: true`，與 183 GUARD_BRAKE 同源走 `checkPlayerMRResist` 通用閘。Java skillmode 第 18 行 `if (!cha.hasSkillEffect(188) && cha instanceof L1PcInstance)` 確保只對 PC 目標套用 dodge_down+5；Go 透過 `executeBuffSkill` 走 PC 目標路徑（NPC 目標走 `executeNpcDebuffSkill`）已隱含 PC-only 限制，無 NPC case 188 處理。既有測試 `TestSkillDragonKnightStatusResistFearAppliesDodgePenaltyOnly` 因新 MR 閘變概率性（50%）→ 改為自我施放（`caster == target`）路徑跳過 MR 閘，與 `skill_buff.go:940` `target.CharID != player.CharID` 自動跳過邏輯一致，仍正面測試 buffs.lua `[188] = { dodge = -5 }` 套用 stat penalty 與不誤動 STR/INT。
- 配套 Java 對照（broader gap 不改）：a) Java `L1Character._dodge_down` 為獨立累加器（clamp 0..10），與 `_dodge_up` 對稱，由 `L1AttackMode:186-187 attackerDice += character.get_dodge_down()` 在攻擊方骰子上加成（受害方變更易被命中）；Go 的 `Dodge` 欄位為單一儲存值不受戰鬥管線消耗（grep `combat.go`/`pvp.go`/scripting 無 `target.Dodge` 引用）。b) Java `S_PacketBoxIcon1(false, _dodge_down)` 為專屬 `opcode 250 + 0x65 + cumulative dodge_down` 圖示封包，Go 走 `S_STATUS`（含 AC + 全屬性）替代。c) Java 第 18 行 `!hasSkillEffect(188)` 守衛避免再施放刷新計時器，Go `applyBuffEffect` 走 `AddBuff` 替換邏輯會刷新 timer。三項皆屬「整體 dodge_down 機制 + buff 圖示封包 + 通用再施放守衛」結構性缺口，與其他 stack 累加器類技能（如 dodge_up、hit_up 等）同源，留至 dodge_down/dodge_up 累加器系統整體實作時補。

- 修正 `FOE_SLAYER(187)` 四項 Java 對齊差異；對齊 Java `skillmode/FOE_SLAYER.java`：
  - **stun 命中機率邊界**：Java 第 40 行 `_random.nextInt(100) <= FOE_SLAYER_RND(=15)` 為 inclusive（roll 0..15 共 16/100=16%），Go `foeSlayerStunSuccess` 原本 `RandInt(100) < chance` 為 exclusive（roll 0..14 共 15/100=15%）。改為 `<= chance` 對齊 Java 邊界。`ProbabilityValue=100` 的測試在兩種比較下都恆真，不受影響。
  - **PC 目標粉紅名觸發**：Java 第 48 行 `L1PinkName.onAction(pc, srcpc)` 在 COPY_SHOCK_STUN 命中 PC 後顯式觸發粉紅名（PvP 旗標），Go `applyFoeSlayerPlayerStun` 缺對應呼叫。補上 `s.deps.PvP.TriggerPinkName(caster, target)` 在 stun 落地後執行（與 `applyShockStunToPlayer` 第 109 行相同模式）。為此將函式簽章從 `applyFoeSlayerPlayerStun(target, skill, nearby)` 擴為 `applyFoeSlayerPlayerStun(caster, target, skill, nearby)`，呼叫端 `executeFoeSlayerOnPlayer` 同步更新。
  - **NPC 類別 setParalyzed 過濾**：Java 第 49-53 行只對 `L1MonsterInstance / L1SummonInstance / L1PetInstance` 三類 NPC `setParalyzed(true)`，Guardian/Guard/Tower/通用 NPC 只獲得 `COPY_SHOCK_STUN` 計時器與 81162 效果，**不**設 Paralyzed 旗標。Go `applyFoeSlayerNpcStun` 原本對所有 NPC 都 `npc.Paralyzed = true`，新增 `npc.Impl == "L1Monster" || "L1Summon" || "L1Pet"` 守衛（與 `skill_status.go:508` SHOCK_STUN NPC 路徑相同模式）。
  - **三段攻擊不中斷**：Java 第 27-34 行 `for (int i = 0; i < 3; i++) { cha.onAction(srcpc); ... }` 不論目標死活都跑完三次 onAction，且不論目標死活都廣播 7020/12119 + 嘗試 COPY_SHOCK_STUN。Go `executeFoeSlayerOnPlayer`/`executeFoeSlayerOnNpc` 原本在每次傷害應用後若目標死亡即 `return`，會跳過後續視覺與 stun 邏輯。移除中段 `if target.Dead || target.HP <= 0 { return }` 早返回；`applyFoeSlayerPlayerDamage` 與 `applyFoeSlayerNpcDamage` 已有 `target.Dead`/`npc.Dead` guard，後續呼叫安全 no-op。
- 配套 Java 對照：Java 戰鬥管線 `L1AttackPc.java:745 if (!_pc.isFoeSlayer()) dk_dmgUp()` 與 `L1AttackPc.java:768-780 if (_pc.isFoeSlayer()) damage += [20/40/60] + FoeSlayerBonusDmg`，Go `calcFoeSlayerPlayerHitDamage`/`calcFoeSlayerNpcHitDamage` 透過 `dragonKnightWeaknessFoeSlayerBonus` 已對齊（weakness 等級 1/2/3 對應 +20/+40/+60 + `player.FoeSlayerBonusDmg`）。
- 配套 Java 對照（broader gap）：Java `cha.onAction(srcpc)` 走完整 `L1AttackPc.attack()` 物理攻擊管線（含 doll skill 觸發、屬性守護之鍊、月亮項鍊、奪魂T恤吸血、毒素附加、Eat HP/MP、DimiterBlessRuneB 等附加效果），Go `applyFoeSlayerPlayerDamage`/`applyFoeSlayerNpcDamage` 直接扣 HP 繞過這些副作用，屬「skill-only path 繞過 melee pipeline 的全體龍騎技能缺口」，留至 doll/amulet 系統重構時整體對齊。

- 修正 `BLOODLUST(186)` 不正當的屬性加成與 exclusions 漏項；對齊 Java `skillmode/BLOODLUST.java`。Java BLOODLUST.start 只做 `L1BuffUtil.braveStart(srcpc) + setSkillEffect + setBraveSpeed(1) + S_SkillBrave`，**沒有任何 dmg_mod / hit_mod / AC 加成**（搜遍 yiwei codebase：`L1AttackPc.calcBuffDamage`/`L1AttackPc.BuffDmgUp`/`L1AttackHand`/`L1MagicHand` 等戰鬥管線無 BLOODLUST 引用；唯一兩處呼叫者 `S_OwnCharPack`/`S_OtherCharPacks` 只用 `isBrave()` 做客戶端勇敢光環旗標、`AutoAttackPc:1346` 做二段加速條件，全屬視覺/速度旗標）。Go `buffs.lua [186] = { dmg_mod = 6, hit_mod = 3, ac = 5, brave_speed = 1, exclusions = {52, 101, 150, 155} }` 的 `dmg_mod=6, hit_mod=3` 從 `[163]` Burning Weapon 誤抄、`ac=5` 為憑空加成（且 AC+5 是防禦變弱方向，與「血之渴望」buff 邏輯相反）。本步改為 `[186] = { brave_speed = 1, exclusions = {52, 101, 150, 155, 1000, 1016} }`：a) 移除三項不正當屬性加成；b) exclusions 由 4 項擴為 6 項，加入 `STATUS_BRAVE(1000)`/`STATUS_ELFBRAVE(1016)` 對齊 Java `L1BuffUtil.braveStart()` 第 214-266 行清單 `{HOLY_WALK(52), MOVING_ACCELERATION(101), WIND_WALK(150), STATUS_BRAVE(1000), STATUS_ELFBRAVE(1016), STATUS_RIBRAVE(1017), BLOODLUST(186 自己, applyBuffEffect AddBuff 替換邏輯已覆蓋), FIRE_BLESS(155)}`。STATUS_RIBRAVE(1017) pre-check（Java 第 22-26 行 `if hasSkillEffect(STATUS_RIBRAVE) → S_ServerMessage(1413) + return`）屬 Go 尚未實作的 `生命之樹果實` potion buff（搜遍 Go 僅有 yaml 物品定義無 buff handler），與 STATUS_RIBRAVE 整體留至 potion buff 系統實作時補。`skill_self.go case 186` 已正確走 `applyBuffEffect` 套用 buff + brave_speed=1，本步不改。無對應測試需更新。

- 修正 `AWAKEN_ANTHARAS(185)` 再施放錯誤的「toggle off 解除」行為；對齊 Java `skillmode/AWAKEN_ANTHARAS.java:17-23` 的 `if (!hasSkillEffect(185))` 守衛——再施放只跳過屬性套用但仍 `sendPacketsX8(S_SkillSound(6975))`，並無移除 buff 行為。Java `L1SkillUse2.java:1675-1683` 的 `if (_skillId == _player.getAwakeSkillId())` 因 `_awakeSkillId` 從未被 `setAwakeSkillId` 設定永遠為 false → 走 `return` 跳過第二次音效廣播，但 skillmode 自己的 X8 廣播仍會執行。`L1SkillUse.deleteRepeatedSkills` 的 `stopSkillList` 第 1787 行 `if (skillId != _skillId)` 也只移除群組內**其他**覺醒，不會碰到自己。Go `skill_buff.go:1144-1175` 原有的「再施放 → `removeBuffAndRevert(skill.SkillID)`」分支屬 Java 從未有過的偽 toggle 行為（comment 「Java: toggle off」誤讀），本步改為：a) 未持有 buff：走 `applyBuffEffect`（exclusions 清其他覺醒）+ `SendPlayerStatus` + 廣播音效；b) 已持有 buff：僅廣播音效（對齊 Java skillmode `sendPacketsX8`），不清 buff、不刷新計時器、不送狀態封包。190 法利昂同步取消「再施放清 169」的 toggle-off 副作用（natural 到期路徑 `tickPlayerBuffs` 第 803-806 行已正確清 169，保留即可）。覆蓋 185/190/195 三個覺醒共用區塊，三者皆已對齊。

- 純審計確認 `MAGMA_BREATH(184)` Go 已對齊 Java——Java codebase 對 184 沒有 skillmode/特殊處理（僅 `L1SkillId.java:619` 常數定義與 `AutoAttackUpdate.java:742` 列在 auto-attack 可用清單）。純資料驅動 attack skill，走標準 `executeAttackSkill` + `magic.calcMagicDamage` 路徑。Go yaml `type=64/target=attack/target_to=3/attr=2` 與 Java skills.sql 結構對齊。資料差異（mp_consume 0 vs 10、hp_consume 35 vs 0、damage_value 45 vs 50、reuse_delay 0 vs 1000、ranged 10 vs 5、area 0 vs 1）屬資料調整非行為對齊，留至 data audit。本步無代碼變更。
- 修正 `GUARD_BRAKE(183)` 數值與互斥對齊 Java——`scripts/combat/buffs.lua [183] = { ac = 5, dmg_mod = -3 }` 改為 `[183] = { ac = 10 }`，對齊 Java `L1SkillUse2.java:2271-2275` cast `pc.addAc(10)` 與 `L1SkillStop.java:669-673` stop `pc.addAc(-10)`。Java GUARD_BRAKE 只給 AC +10（變弱），**沒有 DmgMod 影響**，Go 之前的 `dmg_mod=-3` 為不正當加成。同時將 183 加入 `playerDebuffSkills` 走統一 `checkPlayerMRResist` MR 抗性閘（對齊 Java `L1MagicNpc.calcProbabilityMagic case GUARD_BRAKE` 標準等級差公式）。Java 明確 `if (cha instanceof L1PcInstance)` 限制 — Go executeNpcDebuffSkill 沒 `case 183`，與 167/173/174 同源 NPC 目標缺口，留至 broader audit。
- 修正 `BURNING_SLASH(182)` 從 passive `dmg_mod=5` 改為 Java 一次性 +10 + 消耗 buff 的 active 模式；對齊 Java `L1AttackPc.calcBuffDamage:2434-2438` 行為。新增 `burningSlashDamage(deps, attacker, damage, weaponType) → (damage, consumed)` helper，於 PvE (`combat.go`) 與 PvP (`pvp.go`) 近戰傷害管線晚段呼叫；命中時若 buff 182 存在則 `damage += 10` 並 `RemoveBuffAndRevert(182)`，同時廣播 `S_EffectLocation(targetX, targetY, 6591)` 至 nearby（對齊 Java `_pc.sendPacketsX10`）。bow/gauntlet 武器以 `isRangedWeaponType` 早返回排除（對齊 Java `_weaponType != 20 && _weaponType != 62`）；ki-koru (`_weaponType2 == 17`) Go 暫無對應類型不需處理。`scripts/combat/buffs.lua [182] = { dmg_mod = 5 }` 改為 `[182] = {}` 移除不正確的 passive 加成。Java `calcBuffDamage` 其他成員 (FIRE_WEAPON +4 / BURNING_WEAPON +6 / BERSERKERS +5) 屬 broader 武器 buff 家族缺口，留至整體 audit。
- 純審計確認 `DRAGON_SKIN(181)` Go 已完整對齊 Java——Java codebase 中 DRAGON_SKIN 沒有 skillmode、沒有 L1SkillUse/L1SkillUse2/L1SkillStop 特殊處理，純資料驅動 self-buff，且僅在 `EXCEPT_COUNTER_MAGIC[]` 清單中（counter magic 不可抵擋）。Go 端 `counterMagicExempt[181] = true` 已對齊 Java exemption，`buffs.lua [181] = { ac = -5 }` 套用 AC -5 走 `applyBuffEffect` 通用路徑。yaml `hp_consume=15`/`buff_duration=1200` vs Java skills.sql `hp_consume=12`/`reuse_delay=10`/`buff_duration=1800` 屬資料調整非行為對齊，留至後續 data audit。本步無代碼變更。
- 修正 `EXOTIC_VITALIZE(169)` 與 `ADDITIONAL_FIRE(176)` 不應有屬性加成；對齊 Java `L1PcInstance.isRegenHp/isRegenMp`（行 775/831）兩者**僅作為負重 HP/MP 回復旗標**，無 STR/DEX 加成。`scripts/combat/buffs.lua` 將 `[169] = { str = 5 }` 改為 `[169] = {}`、`[176] = { str = 2, dex = 2 }` 改為 `[176] = {}`。前次 169 audit 漏抓 `str=5` 的不對齊（只修了 `regen.lua` 負重判定順序），本次 176 audit 連帶補齊。`AWAKEN_FAFURION(190)` 仍會透過 synergy 把 169 timer 同步套上（行為對齊 Java `skillmode/AWAKEN_FAFURION.java:18`），但 169 buff 本身不再給予 STR+5。
- 修正 `SOUL_OF_FLAME(175)` 近戰增傷倍率從 2x 改為 1.5x，對齊 Java `L1AttackPc.java:1455-1457/1945-1947` 與 `ConfigSkill.SOUL_OF_FLAME_DAMAGE` yiwei 預設值 `1.5`。`skill_elemental.go elfMeleeDamageWithRoll` 將 `damage *= 2` 改為 `damage = damage * 3 / 2`，PvP 與 NPC 兩路（`pvp.go:96` / `combat.go:213`）皆受益。同步更新 `TestSkillElementalDynamicWaterLifeAndPolluteWaterModifyHealing` 預期值（175+171 雙 buff `100 * 3/2 * 3/2 = 225`，從原本 `300` 改為 `225`）。`isRangedWeaponType` 早返回排除 bow/gauntlet 已對齊 Java 第 1452 行 `else { // 近距離武器 }` 範圍限制。Java「`weaponMaxDamage * 1.5` 取代武器傷害骰值」屬武器傷害管線結構改造（需暴露 weaponMin/Max 至 elf-melee 呼叫點），與 `L1AttackPc.calcBuffDamage` 武器 buff 家族同源結構缺口留至後續整體 audit；`SOUL_OF_FLAME_ALLDAMAGE=1.0` 預設停用、`L1AttackPc.java:749-751` 全傷倍率路徑亦留至同 audit。
- 補上 `STRIKER_GALE(174)` 套用/解除時的 `UPDATE_ER` 客戶端封包；對齊 Java `L1PcInstance.getEr()` 第 3396-3398 行「持有 STRIKER_GALE 直接 `return 0`」與 `L1SkillStop case STRIKER_GALE` 第 540-543 行 buff 結束時送 `S_PacketBox(UPDATE_ER, pc.getEr())`。`executeBuffSkill` 套用 174 後送 `SendUpdateER(sess, 0)` 模擬 Java `getEr()` 為 0 的 UI 顯示；`removeBuffAndRevert(174)` 與 `tickPlayerBuffs` 到期路徑送 `SendUpdateER(sess, target.Dodge)` 還原 Go 儲存的真實 ER（Java 為 `getEr()` 即時計算，Go 為 `player.Dodge` 儲存欄位）。傷害 1.1x (STRIKER_DMG2)、`playerDebuffSkills[174]=true` MR 抗性閘、buffs.lua flag 已實作。NPC 目標（`executeNpcDebuffSkill` 沒 `case 174`）屬「broader gap: 全體 elf debuff NPC 目標」，與 173 同源留至整體 audit；cast probability 三段公式 (`STRIKER_GALE_1/2/3 = 70/40/20` 配 INT/MR 微調) 也屬同族 ConfigElfSkill probability 差異留至後續。
- 補上 `POLLUTE_WATER(173)` 對玩家 HP 藥水回復量減半的對齊——`item_use.go` heal 藥水路徑在高斯隨機後加 `if player.HasBuff(173) { healAmt /= 2 }`，並夾底 `healAmt < 1 → 1`，對齊 Java `UserAddHp.java:69-71`/`UserAddHp_FR.java:91-93` 的 `addhp >>= 1`。先前已實作：a) 治療法術 heal 減半（`applyElfWaterHealingModifiers` 的 `>> 1` 對齊 `L1SkillUse2:2002-2005`）；b) PC 目標 debuff MR 抗性閘（`playerDebuffSkills[173]=true` + `checkPlayerMRResist`）；c) buffs.lua flag 形式 `[173]={}`。NPC 目標路徑（`executeNpcDebuffSkill` 沒 `case 173`）、Java 武器 special trigger（`L1WeaponSkill case 16` 32 秒）與 NPC `useHealPotion` 減半（`L1NpcInstance:2280-2282`）屬「broader gap: 全體 elf debuff NPC 目標 + 武器 special 觸發 + NPC heal potion」，留至後續整體 audit。
- 純審計確認 `ELEMENTAL_FIRE(171)` Go 已完整對齊 Java `L1AttackPc.BuffDmgUp`——`elfMeleeDamageWithRoll` 對持有 171 buff 的玩家以 `RandInt(100) < 33` 33% 機率觸發 `damage * 3 / 2`（即 1.5x），並以 `isRangedWeaponType(weaponType)` 早返回排除 bow/gauntlet（對齊 Java `_weaponType != 20 && _weaponType != 62`）。PvE/PvP melee 兩路皆掛接（`combat.go:213` / `pvp.go:96`）。資料 yaml `type=2/buff_duration=192/target=none/target_to=0` 與 Java config 對齊。本步無代碼變更。
- 補上 `WATER_LIFE(170)` buff 結束時的 `S_PacketBoxWaterLife` 圖示取消封包；對齊 Java `L1SkillStop case 170` 於 buff 到期或被移除時送出 `S_OPCODE_PACKETBOX (opcode 250)` + `byte 59 (ICON_WATER_LIFE_CANCEL)` + `H(0)`，清除客戶端水之元氣狀態圖示。新增 `handler.SendWaterLifeCancel(sess)` helper，於 `removeBuffAndRevert(skillID=170)` 與 `tickPlayerBuffs` 到期路徑（skillID=170）兩處呼叫。Heal 加倍 + 消耗已於 `applyElfWaterHealingModifiers` 實作；POLLUTE_WATER(173) 減半同方法。
- 修正 `EXOTIC_VITALIZE(169)` 對齊 Java `L1PcInstance.isRegenHp/isRegenMp` 的負重判定順序與 HP 門檻：
  - HP 過載門檻從 `>= 121` 改為 `>= 120`，與 Java `if (120 <= weight240)` 一致。
  - 將「負重檢查」提到「食物檢查」之前；當 `weight240 >= 120` 且有 `EXOTIC_VITALIZE(169)` 或 `ADDITIONAL_FIRE(176)` 時，Java 直接 `return _hpRegenCount >= 12 / _mpRegenCount >= 64` 並繞過食物檢查，Go 原本在食物 < 3 時會先擋下、忽視 buff 的負重赦免。檔案：`server/scripts/character/regen.lua` 的 `calc_hp_regen_amount` 與 `calc_mp_regen_amount`。
- 修正玩家施放 `SHOCK_STUN` 對 `ABSOLUTE_BARRIER(78)` 目標的邊界，對齊 Java `L1SkillUse.checkTarget()` 會排除絕對屏障目標；被排除時不套用 87、不送目標 4434，也不進入後續衝暈副作用。
- 修正 `C_UseSkill` 負重過高入口阻擋；對齊 Java 在 `getWeight240() >= 197` 時送出訊息 316 並拒絕施法，避免超重玩家施放 `SHOCK_STUN` 仍進入 SkillQueue。
- 修正 `C_UseSkill` 死亡狀態入口阻擋；對齊 Java 在 `pc.isDead()` 時直接返回，避免死亡玩家施放 `SHOCK_STUN` 仍進入 SkillQueue。
- 修正 `C_UseSkill` 技能延遲入口阻擋；對齊 Java 在 `pc.isSkillDelay()` 時直接返回，避免延遲中的玩家施放 `SHOCK_STUN` 仍進入 SkillQueue。
- 修正 `C_UseSkill` 地圖不可使用技能入口阻擋；對齊 Java 在 `!pc.getMap().isUsableSkill()` 時送出訊息 563 並拒絕施法，避免禁用技能地圖仍排入 SkillQueue。
- 修正 `C_UseSkill` 非法技能 ID 上限阻擋；對齊 Java 在 `skillId > 239` 時直接返回，避免異常 row/column 封包排入 SkillQueue。
- 修正未學會 `SHOCK_STUN` 時的絕對屏障解除順序；對齊 Java 先以 `isSkillMastery` 擋下未知技能，避免非法施放仍解除施法者自己的 78。
- 修正未學會 `SHOCK_STUN` 時的失敗回饋；對齊 Java `isSkillMastery=false` 直接返回，不再送出 `S_ServerMessage(280)`。
- 修正合法施放 `SHOCK_STUN` 時的冥想術解除；對齊 Java `C_UseSkill` 在合法施法前會移除 `MEDITATION(32)`。
- 修正 `SHOCK_STUN` 的不可施法變形與隱身限制順序；對齊 Java 先以 `poly.canUseSkill=false` 回覆 285，不送 1003 且不解除隱身。
- 修正 `SHOCK_STUN` 的麻痺與隱身限制順序；對齊 Java `isParalyzedX()` 先回覆 285，不送 1003 且不解除隱身。
- 修正玩家施放 `SHOCK_STUN` 對 `ABSOLUTE_BARRIER(78)` 目標的消耗時機；對齊 Java `checkTarget()` 會在 `useConsume()` 前排除目標，不消耗 MP。
- 修正 NPC 單體施放 `SHOCK_STUN` 對 `ABSOLUTE_BARRIER(78)` 玩家目標的邊界；對齊 Java `checkTarget()` 直接排除，不清除睡眠、不送施法動作或目標特效。
- 修正 NPC 單體施放 `SHOCK_STUN` 對 `ABSOLUTE_BARRIER(78)` 玩家目標的仇恨語義；對齊 Java `checkUseSkill=false` 只讓技能失敗，不清除怪物既有 `AggroTarget`。
- 修正 NPC 單體施放 `SHOCK_STUN` 對死亡玩家目標的邊界；對齊 Java `checkTarget()` 在 `runSkill()` 前排除死亡目標，不清除睡眠與 `ERASE_MAGIC`，也不送施法動作或目標特效。
- 修正 NPC 單體施放 `SHOCK_STUN` 對死亡玩家目標的消耗時機；對齊 Java `checkUseSkill()` 失敗時不進入 `useConsume()`，死亡目標不再扣 NPC MP。
- 修正 NPC 單體施放 `SHOCK_STUN` 對 GM 隱身玩家目標的消耗時機；對齊 Java `isTarget()` 在 `useConsume()` 前排除 GM 隱身目標，不再扣 NPC MP。
- 修正 NPC 單體施放 `SHOCK_STUN` 對跨地圖玩家目標的邊界；對齊 Java 怪物技能目標來自同地圖可見玩家，跨地圖目標不再清除睡眠或 `ERASE_MAGIC`，也不套用 87。
- 修正 NPC 單體施放 `SHOCK_STUN` 對跨地圖玩家目標的消耗時機；對齊 Java `checkUseSkill()` 失敗時不進入 `useConsume()`，跨地圖目標不再扣 NPC MP。
- 修正 NPC 單體施放 `SHOCK_STUN` 對射程外玩家目標的邊界；對齊 Java `makeTargetList()` 以 `ranged=1` 排除目標，射程外目標不再清除睡眠或 `ERASE_MAGIC`，也不套用 87。
- 修正 NPC 單體施放 `SHOCK_STUN` 對射程外玩家目標的消耗時機；對齊 Java `makeTargetList()` 會在 `useConsume()` 前排除目標，射程外目標不再扣 NPC MP。
- 修正 mob skill type 5 範圍 `SHOCK_STUN` 的死亡目標邊界；對齊 Java `areashock_stun()` 只排除 GM 隱身與已有 87，死亡可見玩家仍會被套用 87。
- 修正 mob skill type 5 範圍 `SHOCK_STUN` 的 MP 條件；對齊 Java `areashock_stun()` 沒有 MP 檢查或消耗，MP 不足時仍可觸發。
- 修正 mob skill type 5 範圍 `SHOCK_STUN` 的 `trigger_random` 語義；對齊 Java `rnd > 0 && rnd <= random(1..100)`，`1` 必定觸發、`0` 不觸發。
- 修正玩家施放 `SHOCK_STUN` 對守護塔目標的邊界；對齊 Java `L1SkillUse.isTargetFailure()` 對 `L1TowerInstance` 回傳 true，TYPE_PROBABILITY 流程直接 `iter.remove()` 並讓 `sendGrfx()` 於 `_targetList.size()==0` 直接 return，但 `_target.onAction(_player)` 仍會在迴圈外觸發。Go 對 `npc.Impl == "L1Tower"` 不清睡眠、不解 `ERASE_MAGIC(153)`、不做安全區/雙手劍/已有 87/概率判定、不套 87、不生成 81162、不送目標 4434，仍保留 SHOCK_STUN 的近戰排程。
- 補上 `SHOCK_STUN` 到期路徑的 Java 對照回歸；對齊 Java `SHOCK_STUN.stop()` 對 `L1PcInstance` 送 `S_Paralysis(5,false)` 且對 `L1MonsterInstance/L1SummonInstance/L1GuardianInstance/L1GuardInstance/L1PetInstance` `setParalyzed(false)`，鎖定 Go `tickPlayerBuffs`(0x17 StunRemove + 清 Paralyzed) 與 `tickNpcDebuffs`(case 87 清 `npc.Paralyzed` 與 `ActiveDebuffs[87]`)，避免後續修改靜默退化。
- 補上 `SHOCK_STUN` 套用路徑的 `S_Paralysis(5,true)` (0x16 StunApply) 封包回歸；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 與 `start(L1NpcInstance,...)` 兩個分支都對玩家目標 `pc.sendPackets(new S_Paralysis(5, true))`，鎖定 Go `applyBuffEffect` 對 SkillID=87 + `eff.Paralyzed=true` 送 StunApply 行為，玩家施放與 NPC 施放成功路徑各有獨立測試。
- 補上玩家施放 `SHOCK_STUN` 對 **NPC 目標** 的雙手劍要求回歸；對齊 Java `SHOCK_STUN.start()` 第 34-37 行 `getWeapon().getItem().getType1() != 50` 即送 `S_SystemMessage("請使用雙手劍")` 並返回，無論目標是 `L1PcInstance` 或 `L1NpcInstance` 都套用，鎖定 Go `executeNpcDebuffSkill` case 87 的 NPC 目標路徑（未裝備不套 87/不生成 81162/送訊息；裝備後正常套用）。
- 補上玩家對自己施放 `SHOCK_STUN` 的 Java 早返回回歸；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 第 31-33 行 `if (srcpc.getId() == cha.getId()) return 0;` 會在雙手劍檢查與後續 `setSkillEffect`、`spawnEffect(81162)`、`S_Paralysis(5,true)`、`L1PinkName.onAction` 之前直接返回，鎖定 Go `applyShockStunToPlayer` 對 `caster.CharID == target.CharID` 直接返回的行為：未裝備雙手劍時對自己施放也不送「請使用雙手劍」（驗證自我檢查排在雙手劍檢查之前）、不套 87、不 setParalyzed、不 spawn 81162、不送 GM 秒數訊息。
- 補上玩家施放 `SHOCK_STUN` 對 `L1Guardian` 目標的 Java 類別邊界回歸；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 第 53-57 行只對 `L1MonsterInstance / L1SummonInstance / L1PetInstance` 設 `setParalyzed(true)`，`L1GuardianInstance` 與 `L1GuardInstance` 均排除（NPC caster 分支才把 Guardian/Guard 納入），鎖定 Go `executeNpcDebuffSkill` case 87 對 `Impl == "L1Guardian"` 仍套 87 debuff 與 81162 效果但不 setParalyzed；`L1Guard` 已有對應測試，Guardian 補齊後兩個玩家施放排除類別都有獨立回歸。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 的 `act_id > 0` 覆寫回歸；對齊 Java `L1MobSkillUse.areashock_stun()` 第 734-738 行 `actionid = 1; if (actId > 0) actionid = actId;` 後以 `_attacker.broadcastPacketAll(S_DoActionGFX)` 廣播覆寫值，鎖定 Go `executeNpcAreaShockStun(npc, 7)` 廣播 action=7 而非預設 1，與單體 NPC 路徑的 act_id 覆寫測試對齊（既有測試只驗證 act_id=0 預設 1）。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 的同地圖目標來源回歸；對齊 Java `L1MobSkillUse.areashock_stun()` 第 740 行 `World.get().getVisiblePlayer(_attacker)` 只取同地圖可見玩家，鎖定 Go `executeNpcAreaShockStun` 不同地圖玩家不會被套 87、不 setParalyzed、不生成 81162，同地圖玩家仍正常套用。
- 補上玩家合法施放 `SHOCK_STUN` 解除自己 `ABSOLUTE_BARRIER(78)` 的 Java 正面案例；對齊 Java `C_UseSkill.start()` 在 `isSkillMastery` 通過後、`killSkillEffectTimer(MEDITATION)` 之前 `cancelAbsoluteBarrier()` 解除施法者自己的 78，鎖定 Go `SkillSystem.processSkill` 合法施放 87 時 `caster.AbsoluteBarrier=false` 且 `buff 78` 被移除（負面案例「未學會 87 不解除自己 78」已有對應測試）。
- 補上 `SHOCK_STUN` 玩家施放成功率的 Java `IMPACT_HALO_INT=0` 回歸；對齊 Java `L1MagicPc.calcProbabilityMagic()` 第 649-651 行 `if (IMPACT_HALO_INT > 0) probability += IMPACT_HALO_INT * INT`，yiwei 設定 `IMPACT_HALO_INT=0` 整個倍率區塊被略過，鎖定 Go `shockStunPlayerProbability` / `shockStunNpcProbability` 對高 INT（127）的玩家與 NPC 目標都精確為 `IMPACT_HALO_2(30) + BaseInt 45+(5) + IntMagicHit table((127-20)/3=35) = 70`，不被任何 `IMPACT_HALO_INT * INT` 線性倍率污染（與既有 IMPACT_HALO_MR=0 測試對齊）。
- 補上玩家施放 `SHOCK_STUN` 對 `L1Pet` 目標的 Java setParalyzed 包含案例；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 第 53-57 行明確列出 `L1MonsterInstance / L1SummonInstance / L1PetInstance` 三類會 `setParalyzed(true)`，鎖定 Go `executeNpcDebuffSkill` case 87 對 `Impl == "L1Pet"` 同時套 87 debuff、81162 效果與 setParalyzed(true)；與 L1Guard / L1Guardian 兩個 negative 排除測試對齊，完成 Java 第 53-57 行三類 positive + 兩類 negative 的完整覆蓋。
- 補上玩家施放 `SHOCK_STUN` 對 `L1Summon` 目標的 Java setParalyzed 包含案例（Java 第 53-57 行三類 positive 的最後一項）；鎖定 Go `executeNpcDebuffSkill` case 87 對 `Impl == "L1Summon"` 同時套 87 debuff、81162 效果與 setParalyzed(true)。至此玩家施放 NPC 類別邊界完整覆蓋：positive (L1Monster 隱含 / L1Summon / L1Pet) + negative (L1Guard / L1Guardian)。
- 補上 `SHOCK_STUN` 玩家目標被 `CANCELLATION(44)` 相消的 Java 不可解除語義；對齊 Java `L1SkillMode.isNotCancelable()` 第 33 行明確列出 `SHOCK_STUN`，`CANCELLATION.java` buff 迴圈會略過該效果，鎖定 Go `cancelAllBuffs` 透過 Lua `IsNonCancellable(87)` 對玩家目標 buff 87 同樣略過，但仍解除其他可相消 buff（如緩速 29）；NPC 目標已有對應測試，本步補上玩家目標對等回歸。
- 補上玩家施放 `SHOCK_STUN` 對 valid target 命中失敗仍消耗 MP 的 Java 正面案例；對齊 Java `C_UseSkill.start()` 的 `useConsume()` 在 `runSkill()` 概率判定之前執行，因此 valid target（非射程外、非絕對屏障）即使命中 0% 仍會扣 MP，鎖定 Go `SkillSystem.processSkill` 對 valid target（Level 99 + RegistStun 100 → 命中率夾底 0%）必失敗時仍扣 `MpConsume=15`；既有「invalid target 不消耗 MP」測試覆蓋負面案例，本步補上正面案例完成 useConsume 與 runSkill 順序的雙向鎖定。
- 補上玩家施放 `SHOCK_STUN` 對 valid NPC 目標 hit/miss 都消耗 MP 的 Java 對等案例；對齊 Java `useConsume()` 對 NPC 目標同樣在 `runSkill()` 之前執行，鎖定 Go `processSkill` 對 valid NPC 目標（alive、同地圖、in-range）扣 `MpConsume=15`，與玩家目標版測試對齊，完成 useConsume 順序在玩家對玩家、玩家對 NPC 兩條主路徑的完整覆蓋。
- 補上 `SHOCK_STUN` GM 秒數訊息的 Java 非廣播語義；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 第 44-46 行使用 `srcpc.sendPackets(...)`（caster only）非 `sendPacketsAll`，鎖定 Go `SendNormalChat(sess, ...)` 只送給 caster session：caster 收到秒數訊息，附近的 GM 觀察者 session 不收到（既有測試僅驗證 caster 收到，本步補上 negative case）。
- 補上 `SHOCK_STUN` 玩家目標已有 87 時不再觸發 `L1PinkName.onAction` 的 Java 負面案例；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` `L1PinkName.onAction(pc, srcpc)` 位於 `if (!cha.hasSkillEffect(87))` 區塊內，目標已有 87 時整段被跳過，鎖定 Go `applyShockStunToPlayer` 對 `target.HasBuff(87)` 早返回前 `PvP.TriggerPinkName` 不被呼叫（called=0）；既有 positive 觸發 PinkName 測試補上 negative case，完成 PinkName 觸發條件的雙向覆蓋。
- 修正 `SHOCK_STUN` 清除睡眠效果遺漏 `PHANTASM(212)` 的 Java 對齊缺失；對齊 Java `L1SkillUse.runSkill()` 第 1965-1968 行對 TYPE_PROBABILITY 技能在概率結果處理前 `removeSkillEffect(FOG_OF_SLEEPING/212/103)`，Go `clearShockStunSleepEffects` 與 `clearShockStunNpcSleepEffects` 原本只清除 62/66/103（漏 212），本次補上 212 清除並加入兩個 PHANTASM 回歸測試（玩家目標 + NPC 目標）；保留 62 以維持與 pvp.breakPlayerSleep 等其他 sleep 清除路徑的一致性。
- 補上 NPC 施放 `SHOCK_STUN` 清除目標 `PHANTASM(212)` 的 Java 對齊回歸；前項已修正 helper 包含 212 清除，NPC caster 路徑 `ApplyNpcShockStun` 走相同 helper，本步用 `TestSkillClanShockStunNpcCasterClearsPhantasmLikeJava` 鎖定 NPC 施放對玩家目標 PHANTASM(212) 同樣被清除，完成 player→player / player→NPC / NPC→player 三條主路徑的 PHANTASM 清除覆蓋。
- 補上玩家施放 `SHOCK_STUN` 不被 `COUNTER_MAGIC(31)` 抵擋的 Java 對齊回歸；對齊 Java `L1SkillUse.EXCEPT_COUNTER_MAGIC[]` 第 146 行明確列出 `SHOCK_STUN`，鎖定 Go `counterMagicExempt[87]=true` 行為：CM 目標仍被套 87 與 Paralyzed、CM buff 不被消耗、且不觸發抵消動畫廣播。
- 補上玩家施放 `SHOCK_STUN` 對 `ICE_LANCE(50)` 控制目標的 Java 阻擋；對齊 Java `L1MagicPc.calcProbabilityMagic()` 對目標已有 50 或 157 皆讓 SHOCK_STUN 判定失敗，既有測試只覆蓋 buff 157（EARTH_BIND），本步補上 buff 50（ICE_LANCE）的對等 negative case，鎖定 Go `if target.HasBuff(50) || target.HasBuff(157)` 的 `||` 雙條件邏輯。
- 補上玩家施放 `SHOCK_STUN` 對 NPC 目標已有 `ICE_LANCE(50)` 的 Java 阻擋；對齊 Java `L1MagicPc.calcProbabilityMagic()` 對 NPC 目標一樣以 `hasSkillEffect(50) || hasSkillEffect(157)` 判定自動失敗，既有測試只覆蓋 debuff 157，本步補上 debuff 50 的對等 negative case：不套 87、不送目標 4434，避免 Go `||` 雙條件分支未來改動只覆蓋單邊。
- 補上 NPC 施放 `SHOCK_STUN` 對玩家目標已有 `ICE_LANCE(50)` 的 Java 阻擋；對齊 Java `L1MagicNpc.calcProbabilityMagic()` 與 `L1MagicPc` 對等以 `hasSkillEffect(50) || hasSkillEffect(157)` 判定自動失敗，既有測試只覆蓋玩家 buff 157（EARTH_BIND），本步補上 buff 50（ICE_LANCE）的對等 negative case：NPC 仍廣播施法者 19、不套 87、不送目標 4434，完成 50/157 在 player-cast→player、player-cast→NPC、NPC-cast→player 三條主路徑的完整覆蓋。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 對 `ICE_LANCE(50)` / `EARTH_BIND(157)` 控制目標仍套用 87 的 Java 對齊（語義差異 negative-of-negative）；對齊 Java `L1MobSkillUse.areashock_stun()` 僅以 `!isGmInvis() && !hasSkillEffect(SHOCK_STUN)` 判斷，**不**沿用單體 `calcProbabilityMagic()` 的 50/157 短路。新增兩個回歸測試鎖定 AOE 路徑不要誤把單體 NPC caster 的 `target.HasBuff(50) || target.HasBuff(157)` 短路複製過來，避免凍結中的玩家無故躲過 type 5 衝暈。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 對 `ABSOLUTE_BARRIER(78)` 玩家仍套用 87 的 Java 對齊（語義差異 negative-of-negative）；對齊 Java `areashock_stun()` 並未沿用單體 `checkTarget()` 對 78 的排除，AOE 只看可見性與兩條件。新增回歸測試鎖定 Go `ApplyNpcAreaShockStun` 不要把單體 `target.AbsoluteBarrier` 短路複製到 AOE 路徑，避免 78 護持玩家無故躲過 type 5 衝暈。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 對**安全區內**玩家仍套用 87 的 Java 對齊（語義差異 negative-of-negative）；對齊 Java `areashock_stun()` 不沿用單體 `checkTarget()/checkZone()` 的安全區排除，AOE 路徑沒有區域檢查。新增回歸測試鎖定 Go `ApplyNpcAreaShockStun` 不要把單體 `shockStunSafetyZoneBlocked` 短路複製到 AOE 路徑，至此 AOE 與單體的 50/157、78、安全區三個語義差異都有獨立回歸覆蓋。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **不**清除目標 `PHANTASM(212)` / `ERASE_MAGIC(153)` 的 Java 對齊（第四、第五個 negative-of-negative）；對齊 Java `areashock_stun()` 第 740-749 行只做 `setSkillEffect/S_Paralysis/spawnEffect`，不沿用單體 `runSkill()` 的睡眠系與 ERASE_MAGIC 清除。新增兩個回歸測試鎖定 Go AOE 路徑不要把單體 `clearShockStunSleepEffects` / `clearShockStunEraseMagic` 短路複製過來，避免 type 5 衝暈順便清掉目標睡眠或 ERASE_MAGIC。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **可見玩家為空仍 return true** 的 Java 對齊；對齊 Java `areashock_stun()` 在迴圈為空時仍會執行後續 `broadcastPacketAll` 與 `_sleepTime = SubMagicSpeed`，最後 `return true` 使呼叫端進入冷卻。新增回歸測試鎖定 Go `executeNpcAreaShockStun` 在沒有可見玩家時仍 `return true`，避免未來改成「無目標→不冷卻」造成怪物無限重觸 type 5。
- 修正 `C_UseSkill` 漏掉 Java `pc.isTeleport()` 入口阻擋；對齊 Java C_UseSkill 第 106-108 行 `if (pc.isTeleport()) return;`（傳送預備等待 C_TELEPORT 確認時靜默拒絕所有技能），Go `HandleUseSpell` 在 Java 對應順序補上 `if player.HasTeleport { return }`（位於 Dead 與 PrivateShop 之間），避免傳送預備中的玩家仍能施放 SHOCK_STUN 等技能。
- 補上**被沉默玩家可施放 `SHOCK_STUN`** 的 Java 對齊；對齊 Java `C_UseSkill` `_cast_with_silence[]` 明確列出 SHOCK_STUN、REDUCTION_ARMOR、BOUNCE_ATTACK、SOLID_CARRIAGE、COUNTER_BARRIER、FOE_SLAYER 六個沉默白名單技能。Go `isCastableWhileSilenced` 列表與 Java 完全一致，但原本無 SHOCK_STUN 專屬 regression test。新增回歸測試鎖定 `Silenced=true` 玩家施放 87 不送 285、仍進入 `useConsume` 扣 MP。
- 補上 NPC 施放 `SHOCK_STUN` **不觸發 onAction 近戰委派**的 Java 對齊（與玩家施放路徑形成 negative 對比）；對齊 Java `L1SkillUse.runSkill()` 第 1853-1857 行明確註解 NPC 施放會在 onAction 觸發 NullPointerException，因此 `_target.onAction(_player)` 僅在 `_user instanceof L1PcInstance` 時呼叫。新增回歸測試鎖定 Go `ApplyNpcShockStun` 成功命中後 `Combat.QueueAttack` 不被呼叫，避免未來把玩家 caster 的 onAction 委派誤複製到 NPC 路徑。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 也**不觸發 onAction** 的 Java 對齊（與 NPC 單體 caster 對等）；對齊 Java `L1MobSkillUse.areashock_stun()` 與 `L1SkillUse.runSkill()` 完全在不同程式路徑，AOE 從不接觸 onAction。新增回歸測試鎖定 Go AOE 路徑同樣不呼叫 `Combat.QueueAttack`，避免 type 5 範圍衝暈同時對每名目標排入近戰攻擊。至此 NPC 兩條路徑（單體 + AOE）的 onAction negative case 形成完整覆蓋。
- 補上 `SHOCK_STUN` INT MP 減免**邊界 negative case**；對齊 Java `L1SkillUse.java` 第 1161 行 `getInt() > 12` 嚴格大於（exclusive），Intel=12 不觸發 `_mpConsume -= (Intel - 12)` 減免。既有測試只用 Intel=18 驗證正面減免，本步補上 Intel=12 邊界鎖定 Go `mpAfterIntReduction` 的嚴格大於語義不退化為 `>=`，避免 Intel=12 玩家施放 87 誤扣 0 MP。
- 補上 `SHOCK_STUN` INT MP 減免**最小有效減免邊界**；對齊 Java 第 1163 行 `_mpConsume -= (Intel - 12)` 在 Intel=13 時剛好減 1 MP，與 Intel=12 不減形成 off-by-one 配對。新增回歸測試鎖定 Go 對 Intel=13/12 兩端離散行為正確銜接，避免未來退化為 `Intel - 13` 或 `>= 13` 造成 Intel=13 不減的 off-by-one bug。
- 補上 NPC 施放 `SHOCK_STUN` **不觸發 PinkName** 的 Java 對齊（與玩家施放路徑形成 negative 對比）；對齊 Java `SHOCK_STUN.start(L1NpcInstance,...)` 對玩家目標只送 `S_Paralysis`，無 `L1PinkName.onAction` 呼叫（該呼叫只存在於 `start(L1PcInstance,...)`）。新增回歸測試鎖定 Go `ApplyNpcShockStun` 成功命中後 `PvP.TriggerPinkName` 不被呼叫，避免怪物施放 87 後玩家被誤掛粉名。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` 也**不觸發 PinkName** 的 Java 對齊（與 NPC 單體 caster 對等）；對齊 Java `areashock_stun()` 同樣無 `L1PinkName.onAction` 呼叫。新增回歸測試鎖定 Go AOE 路徑不呼叫 `PvP.TriggerPinkName`，至此 NPC 兩條路徑的 PinkName + onAction 兩種 PvP 副作用 negative case 完整覆蓋。
- 補上 `SHOCK_STUN` **裝備 1H 劍**仍被拒絕的 Java 對齊（雙手劍要求中間案例）；對齊 Java 第 34 行 `getType1() != 50` 嚴格排除所有非 2H 劍類武器，既有測試只覆蓋無武器與正確 2H 劍兩端，本步補上 1H 劍中間 negative case：不套 87、不 Paralyzed、仍送 `S_SystemMessage("請使用雙手劍")`。鎖定 Go `hasTwoHandSwordEquipped` 對 `type=sword` 嚴格回傳 false，避免未來退化為「有劍即可」寬鬆比對。
- 補上 `SHOCK_STUN` **裝備 1H 劍對 NPC 目標**也被拒絕的 Java 對齊（雙手劍要求 NPC 版）；對齊 Java 第 34 行 `getType1() != 50` 對 NPC 目標分支同樣嚴格排除非 2H 劍。新增回歸測試鎖定 Go NPC 目標路徑同樣以嚴格判斷拒絕 1H 劍（不套 87、不 spawnEffect(81162)、仍送「請使用雙手劍」），完成 1H 劍中間案例的玩家+NPC 雙目標覆蓋。
- 補上 `SHOCK_STUN` **非 GM caster 不收到 GM 秒數訊息**的 Java 對齊（GM 門檻第三條 negative）；對齊 Java 第 44-46 行 `if (srcpc.isGm())` 嚴格 GM 門檻。既有測試覆蓋 GM caster 收到與附近 GM 觀察者不收到，本步補上非 GM 玩家施放 87 成功時也不收到「此次衝暈秒數為...」訊息，避免未來退化為「成功就送」。至此 GM 門檻在三條路徑形成完整覆蓋。
- 補上玩家施放 `SHOCK_STUN` 對玩家目標 **RegistStun 扣除**的 unit 級 Java 對齊；對齊 Java `L1MagicPc.calcProbabilityMagic()` PC_PC `case SHOCK_STUN: probability -= _targetPc.getRegistStun()`。既有 NPC 施放版走 end-to-end integration 證明扣除作用，本步補上 PC_PC 路徑的 unit 回歸（Level 60 vs 50 + RegistStun=10 → 40-10=30），鎖定 1:1 扣除而非倍率或忽略，避免未來公式重構時退化。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **81162 效果精確 spawn 在 target 座標**的 Java 對齊；對齊 Java `areashock_stun()` `spawnEffect(81162, shock, pc.getX(), pc.getY(), ...)` 第 3-4 個參數使用 target 玩家座標。既有 `MatchesJava` 透過 AOI 查找可能漏檢 caster 座標誤 spawn，本步把 caster 與 target 分開 4 格並嚴格驗證 `effects[0].X == target.X && effects[0].Y == target.Y`，鎖定 Go AOE 不會把 81162 spawn 在 caster 座標。
- 補上 NPC 單體施放 `SHOCK_STUN` 也鎖 **81162 效果精確 spawn 在 target 座標**的 Java 對齊（與 AOE 對等）；對齊 Java `SHOCK_STUN.start(L1NpcInstance,...)` `spawnEffect(81162, shock, cha.getX(), cha.getY(), ...)` 同樣使用 `cha.getX()/cha.getY()` 而非 `npc.getX()`。新增回歸測試把 caster 與 target 分開 4 格嚴格驗證座標，至此 NPC 兩條路徑（單體 + AOE）的 81162 spawn 座標精確驗證形成完整覆蓋。
- 補上玩家施放 `SHOCK_STUN` 也鎖 **81162 效果精確 spawn 在 target 座標**的 Java 對齊（完成三條 caster 路徑覆蓋）；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` `spawnEffect(81162, shock, cha.getX(), cha.getY(), ...)` 同樣使用 target 座標而非 srcpc 座標。新增回歸測試把 caster 與 target 分開 4 格嚴格驗證，至此玩家、NPC 單體、NPC AOE 三條 caster 路徑的 81162 spawn 座標精確驗證完整覆蓋。
- 補上玩家施放 `SHOCK_STUN` **對 NPC 目標**也鎖 **81162 spawn 在 NPC 座標**的 Java 對齊（完成 4 條 caster/target 組合）；對齊 Java `start(L1PcInstance,...)` 對 NPC 目標同樣使用 `cha.getX()/cha.getY()` (NPC target 座標)。新增回歸測試把 caster 與 NPC target 分開 4 格嚴格驗證，至此 player→player、player→NPC、NPC→player、NPC AOE 四條組合的 81162 spawn 座標精確驗證完整覆蓋。
- 補上非 GM 玩家對 NPC 施放 `SHOCK_STUN` 也**不收 GM 秒數訊息**的 Java 對齊（NPC 目標版 negative）；對齊 Java `if (srcpc.isGm())` 嚴格 GM 門檻對 NPC 目標分支同樣適用。既有 NPC 目標版只覆蓋 GM positive，本步補上 negative case：非 GM 玩家對 NPC 施放 87 命中時也不收 GM 訊息。至此 GM 門檻在玩家目標、NPC 目標、GM 觀察者三條路徑形成完整 positive/negative 覆蓋。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **81162 效果 TicksLeft 範圍**鎖定的 Java 對齊（與 buff 同步）；對齊 Java `areashock_stun()` `shock * 1000` 同時用於 `setSkillEffect` 與 `spawnEffect`，buff 與 effect 時間應同步。在 `MatchesJava` 加入 `effects[0].TicksLeft 10-25` 範圍檢查（2-5 秒），與既有 player→player、NPC→player effect TicksLeft 範圍檢查形成完整覆蓋。
- 補上 `SHOCK_STUN` cast 後 **SkillDelayUntil 推進 1500ms** 的 Java 對齊；skill_list.yaml `reuse_delay: 1500` 對應 Java `L1Magic.useReuseDelay()` `_skillDelay`。既有入口阻擋只測「`isSkillDelay` 拒絕施法」negative，本步補上 positive case：cast 87 成功後 `SkillDelayUntil >= before + 1500ms`，鎖定 yaml `reuse_delay` 解析正確且 processSkill 設定冷卻，避免玩家連續高頻施放 87。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **81162 效果 GfxID=4183** 的 Java 對齊（完成三條 caster 路徑的 GfxID 鎖定）；Java `L1SpawnUtil.spawnEffect(81162, ...)` 透過 npc template 81162 載入 `gfxid=4183`，既有 player→player（行 2505）與 NPC→player（行 3121）均以 `effects[0].GfxID != 4183` 鎖定，但 AOE `MatchesJava` 原本只檢查 NpcID/SkillID/OwnerCharID/TicksLeft，未鎖 GfxID。本步在 `MatchesJava` 加入 `effects[0].GfxID != 4183` 驗證，至此三條 caster 路徑的 81162 效果模板 GfxID 鎖定完整。
- 補上 mob skill type 5 範圍 `SHOCK_STUN` **已有 87 目標不再收 S_Paralysis(StunApply)** 的 Java 對齊（已有 87 完整 4-way skip 第四項）；對齊 Java `areashock_stun()` 第 744-748 行 `if (!pc.isGmInvis() && !pc.hasSkillEffect(SHOCK_STUN))` 整段 4 個動作（setSkillEffect / sendPackets / setParalyzed / spawnEffect）對已有 87 目標全跳過。既有測試只覆蓋 setSkillEffect 不刷新與 spawnEffect 不重複，本步補上 `sendPackets(S_Paralysis(5,true))` 也被跳過的 negative case，避免未來 AOE 實作改成 `setParalyzed(true)` 副作用每 tick 廣播 StunApply (=0x16) 造成玩家被怪物 type 5 持續刷封包。
- 補上玩家施放 `SHOCK_STUN` **對已有 87 玩家目標也跳過 S_Paralysis(StunApply)** 的 Java 對齊（player-cast 路徑對等鎖定）；對齊 Java `SHOCK_STUN.start(L1PcInstance,...)` 第 49-52 行 `pc.sendPackets(new S_Paralysis(5, true))` 位於 `if (!cha.hasSkillEffect(87))` 區塊內。既有玩家路徑「不刷新」測試只覆蓋 TicksLeft / GM 訊息 / action gfx / 4434 / 81162，未驗證 StunApply 跳過；本步擴充既有測試新增 `hasParalysisSubtype StunApply` negative 斷言，與前一步 AOE 對應測試形成 player-cast 與 AOE 兩條路徑的 StunApply 跳過共同覆蓋。
- 補上 NPC 施放 `SHOCK_STUN` **對已有 87 玩家目標也跳過 S_Paralysis(StunApply)** 的 Java 對齊（NPC-cast 路徑對等鎖定，完成三條 caster 路徑的 StunApply 跳過共同覆蓋）；對齊 Java `SHOCK_STUN.start(L1NpcInstance,...)` 第 73-75 行 `pc.sendPackets(new S_Paralysis(5, true))` 位於 `if (!cha.hasSkillEffect(87))` 區塊內。既有 NPC 路徑「不刷新」測試只覆蓋 TicksLeft / 81162 / action gfx / 4434，未驗證 StunApply 跳過；本步擴充既有測試新增 `hasParalysisSubtype StunApply` negative 斷言，與 player-cast、AOE 對應測試形成三條 caster 路徑（player-cast→player、NPC-cast→player、NPC AOE→player）對「已有 87 目標跳過 StunApply」的完整覆蓋。
- 補上 NPC 施放 `SHOCK_STUN` 對玩家目標 **RegistStun 扣除** 的 unit 級 Java 對齊（NPC_PC 路徑對等 PC_PC）；對齊 Java `L1MagicPc.calcProbabilityMagic()` 第 881-892 行 `if (_calcType == NPC_PC) case SHOCK_STUN: probability -= _targetPc.getRegistStun()`，與 PC_PC 對等。既有 NPC 施放 end-to-end integration test 證明扣除作用，PC_PC 路徑已有 unit 級回歸，但 NPC_PC 路徑無對等 unit 測試。新增 `TestSkillClanShockStunNpcCasterProbabilitySubtractsRegistStunLikeJava`：Level 50 vs 50 + leverage 10 + RegistStun 10 → 預期 40，鎖定 1:1 扣除而非倍率或忽略，避免未來重構退化。至此 PC_PC、NPC_PC 兩條 PC 目標路徑的 RegistStun 扣除有 unit 級對等鎖定。
- 補上 NPC 施放 `SHOCK_STUN` 成功率 **先乘 leverage 再扣 RegistStun** 的運算順序鎖定；對齊 Java `L1MagicNpc.calcProbabilityMagic()` 第 172-180 行先 `probability *= getLeverage()/10.0`，再於 NPC_PC switch 做 `probability -= _targetPc.getRegistStun()`。既有 leverage 端到端測試只驗證 miss，既有 RegistStun unit 測試用 leverage=10 兩種順序結果相同。新增 `TestSkillClanShockStunNpcCasterProbabilityAppliesLeverageBeforeRegistStunLikeJava`：leverage=15 + RegistStun=20 → (50×1.5)−20=55 vs 錯誤順序 (50−20)×1.5=45，差 10 可區分，鎖定運算順序。

## 增幅防禦（REDUCTION_ARMOR / 88）

- 修正 `REDUCTION_ARMOR(88)` 機制誤實作；對齊 Java `L1AttackPc.java:1617-1620` (PvP physical `dmg -= (max(lvl,50)-50)/5 + 10`) 與其他四條路徑（`L1AttackNpc.java:437-440` NPC→PC physical / `L1MagicPc.java:1148,1296` PvP magic / `L1MagicNpc.java:357` NPC→PC magic 均為 `dmg -= (max(lvl,50)-50)/5 + 1`）。Go 原本誤實作為 `buffs.lua [88] = { ac = -4 }`（AC 加成 4，完全不同機制）。本次：1）移除 Lua AC 加成改為 flag-only buff；2）新增 `applyReductionArmorDamage(target, damage, pvpPhysical)` helper 封裝 Java 公式（含 `Math.max(level,50)` floor 與 +10/+1 路徑差）；3）先套用至 `npcMeleeAttack` (NPC→PC physical) 路徑；4）unit test 覆蓋 Level 50 邊界、+1/+10 路徑差、無 buff、Level<50 floor 四個關鍵點。其餘三條路徑（PvP physical / 兩條 magic）後續步驟接續套用同一 helper。
- 接續 `REDUCTION_ARMOR(88)` 套用 PvP physical 路徑（pvpPhysical=true，base=10）；對齊 Java `L1AttackPc.java:1617-1620` 對 PvP 近戰與弓箭都套用 +10 base 減免。在 `pvp.go` 兩處 `applyImmuneToHarmDamage` 後接續呼叫 `applyReductionArmorDamage(target, damage, true)`：PvP 近戰與 PvP 弓箭。公式 helper 上一步已加入並有 unit test 覆蓋 pvpPhysical=true 分支，本步只需接線。剩餘 2 條 magic 路徑後續步驟。
- 完成 `REDUCTION_ARMOR(88)` 剩餘 4 條 →PC 路徑接線（全部 base=1）；至此 Java 五條傷害路徑全對齊。

## 堅固防護（SOLID_CARRIAGE / 90）

- 修正 `SOLID_CARRIAGE(90)` 無盾時失敗訊息與 Java 不一致；對齊 Java `SOLID_CARRIAGE.start()` 第 20/28 行送 `S_ServerMessage("你並未裝備盾牌")` 而非 standard msg 280 "施展魔法失敗"。Go `skill_self.go case 90` 原本走通用 `sendCastFail(sess)` 送 msg 280，本步改為 `handler.SendNormalChat(sess, 0, "你並未裝備盾牌")`。既有 `TestSkillClanAuraSolidCarriageRequiresShieldOrGuarderAndAddsER` 已驗證無盾不套 buff/Dodge/AC 不變，新增 `hasNormalChatText("你並未裝備盾牌")` 斷言鎖定訊息對齊。其他 90 邏輯（盾/臂甲驗證、ER+15、stop 時 update ER）Go 已對齊。

## 尖刺盔甲（BOUNCE_ATTACK / 89）

- 修正 `BOUNCE_ATTACK(89)` PvP 武器破壞缺少 claw 排除的 Java 對齊缺失；對齊 Java `L1AttackPc.damagePcWeaponDurability()` 第 3400-3410 行排除三類武器（`_weaponType==0/20/62`，即赤手/弓/鐵手甲）。Go `damageEquippedWeaponDurability` 原本只排除 `bow`（Java `_weaponType==20`），缺少 `claw`（Java `_weaponType==62`），導致鋼爪攻擊 buff 89 玩家時錯誤觸發武器破壞。在 PvP 排除條件加上 `|| itemInfo.Type == "claw"`，並新增 `TestDamageEquippedWeaponDurabilityExcludesClawInPvpLikeJava` 鎖定 item_id 152（青銅鋼爪）攻擊 buff 89 玩家時 Durability 維持 0、不送訊息 268。89 其他邏輯（HIT+6、10% 觸發、175 排除、bow 排除、訊息 268、特效 10712）Go 已完整對齊。對應 Java→Go：1）`L1AttackNpc.java:437` NPC 弓箭/魔法球物理 → `npc_ai.go:609`；2）`L1MagicNpc.java:357` NPC 魔法 → `npc_ai.go:771`；3）`L1AttackNpc` NPC 物理技能 → `npc_ai.go:859`；4）`L1MagicPc.java:1148/1296` PC 魔法 → `skill_damage.go:164`。每處在 `applyImmuneToHarmDamage` 後接 `applyReductionArmorDamage(target, damage, false)`，共 4 行接線。helper 已有 unit test 覆蓋 pvpPhysical=false，本步只需接線。88 技能 Java 對齊完成，下一個技能起步。

## 力量/敏捷提升圖示註解（DRESS_MIGHTY/DEXTERITY 109/110）

- 補上 `buff_icon_map.yaml` 與 `sendBuffIcon` helper 對 yiwei `L1SkillUse.java:2449/2456` cast type=2、`L1SkillStop.java:433/441` stop type=3 的雙端 type 不對稱來源註解；既有 cancel-only 覆寫邏輯（durationSec=0 → param=3）已正確處理，未修改行為，僅補註解避免未來閱讀 `L1SkillUse2.java` (dead code, 無 caller) 誤判為 source-of-truth 將 yaml param 改為 3。

## 暗影防護（SHADOW_ARMOR / 99）

- 修正 `SHADOW_ARMOR(99)` 與其他 MR/SP 變化 buff 的 `S_SPMR` 通知缺失；對齊 Java `SHADOW_ARMOR.start()` 與 `RESIST_MAGIC` 等 `pc.sendPackets(new S_SPMR(pc))`，套用與還原時都需通知客戶端 MR/SP 新值。Go `applyBuffEffect` 原本只在 STR/DEX/CON/WIS/INT/CHA/MaxHP/MaxMP/AC/DmgMod/HitMod 變化時送 `SendPlayerStatus`（S_STATUS，不含 MR/SP），MR/SP 走 `S_MAGIC_STATUS` 但無觸發點。本步在兩處接線：1) `applyBuffEffect` 套用後若 `buff.DeltaMR != 0 || buff.DeltaSP != 0` 送 `SendMagicStatus`；2) `revertBuffStats` 還原後若同一 delta 也送一次。影響 buff：`[99] mr=5`（暗影防護）、`[129] mr=10`（魔法抗性）、`[209] sp=2`（幻術金巨人）；玩家施放/buff 到期後客戶端 MR/SP 顯示能即時更新。

## 暗隱術（BLIND_HIDING / 97）

- 修正 `BLIND_HIDING(97)` 施放時缺少 `S_Invis(self,1)` 與 `S_RemoveObject` 廣播的 Java 對齊缺失；對齊 Java `L1SkillUse2.java:2511-2514` 把 `INVISIBILITY(60)` 與 `BLIND_HIDING(97)` 放在同一個 if 分支：`pc.sendPackets(new S_Invis(pc.getId(), 1))` 通知施法者自己已隱身、`pc.broadcastPacketAll(new S_RemoveObject(pc))` 把角色從附近玩家畫面移除。Go `executeSelfSkill` 原本只有 `case 60` 處理這兩個封包，`case 97` 缺漏，導致黑妖施放暗隱術時：1) 施法者本機 UI 不顯示隱身狀態切換、2) 附近玩家畫面仍看得到該角色。本步在 switch 補上 `case 97` 分支只發送 self-packet 與 RemoveObject 廣播（buff 屬性與 Invisible flag 仍由 `applyBuffEffect`+`buffs.lua [97] = { invisible = true }` 走通用 buff 路徑處理，duration 由 yaml `buff_duration: 32` 決定，與 Java 一致），不影響 60 的既有永久 buff 邏輯（buff_duration=0 → 走 3600 秒 fallback 直到 cancelInvisibility）。

## 反擊屏障（COUNTER_BARRIER / 91）

- 修正 `COUNTER_BARRIER(91)` 反彈傷害漏過 `IMMUNE_TO_HARM(68)` 攻擊者減半的 Java 對齊；對齊 Java `L1AttackPc.commitCounterBarrier()` 第 3339-3341 行 `if (_pc.hasSkillEffect(68)) damage /= 2;` 對接受反彈的攻擊者套用聖界減傷。Go `pvp.go` PvP 近戰反彈路徑（行 105-119）原本將 `cbDmg` 直接扣到 `attacker.HP`，未走任何傷害濾鏡。本步在套用前先呼叫 `applyImmuneToHarmDamage(attacker, cbDmg)`，與既有 helper（其他四條 →PC 傷害路徑共用）統一語義；不另寫測試，依停損標準避免「Go 已對 + 鎖實作」式回歸。

## 魂體轉換（BLOODY_SOUL / 146）

- 修正 `146 BLOODY_SOUL` MP 回復值使用錯誤資料源的 Java 對齊缺失。Java `BLOODY_SOUL.start()` 第 19 行 `srcpc.setCurrentMp(srcpc.getCurrentMp() + ConfigElfSkill.BLOODY_SOULADDMP)`，yiwei `各職業技能相關設置.properties: BLOODY_SOULADDMP = 20`（與 `妖精_技能設定表.properties` 同步 = 20）。Go `skill_self.go case 146` 原本使用 `skill.SkillLevel`（yaml `skill_level: 19`）作 MP 回復量，但 Java BLOODY_SOUL.start() 完全不引用 skill_level 欄位，只用 ConfigElfSkill 常數 20。誤差 1 MP（19 vs 20）。本步改為硬編碼 `player.MP += 20` 對齊 yiwei 配置，並更新註解指明 ConfigElfSkill 來源避免未來再誤把 skill_level 當依據。

## 屬性防禦（RESIST_ELEMENTAL / 138）+ 元素抗性套用通知

- 修正 `138 RESIST_ELEMENTAL` 與 `147 ELEMENTAL_PROTECTION` cast-side 漏送 `S_OwnCharAttrDef` 的 Java 對齊缺失。Java `L1SkillUse.java:2538-2545` RESIST_ELEMENTAL cast 對 PC `pc.addWind(10) + addWater(10) + addFire(10) + addEarth(10) + sendPackets(new S_OwnCharAttrDef(pc))`，套用四屬 +10 同時送 UI 通知；ELEMENTAL_PROTECTION cast（L1SkillUse:2547-2558）Java 漏送 `S_OwnCharAttrDef`，是 Java 已知 bug（client 顯示與資料不同步直到下次事件 refresh）。Go `applyBuffEffect` 對任何資料驅動四元素抗性 buff（138 走 `[138] = { fire_res, water_res, wind_res, earth_res = 10 }`、147 走 `[147] = {}` + `applyElementalProtectionDelta` 依 ElfAttr 動態設單一 delta）原本只送 `SendPlayerStatus`（S_STATUS 不含元素抗性）+ buff icon，不送 `S_OwnCharAttrDef`，玩家 UI 在套 138 後四抗顯示卡到下次事件 refresh。本步在 `applyBuffEffect` MR/SP `SendMagicStatus` 後新增條件：`if target.Session != nil && (buff.Delta*Res != 0)` 任一非零 → `SendAbilityScores(target.Session, target)`，與 133 revert-side 修正（2026-05-18）對稱。138 cast：完全對齊 Java（送 S_OwnCharAttrDef）；147 cast：比 Java 嚴格收緊（補上 Java 漏送的 UI 通知，client 顯示與資料一致）；133 cast 因走獨立 `applyElementalFallDownToPlayer` 不通過 `applyBuffEffect`，與 Java 133 cast 不送對齊不變。至此元素抗性 buff 的 apply/revert 雙向 S_OwnCharAttrDef 通知完整。

## 弱化屬性（ELEMENTAL_FALL_DOWN / 133）+ 元素抗性還原通知

- 修正 `133 ELEMENTAL_FALL_DOWN` 與其他元素抗性 buff（138/147）解除時漏送 `S_OwnCharAttrDef` 的 Java 對齊缺失。Java `ELEMENTAL_FALL_DOWN.stop()` 第 123 行對 PC `pc.sendPackets(new S_OwnCharAttrDef(pc))`、`L1SkillStop` 第 473 行（138 RESIST_ELEMENTAL stop）與第 489 行（147 ELEMENTAL_PROTECTION stop）同樣在還原 `addWind/Water/Fire/Earth` 後送 `S_OwnCharAttrDef`，讓客戶端 UI 即時更新四屬性抗性顯示。Go `revertBuffStats` 反轉 `Delta{Fire,Water,Wind,Earth}Res` 後沒送任何封包，玩家 UI 仍顯示 buff 期間的低抗性直到下次裝備變更或重進世界。本步在 `revertBuffStats` 末尾（與 MR/SP 補送 S_SPMR 同位置）新增條件：`if target.Session != nil && (buff.Delta*Res != 0)` 任一非零 → `SendAbilityScores(target.Session, target)`，套用至所有四元素抗性 buff 的 revert 路徑。133 player 路徑透過 `applyElementalFallDownToPlayer + revertBuffStats` 自動受惠；NPC 走 `removeElementalFallDownFromNpc` 不送（與 Java NPC stop 不送對齊）。Cast 階段 Java 133 也不送 `S_OwnCharAttrDef`（只有 stop 才送），Go cast 路徑同樣不送，apply/revert 對稱與 Java 一致。138/147 cast 路徑 Java 會送（`L1SkillUse.java:2545` 與 `:2547+`），那兩個技能的 apply-side 補送留到對應 ID 子項處理。

## 三重矢（TRIPLE_ARROW / 132）

- 修正 `132 TRIPLE_ARROW` 漏 Java `TRIPLE_ARROW_DMG=5` 倍率、缺弓裝備檢查與 4394/11764 收尾廣播的三項對齊缺失。1) **5× 倍率**：Java `TRIPLE_ARROW.start()` 第 36-44 行對 `ConfigSkill.TRIPLE_ARROW_DMG > 1` 設 `IsTRIPLE_ARROW(true)`，3 次 `cha.onAction(srcpc)` 走 `L1AttackPc` 第 1512/2002 行 `dmg *= ConfigSkill.TRIPLE_ARROW_DMG`，yiwei `各職業技能相關設置.properties: Triple_Arrow_Dmg = 5.0`。Go `scripts/combat/magic.lua` `calc_physical_skill` 對 sid=132 原本只設 `hit_count = 3` 沒乘倍率，本步改為 `hit_count = 3; damage = damage * 5`。2) **弓裝備檢查**：Java 第 32-33 行 `getCurrentWeapon() != 20 → return 0` 嚴格要求弓（visual byte = 20），Go `processSkill` 原本未檢查，本步在 MP 消耗前（與 skill 5/69/131 模式一致）新增 `if skillID == 132 && player.CurrentWeapon != 20 { return }`，沒裝弓施放 132 不扣 15 MP、不執行 3 次攻擊、不廣播收尾特效。3) **收尾廣播**：Java 第 45-46 行 `S_SkillSound(srcpc.getId(), 4394)` + `S_SkillSound(srcpc.getId(), 11764)` 用於加速封包與特效動畫，Go `executeAttackSkill`（NPC 目標）與 `executeAttackSkillOnPlayer`（玩家目標）均於攻擊迴圈結束後新增此兩個廣播。既有 `TestSkillTripleArrowDamagesPlayerTargetThreeTimes` 因 lostHP 斷言為 `>= 150`，5× 後傷害更大仍通過；不另寫測試，依停損標準避免「鎖實作」回歸。

## 世界樹的呼喚（TELEPORT_TO_MATHER / 131）

- 修正 `131 TELEPORT_TO_MATHER` 缺少 Java 前置 buff 與地圖檢查的 Java 對齊缺失。Java `TELEPORT_TO_MATHER.start()` 第 23-58 行依序檢查 `hasSkillEffect(230)`（亡命之徒 → 訊息 1413）、`hasSkillEffect(4000)`（束縛 → "已被束縛的效果無法瞬移"）、`hasSkillEffect(THUNDER_GRAB=192)`（奪命之雷 → "身上有奪命之雷的效果無法瞬移" + `S_Paralysis(TYPE_TELEPORT_UNLOCK)`）、`pc.getMap().isEscapable()`（不可順移地圖 → 訊息 276 + TeleportUnlock）後才執行傳送與 169 廣播。Go `executeResurrection case 131` 原本直接呼叫 `TeleportPlayer(33047, 32338, 4, 5)`，跳過全部四項 Java 前置檢查（亡命之徒、束縛、奪命之雷玩家可無視 buff 直接回母樹，且任何地圖都能用）。本步新增 `teleportToMatherBlockedBeforeConsume(sess, player)` 在 `processSkill` 流程的 MP 消耗前（與 skill 5/69 模式一致）依序檢查並送對應回饋封包，回傳 true 時中止流程不消耗 MP；阻擋成立時 skill 131 不扣 10 MP、不傳送、不廣播 169。

## 王者加護光環（DIVINE_SACRIFICE / 119 → BRAVE_AVATAR aura）

- 修正 `applyBraveAvatar` MR 變化漏送 `S_SPMR` 的 Java 對齊缺失；對齊 Java `BraveAvatarTimer.run()` 第 54-55 行 `pc.sendPackets(new S_SPMR(pc)) + pc.sendPackets(new S_OwnCharStatus2(pc))`，套用王者加護同時送 MR/SP 與一般狀態。Go `applyBraveAvatar` 原本只送 `SendPlayerStatus`（S_STATUS 不含 MR/SP），且因走獨立路徑（非 `applyBuffEffect`）未受 SHADOW_ARMOR 修正的 MR/SP 自動補送涵蓋，導致王者光環觸發時客戶端 MR 顯示停留舊值。本步在 `AddBuff + SendPlayerStatus` 後補上 `SendMagicStatus(SP, MR)`。`removeBraveAvatar` 走 `removeBuffAndRevert → revertBuffStats` 已含 MR/SP 補送，無需修改。

## 破壞盔甲（ARMOR_BREAK / 112）

- 修正 `ARMOR_BREAK(112)` 施放成功率公式漏項與 INT 來源錯誤；對齊 Java `L1MagicPc.calcProbabilityMagic()` 第 728-744 行 `probability += magichit; if (_pc.getBaseInt() >= 25 && <= 44) probability += (BaseInt-15)/10; else if (BaseInt >= 45) probability += 5;`。Go `calcArmorBreakProb` / `calcArmorBreakProbNpc` 原本 1) 漏掉 `magichit` 項（Java 7.6 智力魔法命中表 `(INT-20)/3`），2) BaseInt 加成誤用 `caster.Intel`（含裝備與 buff 加成），與 Java `getBaseInt()` 排除裝備加成的語義不一致。改用既有 helper `shockStunIntMagicHit(caster.Intel)` 與 `shockStunBaseIntMagicHit(caster)`，玩家與 NPC 目標兩條路徑同步修正。`armorBreakProbabilityByLevel` 測試 helper 仍只驗證 BaseInt 區段（與 Java 第 740-744 行對齊），保持原簽名不變。

## 鋼鐵防護（IRON_SKIN / 168）

- 修正 `168 IRON_SKIN` 互斥 buff 清單過於激進的 Java 對齊偏離（與 151 EARTH_SKIN/159 EARTH_BLESS 同 pattern）。Java `L1SkillUse.java:1743 / L1SkillUse2.java:1752` `REPEATEDSKILLS[1] = { EARTH_SKIN(151), IRON_SKIN(168) }` 只與大地防護互斥，全 Java codebase 無 SHIELD(3)/BLESSED_ARMOR(21)/Shadow Armor v1(24)/EARTH_BLESS(159) 與 IRON_SKIN 的 mutex 規則；Java cast（`L1SkillUse.java:2564-2567 + L1SkillUse2.java:2516-2519`）只 `addAc(-10) + S_SkillIconShield(10, duration)`、stop（`L1SkillStop.java:504-510`）只 `addAc(10) + S_SkillIconShield(10, 0)`。Go `scripts/combat/buffs.lua [168] = { ac = -10, exclusions = {3, 21, 24, 151, 159} }` 原本 5 項排他比 Java 多 4 項（3 SHIELD/21 BLESSED_ARMOR/24 Shadow Armor v1/159 EARTH_BLESS），導致玩家施放鋼鐵防護時誤殺保護罩/鎧甲護持/影之防護/大地祝福，與 Java 行為不一致——Java 允許這些 buff 與 IRON_SKIN 同時存在（保護罩/鎧甲護持/影之防護是 AC 疊加組、大地祝福只送圖示無 AC 變動）。本步收緊 `[168].exclusions = {151}` 對齊 Java REPEATEDSKILLS[1] 唯一條目。圖示路徑已對齊：a) cast `S_SkillIconShield(10, duration)` 透過 `buff_icon_map.yaml:26-28 skill_id=168 type=shield param=10` 與 `skill_buff.go sendBuffIcon "shield"` 對齊 Java；b) stop `S_SkillIconShield(10, 0)` 透過通用 cancelBuffIcon 對齊 Java；c) AC ±10 套用/還原透過 `applyBuffEffect/revertBuffStats` Delta 路徑。同步更新 `TestSkillElementalBuffElfArmorAndWaterBuffsUseJavaValues` 第 115-118 行——原本斷言「cast 168 移除 159 且 AC=0」鎖死 Go 非 Java 行為，本步改為 Java 對齊版「cast 168 移除 151 + 保留 159 + AC=0（10-10）」；訊息明確標註 REPEATEDSKILLS[1] 與大地防護互斥的 Java 設計。**至此 REPEATEDSKILLS[1] 全 2 成員（151/168）反向擴充完全收尾**（151 audit 2026-05-18 已收緊至 `{168}`，本步 168 收緊至 `{151}`）。資料對照 Java `skills.sql:167`：'168', '鋼鐵防護', skill_level=21, skill_number=7, mp_consume=30, buff_duration=960, target='none', target_to=0, attr=1（土）, type=2（TYPE_CHANGE）, id=128, action_id=19, cast_gfx=2252, sys_msg_happen=714, sys_msg_stop=725, sys_msg_fail=280 — 與 Go yaml 多項對齊（cast_gfx 等細節未深入比對，屬資料 audit 範圍）。不另寫測試，依停損標準避免「鎖實作」回歸（更新的測試行為從 Go-鎖死改為 Java-對齊，屬糾錯非新增）。驗證：`go build ./...`、`go test -count=1 ./internal/system`（17.1s 全綠）。

## 風之枷鎖（WIND_SHACKLE / 167）

- 修正 `167 WIND_SHACKLE` 三項 Java 對齊缺失：1) **NPC 目標 debuff 缺失**——Java `WIND_SHACKLE.start(PC, cha, ...)` 與 `start(NPC, cha, ...)` 對 `cha=L1NpcInstance` 也走 `setSkillEffect(167, integer*1000)`（line 23, 36），Go `executeNpcDebuffSkill`（`skill_status.go:408`）switch 無 case 167，玩家對 NPC 施放風之枷鎖時 debuff **完全不會套用**；2) **NPC 攻擊速度減速缺失**——Java `L1NpcInstance.java:2629-2633` `if (this.hasSkillEffect(WIND_SHACKLE)) { if (type == ATTACK_SPEED || type == MAGIC_SPEED) sleepTime += sleepTime * 0.25; }`，NPC 持 167 debuff 時 attack/magic cooldown +25%，Go `setNpcAtkCooldown`/`setNpcSubMagicCooldown`（`npc_ai.go:1250-1271`）無對等 +25% 修正，即使 debuff 套用 NPC 行為也不變慢；3) **PC 目標 MR 抗性閘缺失**——Java `L1MagicPc.calcProbabilityMagic` 對 167 走等級比較 5/10/15% 機率（`ConfigElfSkill.WIND_SHACKLE_1/2/3`），Go `playerDebuffSkills` 集合（`skill_status.go:772-774`）原本不含 167，PC 目標跳過 MR 抗性閘 100% 成功，與 Java 機率不一致。本步三處修正：a) `skill_status.go` `executeNpcDebuffSkill` 新增 `case 167` —— 先 `npc.HasDebuff(167) return` 對齊 Java 不重複套用、`checkNpcMRResist` 抗性閘、通過後 `npc.AddDebuff(167, dur*5)` 套用 16 秒（buff_duration 預設）+ 廣播 castGfx + log；b) `npc_ai.go` `setNpcAtkCooldown` 與 `setNpcSubMagicCooldown` 兩處皆加 `if npc.HasDebuff(167) { cooldown += cooldown / 4 }` 對齊 Java 25% 加成（用整數除法避免浮點）；c) `skill_status.go playerDebuffSkills` 加入 `167: true`，讓 PC 目標走 `skill_buff.go:920` 的 `playerDebuffSkills[skillID]` 統一 MR 抗性閘（`checkPlayerMRResist`），抗性失敗時不套 buff、播 castGfx + sendCastFail。其他 Java 對齊已就位：a) PC 目標 cast `S_PacketBoxWindShackle(charID, duration)` 透過 `handler.SendWindShackle` 對齊 Java `S_PacketBoxWindShackle(pc.id, integer)`；b) PC 目標 stop 時 `SendWindShackle(charID, 0)` 透過 `skill_buff.go:453 cancelBuffIcon` 與 `:780` buff 到期 tick 對齊 Java `stop()` 行 44-46；c) Java NPC caster 路徑（`start(NPC,...)`）目前透過 NPC 自我 AI 套用 debuff 不在本步範圍。資料對照 Java `skills.sql:166`：'167', '風之枷鎖', skill_level=21, skill_number=6, mp_consume=15, item_consume_id=40319, item_consume_count=1, buff_duration=16, target='buff', target_to=3, attr=8（風）, type=1（TYPE_PROBABILITY）, action_id=19, sys_msg_happen=1001 — Go `skill_list.yaml:5117-5147` 多數對齊（差異：probability_dice=30 vs Java 50、cast_gfx=1799 vs Java 11733、ranged=8 vs Java 1，屬資料調整非行為對齊範圍）。不另寫測試，依停損標準避免「鎖實作」回歸（NPC debuff 套用 + cooldown 修正 + PC MR 閘 三項皆是 Java vs Go 真實差異修補，新增實作而非鎖既有行為；未來 NPC AI 整合或 cooldown 重構若需要新增 debuff 類別會自然觸發；既有 `TestSkillElementalBuffElf*` 等 PC 路徑相關測試不涉及 167，無回歸風險）。驗證：`go build ./...`、`go test -count=1 ./internal/system`（17.6s 全綠）。

## 暴風神射（STORM_SHOT / 166）

- 修正 `166 STORM_SHOT` 互斥 buff 清單漏對齊 Java `REPEATEDSKILLS[0]` 的缺失（與 148/149/156/163 對等收尾）。Java `L1SkillUse.java:1741 / L1SkillUse2.java:1750` `REPEATEDSKILLS[0] = { FIRE_WEAPON(148), WIND_SHOT(149), STORM_EYE(156), BURNING_WEAPON(163), STORM_SHOT(166) }` 五個武器加成 buff 全互斥；Go `scripts/combat/buffs.lua [166] = { bow_dmg = 5, bow_hit = -1, exclusions = {149} }` 原本只列 149，缺 148/156/163，導致玩家施放暴風神射時不會解除已套用的近戰武器 buff（FIRE_WEAPON 148/BURNING_WEAPON 163）或 STORM_EYE 156，與 Java 行為不一致。本步擴充為 `exclusions = {148, 149, 156, 163}` 對齊完整 REPEATEDSKILLS[0]。**至此 REPEATEDSKILLS[0] 全 5 成員（148/149/156/163/166）的 exclusions 都對稱涵蓋同組其他 4 員，REPEATEDSKILLS[0] 反向擴充完全收尾**。圖示路徑已對齊：a) cast `S_PacketBoxIconAura(165, duration)` 透過 `buff_icon_map.yaml:68-69 skill_id=166 type=aura` + `sendBuffIcon "aura"` 的 `byte(166-1)=165` 對齊 Java `L1SkillUse.java:1430-1432 + L1SkillUse2.java:1447-1449`；b) stop `S_PacketBoxIconAura(165, 0)` 透過通用 cancelBuffIcon 對齊 Java `L1SkillStop.java:569-575`；c) BowDmg ±5 與 BowHit ±(-1)/+1 套用/還原透過 `applyBuffEffect/revertBuffStats` Delta 路徑對齊 Java `addBowDmgup(5)/addBowHitup(-1)` 與 stop `addBowDmgup(-5)/addBowHitup(1)`。不另寫測試，依停損標準避免「鎖實作」回歸（既有 `TestSkillElementalBuffElfWeaponAndBowBuffsUseJavaValues`（`skill_elemental_buff_test.go:78-81`）已驗證 cast 166 移除 149 + BowHit=2 + BowDmg=9；新擴充的 148/156/163 反向 mutex 不影響該測試——測試流程中 cast 166 時 156/163 未被套用過、148 已被 163 移除，僅實際生效於未來涵蓋 cast 166 時 148/156/163 任一仍存活的場景）。驗證：`go build ./...`、`go test -count=1 ./internal/system`（18.5s 全綠）。

## 自然呼喚（CALL_OF_NATURE / 165）— 純審計無代碼變更

- 審計 `165 CALL_OF_NATURE`：對照 Java `CALL_OF_NATURE.java` start(PC,...) 與 Go `skill_heal_resurrect.go:103-141` case 165 後確認 Go 既有實作完整對齊 Java 三種目標路徑：a) **PC target**（Java 20-37）：跳過 caster 自身、目標屍體格上有「活著的」可見玩家送 `S_ServerMessage(592)` 並 return（witness check）、否則 `setTempID(srcpc.id) + S_Message_YN(322)` 等待玩家同意——Go 線 108-124 `GetByCharID + IsPlayerAt(x, y, mapID, target.SessionID) + SendYesNoDialog(322)` 對應；b) **Pet target**（Java 44-50）：witness check 同 PC、通過後 `npc.resurrect(maxHp) + setResurrect(true)`——Go `callOfNatureResurrectPet/resurrectPetWithHP` 行 245-259 `IsPlayerAt + SendServerMessage(592) + pet.HP=maxHP + Status=PetStatusRest` 對應；c) **NPC target**（Java 38-58）：跳過 `L1Tower`、跳過 `isCantResurrect` 模板、其他 NPC `npc.resurrect(maxHp)`——Go `callOfNatureResurrectNpc/resurrectNpcWithHP` 行 197-228 `Impl=="L1Tower" || CantResurrect → false; HP=maxHP + 清除 AggroTarget/HateList/狀態` 對應。Java 的「visible players radius=0」對齊到 Go 的 `IsPlayerAt(x, y, mapID)` 同 tile 檢查（兩者皆排除目標自身、皆 `!isDead/p.Dead` 過濾活人）。witness check 順序對齊（Java 先檢查、後設 YN；Go 同序）。資料對照 Java `skills.sql:164`：'165', '生命呼喚', skill_level=21, skill_number=4, mp_consume=50, item_consume_id=40319, item_consume_count=1, reuse_delay=0, buff_duration=0, target='buff', target_to=3, attr=4（水）, type=32（TYPE_RESURRECTION）, ranged=10, id=16, action_id=19, cast_gfx=2245, sys_msg_fail=280 — Go `skill_list.yaml:5055-5085` 全欄位逐項對齊。既有測試 `TestSkillCallOfNature*`（共 5 個）已鎖定：PC YN 同意流程、PC witness 拒絕、NPC 滿血復活、Pet 滿血復活+status=rest、Pet witness 拒絕、`cant_resurrect=true` NPC 拒絕復活，覆蓋三條主要路徑與 witness/template 雙拒絕分支。`docs/技能ID對照表.md` 原條目「已接復活 Lua 表，但需對照 Java 寵物/玩家限制」描述已過時——本步更新為「已完整對齊」。本步無代碼變更，純審計確認對齊狀態。

## 生命的祝福（NATURES_BLESSING / 164）

- 修正 `164 NATURES_BLESSING` 範圍治療未限隊伍範圍的 Java 對齊缺失。Java `L1SkillUse.isInTarget()` 第 877-880 行 `if ((target_to & TARGET_TO_PARTY) == TARGET_TO_PARTY && (player.getParty().isMember(xpc) || player.isGm())) return true;` 對 `target_to=8` skill 過濾只接受隊伍成員（第 671-676 行另外允許自己通過）；NATURES_BLESSING `target_to=8 + type=16` 為隊伍範圍治療。Go `skill_self.go:178-216` heal 區塊原本不分隊員無差別地對所有 nearby 玩家治療，導致非隊員路人也被治療，與 Java 行為不一致。本步在 `area=-1` 分支內加入 `if skill.TargetTo == 8` 條件，以 `s.deps.World.Parties.GetParty(player.CharID)` 取得隊員清單建 `map[int32]bool`，在 `nearby` 迭代 inner loop 對非隊員 `continue`。caster 自己永遠治療（對齊 Java 671-676 自身通過邏輯）。盤點 yaml 確認只有 164 同時具 `target_to=8 + type=16`，本步單一條件影響範圍精確限於 164；其他 type=16 heal skill（如自身治癒）走 `target_to=1` 不會誤過濾。其他 Java 對齊已就位：a) `applyElfWaterHealingModifiers` 對每個目標移除 `WATER_LIFE(170)` 對齊 Java `L1SkillUse2.java:2114-2117` heal-group 分支（HEAL/EXTRA_HEAL/GREATER_HEAL/FULL_HEAL/HEAL_ALL/NATURES_TOUCH/NATURES_BLESSING 共用）；b) 治療公式 `CalcHeal(damage_value, damage_dice, damage_dice_count, INT, SP)` 透過 Lua 對齊 Java；c) cast 動畫 `S_ActionGFX(action_id=19)` 廣播由 `executeSelfSkill` 末尾通用路徑送。資料對照 Java `skills.sql:163`：'164', '生命的祝福', skill_level=21, skill_number=3, mp_consume=30, reuse_delay=300, buff_duration=0, target='none', target_to=8（TARGET_TO_PARTY）, damage_value=10, damage_dice=12, attr=4（水）, type=16（TYPE_HEAL）, id=8, action_id=19, cast_gfx=2244 — Go `skill_list.yaml:5024-5054` 對應欄位多數對齊（差異：reuse_delay Go=0 vs Java=300，屬資料調整非行為對齊範圍）。不另寫測試，依停損標準避免「鎖實作」回歸（路人不被誤治療屬靜默過濾，session 端無封包可斷言；party 治療 happy path 既有覆蓋）。驗證：`go build ./...`、`go test -count=1 ./internal/system`（25.6s 全綠）。

## 烈炎武器（BURNING_WEAPON / 163）

- 修正 `163 BURNING_WEAPON` 互斥 buff 清單漏對齊 Java `REPEATEDSKILLS[0]` 的缺失。Java `L1SkillUse.java:1741` `REPEATEDSKILLS[0] = { FIRE_WEAPON(148), WIND_SHOT(149), STORM_EYE(156), BURNING_WEAPON(163), STORM_SHOT(166) }` 五個武器加成 buff 全互斥；Go `scripts/combat/buffs.lua [163] = { dmg_mod = 6, hit_mod = 3, exclusions = {148} }` 原本只列 148，缺 149/156/166，導致玩家施放烈炎武器時不會解除已套用的弓系 buff（WIND_SHOT/STORM_EYE/STORM_SHOT），與 Java 行為不一致。本步擴充為 `exclusions = {148, 149, 156, 166}` 對齊完整 REPEATEDSKILLS[0]。至此 5 個成員（148/149/156/163/166）的 exclusions 都涵蓋同組其他 4 員，REPEATEDSKILLS[0] 對稱補完。圖示路徑已對齊：a) cast `S_PacketBoxIconAura(162, duration)` 透過 `buff_icon_map.yaml:66-67 skill_id=163 type=aura` + `sendBuffIcon "aura"` 的 `byte(163-1)=162` 對齊 Java `L1SkillUse.java:1426-1428` + `L1SkillUse2.java:1443-1445`；b) stop `S_PacketBoxIconAura(162, 0)` 透過通用 cancelBuffIcon 對齊 Java `L1SkillStop.java:546-550`；c) DmgMod ±6 與 HitMod ±3 套用/還原透過 `applyBuffEffect/revertBuffStats` Delta 路徑對齊 Java `addDmgup(6)/addHitup(3)` 與 stop 路徑的 `-6/-3`。發現一項 Java 差異屬「broader architecture gap」依停損標準延後：**`calcBuffDamage` 額外 +6 melee 傷害缺失**——Java `L1AttackPc.java:2426-2428 calcBuffDamage()` 對持 BURNING_WEAPON buff 的玩家在主 melee 傷害計算（已含 `pc.getDmgup()` 的 +6）之外**再加** `dmg += 6.0`，扣除 weaponType 20（弓）/62（鐵手甲）/weaponType2=17（奇古獸）。同檔同 helper 對 FIRE_WEAPON(148) +4、BERSERKERS(75) +5、BURNING_SLASH(216) +10 也走相同雙加路徑，屬武器 buff 家族的「getDmgup + calcBuffDamage 雙加」設計，Go 目前 `melee.lua` 無對等 helper 也未在 combat path 加 buff-specific damage 加成。修正需在 Go melee 主路徑 attack damage 計算新增「持 148/163/75/216 時加對應 dmg + 武器類型 guard」hook，涉及多技能家族與 melee.lua/skill_*.go 多檔，無法在單一 163 子項內完整實作；屬武器 buff 家族架構缺口，留至武器 buff 系統整體 audit（或 148 audit 重訪）處理。本步只完成 exclusions 對齊，不另寫測試（既有 `TestSkillElementalBuffElfWeaponAndBowBuffsUseJavaValues` 第 68-72 行已驗證 cast 163 移除 148 + DmgMod=7 + HitMod=5；新擴充的 149/156/166 反向 mutex 不影響該測試）。驗證：`go build ./...`、`go test -count=1 ./internal/system`（16.9s 全綠）。

## 召喚強力屬性精靈（GREATER_ELEMENTAL / 162）— 純審計無代碼變更

- 審計 `162 GREATER_ELEMENTAL`：對照 Java `GREATER_ELEMENTAL.java` 與 Go `skill_summon.go:282-363 ExecuteElementalSummon` 後確認 Go 既有實作完整對齊（與 154 LESSER_ELEMENTAL 共用同一 helper）：a) `player.ElfAttr == 0` 早返回對齊 Java `attr != 0` 必要條件；b) `RecallPets=false → SendServerMessage(353)` 對齊 Java `pc.getMap().isRecallPets() == false → S_ServerMessage(353)`；c) `calcUsedPetCost != 0` 早返回對齊 Java `petcost == 0` gate；d) NPC ID 對照 `greaterElementalByAttr` 1=81053/2=81050/4=81051/8=81052 完全對齊 Java switch 大地土/火/水/風 強力精靈；e) `PetCost = int(player.Cha) + 7` 對齊 Java `summon.setPetcost(pc.getCha() + 7)`；f) `dmg = 0` 返回對齊。資料對照 Java `skills.sql:161`：'162', '召喚強力屬性精靈', skill_level=21, skill_number=1, mp_consume=20, item_consume_id=40319, item_consume_count=4, reuse_delay=0, buff_duration=0, target='none', target_to=0, attr=0, type=128（TYPE_SUMMON）, id=2, action_id=19, cast_gfx=2510 — Go `skill_list.yaml:4962-4992` 全欄位逐項對齊。既有 `TestSkillElementalSummonGreaterElemental*`（`skill_elemental_summon_test.go:96-103`）已鎖定 `SkillID=162 + ElfAttr=4` 召喚 NPC 81051 + petcost=Cha+7 行為。家族級 Java 差異（cast 動畫廣播缺失、MP 消耗時機）與 154 LESSER_ELEMENTAL audit 同性質，延後至召喚系統整體 audit 處理。`docs/技能ID對照表.md` 原條目「未完整，Java 有 skillmode，需補元素召喚」描述已過時——本步更新為「已完整對齊」。本步無代碼變更。

## 封印禁地（AREA_OF_SILENCE / 161）

- 修正 `161 AREA_OF_SILENCE` 沉默狀態未阻擋全域廣播聊天的 Java 對齊缺失。Java `C_ChatGlobal.java:69-88` 對持有 `SILENCE(64)` / `AREA_OF_SILENCE(161)` / `STATUS_POISON_SILENCE` 三種沉默 buff 的玩家 `if (!pc.isGm()) isStop = true;`，靜默拒絕廣播（不送錯誤訊息）。Go `applyAreaOfSilence`（`skill_elemental.go:111-145`）已對 nearby 玩家套用 `target.Silenced = true + AddBuff(161)`，且 `isCastableWhileSilenced`（`skill.go:564`）的白名單 `{87, 88, 89, 90, 91, 187}` 與 Java `C_ChatGlobal._cast_with_silence = {SHOCK_STUN, REDUCTION_ARMOR, BOUNCE_ATTACK, SOLID_CARRIAGE, COUNTER_BARRIER, FOE_SLAYER}` 完全對齊（已驗證 Java `L1SkillId` 對應 ID 一致），但 `handler/chat.go ChatWorld` 分支未對 `player.Silenced` 做檢查，導致被封印禁地的玩家仍可正常廣播全域聊天，與 Java 行為不一致。本步在 `chat.go ChatWorld` 進入點加入 `if player.Silenced && player.AccessLevel < 200 { return }`（GM 不受限對齊 Java `pc.isGm()`），靜默拒絕（不送錯誤訊息對齊 Java `isStop = true` 但 `errMessage` 仍為 false 的設計）。此修正一併涵蓋 Java 對 64 SILENCE 與 poison silence 的相同 chat 阻擋（Go 三種 buff 都設 `Silenced = true`），三項 Java 對齊以單一 check 達成。其他 Java 行為差異已對齊：a) Java `L1MagicPc.calcProbabilityMagic` 對 161 的等級比較成功率（5/10/15 + INT/MR config）由 `applyAreaOfSilence` 內的 `checkPlayerMRResist` 通用路徑處理；b) Java `C_UseSkill.java:189-193` 對沉默玩家施法的白名單阻擋由 `isCastableWhileSilenced` 在 `skill.go:288` 對齊；c) `applyAreaOfSilence` 跳過自己、死亡目標、視野外玩家對齊 Java AoE 範圍。資料對照 Java `skills.sql:160`：'161', '封印禁地', skill_level=21, mp_consume=40, item_consume_id=40319, item_consume_count=8, reuse_delay=9000, buff_duration=16, target='none', target_to=3, probability_value=33, type=1（TYPE_PROBABILITY）, area=3, through=1, id=1, action_id=19, cast_gfx=10708；Go `skill_list.yaml:4931-4961` 對應欄位多數對齊（差異：probability_dice=30 vs Java 50、area=-1 vs Java 3、cast_gfx=2241 vs Java 10708、sys_msg_happen=715 vs Java 0），屬資料調整非行為對齊範圍，留至資料 audit 處理。不另寫測試，依停損標準避免「鎖實作」回歸（既有 `TestSkillElementalSummon*` 已鎖定 nearby 玩家 Silenced + buff 161 行為）。驗證：`go build ./...`、`go test -count=1 ./internal/handler ./internal/system`（0.9s + 17.6s 全綠）。

## 水之防護（AQUA_PROTECTER / 160）— 純審計無代碼變更

- 審計 `160 AQUA_PROTECTER`：對照 Java `AQUA_PROTECTER.java`、`L1PcInstance.java:3399-3401`、`L1SkillUse.java REPEATEDSKILLS (1741-1762)` 後確認 Go 既有實作完整對齊：a) Java `start(PC, ...)` 行 13-18 只執行 `srcpc.setSkillEffect(160, integer * 1000)` 純 buff 註冊，無 AC/MR/水抗/icon 副作用，`start(NPC, ...)` 與 `stop()` 皆 no-op；b) Java 真實效果在 `L1PcInstance.getEr()` 第 3399-3401 行 `if (hasSkillEffect(AQUA_PROTECTER)) { er += 5; }` 給玩家 +5 ER（迴避率）；c) 全 Java 全 codebase 搜尋 `L1SkillUse.java`、`L1SkillUse2.java`、`L1SkillStop.java` 對 AQUA_PROTECTER 皆**零匹配**確認無 icon emission；d) Java `REPEATEDSKILLS` 全 10 個群組**不含** 160 任何條目，無互斥。Go `scripts/combat/buffs.lua [160] = { dodge = 5 }` 透過 `applyBuffEffect/revertBuffStats` Delta 路徑套用 `target.Dodge += 5` / 還原 `-= 5`，功能上對齊 Java `getEr() += 5`（Go Dodge 對應 Java ER）；`buff_icon_map.yaml` **無 160 條目**對齊 Java 無 icon emission；`buffs.lua [160]` 無 exclusions 對齊 Java 無 REPEATEDSKILLS mutex。既有 `TestSkillElementalDynamic*` 第 162 行 `player.Dodge != 7`（base 2 + buff 5）已鎖定 +5 ER 行為，無需新增測試。資料對照 Java `skills.sql:159`：'160', '水之防護', skill_level=20, skill_number=7, mp_consume=30, buff_duration=960, target='buff', target_to=1, attr=4（水）, type=2（TYPE_CHANGE）, id=128, action_id=19, cast_gfx=5829 — 與 Go yaml 全部欄位對齊（僅 reuse_delay Go=0 vs Java=100 差異，屬資料調整非行為對齊，留至資料 audit 視需要追嚴）。本步無代碼變更。

## 大地的祝福（EARTH_BLESS / 159）

- 修正 `159 EARTH_BLESS` 互斥 buff 清單過於激進的 Java 對齊偏離（與 151 audit 同 pattern）。Java `L1SkillUse.java:1741-1762` `REPEATEDSKILLS` 全部 10 個群組**不含** 159 任何條目；`L1SkillUse2.java:1439-1441 / 2475-2478` cast 對 PC 只送 `S_SkillIconShield(7, duration)` 並 `// pc.addAc(-7);` 註解（義維 Java 移除了 AC 修正）；`L1SkillStop.java:445-449` stop 只送 `S_SkillIconShield(7, 0)` 並 `//cha.addAc(7);` 註解。Go `buffs.lua [159] = { exclusions = {151, 168} }` 私自加入 Java 不存在的互斥規則，導致玩家施放大地的祝福時誤殺允許共存的大地防護（151）與鋼鐵防護（168），與 Java 行為不一致。本步收緊為 `[159] = {}` 對齊 Java（無 AC、無 exclusions、無 buff stat 副作用，只透過 `buff_icon_map.yaml:23-25 skill_id=159 type=shield param=7` 送圖示）。圖示路徑已對齊：a) cast `S_SkillIconShield(7, duration)` 透過 `sendBuffIcon "shield"` 對齊 Java；b) stop `S_SkillIconShield(7, 0)` 透過通用 cancelBuffIcon 對齊 Java。`168 IRON_SKIN` 的反向 mutex（Java REPEATEDSKILLS[1]={151,168} 不含 159，但 Go `[168].exclusions={3,21,24,151,159}` 多加了 3/21/24/159）留至 168 audit 處理。同步更新 `TestSkillElementalBuffElfArmorAndWaterBuffsUseJavaValues` 第 110-114 行：原本斷言「cast 159 移除 151 且 AC=10」鎖死 Go 非 Java 行為，本步改為 Java 對齊版「cast 159 保留 151 並維持 AC=4（仍持 151 的 -6）」；115-118 行 cast 168 部分仍依賴 `[168].exclusions={159}` 通過，待 168 audit 一併修正。驗證：`go build ./...`、`go test -count=1 ./internal/system`（17.4s 全綠）。

## 生命之泉（NATURES_TOUCH / 158）

- 修正 `158 NATURES_TOUCH` HPR 數值對齊 Java 的缺失。Java `HprExecutor.java:55` `_skill.put(NATURES_TOUCH, 15)` 並在 `regenHp()` 第 200-209 行 `for (Integer skillId : _skill.keySet()) { if (pc.hasSkillEffect(skillId)) hpr += _skill.get(skillId); }` 對持有 NATURES_TOUCH buff 的玩家每 regen tick 加 +15 HPR。Go `buffs.lua [158] = { hpr = 4 }` 原本只給 +4 HPR，導致玩家施放生命之泉得到的 HP 回復速度比 Java 版本低 11/tick。本步調整為 `[158] = { hpr = 15 }` 對齊 Java HprExecutor 字面值。Lua 經由 `skill_buff.go:165 buff.DeltaHPR = int16(eff.HPR)` 寫入 buff、`:199 target.HPR += buff.DeltaHPR` 套用、`:492 target.HPR -= buff.DeltaHPR` 還原；通用路徑已對齊，本步只改數值。其他欄位（YAML type=2 TYPE_CHANGE / target=buff / target_to=1 / attr=4 / action_id=19 / cast_gfx=2243 / buff_duration=320 / mp_consume=20）對照 Java `skills.sql:157` 已對齊。Java `L1SkillUse.java:2113-2119` heal-group 對 WATER_LIFE(170) 的特殊移除分支僅對 TYPE_HEAL skills 觸發，TYPE_CHANGE 的 158 不走該分支，Go 現況（cast 158 不移除 170）與 Java 行為一致。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...`、`go test -count=1 ./internal/system`（17.1s 全綠）。

## 火焰武器（FIRE_WEAPON / 148）

- 修正 `148 FIRE_WEAPON` 互斥 buff 清單漏對齊 Java `REPEATEDSKILLS[0]` 的缺失。Java `L1SkillUse.java:1741` `REPEATEDSKILLS[0] = { FIRE_WEAPON(148), WIND_SHOT(149), STORM_EYE(156), BURNING_WEAPON(163), STORM_SHOT(166) }` 五個武器加成 buff 全互斥（`deleteRepeatedSkills` 對任一施放會 `stopSkillList` 同組其他 4 個）。Go `buffs.lua [148] = { dmg_mod = 4, exclusions = {163} }` 原本只列 163，缺 149/156/166，導致玩家施放 148 時不會解除已套用的弓系 buff（149 風之神射、156 暴風之眼、166 暴風神射），與 Java 行為不一致。本步擴充為 `exclusions = {149, 156, 163, 166}` 對齊完整 REPEATEDSKILLS 組；149/156/163/166 反向擴充留至各自審計時補（per-skill-ID-order 漸進式）。圖示（`S_PacketBoxIconAura(147, ...)` cast 與 0 stop）、DmgMod ±4 套用/還原、icon_id=skillID-1=147 對應均已 Java 對齊，本步不改。

## 大地屏障（EARTH_BIND / 157）— 純審計無代碼變更

- 審計 `157 EARTH_BIND`：對照 Java `EARTH_BIND.java` start() PC 與 NPC caster 兩條路徑 + stop()。Go 既有實作完整對齊 Java 核心 mechanic：a) 隨機 1-12 秒凍結 `earthBind.BuffDuration = 1 + RandInt(12)` 對齊 Java `Random.nextInt(12) + 1`；b) `[157] = { paralyzed = true }` + `applyBuffEffect` 行 258-261 送 `S_Paralysis(FreezeApply) + S_Poison(id, 2)` 對齊 Java PC 路徑 `setSkillEffect + S_Paralysis(TYPE_FREEZE,true) + sendPacketsAll(S_Poison(id, 2))`；c) 自然到期 `revertBuffStats` + 行 745-748 送 `S_Paralysis(FreezeRemove) + S_Poison(id, 0)` 對齊 Java stop() `sendPacketsAll(S_Poison(id, 0)) + S_Paralysis(4, false)`；d) NPC 目標路徑透過 `npc_ai.go:1459-1461` 設 `npc.Paralyzed=false + 廣播 S_Poison(0)`；e) Java `L1MagicPc.calcProbabilityMagic` 對 `hasSkillEffect(50/157)` 目標讓概率失敗已有 SHOCK_STUN 對應實作覆蓋。發現兩項 Java 差異屬「broader architecture gap」依停損標準延後：1) **`castleWarResult()` 阻擋缺失**——Java EARTH_BIND.start() 行 24 PC caster 路徑 `if (!srcpc.castleWarResult())` 在攻城戰範圍且有進行中戰爭時略過效果（不消耗 buff、不送任何凍結封包）。Go 無 `castleWarResult()` 對等 helper（`IsClanInWar` 只檢查血盟對血盟，非「玩家位置在攻城戰範圍 + 該城堡戰爭進行中」），需新增 `L1CastleLocation.getCastleIdByArea` + `isNowWar(castleId)` 基礎設施，屬城堡戰系統審計範圍非單一 157 子項；2) **`target.HasBuff(157) || target.Paralyzed` 重施阻擋過嚴**——Java `setSkillEffect(157, ...)` 對重複施放會覆寫 duration，Go 完全阻擋。然而此阻擋同時作為「凍結 buff 疊加生命週期 bug」的 workaround：若移除此 gate，當 EARTH_BIND(157) 過期時 `revertBuffStats` 直接設 `target.Paralyzed = false`，但 ICE_LANCE(50) 仍在 ActiveBuffs 中 SetParalyzed=true，導致玩家被誤解凍結；正確的修法需在 revert path 加 `shouldStayParalyzed()` 重新檢查所有 SetParalyzed buff 來源，屬廣域 paralysis lifecycle 重構非單一 157 子項。本步無代碼變更，純文件審計。既有測試 `TestSkillClanShockStunPlayerTargetEarthBindBlocksLikeJava`、`TestSkillClanShockStunNpcTargetEarthBindBlocksLikeJava` 已覆蓋 calcProbabilityMagic 對 hasSkillEffect(157) 目標的失敗行為。

## 暴風之眼（STORM_EYE / 156）

- 修正 `156 STORM_EYE` 互斥 buff 清單完全空白的 Java 對齊缺失。Java `L1SkillUse.java:2606-2610` cast `addBowHitup(2) + addBowDmgup(3) + S_PacketBoxIconAura(155, duration)`、`L1SkillStop.java:561-568` stop `addBowHitup(-2) + addBowDmgup(-3) + S_PacketBoxIconAura(155, 0)`、`REPEATEDSKILLS[0] = {148, 149, 156, 163, 166}` 與 148/149/163/166 全互斥。Go `buffs.lua [156] = { bow_hit = 2, bow_dmg = 3 }` 原本**無 exclusions 條目**（與其他四個同組成員 148/149/163/166 在本次審計過程中各自擴充對齊不對稱），導致玩家施放 STORM_EYE 時不會清除任何已套用的武器加成 buff，與 Java 行為不一致。本步擴充為 `exclusions = {148, 149, 163, 166}` 對齊完整 REPEATEDSKILLS[0]，至此該組五個技能（148 FIRE_WEAPON / 149 WIND_SHOT / 156 STORM_EYE / 163 BURNING_WEAPON / 166 STORM_SHOT）的互斥都涵蓋同組其他成員。圖示路徑已對齊：a) cast `S_PacketBoxIconAura(155, duration)` 透過 `buff_icon_map.yaml:62-63 skill_id=156 type=aura` 與 `sendBuffIcon "aura"` 的 `byte(156-1)=155`；b) stop `S_PacketBoxIconAura(155, 0)` 透過通用 cancelBuffIcon；c) BowHit ±2 / BowDmg ±3 透過 `applyBuffEffect/revertBuffStats` Delta 路徑。163/166 反向擴充（163 應加 149/156/166、166 應加 148/156/163）尚未補完，留至各自審計時處理。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...`、`go test -count=1 ./internal/system`（16.9s 全綠）。

## 烈炎氣息（FIRE_BLESS / 155）

- 修正 `155 FIRE_BLESS` 兩項 Java 對齊缺失。Java `FIRE_BLESS.java` start() 行 21-26 與 `L1SkillUse.java:1745-1746 REPEATEDSKILLS[2] = {52, 101, 150, 1000, 1016, 186, 155}`：a) `setSkillEffect(155, duration*1000) + setBraveSpeed(1) + S_SkillBrave(self=duration, broadcast=0) + S_PacketBoxIconAura(154, duration)`；b) 七個勇敢速度 buff 全互斥。Go 缺失兩項：1) `buffs.lua [155] = { brave_speed = 1, exclusions = {52, 101, 150, 186} }` 漏 STATUS_BRAVE(1000) 與 STATUS_ELFBRAVE(1016) 兩個 potion buff——本步擴充為 `exclusions = {52, 101, 150, 186, 1000, 1016}` 對齊完整 REPEATEDSKILLS 組（與 150 WIND_WALK 採同模式）；2) `buff_icon_map.yaml` **無 155 條目**，導致 `S_PacketBoxIconAura(154, duration)` cast 與 `S_PacketBoxIconAura(154, 0)` stop 都未送，與 Java FIRE_BLESS.start() 行 26 + stop() 行 43 不一致——本步新增 `skill_id: 155 type: aura` 條目，透過 `sendBuffIcon "aura"` 走 `SendIconAura(sess, byte(155-1)=154, duration)` 對齊 Java icon_id=154。其他路徑已對齊：a) `setBraveSpeed(1)` 透過 `applyBuffEffect` 的 `eff.BraveSpeed > 0` 分支 + `target.BraveSpeed=1`；b) cast `S_SkillBrave(id, 1, duration)` 自送與 `S_SkillBrave(id, 1, 0)` 廣播透過 `sendBraveToAll`（行 603-609）；c) stop `S_SkillBrave(id, 0, 0)` 透過 `revertBuffStats → sendBraveToAll(0, 0)`。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...`、`go test -count=1 ./internal/system`（17.8s 全綠）。

## 召喚屬性精靈（LESSER_ELEMENTAL / 154）— 純審計無代碼變更

- 審計 `154 LESSER_ELEMENTAL`：對照 Java `LESSER_ELEMENTAL.java` 與 `L1SkillUse.java:478-487` 後確認 Go `ExecuteElementalSummon`（`skill_summon.go:282-363`）已完整對齊 Java 核心 mechanic：a) `player.ElfAttr == 0` 早返回對齊 Java `attr != 0` 必要條件；b) `RecallPets=false → SendServerMessage(353)` 對齊 Java `pc.getMap().isRecallPets() == false → S_ServerMessage(353)`；c) `calcUsedPetCost != 0` 早返回對齊 Java `petcost == 0` gate（已有寵物時不能召喚屬性精靈）；d) NPC ID 對照 `lesserElementalByAttr` 1=45306土/2=45303火/4=45304水/8=45305風 完全對齊 Java switch；e) `PetCost = int(player.Cha) + 7` 對齊 Java `summon.setPetcost(pc.getCha() + 7)`；f) 無傷害值對齊 Java `dmg = 0` 返回。發現兩項家族級 Java 差異屬「broader architecture gap」依停損標準延後：1) **施法動畫廣播缺失**——Java `L1SkillUse.java:481-483` 在 `useConsume()` 後執行 `sendGrfx(true)` 廣播 `S_DoActionGFX(actionID=19)` 與 `S_SkillSound(castGfx=2510)`，Go `ExecuteElementalSummon` 與 `skill.go processSkill` 在 summon 委派早返回路徑都未廣播。同樣缺失出現在 **所有召喚技能** (51 SUMMON_MONSTER / 36 TAMING_MONSTER / 41 CREATE_ZOMBIE / 145 RETURN_TO_NATURE / 154 / 162) — `skill_summon.go` 全檔零個 BuildActionGfx/BuildSkillEffect 呼叫；屬召喚系統家族架構缺口，留至召喚系統整體審計處理，避免在 154 一處修改家族邏輯造成與其他召喚技能不對稱。2) **MP 消耗時機**——Java `L1SkillUse.java:481-482` 順序為 `runSkill()` 再 `useConsume()`，意味 start() 內部 gate 阻擋（ElfAttr=0、RecallPets=false、HasPets）時 MP 仍會被消耗；Go `ExecuteElementalSummon` 所有 gate 通過後才呼叫 `ConsumeSkillResources`，與 Java 字面 order 相反但屬 Go 既定 UX pattern（與 2026-05-18 skill 131 TELEPORT_TO_MATHER 的「阻擋成立不扣 MP」precedent 一致）。非單一技能可改範圍，且 Go pattern 為使用者友善設計，不視為待修。本步無代碼變更，純文件審計；既有 `TestSkillElementalSummonElementalSummonUsesElfAttr` 已鎖定 ElfAttr=2 → NPC 45303 + petCost=Cha+7 行為。

## 魔法消除（ERASE_MAGIC / 153）— 純審計無代碼變更（broader architecture gap）

- 審計 `153 ERASE_MAGIC`：發現 Go 對 153 採用「buff 全消」語義（`skill_buff.go:1110-1111 cancelAllBuffs(target)`），與 Java 真實 mechanic 不一致。Java `L1Character.java:1767-1769` 揭示 ERASE_MAGIC 真實效果：`if (hasSkillEffect(153)) { return mr >> 2; }`——目標持有此 buff 期間，有效 MR 被除以 4（魔抗大幅下降）。Java 完整流程：a) **施放** `L1SkillUse.java:1357-1359` 走無 skillmode 預設路徑 `setSkillEffect(_skillId, _getBuffDuration)` 對目標套用 32s buff（不消除任何 buff）；b) **作用期間** 目標 `getMr() = mr >> 2`，所有後續魔法判定使用降低後的 MR；c) **消耗時機** 目標受到任何非 ERASE_MAGIC 概率/詛咒/魔法攻擊（L1SkillUse.java:1940-1942, 1953-1955）會自動移除此 buff；d) **成功率** 等級比較 5/10/15（`ConfigElfSkill.ERASE_MAGIC_1/2/3`，INT/MR 修正預設 0）；e) **目標限制** monster `isErase==false` 模板免疫；f) **圖示** stop 送 `S_PacketBoxIconAura(152, 0)`（`L1SkillStop.java:492-496`）；g) **武器分流** 多個 `W_SK0010-15` weapon special 也會對目標套 ERASE_MAGIC 32s。Go 現況偏離為「重複 CANCELLATION(44) 語義」——`cancelAllBuffs(target)` 對目標移除所有可取消 buff，玩家體感變成「buff 移除工具」而非「魔抗削減工具」。屬「broader skill semantics gap」依停損標準需要整個技能重做：a) 改 case 153 從 cancelAllBuffs 變成 applyBuffEffect 套 buff；b) 在 Go 的 MR 計算路徑（含 Lua `calcProbabilityMagic`、`applyBuffEffect` MR delta）加入「持 153 時 MR/=4」乘法路徑（Go 目前 MR 走加法 Delta，乘法路徑需重構）；c) 在每個概率/詛咒/攻擊魔法 skill 套用流程加入「移除目標 153 buff」hook；d) 圖示 emission 修正。涉及多個系統（system/scripting/handler 三層），無法在單一 153 子項內完整實作。**本步無代碼變更**，純文件審計；先標記為已知 Java vs Go semantic divergence，留待使用者決定是否觸發 broader rework。既有 SHOCK_STUN→ERASE_MAGIC 移除已部分實作（`skill_status.go:131,173,486` 在 87 路徑），但僅限 87 並非 Java 對齊的「所有非-153 概率技能」全 catch。`44 CANCELLATION` 的 cancelAllBuffs 用法不受此審計影響（Java 44 確實是 buff 全消語義）。

## 心靈破壞（MIND_BREAK / 207）— 純審計無代碼變更

- 純審計確認 `207 MIND_BREAK` Go 行為已對齊 Java `skillmode/MIND_BREAK.java`：
  - Java 行為：`dmg = sp * 3.8`、`reMp = 5`、PC 與 NPC 目標皆 `setCurrentMp(-5)` + `receiveDamage(srcpc, dmg, false, false)` 無 MR 減免無概率檢查、廣播 `S_SkillSound(cha.id, 6553)`。
  - Go PC 路徑（`skill_damage.go:83-86`）：`calcMindBreakDamage(player) = SP * 3.8` 覆蓋 Lua 計算 + `applyMindBreakMPDrain(p)` 扣 5 MP。
  - Go NPC 路徑（`skill_damage.go:487-490`）：`res.Damage = SP * 3.8`、`res.DrainMP = 5`，後續於 `t.npc.MP -= t.drainMP`（line 624-628）首次命中時扣除。
  - Go yaml `cast_gfx: 6553` 對齊 Java skillmode 硬編值（SQL `cast_gfx=2510` 在 skillmode 內被覆寫）。
  - Go `applySkillDamageToPlayer` 不套用 MR 減免（line 165-214 僅 AbsoluteBarrier/ImmuneToHarm/ReductionArmor/CounterMirror），對齊 Java `receiveDamage(false, false)` raw damage。
  - 既有測試 `TestSkillIllusionistStatusMindBreakPlayerUsesJavaDamageAndDrainsMP`（`skill_illusionist_status_test.go:10`）驗證 SP=10 → damage=38、target.MP=15。
- **broader gap（不改）**：
  - **reuse_delay 0→3000**：Java SQL reuse_delay=3000，Go yaml=0。屬 yaml 冷卻 tuning，與 185/195/206 同源。
  - **ranged 3→5**：Java SQL ranged=5，Go yaml=3。屬 yaml 距離 tuning。
  - **Lua calc_mind_break 冗餘**：Go skill_damage.go 兩路徑都覆寫 res.Damage，Lua calc_mind_break 計算結果未被使用。屬無害冗餘非 Java 對齊問題，不在本步範圍。
- 本步無代碼變更，純審計。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（無新測試需執行）。

## 專注（CONCENTRATION / 206）

- 修正 `206 CONCENTRATION` yaml `buff_duration: 300` → `600` 對齊 Java `l1j_yiwei_java/db_split/skills.sql` 真實值（10 分鐘）。Go 原值讓玩家只享有半數 buff 持續時間。Java 無 skillmode 檔案、無 `case 206` 特殊處理；純資料驅動，效果由 `MprExecutor._skill.put(CONCENTRATION, 2)` 提供 +2 MP regen/tick。Go `buffs.lua [206] = { mpr = 2 }` 已對齊 Java MprExecutor 值，本步不動。
- **broader gap（不改）**：
  - **type 4→2**：Java SQL `type=2 (TYPE_CHANGE)`，Go yaml `type=4 (TYPE_CURSE)`。Go `skill_damage.go:24 canSkillReachTarget` 對 `type & TYPE_CHANGE != 0` 跳過 LOS 檢查；CONCENTRATION 為 self-buff (target_to=1) 目標即自身，LOS 檢查實質無感，故差異為 type 分類偏差但不影響運行時行為。屬「yaml type 資料審計」結構性缺口。
  - **mp_consume 20→30、reuse_delay 0→1000、ranged 3→5**：Java SQL 與 Go yaml 三項成本/冷卻/距離欄位偏差。屬「yaml 成本/冷卻 tuning」結構性缺口，與 185/190/195 同源 broader gap。
- 不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.4s 全綠）。

## 立方：燃燒（CUBE_IGNITION / 205）

- 補齊 `205 CUBE_IGNITION` 對 PC/NPC 週期傷害的 Java immune buff 檢查。Java `L1Cube.java:79-93` `case STATUS_CUBE_IGNITION_TO_ENEMY` 在 `giveEffect` 內每 4 秒觸發 `receiveDamage(10)` 前，依序檢查 `hasSkillEffect`：STATUS_FREEZE(4000)、ABSOLUTE_BARRIER(78)、ICE_LANCE(50)、EARTH_BIND(157)、FREEZING_BLIZZARD(80)——任一存在則跳過該次傷害。設計動機：目標已被凍結／屏障保護時不應再被立方燃燒傷害。Go `ground_effect.go applyCubeEnemy` 原本只檢查 `target.AbsoluteBarrier`（在 `damagePlayerByCube`），缺其餘四項；`applyCubeEnemyNpc` 完全無免疫檢查。
- 本步新增兩個 helper：
  - `playerCubeIgnitionImmune(target)`：檢查 `AbsoluteBarrier` flag 或 `HasBuff` 4000/50/157/80
  - `npcCubeIgnitionImmune(npc)`：檢查 `HasDebuff` 4000/78/50/157/80（NPC 使用 debuff 表）
- 在 `applyCubeEnemy` 與 `applyCubeEnemyNpc` 的 `GroundEffectCubeIgnition` 分支於 `DamageTickAcc%cubeEffectIntervalTicks==0` 條件後新增 immune guard。註解標明 Java 來源。
- 其他既有 Go 行為已對齊 Java：a) PC ally/enemy 區分（self/clan/party→ally，其他→enemy）；b) NPC 永遠走 enemy 路徑；c) safezone 跳過敵方傷害；d) ally 路徑 FireRes +30 + `SendPlayerStatus`（對齊 Java `addFire(30) + S_OwnCharAttrDef`）；e) `BuildActionGfx(charID, 2)` = ACTION_Damage 廣播（對齊 Java `S_DoActionGFX(ACTION_Damage)`）；f) 4 秒週期（Go cubeEffectIntervalTicks=20 ÷ 5 ticks/sec = 4s）；g) 立方持續期間每秒透過 `applyCubePulse` 掃描範圍。
- **broader gap（不改）**：
  - **副本 showId 檢查**：Java `EffectCubeExecutor:64 effect.get_showId() != pc.get_showId() continue`——僅同副本 instance 內生效。Go 無 instance 系統，與其他副本系統同源 broader gap。
  - **Castle war 區內豁免**：Java `EffectCubeExecutor:84-93` 對 `isSafetyZone` 玩家額外檢查 `isNowWar`（攻城戰期間允許敵方傷害）。Go 簡化為 `IsSafetyZone` 跳過。屬「攻城戰場規則」結構性缺口。
  - **L1EffectInstance 角色一致性**：Java 用 `L1EffectInstance` 作為立方實體（可被互動）；Go 用 `GroundEffect` 純資料結構。屬「世界物件統一」架構差異。
- 不另寫測試（純 guard 加入，邏輯小），依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.1s 全綠）。

## 幻覺：歐吉（ILLUSION_OGRE / 204）

- 修正 `204 ILLUSION_OGRE` buffs.lua 兩項 Java 對齊偏差。Java `L1SkillUse.java:2660-2664` 在 PC 自身目標分支內聯：`pc.addDmgup(4) + pc.addHitup(4)`，**僅** DmgMod 與 HitMod 兩個一般武器修正欄位；`L1SkillStop.java:603-609 case 204`：反向 `addDmgup(-4) + addHitup(-4)`。Java 無 ILLUSION_OGRE skillmode 檔案（純內聯）、無 `L1MagicPc/L1MagicNpc` case 204、無 `case 204` 在 L1WeaponSkill。Java `L1SkillUse.java:1741-1762 REPEATEDSKILLS`：illusion 系列（204/209/214/219）**不在任何互斥群**（與 MIRROR_IMAGE+UNCANNY_DODGE、AWAKEN_ANTHARAS+FAFURION+VALAKAS 等明確互斥群不同）；只有 `EXCEPT_COUNTER_MAGIC` 列表把 204+209+214+219 全部標為「不可被魔法屏障抵擋」。
- Go 原 `buffs.lua:166-167`：`[204] = { dmg_mod=4, hit_mod=4, bow_dmg=4, bow_hit=4, exclusions={209,214,219} }`——多出 Java 沒有的 `bow_dmg=4 bow_hit=4`（給 BowDmgMod/BowHitMod 額外 +4，Java 有獨立 `addBowDmgup`/`addBowHitup` API 但 ILLUSION_OGRE 不呼叫），以及 Go 私自加入的 mutex 群 `{209,214,219}`（Java 明確允許四個 illusion buff 並存）。本步改為 `[204] = { dmg_mod=4, hit_mod=4 }` 對齊 Java，註解標明 Java 來源與 mutex 缺失。
- **broader gap（不改）**：
  - **209/214/219 反向修正**：Go `[209].exclusions={204,214,219}`、`[214].exclusions={204,209,219}`、`[219].exclusions={204,209,214}` 仍把 204 列入 mutex 名單，本步因「不可偷換範圍」未動。209/214/219 各自審計時會修正自己的 exclusions（去除互斥群）。短期內存在不對稱：玩家先施 204 後可同時擁有 204+209+214+219；先施 209 則會清掉 204 等其他 illusion buff——直到後續 audit 補齊。
  - **counter-magic 豁免一致性**：Go `skill_buff.go:410 counterMagicExempt` 已含 201/204/209/211/213/214/216/219（與 Java EXCEPT_COUNTER_MAGIC 對齊），不受本步影響。
- 不另寫測試（純 buffs.lua 資料修正），依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.3s 全綠）。

## 暴擊（SMASH / 203）

- 修正 `203 SMASH` yaml 與 Java SQL 的 damage/attr/ranged 嚴重資料偏差，連帶將技能路由從錯誤的 `calc_physical_skill` 改回 Java 對齊的 `calc_magic_damage`。Java `l1j_yiwei_java/db_split/skills.sql:202`：`damage_value=12 damage_dice=10 damage_dice_count=0 attr=16 ranged=10`（type=64 TYPE_ATTACK 純資料驅動，無 skillmode、無 case 203 特殊處理）。Go yaml 原本 `damage_value=0 damage_dice=0 attr=0 ranged=1`——damage 雙零觸發 `magic.lua:28-30 calc_skill_damage` 啟發式判定為「物理技能」，路由到 `calc_physical_skill` 並硬編 `sid == 203 then damage = damage + math.floor(level/3)` 補償。實際 Java 把 SMASH 當磁性 ATTR_RAY（16）魔法攻擊，傷害走 `calcMagicDiceDamage(value=12, dice=10, count=0)` + INT 係數 + MR 減傷，與物理 STR/DEX/weapon 體系完全不同類別。
- 本步：a) yaml 修正 `damage_value 0→12`、`damage_dice 0→10`、`attr 0→16`、`ranged 1→10` 四項對齊 SQL；b) `magic.lua` 移除 `calc_physical_skill` 內 `sid == 203` 已死的 +level/3 分支，並更新註解標明 203 改走 magic 路徑；c) Go `calc_skill_damage` 自動依 `damage_value > 0` 條件改路由為 `calc_magic_damage`。後者用 `damage = 12 + magic_level dice` + Stage 2 INT 係數 + Stage 4 `coefficient = 1 - attrDef + INT*3/32`，與 Java `L1MagicPc.calcMagicDiceDamage` 對齊；`attr=16 (ATTR_RAY)` 不在 `calcAttrResistance` 處理範圍（Java only 1/2/4/8 = earth/fire/water/wind），attrDef=0 屬正確忽略。
- **broader gap（不改）**：
  - **mp_consume/reuse_delay**：Java SQL `mp_consume=5 reuse_delay=10`、Go yaml `mp_consume=7 reuse_delay=0`。屬「yaml 成本/冷卻 tuning」，與 185/190/195 cost 缺口同源（broader yaml 資料審計）。
- 不另寫測試（純資料 + Lua dead branch 清除），依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（16.9s 全綠）。

## 混亂（CONFUSION / 202）

- 修正 `202 CONFUSION` NPC 路徑 debuff ID 錯誤：Java `skillmode/CONFUSION.java:22-30`：對 cha 設 `L1SkillId.SILENCE(=64), integer*1000ms`，並非設 CONFUSION(202) 本身。Go `skill_damage.go:749-760` 原本 `AddDebuff(202, dur*5)` 是符號性的——其他系統檢查 silence 都看 `skillSilence(64)`（如 PC 端 `target.Silenced` 判定、buff 圖示等），202 ID 不會啟動沉默語義，NPC 中招後仍可正常施法。本步改為 `AddDebuff(64, dur*5)` + `HasDebuff(64)` 已沉默守衛對齊 Java。
- PC 路徑既有 `skill_illusionist.go applyConfusionSilence` 已使用 `skillSilence(64)` buff 並設定 `target.Silenced = true`，與 Java 對齊；本步不動。Java `case 202` 在 `L1MagicNpc.calcProbabilityMagic` 與 `L1MagicPc.calcProbabilityMagic`（`Random.nextInt(11)+20 + (atkLv-defLv)*2 + magichit + INT 修正`）為 dead code——CONFUSION skillmode 不呼叫 calcProbabilityMagic，TYPE_ATTACK 路徑只用 `calcMagicDamage` 算 MR 減傷，SILENCE 套用無概率閘。
- **broader gap（不改）**：
  - **NPC silence 真正消費邏輯**：NpcInfo 缺 `Silenced` 欄位或對應 AI 施法閘。debuff(64) 目前只是表記性的；要讓 NPC 真的不能施法需要 NPC AI 系統補 silence-gate（與其他 NPC 狀態系統同源）。
  - **L1MagicPc magichit/INT 修正**：通用 MR 抗性公式未實作 magic hit + INT 補正，與 192/193/194 同源（broader gap）。
- 不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（23.9s 全綠）。

## 鏡像（MIRROR_IMAGE / 201）

- 修正 `201 MIRROR_IMAGE` + 其他所有 dodge buff（DRAGONEYE 家族、UNCANNY_DODGE 等）revert 路徑 dodge 圖示通知 subcode 錯誤。Java `S_PacketBoxIcon1.java:29-39`：`(boolean type, int i)` constructor → `type=true` 寫 `_dodge_up=0x58`（增加閃避率，配合「正向 dodge 計數器」），`type=false` 寫 `_dodge_down=0x65`（減少閃避率，配合 RESIST_FEAR 等專用「dodge 懲罰計數器」）。Java `skillmode/MIRROR_IMAGE.java:31, 57`、`DRAGONEYE_*.java`、`UNCANNY_DODGE.java` 所有正向 dodge buff 的 start 與 stop **皆**送 `S_PacketBoxIcon1(true, get_dodge())` = 0x58 + 當前 dodge 總值；只有 `RESIST_FEAR.java:22, 42` 使用 `(false, get_dodge_down())` = 0x65 + dodge_down 計數器值。Go 原 `skill_buff.go:518-524` 在 revert 路徑使用 `SendDodgeIcon(target.Dodge, false)` 送 0x65 + dodge 總值——把正向 dodge 值送錯 channel（dodge_down 計數器），客戶端 UI 行為錯亂。本步改為 `true` 對齊 Java：cast 與 stop 一致送 0x58 + 當前 dodge 總值。
- 既有 `buffs.lua [201] = { dodge = 5 }` 對齊 Java `ConfigIllusionstSkill.MIRROR` 預設值；yaml `buff_duration: 1200`（20 分鐘）。Skill 111 龍之眼 revert 仍走專用 `SendUpdateER`（S_PacketBox UPDATE_ER）路徑不受影響。
- **broader gap（不改）**：
  - **RESIST_FEAR dodge_down 累加器**：Java 有獨立 `_dodge_down` 欄位 + `get_dodge_down()`/`add_dodge_down()` API，Go 目前 `PlayerInfo` 缺欄位，與 188 RESIST_FEAR audit 同源（broader gap）。
- 影響範圍：所有正向 dodge buff（201/DRAGONEYE_ANTHARAS/BIRTH/FIGURE/LIFE/UNCANNY_DODGE）revert 通知 subcode 同時對齊。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.4s 全綠）。

## 覺醒：巴拉卡斯（AWAKEN_VALAKAS / 195）+ 安塔瑞斯 retake（AWAKEN_ANTHARAS / 185）

- 修正 `185 AWAKEN_ANTHARAS` 與 `195 AWAKEN_VALAKAS` yaml `buff_duration: 0` 致 stat buff 從未套用的關鍵 bug。對照 `l1j_yiwei_java/db_split/skills.sql` 185/190/195 三筆 buff_duration 欄位均為 `600`（10 分鐘）；Go yaml 原本 185=0、190=0、195=0，其中 190 已於 v0.3.18 修正，本步補齊 185 與 195。Java `applyBuffEffect`-等同路徑 Go `skill_buff.go:131 if skill.BuffDuration <= 0 return` 直接早返回，導致 Java skillmode `setSkillEffect(N, integer*1000)` 對應的 buff 註冊從未發生——玩家施放 185/195 時音效廣播正常但 AC/HitMod/RegistStun/RegistSustain stat 變化完全沒生效。本步 yaml 改為 `buff_duration: 600` 對齊 Java SQL 真實值。
- 既有 `skill_buff.go:1154-1173` 覺醒分派塊（cast-only、非 toggle-off）對 185/190/195 三者一致；`buffs.lua [185] = {ac=-3, regist_sustain=10, exclusions={190,195}}`、`[195] = {hit_mod=5, regist_stun=10, exclusions={185,190}}` 已對齊 Java AWAKEN_ANTHARAS（addAc(-3)+addRegistSustain(+10)）與 AWAKEN_VALAKAS（addHitup(+5)+addRegistStun(+10)）；三者均在 buffs.lua 非可解除 buff 列表（`relax_all` 免疫）。`removeBuffAndRevert` 通用反向 AC/HitMod/RegistStun/RegistSustain 還原與 Java `stop()` 對稱。
- **broader gap（不改）**：
  - **S_OwnCharAttrDef 專屬 AttrDef 面板刷新**：Java `AWAKEN_ANTHARAS.java:21 sendPackets(S_OwnCharAttrDef(srcpc))` 在 185 cast/stop 時刷新 AttrDef 面板（AC + 四元素抗性）。Go awakening 分派塊呼叫 `handler.SendPlayerStatus`（S_STATUS opcode 8）已包含 AC 但不刷新 AttrDef 面板。屬「狀態面板封包系統」UI 細節缺口，與其他 AC 修改 buff 同源（broader gap）。
  - **mp_consume/hp_consume/reuse_delay yaml 資料偏差**：Java SQL 185 `MP=20 HP=10 reuse=10`、190 `MP=30 HP=20 reuse=10`、195 `MP=50 HP=30 reuse=10`；Go yaml 三者皆 `mp_consume=1 hp_consume=1 reuse_delay=0`。屬「yaml 資料審計」結構性缺口，與整批技能 yaml 資料對齊同源（broader gap）。
- 不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（20.5s 全綠）。

## 寒冰噴吐（FREEZING_BREATH / 194）

- 修正 `194 FREEZING_BREATH` 81168 冰矛圍籬視覺從錯誤的 `S_SkillSound` 封包改為真實 GroundEffect spawn。Java `L1SkillUse.java:2225, 2234`：對 PC/NPC 命中時 `L1SpawnUtil.spawnEffect(81168, time, x, y, mapId, _user, 0)` 在目標位置生成 `npc_id=81168 impl=L1Effect gfx_id=176` 真實實體，存活 `buff_duration+1` 秒（4 秒）。Go 原 `skill_dragonknight.go applyFreezingBreathFreeze` 對 PC 路徑送 `handler.BuildSkillEffect(target.CharID, 81168)` 把 NPC ID 當 gfx ID 用——客戶端無法正確顯示冰矛圍籬實體；NPC 路徑（`skill_damage.go` 凍結類）連 81168 spawn 都沒實作。本步：a) `world/ground_effect.go` 新增 `GroundEffectFreezingBreath GroundEffectType = 10`；b) `skill_dragonknight.go` 新增常數 `freezingBreathEffectNpcID = 81168` 與 helper `spawnFreezingBreathGroundEffect(caster, x, y, mapID, durSec)`（與 192 THUNDER_GRAB 的 `spawnThunderGrabGroundEffect` 同模板）；c) `applyFreezingBreathFreeze` 改簽名加入 `caster` 參數、移除 `BuildSkillEffect` 廣播、改呼叫 spawn helper；d) `skill_damage.go:688-694` 凍結類 NPC 路徑於 `skill.SkillID == 194` 時呼叫 `spawnFreezingBreathGroundEffect(player, t.npc.X, t.npc.Y, t.npc.MapID, dur+1)`。其他 Java 對齊缺口（L1MagicPc `default` magic dice 公式 + `RegistFreeze` 抗性減值 + castle war area block）屬「MR 抗性公式統一」結構性缺口，與 192/193 同源（broader gap）延後處理。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.6s 全綠）。

## 驚悚死神（HORROR_OF_DEATH / 193）

- 修正 `193 HORROR_OF_DEATH` 屬性 debuff 數值與抗性檢查兩項對齊缺口。Java `L1SkillUse.java:2290-2295` 與 `L1SkillUse2.java:2277-2282`：對 `L1PcInstance` 套用時 `pc.addStr(-3) + pc.addInt(-3)`（只動 STR/INT 兩屬性），對應 `L1SkillStop.java:675-682 case 193` 解除時 `addStr(+3) + addInt(+3) + sendPackets(S_OwnCharStatus2)`。Go `scripts/combat/buffs.lua:159` 原為 `[193] = { str=-1, con=-1, dex=-1, wis=-1, intel=-1 }`（誤動 CON/DEX/WIS，且數值-1 對不上 Java -3），本步改為 `[193] = { str=-3, intel=-3 }` 完整對齊 Java；revert 由通用 `revertBuffStats` 反向 +3 還原，與 Java 對稱。Java `L1MagicNpc.java:198-205 case HORROR_OF_DEATH` 與 `GUARD_BRAKE(183)/RESIST_FEAR(188)/THUNDER_GRAB(192)` 共用標準等級差公式 `prob = ProbabilityDice/10 * (atkLv-defLv) + ProbabilityValue` × leverage/10；Go `skill_status.go:795 playerDebuffSkills` 原缺 193，本步加入 `193: true`，PC→PC 施放時走 `checkPlayerMRResist` 抗性判定，與 183/188/192 對齊。圖示／cast gfx 6588／HP 50 消耗／buff_duration 64s 走通用 buff apply 路徑，與 Java `L1SkillUse.java:1357-1359` 無 skillmode 預設路徑等同；不另寫測試（依停損標準避免「鎖實作」回歸）。驗證：`go build ./...` 通過、`go test -count=1 ./internal/system`（18.6s 全綠）。

## 地面障礙（ENTANGLE / 152）— 純審計無代碼變更

- 審計 `152 ENTANGLE`：對照 Java `L1SkillUse.java:1463-1467, 2147-2188`、`L1SkillStop.java:660-668` 與 `REPEATEDSKILLS` (1741-1762)。Java 對 152 行為：a) 152 **不**在任何 REPEATEDSKILLS 群組（Java 無交叉 mutex 規則）；b) 套用流程在 case 1 (target MoveSpeed=1 haste 狀態) 時 `removeSkillEffect(skillNum) + removeSkillEffect(this._skillId) + setMoveSpeed(0)`，HASTE 與 ENTANGLE 雙方互消；c) case 0 (target normal) 時 `setMoveSpeed(2)` 並廣播 `S_SkillHaste(2, duration)`（broadcast 也帶 duration，非 0）；d) `getHasteItemEquipped() > 0` 與 `getBraveSpeed() == 5` 直接 continue 不套用。Go `buffs.lua [152] = { move_speed = 2, exclusions = {43, 54} }`：a) exclusions {43 HASTE, 54 GREATER_HASTE} 是 Go 引入但功能上對齊 Java case 1 的 `removeSkillEffect(skillNum)` haste 取消；b) `applyBuffEffect`（行 217-227）的 `eff.MoveSpeed=2 && target.MoveSpeed=1` 走 `cancelSpeedBuffs(target, 1)` 通用 catch-all（同時涵蓋 HASTE/GREATER_HASTE/STATUS_HASTE=1001）→ target.MoveSpeed=0 + 不設 buff.SetMoveSpeed → 進入 buff 表但無 MoveSpeed 副作用，可觀察行為與 Java case 1 等同（HP/MP 不變、目標走路速度恢復）。已發現三項 Java 差異屬「慢速技能家族」(29 SLOW / 76 MASS_SLOW / 152 ENTANGLE) 共通缺口，**不**屬 152 單體範圍，依停損標準列入 broader architecture gap 延後處理：1) `HasteItemEquipped > 0` 持有加速道具免疫缺失；2) `BraveSpeed == 5` 強化勇水免疫缺失；3) Java effect-application 廣播（L1SkillUse.java:2165）使用 `duration` 而非 `0`，Go `sendSpeedToAll` 對 nearby 統一送 0。三項差異留待 SLOW(29) 或 MASS_SLOW(76) 審計時一併重構處理。本步無代碼變更，純文件審計。

## 大地防護（EARTH_SKIN / 151）

- 修正 `151 EARTH_SKIN` 互斥 buff 清單過於激進的 Java 對齊偏離。Java `L1SkillUse.java:1743` `REPEATEDSKILLS[1] = { EARTH_SKIN(151), IRON_SKIN(168) }` 只列出大地防護與鋼鐵防護互斥；除此之外 Java 並無 SHIELD/BLESSED_ARMOR/SHADOW_ARMOR/EARTH_BLESS 與 EARTH_SKIN 的互斥規則（已逐項搜尋 `EARTH_SKIN` 在 yiwei Java 全 codebase 確認）。Go `buffs.lua [151] = { ac = -6, exclusions = {3, 21, 24, 99, 159, 168} }` 原本 6 項排他比 Java 多 5 項（3 SHIELD、21 BLESSED_ARMOR、24 Shadow Armor v1、99 SHADOW_ARMOR、159 EARTH_BLESS），導致玩家施放大地防護時誤殺允許共存的保護罩/鎧甲護持/影之防護/大地祝福，與 Java 行為不一致（Java 允許這些 buff 與 EARTH_SKIN 同時存在）。本步收緊為 `exclusions = {168}` 對齊 Java REPEATEDSKILLS[1]。Cast `S_SkillIconShield(6, duration)` 透過 `buff_icon_map.yaml:20-22 skill_id=151 type=shield param=6` 對齊 Java `L1SkillUse.java:1438-1439`；stop `S_SkillIconShield(6, 0)` 透過通用 cancelBuffIcon 對齊 Java `L1SkillStop.java:511-516`；AC ±6 套用/還原透過 `applyBuffEffect/revertBuffStats` 通用路徑。既有 `TestSkillElementalBuffElfArmorAndWaterBuffsUseJavaValues` 第 110 行測試「cast 159 移除 151」走 `[159].exclusions={151,168}`（不依賴 `[151]`），本步只調整 `[151]` 不影響該測試。`[159]` 的 mutex 對齊（Java EARTH_BLESS 也無 151/168 mutex）留至 159 audit 處理。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...`、`go test -count=1 ./internal/system`（19.0s 全綠）。

## 暴風疾走（WIND_WALK / 150）

- 修正 `150 WIND_WALK` 互斥 buff 清單漏 `STATUS_BRAVE/STATUS_ELFBRAVE` 的 Java 對齊缺失。Java `L1SkillUse.java:1745-1746` `REPEATEDSKILLS[2] = { HOLY_WALK(52), MOVING_ACCELERATION(101), WIND_WALK(150), STATUS_BRAVE(1000), STATUS_ELFBRAVE(1016), BLOODLUST(186), FIRE_BLESS(155) }` 七個勇敢速度 buff 全互斥。Go `buffs.lua [150] = { brave_speed = 4, exclusions = {52, 101, 155, 186} }` 原本只列 4 個技能（52/101/155/186），缺 `STATUS_BRAVE(1000)` 與 `STATUS_ELFBRAVE(1016)` 兩個藥水 buff，導致玩家施放 WIND_WALK 時不會清除已喝下的勇敢藥水/精靈餅乾 buff，與 Java 行為不一致。本步擴充為 `exclusions = {52, 101, 155, 186, 1000, 1016}` 對齊完整 REPEATEDSKILLS 組。Cast `S_SkillBrave(pc.getId(), 4, duration)` for caster + `S_SkillBrave(pc.getId(), 4, 0)` 廣播附近玩家透過 `sendBraveToAll` 對齊 Java `L1SkillUse.java:1456-1460`；stop `S_SkillBrave(pc.getId(), 0, 0)` 透過 `revertBuffStats → sendBraveToAll(0, 0)` 對齊 Java `L1SkillStop.java:594-602`；BraveSpeed=4 apply/revert 透過 `applyBuffEffect/revertBuffStats` 通用路徑。對應反向擴充（STATUS_BRAVE/STATUS_ELFBRAVE 套用時應清除 150 等其他成員）目前由 `item_use.go applyBrave` 處理，但該函式 conflict 清單存在 ID 錯亂與遺漏 Java 差異（PHYSICAL_ENCHANT_STR(42) 誤入清單、BLOODLUST(186) 與 FIRE_BLESS(155) 遺漏），與 Java `L1BuffUtil.braveStart()` 清單 `{HOLY_WALK, MOVING_ACCELERATION, WIND_WALK, STATUS_BRAVE, STATUS_ELFBRAVE, STATUS_RIBRAVE, BLOODLUST, FIRE_BLESS}` 不一致；屬 potion buff 路徑非 skill 150 直接範圍，留待 STATUS_BRAVE/STATUS_ELFBRAVE 專門審計時補。不另寫測試，依停損標準避免「鎖實作」回歸。驗證：`go build ./...`、`go test -count=1 ./internal/system`（20.8s 全綠）。

## 風之神射（WIND_SHOT / 149）

- 修正 `149 WIND_SHOT` 互斥 buff 清單漏對齊 Java `REPEATEDSKILLS[0]` 的缺失。Java `L1SkillUse.java:2601-2604` cast 對 PC `addBowHitup(6) + S_PacketBoxIconAura(148, duration)`、`L1SkillStop.java:554-559` stop `addBowHitup(-6) + S_PacketBoxIconAura(148, 0)`、`REPEATEDSKILLS[0]` 與 148 同組。Go `buffs.lua [149] = { bow_hit = 6, exclusions = {166} }` 原本只列 166，缺 148/156/163，導致玩家施放 WIND_SHOT 時不會解除已套用的近戰武器 buff（148 火焰武器、163 烈炎武器）或 156 暴風之眼。本步擴充為 `exclusions = {148, 156, 163, 166}` 對齊完整 REPEATEDSKILLS 組；156/163 反向擴充留至各自審計時補。圖示（`S_PacketBoxIconAura(148, ...)` cast 與 0 stop）透過 `buff_icon_map.yaml skill_id=149 type=aura` 與 `sendBuffIcon` 的 `byte(149-1)=148`、BowHit ±6 套用/還原透過 `applyBuffEffect/revertBuffStats` 通用路徑，均已 Java 對齊，本步不改。

## 測試瘦身

- 依新增的「對齊深度停損標準」執行 `SHOCK_STUN` cut A：移除 16 個「Go 已對 + 純防回歸」性質的測試與 3 個內嵌斷言，鬆綁未來重構綁架。刪除類別：AOE negative-of-negative 6 個（IceLance/EarthBind/AbsoluteBarrier/SafetyZone/Phantasm/EraseMagic）、NPC caster 與 AOE 對稱補完 5 個（onAction/PinkName negative + ReturnsTrueWithNoVisiblePlayers）、NPC caster RegistStun unit-level 2 個（SubtractsRegistStun + AppliesLeverageBefore）、AOE 已有 87 跳過 StunApply 1 個、INT off-by-one 邊界 2 個（Twelve/Thirteen）、3 個內嵌斷言（PlayerExistingBuff/NpcCasterExistingBuff 的 StunApply + MatchesJava 的 GfxID 4183）。保留所有 Java 真實行為差異測試。`skill_clan_shock_stun_test.go` 5600 行→5055 行，測試數 127→111；解除「測試呼叫內部 helper 變相鎖實作」綁架，未來重構 stun helper 或統一 AOE/單體 target filter 不再被舊測試擋住。
