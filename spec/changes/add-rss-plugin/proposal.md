# 提案: RSS プラグインの追加

## なぜ

**背景**:
- Phase 1 で GitHub 通知は sentei 化できたが、もう一方の情報源である **RSS / Atom フィード** (ブログ・技術記事) は未対応
- 個人の興味領域 (AI / LLM / SWE / Claude Code / TypeScript / React) のフィードを 1 日に数十件見るが、全部読むのは非現実的。urgency + category で色付けすれば「今日読むべき記事」を数秒で判定できる
- 基盤側は拡張を前提にした設計になっている: Core の `Plugin` インターフェース、Bonsai の grammar 差し替え機構 (`GrammarProvider`)、category 列の制約なし (TEXT) 等がプラグイン固有 enum を受け入れる

**現状**: `plugins/git/` のみ。RSS プラグインおよび `spec/specs/rss-plugin/` が存在しない。

**目指す状態**: 登録した RSS/Atom フィードの新着エントリを 15 分間隔でポーリングし、Bonsai で urgency + category をラベリングして `sentei list` および macOS アプリに表示できる。デーモン再起動後も取りこぼし無し。

## コンセプト

**git プラグインの「通知トリガー型」と対称な構成**。サーベイ (batch 集約) は入れず、新着エントリ = 通知として 1 件ずつ Submit する。

- feed list は `config.toml` の `[[plugins.rss.feeds]]` テーブル配列で宣言 (フィード個別設定を許す形式)
- 差分検出は DB の `UNIQUE(source, source_id)` 制約と `LastLabeledAtBySource("rss")` に任せる (in-memory seen-ids は使わない)
- 起動時刻または最後のラベリング時刻より古いエントリは skip (初回洪水防止と再起動ギャップ復旧の両立)
- Bonsai grammar は git と同じ JSON 形状 `{urgency, category, summary}`、**RSS は urgency を 3 値に絞る** (`urgent` は無し)
- per-feed の `urgency_floor` を metadata 経由で Core に渡し、Core 側で汎用的に post-process

## 決定事項 (grill 結果、14 項目)

| # | 論点 | 決定 |
|---|---|---|
| 1 | category enum の所有者 | **rss-plugin/spec.md のみ**。core spec には配置しない |
| 2 | 初期フィード | 11 件 (高信号 3 + Zenn topic 5 + Qiita tag 3) |
| 3 | 分類優先順位 | prompt に番号付き rule: `llm_research > llm_news > dev_tools > swe > other` |
| 4 | urgency スケール | **3 値** (`should_check` / `can_wait` / `ignore`)。grammar で制約 |
| 5 | per-feed 設定 | `[[plugins.rss.feeds]]` 形式で `url` / `name` / `urgency_floor` を持たせる |
| 6 | 再起動ギャップ対策 | 閾値 `max(LastLabeledAtBySource("rss"), now - 24h)` より新しい pubDate のみ submit |
| 7 | source_id | `sha256(normalized_url)` (feed-host prefix なし)。複数フィードで同記事をまとめる |
| 8 | ポーリング戦略 | 並列 fetch + 直列 Submit (pubDate 順) |
| 9 | Item.Content | 本文抜粋 400 字のみ (HTML 除去)。title/feed_name は JSON 注入経由 |
| 10 | 失敗扱い | 最小 (warn ログ) + 429 rate-limit のみ特別扱い (`Retry-After` 尊重) |
| 11 | prompt | 5 few-shot examples 付きの明示 rule 型 (git の先例踏襲) |
| 12 | Storage 経路 | RssPlugin のコンストラクタに `StorageReader` を注入 (Core interface 不変) |
| 13 | urgency_floor 適用 | Bonsai 応答 label の urgency を、`item.Metadata["urgency_floor"]` と比較して Core 側で post-process。汎用機構に一般化 |
| 14 | User-Agent / name fallback / HTML除去 / validation / CLI 表示 | 実装レベルで固定、tasks.md に記載 |

## 初期フィード構成

```toml
[plugins.rss]
enabled = false   # opt-in (既存 config を壊さない)
poll_interval_sec = 900

# 高信号 (固定)
[[plugins.rss.feeds]]
url = "https://www.anthropic.com/news/rss.xml"
name = "Anthropic News"
urgency_floor = "should_check"

[[plugins.rss.feeds]]
url = "https://lilianweng.github.io/index.xml"
name = "Lil'Log (Lilian Weng)"
urgency_floor = "should_check"

[[plugins.rss.feeds]]
url = "https://simonwillison.net/atom/everything/"
name = "Simon Willison"

# Zenn topic
[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/claudecode/feed"
name = "Zenn - Claude Code"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/typescript/feed"
name = "Zenn - TypeScript"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/ai/feed"
name = "Zenn - AI"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/llm/feed"
name = "Zenn - LLM"

[[plugins.rss.feeds]]
url = "https://zenn.dev/topics/react/feed"
name = "Zenn - React"

# Qiita tag
[[plugins.rss.feeds]]
url = "https://qiita.com/tags/typescript/feed"
name = "Qiita - TypeScript"

[[plugins.rss.feeds]]
url = "https://qiita.com/tags/react/feed"
name = "Qiita - React"

[[plugins.rss.feeds]]
url = "https://qiita.com/tags/llm/feed"
name = "Qiita - LLM"
```

Zenn topic slug (`claudecode` など) は実装時に URL で存在確認、404 なら `claude-code` 等の代替スラッグに差し替え。

