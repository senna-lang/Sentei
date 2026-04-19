# 仕様差分: Git プラグイン

対象仕様: `spec/specs/git-plugin/spec.md`

Git プラグインは 2 つの動作モードを持つ:
- **通知トリガー型（リアルタイム）**: GitHub 通知を 1 分間隔でポーリングし、Bonsai で urgency + category をラベル付け
- **バッチサーベイ型（定期）**: 監視リポジトリの活動を収集し、各アイテムに Bonsai でラベル付け、テンプレートで掲示板を生成

Bonsai は「ラベリングエンジン」として urgency + category の分類のみ担当。
掲示板テキストの生成は Bonsai に頼らず、ラベル付き構造化データをテンプレートで整形する。

---

## 追加要件

# 通知トリガー型

### 要件: GitHub 通知ポーリング
WHILE Git プラグインが稼働中の場合、
システムは設定された間隔（デフォルト 1 分）で GitHub Notifications API をポーリングしなければならない (SHALL)。
新着通知がある場合のみ Bonsai へ連携し、変化がなければ何もしない。

#### シナリオ: 新着通知の検出
GIVEN Git プラグインに有効な GITHUB_TOKEN が設定されている
AND ポーリング間隔が 1 分である
WHEN 前回ポーリングから 1 分が経過する
THEN システムは GitHub API から未読通知を取得する
AND 前回取得分との差分を検出する
AND 新着通知のみを Item に変換する
AND 各 Item を `Core.Submit()` でコアに送信して Bonsai ラベリングを実行する

#### シナリオ: 新着通知なし
GIVEN 前回ポーリング以降に新着通知がない
WHEN ポーリングサイクルが実行される
THEN システムは Item を送信せず完了する

#### シナリオ: GitHub API レートリミット
GIVEN GitHub API がレートリミットヘッダ付きの HTTP 403 を返す
WHEN ポーリングサイクルがエラーに遭遇する
THEN システムはリセット時刻を含む警告をログに記録する
AND レートリミットがリセットされるまで次のポーリングを待つ

#### シナリオ: 無効なトークン
GIVEN GITHUB_TOKEN が無効または期限切れである
WHEN 最初のポーリングサイクルが実行される
THEN システムは "GitHub 認証に失敗しました" というエラーをログに記録する
AND 次のポーリングサイクルでリトライする

---

### 要件: 通知タイプフィルタリング
IF 設定で監視する通知タイプが指定されている場合、
システムは指定されたタイプに一致する通知のみを処理しなければならない (SHALL)。

#### シナリオ: タイプによるフィルタリング
GIVEN 設定に `notifications: [review_requested, mentioned]` が含まれている
AND "subscribed" タイプの通知が到着する
WHEN プラグインが通知を処理する
THEN "subscribed" 通知はスキップされる
AND "review_requested" と "mentioned" の通知のみが Item になる

#### シナリオ: フィルタ未設定
GIVEN 設定に通知タイプが指定されていない
WHEN プラグインが通知を処理する
THEN すべての通知タイプが Item に変換される

---

### 要件: 通知のラベリング
WHEN GitHub 通知が Item としてコアに送信された場合、
Bonsai は通知の metadata を元に urgency と category をラベル付けしなければならない (SHALL)。

Git プラグインの category enum:
- `pr`: PR レビュー依頼、マージ、コメント
- `issue`: Issue 起票、メンション、コメント
- `ci`: CI の成功・失敗
- `release`: 新規リリース
- `discussion`: Discussion
- `other`: その他

#### シナリオ: メンターからのレビュー依頼
GIVEN notification_type "review_requested" の通知 Item
WHEN Bonsai がラベリングする
THEN urgency は "urgent"、category は "pr" になる

#### シナリオ: CI 失敗通知
GIVEN notification_type "ci_activity"、title に "failed" を含む通知 Item
WHEN Bonsai がラベリングする
THEN urgency は "should_check"、category は "ci" になる

#### シナリオ: リリース PR のマージ
GIVEN notification_type "subscribed"、title に "release" を含む通知 Item
WHEN Bonsai がラベリングする
THEN category は "release" になる（"pr" ではなく）

#### シナリオ: subscribed のみの更新
GIVEN notification_type "subscribed" の通知 Item
WHEN Bonsai がラベリングする
THEN urgency は "can_wait" または "ignore" になる

---

### 要件: 通知 Item 正規化
WHEN GitHub 通知を Item に変換する場合、
システムは source を "git"、source_id を通知 ID、title を通知サブジェクトから設定し、metadata にプラグイン固有フィールドを含めなければならない (SHALL)。

#### シナリオ: レビュー依頼の正規化
GIVEN リポジトリ "arxiv-compass" でユーザー "mentor-username" による "review_requested" タイプの GitHub 通知がある
WHEN 通知が正規化される
THEN Item は以下の値を持つ:
- source: "git"
- source_id: 通知スレッド ID
- title: 通知サブジェクトの PR タイトル
- url: PR の URL
- metadata.repo: "arxiv-compass"
- metadata.notification_type: "review_requested"
- metadata.author: "mentor-username"

