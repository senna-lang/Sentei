# macOS アプリ仕様

sentei の macOS メニューバー常駐アプリに関する仕様。

## Requirements

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
システムはサイドバー付きのメインウィンドウでアイテム一覧とサマリーを表示しなければならない (SHALL)。
ウィンドウの閉じるボタン（×）はウィンドウを隠すだけで、アプリは終了しない。

#### Scenario: サイドバーナビゲーション
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーがサイドバーを操作する
THEN 以下のセクションが選択可能である:
- 「アイテム」セクション（urgency 別のカウントバッジ付き: urgent / should_check / 全件）
- 「サマリー」セクション（リポジトリ名が個別にネスト表示）
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

#### Scenario: サマリーの表示
GIVEN ダッシュボードで特定リポジトリのサマリーが選択されている
WHEN サマリーが表示される
THEN サマリーテキストが monospace プレーンテキストで表示される

#### Scenario: ウィンドウを閉じる
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーが閉じるボタン（×）をクリックする
THEN ウィンドウが非表示になる
AND アプリはメニューバーに常駐し続ける

---

### Requirement: ダッシュボードウィンドウの単一インスタンス
WHILE アプリが起動中の場合、
システムはダッシュボードウィンドウを常に単一インスタンスで維持しなければならない (SHALL)。

#### Scenario: 開いている状態で再度開く
GIVEN ダッシュボードウィンドウが既に表示されている
WHEN ユーザーがポップオーバーから「ダッシュボードを開く」を再度クリックする
THEN 既存のウィンドウがフォアグラウンドに来る
AND 新しいウィンドウは作成されない

（実装メモ: SwiftUI の `WindowGroup` は複製されるため `Window` を使う）

---

### Requirement: アイテム一覧とサマリーの棲み分け
WHEN macOS アプリがデータを表示する場合、
システムは「アイテム一覧」と「サマリー」で異なる責務のデータを表示しなければならない (SHALL)。

- **アイテム一覧** = 自分宛の GitHub 通知のみ（`survey_type` メタデータを持たないアイテム）
- **サマリー** = 監視対象リポジトリの**今日**の活動（`survey_type` を持ち、`timestamp >= startOfToday` のアイテム）

#### Scenario: アイテム一覧は自分宛のみ
GIVEN ストレージに通知由来 (survey_type なし) とサーベイ由来 (survey_type: merged_pr など) が混在する
WHEN アイテム一覧がレンダリングされる
THEN 通知由来のアイテムだけが表示される
AND urgentCount / recentItems / サイドバーの urgency 別カウントも通知のみを対象とする

#### Scenario: サマリーは今日の活動のみ
GIVEN サーベイ由来アイテムが過去 30 日分ストレージに存在する
WHEN `GET /api/summary` が呼ばれる
THEN `timestamp >= 今日の 00:00` のサーベイアイテムだけが各リポジトリのサマリーに含まれる
AND 過去に一度でもサーベイされたリポジトリは、今日の活動がゼロでもサマリー枠が返る（本文は「特に動きはありません」）

#### Scenario: 通知はサマリーに混ざらない
GIVEN `ncdcdev/foo` から通知だけ来ており、survey 対象には入っていない
WHEN サマリー API が呼ばれる
THEN `ncdcdev/foo` のサマリーは返らない（survey_type のあるアイテムが存在しないため）

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

### Requirement: UserNotifications の bundle ガード
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

---

### Requirement: デザインテーマ
システムは盆栽 / 剪定モチーフの緑基調ダークテーマで表示しなければならない (SHALL)。
背景は純黒ではなく緑寄りの墨色を基調とし、アクセントには苔色の低彩度グリーンを用いる。
urgency の "urgent" のみ唯一の暖色（柿渋 / 古銅）を割り当て、他のラベルは自然系（苔黄・竹鼠・涸）でまとめる。
テキストは純白を避け、和紙色 (`#E8EDE5`) を最高階層とする。

#### Scenario: ダークモード（緑基調墨色）
GIVEN macOS アプリが起動する
WHEN ウィンドウとポップオーバーが描画される
THEN ダークカラースキームが適用される
AND 背景は墨色 `#141815`（primary）/ `#1A1F1C`（secondary）/ `#222823`（card）で構成される
AND テキストは和紙色 `#E8EDE5`（primary）/ `#A8B0A4`（secondary）/ `#6B756A`（tertiary）の階層で描画される

#### Scenario: 主要アクセントは苔色
GIVEN 接続状態 / focus / primary action が描画される場面
WHEN アクセント色が必要になる
THEN 苔色 `#7BA05B`（accentPrimary）が用いられる
AND hover / セクションヘッダー等の二次アクセントは常緑 `#4A6B3D` を用いる

#### Scenario: urgency の自然色マッピング
GIVEN アイテムが表示される場面（ポップオーバーまたはダッシュボード）
WHEN urgency ラベルが描画される
THEN "urgent" は柿渋 `#C97E4A`（古銅、唯一の暖色）で表示される
AND "should_check" は苔黄 `#C9B560` で表示される
AND "can_wait" は竹鼠 `#7A8A78` で表示される
AND "ignore" は涸 `#4A524A` で表示される

#### Scenario: コントラスト確保
GIVEN 上記パレットでテキストと背景が組み合わされる
WHEN 主要テキストが描画される
THEN textPrimary on backgroundPrimary のコントラスト比は 4.5:1 以上（WCAG AA 準拠）を満たす
AND urgent on backgroundPrimary のコントラスト比は 3:1 以上を満たす（UI コンポーネント基準）

#### Scenario: 純黒・純白の不使用
GIVEN テーマパレットが定義される
WHEN 背景色および主要テキスト色が指定される
THEN 純黒 `#000000` と純白 `#FFFFFF` は背景・主要テキストいずれにも用いられない