## category enum (5 値)

- `llm_research` — 研究・論文・アーキテクチャ/訓練/評価手法の深掘り
- `llm_news` — LLM 製品の発表、リリース、ベンチマーク
- `dev_tools` — ツール / ライブラリ / CLI / エディタ拡張の紹介・レビュー
- `swe` — 言語 / FW / 設計 / パターン等、ツール非依存の技術記事
- `other` — 上記いずれにも該当しない

## 変更内容

### Go 側
- `plugins/rss/plugin.go` — Plugin インターフェース実装、pollOnce ループ、閾値計算 (Q6b)
- `plugins/rss/fetch.go` — HTTP fetch + gofeed parse、429 特別扱い、10 秒 timeout
- `plugins/rss/strip.go` — HTML タグ除去 + 400 字 truncate
- `plugins/rss/sourceid.go` — normalize URL + sha256 短縮 hash (Q7)
- `plugins/rss/grammar.go` — GBNF grammar (urgency 3 値 / category 5 値) と prompt (5 few-shot)
- `plugins/rss/*_test.go` — 各ファイル対応テスト
- `cmd/sentei/serve.go` — RSS プラグイン登録 + `engine.Storage()` を `StorageReader` として注入
- `cmd/sentei/plugin_list.go` — rss case 追加 (ポーリング間隔・フィード件数・urgency_floor 表示)
- `cmd/sentei/main.go` — `const Version = "0.1.0"` 追加 (User-Agent 用)
- `internal/config/config.go` — `RssConfig` / `FeedConfig` 構造体、TOML 雛形追加、validation
- `internal/storage/storage.go` — `LastLabeledAtBySource(source)` メソッド追加
- `internal/plugin/plugin.go` — `UrgencyRank` map 追加 (post-process の比較に使用)
- `internal/core/core.go` — `Submit` 内で Bonsai 応答後に `urgency_floor` metadata を解釈して格上げ
- `go.mod` — `github.com/mmcdole/gofeed`, `golang.org/x/net/html` を追加

### spec
- `spec/specs/rss-plugin/spec.md` を新規作成 (本 change の rss-plugin/spec-delta.md から)
- `spec/specs/core/spec.md` に「Metadata 経由の urgency floor 一般化」要件を ADDED

### config
- 初期値 `enabled = false` で既存 config を壊さない
- 雛形は 11 フィード分のコメント付きテンプレート

## 影響範囲

### 影響する仕様
- `spec/specs/rss-plugin/spec.md` — **新規追加** (ADDED)
- `spec/specs/core/spec.md` — 「Metadata ベースの urgency floor 適用」要件を ADDED
- `spec/specs/cli/spec.md` — 変更なし (既に `rss` source 対応済み)

### 影響するコード
- `plugins/rss/` — 新規ディレクトリ (6 ファイル + tests)
- `internal/storage/storage.go` — `LastLabeledAtBySource` 追加のみ (既存挙動変更なし)
- `internal/core/core.go` — Submit 内に 5-10 行の post-process 追加
- `internal/plugin/plugin.go` — `UrgencyRank` 定数追加
- `internal/config/config.go` — TOML 雛形と struct 拡張
- `cmd/sentei/serve.go` / `plugin_list.go` / `main.go`
- `go.mod` / `go.sum` — gofeed + x/net/html

### ユーザー影響
- 既存ユーザーの config は壊れない (`enabled = false` 初期値)
- 有効化後は 15 分ごとに指定フィードへ HTTP リクエスト
- Bonsai のラベリング負荷が増える (11 フィードで初回 ~40-80 件、以降平均 10-20 件/日)

### API 変更
- なし (既存の `/api/items?source=rss` で取れる)

### マイグレーション
- 不要 (config 初期値 `enabled = false`、DB スキーマ変更なし)

## 規模見積り

Medium。並列 fetch + 直列 Submit の構造は自明で、grammar / prompt も git の先例をなぞれる。urgency_floor の core 側汎用化が最も繊細な箇所。半日〜1 日。

## リスク

- **Zenn/Qiita topic slug の実在性**: URL が 404 なら当該フィードだけ warn ログで skip。実装時に 1 度確認し、違えば config 更新
- **Bonsai 初回 submit の渋滞**: 11 フィード × 24h 分で 40-80 件 × 3 秒/件 = 2-4 分。serial submit で自然に解消するが、UI からは「sentei list が gradually 埋まる」体験
  - 緩和: `slog.Info` で進捗をログ出力、ユーザーが待っているか判別可
- **urgency_floor の汎用化が core spec の責務を広げる**: これまで core は「Item を受ける / Bonsai に投げる / DB に保存」だけだった。Metadata の解釈 (たとえ floor hint だけでも) を core に入れるのは abstraction の拡張
  - 緩和: metadata hint の interpretation はオプショナル動作 (metadata に無ければ no-op)。将来他プラグインが同じ機構を使えるように 1 点だけ開ける
- **複数フィード同一記事の Submit レース**: 並列 fetch 中、Simon Willison の記事が他フィードにもあると、並列 goroutine が集約した後 serial submit で処理されるが、source_id が同じ → DB UNIQUE 制約で先行分のみ採用、後続は slog.Debug で無害 skip
- **pubDate の parse 失敗**: gofeed が parse できないフィード (稀) は `Timestamp = zero`。閾値比較で false になり submit されない → 実質 ignore される。ログで可視化
