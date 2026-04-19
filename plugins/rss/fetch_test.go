package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/senna-lang/sentei/internal/config"
)

const rssSample = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Example RSS</title>
    <item>
      <title>Post 1</title>
      <link>https://example.com/posts/1</link>
      <guid>https://example.com/posts/1</guid>
      <pubDate>Sat, 19 Apr 2026 10:00:00 +0000</pubDate>
      <description>Body 1</description>
    </item>
  </channel>
</rss>`

const atomSample = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Example Atom</title>
  <entry>
    <title>Atom Post</title>
    <link href="https://example.com/atom/1"/>
    <id>https://example.com/atom/1</id>
    <published>2026-04-19T10:00:00Z</published>
    <summary>Summary</summary>
  </entry>
</feed>`

func newTestFetcher() *Fetcher {
	return NewFetcher("sentei-test/0.0")
}

func TestFetch_RSS2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "sentei-test/0.0" {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(rssSample))
	}))
	defer srv.Close()

	f := newTestFetcher()
	feed, err := f.Fetch(context.Background(), config.FeedConfig{URL: srv.URL})
	if err != nil {
		t.Fatalf("Fetch error = %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(feed.Items))
	}
	if feed.Items[0].Title != "Post 1" {
		t.Errorf("item title = %q, want %q", feed.Items[0].Title, "Post 1")
	}
}

func TestFetch_Atom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(atomSample))
	}))
	defer srv.Close()

	f := newTestFetcher()
	feed, err := f.Fetch(context.Background(), config.FeedConfig{URL: srv.URL})
	if err != nil {
		t.Fatalf("Fetch error = %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(feed.Items))
	}
	if feed.Items[0].Title != "Atom Post" {
		t.Errorf("title = %q", feed.Items[0].Title)
	}
}

func TestFetch_404_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := newTestFetcher()
	_, err := f.Fetch(context.Background(), config.FeedConfig{URL: srv.URL})
	if err == nil {
		t.Fatal("404 でエラーを期待")
	}
	if IsRateLimit(err) {
		t.Error("404 が RateLimit として扱われている")
	}
}

func TestFetch_429_ReturnsRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "300")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	f := newTestFetcher()
	_, err := f.Fetch(context.Background(), config.FeedConfig{URL: srv.URL})
	if err == nil {
		t.Fatal("429 でエラーを期待")
	}
	if !IsRateLimit(err) {
		t.Errorf("err = %v, RateLimitError を期待", err)
	}

	var rle *RateLimitError
	if !errorAs(err, &rle) {
		t.Fatal("errors.As(&RateLimitError) が false")
	}
	if rle.RetryAfterSec != 300 {
		t.Errorf("RetryAfterSec = %d, want 300", rle.RetryAfterSec)
	}
}

func TestFetch_Timeout(t *testing.T) {
	// 10 秒タイムアウトを超える応答でエラーになるか。実テストでは短い timeout を使う。
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(rssSample))
	}))
	defer srv.Close()

	f := &Fetcher{
		client:    &http.Client{Timeout: 50 * time.Millisecond},
		userAgent: "sentei-test/0.0",
	}
	_, err := f.Fetch(context.Background(), config.FeedConfig{URL: srv.URL})
	if err == nil {
		t.Fatal("timeout でエラーを期待")
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter(""); got != 0 {
		t.Errorf("empty = %d, want 0", got)
	}
	if got := parseRetryAfter("120"); got != 120 {
		t.Errorf("numeric = %d, want 120", got)
	}
	if got := parseRetryAfter("not a number"); got != 0 {
		t.Errorf("garbage = %d, want 0", got)
	}
}

// errorAs は testing-only の薄い wrapper (errors.As へのアクセス簡略化)
func errorAs(err error, target any) bool {
	type asable interface {
		As(any) bool
	}
	if ae, ok := err.(asable); ok {
		return ae.As(target)
	}
	// 標準 errors.As の代わりに型アサーション相当を直接使うが、
	// RateLimitError は pointer receiver なので直接アサートでいい。
	if rle, ok := err.(*RateLimitError); ok {
		if ptr, ok := target.(**RateLimitError); ok {
			*ptr = rle
			return true
		}
	}
	return false
}
