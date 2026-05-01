# 待推送變更
## 技能系統整合
- 將正式技能實作從 `skill_batch_*.go` 搬入 owner/domain 檔案，並將測試檔同步改為 owner/domain 命名。
- 新增 `skill_damage.go`、`skill_buff.go`、`skill_heal_resurrect.go`、`skill_status.go`、`skill_weapon.go`、`skill_polymorph.go`、`skill_ground_effect.go`、`skill_elemental.go`、`skill_clan.go`、`skill_dragonknight.go`、`skill_illusionist.go`、`skill_self.go`、`skill_teleport.go` 作為維護入口。
- 進一步瘦身 `skill.go`，將復活/治療、武器與防具賦予、Buff 管理、NPC debuff、自身技能路由抽到 owner 檔。
- 清除正式程式碼中的 `applyBatch*` 命名殘留，改為幻術師/龍騎士語意函式名。
- 清除技能測試中的 `Batch*` helper 與測試函式命名殘留，保留原有回歸覆蓋但改用語意名稱。
- 將 `executeAttackSkill()` 與起死回生術移入 `skill_damage.go`，讓 `skill.go` 保持入口與 dispatch 職責。
- 在 `skill.go` 的關鍵 dispatch 補上 `Owner:` 註解，讓技能 ID 可直接追到主要實作檔。
- 新增 `docs/技能系統檔案索引.md`，記錄 owner 規則與各技能檔責任邊界。

## 技能批次 BB
- 補齊 `SHAPE_CHANGE` 67：自我施放開啟 monlist、對控制戒指玩家開啟目標端列表、對無戒指玩家依 Java 固定清單隨機變形。
- 補齊 NPC 變形術：限制 50 級以下與 Java 黑名單 NPC，保存原始 GFX，並在 debuff 到期、魔法相消、重生、復活時還原。
- 修正技能變形 monlist 選擇階段二次消耗魔法寶石的問題，改由施法階段單次消耗。
- 新增批次 BB 測試與文件 `docs/技能批次BB實作計畫.md`，覆蓋自我、玩家、控制戒指與 NPC 變形/還原。

## 技能批次 T
- 補齊精靈 `LESSER_ELEMENTAL` 154 與 `GREATER_ELEMENTAL` 162，依 `ElfAttr` 召喚對應屬性精靈並使用 `CHA + 7` 作為 petcost。
- 補齊 `AREA_OF_SILENCE` 161 範圍沉默，套用玩家沉默狀態並保留 MR 抵抗判定，施法者本身不受影響。
- 補齊 `COUNTER_MIRROR` 134 玩家魔法傷害反彈：以 WIS 對 random(100) 判定，觸發後反彈傷害、原傷害歸零並移除 buff。
- 新增批次 T 測試與文件 `docs/技能批次T實作計畫.md`，覆蓋元素召喚、封印禁地與鏡反射。

## GM 指令擴充
- 新增 `.poison [damage|silence|para]` — 對自己施加中毒；預設沉默毒（即「卡司特毒」，對應 NPC 卡司特 npc_id 45213 的 poison_atk: 2）。
- 新增 `.broken [數值1-127]` — 將自己當前裝備武器的 `Durability` 設為指定值；預設 127（極限損壞）。
- 擴充 `.item` — 同時支援數字 itemID 與中文物品名稱模糊查詢；完全相符優先，多筆候選列前 10 筆。
- 新增 `data.ItemTable.FindByName()` 中文模糊查詢、`system.GMApplyPoison()` 三毒類型施加 helper、`system.GMBreakWeapon()` 武器耐久強制設定 helper，並透過 `GMCommandManager` 介面暴露 `ApplyPoison` / `BreakWeapon`。

## 技能批次 S
- 補齊妖精技能動態邏輯：`ELEMENTAL_FALL_DOWN` 依施法者 `ElfAttr` 對玩家/NPC 降低單一屬性抗性 50，並在效果移除時還原。
- 修正 `AQUA_PROTECTER` 為 ER/Dodge +5；`WATER_LIFE` 讓下一次受治療加倍後移除，`POLLUTE_WATER` 讓受治療減半且不自動移除。
- 將 `ELEMENTAL_FIRE`、`STRIKER_GALE`、`SOUL_OF_FLAME` 改為戰鬥結算旗標，避免誤加固定能力值；近戰與遠程傷害結算已接入對應倍率。
- 新增批次 S 測試與文件 `docs/技能批次S實作計畫.md`，覆蓋屬性弱化、治療倍率與妖精戰鬥 buff helper。


