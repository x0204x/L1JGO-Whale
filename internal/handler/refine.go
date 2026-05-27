package handler

import (
	"fmt"
	"math"

	"github.com/l1jgo/server/internal/data"
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"github.com/l1jgo/server/internal/world"
	"go.uber.org/zap"
)

// handleRefineResolve 處理火神精煉（C_PledgeContent type=13）— 分解物品為結晶。
// Java 815: C_PledgeContent case 13
// 封包格式：[D npcObjID][D itemObjID][D assistItemObjID]
//
// 流程：
//  1. 驗證 NPC 存在且在範圍內
//  2. 從背包取得物品，查火結晶表計算結晶數量
//  3. 移除原物品
//  4. 給予魔法結晶體（item 41246）
func handleRefineResolve(sess *net.Session, r *packet.Reader, player *world.PlayerInfo, deps *Deps) {
	npcObjID := r.ReadD()
	itemObjID := r.ReadD()
	assistObjID := r.ReadD()

	deps.Log.Info("火神精煉/收 data=13",
		zap.String("char", player.Name),
		zap.Int32("npcObjID", npcObjID),
		zap.Int32("itemObjID", itemObjID),
		zap.Int32("assistObjID", assistObjID),
	)

	// 驗證 NPC 存在且在範圍內
	npc := deps.World.GetNpc(npcObjID)
	if npc == nil {
		deps.Log.Warn("火神精煉/拒：NPC 不存在", zap.Int32("npcObjID", npcObjID))
		return
	}
	dx := int32(math.Abs(float64(player.X - npc.X)))
	dy := int32(math.Abs(float64(player.Y - npc.Y)))
	if dx > 5 || dy > 5 {
		deps.Log.Warn("火神精煉/拒：超出範圍", zap.Int32("dx", dx), zap.Int32("dy", dy))
		return
	}

	// 查找玩家背包中的物品
	item := player.Inv.FindByObjectID(itemObjID)
	if item == nil {
		deps.Log.Warn("火神精煉/拒：背包查無此物品", zap.Int32("itemObjID", itemObjID))
		sendGlobalChat(sess, 9, "\\f3無法精煉該物品。")
		return
	}
	if item.Equipped {
		deps.Log.Warn("火神精煉/拒：物品已裝備", zap.Int32("itemID", item.ItemID))
		sendGlobalChat(sess, 9, "\\f3無法精煉該物品。")
		return
	}

	// 火結晶表未載入
	if deps.FireCrystals == nil {
		deps.Log.Warn("火神精煉/拒：FireCrystals 未載入")
		sendGlobalChat(sess, 9, "\\f3精煉系統尚未啟用。")
		return
	}

	itemInfo := deps.Items.Get(item.ItemID)
	if itemInfo == nil || itemInfo.Category == data.CategoryEtcItem {
		deps.Log.Warn("火神精煉/拒：物品類別不可精煉",
			zap.Int32("itemID", item.ItemID),
			zap.Any("category", itemInfo),
		)
		sendGlobalChat(sess, 9, "\\f3此物品無法精煉。")
		return
	}

	// 計算基礎 item ID（去除祝福/詛咒偏移）
	// Java: bless==0 → itemId-100000; bless==2 → itemId-200000
	lookupID := item.ItemID
	if item.Bless == 0 { // 祝福狀態
		candidateID := item.ItemID - 100000
		if ci := deps.Items.Get(candidateID); ci != nil && ci.Name == itemInfo.Name {
			lookupID = candidateID
		}
	} else if item.Bless == 2 { // 詛咒狀態
		candidateID := item.ItemID - 200000
		if ci := deps.Items.Get(candidateID); ci != nil && ci.Name == itemInfo.Name {
			lookupID = candidateID
		}
	}

	entry := deps.FireCrystals.Get(lookupID)
	if entry == nil {
		deps.Log.Warn("火神精煉/拒：fire_crystal_list 無對應條目",
			zap.Int32("itemID", item.ItemID),
			zap.Int32("lookupID", lookupID),
			zap.Int32("bless", int32(item.Bless)),
		)
		sendGlobalChat(sess, 9, "\\f3此物品無法精煉。")
		return
	}

	crystalCount := entry.GetCrystalCount(int(item.EnchantLvl), int(itemInfo.Category), itemInfo.SafeEnchant)
	if crystalCount <= 0 {
		deps.Log.Warn("火神精煉/拒：crystalCount<=0（強化等級不足）",
			zap.Int32("itemID", item.ItemID),
			zap.Int8("enchant", item.EnchantLvl),
			zap.Int("safeEnchant", itemInfo.SafeEnchant),
		)
		sendGlobalChat(sess, 9, "\\f3此物品無法精煉。")
		return
	}

	// 火神之淚加成：assist 槽放入火神之淚 → 結晶數 ×（1+bonus）+ 扣 1 個火神之淚
	usedTear := false
	if assistObjID > 0 {
		if tear := player.Inv.FindByObjectID(assistObjID); tear != nil && tear.ItemID == data.FireSmithTearItemID {
			baseCount := crystalCount
			crystalCount = int32(float64(crystalCount) * (1 + data.FireSmithRefineTearBonus))
			usedTear = true
			deps.Log.Info("火神精煉/淚加成",
				zap.String("char", player.Name),
				zap.Int32("base", baseCount),
				zap.Int32("final", crystalCount),
				zap.Float64("bonus", data.FireSmithRefineTearBonus),
			)
		}
	}

	// 委派系統執行分解（移除裝備 + 給予火神結晶體 80029）
	deps.NpcSvc.Refine(sess, player, item, data.FireSmithCrystalItemID, crystalCount)

	// 扣火神之淚（Refine 已執行完才扣；確保失敗早退時不誤扣）
	if usedTear {
		if !deps.NpcSvc.ConsumeItem(sess, player, assistObjID, 1) {
			deps.Log.Warn("火神精煉/淚扣除失敗",
				zap.String("char", player.Name),
				zap.Int32("assistObjID", assistObjID),
			)
		}
	}

	// 系統訊息：獲得 X 個火神結晶體
	sendGlobalChat(sess, 9, fmt.Sprintf("\\f2獲得 %d 個火神結晶體。", crystalCount))
	deps.Log.Info(fmt.Sprintf("火神精煉  角色=%s  物品=%d(+%d)  結晶=%d", player.Name, item.ItemID, item.EnchantLvl, crystalCount))
}

