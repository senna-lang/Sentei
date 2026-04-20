# 提案: RSS の初回バックフィル機構

## なぜ

**背景**:
- 現 RSS プラグインは閾値 `max(LastLabeledAtBySource("rss"), now - 24h)` より新しい pubDate のエントリだけを Submit する (add-rss-plugin で決めたロジック)
- この設計は「再起動ギャップ復旧」と「初回洪水防止」を両立するが、**低頻度・高信号フィード**で致命的に困る:
  - **Karpathy Blog**: 数本/年
  - **Lil'Log (Lilian Weng)**: 月 1-2 本
  - **Simon Willison**: 週数本だが深い記事は月 1-2
- フィード追加直後、24h 窓に記事が入らないと **当該フィードから 1 件も届かない**。Karpathy の場合、次投稿まで数ヶ月「無表示」の可能性
- ユーザーから見ると「フィード追加したのに動いてない」と誤認しやすい

**現状**: `add-rss-plugin` archive 時点で運用開始。2026-04-20 に Karpathy Blog を初期フィードに追加したが、直近投稿が 10 日前なのでラベリング対象にならない。

**目指す状態**: フィード追加時に任意の日数 (例 90 日) 分を **1 回だけ** バックフィルできる。通常運用時の洪水抑制は維持。

## 選択肢

| 案 | 仕組み | 長所 | 短所 |
|---|---|---|---|
| A | per-feed config `initial_backfill_days` | 宣言的、config 再読み込みのみで適用 | 「初回か」の判定が必要。feed 追加後に daemon が 1 回 backfill したら二度としないようにする仕組みが要る |
| B | CLI `sentei rss backfill <url> --days N` コマンド | 明示的、1 回コマンドで済む、閾値ロジックを汚さない | CLI 操作が必要 (ユーザーが忘れがち) |
| C | (A) + (B) のハイブリッド | 自動化 + 手動リカバリ | 実装範囲が広がる |

## 推奨: 案 B (CLI コマンド) のみ

**理由**:
- 宣言的 config (A) は「初回バックフィル済みフラグ」を永続化する必要があり、**config file 単独の設計を壊す** (副作用フラグが DB 等に漏れる)
- CLI 方式は閾値ロジックに一切触れず、単発の admin 操作として隔離できる
- ユーザーが Karpathy Blog を追加した直後に `sentei rss backfill https://karpathy.github.io/feed.xml --days 90` を 1 回実行するだけ
- 大量の古い記事を読み込むリスクは、--days で明示的に制御できる

**CLI 仕様案**:
```
sentei rss backfill <feed-url> [--days N]

Options:
  --days N         バックフィル期間 (デフォルト 30、最大 365)
  --dry-run        取得はするが Submit しない (件数のみ表示)
  --all            --days の代わりに全エントリを対象 (危険、dry-run 推奨)
```

動作:
1. CLI が daemon の新エンドポイント `POST /api/rss/backfill` を呼ぶ
2. daemon がそのフィードを通常 fetch (gofeed parse)
3. pubDate 閾値チェックを **bypass** して、指定期間内の全エントリを Submit キューへ
4. DB の UNIQUE 制約で既存分はスキップ、新規のみ Bonsai ラベリング
5. 結果件数を CLI に返す

## 影響範囲

### 影響する仕様
- `spec/specs/rss-plugin/spec.md` — ADDED: 「手動バックフィル」要件
- `spec/specs/cli/spec.md` — ADDED: `rss backfill` サブコマンド要件
- `spec/specs/core/spec.md` — 変更なし (閾値ロジックは維持)

### 影響するコード
- `plugins/rss/backfill.go` (新設) — pollOnce の閾値を受け取らないバリアントを実装
- `internal/server/server.go` — `POST /api/rss/backfill` ハンドラ追加
- `cmd/sentei/rss_backfill.go` (新設) — CLI サブコマンド
- `cmd/sentei/main.go` — `rss` サブコマンドグループ + `backfill` 登録

### ユーザー影響
- 既存ユーザーに自動変化なし (明示実行のみ)
- 新フィード追加時のワークフローに 1 ステップ追加:
  1. config 編集
  2. `sentei serve` 再起動
  3. `sentei rss backfill <url> --days 90` (← 新規)

### API 変更
- `POST /api/rss/backfill` 新規。Body: `{url: string, days: int, dry_run: bool}`。Response: `{fetched: int, submitted: int, skipped: int}`

### マイグレーション
- 不要 (既存フィードは通常動作、手動実行のみ新規機能)

## 規模見積り

Small (半日)。既存 Fetcher と entriesToItems の流用で済む。新しい grammar / prompt は不要。

## リスク

- **誤って全フィードを `--days 365` で backfill する**: 数千エントリ × Bonsai 3 秒 = 数時間ブロック。`--dry-run` の周知で緩和。`--days` の上限 365 で多少抑制
- **バックフィル中に通常 polling が走る**: core.Submit の `sync.Mutex` で直列化されるため並行実行は安全だが遅延が伸びる。UX 的には問題小
- **古い記事の pubDate が localStorage 的 LastLabeledAtBySource を更新する**: backfill 完了後、次ポーリングの閾値が「古い pubDate」になるバグ。注意: Submit 時の `labeled_at = time.Now()` なので問題なし。ただし要確認

## メモ

- Karpathy Blog 追加 (2026-04-20) に起因した気付き。実装着手するときはこの proposal を読み返す
- 将来、feed 追加 GUI (macOS アプリ側) を作る時はこの backfill を自動実行する UX を検討 (ボタン 1 個で追加 + backfill)