## 技能邏輯批次 R
- 修正精靈 buff 數值：`RESIST_MAGIC`、`CLEAR_MIND`、`RESIST_ELEMENTAL`、`FIRE_WEAPON`、`BURNING_WEAPON`、`WIND_SHOT`、`STORM_EYE`、`STORM_SHOT`、`EARTH_SKIN`、`IRON_SKIN` 對齊 Java 數值。
- 修正 `ELEMENTAL_PROTECTION` 147：新增 `PlayerInfo.ElfAttr` 記憶體欄位，依 Java 精靈屬性只套用單一屬性抗性 +50，無屬性時施法失敗送 280。
- 修正 `FIRE_BLESS` 155：改為勇敢速度 1 並與其他勇敢速度類互斥，不再錯誤提供弓命中/弓傷害。
- 修正 `EARTH_BLESS` 159 與 `AQUA_PROTECTER` 160：依義維 Java 只保留狀態/圖示用途，不再錯誤調整 AC 或水抗。

## 技能邏輯批次 Q
- 補齊技能 108 `FINAL_BURN` / 會心一擊的 Java 特殊資源規則：HP 必須大於 100 才可施放，否則送 279 並取消。
- 成功施放技能 108 時不再照 YAML 固定扣 1 HP/MP，而是將 HP 扣到 100、MP 扣到 1，並同步發送 HP/MP 更新。

## 技能邏輯批次 P
- 補齊 PoisonSystem 的毒抵抗判定，玩家已有毒、`VENOM_RESIST` 104 或生命魔眼 6687 時不可再被施毒。
- 將 NPC 毒攻擊改為共用毒判定流程，傷害毒、沉默毒與麻痺毒都會尊重毒抵抗。
- 補齊 `ENCHANT_VENOM` 98 對玩家與 NPC 近戰命中後的 10% 傷害毒觸發，玩家/NPC 目標皆使用 3000ms、5 傷害的 Java 行為。
- 將 PvP 近戰與玩家對 NPC 近戰接入共用附加劇毒流程，避免兩套毒狀態設定分歧。
- 修正中毒視覺對標：毒性麻痺改用 yiwei 的 `S_Paralysis` subtype `0x02/0x03`，玩家與 NPC 初始物件封包會帶 poison status bit。

## 技能邏輯批次 O
- 修正黑暗妖精 buff 數值：`SHADOW_ARMOR` 改為 MR +5，`DRESS_MIGHTY` / `DRESS_DEXTERITY` 改為 STR/DEX +3，`DRESS_EVASION` 改為 ER/Dodge +18 且不修改 AC。
- 將 `BURNING_SPIRIT` 與 `DOUBLE_BREAK` 改為觸發型增傷旗標，不再提供固定 SP、Hit 或 Dmg 加成。
- 新增黑妖物理觸發傷害 helper，近戰與 PvP 近戰傷害會依燃燒鬥志與雙重破壞機率加成。
- 修正 `DARK_BLIND` 玩家與 NPC 目標皆依 Java 使用 66 睡眠效果，而不是保留 103 空狀態。

## 技能邏輯批次 N
- 補齊 `CALL_CLAN` / `RUN_CLAN` 血盟目標專用流程，依 Java 以角色名稱或目標 ID 找線上血盟成員，不再受一般 buff 目標的同地圖與距離限制。
- 修正 `CALL_CLAN` 對目標盟友設定 `PendingYesNoType=729` 與呼喚者 ID，並發送 729 同意視窗。
- 修正 `RUN_CLAN` 只允許傳送到目標地圖 `0/4/304` 且非攻城區，施法者目前地圖不可逃離時送 647，目標地點不合法時送 1192 並解除傳送鎖。
- 補齊 `SHOCK_STUN` 玩家目標雙手劍限制、暈眩套用與隨機秒數，並讓 NPC 目標共用雙手劍檢查。
- 驗證 `BOUNCE_ATTACK` 依 Java 只提供 Hit +6，重複施放不會疊加。

