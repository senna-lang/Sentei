# 実装タスク

Grill 結果を反映し、以下の順序で進める。テストは各 Phase で同時に書く (CLAUDE.md TDD)。

## Phase 0: 依存と config スケルトン

1. `go get github.com/mmcdole/gofeed@latest`、`go get golang.org/x/net/html` (後者は未採用なら追加しない。通常は std ライブラリ)
2. `go mod tidy` で依存整理
3. `cmd/sentei/main.go` に `const Version = "0.1.0"` を追加 (User-Agent 用)
4. `internal/config/config.go` に `RssConfig` / `FeedConfig` 構造体を追加
   - `FeedConfig { URL string; Name string; UrgencyFloor string }`
   - `RssConfig { Enabled bool; PollIntervalSec int; Feeds []FeedConfig }`
   - `PluginsConfig` に `Rss RssConfig` フィールド追加
5. `DefaultTOML()` に 11 フィードの雛形 (コメント付き) を追記、`enabled = false`
6. `config.Load()` 内に validation を追加: URL が `http(s)://` で始まらなければ fatal、`urgency_floor` が enum 外なら fatal、`poll_interval_sec < 60` は warn ログ
7. `config_test.go` で: enabled/disabled、URL 不正、urgency_floor 不正、名前 fallback、poll 未指定の各ケース

## Phase 1: 基礎部品

8. `plugins/rss/sourceid.go` 新設: `normalizeURL(u string) string` + `buildSourceID(entry) string`
   - 正規化: scheme=https、host lowercase、`www.` 除去、`utm_*` / `fbclid` / `gclid` / `ref` 除去、fragment 除去、trailing slash 除去
   - source_id: sha256(normalized_url) の先頭 16 文字。GUID が `http(s)://` で始まる場合は GUID を URL 扱い、それ以外は entry.Link を使う。両方空なら `title + published` の hash
9. `sourceid_test.go`: 正規化 6 ケース + buildSourceID 4 ケース (GUID URL / GUID 非 URL / Link のみ / 完全破綻)
10. `plugins/rss/strip.go` 新設: `stripHTML(s string, maxLen int) string`
    - `golang.org/x/net/html.NewTokenizer` で text-only 抽出 (`<script>` / `<style>` 内容は除去)
    - 連続空白と改行を単一空白に圧縮
    - maxLen で truncate + 末尾 `...`
11. `strip_test.go`: プレーン / `<p>` / `<a>` / `<script>` 除去 / 改行圧縮 / truncate の 5-6 ケース

## Phase 2: フィード取得

12. `plugins/rss/fetch.go` 新設: `fetchFeed(ctx, fc FeedConfig) ([]*gofeed.Item, error)`
    - `http.Client{Timeout: 10*time.Second}`
    - User-Agent: `sentei/<Version> (+https://github.com/senna-lang/Sentei)`
    - ステータスコード 429 → `Retry-After` を読んで `rateLimitErr{fc.URL, retryAfter}` を返す (専用 error 型)
    - 400+ は一般 error、2xx は gofeed parser で `*gofeed.Feed` に変換
13. `fetch_test.go`: httptest.Server で RSS 2.0 / Atom / 404 / timeout / 429 with Retry-After の 5 ケース

## Phase 3: Plugin 骨格

14. `plugins/rss/plugin.go` 新設
    - 型: `RssPlugin { config RssConfig; storage StorageReader; nextAllowedAt map[string]time.Time }`
    - interface: `type StorageReader interface { LastLabeledAtBySource(source string) (time.Time, error) }`
    - `NewPlugin(cfg RssConfig, st StorageReader) *RssPlugin`
    - `Name() string { return "rss" }`
    - `Start(ctx, core) error`: 起動時刻記録なし (Q6b で不要、閾値は毎 poll で DB 参照)、goroutine 起動
    - `Stop() error`: context cancel
15. `plugin_test.go`: ポーリング 1 回の Submit 件数検証、mockCore / mockStorageReader

## Phase 4: pubDate 閾値と Submit

16. `plugin.go` に `computeThreshold(core plugin.Core) time.Time` を実装
    - `last, _ := p.storage.LastLabeledAtBySource("rss")`
    - `fallback := time.Now().Add(-24 * time.Hour)`
    - `last.IsZero() || last.Before(fallback)` なら `fallback` を返す、そうでなければ `last`
