package notify

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

func newTestItem(title, category string, urgency plugin.Urgency) plugin.LabeledItem {
	return plugin.LabeledItem{
		Item: plugin.Item{
			Source:    "git",
			SourceID:  "test-1",
			Title:     title,
			Timestamp: time.Now(),
		},
		Label: plugin.Label{
			Urgency:  urgency,
			Category: category,
		},
		LabeledAt: time.Now(),
	}
}

func TestDarwinNotifier_BuildCommand_ContainsTitle(t *testing.T) {
	n := &DarwinNotifier{}
	item := newTestItem("メンターからのレビュー依頼", "pr", plugin.UrgencyUrgent)

	args := n.BuildCommand(item)

	if args[0] != "osascript" {
		t.Errorf("command should be osascript, got %q", args[0])
	}
	if args[1] != "-e" {
		t.Errorf("second arg should be -e, got %q", args[1])
	}
	if !strings.Contains(args[2], "sentei") {
		t.Error("script should contain title 'sentei'")
	}
	if !strings.Contains(args[2], "[pr]") {
		t.Error("script should contain category [pr]")
	}
	if !strings.Contains(args[2], "メンターからのレビュー依頼") {
		t.Error("script should contain item title")
	}
}

func TestDarwinNotifier_BuildCommand_EscapesQuotes(t *testing.T) {
	n := &DarwinNotifier{}
	item := newTestItem(`Fix "critical" bug`, "issue", plugin.UrgencyUrgent)

	args := n.BuildCommand(item)

	// ダブルクォートがエスケープされていること
	if strings.Contains(args[2], `"critical"`) {
		t.Error("quotes in title should be escaped")
	}
	if !strings.Contains(args[2], `\"critical\"`) {
		t.Error("quotes should be escaped with backslash")
	}
}

func TestNoopNotifier_DoesNotError(t *testing.T) {
	n := &NoopNotifier{}
	item := newTestItem("test", "pr", plugin.UrgencyUrgent)

	err := n.Notify(item)
	if err != nil {
		t.Errorf("NoopNotifier should not return error, got %v", err)
	}
}

func TestNewPlatformNotifier_ReturnsCorrectType(t *testing.T) {
	n := NewPlatformNotifier()

	switch runtime.GOOS {
	case "darwin":
		if _, ok := n.(*DarwinNotifier); !ok {
			t.Errorf("darwin で DarwinNotifier を期待、got %T", n)
		}
	default:
		if _, ok := n.(*NoopNotifier); !ok {
			t.Errorf("非 darwin で NoopNotifier を期待、got %T", n)
		}
	}
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`hello`, `hello`},
		{`say "hi"`, `say \"hi\"`},
		{`path\to\file`, `path\\to\\file`},
	}

	for _, tt := range tests {
		got := escapeAppleScript(tt.input)
		if got != tt.want {
			t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