## 技能邏輯批次 M
- 補齊王族士氣範圍套用：`GLOWING_AURA`、`SHINING_AURA`、`BRAVE_AURA` 依 `TargetTo` 套用隊伍或血盟附近成員。
- 修正士氣 buff 數值：激勵士氣改為近戰命中/傷害 +5，鋼鐵士氣維持 AC -8，衝擊士氣改為機率增傷旗標。
- 修正 `SOLID_CARRIAGE` 需要盾牌或臂甲，效果改為 ER/Dodge +15，不再錯誤修改 AC。
- 補齊 `TRUE_TARGET` 113 狀態與 `S_TrueTarget` 封包，支援將文字標記送給血盟成員。
- 將 `BRAVE_AURA` 接入玩家物理近戰與遠程傷害流程，依 Java 以 33% 機率造成 1.5 倍傷害。

## 技能邏輯批次 L
- 補齊 `CURSE_BLIND` / `DARKNESS` 的致盲狀態邏輯，依 Java 使用 40 作為共同致盲 buff，並支援解除致盲封包。
- 修正 `REMOVE_CURSE` 解除毒、詛咒麻痺、麻痺/睡眠與致盲狀態。
- 修正 `CANCELLATION` 可對自身生效，清除可相消 buff 並保留不可相消 buff。
- 修正 `HASTE` / `SLOW` 互相抵銷行為：目標有相反速度狀態時只解除舊狀態並回復正常速度，不套用新狀態。

## 技能邏輯批次 K
- 補齊龍騎士鎖鏈劍普通近戰的弱點曝光階段狀態，對齊 Java `dk_dmgUp()` 的目標切換清除與 Lv1/Lv2/Lv3 推進規則。
- 新增 `S_PacketBoxDk` 封包建構與發送，支援弱點曝光階段顯示與清除。
- 補齊屠宰者吃弱點階段加傷：Lv1/Lv2/Lv3 分別加 `20/40/60 + FoeSlayerBonusDmg`，並在技能結束後清除弱點。

## 技能邏輯批次 J
- 補齊 `FOE_SLAYER`（187，屠宰者）專用流程：三段近戰攻擊、額外亂數傷害、7020/12119 技能音效與玩家/NPC 目標處理。
- 補齊屠宰者 `COPY_SHOCK_STUN`（508）機率暈眩，玩家送 Stun 操作鎖，NPC 設定 paralyzed debuff。
- 將 508 納入既有暈眩解除流程，避免 buff 清除或到期後殘留客戶端操作鎖。

## 技能邏輯批次 I
- 補齊 `FREEZING_BREATH`（194，寒冰噴吐）玩家命中後的凍結、灰色凍結色調、Freeze 操作鎖與 `81168` 凍結特效。
- 補齊寒冰噴吐對隱身目標的揭示邏輯，對齊 Java `detection(_player)` 行為。
- 將 194 納入 NPC 凍結 debuff 與玩家凍結解除流程，避免 buff 到期、GM 清除或取消時殘留客戶端控制狀態。

## 技能系統批次 H
- 修正 `RESIST_FEAR`（188）玩家效果，依 Java 改為閃避 -5，不再錯誤調整 STR/INT。
- 補齊 `THUNDER_GRAB`（192）玩家目標束縛效果，依技能機率資料與等級差判定成功後套用 192 束縛狀態，並補上 Bind 解除封包處理。

## 技能系統批次 G
- 補齊 `BONE_BREAK`（208）玩家目標控制效果，依義維版幻術師設定計算命中率，成功後套用 208 麻痺狀態並廣播 13119 技能音效。
- 修正 `JOY_OF_PAIN`（218）行為：施放時只給施法者 16 秒一次性狀態，不再直接傷害目標；玩家目標受傷前依目標既有失血量反傷施法者，反傷後移除 218。
- 修正 `magic.lua` 的 218 公式，避免 Lua 路徑產生非 Java 的直接傷害。

## 技能系統批次 F
- 補齊幻術師玩家目標路徑：`CONFUSION`（202）命中後沉默、`MIND_BREAK`（207）改用 Java `SP * 3.8` 傷害並扣 MP、`PHANTASM`（212）改為睡眠、`ARM_BREAKER`（213）命中後揭示隱身目標。
- 修正 `Mind Break` Lua 公式與 NPC MP 扣除行為，目標 MP 不足時扣到 0。
- 補上批次 F 測試，驗證玩家目標的傷害、MP 扣除、沉默、睡眠與揭隱行為。