17. `plugin.go` に `pollOnce(ctx, core)` 実装 (Q8 (D) の並列 fetch + 直列 Submit)
    - 各フィードを goroutine で並列 fetch (sync.WaitGroup)
    - `nextAllowedAt[url]` が将来なら当該フィードを skip (429 対応)
    - 取得した entry を slice に集約、pubDate で昇順 sort
    - 閾値より古いものは drop
    - 直列で `core.Submit(item)` — item.Metadata に `feed_url` / `feed_name` / `author` / `categories` / `urgency_floor` / `guid` を詰める
18. `plugin_test.go` に閾値テスト追加 (前回 label 無し / ある / 24h 以上前)、429 スキップテスト

## Phase 5: Bonsai grammar + prompt

19. `plugins/rss/grammar.go` 新設: 公開定数 `Grammar` と `PromptTemplate`
    - grammar の urgency enum は `"should_check" | "can_wait" | "ignore"` の 3 値のみ
    - category enum は `"llm_research" | "llm_news" | "dev_tools" | "swe" | "other"`
20. prompt template (Q11 の 5 few-shot つき): `/no_think` prefix、User context、Category/Urgency 列挙、優先順位 rule、5 few-shot、`{notification_json}` placeholder
21. `grammar_test.go`: grammar 文字列に urgent が含まれないこと、category 5 値が含まれること、prompt が `{notification_json}` を持つことの構造テスト

## Phase 6: Core 側の urgency_floor 汎用化 (Q13 (A-3))

22. `internal/plugin/plugin.go` に `UrgencyRank` map を公開定数で追加:
    ```go
    var UrgencyRank = map[Urgency]int{
        UrgencyIgnore: 0, UrgencyCanWait: 1, UrgencyShouldCheck: 2, UrgencyUrgent: 3,
    }
    ```
23. `internal/core/core.go` の `Submit` 内、Bonsai ラベリング成功後・`SaveLabeledItem` 前に:
    ```go
    if floor := item.Metadata["urgency_floor"]; floor != "" {
        if plugin.UrgencyRank[label.Urgency] < plugin.UrgencyRank[plugin.Urgency(floor)] {
            label.Urgency = plugin.Urgency(floor)
        }
    }
    ```
24. `core_test.go` を新設 (現在テスト無し)。最低限: Submit + urgency_floor が期待通り格上げ、metadata 無しなら no-op、floor 文字列が不正値なら no-op

## Phase 7: Storage 拡張

25. `internal/storage/storage.go` に `LastLabeledAtBySource(source string) (time.Time, error)` 追加
    - 既存 `LastLabeledAt()` と同形、WHERE 句で source = ? を追加
26. `storage_test.go` に 3 ケース: 空テーブル / rss のみ / git と混在で rss 最新を取得

## Phase 8: 配線と CLI

27. `cmd/sentei/serve.go` に RSS プラグイン登録を追加
    - `cfg.Plugins.Rss.Enabled` が true なら `rssplugin.NewPlugin(cfg.Plugins.Rss, engine.Storage())` を `engine.RegisterPlugin(p, rssplugin.Grammar, rssplugin.PromptTemplate)` で登録
28. `cmd/sentei/plugin_list.go:52` に rss case を追加
    - 表示: `poll_interval_sec`、フィード件数、各フィード行 `"<name> [floor: <urgency_floor>]"` (floor 未指定は省略)
    - 10 件超過で `...and N more`

## Phase 9: spec 文書化と archive

29. `spec/specs/rss-plugin/spec.md` を新規作成 — `specs/rss-plugin/spec-delta.md` の ADDED 要件をベースに
30. `spec/specs/core/spec.md` に「Metadata ベースの urgency floor 適用」要件を ADDED
31. 実機テスト: `enabled = true` + poll 60 秒にして `sentei serve` 起動、1-2 cycle 観察、`sentei list --source rss` で新着確認、`sentei plugin list` の rss 表示確認、`--category swe` / `--urgency should_check` でフィルタ動作確認
32. archive へ移動: `mv spec/changes/add-rss-plugin spec/archive/YYYY-MM-DD-add-rss-plugin` + `IMPLEMENTED` ファイル (timestamp)

---

**並列化可能な箇所**:
- Phase 1 (基礎部品) / Phase 2 (fetch) は互いに独立
- Phase 5 (grammar) は Phase 1-4 と独立。どの順でも可
- Phase 6 (Core 拡張) と Phase 7 (Storage 拡張) は互いに独立

**着手時の最終確認**:
- 11 フィード URL を `curl -I` で一度確認、特に Zenn topic slug が実在するか
- Bonsai が起動中、`LastLabeledAtBySource` を呼ぶため DB はマイグレーション済みであること
