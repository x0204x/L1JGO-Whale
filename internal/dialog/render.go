package dialog

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// templateFuncs Go text/template 內可用的 helper 函式。
// .htm 檔內以 {{ funcName .Arg }} 呼叫。
var templateFuncs = template.FuncMap{
	"remaining":     tplRemaining,
	"format_hms":    tplFormatHMS,
	"format_minsec": tplFormatMinSec,
	"upper":         strings.ToUpper,
	"lower":         strings.ToLower,
	"color":         tplColor,
	"itoa":          tplItoa,
}

// tplRemaining 把 unix-second 時間戳轉成相對現在的剩餘時間，格式「X 時 Y 分 Z 秒」。
// 過期回 "0 秒"。
//
// 用法：{{ remaining .NextHansBagAt }}
func tplRemaining(targetUnix int64) string {
	remaining := targetUnix - time.Now().Unix()
	if remaining <= 0 {
		return "0 秒"
	}
	h := remaining / 3600
	m := (remaining % 3600) / 60
	s := remaining % 60
	var b strings.Builder
	if h > 0 {
		b.WriteString(strconv.FormatInt(h, 10))
		b.WriteString(" 時 ")
	}
	if h > 0 || m > 0 {
		b.WriteString(strconv.FormatInt(m, 10))
		b.WriteString(" 分 ")
	}
	b.WriteString(strconv.FormatInt(s, 10))
	b.WriteString(" 秒")
	return b.String()
}

// tplFormatHMS 把秒數格式化為 HH:MM:SS。
// 用法：{{ format_hms 3725 }} → "01:02:05"
func tplFormatHMS(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return zeroPad2(h) + ":" + zeroPad2(m) + ":" + zeroPad2(s)
}

// tplFormatMinSec 格式化為 MM:SS。
func tplFormatMinSec(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	m := seconds / 60
	s := seconds % 60
	return zeroPad2(m) + ":" + zeroPad2(s)
}

// tplColor 包成 L1J HTML 的彩色文字標籤。fg 為 hex 字串（不含 #）。
// 用法：{{ color "00ff00" "綠色文字" }}
func tplColor(fg, text string) string {
	return `<font fg="` + fg + `">` + text + `</font>`
}

// tplItoa 整數轉字串（int64 / int32 / int 都接）。
func tplItoa(v any) string {
	switch x := v.(type) {
	case int:
		return strconv.Itoa(x)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	default:
		return ""
	}
}

func zeroPad2(v int64) string {
	if v < 10 && v >= 0 {
		return "0" + strconv.FormatInt(v, 10)
	}
	return strconv.FormatInt(v, 10)
}

// RenderContext 提供 .htm 模板可讀的玩家資料（白名單欄位）。
// 直接讓模板存取整個 PlayerInfo 太危險（可改寫狀態、暴露未公開欄位），
// 改用這個薄層 view 結構控制可見性。
type RenderContext struct {
	// 玩家基本資料
	Level        int16
	ClassID      int32
	Gold         int64
	Lawful       int32
	Name         string

	// 任務/冷卻欄位（白名單，避免反射開放整個 PlayerInfo）
	NextHansBagAt int64

	// 環境
	ShowID int32 // 副本實例 ID（0 = 不在副本）
	MapID  int16

	// NPC 名稱（speaker 已在外層套用 <font> 前綴，這裡是給特殊內文用）
	NpcName string

	// 給模板存取「現在時間」用（避免 .htm 寫硬編碼）
	NowUnix int64
}

// Render 用編譯後的模板 + RenderContext 產出最終 HTML body。
// 失敗時回空字串；呼叫方應 log 並走 fallback。
func (d *DialogTemplate) Render(ctx *RenderContext) (string, error) {
	if d == nil || d.Compiled == nil {
		return "", nil
	}
	var buf bytes.Buffer
	if err := d.Compiled.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WrapBodyWithChrome 把對話內文套上 <html><body> + speaker 前綴。
// speaker / speakerColor 為空時不加前綴。
func WrapBodyWithChrome(rawBody, speaker, speakerColor string) string {
	var b strings.Builder
	b.WriteString("<html><body>")
	if speaker != "" {
		color := speakerColor
		if color == "" {
			color = "ffcc00"
		}
		b.WriteString(`<font fg="`)
		b.WriteString(color)
		b.WriteString(`">`)
		b.WriteString(speaker)
		b.WriteString(":</font> ")
	}
	b.WriteString(rawBody)
	b.WriteString("</body></html>")
	return b.String()
}
