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
