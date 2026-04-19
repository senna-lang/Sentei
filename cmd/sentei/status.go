/**
 * status コマンド — デーモンの動作状態と有効プラグインの情報を表示する
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "デーモンの動作状態を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func runStatus() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/status", apiAddr))
	if err != nil {
		fmt.Println(cli.Error("デーモンは停止中です"))
		return nil
	}
	defer resp.Body.Close()

	var status map[string]any
	json.NewDecoder(resp.Body).Decode(&status)

	// Bonsai 接続状態の色分け
	bonsaiStatus := fmt.Sprintf("%v", status["bonsai"])
	if bonsaiStatus == "ok" {
		bonsaiStatus = cli.Success("ok")
	} else {
		bonsaiStatus = cli.Error(bonsaiStatus)
	}

	fmt.Printf("デーモン状態: %s\n", cli.Success("稼働中"))
	fmt.Printf("Bonsai 接続: %s\n", bonsaiStatus)
	fmt.Printf("有効プラグイン: %v\n", status["plugins"])
	fmt.Printf("保存済みアイテム数: %v 件\n", status["item_count"])

	if lastLabeled, ok := status["last_labeled"]; ok && lastLabeled != nil {
		fmt.Printf("最終ラベリング: %v\n", lastLabeled)
	}
	return nil
}
