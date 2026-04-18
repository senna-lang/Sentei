/**
 * Git プラグインのテスト
 */
package git

import (
	"strings"
	"testing"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

// mockCore は Submit を記録するモック
type mockCore struct {
	submitted []plugin.Item
}

func (m *mockCore) Submit(item plugin.Item) error {
	m.submitted = append(m.submitted, item)
	return nil
}

func TestPluginName(t *testing.T) {
	p := NewPlugin(Config{})
	if p.Name() != "git" {
		t.Errorf("Name() = %q, want %q", p.Name(), "git")
	}
}

func TestPluginGrammar(t *testing.T) {
	p := NewPlugin(Config{})
	grammar := p.Grammar()
	if grammar == "" {
		t.Error("Grammar() が空")
	}
	if !strings.Contains(grammar, "urgent") {
		t.Error("Grammar に urgent が含まれていない")
	}
	if !strings.Contains(grammar, `\"pr\"`) {
		t.Error("Grammar に pr が含まれていない")
	}
}

func TestNormalizeNotification_Basic(t *testing.T) {
	p := NewPlugin(Config{})

	notif := ghNotification{
		ID:        "123",
		Reason:    "review_requested",
		Unread:    true,
		UpdatedAt: time.Now(),
	}
	notif.Subject.Title = "fix: score calculation bug"
	notif.Subject.Type = "PullRequest"
	notif.Subject.URL = "https://api.github.com/repos/senna-lang/arxiv-compass/pulls/47"
	notif.Repository.FullName = "senna-lang/arxiv-compass"
	notif.Repository.HTMLURL = "https://github.com/senna-lang/arxiv-compass"

	item, ok := p.normalizeNotification(notif)
	if !ok {
		t.Fatal("normalizeNotification returned false")
	}

	if item.Source != "git" {
		t.Errorf("Source = %q, want %q", item.Source, "git")
	}
	if item.SourceID != "123" {
		t.Errorf("SourceID = %q, want %q", item.SourceID, "123")
	}
	if item.Title != "fix: score calculation bug" {
		t.Errorf("Title = %q, want %q", item.Title, "fix: score calculation bug")
	}
	if item.Metadata["repo"] != "senna-lang/arxiv-compass" {
		t.Errorf("Metadata[repo] = %q, want %q", item.Metadata["repo"], "senna-lang/arxiv-compass")
	}
	if item.Metadata["notification_type"] != "review_requested" {
		t.Errorf("Metadata[notification_type] = %q, want %q", item.Metadata["notification_type"], "review_requested")
	}
	if item.URL != "https://github.com/senna-lang/arxiv-compass/pull/47" {
		t.Errorf("URL = %q, want HTML URL", item.URL)
	}
}

func TestNormalizeNotification_TypeFilter(t *testing.T) {
	p := NewPlugin(Config{
		Notification: NotificationConfig{
			Types: []string{"review_requested", "mentioned"},
		},
	})

	// review_requested → 通過
	notif := ghNotification{ID: "1", Reason: "review_requested", UpdatedAt: time.Now()}
	notif.Subject.Title = "test"
	_, ok := p.normalizeNotification(notif)
	if !ok {
		t.Error("review_requested がフィルタで除外された")
	}

	// subscribed → 除外
	notif2 := ghNotification{ID: "2", Reason: "subscribed", UpdatedAt: time.Now()}
	notif2.Subject.Title = "test"
	_, ok = p.normalizeNotification(notif2)
	if ok {
		t.Error("subscribed がフィルタを通過した")
	}
}

func TestNormalizeNotification_NoFilter(t *testing.T) {
	p := NewPlugin(Config{})

	notif := ghNotification{ID: "1", Reason: "subscribed", UpdatedAt: time.Now()}
	notif.Subject.Title = "test"
	_, ok := p.normalizeNotification(notif)
	if !ok {
		t.Error("フィルタなしで subscribed が除外された")
	}
}

func TestConvertAPIURLToHTML(t *testing.T) {
	tests := []struct {
		apiURL  string
		repoURL string
		want    string
	}{
		{
			"https://api.github.com/repos/senna-lang/arxiv-compass/pulls/47",
			"https://github.com/senna-lang/arxiv-compass",
			"https://github.com/senna-lang/arxiv-compass/pull/47",
		},
		{
			"https://api.github.com/repos/senna-lang/test/issues/10",
			"https://github.com/senna-lang/test",
			"https://github.com/senna-lang/test/issues/10",
		},
		{
			"",
			"https://github.com/senna-lang/test",
			"https://github.com/senna-lang/test",
		},
	}

	for _, tt := range tests {
		got := convertAPIURLToHTML(tt.apiURL, tt.repoURL)
		if got != tt.want {
			t.Errorf("convertAPIURLToHTML(%q, %q) = %q, want %q", tt.apiURL, tt.repoURL, got, tt.want)
		}
	}
}

func TestDuplicateDetection(t *testing.T) {
	p := NewPlugin(Config{})
	core := &mockCore{}
	p.core = core

	notif := ghNotification{ID: "dup-1", Reason: "mentioned", UpdatedAt: time.Now()}
	notif.Subject.Title = "test"

	item, _ := p.normalizeNotification(notif)

	// 1 回目: 新規
	p.mu.Lock()
	p.seenIDs[item.SourceID] = true
	p.mu.Unlock()

	// 2 回目: 重複
	p.mu.Lock()
	seen := p.seenIDs[item.SourceID]
	p.mu.Unlock()

	if !seen {
		t.Error("重複検出が機能していない")
	}
}

func TestDefaultConfig(t *testing.T) {
	p := NewPlugin(Config{})

	if p.config.Notification.PollInterval != 1*time.Minute {
		t.Errorf("デフォルト PollInterval = %v, want 1m", p.config.Notification.PollInterval)
	}
	if p.config.Survey.Interval != 1*time.Hour {
		t.Errorf("デフォルト Survey.Interval = %v, want 1h", p.config.Survey.Interval)
	}
}

