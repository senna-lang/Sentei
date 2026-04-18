# Spec Delta: macOS アプリ

対象仕様: `spec/specs/macos-app/spec.md`

## ADDED Requirements

### Requirement: メニューバー常駐
WHILE sentei macOS アプリが起動中の場合、
システムは macOS メニューバーにアイコンを常駐表示しなければならない (SHALL)。
Dock にはアプリアイコンを表示しない（LSUIElement）。

#### Scenario: アプリ起動時のメニューバー表示
GIVEN sentei macOS アプリが起動される
WHEN アプリの初期化が完了する
THEN メニューバーにアイコンが表示される
AND Dock にはアイコンが表示されない

#### Scenario: アイコンクリックでポップオーバー表示
GIVEN メニューバーにアイコンが表示されている
WHEN ユーザーがアイコンをクリックする
THEN ポップオーバーが表示される
AND 即座に最新データが fetch される

#### Scenario: メニューバーアイコンの状態表現（通常）
GIVEN sentei デーモンが稼働中で urgent アイテムがない
WHEN メニューバーアイコンが描画される
THEN 通常のアイコンが表示される

#### Scenario: メニューバーアイコンの状態表現（urgent あり）
GIVEN sentei デーモンが稼働中で urgent アイテムが 3 件ある
WHEN メニューバーアイコンが描画される
THEN アイコンに赤いバッジ「3」が表示される

#### Scenario: メニューバーアイコンの状態表現（未接続）
GIVEN sentei デーモンが停止中である
WHEN メニューバーアイコンが描画される
THEN アイコンがグレーアウトされる

---

### Requirement: ポップオーバー表示
WHEN ユーザーがメニューバーアイコンをクリックした場合、
システムは直近のアイテム（最大 10 件、urgency 順）をポップオーバーで表示しなければならない (SHALL)。
ポップオーバーのフッターには「ダッシュボードを開く」と「終了」ボタンを常設する。

#### Scenario: アイテムがある場合のポップオーバー
GIVEN sentei デーモンが稼働中で 15 件のアイテムがある
WHEN ポップオーバーが表示される
THEN urgency 順（urgent → should_check → can_wait → ignore）で最大 10 件が表示される
AND 各アイテムにチェックボタン、urgency バッジ、category アイコン、title が表示される
AND フッターに「ダッシュボードを開く」と「終了」ボタンが表示される

#### Scenario: アイテムがない場合のポップオーバー
GIVEN sentei デーモンが稼働中だがアイテムがない
WHEN ポップオーバーが表示される
THEN 「アイテムがありません」と表示される

#### Scenario: デーモン未接続時のポップオーバー
GIVEN sentei デーモンが停止中である
WHEN ポップオーバーが表示される
THEN 接続エラー状態が表示される

#### Scenario: アイテムクリックで URL を開く
GIVEN ポップオーバーに URL 付きのアイテムが表示されている
WHEN ユーザーがアイテム行をクリックする
THEN デフォルトブラウザで該当 URL が開かれる

#### Scenario: チェックボタンでアイテム削除
GIVEN ポップオーバーにアイテムが表示されている
WHEN ユーザーがチェックボタンをクリックする
THEN DELETE /api/items/{source}/{source_id} が呼ばれる
AND アイテムがリストから消える

#### Scenario: 「ダッシュボードを開く」ボタン
GIVEN ポップオーバーが表示されている
WHEN ユーザーが「ダッシュボードを開く」をクリックする
THEN ダッシュボードウィンドウが開かれる

#### Scenario: 「終了」ボタン
GIVEN ポップオーバーが表示されている
WHEN ユーザーが「終了」をクリックする
THEN デーモンが停止される
AND アプリが終了する

---

### Requirement: ダッシュボードウィンドウ
WHEN ユーザーがダッシュボードウィンドウを開いた場合、
システムはサイドバー付きのメインウィンドウでアイテム一覧と掲示板を表示しなければならない (SHALL)。
ウィンドウの閉じるボタン（×）はウィンドウを隠すだけで、アプリは終了しない。

#### Scenario: サイドバーナビゲーション
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーがサイドバーを操作する
THEN 以下のセクションが選択可能である:
- 「アイテム」セクション（urgency 別のカウントバッジ付き: urgent / should_check / 全件）
- 「掲示板」セクション（リポジトリ名が個別にネスト表示）
- 「ステータス」セクション

#### Scenario: アイテム一覧の表示
GIVEN ダッシュボードで「アイテム」が選択されている
AND 20 件のアイテムがある
WHEN アイテム一覧が表示される
THEN urgency 順でアイテムがカード形式で表示される
AND 各カードにチェックボタン、urgency 色帯、category アイコン、title、author、時刻が含まれる
AND ポップオーバーと同じ ItemRow コンポーネントが使用される

