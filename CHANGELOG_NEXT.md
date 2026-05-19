## MISS-P1-010 火神煉化系統（精煉 + 合成）完整實作（2026-05-20）

3.80C 客戶端原生 SmithUI（type 48 精煉視窗 / type 49 合成視窗）已完整對接。整套流程從 NPC 對話、UI 開窗、拖放素材、扣材料、給成品到日誌診斷全部走通並由使用者實機驗證通過。

---

### 🎮 玩家操作手冊

#### A. 火神精煉（type 48：把裝備分解成火神結晶體）

1. **找到 NPC**：對話火神煉化工匠系列 NPC（如 NPC 111414）。
2. **點 HTML 按鈕**：選擇「物品精煉（itemresolve）」。
3. **開啟精煉視窗**：客戶端跳出 `MakeCrystal_Window`。
4. **拖入裝備**：把要分解的 +6 以上武器/防具拖入視窗。
   - 條件：物品必須在白名單（itemdesc_id 在 CrystalInfo.tbl 147 條中）。
   - 限制：未裝備、+6 以上才有結晶值（+0~+5 會被拒「此物品無法精煉。」）。
5. **確認**：客戶端送出 `C_PledgeContent data=13` → server 扣裝備、給予 41246 火神結晶體（按 `fire_crystal_list.yaml` 對應等級數量）。
   - 範例：+7 大馬士革刀 → 26 個火神結晶體（已實測）。
   - 系統訊息：`\f2獲得 X 個火神結晶體。`

#### B. 火神合成（type 49：用結晶體 + 契約 + 槌子合成裝備）

1. **找到 NPC**：對話同一系列火神工匠 NPC。
2. **點 HTML 按鈕**：選擇「物品合成（itemtransform）」。
3. **開啟合成視窗**：客戶端跳出 `MakeItem_Window`，顯示 11 條配方（紅影雙刀、風刃、死騎劍、冰皇魔杖、寒冰鎖鏈劍、寒冰奇古獸、沙耶弓、蚩尤鎧甲、黑長者長袍、底比斯雙戒）。
4. **選配方**：UI 顯示需要的火神結晶體（80029）、火神契約（80028）數量與基礎成功率。
5. **拖入火神之槌（80027，選填）**：放進 plus 槽，每個 +1% 成功率。
   - 拖入 10 個槌子 → 成功率 +10%、扣 10 個（已實測修正）。
   - 客戶端封包 `[D npcObjID][H actionID][D plusObjID][D plusCount]`。
6. **按合成按鈕**：客戶端送出 `C_PledgeContent data=14` → server：
   - 扣材料：結晶體 + 契約 + 火神之槌（不論成敗都扣）。
   - 擲骰判定：`rand(100) < successRate` → 成功給成品；否則「\f3合成失敗。」。
   - 成功訊息：`\f2成功鑄造。` + 成品入背包。

#### 各配方基礎成功率（MakeInfo.tbl col4）

| actionID | 配方 | 結晶 | 契約 | 基礎率 |
|---|---|---|---|---|
| 1752 | 紅影雙刀 | 400 | 20 | 1% |
| 1820 | 風刃短劍 | 400 | 20 | 1% |
| 992 | 死亡騎士的烈炎之劍 | 200 | 10 | 1% |
| 1094 | 冰之女王魔杖 | 150 | 5 | 1% |
| 12527 | 寒冰鎖鏈劍 | 125 | 5 | 3% |
| 12528 | 寒冰奇古獸 | 125 | 5 | 3% |
| 3551 | 蚩尤鎧甲 | 100 | 5 | 8% |
| 568 | 沙耶之弓 | 100 | 5 | 10% |
| 978 | 黑長者長袍 | 100 | 5 | 10% |
| 3268 | 底比斯賀洛斯戒指 | 75 | 5 | 10% |
| 3269 | 底比斯阿努比斯戒指 | 75 | 5 | 10% |

---

### 🛠 實作摘要（按解決順序）

#### 1. 開窗封包（方向 B：380 S_EquipmentWindow type 48/49）

- 380 `S_EquipmentWindow(type, value)`：`writeC(S_OPCODE_CHARRESET=64) + writeC(type) + writeD(npcObjID) + writeC(0x95) + writeC(0x19)`。尾碼 0x95/0x19 取代舊版 S_Refine 的 0xE3/0x92。
- `handler/npcaction.go::sendRefineUI(sess, npcObjID, refineType)` 採此格式；`case "itemresolve"` → type=48、`case "itemtransform"` → type=49。
- 舊 `sendFireSmithSellList` / `sendCraftItemBlend` 保留為 NPC 111414 的回退路徑（`case "request firecrystal"` / `case "request craft itemblend"`）。

#### 2. 拖放白名單（itemdesc_id 修正）

- 拖放被客戶端 silently reject 的根因：server 送 `S_AddItem` / `S_ADD_INVENTORY_BATCH` 時 `itemdesc_id` 填 0，導致 SmithUI drag-accept filter（檢查 `item+0xb2` itemdesc_id 與 CrystalInfo.tbl 147 條白名單比對）匹配失敗。
- `handler/shop.go::sendAddItem` + `sendInvList` 改為優先使用 YAML `itemInfo.ItemDescID`，fallback 才走 hardcoded 6-item 表。
- 資料層補齊：`scripts/patch_itemdesc.py` / `scripts/diff_itemdesc.py` 比對 yiwei `weapon`/`armor`/`etcitem` SQL 把 YAML `itemdesc_id: 0` 但 yiwei 有非零值的條目補上（etcitem 16 個：40318=166、40319=569、49147–49158 14 個）。
- 結果：itemdesc_id 235（大馬士革刀）等 147 個物品的拖放 filter 現可正確通過。

