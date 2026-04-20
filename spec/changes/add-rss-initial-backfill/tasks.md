# 実装タスク

## Phase 1: サーバー側 backfill エンドポイント

1. `plugins/rss/backfill.go` 新設: `Backfill(ctx, fc config.FeedConfig, days int, dryRun bool) BackfillResult`
   - `BackfillResult { Fetched, Submitted, Skipped int }`
   - pubDate 閾値は `now - days*24h`、それより新しい全エントリを Submit (dryRun なら Submit 飛ばす)
   - Submit 前に `core.Storage().Exists` で DB 重複を先取りして Skipped にカウント
2. `plugins/rss/backfill_test.go`: mockFetcher + mockCore で days ウィンドウ / dryRun / 重複スキップの各テスト
3. `internal/server/server.go` に `POST /api/rss/backfill` ハンドラ追加
   - Body decode + feed URL 検索 (config から当該 FeedConfig 取得、一致しなければ 404)
   - RSS プラグインインスタンスに Backfill を呼ぶ
   - Response JSON: `{fetched, submitted, skipped}`
4. `server_test.go` (あれば) or integration テストで 200 / 404 / 空 body のケース確認

## Phase 2: CLI サブコマンド

5. `cmd/sentei/rss.go` 新設: `rssCmd()` で `rss` グループを定義
6. `cmd/sentei/rss_backfill.go` 新設: `backfillCmd()` を実装
   - Args: `<feed-url>`
   - Flags: `--days N` (default 30, max 365)、`--dry-run`、`--all` (days を無視して最大値を使う shortcut)
   - daemon へ `POST /api/rss/backfill` を叩く
   - 結果を人間可読で出力 (fetched/submitted/skipped + 所要時間)
7. `cmd/sentei/main.go` の `rootCmd` に `rssCmd` を追加、`rssCmd.AddCommand(backfillCmd())` で subcmd 登録

## Phase 3: 仕様と確認

8. `spec/specs/rss-plugin/spec.md` に ADDED 要件「手動バックフィル」追記 (specs/rss-plugin/spec-delta.md を作って本 change に同梱)
9. `spec/specs/cli/spec.md` に ADDED 要件「rss backfill サブコマンド」追記 (specs/cli/spec-delta.md)
10. 実機テスト:
    - `sentei rss backfill https://karpathy.github.io/feed.xml --days 365 --dry-run` → 全エントリ件数表示確認
    - `sentei rss backfill https://karpathy.github.io/feed.xml --days 90` → Karpathy 直近記事が Bonsai ラベリングされて `sentei list --source rss` に現れること確認
    - 重複実行で Skipped が増えること確認
11. archive へ移動: `spec/archive/YYYY-MM-DD-add-rss-initial-backfill/` + `IMPLEMENTED`

---

**メモ**:
- 本提案は実装保留。Karpathy/Lil'Log のような低頻度フィードを運用で多用するようになったら着手
- Phase 1 単独でも daemon 側の機能として使える (CLI なくとも curl で叩ける)
- 将来 macOS アプリの「フィード追加 GUI」を作る時はこのエンドポイントを内部利用する
