# 變更摘要（下次推送用）

## v0.3.29 — 火龍窟副本（MISS-P2-016）完整實裝

### 新功能：火龍巢穴（單人副本，mapid 2600）

完整實裝官方流程：威頓村 NPC 漢 → 漢的袋子（24h cooldown）→ 點擊隨機開出冷冽的氣息 → 找愛德納斯傳送 → 入副本拿真死亡騎士烈炎之劍 → 三區戰鬥 → Boss 三選一 → 通關回奇岩。

- 副本配置：max_users=1、time_limit=1800s、out_stop=true、min_level=60
- 三區域漸進式 Round 推進（Round -1/1/2/3，總 4 個 Round）
- 最終 Boss 三選一隨機出現：伊弗利特(111532) / 不死鳥(4070080) / 巴拉卡斯(4070081)
- 武器需求機制：必須裝備真死亡騎士烈炎之劍(item 850) 才能對副本怪造成傷害
- 離場自動刪除 item 850（cleanup_items），避免帶出副本當一般武器使用

### 框架擴展（副本系統）

- `data.NpcTemplate.WeaponRequired` / `world.NpcInfo.WeaponRequired` 新欄位
- `NpcInfo.CanReceiveDamageFrom(weaponID)` 共用判斷 API
- 戰鬥傷害檢查覆蓋 5 個玩家來源傷害路徑：
  - `combat.go` melee 主路徑
  - `skill_damage.go` 單體技能 + AoE 多 hit
  - `skill_dragonknight.go` FoeSlayer 龍騎士魔法
  - `skill_self_area.go` 自我中心 AoE
  - `weapon_skill.go` 武器技能 AoE proc
- `DungeonSpawn.Auxiliary` 旗標：商人/對話 NPC 不計入 round clear（解決死亡騎士永遠卡住下一輪的問題）
- `DungeonRound.RandomPick` 旗標：spawns 中隨機選 1 條執行（用於 Boss 三選一）
- `QuestInstance.AddAuxiliaryNpc / AuxiliaryNpcs`：副本結束時統一回收
- `QuestWorldSystem.OnNpcDeath` 改為「每次全清只觸發 1 個未出生 on_round_clear round」，避免多 round 同時觸發
- 副本最後 round 清空時自動 `endInstance("last_mob_death")`

### 5 個 template→runtime 站點同步補上 WeaponRequired
- `quest_world_spawn.go` / `npc_respawn.go` / `item_use.go` / `npc_ai.go` / `gmcommand.go`

### NPC 動作 handler（薄層）
- `give_hans_bag`（46180 漢）→ 24h cooldown 檢查 + 給漢的袋子
- `enter_fire_dragon_cave`（46181 愛德納斯）→ 委派 `QuestWorld.Enter(player, 148)`
- `give_flame_sword`（46164 死亡騎士）→ 副本內檢查 + 給 item 850

### 玩家欄位
- `PlayerInfo.NextHansBagAt int64` — 漢的袋子下次可領 Unix 秒（24h cooldown）

### 資料新增
- **`quest_dungeons.yaml`**：id=148 完整副本宣告（4 round、entry/exit、cleanup_items 850）
- **`npc_list.yaml`**：10 個新 NPC（漢/愛德納斯/死亡騎士 + 4 mob 變體 + 3 Boss）
- **`weapon_list.yaml`**：item 850 真死亡騎士烈炎之劍（雙手劍，tradeable=false）
- **`etcitem_list.yaml`**：80001 漢的袋子（24h cooldown）+ 80020 冷冽的氣息 + 3 補項 drop 物品
- **`item_box.yaml`**：漢的袋子開箱 5 種結果（冷氣 30% / 金幣 35% / 紅藥 15% / 橙藥 12% / 純白萬能藥 8%）
- **`drop_list.yaml`**：3 Boss 各 10 種掉落（金幣必掉 + 9 種隨機）
- **`spawn_list.yaml`**：漢 + 愛德納斯 常駐 spawn（威頓村 mapid 4）
- **`npc_action_list.yaml`**：3 NPC 的 normal_action 字串映射

### 刻意不做
- 龍印魔石寶箱 / 龍耀符石寶箱 / 惡魔氣息 / 天幣（4 種 Boss 掉落物 815.sql 未找到 ID）→ 留待物品系統範疇
- 副本內掉落「真死亡騎士烈炎之劍」的對話樹細節 → 用最小可行 give action
- 24h cooldown 持久化到 DB（目前 NextHansBagAt 隨 player 持久化）

### 驗證
- `go build ./...` 通過
- 伺服器啟動成功：副本定義 2 個（含 148）、NPC 模板 3450、道具模板 4111、物品箱 3、掉寶表 1018
