package rss

import (
	"strings"
	"testing"
)

func TestStripHTML_Plain(t *testing.T) {
	got := stripHTML("hello world", 100)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestStripHTML_BasicTags(t *testing.T) {
	in := "<p>Hello <b>world</b></p>"
	want := "Hello world"
	got := stripHTML(in, 100)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripHTML_AnchorText(t *testing.T) {
	in := `<p>See <a href="https://example.com">this link</a> for more</p>`
	want := "See this link for more"
	got := stripHTML(in, 100)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripHTML_ScriptContentRemoved(t *testing.T) {
	in := `<p>Hello</p><script>alert("xss")</script><p>world</p>`
	want := "Hello world"
	got := stripHTML(in, 100)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripHTML_StyleContentRemoved(t *testing.T) {
	in := `<style>body { color: red; }</style><p>Hello</p>`
	want := "Hello"
	got := stripHTML(in, 100)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripHTML_WhitespaceCollapse(t *testing.T) {
	in := "Hello\n\n\t  world\n"
	want := "Hello world"
	got := stripHTML(in, 100)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripHTML_Truncate(t *testing.T) {
	in := strings.Repeat("a", 1000)
	got := stripHTML(in, 400)
	// 400 文字 + "..." = 403 文字
	if len([]rune(got)) != 403 {
		t.Errorf("truncated length = %d, want 403", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncate suffix '...' が付いていない")
	}
}

func TestStripHTML_NoTruncateWhenUnder(t *testing.T) {
	in := "short"
	got := stripHTML(in, 400)
	if got != "short" {
		t.Errorf("got %q, want %q", got, "short")
	}
	if strings.HasSuffix(got, "...") {
		t.Error("truncate 不要なケースで '...' が付いた")
	}
}

func TestStripHTML_MultibyteTruncate(t *testing.T) {
	// 日本語 500 文字を 10 rune に truncate → 10 rune + "..."
	in := strings.Repeat("あ", 500)
	got := stripHTML(in, 10)
	runes := []rune(got)
	if len(runes) != 13 { // 10 + "..." (3 chars)
		t.Errorf("multibyte 切り詰め長 = %d, want 13", len(runes))
	}
}

func TestStripHTML_Empty(t *testing.T) {
	got := stripHTML("", 100)
	if got != "" {
		t.Errorf("empty 入力で %q、空文字列を期待", got)
	}
}
