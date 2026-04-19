# 実装タスク

## Phase 0: 依存と config スケルトン

1. `go get github.com/mmcdole/gofeed@latest` で依存追加、`go mod tidy`
2. `internal/config/config.go` に RSS セクション構造体を追加
   - `RssConfig { Enabled bool; PollIntervalSec int; Feeds []string }`
   - `PluginsConfig` に `Rss RssConfig` フィールド追加
3. `DefaultTOML()` に `[plugins.rss]` セクションを追記 (enabled = false、poll 900、feeds は初期 5 件)
4. `config_test.go` で新セクションの読み込みテスト (enabled / disabled / feeds の 3 パターン)

## Phase 1: フィード取得とパース

5. `plugins/rss/fetch.go` 新設: `fetchFeed(ctx, url) (*gofeed.Feed, error)`。`http.Client{Timeout: 10*time.Second}`、User-Agent は `sentei/<version>`
6. `fetch_test.go`: httptest.Server でモックフィード (RSS 2.0 / Atom) を返し、正常 parse / タイムアウト / 404 / 不正 XML の 4 ケース

## Phase 2: Plugin インターフェース実装

7. `plugins/rss/plugin.go` 新設: `Plugin` / `GrammarProvider` 実装
   - `type RssPlugin struct { config RssConfig; seen map[string]map[string]struct{}; ... }`
   - `Name() string` → `"rss"`
   - `Start(ctx, core)` → 各フィードごとに goroutine で ticker ループ起動
   - `Stop()` → context cancel で全 goroutine 停止
8. `plugin_test.go`: `Submit` が差分のみ呼ばれることを確認 (InMemoryCore で検証)

## Phase 3: 差分検出と初回 skip

9. `resolveSourceID(entry)` 実装: GUID 優先、無ければ URL の SHA256 短縮を使う。prefix は `<feed-host>:<key>`
10. 初回ポーリングフラグ: 最初の 1 回だけは全 entry を seen に登録し Submit しない (proposal 決定 #6)
11. 2 回目以降は seen 差分のみ Submit。seen map はメモリ保持 (デーモン再起動で reset、DB の UNIQUE 制約が冪等性を担保)
12. `plugin_test.go` で初回 skip と差分 Submit の境界テスト

## Phase 4: Bonsai grammar + prompt

13. `plugins/rss/grammar.go` 新設: GBNF grammar (category enum `llm_research` / `llm_news` / `dev_tools` / `other`)
14. prompt テンプレート定義: 入力は `title` / `excerpt (200 字)` / `feed_name`。出力は `{urgency, category, summary}` JSON
15. `grammar_test.go` で grammar 文字列の正当性 (無効な category を Bonsai が返せないこと) を構造的に検証

## Phase 5: Anthropic News ルール (post-Bonsai 格上げ)

16. `plugins/rss/rule.go` 新設: `applyPostLabelRules(feedURL, label) plugin.Label` を実装
17. ルール: feed URL が `anthropic.com/news/rss.xml` を含み、Bonsai が `can_wait` / `ignore` を返した場合、`should_check` に格上げ。`urgent` / `should_check` はそのまま
18. `rule_test.go`: 格上げケース × 4 (各 urgency) + 非 Anthropic feed で格上げされないこと
19. `plugin.go` の Submit 直前で `applyPostLabelRules` を通す (Core 側の挙動には変更を加えない)

## Phase 6: 登録とエンドツーエンド

20. `cmd/sentei/serve.go` で `cfg.Plugins.Rss.Enabled` が true なら `rssplugin.NewPlugin(cfg.Plugins.Rss)` を `engine.RegisterPlugin(p, grammar, prompt)` で登録
21. `sentei plugin list` の RSS 表示確認 (`plugin_list.go` の case 追加: フィード数・ポーリング間隔を表示)
22. 実機テスト: `[plugins.rss] enabled = true` にして `sentei serve` 起動、15 分 (or 手動で poll_interval_sec=60 に短縮して) 待ち、`sentei list --source rss` で新着が取れることを確認

## Phase 7: 仕様・最終化

23. `spec/changes/add-rss-plugin/specs/rss-plugin/spec-delta.md` の内容を `spec/specs/rss-plugin/spec.md` にコピー (新規ファイル)
24. `spec/specs/core/spec.md` の category enum Requirement を更新 (予約 → 採用)。delta は `specs/core/spec-delta.md` を参照
25. tasks.md に完了マーク、proposal の「目指す状態」との突合
26. archive へ移動: `spec/archive/YYYY-MM-DD-add-rss-plugin/` + `IMPLEMENTED` ファイル

---

**メモ**:
- Phase 0-6 は依存関係ほぼ一直線。Phase 4 と Phase 5 だけは並行可能
- tests は各 Phase で同時に書く (CLAUDE.md TDD 規則)
- 実装前にこの tasks.md を読み返して、proposal の「影響範囲」との差分が無いか確認