// handleRefineTransform 處理火神合成（C_PledgeContent type=14）— 材料合成裝備。
// 3.80C 客戶端反編譯（RVA 0x29774D，mode 2 分支）封包格式 "ccdhdd"：
//   [D npcObjID]      // SmithUI+0x1cc
//   [H actionID]      // SmithUI+0x1dc，來源為客戶端 MakeInfo.tbl 第 1 欄
//   [D plusItemObjID] // SmithUI+0x1e4，玩家拖入的火神之槌/火神之淚 objID
//   [D plusItemCount] // SmithUI+0x20c，玩家拖入的火神之槌/火神之淚數量
//
// actionID 由客戶端 MakeInfo.tbl 定義；server 端對齊用 firesmith_recipe_list.yaml 查表。
// plus 數量直接用於成功率加成（每個 hammer +PlusHammerBonus%）並全部消耗。
func handleRefineTransform(sess *net.Session, r *packet.Reader, player *world.PlayerInfo, deps *Deps) {
	npcObjID := r.ReadD()
	actionID := r.ReadH()
	plusItemObjID := r.ReadD()
	plusItemCount := r.ReadD()

	deps.Log.Info("火神合成/收 data=14",
		zap.String("char", player.Name),
		zap.Int32("npcObjID", npcObjID),
		zap.Uint16("actionID", actionID),
		zap.Int32("plusItemObjID", plusItemObjID),
		zap.Int32("plusItemCount", plusItemCount),
	)

	// 驗證 NPC 存在且在範圍內
	npc := deps.World.GetNpc(npcObjID)
	if npc == nil {
		return
	}
	dx := int32(math.Abs(float64(player.X - npc.X)))
	dy := int32(math.Abs(float64(player.Y - npc.Y)))
	if dx > 5 || dy > 5 {
		return
	}

	if deps.FireSmithRecipes == nil {
		sendGlobalChat(sess, 9, "\\f3製作系統尚未啟用。")
		return
	}

	recipe := deps.FireSmithRecipes.Get(int32(actionID))
	if recipe == nil {
		deps.Log.Warn("火神合成/拒：找不到對應配方",
			zap.Int32("npcID", npc.NpcID),
			zap.Uint16("actionID", actionID),
		)
		sendGlobalChat(sess, 9, "\\f3找不到對應的製作配方。")
		return
	}

	deps.NpcSvc.FireSmithCraft(sess, player, recipe, plusItemObjID, plusItemCount)
}