#### Scenario: アイテム一覧のフィルタリング
GIVEN ダッシュボードで「アイテム」が選択されている
WHEN ユーザーが urgency フィルタで "urgent" を選択する
THEN urgency が "urgent" のアイテムのみが表示される

#### Scenario: アイテムのチェックボタン
GIVEN ダッシュボードのアイテム一覧にアイテムが表示されている
WHEN ユーザーがチェックボタンをクリックする
THEN DELETE /api/items/{source}/{source_id} が呼ばれる
AND アイテムがリストから消える

#### Scenario: 掲示板の表示
GIVEN ダッシュボードで特定リポジトリの掲示板が選択されている
WHEN 掲示板が表示される
THEN 掲示板テキストが monospace プレーンテキストで表示される

#### Scenario: ウィンドウを閉じる
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーが閉じるボタン（×）をクリックする
THEN ウィンドウが非表示になる
AND アプリはメニューバーに常駐し続ける

---

### Requirement: デーモン自動起動
WHEN macOS アプリが起動した場合、
システムは sentei デーモンが稼働中か確認し、停止中であれば自動的に `sentei serve` プロセスを起動しなければならない (SHALL)。

#### Scenario: デーモン未起動時の自動起動
GIVEN sentei デーモンが停止中である
WHEN macOS アプリが起動する
THEN アプリは `sentei serve` プロセスをバックグラウンドで spawn する
AND デーモンの起動を待ってから接続する

#### Scenario: デーモン起動済みの場合
GIVEN sentei デーモンが既に稼働中である
WHEN macOS アプリが起動する
THEN アプリは新たなプロセスを spawn しない
AND 既存のデーモンに接続する

#### Scenario: アプリ終了時のデーモン停止
GIVEN アプリがデーモンを spawn して起動した
WHEN ユーザーがアプリを終了する
THEN アプリは spawn したデーモンプロセスを停止する

---

### Requirement: API ポーリング（動的間隔）
WHILE macOS アプリが起動中の場合、
システムは動的な間隔で sentei デーモンの REST API をポーリングし、表示データを更新しなければならない (SHALL)。

#### Scenario: バックグラウンドポーリング
GIVEN ポップオーバーもウィンドウも表示されていない
WHEN ポーリング間隔が経過する
THEN 60 秒間隔で GET /api/items と GET /api/status を呼び出す
AND メニューバーアイコンのバッジを更新する

#### Scenario: フォアグラウンドポーリング
GIVEN ポップオーバーまたはダッシュボードウィンドウが表示中である
WHEN ポーリング間隔が経過する
THEN 15 秒間隔で GET /api/items と GET /api/status を呼び出す
AND 表示データが最新の状態に更新される

#### Scenario: 表示時の即時 fetch
GIVEN ポップオーバーまたはダッシュボードウィンドウが開かれる
WHEN 画面が表示される瞬間
THEN 即座に GET /api/items を呼び出す
AND 最新データで表示を更新する

#### Scenario: デーモン接続断
GIVEN アプリが起動中で sentei デーモンが停止する
WHEN ポーリングが接続エラーを検知する
THEN アプリは「未接続」状態を表示する
AND ポーリングを継続し、デーモン再起動時に自動復帰する

#### Scenario: デーモン再接続
GIVEN アプリが「未接続」状態である
WHEN sentei デーモンが起動してポーリングが成功する
THEN アプリは「接続中」状態に復帰する
AND 最新データで表示を更新する

---

### Requirement: urgent 通知（UserNotifications）
WHEN 新しい urgency "urgent" のアイテムが検出された場合、
システムは macOS の UserNotifications framework で通知を表示しなければならない (SHALL)。

#### Scenario: urgent アイテムの通知
GIVEN アプリが起動中でポーリングが実行される
WHEN 前回のポーリング以降に新しい urgent アイテムが検出される
THEN macOS 通知が表示される（タイトル: "sentei"、本文: "[category] title"）

#### Scenario: 非 urgent アイテムは通知しない
GIVEN アプリが起動中でポーリングが実行される
WHEN 新しいアイテムの urgency が "should_check" 以下である
THEN macOS 通知はトリガーされない

---

### Requirement: デザインテーマ
システムはダーク寄りモダンデザイン（Raycast/Linear 風）で表示しなければならない (SHALL)。

