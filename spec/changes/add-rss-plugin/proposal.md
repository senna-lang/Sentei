# 提案: RSS プラグインの追加

## なぜ

**背景**:
- Phase 1 で GitHub 通知は sentei 化できたが、もう一方の情報源である **RSS / Atom フィード** (ブログ・技術記事) は未対応
- 個人の興味領域 (AI / LLM / SWE) のフィードを 1 日に数十件見るが、全部読むのは非現実的。urgency + category で色付けすれば「今日読むべき記事」を数秒で判定できる
- core spec は既に RSS 用 category enum (`llm_research` / `llm_news` / `dev_tools` / `other`) を**予約済み**で、本変更で正式採用に格上げする
- CLI spec も `rss` source のフィルタ例を既に記載済み、サーバー側は `/api/items?source=rss` で動作する見込み

**現状**: `plugins/git/` のみ。RSS プラグインおよび `spec/specs/rss-plugin/` が存在しない。

**目指す状態**: 登録した RSS/Atom フィードの新着エントリを 15 分間隔でポーリングし、Bonsai で urgency + category をラベリングして `sentei list` および macOS アプリに表示できる。

## コンセプト

**git プラグインの「通知トリガー型」と対称な構成**。サーベイ (batch 集約) は入れず、新着エントリ = 通知として 1 件ずつ Submit する。

- feed list は `config.toml` で宣言 (enabled フラグ + フィード URL 配列 + ポーリング間隔)
- 初回起動時は全エントリを seen-ids に登録するだけで Submit しない (起動前の記事が洪水にならないようにする)
- 以降は差分 (seen に無い GUID) のみ Submit
- Bonsai grammar は git と同じ JSON 形状 `{urgency, category, summary}`、category enum のみ差し替え

## 決定事項 (設計判断の記録)

| # | 論点 | 決定 |
|---|---|---|
| 1 | フィード管理 | `[plugins.rss] feeds = [...]` を `config.toml` で列挙。初期 5 件: Zenn / Qiita 人気 / Simon Willison / Anthropic News / Lilian Weng |
| 2 | 重複検出 | entry GUID (Atom の `<id>` / RSS の `<guid>`) 優先、無い場合は URL を fallback。`source_id` は `"<feed-host>:<guid-or-url-hash>"` |
| 3 | 本文抽出 | feed 内の description / content のみ使用。フル記事 HTTP fetch は v1 では行わない (別 proposal で検討) |
| 4 | サーベイ | **無し**。通知トリガー型のみ。週次ダイジェスト等は将来検討 |
| 5 | ポーリング間隔 | デフォルト 15 分 (`poll_interval_sec = 900`)、config で override 可 |
| 6 | 初回 fetch | 全エントリを seen-ids に登録し Submit しない。次ポーリング以降の差分のみ Submit |
| 7 | ライブラリ | `github.com/mmcdole/gofeed` (RSS / Atom / JSON Feed 統一) |
| 8 | Bonsai 特殊ルール | **Anthropic News フィード由来のエントリは urgency を強制的に `should_check` 以上にする** (fixed rule、Bonsai の判定が `can_wait` / `ignore` を返しても `should_check` に格上げ)。他のフィードは完全に Bonsai 任せ |

## 変更内容

### Go 側
- `plugins/rss/plugin.go` — `Plugin` インターフェース実装、feed ポーリングループ、GUID 差分検出
- `plugins/rss/fetch.go` — `gofeed.Parser` ラッパー、タイムアウト付き HTTP fetch、エラー時リトライ
- `plugins/rss/grammar.go` — RSS 用 GBNF grammar (category enum: `llm_research` / `llm_news` / `dev_tools` / `other`) と prompt テンプレート
- `plugins/rss/rule.go` — Anthropic News 由来の urgency 格上げルール (post-Bonsai フック)
- `plugins/rss/*_test.go` — 各ファイル対応テスト (gofeed モック / 差分検出 / grammar / rule)
- `cmd/sentei/serve.go` — RSS プラグイン登録 (config.Plugins.Rss.Enabled のみで gate)
- `internal/config/config.go` — `[plugins.rss]` セクション追加 (enabled / poll_interval_sec / feeds)
- `go.mod` — `github.com/mmcdole/gofeed` を追加

### spec
- `spec/specs/rss-plugin/spec.md` を新規作成 (本 change の `specs/rss-plugin/spec-delta.md` から生成)
- `spec/specs/core/spec.md` を更新 (`specs/core/spec-delta.md` 経由): category enum の「予約」を「採用」へ格上げ

### config
- `~/.config/sentei/config.toml` の雛形に `[plugins.rss]` セクション追加
  - `enabled = false` (初期は opt-in、既存ユーザーの config を壊さない)
  - `poll_interval_sec = 900`
  - `feeds = [ ... 5 件 ... ]`

## 影響範囲

### 影響する仕様
- `spec/specs/rss-plugin/spec.md` — **新規追加** (ADDED)
- `spec/specs/core/spec.md` — Requirement「category enum の範囲」を MODIFIED (予約 → 採用)
- `spec/specs/cli/spec.md` — 変更なし (既に `rss` source を想定した記述あり)

### 影響するコード
- `plugins/rss/` — 新規ディレクトリ
- `cmd/sentei/serve.go` — プラグイン登録追加のみ
- `internal/config/config.go` — 構造体と TOML 雛形に RSS セクション追加
- `go.mod` / `go.sum` — gofeed 依存追加

### ユーザー影響
- 既存ユーザーの config は壊れない (`enabled = false` 初期値)
- 有効化後は 15 分ごとに指定フィードへ HTTP リクエスト
- Bonsai のラベリング負荷が増える (5 フィード × 平均 5 件/日 = 25 件/日増、M2 で数十秒/日)

### API 変更
- なし (既存の `/api/items?source=rss` で取れる)

### マイグレーション
- 不要 (config の初期値は `enabled = false`、DB スキーマ変更なし)

## 規模見積り

Medium。ポーリングループと差分検出は git プラグインの pattern をほぼ踏襲できる。gofeed の扱い + Anthropic 特殊ルール + spec 文書化で半日〜1 日程度。

## リスク

- **フィード側の仕様揺れ**: RSS 2.0 / Atom / JSON Feed で微妙に異なる。`gofeed` で吸収される想定だが、一部フィード (特に自作 RSS) は `<guid>` が無く URL fallback に頼ることになる
  - 緩和: fallback 経路を最初から実装。テストで両 pattern をカバー
- **Bonsai のラベリング精度**: RSS 記事のタイトルと抜粋だけで urgency を正確に判定するのは git 通知より難しい (文脈が薄い)
  - 緩和: 運用しながら prompt を調整。Anthropic News の urgency 格上げ rule で少なくとも 1 つは signal を担保
- **HTTP fetch の失敗伝播**: 1 フィードのネットワーク障害が全ポーリングを止める risk
  - 緩和: フィード単位で try/catch、失敗はログ出力のみ、次フィードへ進む。timeout は 10 秒
- **洪水**: 初回 fetch で数百エントリが Submit される事故
  - 緩和: 初回 seen-ids 登録で Submit skip (上記 #6)
- **レート制限**: Zenn / Qiita は RSS に対して明示的なレート制限は公表していないが、過剰ポーリングは配信元に迷惑
  - 緩和: 15 分間隔を下限として config で過小設定を許さない (Clamp しないが推奨値をコメントに明記)
