# 仕様差分: コアデーモン

対象仕様: `spec/specs/core/spec.md`

## 設計方針: Bonsai はラベリングエンジン

Bonsai の役割は全プラグイン共通の「ラベリングエンジン」に特化する。
各アイテムに `{urgency, category}` をラベル付けし、GBNF grammar で出力スキーマを保証する。
自由テキスト生成（要約、掲示板テキスト等）は Bonsai に頼らず、テンプレートエンジンが担う。
summary フィールドは「おまけ」として出力するが、品質が低い場合は title で代替する。

## 追加要件

### 要件: プラグインインターフェース
システムは `Name() string`, `Start(ctx, core) error`, `Stop() error` メソッドを持つ Plugin インターフェースを定義しなければならない (SHALL)。

#### シナリオ: プラグイン登録
GIVEN 有効なプラグイン実装がある
WHEN コアデーモンが起動する
THEN システムはプラグインを名前で登録する
AND 有効な context と core 参照で `Start()` を呼び出す

#### シナリオ: シャットダウン時のプラグイン停止
GIVEN プラグインが実行中である
WHEN コアデーモンがシャットダウンシグナルを受信する
THEN システムは各登録プラグインの `Stop()` を呼び出す
AND グレースフルに完了を待つ

---

### 要件: Item 送信
WHEN プラグインが `Core.Submit(item)` で Item を送信した場合、
システムは Item のフィールドを検証し、Bonsai 判定キューに追加しなければならない (SHALL)。

#### シナリオ: 正常な Item 送信
GIVEN プラグインが source "git"、source_id "notif-123"、title "Review request"、content 本文を持つ Item を送信する
WHEN Item がバリデーションを通過する
THEN システムは Item を Bonsai 判定キューに追加する
AND エラーを返さない

#### シナリオ: 重複 Item の処理
GIVEN source "git"、source_id "notif-123" の Item が既にデータベースに存在する
WHEN プラグインが同じ source と source_id の Item を送信する
THEN システムは Bonsai 判定をスキップする
AND エラーを返さない（冪等性）

#### シナリオ: 不正な Item の拒否
GIVEN title が空の Item がある
WHEN プラグインがその Item を送信する
THEN システムはバリデーションエラーを返す
AND Item をキューに追加しない

---

### 要件: Bonsai ラベリング（GBNF grammar による構造化出力）
WHEN Item が判定キューに入った場合、
システムは GBNF grammar 付きリクエストを Bonsai（localhost:8080 の `/completion` エンドポイント）に送信し、
構造化された JSON レスポンス（urgency, category, summary）を取得しなければならない (SHALL)。

urgency と category は GBNF grammar で enum 値に制約する。
summary は自由テキストだが品質が不安定なため、表示時は title をプライマリ、summary をセカンダリとして扱う。

#### GBNF grammar 定義

category の enum はプラグインごとに異なる。Git プラグインの場合:

```gbnf
root     ::= "{" ws "\"urgency\":" ws urgency "," ws "\"category\":" ws category "," ws "\"summary\":" ws summary "}" ws
urgency  ::= "\"urgent\"" | "\"should_check\"" | "\"can_wait\"" | "\"ignore\""
category ::= "\"pr\"" | "\"issue\"" | "\"ci\"" | "\"release\"" | "\"discussion\"" | "\"other\""
summary  ::= "\"" char char char char char+ "\""
char     ::= [^"\\\n]
ws       ::= [ \t\n]*
```

将来のプラグインでは category の enum を差し替える:
- RSS: `"llm_research"` | `"llm_news"` | `"dev_tools"` | `"other"`
- arxiv: `"llm"` | `"ml"` | `"systems"` | `"other"`

#### ラベリング精度を上げる実装方針

- **few-shot examples**: プロンプトに 2-3 個の分類例を含める
- **temperature 0.2**: 分類タスクなので低温に設定
- **metadata ルール明記**: 「mentor からの review_requested → urgent」等をプロンプトに記載
- **`/no_think` プレフィックス**: Qwen3 ベースの thinking モードを無効化
- **`/completion` エンドポイント使用**: `/v1/chat/completions` ではなく `/completion` で grammar を確実に適用
- **`n_predict: 150`**: 出力トークン数を制限して暴走を防止
- **フォールバック**: パース失敗時は urgency を "should_check"、category を "other" にデフォルト