#### Scenario: urgency 色分け
GIVEN アイテムが表示される場面（ポップオーバーまたはダッシュボード）
WHEN urgency ラベルが描画される
THEN "urgent" は赤系 (#FF4444) で表示される
AND "should_check" は黄系 (#FFB020) で表示される
AND "can_wait" はグレー (#888888) で表示される
AND "ignore" は薄グレー (#555555) で表示される

#### Scenario: ダークモード
GIVEN macOS アプリが起動する
WHEN ウィンドウとポップオーバーが描画される
THEN ダークカラースキームが適用される

---

## MODIFIED Requirements

### Requirement: 緊急通知（コアデーモン仕様）
**Previous**: urgency "urgent" のアイテムに対して osascript で macOS 通知を発行する

WHEN Item が urgency "urgent" とラベリングされた場合、
コアデーモンは通知を発行しない (SHALL NOT)。
通知の責務は macOS アプリの UserNotifications に委譲する。

#### Scenario: デーモン単体での通知なし
GIVEN macOS アプリが起動していない
WHEN urgency "urgent" のアイテムがラベリングされる
THEN デーモンは通知を発行しない

---

### Requirement: アイテム削除 API（コアデーモン仕様）

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

### Requirement: アイテム一覧と掲示板の棲み分け (2026-04-18 追加)
WHEN macOS アプリがデータを表示する場合、
システムは「アイテム一覧」と「掲示板」で異なる責務のデータを表示しなければならない (SHALL)。

- **アイテム一覧** = 自分宛の GitHub 通知のみ（`survey_type` メタデータを持たないアイテム）
- **掲示板** = 監視対象リポジトリの**今日**の活動（`survey_type` を持ち、`timestamp >= startOfToday` のアイテム）

#### Scenario: アイテム一覧は自分宛のみ
GIVEN ストレージに通知由来 (survey_type なし) とサーベイ由来 (survey_type: merged_pr など) が混在する
WHEN アイテム一覧がレンダリングされる
THEN 通知由来のアイテムだけが表示される
AND urgentCount / recentItems / サイドバーの urgency 別カウントも通知のみを対象とする

#### Scenario: 掲示板は今日の活動のみ
GIVEN サーベイ由来アイテムが過去 30 日分ストレージに存在する
WHEN `GET /api/board` が呼ばれる
THEN `timestamp >= 今日の 00:00` のサーベイアイテムだけが各リポジトリの掲示板に含まれる
AND 過去に一度でもサーベイされたリポジトリは、今日の活動がゼロでも掲示板枠が返る（本文は「特に動きはありません」）

#### Scenario: 通知は掲示板に混ざらない
GIVEN `ncdcdev/foo` から通知だけ来ており、survey 対象には入っていない
WHEN 掲示板 API が呼ばれる
THEN `ncdcdev/foo` の掲示板は返らない（survey_type のあるアイテムが存在しないため）

---

### Requirement: Git サーベイの対象種別 (2026-04-18 追加)
WHEN Git プラグインがサーベイを実行する場合、
システムは以下 4 種のレポジトリ活動を取得して `survey_type` メタデータで識別しなければならない (SHALL)。

- `merged_pr`: 過去 30 日以内にマージされた PR
- `open_pr`: state=open かつ直近 30 日以内に更新された PR
- `new_issue`: 直近 30 日以内に作成された Issue
- `release`: 直近 30 日以内の published リリース（Draft 除外）

#### Scenario: 4 種別の収集
GIVEN サーベイ対象リポジトリに関連 PR / Issue / リリースがある
WHEN `surveyRepo` が呼ばれる
THEN 各項目がそれぞれ対応する `survey_type` メタデータ付きで Submit される

---

### Requirement: 掲示板サマリー（テンプレート生成） (2026-04-18 更新)
**Previous**: Bonsai のフリーテキスト生成で 1-2 文のサマリーを得る

WHEN 掲示板がレンダリングされる場合、
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
THEN サマリーは空文字列を返し、掲示板にはサマリー行が出力されない

---

### Requirement: ダッシュボードウィンドウの単一インスタンス (2026-04-18 追加)
WHILE アプリが起動中の場合、
システムはダッシュボードウィンドウを常に単一インスタンスで維持しなければならない (SHALL)。

#### Scenario: 開いている状態で再度開く
GIVEN ダッシュボードウィンドウが既に表示されている
WHEN ユーザーがポップオーバーから「ダッシュボードを開く」を再度クリックする
THEN 既存のウィンドウがフォアグラウンドに来る
AND 新しいウィンドウは作成されない

（実装メモ: SwiftUI の `WindowGroup` は複製されるため `Window` を使う）

---

### Requirement: UserNotifications の bundle ガード (2026-04-18 追加)
WHILE アプリが `swift run` など .app 外から起動されている場合、
システムは UNUserNotificationCenter を呼ばず通知をスキップしなければならない (SHALL)。

非バンドル実行下で `UNUserNotificationCenter.current()` を呼ぶと NSException で強制終了するため、
`Bundle.main.bundlePath.hasSuffix(".app")` で判定する。`bundleIdentifier != nil` 判定は swift run でも true を返すため不可。

#### Scenario: bundle 外実行時のスキップ
GIVEN アプリが `swift run` で起動されている
WHEN urgent アイテムを検出する
THEN UN API は呼ばれず、アプリはクラッシュしない

#### Scenario: bundle 内実行時の通知
GIVEN アプリが `Sentei.app` として起動されている
WHEN urgent アイテムを検出する
THEN UserNotifications で通知が発行される
