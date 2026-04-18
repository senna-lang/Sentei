/**
 * サマリーテンプレートエンジンのテスト
 */
package summary

import (
	"strings"
	"testing"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

func TestRender_EmptyItems(t *testing.T) {
	result := Render(SummaryData{
		Repo: "senna-lang/test-repo",
		Date: "2026-04-17",
	})
	if !strings.Contains(result, "特に動きはありません") {
		t.Errorf("空アイテムで '特に動きはありません' が含まれていない: %s", result)
	}
	if !strings.Contains(result, "test-repo") {
		t.Errorf("リポジトリ名が含まれていない: %s", result)
	}
}

func TestRender_GroupedByCategory(t *testing.T) {
	items := []plugin.LabeledItem{
		{
			Item: plugin.Item{
				Source: "git", SourceID: "1", Title: "#7 refactor: scripts リアーキ",
				Timestamp: time.Now(),
				Metadata:  map[string]string{"author": "senna", "survey_type": "merged_pr"},
			},
			Label: plugin.Label{Category: "pr"},
		},
		{
			Item: plugin.Item{
				Source: "git", SourceID: "2", Title: "#46 SPECTER2 対応",
				Timestamp: time.Now(),
				Metadata:  map[string]string{"author": "researcher", "survey_type": "new_issue"},
			},
			Label: plugin.Label{Category: "issue"},
		},
		{
			Item: plugin.Item{
				Source: "git", SourceID: "3", Title: "#6 CI workflow 追加",
				Timestamp: time.Now(),
				Metadata:  map[string]string{"author": "senna", "survey_type": "merged_pr"},
			},
			Label: plugin.Label{Category: "ci"},
		},
	}

	result := Render(SummaryData{
		Repo:  "senna-lang/arxiv-compass",
		Date:  "2026-04-17",
		Items: items,
	})

	prIdx := strings.Index(result, "pr (")
	issueIdx := strings.Index(result, "issue (")
	ciIdx := strings.Index(result, "ci (")

	if prIdx >= issueIdx {
		t.Error("pr が issue より後に表示されている")
	}
	if issueIdx >= ciIdx {
		t.Error("issue が ci より後に表示されている")
	}
}

func TestRender_ContainsItemDetails(t *testing.T) {
	items := []plugin.LabeledItem{
		{
			Item: plugin.Item{
				Source: "git", SourceID: "1", Title: "#45 Fix score calculation",
				Timestamp: time.Now(),
				Metadata:  map[string]string{"author": "contributor", "survey_type": "merged_pr"},
			},
			Label: plugin.Label{Category: "pr"},
		},
	}

	result := Render(SummaryData{
		Repo:  "test-repo",
		Date:  "2026-04-17",
		Items: items,
	})

	if !strings.Contains(result, "Fix score calculation") {
		t.Error("title が含まれていない")
	}
	if !strings.Contains(result, "@contributor") {
		t.Error("author が含まれていない")
	}
}

func TestRender_StatsLine(t *testing.T) {
	items := []plugin.LabeledItem{
		{
			Item:  plugin.Item{Source: "git", SourceID: "1", Title: "PR1", Timestamp: time.Now(), Metadata: map[string]string{"survey_type": "merged_pr", "author": "senna"}},
			Label: plugin.Label{Category: "pr"},
		},
		{
			Item:  plugin.Item{Source: "git", SourceID: "2", Title: "PR2", Timestamp: time.Now(), Metadata: map[string]string{"survey_type": "merged_pr", "author": "other"}},
			Label: plugin.Label{Category: "pr"},
		},
		{
			Item:  plugin.Item{Source: "git", SourceID: "3", Title: "Issue1", Timestamp: time.Now(), Metadata: map[string]string{"survey_type": "new_issue", "author": "user"}},
			Label: plugin.Label{Category: "issue"},
		},
	}

	result := Render(SummaryData{
		Repo:   "test-repo",
		Date:   "2026-04-17",
		Items:  items,
		MyUser: "senna",
	})

	if !strings.Contains(result, "マージ 2件") {
		t.Errorf("マージ数が含まれていない: %s", result)
	}
	if !strings.Contains(result, "あなた担当 1件") {
		t.Errorf("担当数が含まれていない: %s", result)
	}
	if !strings.Contains(result, "Issue 1件") {
		t.Errorf("Issue 数が含まれていない: %s", result)
	}
}

func TestRender_WithSummary(t *testing.T) {
	items := []plugin.LabeledItem{
		{
			Item:  plugin.Item{Source: "git", SourceID: "1", Title: "PR1", Timestamp: time.Now(), Metadata: map[string]string{"survey_type": "merged_pr"}},
			Label: plugin.Label{Category: "pr"},
		},
	}

	result := Render(SummaryData{
		Repo:    "test-repo",
		Date:    "2026-04-17",
		Items:   items,
		Summary: "scripts/ のリアーキが完了しました",
	})

	if !strings.Contains(result, "scripts/ のリアーキが完了しました") {
		t.Error("Bonsai サマリーが含まれていない")
	}
}

func TestRender_SkipsEmptyCategoryGroups(t *testing.T) {
	items := []plugin.LabeledItem{
		{
			Item:  plugin.Item{Source: "git", SourceID: "1", Title: "test", Timestamp: time.Now(), Metadata: map[string]string{}},
			Label: plugin.Label{Category: "ci"},
		},
	}

	result := Render(SummaryData{Repo: "repo", Date: "2026-04-17", Items: items})

	if strings.Contains(result, "pr (") {
		t.Error("アイテムがない pr セクションが表示されている")
	}
	if !strings.Contains(result, "ci (") {
		t.Error("アイテムがある ci セクションが表示されていない")
	}
}

func TestCalcStats(t *testing.T) {
	items := []plugin.LabeledItem{
		{Item: plugin.Item{Metadata: map[string]string{"survey_type": "merged_pr", "author": "senna"}}, Label: plugin.Label{Category: "pr"}},
		{Item: plugin.Item{Metadata: map[string]string{"survey_type": "merged_pr", "author": "other"}}, Label: plugin.Label{Category: "pr"}},
		{Item: plugin.Item{Metadata: map[string]string{"survey_type": "new_issue", "author": "user"}}, Label: plugin.Label{Category: "issue"}},
		{Item: plugin.Item{Metadata: map[string]string{"survey_type": "", "author": ""}}, Label: plugin.Label{Category: "release"}},
	}

	stats := CalcStats(items, "senna")

	if stats.MergedPRs != 2 {
		t.Errorf("MergedPRs = %d, want 2", stats.MergedPRs)
	}
	if stats.MyPRs != 1 {
		t.Errorf("MyPRs = %d, want 1", stats.MyPRs)
	}
	if stats.NewIssues != 1 {
		t.Errorf("NewIssues = %d, want 1", stats.NewIssues)
	}
	if stats.Releases != 1 {
		t.Errorf("Releases = %d, want 1", stats.Releases)
	}
}
