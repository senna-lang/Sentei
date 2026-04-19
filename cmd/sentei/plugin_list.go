/**
 * plugin list コマンド — 有効なプラグインとその設定を表示する
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/spf13/cobra"
)

func pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "プラグイン一覧を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginList()
		},
	}
}

func runPluginList() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/status", apiAddr))
	if err != nil {
		fmt.Println(cli.Error("デーモンに接続できません。sentei serve を起動してください"))
		return nil
	}
	defer resp.Body.Close()

	var status map[string]any
	json.NewDecoder(resp.Body).Decode(&status)

	plugins, ok := status["plugins"].([]any)
	if !ok || len(plugins) == 0 {
		fmt.Println("有効なプラグインがありません")
		return nil
	}

	// 設定から追加情報を取得
	cfg := loadConfig()

	fmt.Println(cli.Bold("プラグイン一覧:"))
	fmt.Println()

	for _, p := range plugins {
		name := fmt.Sprintf("%v", p)
		fmt.Printf("  %s %s\n", cli.Success("●"), cli.Bold(name))

		if name == "git" && cfg.Plugins.Git.Enabled {
			fmt.Printf("    通知ポーリング: %d秒間隔\n", cfg.Plugins.Git.PollIntervalSec)
			fmt.Printf("    サーベイ: %d秒間隔\n", cfg.Plugins.Git.SurveyIntervalSec)
			fmt.Printf("    監視リポジトリ:\n")
			for _, repo := range cfg.Plugins.Git.Repos {
				fmt.Printf("      - %s\n", repo)
			}
		}
	}

	return nil
}
