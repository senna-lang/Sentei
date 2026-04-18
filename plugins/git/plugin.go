/**
 * Git プラグイン - 通知トリガー型
 * GitHub Notifications API を 1 分間隔でポーリングし、
 * 新着通知を Item に正規化してコアに送信する
 */
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

// Config は Git プラグインの設定
type Config struct {
	Enabled      bool
	GithubToken  string
	Notification NotificationConfig
	Survey       SurveyConfig
}

// NotificationConfig は通知トリガー型の設定
type NotificationConfig struct {
	PollInterval time.Duration
	Types        []string // フィルタする通知タイプ
}

// SurveyConfig はバッチサーベイ型の設定
type SurveyConfig struct {
	Interval time.Duration
	Repos    []RepoConfig
}

// RepoConfig は監視リポジトリの設定
type RepoConfig struct {
	GitHub string // "owner/repo" 形式
}

// Plugin は Git プラグインの実装
type Plugin struct {
	config    Config
	core      plugin.Core
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	lastPoll  time.Time
	seenIDs   map[string]bool // 重複検出用
	mu        sync.Mutex
}

// NewPlugin は Git プラグインを生成する
func NewPlugin(config Config) *Plugin {
	if config.Notification.PollInterval == 0 {
		config.Notification.PollInterval = 1 * time.Minute
	}
	if config.Survey.Interval == 0 {
		config.Survey.Interval = 1 * time.Hour
	}
	return &Plugin{
		config:  config,
		seenIDs: make(map[string]bool),
	}
}

// Name はプラグイン名を返す
func (p *Plugin) Name() string {
	return "git"
}

// Grammar は GBNF grammar を返す
func (p *Plugin) Grammar() string {
	return GitGrammar
}

// Start はプラグインを起動する
func (p *Plugin) Start(ctx context.Context, core plugin.Core) error {
	p.core = core

	ctx, p.cancel = context.WithCancel(ctx)

	// 通知トリガー型ポーリング開始
	if len(p.config.Notification.Types) > 0 || p.config.Notification.PollInterval > 0 {
		p.wg.Add(1)
		go p.pollNotifications(ctx)
	}

	// バッチサーベイ型開始
	if len(p.config.Survey.Repos) > 0 {
		p.wg.Add(1)
		go p.runSurvey(ctx)
	}

	return nil
}

// Stop はプラグインを停止する
func (p *Plugin) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	return nil
}

// pollNotifications は GitHub Notifications API をポーリングする
func (p *Plugin) pollNotifications(ctx context.Context) {
	defer p.wg.Done()

	slog.Info("Git 通知ポーリング開始", "interval", p.config.Notification.PollInterval)

	// 初回は即実行
	p.fetchAndSubmitNotifications()

	ticker := time.NewTicker(p.config.Notification.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Git 通知ポーリング停止")
			return
		case <-ticker.C:
			p.fetchAndSubmitNotifications()
		}
	}
}

// fetchAndSubmitNotifications は GitHub 通知を取得して Submit する
func (p *Plugin) fetchAndSubmitNotifications() {
	notifications, err := p.fetchNotifications()
	if err != nil {
		slog.Error("GitHub 通知取得失敗", "error", err)
		return
	}

	newCount := 0
	for _, notif := range notifications {
		item, ok := p.normalizeNotification(notif)
		if !ok {
			continue
		}

		// 差分検出
		p.mu.Lock()
		if p.seenIDs[item.SourceID] {
			p.mu.Unlock()
			continue
		}
		p.seenIDs[item.SourceID] = true
		p.mu.Unlock()

		if err := p.core.Submit(item); err != nil {
			slog.Error("Item Submit 失敗", "source_id", item.SourceID, "error", err)
			continue
		}
		newCount++
	}

	if newCount > 0 {
		slog.Info("新着通知を処理", "count", newCount)
	}
	p.lastPoll = time.Now()
}

// ghNotification は GitHub API の通知レスポンス
type ghNotification struct {
	ID        string    `json:"id"`
	Reason    string    `json:"reason"`
	Unread    bool      `json:"unread"`
	UpdatedAt time.Time `json:"updated_at"`
	Subject   struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Type  string `json:"type"` // PullRequest, Issue, CheckSuite, Release, Discussion
	} `json:"subject"`
	Repository struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
}

// fetchNotifications は gh CLI で GitHub 通知を取得する
func (p *Plugin) fetchNotifications() ([]ghNotification, error) {
	cmd := exec.Command("gh", "api", "/notifications", "--paginate")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "401") || strings.Contains(stderr, "403") {
				return nil, fmt.Errorf("GitHub 認証に失敗しました: %s", stderr)
			}
			if strings.Contains(stderr, "rate limit") {
				return nil, fmt.Errorf("GitHub API レートリミット: %s", stderr)
			}
			return nil, fmt.Errorf("gh api エラー: %s", stderr)
		}
		return nil, fmt.Errorf("gh コマンド実行失敗: %w", err)
	}

	var notifications []ghNotification
	if err := json.Unmarshal(output, &notifications); err != nil {
		return nil, fmt.Errorf("通知 JSON パース失敗: %w", err)
	}

	return notifications, nil
}

// normalizeNotification は GitHub 通知を Item に正規化する
func (p *Plugin) normalizeNotification(notif ghNotification) (plugin.Item, bool) {
	// 通知タイプフィルタリング
	if len(p.config.Notification.Types) > 0 {
		matched := false
		for _, t := range p.config.Notification.Types {
			if notif.Reason == t {
				matched = true
				break
			}
		}
		if !matched {
			return plugin.Item{}, false
		}
	}

	// subject URL から HTML URL を生成
	htmlURL := convertAPIURLToHTML(notif.Subject.URL, notif.Repository.HTMLURL)

	item := plugin.Item{
		Source:    "git",
		SourceID: notif.ID,
		Title:    notif.Subject.Title,
		Content:  fmt.Sprintf("%s: %s", notif.Reason, notif.Subject.Title),
		URL:      htmlURL,
		Timestamp: notif.UpdatedAt,
		Metadata: map[string]string{
			"repo":              notif.Repository.FullName,
			"notification_type": notif.Reason,
			"subject_type":      notif.Subject.Type,
		},
	}

	return item, true
}

// convertAPIURLToHTML は GitHub API URL を HTML URL に変換する
func convertAPIURLToHTML(apiURL, repoHTMLURL string) string {
	if apiURL == "" {
		return repoHTMLURL
	}
	// https://api.github.com/repos/owner/repo/pulls/123
	// → https://github.com/owner/repo/pull/123
	apiURL = strings.Replace(apiURL, "https://api.github.com/repos/", "https://github.com/", 1)
	apiURL = strings.Replace(apiURL, "/pulls/", "/pull/", 1)
	return apiURL
}
