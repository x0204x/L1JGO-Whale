package data

import (
	"path/filepath"
	"testing"
)

func TestMobSkillYamlMatchesYiweiColumnOrderForMagicMobs(t *testing.T) {
	table, err := LoadMobSkillTable(filepath.Join("..", "..", "data", "yaml", "mob_skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入 mob_skill_list.yaml 失敗: %v", err)
	}

	cat := table.Get(45039)
	if len(cat) != 2 {
		t.Fatalf("45039 應有 2 筆 yiwei mob skill，got=%d", len(cat))
	}
	if cat[0].Type != 1 || cat[0].TriggerRandom != 25 || cat[0].TriggerRange != -14 ||
		cat[0].ChangeTarget != 0 || cat[0].Range != 1 || cat[0].AreaWidth != 0 ||
		cat[0].AreaHeight != 0 || cat[0].ActID != 1 {
		t.Fatalf("45039 act01 欄位順序不符 yiwei: %+v", cat[0])
	}
	if cat[1].Type != 2 || cat[1].TriggerRandom != 25 || cat[1].TriggerRange != -14 ||
		cat[1].SkillID != 11003 || cat[1].Leverage != 0 {
		t.Fatalf("45039 魔法技能應讀入 yiwei SkillId=11003 且 leverage=0: %+v", cat[1])
	}

	harpy := table.Get(45264)
	if len(harpy) != 2 {
		t.Fatalf("45264 應有 2 筆 yiwei mob skill，got=%d", len(harpy))
	}
	if harpy[0].Type != 2 || harpy[0].SkillID != 28 || harpy[0].TriggerRandom != 25 {
		t.Fatalf("45264 吸血鬼之吻欄位順序不符 yiwei: %+v", harpy[0])
	}
	if harpy[1].Type != 2 || harpy[1].SkillID != 33 || harpy[1].TriggerRandom != 25 {
		t.Fatalf("45264 黑闇之影欄位順序不符 yiwei: %+v", harpy[1])
	}

	darkElf := table.Get(45245)
	if len(darkElf) != 2 {
		t.Fatalf("45245 應有 2 筆 yiwei mob skill，got=%d", len(darkElf))
	}
	if darkElf[1].ActNo != 0 || darkElf[1].Type != 3 || darkElf[1].SummonID != 45244 ||
		darkElf[1].ReuseDelay != 10000 {
		t.Fatalf("45245 召喚闇之精靈應帶入 yiwei reuseDelay=10000: %+v", darkElf[1])
	}

	bear := table.Get(45040)
	if len(bear) == 0 {
		t.Fatalf("45040 應有 yiwei mob skill")
	}
	if bear[0].Type != 1 || bear[0].Range != 1 || bear[0].AreaWidth != 3 || bear[0].AreaHeight != 2 {
		t.Fatalf("45040 熊範圍物理技能應帶入 yiwei range/area_width/area_height: %+v", bear[0])
	}
}

func TestMobSkillReferencedSkillTemplatesExistLikeYiwei(t *testing.T) {
	mobSkills, err := LoadMobSkillTable(filepath.Join("..", "..", "data", "yaml", "mob_skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入 mob_skill_list.yaml 失敗: %v", err)
	}
	skills, err := LoadSkillTable(filepath.Join("..", "..", "data", "yaml", "skill_list.yaml"))
	if err != nil {
		t.Fatalf("載入 skill_list.yaml 失敗: %v", err)
	}

	// 這四筆存在於 yiwei mobskill.sql，但 yiwei skills.sql 本身沒有對應模板。
	knownMissingInYiweiSkills := map[int]bool{
		4017:  true,
		41583: true,
		41606: true,
		41607: true,
	}

	missing := make(map[int]int)
	for _, list := range mobSkills.skills {
		for _, sk := range list {
			if sk.SkillID == 0 || knownMissingInYiweiSkills[sk.SkillID] {
				continue
			}
			if skills.Get(int32(sk.SkillID)) == nil {
				missing[sk.SkillID]++
			}
		}
	}
	if len(missing) != 0 {
		t.Fatalf("mob_skill_list.yaml 引用的 yiwei 技能模板缺少 %d 種，會讓怪物退回近戰: %v", len(missing), missing)
	}
}
