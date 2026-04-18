/**
 * プラグインインターフェースと共通型の定義
 * すべてのプラグイン（Git, RSS, arxiv 等）はこのインターフェースを実装する
 */
package plugin

import (
	"context"
	"time"
)

// Urgency はアイテムの緊急度を表す
type Urgency string

const (
	UrgencyUrgent      Urgency = "urgent"
	UrgencyShouldCheck Urgency = "should_check"
	UrgencyCanWait     Urgency = "can_wait"
	UrgencyIgnore      Urgency = "ignore"
)

// Item はプラグインからコアに送信される正規化された情報アイテム
type Item struct {
	Source    string            // プラグイン識別子 (例: "git")
	SourceID  string            // ソース内一意ID
	Title     string            // タイトル（表示のプライマリテキスト）
	Content   string            // 本文（Bonsai 判定の入力に使用）
	URL       string            // 元情報へのリンク
	Timestamp time.Time         // アイテムの発生時刻
	Metadata  map[string]string // プラグイン固有情報
}

// Label は Bonsai によるラベリング結果
type Label struct {
	Urgency  Urgency // urgent, should_check, can_wait, ignore
	Category string  // プラグインごとに異なる enum (例: pr, issue, ci)
	Summary  string  // Bonsai 生成の要約（品質不安定、title をフォールバック）
}

// LabeledItem はラベリング済みの Item
type LabeledItem struct {
	Item      Item
	Label     Label
	LabeledAt time.Time
}

// Core はプラグインがアイテムを送信するためのインターフェース
type Core interface {
	Submit(item Item) error
}

// Plugin はすべてのプラグインが実装するインターフェース
type Plugin interface {
	Name() string
	Start(ctx context.Context, core Core) error
	Stop() error
}

// GrammarProvider はプラグインごとの GBNF grammar を提供するインターフェース
type GrammarProvider interface {
	Grammar() string // GBNF grammar 文字列を返す
}
