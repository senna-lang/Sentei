# コアデーモン仕様

sentei のコアデーモン（`sentei serve`）の振る舞いに関する仕様。
ラベリングパイプライン、REST API、ストレージ、サマリーレンダリングを含む。

## Requirements

### Requirement: 緊急通知の委譲
WHEN Item が urgency "urgent" とラベリングされた場合、
コアデーモンは通知を発行しない (SHALL NOT)。
通知の責務は macOS アプリの UserNotifications に委譲する。

#### Scenario: デーモン単体での通知なし
GIVEN macOS アプリが起動していない
WHEN urgency "urgent" のアイテムがラベリングされる
THEN デーモンは通知を発行しない

---

### Requirement: アイテム削除 API
WHEN クライアントが DELETE /api/items/{source}/{source_id} を送信した場合、
システムは該当アイテムを SQLite から物理削除しなければならない (SHALL)。

#### Scenario: アイテムの正常な削除
GIVEN source "git"、source_id "notif-123" のアイテムが存在する
WHEN DELETE /api/items/git/notif-123 が送信される
THEN アイテムが SQLite から物理削除される
AND HTTP 200 が返される

#### Scenario: 存在しないアイテムの削除
GIVEN source "git"、source_id "nonexistent" のアイテムが存在しない
WHEN DELETE /api/items/git/nonexistent が送信される
THEN HTTP 404 が返される

---

### Requirement: サマリー（テンプレート生成）
WHEN サマリーがレンダリングされる場合、
システムは `survey_type` / カテゴリ / urgency / author から決定的テンプレートでサマリー文を生成しなければならない (SHALL)。
LLM フリー生成は品質不安定なため使用しない。

#### Scenario: テンプレートサマリーの生成
GIVEN 今日の活動が「マージ 2 件 / 新規 Issue 1 件 / @alice と @bob が関与」である
WHEN サマリーが生成される
THEN 例: `PR が 2 件マージされました、Issue が 1 件起票。 関与: @alice, @bob。`
AND LLM 呼び出しは発生しない

#### Scenario: 活動ゼロ時
GIVEN 今日の活動がゼロ
WHEN サマリーが生成される
THEN サマリーは空文字列を返し、サマリー出力にはサマリー行が出力されない

---

### Requirement: Metadata ベースの urgency floor 適用
WHEN Bonsai がラベリングを完了した Item に `Metadata["urgency_floor"]` が設定されている場合、
Core は Bonsai 応答の urgency と floor を比較し、floor より低ければ floor に格上げしなければならない (SHALL)。
比較には `plugin.UrgencyRank` (ignore=0, can_wait=1, should_check=2, urgent=3) を使う。

`urgency_floor` が空文字列、未設定、または `plugin.UrgencyRank` に存在しない値の場合、Core は格上げを行わない (no-op)。

post-process は `SaveLabeledItem` の直前で実行する。DB に保存される urgency は格上げ後の値。

プラグイン固有のルールを Core に hard-code せず、各プラグインが metadata で宣言する契約方式とする。

#### Scenario: can_wait → should_check の格上げ
GIVEN Item の Metadata に `urgency_floor = "should_check"` が含まれる
AND Bonsai が urgency = `can_wait` を返す
WHEN Core が post-process を実行する
THEN 最終 label の urgency は `should_check` になる
AND DB には `should_check` として保存される

#### Scenario: 格上げ不要 (既に floor 以上)
GIVEN Item の Metadata に `urgency_floor = "can_wait"` が含まれる
AND Bonsai が urgency = `should_check` を返す
WHEN Core が post-process を実行する
THEN urgency は `should_check` のまま変化しない

#### Scenario: Metadata に urgency_floor が無い
GIVEN Item の Metadata に `urgency_floor` キーが存在しない (または空文字列)
WHEN Core が post-process を実行する
THEN urgency は Bonsai 応答のまま変化しない (no-op)

#### Scenario: 不正な urgency_floor 値
GIVEN Item の Metadata に `urgency_floor = "super_urgent"` (enum 外) が含まれる
AND Bonsai が urgency = `can_wait` を返す
WHEN Core が post-process を実行する
THEN urgency は `can_wait` のまま変化しない (不正値は無視して no-op)
AND 警告は出さない (ラベリング経路を止めない)

#### Scenario: urgent は常に最上位として扱う
GIVEN Item の Metadata に `urgency_floor = "should_check"` が含まれる
AND Bonsai が urgency = `urgent` を返す
WHEN Core が post-process を実行する
THEN urgency は `urgent` のまま変化しない (UrgencyRank[urgent]=3 > UrgencyRank[should_check]=2)
