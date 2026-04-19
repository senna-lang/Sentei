# 仕様差分: RSS プラグイン

対象仕様: `spec/specs/rss-plugin/spec.md` (**新規作成**)

RSS プラグインは git 通知トリガー型と対称な構成を取る。登録された RSS/Atom フィードをポーリングし、新着エントリを 1 件ずつ Bonsai に渡して urgency + category をラベリングする。サーベイ (batch 集約) は持たない。

**設計のキモ**:
- urgency は `should_check` / `can_wait` / `ignore` の 3 値 (git と違い `urgent` を含まない — 通知発火の責務は git のみが持つ)
- category は 5 値 (`llm_research` / `llm_news` / `dev_tools` / `swe` / `other`)
- 再起動時の取りこぼしを防ぐため、閾値 = `max(LastLabeledAtBySource("rss"), now - 24h)` より新しい pubDate のエントリのみ submit
- 並列 fetch + 直列 Submit で Bonsai の処理を律速させる
- per-feed の `urgency_floor` は metadata 経由で Core に渡し、Core の汎用 post-process が格上げを適用する

---

## ADDED 要件

### 要件: フィードのポーリング
WHILE RSS プラグインが稼働中の場合、
システムは `config.Plugins.Rss.Feeds` に列挙された各フィードを、`PollIntervalSec` (デフォルト 900 秒) 間隔でポーリングしなければならない (SHALL)。
取得はフィードごとに並列、Submit は pubDate 昇順で直列実行する。

#### シナリオ: 新着エントリの検出
GIVEN RSS プラグインに有効なフィード URL が 1 件以上設定されている
AND `PollIntervalSec` が 900 である
WHEN 前回ポーリングから 900 秒が経過する
THEN システムは全フィードを並列に HTTP GET で取得する
AND `gofeed.Parser` でエントリ一覧を parse する
AND 後続要件「閾値による絞り込み」で決まる閾値より新しい pubDate のエントリのみ選択する
AND 選択されたエントリを pubDate 昇順に sort する
AND 各 Item を `Core.Submit()` に直列で送信する

#### シナリオ: 新着エントリなし
GIVEN 前回ポーリング以降に新着エントリがない
WHEN ポーリングサイクルが実行される
THEN システムは Item を送信せず完了する

#### シナリオ: 1 フィードの HTTP エラー
GIVEN 1 つのフィードが HTTP 500 を返す
WHEN ポーリングサイクルが実行される
THEN システムは警告をログに記録する
AND 他のフィードの fetch と Submit は継続する
AND 次のポーリング周期で自動再試行する

#### シナリオ: 1 フィードの HTTP タイムアウト
GIVEN 1 つのフィードが 10 秒以上応答しない
WHEN ポーリングサイクルがそのフィードを取得しようとする
THEN システムは当該フィードの fetch を中断する
AND 警告をログに記録し、他のフィードの処理を続行する

---

### 要件: 閾値による絞り込み (初回洪水 + 再起動ギャップ対策)
WHEN RSS プラグインがポーリングサイクルで Submit 対象エントリを選択する場合、
システムは閾値 `T = max(LastLabeledAtBySource("rss"), now - 24h)` を計算し、
`pubDate > T` のエントリのみを Submit しなければならない (SHALL)。

`LastLabeledAtBySource("rss")` は DB に保存された source="rss" の最新 `labeled_at` を返す。
rss アイテムが未保存の場合は zero time を返し、結果として閾値は `now - 24h` になる。

#### シナリオ: 初回起動 (rss アイテム未保存)
GIVEN rss プラグインが初めて起動する
AND DB に source="rss" のアイテムが 0 件
AND フィードに過去 1 週間分のエントリが含まれる
WHEN 初回ポーリングが実行される
THEN 閾値は `now - 24h`
AND 過去 24 時間以内に published されたエントリのみが Submit される
AND それより古いエントリは skip される

#### シナリオ: 再起動ギャップの復旧
GIVEN 前回のセッションで最後にラベリングされた rss アイテムの `labeled_at = T_last`
AND T_last が now - 24h より新しい
AND フィードには T_last 以降に published されたエントリが含まれる
WHEN 再起動後の最初のポーリングが実行される
THEN 閾値は T_last
AND T_last より新しい pubDate のエントリが Submit される (ダウンタイム中の記事を復旧)

