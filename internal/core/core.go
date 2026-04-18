/**
 * コアエンジン
 * プラグインからの Item 受信 → Bonsai ラベリング → SQLite 保存のパイプラインを管理する
 */
package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/senna-lang/sentei/internal/bonsai"
	"github.com/senna-lang/sentei/internal/plugin"
	"github.com/senna-lang/sentei/internal/storage"
)

// Engine はコアエンジン。Plugin インターフェースの Core を実装する
type Engine struct {
	storage  *storage.Storage
	bonsai   *bonsai.Client
	plugins  []plugin.Plugin
	grammar  map[string]string // source → GBNF grammar
	prompt   map[string]string // source → プロンプトテンプレート
	onSubmit []func(plugin.LabeledItem) // Submit 成功後のコールバック
	mu       sync.Mutex
	cancel   context.CancelFunc
}

// Config はコアエンジンの設定
type Config struct {
	DBPath    string
	BonsaiURL string
}

// New はコアエンジンを生成する
func New(cfg Config) (*Engine, error) {
	st, err := storage.New(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("ストレージ初期化失敗: %w", err)
	}

	bc := bonsai.NewClient(cfg.BonsaiURL)

	return &Engine{
		storage: st,
		bonsai:  bc,
		grammar: make(map[string]string),
		prompt:  make(map[string]string),
	}, nil
}

// OnSubmit は Submit 成功後に呼び出されるコールバックを登録する
func (e *Engine) OnSubmit(fn func(plugin.LabeledItem)) {
	e.onSubmit = append(e.onSubmit, fn)
}

// RegisterPlugin はプラグインを登録する
func (e *Engine) RegisterPlugin(p plugin.Plugin, grammar, promptTemplate string) {
	e.plugins = append(e.plugins, p)
	e.grammar[p.Name()] = grammar
	e.prompt[p.Name()] = promptTemplate
}

// Start は全プラグインを起動する
func (e *Engine) Start(ctx context.Context) error {
	ctx, e.cancel = context.WithCancel(ctx)

	for _, p := range e.plugins {
		slog.Info("プラグイン起動", "name", p.Name())
		if err := p.Start(ctx, e); err != nil {
			return fmt.Errorf("プラグイン %s 起動失敗: %w", p.Name(), err)
		}
	}
	return nil
}

// Stop は全プラグインを停止する
func (e *Engine) Stop() error {
	if e.cancel != nil {
		e.cancel()
	}

	var firstErr error
	for _, p := range e.plugins {
		slog.Info("プラグイン停止", "name", p.Name())
		if err := p.Stop(); err != nil {
			slog.Error("プラグイン停止失敗", "name", p.Name(), "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if err := e.storage.Close(); err != nil {
		slog.Error("ストレージ Close 失敗", "error", err)
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Submit は Item を受信し、Bonsai ラベリング → SQLite 保存を実行する（同期）
func (e *Engine) Submit(item plugin.Item) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 冪等性チェック
	exists, err := e.storage.Exists(item.Source, item.SourceID)
	if err != nil {
		return fmt.Errorf("冪等性チェック失敗: %w", err)
	}
	if exists {
		slog.Debug("重複アイテム、スキップ", "source", item.Source, "source_id", item.SourceID)
		return nil
	}

	// バリデーション
	if item.Title == "" {
		return fmt.Errorf("title が空です")
	}
	if item.Source == "" {
		return fmt.Errorf("source が空です")
	}
	if item.SourceID == "" {
		return fmt.Errorf("source_id が空です")
	}

	// Bonsai ラベリング
	grammar := e.grammar[item.Source]
	promptTemplate := e.prompt[item.Source]

	label, err := e.bonsai.Label(item, grammar, promptTemplate)
	if err != nil {
		slog.Warn("Bonsai ラベリング失敗、フォールバック使用",
			"source", item.Source,
			"source_id", item.SourceID,
			"error", err,
		)
		// フォールバックラベルは bonsai.Label() 内で設定済み
	}

	labeledItem := plugin.LabeledItem{
		Item:      item,
		Label:     label,
		LabeledAt: time.Now(),
	}

	// SQLite 保存
	if err := e.storage.SaveLabeledItem(labeledItem); err != nil {
		return fmt.Errorf("アイテム保存失敗: %w", err)
	}

	slog.Info("アイテム処理完了",
		"source", item.Source,
		"title", item.Title,
		"urgency", label.Urgency,
		"category", label.Category,
	)

	// Submit 成功後のコールバック呼び出し
	for _, fn := range e.onSubmit {
		fn(labeledItem)
	}

	return nil
}

// Storage はストレージへの参照を返す（API ハンドラ用）
func (e *Engine) Storage() *storage.Storage {
	return e.storage
}

// BonsaiClient は Bonsai クライアントへの参照を返す（ヘルスチェック用）
func (e *Engine) BonsaiClient() *bonsai.Client {
	return e.bonsai
}

// PluginNames は登録されたプラグイン名一覧を返す
func (e *Engine) PluginNames() []string {
	names := make([]string, len(e.plugins))
	for i, p := range e.plugins {
		names[i] = p.Name()
	}
	return names
}