## GM 指令
- 新增 `.clearbuff` — 全清狀態：所有 buff（含不可取消的覺醒類）+ 中毒 + 詛咒 + 麻痺/凍結/睡眠/隱身/絕對屏障 + 沉默
- 新增 `system.SkillSystem.GMClearAllStatuses()`（暴露於 `SkillManager` 介面），含完整客戶端通知封包，補足 `ClearAllBuffsOnDeath()` 在非死亡情境下會留下 freeze/paralysis/invis 殘影、HasteTicks/WisdomSP/AbsoluteBarrier 等旗標未清的問題

## 技能封包
- 擴充 `SkillRequest`，保留技能封包的目標座標、地圖 ID、書籤 ID、召喚選擇值、目標角色名與文字內容。
- 補齊 `C_USE_SPELL` 對照 Java `C_UseSkill.java` 的特殊讀取路徑：指定/集體傳送、精準目標、呼喚/援護盟友、火牢/生命之泉、召喚術。
- 新增 `HandleUseSpell` 單元測試，覆蓋特殊技能封包欄位解析。
- 補上 `SkillSystem` 對 `SkillRequest.TargetName` 的目標玩家解析，讓呼喚盟友/援護盟友可由角色名轉成目標 CharID。
- 新增 `SkillSystem` 目標解析單元測試。

## 技能系統
- 補齊壞物術（skill 27）玩家目標邏輯：依 Java `INT / 3` 隨機增加目標已裝備武器耐久，並發送訊息 268 與背包更新。
- 補齊壞物術 NPC 目標邏輯：新增 `NpcInfo.WeaponBroken`，NPC 物理攻擊傷害套用減半，魔法相消術會清除狀態。
- 新增壞物術耐久、NPC 傷害減半與相消清除狀態的單元測試。

## 技能系統
- 建立技能批次 A 實作計畫，批次範圍為法師基礎治癒與基礎攻擊技能。
- 補齊攻擊型技能對玩家目標的共用傷害流程，讓光箭、冰箭、風刃、火箭、地獄之牙、極光雷電、燃燒的火球、地裂術、烈炎術、龍捲風、暴風雪等同類技能可共用同一路徑。
- 補齊自身中心範圍攻擊對附近玩家的傷害流程，保留既有 NPC 範圍傷害。
- 新增 A 批次技能測試，覆蓋玩家目標攻擊、範圍攻擊、單體治癒封頂與全部治癒範圍。

## 技能補齊：批次 B
- 補齊 `SILENCE`（64）buff 與 `PlayerInfo.Silenced` 的套用/解除同步，讓沉默狀態確實阻止施法並能在 buff 移除時恢復。
- 補齊 `IMMUNE_TO_HARM`（68）聖結界共用減傷邏輯，玩家魔法、PvP 近戰/遠程、NPC 近戰/遠程/技能對玩家傷害會依 Java 預設倍率減半。
- 補上批次 B 測試，覆蓋沉默狀態同步、聖結界玩家魔法減傷、共用減傷 helper、魔法屏障抵消玩家攻擊魔法。

## 技能補齊：批次 C
- 補齊 `TELEPORT`（5）與 `MASS_TELEPORT`（69）使用 `SkillRequest.BookmarkID` 執行書籤傳送，避免誤用一般目標 ID。
- 補齊 `SUMMON_MONSTER`（51）使用 `SkillRequest.SummonID` 委派召喚系統，對齊 Java `C_UseSkill.java` 額外欄位解析。
- 新增批次 C 測試，覆蓋指定傳送書籤 ID 與召喚術 summon ID 的 System 委派行為。

## 技能補齊：批次 D
- 新增地面技能效果模型與 `GroundEffectSystem`，支援地面效果生成、AOI 顯示、過期移除與 tick 處理，且不占用實體碰撞格。
- 補齊 `FIRE_WALL`（58）火牢依目標座標生成 81157 地面效果，並定期對 1 格內目標造成火屬性傷害。
- 補齊 `LIFE_STREAM`（63）生命之泉 81169 地面效果，玩家在 4 格內時 HP 回復額外 +3。
- 新增批次 D 與 world 地面效果測試，覆蓋生命週期、重複座標查詢、火牢生成、火牢傷害與生命之泉 HPR 加成。

