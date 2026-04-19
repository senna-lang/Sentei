/**
 * summary コマンド — リポジトリ別サマリーを表示する
 * 引数なしで全リポジトリ、引数ありで特定リポジトリのサマリーを表示する
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/spf13/cobra"
)

func summaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary [repo]",
		Short: "リポジトリ別サマリーを表示する",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return runSummaryRepo(args[0])
			}
			return runSummaryAll()
		},
	}
}

func runSummaryAll() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/summary", apiAddr))
	if err != nil {
		fmt.Println(cli.Error("デーモンに接続できません。sentei serve を起動してください"))
		return nil
	}
	defer resp.Body.Close()

	var summaries []map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&summaries); err != nil {
		return fmt.Errorf("レスポンスのパース失敗: %w", err)
	}

	if len(summaries) == 0 {
		fmt.Println("サマリーがまだ生成されていません。デーモン起動後、最初のサーベイ完了をお待ちください")
		return nil
	}

	for i, s := range summaries {
		fmt.Print(s["summary"])
		if i < len(summaries)-1 {
			fmt.Println()
		}
	}

	return nil
}

func runSummaryRepo(repo string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/summary/%s", apiAddr, repo))
	if err != nil {
		fmt.Println(cli.Error("デーモンに接続できません。sentei serve を起動してください"))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("リポジトリ %s のサマリーがありません\n", repo)
		return nil
	}

	var summary map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return fmt.Errorf("レスポンスのパース失敗: %w", err)
	}

	fmt.Print(summary["summary"])
	return nil
}
