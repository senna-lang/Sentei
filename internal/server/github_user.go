/**
 * 自分の GitHub ログイン名を gh CLI から 1 回だけ取得してキャッシュする。
 * サマリーの「あなた担当 PR」判定で使う。未取得・取得失敗時は空文字を返し、
 * CalcStats 側で「担当判定スキップ」にフォールバックする。
 */
package server

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ghUserTimeout は gh api user 呼び出しに許す時間
const ghUserTimeout = 5 * time.Second

// githubUserCache は gh api user の結果を 1 回だけ評価するためのキャッシュ
type githubUserCache struct {
	once sync.Once
	user string
}

// get は gh api user --jq .login を 1 度だけ実行し、ログイン名を返す。
// 失敗時は空文字を返し、以降の呼び出しでも再試行しない (デーモンライフタイム固定)。
func (c *githubUserCache) get() string {
	c.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), ghUserTimeout)
		defer cancel()
		out, err := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login").Output()
		if err != nil {
			slog.Warn("gh api user 取得失敗、担当 PR 判定は無効化", "error", err)
			return
		}
		c.user = strings.TrimSpace(string(out))
		if c.user != "" {
			slog.Info("GitHub user 確定", "user", c.user)
		}
	})
	return c.user
}
