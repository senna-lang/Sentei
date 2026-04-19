/**
 * RSS プラグイン本体。
 * 登録された複数フィードを定期ポーリングし、pubDate 閾値を超える新着エントリを
 * Bonsai ラベリング用に Core.Submit() へ直列で投入する。
 *
 * 閾値: max(storage.LastLabeledAtBySource("rss"), now - 24h)
 *   → 再起動ギャップの復旧と初回洪水防止を同時に満たす (Q6b)。
 */
package rss

import (
	"context"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/senna-lang/sentei/internal/config"
	"github.com/senna-lang/sentei/internal/plugin"
)

// StorageReader は Core の Storage のうち、RSS プラグインが利用する最小 API。
// コンストラクタ注入で RssPlugin に渡す (Q12)。
type StorageReader interface {
	LastLabeledAtBySource(source string) (time.Time, error)
}

// feedFetcher は fetch.go の *Fetcher と同じメソッドを持つ interface。
// テストでの差し替えを可能にする。
type feedFetcher interface {
	Fetch(ctx context.Context, fc config.FeedConfig) (*gofeed.Feed, error)
}

// RssPlugin は sentei の Plugin インターフェース実装。
type RssPlugin struct {
	config  config.RssConfig
	storage StorageReader
	fetcher feedFetcher

	mu            sync.Mutex
	nextAllowedAt map[string]time.Time // feed URL → 次に fetch 可能な時刻 (429 対応)

	cancel context.CancelFunc
}

// NewPlugin は RssPlugin を生成する。userAgent は HTTP User-Agent ヘッダに使う。
func NewPlugin(cfg config.RssConfig, storage StorageReader, userAgent string) *RssPlugin {
	return &RssPlugin{
		config:        cfg,
		storage:       storage,
		fetcher:       NewFetcher(userAgent),
		nextAllowedAt: make(map[string]time.Time),
	}
}

// Name はプラグイン識別子を返す。
func (p *RssPlugin) Name() string { return "rss" }

// Start はポーリング loop を起動する。ctx がキャンセルされるまで継続する。
func (p *RssPlugin) Start(ctx context.Context, core plugin.Core) error {
	ctx, p.cancel = context.WithCancel(ctx)

	interval := time.Duration(p.config.PollIntervalSec) * time.Second
	if interval <= 0 {
		interval = 900 * time.Second
	}

	go p.loop(ctx, core, interval)
	slog.Info("RSS プラグイン起動", "feeds", len(p.config.Feeds), "interval_sec", int(interval.Seconds()))
	return nil
}

// Stop はポーリング loop を停止する。
func (p *RssPlugin) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *RssPlugin) loop(ctx context.Context, core plugin.Core, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.pollOnce(ctx, core)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollOnce(ctx, core)
		}
	}
}

// computeThreshold は「この時刻より新しい pubDate のエントリのみ Submit する」閾値を返す。
// 閾値 = max(LastLabeledAtBySource("rss"), now - 24h)
func (p *RssPlugin) computeThreshold() time.Time {
	fallback := time.Now().Add(-24 * time.Hour)
	last, err := p.storage.LastLabeledAtBySource("rss")
	if err != nil || last.IsZero() || last.Before(fallback) {
		return fallback
	}
	return last
}

// pollOnce は 1 ポーリング周期を実行する。
// 1. 全フィードを並列 fetch (個別エラーは skip)
// 2. 閾値より新しいエントリを収集、pubDate 昇順 sort
// 3. 直列で core.Submit() (Bonsai 律速を自然に解消)
func (p *RssPlugin) pollOnce(ctx context.Context, core plugin.Core) {
	threshold := p.computeThreshold()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		all     []plugin.Item
	)

	for _, fc := range p.config.Feeds {
		if p.isRateLimited(fc.URL) {
			slog.Debug("rss rate-limit 待機中、skip", "feed", fc.URL)
			continue
		}
		wg.Add(1)
		go func(fc config.FeedConfig) {
			defer wg.Done()
			feed, err := p.fetcher.Fetch(ctx, fc)
			if err != nil {
				p.handleFetchError(fc.URL, err)
				return
			}
			items := p.entriesToItems(feed, fc, threshold)
			mu.Lock()
			all = append(all, items...)
			mu.Unlock()
		}(fc)
	}
	wg.Wait()

	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})

	for _, item := range all {
		if err := core.Submit(item); err != nil {
			slog.Warn("rss submit 失敗", "title", item.Title, "error", err)
		}
	}
}

