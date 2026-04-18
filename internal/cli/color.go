/**
 * ターミナル色出力ヘルパー
 * urgency に応じた色分けと ANSI エスケープシーケンスを提供する
 * NO_COLOR 環境変数で色出力を無効化できる
 */
package cli

import (
	"fmt"
	"os"

	"github.com/senna-lang/sentei/internal/plugin"
)

// Color は ANSI カラーコード
type Color string

const (
	ColorReset Color = "\033[0m"
	ColorRed   Color = "\033[31m"
	ColorYellow Color = "\033[33m"
	ColorGreen Color = "\033[32m"
	ColorGray  Color = "\033[90m"
	ColorBold  Color = "\033[1m"
)

// colorEnabled は色出力が有効かどうかを返す
func colorEnabled() bool {
	_, noColor := os.LookupEnv("NO_COLOR")
	return !noColor
}

// Colorize は文字列を指定した色で装飾する
func Colorize(s string, c Color) string {
	if !colorEnabled() {
		return s
	}
	return fmt.Sprintf("%s%s%s", c, s, ColorReset)
}

// UrgencyColor は urgency に対応する色を返す
func UrgencyColor(u plugin.Urgency) Color {
	switch u {
	case plugin.UrgencyUrgent:
		return ColorRed
	case plugin.UrgencyShouldCheck:
		return ColorYellow
	case plugin.UrgencyCanWait:
		return ColorReset
	case plugin.UrgencyIgnore:
		return ColorGray
	default:
		return ColorReset
	}
}

// FormatUrgency は urgency を色付きで表示する
func FormatUrgency(u plugin.Urgency) string {
	return Colorize(string(u), UrgencyColor(u))
}

// Success は緑色の成功メッセージを返す
func Success(s string) string {
	return Colorize(s, ColorGreen)
}

// Error は赤色のエラーメッセージを返す
func Error(s string) string {
	return Colorize(s, ColorRed)
}

// Bold は太字テキストを返す
func Bold(s string) string {
	return Colorize(s, ColorBold)
}
