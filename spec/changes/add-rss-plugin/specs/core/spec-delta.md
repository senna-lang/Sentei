# 仕様差分: コアデーモン

対象仕様: `spec/specs/core/spec.md`

RSS プラグインの per-feed `urgency_floor` を受けるため、Core 側に「Item の metadata による urgency 格上げヒント」を解釈する汎用機構を ADDED する。

プラグイン固有の rule を Core に hard-code せず、各プラグインが metadata で宣言する契約方式を採る。これにより将来のプラグイン (arxiv / Slack / newsletter 等) も同じ機構で独自の格上げルールを持たせられる。

---

## ADDED 要件

### 要件: Metadata ベースの urgency floor 適用
WHEN Bonsai がラベリングを完了した Item に `Metadata["urgency_floor"]` が設定されている場合、
Core は Bonsai 応答の urgency と floor を比較し、floor より低ければ floor に格上げしなければならない (SHALL)。
比較には `plugin.UrgencyRank` (ignore=0, can_wait=1, should_check=2, urgent=3) を使う。

`urgency_floor` が空文字列、未設定、または `plugin.UrgencyRank` に存在しない値の場合、Core は格上げを行わない (no-op)。

post-process は `SaveLabeledItem` の直前で実行する。DB に保存される urgency は格上げ後の値。

#### シナリオ: can_wait → should_check の格上げ
GIVEN Item の Metadata に `urgency_floor = "should_check"` が含まれる
AND Bonsai が urgency = `can_wait` を返す
WHEN Core が post-process を実行する
THEN 最終 label の urgency は `should_check` になる
AND DB には `should_check` として保存される

#### シナリオ: 格上げ不要 (既に floor 以上)
GIVEN Item の Metadata に `urgency_floor = "can_wait"` が含まれる
AND Bonsai が urgency = `should_check` を返す
WHEN Core が post-process を実行する
THEN urgency は `should_check` のまま変化しない

#### シナリオ: Metadata に urgency_floor が無い
GIVEN Item の Metadata に `urgency_floor` キーが存在しない (または空文字列)
WHEN Core が post-process を実行する
THEN urgency は Bonsai 応答のまま変化しない (no-op)

#### シナリオ: 不正な urgency_floor 値
GIVEN Item の Metadata に `urgency_floor = "super_urgent"` (enum 外) が含まれる
AND Bonsai が urgency = `can_wait` を返す
WHEN Core が post-process を実行する
THEN urgency は `can_wait` のまま変化しない (不正値は無視して no-op)
AND 警告は出さない (ラベリング経路を止めない)

#### シナリオ: urgent は常に最上位として扱う
GIVEN Item の Metadata に `urgency_floor = "should_check"` が含まれる
AND Bonsai が urgency = `urgent` を返す
WHEN Core が post-process を実行する
THEN urgency は `urgent` のまま変化しない (UrgencyRank[urgent]=3 > UrgencyRank[should_check]=2)
