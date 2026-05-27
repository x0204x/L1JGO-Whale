package data

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FireSmithRecipe 火神精煉合成配方（MakeItem_Window / type 49 用）。
// 來源：客戶端 MakeInfo.tbl + 反編譯封包格式（ccdhdd）。
// 配方以 ActionID 為唯一鍵；客戶端送 actionID，server 查表取得材料與成品。
type FireSmithRecipe struct {
	ActionID        int32  `yaml:"action_id"`
	Note            string `yaml:"note"`
	NewItemID       int32  `yaml:"new_item_id"`
	NewItemCount    int32  `yaml:"new_item_count"`
	NewEnchantLvl   int8   `yaml:"new_enchant_lvl"`
	NewBless        int8   `yaml:"new_bless"`
	CrystalCount    int32  `yaml:"crystal_count"`     // 火神結晶體（80029）消耗量
	ContractCount   int32  `yaml:"contract_count"`    // 火神契約（80028）消耗量
	CatalystItemID  int32  `yaml:"catalyst_item_id"`  // 玩家拖入 catalyst slot 的物品
	CatalystCount   int32  `yaml:"catalyst_count"`    // catalyst 物品需要數量
	SuccessRate     int32  `yaml:"success_rate"`      // 基礎成功率 0~100
	PlusTearBonus   int32  `yaml:"plus_tear_bonus"`   // 火神之淚（80030）加成成功率
	PlusHammerBonus int32  `yaml:"plus_hammer_bonus"` // 火神之槌（80027）加成
}

type fireSmithRecipeFile struct {
	Recipes []FireSmithRecipe `yaml:"recipes"`
}

// FireSmithRecipeTable 配方表 — 索引 actionID → FireSmithRecipe。
type FireSmithRecipeTable struct {
	byActionID map[int32]*FireSmithRecipe
}

// LoadFireSmithRecipeTable 從 YAML 讀取火神合成配方表。
func LoadFireSmithRecipeTable(path string) (*FireSmithRecipeTable, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read firesmith_recipe_list: %w", err)
	}
	var f fireSmithRecipeFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parse firesmith_recipe_list: %w", err)
	}
	t := &FireSmithRecipeTable{byActionID: make(map[int32]*FireSmithRecipe, len(f.Recipes))}
	for i := range f.Recipes {
		r := f.Recipes[i]
		t.byActionID[r.ActionID] = &r
	}
	return t, nil
}

// Get 透過 actionID 查配方；查無回傳 nil。
func (t *FireSmithRecipeTable) Get(actionID int32) *FireSmithRecipe {
	if t == nil {
		return nil
	}
	return t.byActionID[actionID]
}

// Count 已載入配方數。
func (t *FireSmithRecipeTable) Count() int {
	if t == nil {
		return 0
	}
	return len(t.byActionID)
}

// 火神精煉系統共用材料 item_id（與 etcitem_list.yaml 對齊）
const (
	FireSmithCrystalItemID  int32 = 80029 // 火神結晶體
	FireSmithContractItemID int32 = 80028 // 火神契約
	FireSmithHammerItemID   int32 = 80027 // 火神之槌
	FireSmithTearItemID     int32 = 80030 // 火神之淚
)

// FireSmithRefineTearBonus 火神精煉時 assist 槽放入火神之淚的結晶數量加成倍率。
// 公式：finalCount = baseCount × (1 + FireSmithRefineTearBonus)
// 例：0.2 → +20%（baseCount=26 → 31）；0.5 → +50%；1.0 → +100%（×2）。
// 客戶端封包 ccddd 只送 1 個 assistObjID，每次煉化最多用 1 個火神之淚。
// 此值為專案自訂（客戶端原始公式未確認），可依平衡需要調整。
const FireSmithRefineTearBonus = 0.2