---

# バッチサーベイ型

### 要件: リポジトリ定期サーベイ
WHILE Git プラグインが稼働中の場合、
システムは設定された間隔（デフォルト 1 時間）で監視対象リポジトリの最新状況を GitHub API から収集しなければならない (SHALL)。

#### シナリオ: 定期サーベイの実行
GIVEN 監視リポジトリ "senna-lang/arxiv-compass" が設定されている
AND サーベイ間隔が 1 時間である
WHEN 前回サーベイから 1 時間が経過する
THEN システムは以下の情報を GitHub API から取得する:
- 前回サーベイ以降にマージされた PR
- 前回サーベイ以降に起票された Issue
- 自分宛のメンション
- 自分宛のレビュー依頼
- 最近のリリース（あれば）
AND 各アイテムを Item としてコアに送信し、Bonsai でラベリングする

#### シナリオ: 変化なし
GIVEN 前回サーベイ以降にリポジトリに変化がない
WHEN サーベイサイクルが実行される
THEN 掲示板は更新されない（前回の内容を保持）

#### シナリオ: 複数リポジトリのサーベイ
GIVEN 3 つの監視リポジトリが設定されている
WHEN サーベイサイクルが実行される
THEN 各リポジトリに対して個別にサーベイが実行される
AND 各リポジトリの掲示板が独立して更新される

---

### 要件: リポジトリ掲示板生成（テンプレートベース）
WHEN サーベイ結果が収集・ラベリングされた場合、
システムはラベリング済みの構造化データをテンプレートで整形し、リポジトリ別の掲示板を生成しなければならない (SHALL)。

掲示板は Bonsai による自由テキスト生成ではなく、ラベル付きアイテムを urgency 順にグループ化したテンプレート出力である。

#### シナリオ: 活発なリポジトリの掲示板
GIVEN "arxiv-compass" で PR 2 件マージ、Issue 1 件起票、レビュー依頼 1 件がある
AND 各アイテムに Bonsai が urgency + category をラベル付け済み
WHEN 掲示板テンプレートが適用される
THEN 以下の形式で出力される:
```
📋 senna-lang/arxiv-compass (2026-04-17 10:00)

⚡ urgent (1)
  [pr] #47 レビュー依頼: API最適化 (@mentor)

📌 should_check (1)
  [issue] #46 SPECTER2対応 (@researcher)

💤 can_wait (2)
  [pr] #45 マージ済: 丸め誤差修正 (@contributor)
  [pr] #44 マージ済: カバレッジ追加 (@senna)
```

#### シナリオ: 静かなリポジトリの掲示板
GIVEN "my-utils" で前回サーベイ以降に変化がない
WHEN 掲示板テンプレートが適用される
THEN "変化なし" と直近の状態（最終更新日、オープン Issue 数）が表示される

---

### 要件: 掲示板の永続化
WHEN 掲示板が生成された場合、
システムはリポジトリごとの掲示板データ（ラベリング済みアイテムリスト）を SQLite に保存しなければならない (SHALL)。

#### シナリオ: 掲示板の保存
GIVEN "arxiv-compass" の掲示板が生成された
WHEN システムが保存する
THEN リポジトリ識別子、サーベイ時刻、ラベリング済みアイテムの構造化データが SQLite に保存される

#### シナリオ: 掲示板の履歴
GIVEN "arxiv-compass" の掲示板が複数回生成されている
WHEN ユーザーが過去の掲示板を参照する
THEN 直近 N 回分の掲示板が時系列で参照可能である

---

# 共通

### 要件: 設定読み込み
WHEN Git プラグインが起動する場合、
システムは `~/.config/sentei/config.toml` の `[plugins.git]` セクションから通知トリガー型・バッチサーベイ型の両方の設定を読み込まなければならない (SHALL)。

#### シナリオ: 正常な設定
GIVEN 以下の config.toml がある:
```toml
[plugins.git]
enabled = true
github_token = "${GITHUB_TOKEN}"

# 通知トリガー型（リアルタイム）
[plugins.git.notification]
poll_interval = "1m"
types = ["review_requested", "mentioned", "commented"]

# バッチサーベイ型（定期）
[plugins.git.survey]
interval = "1h"

[[plugins.git.survey.repos]]
github = "senna-lang/arxiv-compass"

[[plugins.git.survey.repos]]
github = "senna-lang/logosyncx"

[[plugins.git.survey.repos]]
github = "senna-lang/bonsai-coworker"
```
WHEN Git プラグインが起動する
THEN プラグインは通知ポーリングとサーベイサイクルの両方を指定された設定で開始する

#### シナリオ: トークン未設定
GIVEN 設定に `github_token: ${GITHUB_TOKEN}` があるが環境変数が未設定
WHEN Git プラグインが起動する
THEN システムは "GITHUB_TOKEN 環境変数が設定されていません" というエラーを返す
AND プラグインは起動しない

#### シナリオ: サーベイのみ有効
GIVEN notification セクションが省略されている
WHEN Git プラグインが起動する
THEN バッチサーベイ型のみ動作する
AND 通知ポーリングはスキップされる
