package rss

import (
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"http to https", "http://example.com/a", "https://example.com/a"},
		{"host lowercase", "https://Example.COM/A", "https://example.com/A"},
		{"www removed", "https://www.example.com/a", "https://example.com/a"},
		{"utm removed", "https://example.com/a?utm_source=twitter&utm_campaign=foo", "https://example.com/a"},
		{"fbclid removed", "https://example.com/a?fbclid=xxx", "https://example.com/a"},
		{"trailing slash", "https://example.com/a/", "https://example.com/a"},
		{"fragment removed", "https://example.com/a#frag", "https://example.com/a"},
		{"root slash kept", "https://example.com/", "https://example.com/"},
		{"non-tracking query kept", "https://example.com/a?id=42", "https://example.com/a?id=42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeURL(tt.in)
			if got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildSourceID_GUID_URL(t *testing.T) {
	e := &gofeed.Item{GUID: "https://example.com/posts/42"}
	got := buildSourceID(e)
	if len(got) != 16 {
		t.Errorf("source_id 長さ = %d, want 16", len(got))
	}
}

func TestBuildSourceID_GUID_NonURL_FallsBackToLink(t *testing.T) {
	// GUID が URL でない場合 Link が優先される
	e1 := &gofeed.Item{GUID: "abc123-xyz", Link: "https://example.com/posts/42"}
	e2 := &gofeed.Item{Link: "https://example.com/posts/42"}

	id1 := buildSourceID(e1)
	id2 := buildSourceID(e2)

	if id1 != id2 {
		t.Errorf("GUID 非 URL の場合、Link ベースと一致するべき: %q vs %q", id1, id2)
	}
}

func TestBuildSourceID_Link_Only(t *testing.T) {
	e := &gofeed.Item{Link: "https://example.com/posts/42"}
	got := buildSourceID(e)
	if got == "" {
		t.Error("link のみで source_id が空")
	}
}

func TestBuildSourceID_Fallback_TitleAndPublished(t *testing.T) {
	e := &gofeed.Item{Title: "Announcing Claude 4", Published: "2026-04-19"}
	got := buildSourceID(e)
	if got == "" {
		t.Error("title+published フォールバックで source_id が空")
	}
}

func TestBuildSourceID_DedupsAcrossTrackingParams(t *testing.T) {
	// utm_source 違いの同じ記事が同じ source_id になること
	e1 := &gofeed.Item{Link: "https://example.com/a?utm_source=twitter"}
	e2 := &gofeed.Item{Link: "https://example.com/a?utm_source=mail"}
	if buildSourceID(e1) != buildSourceID(e2) {
		t.Error("tracking param 違いで source_id が変わってはいけない")
	}
}

func TestBuildSourceID_NilEntry(t *testing.T) {
	got := buildSourceID(nil)
	if got != "" {
		t.Errorf("nil entry で source_id = %q, want empty", got)
	}
}

func TestShortHash_Deterministic(t *testing.T) {
	a := shortHash("hello")
	b := shortHash("hello")
	if a != b {
		t.Error("same input で hash が異なる")
	}
	if strings.ContainsAny(a, "GHIJKLMNOPQRSTUVWXYZ") {
		t.Error("hex 出力は小文字 a-f のみのはず")
	}
}
