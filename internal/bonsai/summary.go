/**
 * リポジトリ活動のサマリー生成
 * Bonsai のフリーテキスト生成は品質不安定のため、テンプレートベースの決定的サマリーを返す
 * 入力: 1 日分の LabeledItem リスト / 出力: 1 文の日本語サマリー（空なら非表示）
 */
package bonsai

import (
	"fmt"
	"sort"
	"strings"

	"github.com/senna-lang/sentei/internal/plugin"
)

// GenerateSummary はリポジトリの活動からサマリー文を生成する
// テンプレートベース（LLM を使わない）ため、常に安定した出力を返す
// 受け手：Client をレシーバに取るのは API 互換維持のため
func (c *Client) GenerateSummary(repo string, items []plugin.LabeledItem) string {
	return BuildTemplateSummary(items)
}

// BuildTemplateSummary は items からテンプレートベースのサマリーを組み立てる
// LLM 非依存なのでテスト可能・安定動作する
func BuildTemplateSummary(items []plugin.LabeledItem) string {
	if len(items) == 0 {
		return ""
	}

	merged := 0
	newPRs := 0
	newIssues := 0
	releases := 0
	urgent := 0
	authorSet := map[string]struct{}{}

	for _, li := range items {
		surveyType := li.Item.Metadata["survey_type"]
		switch {
		case surveyType == "merged_pr":
			merged++
		case li.Label.Category == "pr":
			newPRs++
		case li.Label.Category == "issue":
			newIssues++
		case li.Label.Category == "release":
			releases++
		}
		if li.Label.Urgency == plugin.UrgencyUrgent {
			urgent++
		}
		if a := li.Item.Metadata["author"]; a != "" {
			authorSet[a] = struct{}{}
		}
	}

	var parts []string
	if releases > 0 {
		parts = append(parts, fmt.Sprintf("リリースが %d 件出ています", releases))
	}
	if merged > 0 {
		parts = append(parts, fmt.Sprintf("PR が %d 件マージされました", merged))
	}
	if newPRs > 0 {
		parts = append(parts, fmt.Sprintf("新規 PR が %d 件", newPRs))
	}
	if newIssues > 0 {
		parts = append(parts, fmt.Sprintf("Issue が %d 件起票", newIssues))
	}
	if urgent > 0 {
		parts = append(parts, fmt.Sprintf("urgent 扱いが %d 件", urgent))
	}

	if len(parts) == 0 {
		return ""
	}

	summary := joinJapanese(parts) + "。"

	// 参加者が複数なら追記（単独 or 不明は省略）
	if names := sortedAuthors(authorSet); len(names) >= 2 {
		summary += fmt.Sprintf(" 関与: %s。", strings.Join(names, ", "))
	}

	return summary
}

// joinJapanese は複数の断片を日本語として読みやすく連結する
// 例: ["A", "B", "C"] → "A、B、C"
func joinJapanese(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], "、") + "、" + parts[len(parts)-1]
}

func sortedAuthors(set map[string]struct{}) []string {
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, "@"+n)
	}
	sort.Strings(names)
	if len(names) > 5 {
		names = names[:5]
	}
	return names
}
