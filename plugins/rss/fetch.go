/**
 * RSS / Atom フィードを HTTP で取得し、gofeed で parse する。
 * 10 秒タイムアウト、User-Agent 設定、HTTP 429 の Retry-After 尊重を扱う。
 */
package rss

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/senna-lang/sentei/internal/config"
)

// Fetcher は HTTP 経由でフィードを取得する。
type Fetcher struct {
	client    *http.Client
	userAgent string
}

// NewFetcher は 10 秒タイムアウトのデフォルト Fetcher を返す。
func NewFetcher(userAgent string) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: userAgent,
	}
}

// RateLimitError は HTTP 429 を受けたことを示す専用エラー。
// プラグインはこれを検出して当該フィードの次回 fetch を Retry-After 時刻まで保留する。
type RateLimitError struct {
	FeedURL       string
	RetryAfterSec int // 0 の場合は Retry-After ヘッダが無い
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited: %s (retry after %d sec)", e.FeedURL, e.RetryAfterSec)
}

// IsRateLimit は err が RateLimitError かを判定する。
func IsRateLimit(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// Fetch は単一フィードを取得して parse する。
// HTTP 429 は RateLimitError として返し、それ以外の 4xx/5xx は一般 error にする。
func (f *Fetcher) Fetch(ctx context.Context, fc config.FeedConfig) (*gofeed.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fc.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("request 生成失敗: %w", err)
	}
	if f.userAgent != "" {
		req.Header.Set("User-Agent", f.userAgent)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch 失敗 %s: %w", fc.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{FeedURL: fc.URL, RetryAfterSec: retryAfter}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, fc.URL)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("feed parse 失敗 %s: %w", fc.URL, err)
	}
	return feed, nil
}

// parseRetryAfter は Retry-After ヘッダを秒数として解釈する。
// 値が数値なら秒、HTTP 日付の場合は差分秒、解釈不能なら 0。
func parseRetryAfter(v string) int {
	if v == "" {
		return 0
	}
	if n, err := strconv.Atoi(v); err == nil && n >= 0 {
		return n
	}
	if t, err := http.ParseTime(v); err == nil {
		diff := time.Until(t).Seconds()
		if diff > 0 {
			return int(diff)
		}
	}
	return 0
}
