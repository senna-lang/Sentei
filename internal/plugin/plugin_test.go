/**
 * Plugin 型のテスト
 */
package plugin

import (
	"testing"
	"time"
)

func TestUrgencyConstants(t *testing.T) {
	tests := []struct {
		urgency Urgency
		want    string
	}{
		{UrgencyUrgent, "urgent"},
		{UrgencyShouldCheck, "should_check"},
		{UrgencyCanWait, "can_wait"},
		{UrgencyIgnore, "ignore"},
	}

	for _, tt := range tests {
		if string(tt.urgency) != tt.want {
			t.Errorf("Urgency = %q, want %q", tt.urgency, tt.want)
		}
	}
}

func TestItemValidFields(t *testing.T) {
	item := Item{
		Source:    "git",
		SourceID: "notif-123",
		Title:    "Review request",
		Content:  "PR description",
		URL:      "https://github.com/...",
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"repo":              "arxiv-compass",
			"notification_type": "review_requested",
		},
	}

	if item.Source != "git" {
		t.Errorf("Source = %q, want %q", item.Source, "git")
	}
	if item.Metadata["repo"] != "arxiv-compass" {
		t.Errorf("Metadata[repo] = %q, want %q", item.Metadata["repo"], "arxiv-compass")
	}
}

func TestLabeledItem(t *testing.T) {
	li := LabeledItem{
		Item: Item{
			Source:    "git",
			SourceID: "1",
			Title:    "test",
			Timestamp: time.Now(),
		},
		Label: Label{
			Urgency:  UrgencyUrgent,
			Category: "pr",
			Summary:  "test summary",
		},
		LabeledAt: time.Now(),
	}

	if li.Label.Urgency != UrgencyUrgent {
		t.Errorf("Urgency = %q, want %q", li.Label.Urgency, UrgencyUrgent)
	}
	if li.Label.Category != "pr" {
		t.Errorf("Category = %q, want %q", li.Label.Category, "pr")
	}
}