#### シナリオ: ラベリング成功
GIVEN Bonsai が localhost:8080 で稼働中
AND title "メンターからの PR レビュー依頼"、metadata `{"notification_type": "review_requested"}` の Item がある
WHEN システムが GBNF grammar 付きでラベリングリクエストを送信する
THEN Bonsai が `{"urgency": "urgent", "category": "pr", "summary": "..."}` を返す
AND grammar 制約により urgency は必ず 4 値のいずれか、category は必ず 6 値のいずれかである
AND システムはラベリング結果を Item に付与する

#### シナリオ: Bonsai 停止中
GIVEN Bonsai が localhost:8080 で稼働していない
WHEN システムがラベリングを試行する
THEN システムは指数バックオフで最大 3 回リトライする
AND すべてリトライが失敗した場合、Item を "pending_label" としてマークする
AND 警告をログに記録する

#### シナリオ: summary 品質が低い場合
GIVEN Bonsai が JSON を返したが summary が不自然または title の繰り返しである
WHEN 表示層がアイテムを描画する
THEN title をプライマリテキストとして表示する
AND summary は補助情報として小さく表示する（または非表示）

#### シナリオ: Bonsai 接続タイムアウト
GIVEN Bonsai への HTTP リクエストが 30 秒以内に応答しない
WHEN システムがタイムアウトを検知する
THEN システムはリトライキューに Item を戻す
AND 警告をログに記録する

---

### 要件: SQLite 永続化
WHEN ラベリング結果が得られた場合、
システムは Item とラベリング結果を UNIQUE(source, source_id) 制約付きで SQLite に永続化しなければならない (SHALL)。

#### シナリオ: 正常な永続化
GIVEN urgency "should_check" のラベリング済み Item がある
WHEN システムが SQLite に保存する
THEN urgency, category, summary, labeled_at を含む全フィールドが保存される
AND source、urgency、category でのフィルタ取得が可能である

#### シナリオ: データベース初期化
WHEN コアデーモンが初回起動する場合、
システムは `~/.sentei/db.sqlite` に SQLite データベースファイルを作成しなければならない (SHALL)
AND スキーマ（items テーブルとインデックス）を適用する。

---

### 要件: 緊急通知
WHEN Item が urgency "urgent" とラベリングされた場合、
システムは Item のタイトルとソースを含む macOS 通知を発行しなければならない (SHALL)。

#### シナリオ: macOS 通知の送信
GIVEN urgency "urgent"、title "メンターからのレビュー依頼" のラベリング済み Item がある
WHEN ラベリング結果が永続化される
THEN タイトル "sentei"、本文 "[pr] メンターからのレビュー依頼" の macOS 通知が表示される

#### シナリオ: 非緊急アイテムは通知なし
GIVEN urgency "should_check" のラベリング済み Item がある
WHEN ラベリング結果が永続化される
THEN macOS 通知はトリガーされない

---

### 要件: デーモンライフサイクル
システムは `sentei start` で起動し `sentei stop` で停止するバックグラウンドデーモンとして動作しなければならない (SHALL)。

#### シナリオ: デーモン起動
GIVEN デーモンが起動していない
WHEN ユーザーが `sentei start` を実行する
THEN デーモンプロセスがバックグラウンドで起動する
AND PID を `~/.sentei/daemon.pid` に書き込む
AND プラグインのポーリングサイクルを開始する

#### シナリオ: デーモン停止
GIVEN デーモンが起動中である
WHEN ユーザーが `sentei stop` を実行する
THEN デーモンは全プラグインをグレースフルにシャットダウンする
AND PID ファイルを削除する

#### シナリオ: LaunchAgent 自動起動
GIVEN LaunchAgent plist がインストールされている
WHEN macOS が起動する
THEN Bonsai llama.cpp サーバーとコアデーモンが自動的に起動する