#### シナリオ: 冪等性の保護
GIVEN 同じ source_id のエントリが既に DB に存在する
WHEN 閾値内であっても再度 Submit される
THEN DB の UNIQUE(source, source_id) 制約により重複保存はスキップされる
AND 再ラベリングは発生しない (core.Submit 側の冪等ロジック)

---

### 要件: エントリの一意識別
WHEN RSS エントリを Item に変換する場合、
システムは以下の優先順位で `source_id` のキーを決定しなければならない (SHALL):

1. entry の GUID が `http(s)://` で始まる場合: GUID を URL として正規化し、SHA256 の先頭 16 文字を使う
2. それ以外で entry.Link が存在する: entry.Link を正規化して SHA256 短縮
3. それも無い: `title + published` の SHA256 短縮 (フォールバック)

URL 正規化は以下を適用する:
- scheme を https に統一
- host を lowercase、`www.` prefix を除去
- クエリパラメータ `utm_*` / `fbclid` / `gclid` / `ref` を除去
- fragment 除去
- path の trailing slash 除去

feed-host の prefix は source_id に含めない (同一記事が複数フィードに掲載されても 1 件として扱う)。

#### シナリオ: GUID が URL 形式
GIVEN エントリに `<guid>https://example.com/posts/42</guid>` がある
WHEN source_id が生成される
THEN URL として正規化された上で SHA256 短縮される

#### シナリオ: GUID が数値/ハッシュ形式
GIVEN エントリに `<guid>d5f3a7...</guid>` (URL でない) がある
AND entry.Link が `https://example.com/posts/42`
WHEN source_id が生成される
THEN entry.Link を URL 正規化した hash が使われる

#### シナリオ: トラッキングパラメータの除去
GIVEN URL `https://example.com/posts/42?utm_source=twitter`
WHEN 正規化される
THEN 最終 URL は `https://example.com/posts/42`
AND 別セッションで `?utm_source=mail` がついた同記事も同じ source_id になる

---

### 要件: Bonsai grammar と prompt (RSS 専用)
WHEN RSS プラグインが Core に登録される場合、
システムは RSS 専用の GBNF grammar と prompt テンプレートを `Core.RegisterPlugin()` の引数として提供しなければならない (SHALL)。

- **urgency enum は 3 値**: `"should_check" | "can_wait" | "ignore"` (grammar で `urgent` を除外)
- **category enum は 5 値**: `"llm_research" | "llm_news" | "dev_tools" | "swe" | "other"`
- prompt は `/no_think` prefix、分類優先順位 rule、5 件の few-shot example、`{notification_json}` placeholder を持つ

Prompt の優先順位 rule (実装で prompt に埋め込む):
1. 研究・論文深掘り → `llm_research`
2. LLM 製品の発表・リリース → `llm_news`
3. 特定ツール / ライブラリ / CLI / エディタ拡張の紹介 → `dev_tools`
4. 言語 / FW / 設計等のツール非依存記事 → `swe`
5. 上記に当てはまらない → `other`

LLM 関連記事は `llm_*` を優先。ただし主題が「tool X を使ってこう書いた」なら `dev_tools`。

#### シナリオ: urgent を返さない
GIVEN Bonsai が RSS エントリに対してラベリングを実行する
WHEN grammar が適用される
THEN 出力 urgency は `should_check` / `can_wait` / `ignore` のいずれか
AND `urgent` は grammar で生成不能

#### シナリオ: 無効な category の拒否
GIVEN Bonsai が誤って `"anime"` 等の枠外 category を返そうとする
WHEN grammar が適用される
THEN 許可された 5 値のみが返る

---

### 要件: Item.Content と Metadata の構造
WHEN RSS プラグインが entry を Item に変換する場合、
システムは以下を満たさなければならない (SHALL):