#### 3. 精煉處理（C_PledgeContent data=13）

- `handler/refine.go::handleRefineResolve` 從 git 歷史還原並對齊 380：`[D npcObjID][D itemObjID][D assistItemObjID]` → 查 `FireCrystals` 表計算結晶數量 → 移除原物品 → 給予 41246 火神結晶體。
- 每個拒絕分支加 `deps.Log.Warn` 結構化記錄（NPC 不存在、超出範圍、查無物品、已裝備、FireCrystals 未載入、類別不可精煉、無對應條目、crystalCount<=0），方便除錯。
- 「+0~+5 拒絕」為正常設計（`fire_crystal_list.yaml` 大馬士革刀 `enchant_levels: [0,0,0,0,0,0,7,26,90,...]`，+6 起才有值），非 bug。
- **實測**：+7 大馬士革刀 → 26 個火神結晶體 ✓。

#### 4. 合成處理（C_PledgeContent data=14）

**Recipe YAML（`firesmith_recipe_list.yaml`）**

- 11 筆配方由 `MakeInfo.tbl`（`d:/l1jclient-rustx64/target/text-search/MakeInfo.tbl`）抽出：
  - `col1` = actionID（封包 H 欄位）
  - `col2` = 火神結晶體 80029 需要數量
  - `col3` = 火神契約 80028 需要數量
  - `col4` = **基礎成功率（%）** UI 顯示值（先前誤判為 catalyst 數量，使用者反編譯後修正）
  - `col5` = UI display key
- `catalyst_count` 全部歸 0（MakeInfo.tbl 無此欄；5 個 catalyst 槽位的實際素材表待後續補）。

**新增成品武器**

- `weapon_list.yaml` 補 100366 寒冰鎖鏈劍（chainsword/龍騎士、dmg 23/20）+ 100369 寒冰奇古獸（kiringku/幻術師、dmg 25/25），資料源自 yiwei 380 SQL；`firesmith_recipe_list.yaml` actionID 12527/12528 `new_item_id` 連上。

**封包格式**（reverse-32-mcp attach 3.80C 客戶端、RVA 0x29774D 反編譯確認）

```
mode 2 分支 push 順序（ccdhdd）：
  [+0x1cc] D npcObjID
  [+0x1dc] H actionID
  [+0x1e4] D plusItemObjID    ← 玩家拖入的火神之槌/淚 objID
  [+0x20c] D plusItemCount    ← 玩家拖入的數量（非另一個 objID！）
```

- 之前誤把第 4 個 D 當作 mainItemObjID，10 個 hammer → server 只認 objID 不認 count → 固定扣 1（使用者回報「我帶入十個火神之槌結果他只扣一個」）。
- `handler/refine.go::handleRefineTransform` 第 4 個 D 改讀為 `plusItemCount`。

**Server 邏輯（`system/npc_service.go::FireSmithCraft`）**

- 簽章：`FireSmithCraft(recipe, plusItemObjID int32, plusItemCount int32)`。
- 流程：材料檢查（結晶/契約/catalyst）→ plus 物品檢查（objID 找實體、count clamp 到 `plus.Count` 防作弊）→ 計算成功率 `recipe.SuccessRate + bonus × use` → 確認可放成品 → 扣全部材料（含 plus 整批）→ 擲骰 → 成功給成品 / 失敗送訊息。
- 防作弊：client 送的 plusItemCount 若超過實際庫存，clamp 到 `plus.Count`。
- 介面：`handler/context.go::NpcServiceManager` + 兩個 test stub 同步收斂為固定兩個 int32 參數。
- 診斷 log：`plusStackCount` / `plusUseCount` / `count` 三欄方便驗證實際扣除數。

#### 5. 餘留（不在本步處理）

- catalyst 5 個槽位實際素材表（catalyst_item_id / 各 slot 對應物品）尚未還原；目前 `CatalystItemID>0 && CatalystCount>0` 雙條件守衛，全部歸 0 等於跳過檢查。後續可由 server SQL / 反編譯 SmithUI catalyst drag-accept filter 還原。
- itemdesc_id 兩邊都非零但值不同的 4 個 etcitem（40100、240010、240100、49147），留作後續決定要不要切換到 yiwei 值。

---

### ✅ 驗證

- `cd server && go build ./...` ✓
- `go vet ./...` ✓
- `go test ./internal/handler/... ./internal/system/...` 全綠（regression-free）
- 實機：MakeCrystal_Window 開窗 ✓、拖入 +7 大馬士革刀分解 → 26 個火神結晶體 ✓、MakeItem_Window 開窗 ✓、10 個 hammer 一次扣 10 個 + 成功率 +10% ✓

## MISS-P1-009 結婚（C_Propose 行為對齊）

- **盤點結論**：Go `handlePropose` 已實作大致流程（戒指、教堂、Y/N），但與 Java `C_Propose.java` mode=0 比對發現 4 個真實行為差異：
  1. **缺亡靈/死亡守衛**：Java `if (pc.isGhost()) return;` 在第一行就擋；Go 沒擋。
  2. **缺同性結婚拒絕**：Java `if (pc.get_sex() == target.get_sex())` → `S_ServerMessage(661)`；Go 完全沒檢查 → 可用同性求婚。
  3. **等級總和 <50 的訊息錯誤**：Java 用 `S_SystemMessage("雙方總等級未等於50以上。")`；Go 錯用 `SendServerMessage(sess, 661)`（661 是「必須在教堂」的訊息）。
  4. **教堂檢查時機與訊息錯誤**：Java 把教堂檢查放到所有檢查最後，並送 `S_SystemMessage("必須在教堂中才能進行。")`；Go 把教堂放在第一個檢查並送 `S_ServerMessage(661)`，且只檢查求婚者（對象不在教堂時靜默 return）。