## 技能補齊：批次 E
- 補齊幻術師四種立方 `CUBE_IGNITION`（205）、`CUBE_QUAKE`（210）、`CUBE_SHOCK`（215）、`CUBE_BALANCE`（220）的地面效果生成與 3 格內同型立方拒絕邏輯。
- 擴充 `GroundEffectSystem` 處理立方盟友/敵方狀態、燃燒傷害、地裂束縛、衝擊狀態與和諧 MP/HP tick。
- 新增批次 E 測試，覆蓋立方生成、重複施放不消耗材料、燃燒敵我判定與和諧 tick 效果。
# 下次變更摘要

## 技能批次 U
- 修正 `TELEPORT_TO_MATHER` 131：依 Java 行為回母樹座標，而非範圍復活。
- 修正 `CALL_OF_NATURE` 165：玩家死亡目標改送復活同意，屍體格有活人時以訊息 592 拒絕。
- 修正 `EARTH_BIND` 157：玩家凍結持續時間改為 Java 1-12 秒隨機，並納入玩家 debuff MR 判定。
- 修正精靈 `EXOTIC_VITALIZE` 169 / `ADDITIONAL_FIRE` 176：負重狀態下允許 HP/MP 回復。
- 新增批次 U 測試與 `docs/技能批次U實作計畫.md`。

## 技能批次 V
- 補齊 `CALL_OF_NATURE` 165 對死亡 NPC 的滿 HP 復活與狀態重置。
- 補齊 `CALL_OF_NATURE` 165 對死亡寵物的滿 HP 復活、仇恨清除、休息狀態與 HP 條更新。
- 補齊寵物屍體格有活人時以訊息 592 拒絕自然呼喚。
- 新增批次 V 測試與 `docs/技能批次V實作計畫.md`。

## 技能批次 W
- 新增 NPC YAML/template/runtime `cant_resurrect` 欄位。
- 讓 `CALL_OF_NATURE` 165 遵守不可復活 NPC 限制。
- 一般 spawn 與部分動態 NPC 生成會從 template 帶入 `CantResurrect`。
- 新增批次 W 測試與 `docs/技能批次W實作計畫.md`。

## 技能批次 X
- 補齊 `RESURRECTION` 61 / `GREATER_RESURRECTION` 75 對死亡 NPC 的 1/4 HP 復活。
- 補齊 61/75 對死亡寵物的 1/4 HP 復活、仇恨清除、休息狀態與 HP 條更新。
- 61/75 共用 `cant_resurrect` 與寵物屍體格活人阻擋限制。
- 新增批次 X 測試與 `docs/技能批次X實作計畫.md`。

## 技能批次 Y
- 補齊復活卷軸 `40089` / `140089` 的死亡玩家、NPC、寵物目標處理。
- 復活卷軸玩家目標改送 321/322 同意視窗，NPC/寵物直接以 1/4 HP 復活。
- 復活卷軸共用 `cant_resurrect`、塔不可復活、寵物屍體格活人阻擋與地圖不可復活限制。
- 新增批次 Y 測試與 `docs/技能批次Y實作計畫.md`。

## 技能批次 Z
- 新增 `S_OPCODE_RESURRECTION = 85` 與 `handler.BuildResurrection()`。
- 玩家接受 321/322 復活同意後，改送復活音效、`S_Resurrection` 與 `S_CharVisualUpdate` 給自己與附近玩家。
- 新增批次 Z 測試與 `docs/技能批次Z實作計畫.md`。

## 技能批次 AA
- 對標義維版死亡墓碑並加上 `gameplay.enable_tomb_effect` 開關；3.80C 原生客戶端沒有墓碑圖檔，預設不生成墓碑。
- 開關啟用後，玩家死亡會生成 NPC 86126 / GFX 13600 墓碑 GroundEffect，存在 300 秒，名稱與 lawful 取自死亡玩家。
- 墓碑生成後對附近玩家送 `S_NPCPack_Eff` 相容物件封包與 action 4。
- 復活同意、死亡回村、斷線與 GroundEffect 自然過期都會清除墓碑，清除時送 action 8 與 `S_RemoveObject`。
- 新增批次 AA 測試與 `docs/技能批次AA實作計畫.md`。
