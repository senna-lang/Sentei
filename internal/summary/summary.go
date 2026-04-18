/**
 * サマリーテンプレートエンジン
 * サーベイで収集したアイテムをリポジトリ別に整形する
 * - 統計行（PR 新規/担当/マージ数）はテンプレート
 * - サマリー文は Bonsai フリー生成（外部から注入）
 * - アイテム一覧は category 別にグループ化
 */
package summary

import (
	"fmt"
	"strings"

	"github.com/senna-lang/sentei/internal/plugin"
)

var categoryIcon = map[string]string{
	"pr":         "🔀",
	"issue":      "📝",
	"ci":         "🔧",
	"release":    "🚀",
	"discussion": "💬",
	"other":      "📌",
}

var categoryOrder = []string{"pr", "issue", "ci", "release", "discussion", "other"}

// SummaryData はサマリーのレンダリングに必要なデータ
type SummaryData struct {
	Repo    string
	Date    string
	Items   []plugin.LabeledItem
	Summary string // Bonsai 生成のサマリー文（空なら非表示）
	MyUser  string // 自分のユーザー名（担当 PR の判定用）
}

// Stats はサマリーの統計情報
type Stats struct {
	NewPRs    int
	MyPRs     int
	MergedPRs int
	NewIssues int
	Releases  int
}

// CalcStats はアイテムから統計情報を算出する
// survey_type をソースオブトゥルースとし、Bonsai のカテゴリラベルが外れてもカウントは正確に保つ
func CalcStats(items []plugin.LabeledItem, myUser string) Stats {
	var s Stats
	for _, li := range items {
		surveyType := li.Item.Metadata["survey_type"]
		author := li.Item.Metadata["author"]

		switch surveyType {
		case "merged_pr":
			s.MergedPRs++
			if author == myUser || li.Item.Metadata["reviewer"] == myUser {
				s.MyPRs++
			}
		case "open_pr":
			s.NewPRs++
			if author == myUser || li.Item.Metadata["reviewer"] == myUser {
				s.MyPRs++
			}
		case "new_issue":
			s.NewIssues++
		case "release":
			s.Releases++
		default:
			switch li.Label.Category {
			case "release":
				s.Releases++
			}
		}
	}
	return s
}

// Render はサマリーテキストを生成する
func Render(data SummaryData) string {
	if len(data.Items) == 0 {
		return fmt.Sprintf("📋 %s (%s)\n  特に動きはありません\n", data.Repo, data.Date)
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("📋 %s (%s)\n", data.Repo, data.Date))

	stats := CalcStats(data.Items, data.MyUser)
	statParts := []string{}
	if stats.MergedPRs > 0 {
		statParts = append(statParts, fmt.Sprintf("マージ %d件", stats.MergedPRs))
	}
	if stats.NewPRs > 0 {
		statParts = append(statParts, fmt.Sprintf("オープン PR %d件", stats.NewPRs))
	}
	if stats.MyPRs > 0 {
		statParts = append(statParts, fmt.Sprintf("あなた担当 %d件", stats.MyPRs))
	}
	if stats.NewIssues > 0 {
		statParts = append(statParts, fmt.Sprintf("Issue %d件", stats.NewIssues))
	}
	if stats.Releases > 0 {
		statParts = append(statParts, fmt.Sprintf("リリース %d件", stats.Releases))
	}
	if len(statParts) > 0 {
		b.WriteString(fmt.Sprintf("  %s\n", strings.Join(statParts, " / ")))
	}

	if data.Summary != "" {
		b.WriteString(fmt.Sprintf("  %s\n", data.Summary))
	}

	groups := make(map[string][]plugin.LabeledItem)
	for _, li := range data.Items {
		groups[li.Label.Category] = append(groups[li.Label.Category], li)
	}

	for _, cat := range categoryOrder {
		group, ok := groups[cat]
		if !ok {
			continue
		}

		icon := categoryIcon[cat]
		if icon == "" {
			icon = "📌"
		}
		b.WriteString(fmt.Sprintf("\n  %s %s (%d)\n", icon, cat, len(group)))

		for _, li := range group {
			author := li.Item.Metadata["author"]
			line := fmt.Sprintf("    %s", li.Item.Title)
			if author != "" {
				line += fmt.Sprintf(" (@%s)", author)
			}
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}