- `handler/marriage.go::handlePropose` 重新排序並對齊 Java 流程：
  1. `if player.Dead { return }` — Java `isGhost()` 等效。
  2. `player.PartnerID != 0` → 658。
  3. `findNearbyPlayer` → 找不到送 93。
  4. `target.PartnerID != 0` → 658。
  5. **新增** `player.Sex == target.Sex` → 661。
  6. **改正** `int(player.Level)+int(target.Level) < 50` → `SendSystemMessage("雙方總等級未等於50以上。")`（非 661）。
  7. `!hasRing(player)` → 659；`!hasRing(target)` → 660。
  8. **改正** `!inChurch(player) || !inChurch(target)` → `SendSystemMessage("必須在教堂中才能進行。")`（移到最後 + 同時檢查雙方 + 改用 system message）。
  9. 設定 `target.PendingYesNoType=654 / PendingYesNoData=player.CharID` + `sendYesNoDialog(target, 654, player.Name)`。
- 新增 `handler/marriage_test.go` 5 個測試：
  - `TestHandleMarriageProposeSendsYesNoToTarget` — 成功路徑：對象 PendingYesNoType=654、PendingYesNoData=CharID、求婚者不收到 661。
  - `TestHandleMarriageProposeRejectsSameSex` — 同性求婚送 661、對象不收 Y/N。
  - `TestHandleMarriageProposeLevelSumUsesSystemMessage` — 等級總和<50 送 S_SystemMessage（type=9），不送 661。
  - `TestHandleMarriageProposeOutOfChurchUsesSystemMessage` — 教堂外送 S_SystemMessage，不送 661。
  - `TestHandleMarriageProposeWhenDeadSilent` — 死亡時靜默 return（無 658/659/660/661）。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/handler -run TestHandleMarriagePropose` 5/5 ✅、`go test ./internal/...` 全 package 全綠（regression-free）。
- 停損：Java 的 `pc.getQuest().get_step(L1PcQuest.QUEST_MARRY) == 0` 額外閘門（結婚任務未開放時不可求婚）需要 Quest 系統的 QUEST_MARRY step；Go 目前缺結婚任務模型，此 gate 不在本步補。離婚（mode=1）與接受/拒絕回呼（`AcceptProposal` / `ConfirmDivorce`）的 Java 對齊待後續子項。

## MISS-P1-008 住宅售屋流程（agsell）

- **盤點結論**：Go 已有完整 NPC 管家動作（name / tel0-3 / upgrade / hall），唯獨 `agsell`（售屋）被標註「與拍賣系統整合，暫不實作」直接 return false，導致玩家在自家管家點「出售」毫無回應。Java `C_NPCAction.sellHouse + S_SellHouse + C_Amount("agsell")` 完整流程為：條件驗證 → 送 `S_SellHouse` 開啟價格輸入框 → 玩家輸入後 C_Amount 回覆 `agsell {houseId}` → 寫入 AuctionBoardTable。
- 條件鏈（對齊 Java）：玩家在血盟 → 血盟有小屋（`HasHouse != 0`）→ keeperID 對應該小屋 → 君主職業（`ClassType == 0`）→ 盟主（`CharID == clan.LeaderID`）→ 尚未上架。
- `world/state.go`：`PlayerInfo` 新增 `PendingSellHouseID int32`（與 `PendingAuctionHouseID` / `PendingInnNpcObjID` 同層級的待處理槽）。
- `handler/house.go`：
  - `handleHousekeeperAction` switch 補上 `case "agsell"` → 委派 `handleHouseSell`。
  - `handleHouseSell` 跑完整 Java 條件鏈；通過後送 `S_SellHouse` 並設定 `PendingSellHouseID`；已上架時送 `"agonsale"` hypertext（對齊 Java `htmlid = "agonsale"`）。
  - `sendSellHouse(sess, npcObjID, houseID)` 建 `S_OPCODE_INPUTAMOUNT`：npcObjID、`writeD(0)`、`writeD(100000)`、`writeD(100000)`、`writeD(2000000000)`、`writeH(0)`、`writeS("agsell")`、`writeS("agsell {houseID}")`（逐欄對齊 Java `S_SellHouse.java`）。
  - `HandleSellHouseAmount(sess, r, player, deps)` 解析 C_Amount 回覆：`[D npcObjID][D amount][C unknown][S "agsell {houseId}"]`；價格 clamp `[100_000, 2_000_000_000]`；再驗一次盟主/小屋狀態（防 send→reply 間血盟異動）；查 HouseRepo/HouseTable 補上 `HouseName / HouseArea / Location`，建 `AuctionEntry`（deadline = now+5 days、minute/second=0、`OldOwner = player.Name`）後委派 `deps.Auction.CreateSale`。
  - 新增輔助 `sellHouseDeadline(now)`、`lookupHouseAuctionFields`、`houseTownName`、`houseArea`。
- `handler/auction.go`：`AuctionManager` 介面新增 `CreateSale(entry *persist.AuctionEntry) bool`。
- `system/auction_sys.go`：`AuctionSystem.CreateSale` 實作（WAL 保護：先 `InsertAuction` 再加入記憶體快取；同 houseID 已存在則拒絕）。
- `handler/polymorph.go::HandleHypertextInputResult`：在 `PendingAuctionHouseID` 路由之後追加 `PendingSellHouseID > 0` 路由到 `HandleSellHouseAmount`。
- 新增 `handler/house_sell_test.go` 7 個測試：
  - `TestHandleHouseSellSendsS_SellHouseWhenValid` — 成功路徑：通過所有檢查、`PendingSellHouseID` 設值、收到含 `agsell` htmlid 的 `S_OPCODE_INPUTAMOUNT`。
  - `TestHandleHouseSellRejectsNonLeader` — 非盟主拒絕，無 S_SellHouse。
  - `TestHandleHouseSellRejectsNonCrown` — 非君主職業拒絕。
  - `TestHandleHouseSellWrongKeeperNoOp` — keeperID 對不上自家小屋時靜默 no-op。
  - `TestHandleHouseSellAlreadyOnSaleSendsAgonsale` — 已上架時送 `agonsale` hypertext，不送 S_SellHouse。
  - `TestHandleSellHouseAmountCreatesSale` — C_Amount 回覆觸發 `CreateSale`：HouseID / OldOwner / Price / Location（依 houseID 區段=奇岩）/ HouseArea（X1..Y2 計算=121）/ Deadline ≈ now+5 day（hour-truncated）。
  - `TestHandleSellHouseAmountRejectsOutOfRange` — 價格 < 100,000 不上架。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/handler -run 'TestHandleHouseSell|TestHandleSellHouseAmount'` 7/7 ✅、`go test ./internal/...` 全 package 全綠（regression-free）。
