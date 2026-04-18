/**
 * アプリケーション設定の読み込みと管理
 * TOML 形式の設定ファイルからデーモン・Bonsai・プラグインの設定を読み込む
 */
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

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
}

// GitPluginConfig は Git プラグインの設定
type GitPluginConfig struct {
	Enabled           bool     `toml:"enabled"`
	PollIntervalSec   int      `toml:"poll_interval_sec"`
	SurveyIntervalSec int      `toml:"survey_interval_sec"`
	Repos             []string `toml:"repos"`
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
		},
	}
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
