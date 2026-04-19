/**
 * sentei - 常駐型の優先度ラベリングツール
 * Bonsai (1-bit LLM) による自動ラベリングで「今何を見るべきか」を明確にする
 */
package main

import (
	"os"

	"github.com/senna-lang/sentei/internal/config"
	"github.com/spf13/cobra"
)

// Version はバイナリのバージョン番号 (User-Agent 等で使用)
const Version = "0.1.0"

var (
	cfgFile string
	apiAddr string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sentei",
		Short: "常駐型の優先度ラベリングツール",
		Long:  "Bonsai (1-bit LLM) による自動ラベリングで「今何を見るべきか」を明確にする",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config ファイルパス (デフォルト: ~/.config/sentei/config.toml)")
	rootCmd.PersistentFlags().StringVar(&apiAddr, "addr", "127.0.0.1:7890", "API サーバーアドレス")

	// plugin サブコマンドグループ
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "プラグイン管理",
	}
	pluginCmd.AddCommand(pluginListCmd())

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(summaryCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(pluginCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig は設定ファイルを読み込む。ファイルがなければデフォルト値を返す
func loadConfig() config.AppConfig {
	path := cfgFile
	if path == "" {
		path = config.DefaultPath()
	}

	cfg, err := config.Load(path)
	if err != nil {
		// ファイルがない場合はデフォルト設定を使用
		return config.DefaultConfig()
	}

	// --addr フラグが明示的に指定されていたら上書き
	if apiAddr != "127.0.0.1:7890" {
		cfg.Daemon.Addr = apiAddr
	}

	return cfg
}
