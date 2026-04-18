/**
 * Bonsai ラベリングクライアントのテスト
 */
package bonsai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/senna-lang/sentei/internal/plugin"
)

func TestParseLabel_ValidJSON(t *testing.T) {
	input := `{"urgency": "urgent", "category": "pr", "summary": "Review requested"}`
	label, err := parseLabel(input)
	if err != nil {
		t.Fatalf("parseLabel() error = %v", err)
	}
	if label.Urgency != plugin.UrgencyUrgent {
		t.Errorf("Urgency = %q, want %q", label.Urgency, plugin.UrgencyUrgent)
	}
	if label.Category != "pr" {
		t.Errorf("Category = %q, want %q", label.Category, "pr")
	}
	if label.Summary != "Review requested" {
		t.Errorf("Summary = %q, want %q", label.Summary, "Review requested")
	}
}

func TestParseLabel_MissingClosingBrace(t *testing.T) {
	// stop で } が消えるケース
	input := `{"urgency": "should_check", "category": "ci", "summary": "CI failed"`
	label, err := parseLabel(input)
	if err != nil {
		t.Fatalf("parseLabel() error = %v", err)
	}
	if label.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("Urgency = %q, want %q", label.Urgency, plugin.UrgencyShouldCheck)
	}
	if label.Category != "ci" {
		t.Errorf("Category = %q, want %q", label.Category, "ci")
	}
}

func TestParseLabel_InvalidJSON(t *testing.T) {
	input := `not json at all`
	label, err := parseLabel(input)
	if err == nil {
		t.Error("parseLabel() should return error for invalid JSON")
	}
	// フォールバック値が返る
	if label.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("fallback Urgency = %q, want %q", label.Urgency, plugin.UrgencyShouldCheck)
	}
	if label.Category != "other" {
		t.Errorf("fallback Category = %q, want %q", label.Category, "other")
	}
}

func TestFallbackLabel(t *testing.T) {
	label := fallbackLabel()
	if label.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("Urgency = %q, want %q", label.Urgency, plugin.UrgencyShouldCheck)
	}
	if label.Category != "other" {
		t.Errorf("Category = %q, want %q", label.Category, "other")
	}
}

func TestBuildPrompt(t *testing.T) {
	item := plugin.Item{
		Source: "git",
		Title:  "Fix bug",
		Metadata: map[string]string{
			"repo": "test-repo",
		},
	}
	template := "Classify: {notification_json}"
	result := buildPrompt(item, template)

	if result == template {
		t.Error("プロンプトが置換されていない")
	}
	if len(result) <= len("Classify: ") {
		t.Error("プロンプトに item JSON が含まれていない")
	}
}

func TestClient_Label_WithMockServer(t *testing.T) {
	// モック Bonsai サーバー
	mockResp := completionResponse{
		Content: `{"urgency": "urgent", "category": "pr", "summary": "Review needed"}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/completion" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	item := plugin.Item{
		Source:   "git",
		SourceID: "1",
		Title:    "PR review request",
		Metadata: map[string]string{"notification_type": "review_requested"},
	}

	label, err := client.Label(item, "mock-grammar", "Classify: {notification_json}")
	if err != nil {
		t.Fatalf("Label() error = %v", err)
	}
	if label.Urgency != plugin.UrgencyUrgent {
		t.Errorf("Urgency = %q, want %q", label.Urgency, plugin.UrgencyUrgent)
	}
	if label.Category != "pr" {
		t.Errorf("Category = %q, want %q", label.Category, "pr")
	}
}

func TestClient_Label_ServerDown(t *testing.T) {
	client := NewClient("http://127.0.0.1:1") // 存在しないポート
	client.maxRetries = 1                      // テスト高速化

	item := plugin.Item{
		Source:   "git",
		SourceID: "1",
		Title:    "test",
		Metadata: map[string]string{},
	}

	label, err := client.Label(item, "grammar", "prompt {notification_json}")
	if err == nil {
		t.Error("Label() should return error when server is down")
	}
	// フォールバックが返る
	if label.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("fallback Urgency = %q, want %q", label.Urgency, plugin.UrgencyShouldCheck)
	}
}

func TestClient_Ping_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