- 停損：Java agsell 還寫了 `house.setOnSale(true)` + `house.setPurchaseBasement(true)` 兩個 DB 旗標，但 Go 的「on sale」狀態本就是用「auction entry 存在與否」推導（與 `clan.HasHouse != 0` 同樣 derived state），不需要獨立旗標；`setPurchaseBasement(true)` 的 Java 語意（Chinese comment 寫「設置為未購入」但代碼是 `true`）含糊，可能是 Java 設計缺陷或翻譯誤植，不照搬以免引入既有 bug。售屋成交、結算所有權轉移、原屋主退金幣由既有 `AuctionSystem.settleExpired` 自動處理（已對齊 Java 四種結算情境）。

## MISS-P1-007 船運航線（C_Ship 行為對齊）

- **盤點結論**：Go 已有 `shipDocks`（碼頭座標 → 航線群組 + 船票）+ `shipGroupA/B Windows`（16 個時間窗口）+ `shipTicketByCurrentMap`（船地圖 → 船票）+ `CheckShipDock` 完整覆蓋義维 Java `DungeonTable.dg()` SHIP_FOR_* 6 種類型；`HandleEnterShip`（opcode 231）對應 Java `C_Ship.java`。實際 Java 行為差異只有兩處：
  - Java `C_Ship` 在 `consumeItem` 後送 `S_OwnCharPack(pc)` 再呼叫 `L1Teleport.teleport`，Go 漏掉了傳送前的 S_OwnCharPack。
  - Java `L1Teleport.teleport(pc, locX, locY, mapId, 0, false)` 第 5 參數 `heading=0`，Go 硬寫 `5`。
