/**
 * アプリケーション設定の読み込みと管理
 * TOML 形式の設定ファイルからデーモン・Bonsai・プラグインの設定を読み込む
 */
package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// 有効な urgency_floor 値の集合 (空文字列は「floor 無指定」を意味するので OK)
var validUrgencyFloors = map[string]struct{}{
	"":             {},
	"ignore":       {},
	"can_wait":     {},
	"should_check": {},
	"urgent":       {},
}

// AppConfig はアプリケーション全体の設定
type AppConfig struct {
	Daemon  DaemonConfig  `toml:"daemon"`
	Bonsai  BonsaiConfig  `toml:"bonsai"`
	Plugins PluginsConfig `toml:"plugins"`
}

// DaemonConfig はデーモンの設定
type DaemonConfig struct {
	Addr   string `toml:"addr"`
	DBPath string `toml:"db_path"`
}

// BonsaiConfig は Bonsai ラベリングエンジンの接続設定
type BonsaiConfig struct {
	URL string `toml:"url"`
}

// PluginsConfig はプラグイン群の設定
type PluginsConfig struct {
	Git GitPluginConfig `toml:"git"`
	Rss RssConfig       `toml:"rss"`
}

// GitPluginConfig は Git プラグインの設定
type GitPluginConfig struct {
	Enabled           bool     `toml:"enabled"`
	PollIntervalSec   int      `toml:"poll_interval_sec"`
	SurveyIntervalSec int      `toml:"survey_interval_sec"`
	Repos             []string `toml:"repos"`
}

// RssConfig は RSS プラグインの設定
type RssConfig struct {
	Enabled         bool         `toml:"enabled"`
	PollIntervalSec int          `toml:"poll_interval_sec"`
	Feeds           []FeedConfig `toml:"feeds"`
}

// FeedConfig は個別フィードの設定
type FeedConfig struct {
	URL          string `toml:"url"`
	Name         string `toml:"name"`
	UrgencyFloor string `toml:"urgency_floor"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() AppConfig {
	return AppConfig{
		Daemon: DaemonConfig{
			Addr:   "127.0.0.1:7890",
			DBPath: "~/.config/sentei/db.sqlite",
		},
		Bonsai: BonsaiConfig{
			URL: "http://127.0.0.1:8080",
		},
		Plugins: PluginsConfig{
			Git: GitPluginConfig{
				Enabled:           true,
				PollIntervalSec:   60,
				SurveyIntervalSec: 3600,
				Repos: []string{
					"senna-lang/arxiv-compass",
					"senna-lang/Logosyncx",
					"senna-lang/bonsai-TRM",
				},
			},
			Rss: RssConfig{
				Enabled:         false,
				PollIntervalSec: 900,
				Feeds:           nil,
			},
		},
	}
}

// validate は設定値のサニティチェックを実施する。致命的な不備は error、
// 推奨外の値は warn ログに落として許容する。
func validate(cfg *AppConfig) error {
	// RSS の validation
	for i, fc := range cfg.Plugins.Rss.Feeds {
		if !strings.HasPrefix(fc.URL, "http://") && !strings.HasPrefix(fc.URL, "https://") {
			return fmt.Errorf("plugins.rss.feeds[%d].url は http(s):// で始まる必要があります: %q", i, fc.URL)
		}
		if _, ok := validUrgencyFloors[fc.UrgencyFloor]; !ok {
			return fmt.Errorf("plugins.rss.feeds[%d].urgency_floor は ignore/can_wait/should_check/urgent のいずれか (または未指定) である必要があります: %q", i, fc.UrgencyFloor)
		}
	}
	if cfg.Plugins.Rss.Enabled && cfg.Plugins.Rss.PollIntervalSec > 0 && cfg.Plugins.Rss.PollIntervalSec < 60 {
		slog.Warn("plugins.rss.poll_interval_sec が 60 秒未満です。フィード元への負荷を考慮してください",
			"value", cfg.Plugins.Rss.PollIntervalSec)
	}
	return nil
}

// Load は TOML ファイルから設定を読み込む
func Load(path string) (AppConfig, error) {
	expanded := ExpandPath(path)

	data, err := os.ReadFile(expanded)
	if err != nil {
		return AppConfig{}, fmt.Errorf("設定ファイル読み込み失敗: %w", err)
	}

	cfg := DefaultConfig()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("TOML パース失敗: %w", err)
	}

	// DBPath の ~ 展開
	cfg.Daemon.DBPath = ExpandPath(cfg.Daemon.DBPath)

	if err := validate(&cfg); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}

// DefaultTOML はデフォルト設定を TOML 文字列で返す（init コマンド用）
func DefaultTOML() string {
	return `[daemon]
addr = "127.0.0.1:7890"
db_path = "~/.config/sentei/db.sqlite"

[bonsai]
url = "http://127.0.0.1:8080"

[plugins.git]
enabled = true
poll_interval_sec = 60
survey_interval_sec = 3600
repos = [
  "senna-lang/arxiv-compass",
  "senna-lang/Logosyncx",
  "senna-lang/bonsai-TRM",
]

# RSS プラグインは opt-in。enabled = true にして使用する
# RSS は category 分類のみ (urgency は使わない)。urgency_floor は設定可能だが
# 現状 RSS では no-op (将来別プラグイン用に残している)
# (Anthropic は公式 RSS 提供なしのため含めていない。代替が見つかったら追記)
[plugins.rss]
enabled = false
poll_interval_sec = 900

# 高信号 (固定)
[[plugins.rss.feeds]]
url = "https://lilianweng.github.io/index.xml"
name = "Lil'Log (Lilian Weng)"

[[plugins.rss.feeds]]
url = "https://simonwillison.net/atom/everything/"
name = "Simon Willison"

# Zenn topic
[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/claudecode/feed"
name = "Zenn - Claude Code"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/typescript/feed"
name = "Zenn - TypeScript"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/ai/feed"
name = "Zenn - AI"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/llm/feed"
name = "Zenn - LLM"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/react/feed"
name = "Zenn - React"

# Qiita tag
[[plugins.rss.feeds]]
url = "https://qiita.com/tags/typescript/feed"
name = "Qiita - TypeScript"

[[plugins.rss.feeds]]
url = "https://qiita.com/tags/react/feed"
name = "Qiita - React"

[[plugins.rss.feeds]]
url = "https://qiita.com/tags/llm/feed"
name = "Qiita - LLM"
`
}

// ConfigDir は設定ディレクトリのパスを返す
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sentei")
}

// DefaultPath はデフォルトの設定ファイルパスを返す
func DefaultPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// ExpandPath はパス中の ~ をホームディレクトリに展開する
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
