/**
 * RSS エントリの一意識別子を生成する。
 * GUID 優先 (URL 形式の場合) → entry.Link → title+published の順でフォールバック。
 * URL は正規化 (scheme/host/tracking params) してから SHA256 の先頭 16 文字を使う。
 */
package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

// 除去対象のトラッキングクエリパラメータ (完全一致)
var trackingParams = map[string]struct{}{
	"fbclid": {},
	"gclid":  {},
	"ref":    {},
}

// normalizeURL は URL を等価類へ落とし込む正規化を行う。
// - scheme を https に統一
// - host を lowercase、www. prefix を除去
// - utm_* / fbclid / gclid / ref を除去
// - fragment を除去
// - path の trailing slash を除去 (ルート "/" を除く)
func normalizeURL(u string) string {
	parsed, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return u
	}
	if parsed.Scheme == "http" {
		parsed.Scheme = "https"
	}
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Host = strings.TrimPrefix(parsed.Host, "www.")
	parsed.Fragment = ""

	q := parsed.Query()
	for k := range q {
		if strings.HasPrefix(k, "utm_") {
			q.Del(k)
			continue
		}
		if _, hit := trackingParams[k]; hit {
			q.Del(k)
		}
	}
	parsed.RawQuery = q.Encode()

	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}

	return parsed.String()
}

// buildSourceID はエントリから sentei の source_id を構築する。
// 優先順位:
// 1. GUID が http(s):// で始まる → GUID を URL 正規化して hash
// 2. それ以外で entry.Link がある → Link を URL 正規化して hash
// 3. 両方無い → title + published の hash (フォールバック)
func buildSourceID(entry *gofeed.Item) string {
	if entry == nil {
		return ""
	}

	if g := strings.TrimSpace(entry.GUID); g != "" {
		if strings.HasPrefix(g, "http://") || strings.HasPrefix(g, "https://") {
			return shortHash(normalizeURL(g))
		}
	}
	if l := strings.TrimSpace(entry.Link); l != "" {
		return shortHash(normalizeURL(l))
	}

	return shortHash(entry.Title + "|" + entry.Published)
}

// shortHash は input の SHA256 を先頭 16 hex 文字で返す (64 bit 衝突耐性)
func shortHash(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])[:16]
}