- `server/internal/handler/ship.go::HandleEnterShip` 修正：在 `cancelTradeIfActive` 之後、`teleportPlayer` 之前插入 `sendOwnCharPackPlayer(sess, player)`；`teleportPlayer` 的 heading 參數從 `5` 改為 `0`。
- 新增 `server/internal/handler/ship_test.go` 2 個測試：
  - `TestHandleEnterShipMatchesJava` — 持票 + 正確 map 下，驗證消耗 1 張船票、玩家 heading=0、MapID/X/Y 跳到目的地、自己 session 收到 ≥2 次 `S_PUT_OBJECT`（pre-teleport + teleportPlayer 內部的 own-pack）。
  - `TestHandleEnterShipWithoutTicketSilentlyReturns` — 無票時 Java 走 default 不做事，Go 對齊：不消耗、玩家留原地。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/handler -run TestHandleEnterShip` 2/2 ✅、`go test ./internal/handler ./internal/system ./internal/world` 全綠（regression-free）。
- 停損：未開發系統排序盤點所稱「航線移動動畫」對應 l1j_fly 的 `Npc_Ship$Work` — 該類是 `Chapter02R` 副本任務專用的腳本化船 NPC（透過 `L1QuestUser` + `NpcWorkMove` + `Chapter02R.shipReturnStep()` 沿 `Point[]` 路徑移動），屬副本任務內容；本專案的官方參考義维 Java 並無此類，歸 MISS-P2-019（Chapter02R 副本內容）後續，不在 P1 範圍。一般船運（時刻表 / 船票 / 上下船 / 傳送 / 訊息）至此已完全對齊義维 Java。

## MISS-P1-006 魔法娃娃完整效果

- **盤點結論**：Java 共 45 個 `Doll_*` 類別，Go YAML 已宣告 12 種 power type 但 `system/doll.go` 的 `switch p.Type` 只覆蓋 8 種；`dmg_reduction` / `weight` / `hp_regen_tick` / `mp_regen_tick` 四種 silently fall-through，YAML 寫了也沒效果。本步補齊這四種真實 Java 行為差異（CLAUDE.md 對齊深度停損標準的「Java 與 Go 行為實際不同 → 必須修 Go」類）。
- `server/internal/world/doll.go` 的 `DollInfo` 新增 8 個欄位：
  - `BonusDmgReduce int16` — 對應 Java `Doll_DmgDown` 受傷減免（目前接到 `EquipBonuses.DmgReduction`，戰鬥讀取為後續任務；娃娃端先完成跟蹤）。
  - `BonusWeight int16` — 對應 Java `Doll_Weight` 額外負重上限。
  - `RegenHPAmount / RegenHPInterval / RegenHPCounter int` — 對應 Java `DollHprTimer`，每 Interval ticks 觸發回主人 Amount HP。
  - `RegenMPAmount / RegenMPInterval / RegenMPCounter int` — 對應 Java `DollMprTimer`，MP 版本。
- `server/internal/world/state.go` 的 `PlayerInfo` 新增 `WeightBonus int32`；`server/internal/world/inventory.go` 的 `PlayerMaxWeight(p)` 將 `WeightBonus` 累加到回傳值（既有 14/218 buff `+180` 不變）。
- `server/internal/system/doll.go` 的 `UseDoll` switch 補齊四個 case：`dmg_reduction` → `BonusDmgReduce`、`weight` → `BonusWeight`、`hp_regen_tick` → `RegenHPAmount` + `RegenHPInterval = param × 5 ticks`、`mp_regen_tick` 同上。`applyDollBonuses` / `removeDollBonuses` 對稱加入 `player.WeightBonus ± BonusWeight` 與 `player.EquipBonuses.DmgReduction ± BonusDmgReduce`。
- `server/internal/system/companion_ai.go` 的 `tickDolls` 在 timer 倒數之後呼叫新 helper `tickDollRegen(doll, master)`：Counter 遞增 → 達 Interval 即歸零、套用 Amount、上限 MaxHP/MaxMP、發送 `S_HitPoint` / `S_ManaPoint`；主人死亡時暫停回復（對齊 Java）。Amount=0 或 Interval=0 短路不增 Counter。
- 新增 `server/internal/system/doll_bonus_test.go` 7 個測試：apply/remove Weight+DmgReduction、PlayerMaxWeight 含 WeightBonus、HP 回復 Interval=5 ticks 觸發點、HP 截斷至 MaxHP、MP 對稱、Amount/Interval=0 短路、主人死亡暫停。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/system ./internal/handler ./internal/world` 全綠（regression-free）。
- 停損：Java 45 個 Doll_* 中剩餘的 `Doll_X2`（雙倍掉落 / 經驗）、`Doll_PVPDmg`、`Doll_Speed`（移動速度）、`Doll_AddSp` (加魔成長率) 等高階道具因牽涉 NPC 死亡分配 / PVP 戰鬥流程 / 玩家移動公式（屬 P2 範圍），不在本步處理；YAML 端亦尚未引入這些 power type。`Doll_DmgDown` 戰鬥端讀取 `EquipBonuses.DmgReduction` 屬另一條對齊路徑（item_power 已預備此欄位），與本步並列為「先完成數值跟蹤、戰鬥讀取另算」。

## MISS-P1-005 物品強化加成（L1ItemPower）

