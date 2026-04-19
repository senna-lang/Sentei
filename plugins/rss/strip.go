/**
 * RSS エントリの description/content を plaintext に整形する。
 * HTML タグ除去 + 連続空白圧縮 + 最大文字数 truncate を 1 関数で行う。
 */
package rss

import (
	"strings"

	"golang.org/x/net/html"
)

// skipTags は tokenizer で開始タグを検出した時に内部の text を無視する要素。
var skipTags = map[string]struct{}{
	"script": {},
	"style":  {},
	"noscript": {},
}

// stripHTML は入力 s から HTML タグを除去し、連続空白/改行を単一空白に圧縮して、
// maxLen (rune 単位) で truncate する。truncate 時は末尾に "..." を付ける。
func stripHTML(s string, maxLen int) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	tokenizer := html.NewTokenizer(strings.NewReader(s))
	skipDepth := 0

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := tokenizer.TagName()
			if _, hit := skipTags[string(name)]; hit {
				skipDepth++
			}
			// タグ境界で連続テキストが結合しないよう空白を挿入
			// (collapseWhitespace で連続空白は単一空白に畳まれる)
			b.WriteByte(' ')
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			if _, hit := skipTags[string(name)]; hit && skipDepth > 0 {
				skipDepth--
			}
			b.WriteByte(' ')
		case html.TextToken:
			if skipDepth == 0 {
				b.Write(tokenizer.Text())
			}
		}
	}

	text := collapseWhitespace(b.String())
	return truncateRunes(text, maxLen)
}

// collapseWhitespace は ASCII 空白・タブ・改行を単一空白に畳み、
// 前後の空白を除去する。
func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true // 先頭の空白を抑止
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimRight(b.String(), " ")
}

// truncateRunes は s を最大 n rune に切り詰め、超過時は末尾に "..." を付ける。
// n <= 0 の場合は空文字列を返す。
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
