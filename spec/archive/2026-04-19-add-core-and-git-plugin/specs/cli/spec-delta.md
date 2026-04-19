# 仕様差分: CLI

対象仕様: `spec/specs/cli/spec.md`

## 追加要件

### 要件: init コマンド
WHEN ユーザーが `sentei init` を実行した場合、
システムは `~/.sentei/` ディレクトリとデフォルト設定ファイルを作成しなければならない (SHALL)。

#### シナリオ: 初回 init
GIVEN `~/.sentei/` ディレクトリが存在しない
WHEN ユーザーが `sentei init` を実行する
THEN システムは `~/.sentei/` ディレクトリを作成する
AND デフォルトの `config.yaml` を生成する
AND "初期設定が完了しました" と表示する

#### シナリオ: 既に初期化済み
GIVEN `~/.sentei/` ディレクトリが既に存在する
WHEN ユーザーが `sentei init` を実行する
THEN システムは既存設定を上書きしない
AND "既に初期化されています" と表示する

---

### 要件: start コマンド
WHEN ユーザーが `sentei start` を実行した場合、
システムはコアデーモンをバックグラウンドで起動しなければならない (SHALL)。

#### シナリオ: 正常な起動
GIVEN デーモンが停止している
AND init が完了済みである
WHEN ユーザーが `sentei start` を実行する
THEN デーモンがバックグラウンドで起動する
AND "デーモンを起動しました (PID: XXXX)" と表示する

#### シナリオ: 既に起動中
GIVEN デーモンが既に起動している
WHEN ユーザーが `sentei start` を実行する
THEN システムは新しいプロセスを起動しない
AND "デーモンは既に起動中です (PID: XXXX)" と表示する

---

### 要件: stop コマンド
WHEN ユーザーが `sentei stop` を実行した場合、
システムは稼働中のデーモンをグレースフルに停止しなければならない (SHALL)。

#### シナリオ: 正常な停止
GIVEN デーモンが起動中である
WHEN ユーザーが `sentei stop` を実行する
THEN デーモンがグレースフルに停止する
AND "デーモンを停止しました" と表示する

#### シナリオ: デーモン未起動
GIVEN デーモンが起動していない
WHEN ユーザーが `sentei stop` を実行する
THEN "デーモンは起動していません" と表示する

---

### 要件: status コマンド
WHEN ユーザーが `sentei status` を実行した場合、
システムはデーモンの動作状態と有効プラグインの情報を表示しなければならない (SHALL)。

#### シナリオ: 稼働中の状態表示
GIVEN デーモンが起動中で Git プラグインが有効
WHEN ユーザーが `sentei status` を実行する
THEN 以下の情報が表示される:
- デーモン状態: 稼働中 (PID: XXXX)
- Bonsai 接続: 正常 (localhost:8080)
- 有効プラグイン: git
- 最終通知ポーリング: YYYY-MM-DD HH:MM:SS
- 最終サーベイ: YYYY-MM-DD HH:MM:SS
- 保存済みアイテム数: N 件

#### シナリオ: 停止中の状態表示
GIVEN デーモンが停止している
WHEN ユーザーが `sentei status` を実行する
THEN "デーモンは停止中です" と表示する

---

### 要件: list コマンド
WHEN ユーザーが `sentei list` を実行した場合、
システムは保存済みアイテムを一覧表示しなければならない (SHALL)。
各アイテムは title をプライマリテキストとして表示し、Bonsai の urgency + category ラベルを付与する。

#### シナリオ: urgency フィルタ付き一覧
GIVEN urgency "urgent" のアイテムが 2 件、"should_check" が 5 件保存されている
WHEN ユーザーが `sentei list --urgency urgent` を実行する
THEN urgency "urgent" の 2 件のみが表示される
AND 各アイテムには [category] タイトル (@author) 形式で表示される

#### シナリオ: source フィルタ付き一覧
GIVEN source "git" のアイテムが 3 件、"rss" が 2 件保存されている
WHEN ユーザーが `sentei list --source git` を実行する
THEN source "git" の 3 件のみが表示される

#### シナリオ: フィルタなし全件表示
GIVEN アイテムが 10 件保存されている
WHEN ユーザーが `sentei list` を実行する
THEN 全 10 件が urgency 順（urgent → should_check → can_wait → ignore）で表示される

#### シナリオ: アイテムなし
GIVEN 保存済みアイテムがない
WHEN ユーザーが `sentei list` を実行する
THEN "アイテムがありません" と表示する

---

### 要件: board コマンド
WHEN ユーザーが `sentei board` を実行した場合、
システムはバッチサーベイ型で収集・ラベリングされたデータを元に、リポジトリ別掲示板をテンプレートで表示しなければならない (SHALL)。

#### シナリオ: 全リポジトリの掲示板一覧
GIVEN 3 つの監視リポジトリの掲示板がある
WHEN ユーザーが `sentei board` を実行する
THEN 各リポジトリの掲示板が表示される
AND urgent アイテムがあるリポジトリが先頭に来る

#### シナリオ: 特定リポジトリの掲示板
GIVEN 複数リポジトリの掲示板がある
WHEN ユーザーが `sentei board arxiv-compass` を実行する
THEN "arxiv-compass" の掲示板のみが詳細表示される

#### シナリオ: 掲示板が未生成
GIVEN まだサーベイが一度も実行されていない
WHEN ユーザーが `sentei board` を実行する
THEN "掲示板がまだ生成されていません。デーモン起動後、最初のサーベイ完了をお待ちください" と表示する

---

### 要件: plugin list コマンド
WHEN ユーザーが `sentei plugin list` を実行した場合、
システムは利用可能なプラグインとその有効/無効状態を表示しなければならない (SHALL)。

#### シナリオ: プラグイン一覧表示
GIVEN git プラグインが有効、rss プラグインが無効
WHEN ユーザーが `sentei plugin list` を実行する
THEN 以下が表示される:
- git: 有効（通知: 1分間隔 / サーベイ: 1時間間隔）
- rss: 無効

---

### 要件: CLI 出力フォーマット
システムは CLI 出力を色付きで見やすくフォーマットしなければならない (SHALL)。

#### シナリオ: urgency に応じた色分け
GIVEN urgency "urgent" のアイテムがある
WHEN list または board コマンドで表示される
THEN "urgent" のアイテムは赤色で表示される
AND "should_check" は黄色で表示される
AND "can_wait" はデフォルト色で表示される
AND "ignore" はグレーで表示される