- **盤點結論**：Go `equip.go` 只算基礎模板加成 + 武器強化通用 HitMod/DmgMod，**沒有 Java `L1ItemPower` 的 per-item-ID enchant scaling**（如「抗魔法頭盔」每強化 +1 額外 +1 MR、巫妖斗篷 +3 起每階 +1 SP 等共 50+ 物品的特殊規則）。本步覆蓋簡單線性 + 門檻線性物品（≈36 條規則），延後 step-function 物品（體力臂甲/法師臂甲/守護臂甲/火神武器）與 X2/PVPDMG 維度（EquipStats 無對應欄位）。
- 新增 `server/internal/data/item_power.go`：`ItemPowerRule { ItemID / Stat / PerEnchant / MinEnchant / EnchantOffset / Note }` + `ItemPowerTable` + `LoadItemPowerTable(path)`。支援六個 stat dimension（MR / MPR / SP / HIT / HP / DMG_REDUCE）對應 EquipStats 的 MDef / AddMPR / AddSP / HitMod / AddHP / DmgReduction。
- 新增 `server/data/yaml/item_power.yaml`：22 條 MR + 2 條 MPR + 6 條 SP + 1 條 HIT + 2 條 DMG_REDUCE = 33 條規則覆蓋簡單線性與門檻線性物品（巫妖斗篷 +3 起、幻象眼魔 +8 起、馬昆斯斗篷 +7 起、激怒手套 +5 起、石製手套 +7 起）。馬昆斯斗篷同 item_id 同時有 MR+SP 兩條規則。
- `handler.Deps` 新增 `ItemPowers *data.ItemPowerTable` 欄位；`system/equip.go` 的 `applyEquipStats / calcEquipStats` 加入 `itemPowers` 參數。
- 新增 `applyItemPowerBonuses(stats, itemPowers, itemID, enchant)` 輔助函式（抽出為獨立函式以方便回歸測試），在 `calcEquipStats` 內每件裝備計算完基礎加成後呼叫，依 `rule.Bonus(enchant)` 累加至對應 stats 欄位。
- `main.go` 新增 `LoadItemPowerTable("data/yaml/item_power.yaml")` 並 `printStat("物品強化加成", count)`，掛入 `Deps.ItemPowers`。
- 新增 `data/item_power_test.go` 9 個測試：純線性、門檻+offset（巫妖斗篷）、多倍率（混沌斗篷 ×3）、nil rule、合法 YAML 載入、同 item 多 stat、4 種無效規則拒絕、Get 未登錄 + nil table 安全。
- 新增 `system/equip_item_power_test.go` 5 個測試：線性 MR、6 個 stat 維度路由到正確 EquipStats 欄位、同 item 多 stat 同時生效、門檻未達不加分、nil table/nil stats/未登錄 item 安全。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/data -run TestItemPower` 9/9 ✅、`go test ./internal/system -run TestApplyItemPowerBonuses` 5/5 ✅、`go test ./internal/...` 全綠（regression-free）、server 啟動載入「物品強化加成 36」。
- 停損：step-function 物品（體力臂甲 +5/+7/+9 三段、法師臂甲類同、守護臂甲類同、火神武器 8/9/10/11/12 階）+ X2/PVPDMG 維度（需要 EquipStats 新欄位 + 戰鬥公式接入）歸 P1-005 後續或下放至 P2。

## MISS-P0-004 DungeonTable / DungeonRTable 完整規則

- **盤點結論**：Go 已有 `PortalTable`（對應 Java `DungeonTable`）+ `RandomPortalTable`（對應 Java `DungeonRTable`）+ `CheckShipDock` 完整覆蓋 6 種 SHIP_FOR 類型 + 16 個時間窗口 + 6 種船票物品 ID。`HandleMove` 已串接固定/隨機傳送門查詢。**唯一真實 Java 行為缺口：傳送觸發時缺 `ABSOLUTE_BARRIER` 2 秒無敵 + `stopHpRegeneration/stopMpRegeneration` 對齊**。
- 新增 `server/internal/handler/movement.go` 的 `applyDungeonTeleportEffect(player)` helper：
  - 設定 `player.AbsoluteBarrier = true`
  - 重置 `player.RegenHPAcc = 0`（對齊 Java `stopHpRegeneration`；MP regen 為全域 tick 觸發無 per-player accumulator）
  - 加入 `ActiveBuff{SkillID: 78, TicksLeft: 10, SetAbsoluteBarrier: true}`（2 秒 × 5 ticks/秒）
  - 對齊 Java `setSkillEffect` 同 ID 覆寫行為：原 skill 78 buff（如玩家剛 cast 過 12 秒 ABSOLUTE_BARRIER）會被 2 秒覆寫。
- `HandleMove` 兩個傳送分支（固定 `Portals.Get` 和隨機 `RandomPortals.Get`）在 `teleportPlayer` 前各加一行 `applyDungeonTeleportEffect(player)`。
- RegenSystem 已有的 `if p.AbsoluteBarrier { return }` 短路會自動讓 2 秒 HP/MP regen 跳過（既有架構直接覆蓋 Java `stopHpRegeneration/stopMpRegeneration` 語義）。
- 船票驗證失敗（`isDock && !allowed`）走「繼續正常移動」分支，不取得屏障——對齊 Java `dg() returns false` 的路徑。
- 新增 `server/internal/handler/dungeon_teleport_test.go` 4 個測試：helper 設旗+重置 RegenHPAcc+加 10-tick skill 78 buff、nil 玩家無 panic、既有 60-tick AB buff 被覆寫為 10 ticks、portal YAML 載入後 helper 行為驗證。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/handler -run TestApplyDungeonTeleport` 4/4 ✅、`go test ./internal/...` 全 package 全綠（regression-free）。
- 停損：`S_Teleport2` 封包流程（Java 透過 S_Teleport2 + C_Teleport 兩段式握手；Go 用 teleportPlayer 直接傳送）屬架構差異不破壞功能，不對齊。

## MISS-P0-003 Stage E — 第一個示範副本（框架驗證 demo）

