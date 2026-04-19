package rss

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/senna-lang/sentei/internal/config"
	"github.com/senna-lang/sentei/internal/plugin"
)

// mockCore は Submit された Item を保存する。
type mockCore struct {
	mu    sync.Mutex
	items []plugin.Item
}

func (m *mockCore) Submit(item plugin.Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, item)
	return nil
}

func (m *mockCore) snapshot() []plugin.Item {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]plugin.Item, len(m.items))
	copy(out, m.items)
	return out
}

// mockStorage は LastLabeledAtBySource に固定値を返す。
type mockStorage struct {
	last time.Time
	err  error
}

func (m *mockStorage) LastLabeledAtBySource(source string) (time.Time, error) {
	return m.last, m.err
}

// mockFetcher は事前に用意したエントリをそのまま返す。
type mockFetcher struct {
	feed *gofeed.Feed
	err  error
}

func (m *mockFetcher) Fetch(ctx context.Context, fc config.FeedConfig) (*gofeed.Feed, error) {
	return m.feed, m.err
}

func makeEntry(title, link, guid string, published time.Time) *gofeed.Item {
	return &gofeed.Item{
		Title:           title,
		Link:            link,
		GUID:            guid,
		Published:       published.Format(time.RFC3339),
		PublishedParsed: &published,
		Description:     "<p>" + title + " body</p>",
	}
}

func TestComputeThreshold_NoRssItems_UsesFallback(t *testing.T) {
	st := &mockStorage{last: time.Time{}}
	p := NewPlugin(config.RssConfig{}, st, "test-ua")

	th := p.computeThreshold()

	// fallback は now - 24h 付近
	diff := time.Since(th) - 24*time.Hour
	if diff < -time.Second || diff > time.Second {
		t.Errorf("閾値が 24h fallback になっていない: %v (diff=%v)", th, diff)
	}
}

func TestComputeThreshold_RecentLabeled_UsesLast(t *testing.T) {
	last := time.Now().Add(-2 * time.Hour)
	st := &mockStorage{last: last}
	p := NewPlugin(config.RssConfig{}, st, "test-ua")

	th := p.computeThreshold()

	if !th.Equal(last) {
		t.Errorf("閾値 = %v, want %v", th, last)
	}
}

func TestComputeThreshold_OldLabeled_UsesFallback(t *testing.T) {
	// last が fallback より古い (48h 前) → fallback が勝つ
	last := time.Now().Add(-48 * time.Hour)
	st := &mockStorage{last: last}
	p := NewPlugin(config.RssConfig{}, st, "test-ua")

	th := p.computeThreshold()

	// fallback (now - 24h) の方が newer なので採用されている
	if th.Before(last) {
		t.Errorf("古い last が採用されている: last=%v, th=%v", last, th)
	}
	diff := time.Since(th) - 24*time.Hour
	if diff < -time.Second || diff > time.Second {
		t.Errorf("閾値 = %v, want now-24h", th)
	}
}

func TestPollOnce_SubmitsOnlyNewEntries(t *testing.T) {
	// 閾値 = 1 時間前 → それより新しいエントリのみ Submit
	threshold := time.Now().Add(-1 * time.Hour)
	old := threshold.Add(-1 * time.Minute) // 古い
	fresh := threshold.Add(1 * time.Minute) // 新しい

	feed := &gofeed.Feed{
		Items: []*gofeed.Item{
			makeEntry("Old", "https://example.com/a", "https://example.com/a", old),
			makeEntry("Fresh", "https://example.com/b", "https://example.com/b", fresh),
		},
	}

	core := &mockCore{}
	st := &mockStorage{last: threshold}
	p := NewPlugin(config.RssConfig{
		Feeds: []config.FeedConfig{{URL: "https://example.com/feed", Name: "Example"}},
	}, st, "test-ua")
	p.fetcher = &mockFetcher{feed: feed}

	p.pollOnce(context.Background(), core)

	items := core.snapshot()
	if len(items) != 1 {
		t.Fatalf("submit 件数 = %d, want 1", len(items))
	}
	if items[0].Title != "Fresh" {
		t.Errorf("Submit された item = %q, want 'Fresh'", items[0].Title)
	}
}

