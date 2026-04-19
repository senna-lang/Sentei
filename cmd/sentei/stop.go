/**
 * stop コマンド — 稼働中のデーモンを停止する
 * POST /api/shutdown でグレースフルシャットダウンを要求する
 */
package main

import (
	"fmt"
	"net/http"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/spf13/cobra"
)

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "デーモンを停止する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop()
		},
	}
}

func runStop() error {
	resp, err := http.Post(fmt.Sprintf("http://%s/api/shutdown", apiAddr), "", nil)
	if err != nil {
		fmt.Println("デーモンは起動していません")
		return nil
	}
	defer resp.Body.Close()

	fmt.Println(cli.Success("デーモンを停止しました"))
	return nil
}