- `server/data/yaml/quest_dungeons.yaml` 新增 dungeon id=1001 「副本框架驗證 demo」（map_id=6666、time_limit=600 秒、2 個 round：on_enter spawn 2 隻青蛙、on_round_clear spawn 1 隻青蛙；exit teleport 回王城）。本副本不對應任何 Java 副本，僅為驗證 Stage A-D 框架。真實副本（火龍窟、屠龍、Chapter01R）歸 MISS-P2-015/017/019。
- `handler.QuestWorldManager` 介面擴展加入 `Enter(player, dungeonID) *QuestInstance` + `Exit(player) bool`，讓 handler 層可直接委派 system 完成入場/離場。
- 新增 `server/internal/handler/quest_dungeon.go` 提供 demo NPC 動作薄封裝：`enterDemoDungeon` / `exitDemoDungeon`（含「已在副本中拒重複入場」「未在副本中拒退場」基本驗證）。
- `npcaction.go` 新增 `case "enter_demo_dungeon"` 與 `case "exit_demo_dungeon"`，路由到上述兩個函式。
- 新增 `server/internal/system/quest_world_e2e_test.go` 3 個端對端測試覆蓋 dev doc Stage E item 25 的驗收條件：
  - `TestQuestWorldE2ESamePartySeesSameInstance` — 玩家 A Enter + 玩家 B Join → 同 ShowID、互相在 GetNearbyPlayersInShow 視野內。
  - `TestQuestWorldE2EThirdPlayerSeparateInstance` — 第三玩家自行 Enter → 新 ShowID、AOI 雙向隔離（A 看不見 C / C 看不見 A 與 B）、NPC 也跟著 ShowID 隔離（C 副本內看不到 A 副本的 NPC）。
  - `TestQuestWorldE2EFullCycle` — Enter→spawn→Exit 完整生命週期：最後玩家離開觸發 endInstance，玩家 ShowID 歸零、實例註冊表清除、副本 NPC 全清。
  - 測試透過 `net.Pipe` 建立 dummy Session 解開 SendNpcPack 對 session 的依賴。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/system -run TestQuestWorld` 23/23 ✅（20 原始 + 3 E2E）、`go test ./internal/...` 全 package 全綠（regression-free）、`server` 啟動正常載入「副本定義 1」。
- 已知範圍外（不做）：火龍窟副本/屠龍副本/Chapter01R 真實副本內容、Lua hook 實作、進場條件 DSL 實際生效（min_level/required_items/forbidden_buffs 等欄位已 schema 化但 Enter 路徑尚未驗證），歸 MISS-P0-003 後續或下放至 P2/P3 子任務。

## MISS-P0-003 Stage D — Round 引擎與 NPC 生命週期

- 新增 `server/internal/system/quest_world_spawn.go` 實作副本 NPC 出生模組：
  - `buildDungeonNpc(deps, tmpl, x, y, mapID, heading, showID)` — 從 NpcTemplate 建立 `Transient=true` + `ShowID=副本實例 ID` 的副本暫態 NPC（不寫 SpawnX/Y 等重生欄位，避免 NpcRespawnSystem 誤判）。
  - `(s *QuestWorldSystem) spawnRound(inst, round)` — 執行單一 DungeonRound 的所有 spawn 規則；含 area / fixed / group_id 支援；對 `group_id>0` 透過 `data.MobGroupTable` 帶出隊員並全部標 Transient+ShowID。
  - `pickSpawnPoint` 共用 `FindNpcSpawnPoint` 取合法座標；area 模式自動取中心點當 fallback、無 area 則用 fixed 點。
  - `addDungeonNpc` 統一處理：加入世界 + `inst.AddNpc` + `MapData.SetImpassable` + 透過 `GetNearbyPlayersInShow` 廣播給副本內玩家（自動排除主世界玩家）。
- 擴展 `server/internal/world/quest_instance.go` 加 `spawnedRounds map[int32]bool` + `MarkRoundSpawned/IsRoundSpawned` 方法防止 round 重複觸發。
- `server/internal/system/quest_world.go` 接上 Round 引擎與 NPC 生命週期：
  - `Enter` 進場後立即觸發所有 `trigger=on_enter` 的 round；MarkRoundSpawned 保證單次觸發。
  - `Update` 每 tick 檢查所有實例的 `on_timer` round（依 `timer 秒 × 5 ticks/秒` 比對 elapsedTicks），到期且未觸發過 → spawn；時間限制檢查維持 Stage C 行為。
  - `endInstance` 新增 `cleanupDungeonNpcs` 子流程：對 `inst.Npcs()` 逐個廣播 `BuildRemoveObject` + `MapData.SetImpassable(false)` + `world.RemoveNpc` + `inst.RemoveNpc`，確保副本結束時暫態 NPC 全清。
- 新增 `(s *QuestWorldSystem) OnNpcDeath(npc)` API 並加入 `handler.QuestWorldManager` 介面：
  - 從 `inst.npcs` 移除死亡 NPC → 若 `inst.NpcCount()==0` → 觸發所有 `trigger=on_round_clear` 的 round。
- `server/internal/system/combat.go` 的 `handleNpcDeath` 在群體死亡處理後加入 `if npc.ShowID > 0 && deps.QuestWorld != nil → OnNpcDeath(npc)`，把副本怪死亡訊號回傳給 QuestWorldSystem。
- 新增 `server/internal/system/quest_world_spawn_test.go` 7 個測試：on_enter+fixed 出生（含 Transient/ShowID/MapID 驗證）、on_enter+area 出生（座標落點驗證）、不同實例 spawnedRounds 隔離、endInstance NPC 全清、on_round_clear 觸發下一輪、on_timer 在 15 tick（3 秒）觸發且不重複、主世界 NPC OnNpcDeath 無副作用。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/system -run TestQuestWorld` 20/20 ✅（13 原始 + 7 新增）、`go test ./internal/...` 全 package 全綠（regression-free）。

## MISS-P0-003 Stage C — QuestWorldSystem 全域註冊器

- 新增 `server/internal/system/quest_world.go` 實作 `QuestWorldSystem`（Phase PostUpdate）：
  - `NextID()` 流水號從 100 起單調遞增（對齊 Java `WorldQuest._nextId = 100`）。
  - `Enter(player, dungeonID)` 建立新實例 → 分配 ShowID → AddPlayer → 設定 `player.ShowID` → 若 `Entry.TeleportTo` 有值則傳送。
  - `Join(player, showID)` 加入既有實例（同隊伍/血盟成員追入用）。
  - `Exit(player)` 移除玩家、清 `player.ShowID`；若 `out_stop=true` 或剩 0 人 → `endInstance`。
  - `Get / IsQuest / Count` 註冊表查詢。
  - `Update(dt)` 每 tick 累計、檢查 `TimeLimit` 過期（5 tick/秒），到期 → `endInstance`。
  - `RemoveOnDisconnect(player)` 斷線清理（實作 `handler.QuestWorldManager` 介面）。