// entriesToItems は feed の中から閾値を超える entry のみを Item に変換する。
func (p *RssPlugin) entriesToItems(feed *gofeed.Feed, fc config.FeedConfig, threshold time.Time) []plugin.Item {
	if feed == nil {
		return nil
	}

	feedName := fc.Name
	if feedName == "" {
		feedName = hostOf(fc.URL)
	}

	var out []plugin.Item
	for _, e := range feed.Items {
		ts := entryTimestamp(e)
		if ts.IsZero() || !ts.After(threshold) {
			continue
		}
		content := ""
		if e.Content != "" {
			content = stripHTML(e.Content, 400)
		} else if e.Description != "" {
			content = stripHTML(e.Description, 400)
		}

		item := plugin.Item{
			Source:    "rss",
			SourceID:  buildSourceID(e),
			Title:     strings.TrimSpace(e.Title),
			Content:   content,
			URL:       e.Link,
			Timestamp: ts,
			Metadata:  buildMetadata(e, fc, feedName),
		}
		if item.SourceID == "" || item.Title == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

// entryTimestamp は entry.PublishedParsed を優先し、無ければ UpdatedParsed を返す。
func entryTimestamp(e *gofeed.Item) time.Time {
	if e == nil {
		return time.Time{}
	}
	if e.PublishedParsed != nil {
		return *e.PublishedParsed
	}
	if e.UpdatedParsed != nil {
		return *e.UpdatedParsed
	}
	return time.Time{}
}

// buildMetadata は Core / 表示層が利用する metadata を組み立てる。
func buildMetadata(e *gofeed.Item, fc config.FeedConfig, feedName string) map[string]string {
	m := map[string]string{
		"feed_url":  fc.URL,
		"feed_name": feedName,
	}
	if fc.UrgencyFloor != "" {
		m["urgency_floor"] = fc.UrgencyFloor
	}
	if e == nil {
		return m
	}
	if e.Author != nil && e.Author.Name != "" {
		m["author"] = e.Author.Name
	}
	if len(e.Categories) > 0 {
		m["categories"] = strings.Join(e.Categories, ",")
	}
	if e.GUID != "" {
		m["guid"] = e.GUID
	}
	return m
}

// hostOf は feed URL から host (例: "zenn.dev") を返す。解析失敗時は URL をそのまま返す。
func hostOf(u string) string {
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" {
		return u
	}
	return parsed.Host
}

// isRateLimited は当該フィードが Retry-After の期間中かを確認する。
func (p *RssPlugin) isRateLimited(feedURL string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	t, ok := p.nextAllowedAt[feedURL]
	return ok && time.Now().Before(t)
}

// handleFetchError は fetch エラーをログに記録し、rate limit の場合は nextAllowedAt を更新する。
func (p *RssPlugin) handleFetchError(feedURL string, err error) {
	if IsRateLimit(err) {
		var rle *RateLimitError
		if errorAsRateLimit(err, &rle) {
			sec := rle.RetryAfterSec
			if sec <= 0 {
				sec = 600 // ヘッダ無ければ 10 分保留
			}
			p.mu.Lock()
			p.nextAllowedAt[feedURL] = time.Now().Add(time.Duration(sec) * time.Second)
			p.mu.Unlock()
			slog.Warn("rss レートリミット受信、次回 fetch を保留",
				"feed", feedURL, "retry_after_sec", sec)
			return
		}
	}
	slog.Warn("rss fetch エラー", "feed", feedURL, "error", err)
}

// errorAsRateLimit は error を *RateLimitError へ安全にキャスト。
func errorAsRateLimit(err error, target **RateLimitError) bool {
	if rle, ok := err.(*RateLimitError); ok {
		*target = rle
		return true
	}
	return false
}
