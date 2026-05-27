package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadAll 掃描 baseDir（例如 server/data/dialogs/），對每個子資料夾建構一個 Registry。
// 子資料夾名稱規則：`<npc_id>_<helper_name>`，例如 "46180_han"、"46181_eldnas"。
// 首段數字（46180）即為 npc_id，後面助記名僅供人類閱讀。
//
// 載入完成後回傳 map[npc_id]*Registry，可供 Manager.SetAll 使用。
func LoadAll(baseDir string) (map[int32]*Registry, error) {
	out := make(map[int32]*Registry)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil // 還沒任何對話定義；非錯誤
		}
		return nil, fmt.Errorf("dialog: read base dir %s: %w", baseDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "_") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		reg, err := loadNpcFolder(filepath.Join(baseDir, e.Name()), e.Name())
		if err != nil {
			return nil, fmt.Errorf("dialog: load %s: %w", e.Name(), err)
		}
		if reg == nil {
			continue
		}
		if existing, dup := out[reg.NpcID]; dup {
			return nil, fmt.Errorf("dialog: npc_id %d 重複定義（%s 與 %s）",
				reg.NpcID, existing.FolderName, reg.FolderName)
		}
		out[reg.NpcID] = reg
	}
	return out, nil
}

// folderNamePat 從資料夾名解出 npc_id（第一段數字）。
var folderNamePat = regexp.MustCompile(`^(\d+)(?:_.*)?$`)

// loadNpcFolder 載入單一 NPC 資料夾（_routes.yaml + 所有 .htm）。
func loadNpcFolder(dir, folderName string) (*Registry, error) {
	m := folderNamePat.FindStringSubmatch(folderName)
	if m == nil {
		return nil, fmt.Errorf("資料夾名 %q 不符 <npc_id>[_<name>] 格式", folderName)
	}
	npcID64, err := strconv.ParseInt(m[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("解析 npc_id %q: %w", m[1], err)
	}

	// 1) 讀 _routes.yaml
	routesPath := filepath.Join(dir, "_routes.yaml")
	routesData, err := os.ReadFile(routesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 沒 _routes.yaml → 略過這個資料夾（允許暫時擺空殼）
			return nil, nil
		}
		return nil, fmt.Errorf("讀 %s: %w", routesPath, err)
	}
	var raw rawRoute
	if err := yaml.Unmarshal(routesData, &raw); err != nil {
		return nil, fmt.Errorf("解析 %s: %w", routesPath, err)
	}

	reg := &Registry{
		NpcID:        int32(npcID64),
		FolderName:   folderName,
		Speaker:      raw.DefaultSpeaker,
		SpeakerColor: defaultIfEmpty(raw.DefaultSpeakerColor, "ffcc00"),
		OnAction:     make(map[string]*ActionDef, len(raw.OnAction)),
		Dialogs:      make(map[string]*DialogTemplate),
	}

	// 2) on_talk：解析 refresh / duration 字串為 time.Duration
	for i, b := range raw.OnTalk {
		branch := TalkBranch{When: b.When, Send: b.Send}
		if b.Refresh != "" {
			d, err := time.ParseDuration(b.Refresh)
			if err != nil {
				return nil, fmt.Errorf("%s on_talk[%d] refresh %q 不是合法 duration: %w",
					routesPath, i, b.Refresh, err)
			}
			branch.Refresh = d
		}
		if b.Duration != "" {
			d, err := time.ParseDuration(b.Duration)
			if err != nil {
				return nil, fmt.Errorf("%s on_talk[%d] duration %q 不是合法 duration: %w",
					routesPath, i, b.Duration, err)
			}
			branch.Duration = d
		}
		reg.OnTalk = append(reg.OnTalk, branch)
	}

	// 3) on_action：直接搬資料
	for action, def := range raw.OnAction {
		if def == nil {
			continue
		}
		reg.OnAction[strings.ToLower(action)] = &ActionDef{
			Require:  def.Require,
			Effects:  def.Effects,
			ThenSend: def.ThenSend,
		}
	}

	// 4) 掃所有 .htm 檔，每個編譯成一個 DialogTemplate
	htmFiles, err := filepath.Glob(filepath.Join(dir, "*.htm"))
	if err != nil {
		return nil, fmt.Errorf("掃 .htm 檔 %s: %w", dir, err)
	}
	for _, htmPath := range htmFiles {
		key := strings.TrimSuffix(filepath.Base(htmPath), ".htm")
		if strings.HasPrefix(key, "_") || strings.HasPrefix(key, ".") {
			continue
		}
		raw, err := os.ReadFile(htmPath)
		if err != nil {
			return nil, fmt.Errorf("讀 %s: %w", htmPath, err)
		}
		normalized := normalizeHTML(string(raw))
		tpl, err := template.New(key).Funcs(templateFuncs).Parse(normalized)
		if err != nil {
			return nil, fmt.Errorf("編譯模板 %s: %w", htmPath, err)
		}
		reg.Dialogs[key] = &DialogTemplate{
			Key:      key,
			RawHTML:  normalized,
			Compiled: tpl,
		}
	}

	// 5) 驗證 on_talk / on_action 引用的 dialog key 都存在
	for i, b := range reg.OnTalk {
		if _, ok := reg.Dialogs[b.Send]; !ok {
			return nil, fmt.Errorf("%s on_talk[%d] send %q 沒對應的 .htm 檔", routesPath, i, b.Send)
		}
	}
	for action, def := range reg.OnAction {
		if def.ThenSend != "" {
			if _, ok := reg.Dialogs[def.ThenSend]; !ok {
				return nil, fmt.Errorf("%s on_action[%q] then_send %q 沒對應的 .htm 檔",
					routesPath, action, def.ThenSend)
			}
		}
	}

	return reg, nil
}

// normalizeHTML 把 .htm 原始內容轉成可送上線的單行 HTML body：
//   - 去掉 \r
//   - 把 \n 換成單一空白（編輯器換行 = HTML 顯示用空白；要換行請寫 <br>）
//   - 摺疊多重空白（含 tab）為一個空白
//   - 移除 <!-- ... --> 註解
//   - 移除頭尾空白
func normalizeHTML(s string) string {
	// 移除註解
	s = htmlCommentPat.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = multiSpacePat.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

var (
	htmlCommentPat = regexp.MustCompile(`(?s)<!--.*?-->`)
	multiSpacePat  = regexp.MustCompile(`[ ]{2,}`)
)

func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