- `endInstance` 對所有剩餘玩家清 ShowID + 若 `Exit.TeleportTo` 有值且玩家仍在副本地圖 → 傳送出去。
- 新增 `handler.QuestWorldManager` 介面 + `handler.Deps.QuestWorld` 槽。
- `InputSystem` 新增 `SetQuestWorld()` setter + 斷線回呼（仿 `HauntedHouseManager` 模式）。
- `main.go` 在 `HauntedHouseSystem` 後接 `QuestWorldSystem`（共用同一 `dungeonTable`）並註冊到 runner。
- 新增 `server/internal/system/quest_world_test.go` 13 個測試覆蓋：NextID 起始 100 + 單調、Enter 分配 ShowID + 註冊實例、Enter 不存在副本回 nil、Join 加入既有實例、Join 不存在 showID 回 nil、Exit 清 ShowID、最後玩家離開觸發 endInstance、out_stop 模式任一離開即結束、不在副本中 Exit 回 false、RemoveOnDisconnect 同 Exit、時間限制 5 tick 到期、time_limit=-1 永不過期、Update 單調 tick。
- Stage C 不做（留給 Stage D/E）：Round 引擎與 NPC 出生、進場條件 DSL、Lua hook、NPC MarkForDestruction、通關獎勵。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/system -run TestQuestWorld` 13/13 ✅、`go test ./internal/...` 全 package 全綠（regression-free）。

## MISS-P0-003 Stage B — ShowID AOI 隔離

- `world.PlayerInfo` 新增 `ShowID int32` 欄位（`0 = 主世界`，Go 零初值；副本實例 ID `>=100`）。
- `world.NpcInfo` 新增 `ShowID int32` + `Transient bool`（副本暫態 NPC 不進入持久化）。
- `world.State` 新增：
  - `GetNearbyPlayersInShow(x, y, mapID, excludeSession, viewerShowID)` — 過濾 `target.ShowID == viewerShowID`。
  - `GetNearbyNpcsInShow(x, y, mapID, viewerShowID)` — 同上。
- **既有 6 個 AOI 方法保持不變**（314 個呼叫點零回歸風險，主世界 ShowID=0 走原路徑）。
- 新增 `world/showid_aoi_test.go` 11 個測試：主世界互看、跨副本/主世界隔離、同副本可見、不同副本實例隔離、NPC 對應四情境、跳過死亡 NPC、舊方法向下相容驗證、預設零值驗證。
- 設計決策：Java 用 `-1` sentinel，Go 改用 `0` 因副本實例 ID 永遠 `>=100`、Go 零初值=0、`0/-1` 等價無歧義。
- 副本 mapID 通常與主世界不同 → AOI 物理隔離已先擋掉跨界查詢，`InShow` 是「同 mapID 多副本實例」最後一層過濾（Stage E 副本實例化時才會啟用）。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/world/...` + handler + system 三個 package 全綠（既有 AOI 測試零退化）。

## MISS-P0-003 Stage A — 副本框架資料層 + 型別

- 新增 `server/data/yaml/quest_dungeons.yaml` 統一副本宣告檔（含完整 schema 註解，目前 `dungeons: []`）。
- 新增 `server/internal/data/quest_dungeon.go` 載入器：
  - 型別：`DungeonDef`（id/name/map_id/max_users/time_limit/out_stop）+ `DungeonEntrySpec`（進場條件 DSL：等級/職業/物品/任務步驟/禁用 buff/傳送目標/拒絕訊息）+ `DungeonExitSpec`（離場規則：傳送/清理物品/通關獎勵）+ `DungeonRound`（多輪出生：`on_enter`/`on_round_clear`/`on_timer`/`on_lua` 四種觸發）+ `DungeonSpawn`（area/fixed 雙模式）+ `DungeonHookSpec`（Lua hook 路徑）+ `DungeonReward`（item/exp/adena）。
  - API：`LoadDungeonTable(path)` / `Get(id)` / `All()` / `Count()` / `IsDungeonMap(mapID)`（對應 Java `QuestMapTable.isQuestMap`）。
  - 驗證：拒絕重複 ID、`class_mask>127`（3.80C 無戰士）、`trigger=on_timer` 無 timer、spawn 無 area/fixed、hook 格式錯誤。
- 新增 `server/internal/world/quest_instance.go` 運行期型別 `QuestInstance`（對應 Java `L1QuestUser`）：玩家清單 / NPC 清單 / 時間限制 / 分數 / 章節狀態欄位（`ChapterState interface{}`）/ 自訂變數槽（給 Lua hook 用）/ 玩家與 NPC 管理方法。
- 接 loader 到 `server/cmd/l1jgo/main.go`：任務範本載入後執行 `data.LoadDungeonTable("data/yaml/quest_dungeons.yaml")`、`printStat("副本定義", ...)`；本 Stage 用 `_ = dungeonTable` 顯式標記僅載入無 runtime 行為。
- 新增測試 `server/internal/data/quest_dungeon_test.go`：8 個測試覆蓋 empty / 最小 / 完整 / 重複 ID 拒絕 / class_mask=255 拒絕 / on_timer 缺 timer 拒絕 / spawn 缺位置拒絕 / hook 格式錯誤拒絕。
- 同步文件 `docs/實作開發紀錄.md` MISS-P0-003 狀態 ❌→🔶（進行中）並補上 Stage A 子項紀錄。
- 設計文件 `docs/副本系統開發文檔.md` Stage A 工作分解更新為「統一 quest_dungeons.yaml」（取代原本 quest_maps.yaml + quest_spawns.yaml 雙檔設計），對應 4.5 自訂副本擴展性章節。
- 驗證：`cd server && go build ./...` 通過、`go test ./internal/data -run TestLoadDungeon` 8 個測試全綠。
