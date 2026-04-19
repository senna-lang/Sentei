/**
 * list コマンド — 保存済みアイテムを色付きで一覧表示する
 * urgency / source / category でフィルタ可能
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/senna-lang/sentei/internal/plugin"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	var urgency, source, category string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "保存済みアイテムを一覧表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(urgency, source, category)
		},
	}

	cmd.Flags().StringVar(&urgency, "urgency", "", "urgency フィルタ (urgent, should_check, can_wait, ignore)")
	cmd.Flags().StringVar(&source, "source", "", "source フィルタ (例: git)")
	cmd.Flags().StringVar(&category, "category", "", "category フィルタ (例: pr, issue, ci)")

	return cmd
}

func runList(urgency, source, category string) error {
	// クエリパラメータ構築
	params := url.Values{}
	if urgency != "" {
		params.Set("urgency", urgency)
	}
	if source != "" {
		params.Set("source", source)
	}
	if category != "" {
		params.Set("category", category)
	}

	apiURL := fmt.Sprintf("http://%s/api/items", apiAddr)
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println(cli.Error("デーモンに接続できません。sentei serve を起動してください"))
		return nil
	}
	defer resp.Body.Close()

	var items []plugin.LabeledItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return fmt.Errorf("レスポンスのパース失敗: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("アイテムがありません")
		return nil
	}

	// ヘッダー
	fmt.Printf("%-14s %-10s %-6s %s\n",
		cli.Bold("URGENCY"),
		cli.Bold("CATEGORY"),
		cli.Bold("SOURCE"),
		cli.Bold("TITLE"),
	)
	fmt.Println(strings.Repeat("─", 72))

	// アイテム一覧
	for _, item := range items {
		urgencyDisplay := string(item.Label.Urgency)
		if urgencyDisplay == "" {
			urgencyDisplay = "-" // urgency 未使用プラグイン (RSS 等)
		} else {
			urgencyDisplay = cli.FormatUrgency(item.Label.Urgency)
		}
		urgencyStr := fmt.Sprintf("%-14s", urgencyDisplay)
		categoryStr := fmt.Sprintf("%-10s", item.Label.Category)
		sourceStr := fmt.Sprintf("%-6s", item.Item.Source)

		title := item.Item.Title
		if author, ok := item.Item.Metadata["author"]; ok && author != "" {
			title += fmt.Sprintf(" (@%s)", author)
		}

		fmt.Printf("%s %s %s %s\n", urgencyStr, categoryStr, sourceStr, title)
	}

	fmt.Printf("\n%d 件\n", len(items))
	return nil
}