- `Item.Content` は entry の description / content を HTML タグ除去した text、400 字で truncate
- `Item.Title` は entry.Title
- `Item.URL` は entry.Link (正規化前の original)
- `Item.Timestamp` は entry.PublishedParsed、無ければ entry.UpdatedParsed
- `Item.SourceID` は前述の「エントリの一意識別」で生成
- `Item.Metadata` は以下を含む:
  - `feed_url` (必須): ポーリング元のフィード URL
  - `feed_name` (必須): config の `name`、無指定なら feed_url の host
  - `author` (任意): entry.Author.Name
  - `categories` (任意): entry のタグをカンマ区切り
  - `guid` (任意): 元の GUID (デバッグ用、dedup には使わない)
  - `urgency_floor` (任意): config の urgency_floor 値、無指定なら空文字列

#### シナリオ: HTML 除去
GIVEN entry.Content が `<p>Hello <b>world</b></p><script>evil()</script>`
WHEN Item.Content が生成される
THEN `"Hello world"` が Item.Content に入る (script 内容は除去)

#### シナリオ: 長文 truncate
GIVEN entry.Description が 1000 字の HTML
WHEN Item.Content が生成される
THEN HTML 除去後、400 字で truncate され末尾に `...` が付く

#### シナリオ: feed_name fallback
GIVEN config で feed に name が指定されていない (url = `https://example.com/feed`)
WHEN Item.Metadata["feed_name"] が設定される
THEN `"example.com"` (host) が fallback として入る

---

### 要件: per-feed urgency_floor の伝達
WHEN config の feed 定義に `urgency_floor` が指定されている場合、
システムは当該フィードから生成した各 Item の `Metadata["urgency_floor"]` に当該値をコピーしなければならない (SHALL)。
Core は別要件「Metadata ベースの urgency floor 適用」(core spec) に従って post-process する。

#### シナリオ: Anthropic News の urgency 格上げ
GIVEN config に Anthropic News の feed エントリで `urgency_floor = "should_check"` が設定されている
AND Bonsai が当該フィードのエントリに `urgency = "can_wait"` を返した
WHEN Core が Submit を完了する
THEN 保存される Item の urgency は `should_check` に格上げされる

#### シナリオ: floor 無指定のフィードは格上げなし
GIVEN config に urgency_floor 無指定のフィード (Zenn 等)
AND Bonsai が `urgency = "can_wait"` を返す
WHEN Core が post-process を実行する
THEN urgency は `can_wait` のまま変化しない

---

### 要件: レートリミット (429) の尊重
WHEN フィードが HTTP 429 とともに `Retry-After` ヘッダを返した場合、
システムは当該フィードの次回 fetch をヘッダ指定時刻まで保留しなければならない (SHALL)。
他のフィードの処理には影響しない。

#### シナリオ: 429 を受けた直後のポーリング
GIVEN フィード A が `HTTP 429`、`Retry-After: 600` を返した
WHEN 次のポーリングサイクル (15 分後) が実行される
THEN フィード A は 10 分経過していないため skip される
AND 警告ログに `retry_after=600` が記録される
AND 他のフィードは通常通り fetch される

#### シナリオ: Retry-After 経過後の復帰
GIVEN フィード A の `Retry-After` 時刻を超過している
WHEN ポーリングサイクルが実行される
THEN フィード A は通常通り fetch される

---

### 要件: プラグイン設定の opt-in
WHEN `config.toml` に `[plugins.rss]` セクションが存在しない、または `enabled = false` の場合、
システムは RSS プラグインを登録してはならない (SHALL NOT)。

#### シナリオ: 既存ユーザーの config 保護
GIVEN 既存の `config.toml` には `[plugins.rss]` セクションが存在しない
WHEN `sentei serve` が起動する
THEN RSS プラグインは登録されない
AND `sentei plugin list` には `git` のみ表示される
AND ログにエラーや警告は出ない

#### シナリオ: 明示的な有効化
GIVEN `config.toml` に `[plugins.rss] enabled = true` と設定されている
AND `[[plugins.rss.feeds]]` が 1 件以上ある
WHEN `sentei serve` が起動する
THEN RSS プラグインが登録される
AND ログに「RSS プラグイン起動、<n> フィード」と記録される

#### シナリオ: 設定 validation
GIVEN config の `[[plugins.rss.feeds]]` のうち 1 件の URL が `ftp://...`
WHEN `sentei serve` が起動する
THEN fatal ログとともに起動失敗する
AND ユーザーに URL の scheme 修正を促すエラーメッセージが出る
