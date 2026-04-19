# 仕様差分: RSS プラグイン

対象仕様: `spec/specs/rss-plugin/spec.md` (**新規作成**)

RSS プラグインは git 通知トリガー型と対称な構成を取る。登録された RSS/Atom フィードをポーリングし、新着エントリを 1 件ずつ Bonsai に渡して urgency + category をラベリングする。サーベイ (batch 集約) は持たない。

---

## ADDED 要件

### 要件: フィードのポーリング
WHILE RSS プラグインが稼働中の場合、
システムは `config.Plugins.Rss.Feeds` に列挙された各フィードを、`PollIntervalSec` (デフォルト 900 秒) 間隔でポーリングしなければならない (SHALL)。

#### シナリオ: 新着エントリの検出
GIVEN RSS プラグインに有効なフィード URL が 1 件以上設定されている
AND `PollIntervalSec` が 900 である
WHEN 前回ポーリングから 900 秒が経過する
THEN システムは各フィードを HTTP GET で取得する
AND `gofeed.Parser` でエントリ一覧を parse する
AND 前回取得分の seen-ids に含まれないエントリのみ Item に変換する
AND 各 Item を `Core.Submit()` でコアに送信する
AND 送信した entry の GUID / URL を seen-ids に追加する

#### シナリオ: 新着エントリなし
GIVEN 前回ポーリング以降に新着エントリがない
WHEN ポーリングサイクルが実行される
THEN システムは Item を送信せず完了する

#### シナリオ: フィードサーバー障害
GIVEN 1 つのフィードが HTTP 500 を返す
WHEN ポーリングサイクルが実行される
THEN システムは警告をログに記録する
AND 次のフィードの処理に進む (全体を止めない)
AND 次のポーリング周期で再試行する

#### シナリオ: HTTP タイムアウト
GIVEN 1 つのフィードが 10 秒以上応答しない
WHEN ポーリングサイクルがそのフィードを取得しようとする
THEN システムは当該フィードの fetch を中断する
AND 警告をログに記録し、次のフィードへ進む

---

### 要件: 初回ポーリング時の洪水抑制
WHEN RSS プラグインが起動後、最初のポーリングサイクルを実行する場合、
システムは取得した全エントリを seen-ids に登録するのみとし、`Core.Submit()` を呼んではならない (SHALL NOT)。

#### シナリオ: 初回起動
GIVEN RSS プラグインが初めて起動する
AND フィード A に過去 100 件のエントリが含まれる
WHEN 初回ポーリングが実行される
THEN 100 件のエントリは seen-ids に記録される
AND `Core.Submit()` は呼ばれない
AND ログに「初回ポーリング、<n> 件を既読として登録」と記録される

#### シナリオ: 2 回目以降
GIVEN 初回ポーリングが完了している
AND フィード A に新規エントリ 2 件が追加された
WHEN 2 回目のポーリングが実行される
THEN 新規 2 件のみ `Core.Submit()` に送信される

---

### 要件: エントリの一意識別
WHEN RSS エントリを Item に変換する場合、
システムは `<guid>` (Atom の `<id>`) を優先して `source_id` のキーに使用しなければならない (SHALL)。
GUID が空の場合は entry の URL を使用する。

#### シナリオ: GUID が存在する場合
GIVEN エントリに `<guid>d5f3a...</guid>` がある
WHEN Item が生成される
THEN `source_id` は `"<feed-host>:d5f3a..."` の形式になる

#### シナリオ: GUID が空の場合
GIVEN エントリに GUID がなく、URL `https://example.com/posts/42` が存在する
WHEN Item が生成される
THEN `source_id` は URL の SHA256 短縮文字列を用いる
AND prefix は同じく `"<feed-host>:..."` になる

#### シナリオ: 同一 source_id の重複投入
GIVEN 同じ GUID のエントリが既にストレージに保存されている
WHEN 同じエントリが再度 Submit される
THEN ストレージ側の UNIQUE(source, source_id) 制約で冪等に扱われ、再ラベリングは発生しない

---

### 要件: RSS 用 Bonsai grammar と prompt
WHEN RSS プラグインが登録される場合、
システムは core のラベリングパイプラインに、RSS 専用の GBNF grammar と prompt テンプレートを提供しなければならない (SHALL)。

- grammar の `category` enum は `"llm_research" | "llm_news" | "dev_tools" | "other"`
- prompt の入力は `title` / `excerpt (最初の 200 文字)` / `feed_name`
- prompt の出力は `{urgency, category, summary}` JSON

#### シナリオ: 有効な category の出力
GIVEN エントリ "New Claude 3.5 Sonnet Benchmarks" が Bonsai に投入される
WHEN Bonsai がラベリングする
THEN `category` は grammar の 4 値のいずれかに限定される

#### シナリオ: 無効な category の拒否
GIVEN Bonsai が誤って `"anime"` のような枠外の文字列を返そうとする
WHEN grammar が適用されている
THEN `anime` は生成できず、grammar で許可された値のみが返る

---

### 要件: Anthropic News の urgency 格上げルール
WHEN Bonsai が Anthropic News フィード (`anthropic.com/news/rss.xml`) 由来のエントリに対して `can_wait` または `ignore` を返した場合、
システムは urgency を `should_check` に格上げしなければならない (SHALL)。
`urgent` と `should_check` はそのまま保持する。

#### シナリオ: can_wait の格上げ
GIVEN Anthropic News フィード由来のエントリ "Announcing Claude 4"
AND Bonsai が `urgency = "can_wait"` を返す
WHEN ルールが適用される
THEN 最終的な urgency は `should_check` になる

#### シナリオ: urgent は据え置き
GIVEN Anthropic News フィード由来のエントリ "Critical Safety Update"
AND Bonsai が `urgency = "urgent"` を返す
WHEN ルールが適用される
THEN urgency は `urgent` のまま変化しない

#### シナリオ: 他フィードでは格上げしない
GIVEN Zenn フィード由来のエントリ
AND Bonsai が `urgency = "can_wait"` を返す
WHEN ルールが適用される
THEN urgency は `can_wait` のまま変化しない

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
AND `feeds` に 1 件以上の URL がある
WHEN `sentei serve` が起動する
THEN RSS プラグインが登録される
AND ログに「RSS プラグイン起動、<n> フィード」と記録される