func TestPollOnce_MetadataIncludesFeedAndFloor(t *testing.T) {
	fresh := time.Now().Add(-5 * time.Minute)
	feed := &gofeed.Feed{
		Items: []*gofeed.Item{
			makeEntry("P1", "https://example.com/a", "https://example.com/a", fresh),
		},
	}

	core := &mockCore{}
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{
		Feeds: []config.FeedConfig{{
			URL:          "https://example.com/feed",
			Name:         "Example Blog",
			UrgencyFloor: "should_check",
		}},
	}, st, "test-ua")
	p.fetcher = &mockFetcher{feed: feed}

	p.pollOnce(context.Background(), core)

	items := core.snapshot()
	if len(items) != 1 {
		t.Fatalf("submit 件数 = %d", len(items))
	}
	m := items[0].Metadata
	if m["feed_url"] != "https://example.com/feed" {
		t.Errorf("feed_url = %q", m["feed_url"])
	}
	if m["feed_name"] != "Example Blog" {
		t.Errorf("feed_name = %q", m["feed_name"])
	}
	if m["urgency_floor"] != "should_check" {
		t.Errorf("urgency_floor = %q", m["urgency_floor"])
	}
}

func TestPollOnce_FeedNameFallback(t *testing.T) {
	fresh := time.Now().Add(-5 * time.Minute)
	feed := &gofeed.Feed{
		Items: []*gofeed.Item{
			makeEntry("P1", "https://example.com/a", "https://example.com/a", fresh),
		},
	}

	core := &mockCore{}
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{
		Feeds: []config.FeedConfig{{URL: "https://example.com/feed"}}, // name 無指定
	}, st, "test-ua")
	p.fetcher = &mockFetcher{feed: feed}

	p.pollOnce(context.Background(), core)

	items := core.snapshot()
	if len(items) == 0 {
		t.Fatal("submit 0 件")
	}
	if items[0].Metadata["feed_name"] != "example.com" {
		t.Errorf("feed_name fallback = %q, want 'example.com'", items[0].Metadata["feed_name"])
	}
}

func TestPollOnce_SkipsRateLimitedFeed(t *testing.T) {
	fresh := time.Now().Add(-5 * time.Minute)
	feed := &gofeed.Feed{
		Items: []*gofeed.Item{
			makeEntry("P1", "https://example.com/a", "https://example.com/a", fresh),
		},
	}

	core := &mockCore{}
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{
		Feeds: []config.FeedConfig{{URL: "https://example.com/feed"}},
	}, st, "test-ua")
	p.fetcher = &mockFetcher{feed: feed}

	// 将来の nextAllowedAt を設定 → skip される
	p.nextAllowedAt["https://example.com/feed"] = time.Now().Add(10 * time.Minute)

	p.pollOnce(context.Background(), core)

	if len(core.snapshot()) != 0 {
		t.Errorf("rate-limited feed なのに Submit された")
	}
}

func TestHandleFetchError_SetsNextAllowedAtOnRateLimit(t *testing.T) {
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{}, st, "test-ua")

	url := "https://example.com/feed"
	rle := &RateLimitError{FeedURL: url, RetryAfterSec: 300}

	p.handleFetchError(url, rle)

	next, ok := p.nextAllowedAt[url]
	if !ok {
		t.Fatal("nextAllowedAt が設定されていない")
	}
	// 約 5 分後
	diff := time.Until(next) - 5*time.Minute
	if diff < -time.Second || diff > time.Second {
		t.Errorf("nextAllowedAt = %v (diff=%v)", next, diff)
	}
}

func TestHandleFetchError_NonRateLimit_NoStateChange(t *testing.T) {
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{}, st, "test-ua")

	url := "https://example.com/feed"
	p.handleFetchError(url, errors.New("generic error"))

	if _, ok := p.nextAllowedAt[url]; ok {
		t.Error("通常エラーで nextAllowedAt が設定された")
	}
}

func TestPollOnce_ItemContentStrippedAndTruncated(t *testing.T) {
	fresh := time.Now().Add(-5 * time.Minute)
	entry := makeEntry("Post", "https://example.com/a", "https://example.com/a", fresh)
	entry.Description = "<p>Hello <b>world</b></p>"

	feed := &gofeed.Feed{Items: []*gofeed.Item{entry}}
	core := &mockCore{}
	st := &mockStorage{}
	p := NewPlugin(config.RssConfig{
		Feeds: []config.FeedConfig{{URL: "https://example.com/feed"}},
	}, st, "test-ua")
	p.fetcher = &mockFetcher{feed: feed}

	p.pollOnce(context.Background(), core)

	items := core.snapshot()
	if len(items) == 0 {
		t.Fatal("submit 0 件")
	}
	if items[0].Content != "Hello world" {
		t.Errorf("content = %q, want 'Hello world'", items[0].Content)
	}
}

func TestPluginName(t *testing.T) {
	p := NewPlugin(config.RssConfig{}, &mockStorage{}, "test-ua")
	if p.Name() != "rss" {
		t.Errorf("Name() = %q, want 'rss'", p.Name())
	}
}
