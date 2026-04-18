/**
 * macOS デスクトップ通知
 * urgency が urgent のアイテムに対して osascript で通知を表示する
 */
package notify

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/senna-lang/sentei/internal/plugin"
)

// Notifier はシステム通知を送信するインターフェース
type Notifier interface {
	Notify(item plugin.LabeledItem) error
}

// DarwinNotifier は macOS の osascript を使ってデスクトップ通知を送信する
type DarwinNotifier struct{}

// Notify は osascript で macOS 通知を表示する
func (d *DarwinNotifier) Notify(item plugin.LabeledItem) error {
	title := "sentei"
	body := fmt.Sprintf("[%s] %s", item.Label.Category, item.Item.Title)

	// AppleScript の文字列エスケープ
	body = escapeAppleScript(body)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
	return exec.Command("osascript", "-e", script).Run()
}

// BuildCommand は osascript コマンドの引数を返す（テスト用）
func (d *DarwinNotifier) BuildCommand(item plugin.LabeledItem) []string {
	title := "sentei"
	body := fmt.Sprintf("[%s] %s", item.Label.Category, item.Item.Title)
	body = escapeAppleScript(body)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
	return []string{"osascript", "-e", script}
}

// NoopNotifier は何もしない通知実装（テスト・非 macOS 用）
type NoopNotifier struct{}

// Notify は何もせず nil を返す
func (n *NoopNotifier) Notify(item plugin.LabeledItem) error {
	return nil
}

// escapeAppleScript は AppleScript 文字列内の特殊文字をエスケープする
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
