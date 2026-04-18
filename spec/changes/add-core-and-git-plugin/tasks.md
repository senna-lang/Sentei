# 実装タスク

## Step 1: Bonsai Daemon セットアップ ✅ 完了

1. ✅ llama.cpp ビルド（upstream, Metal 対応）
2. ✅ llama.cpp server モードで OpenAI 互換 API の応答確認
3. LaunchAgent plist ファイル作成（Mac 起動時に Bonsai 自動起動）
4. 24 時間常駐テスト（メモリ 2GB 以下、CPU 安定を確認）

## Step 2: Bonsai POC ✅ 完了

POC で判明したこと:
- ✅ GBNF grammar 制約下の enum 分類（urgency, category）は動作する
- ✅ `/completion` エンドポイント + grammar + `n_predict` で構造化 JSON 出力可能
- ✅ M2 で 36.7 tok/s、レイテンシ約 3s
- ❌ 自由テキスト生成（掲示板要約）は品質不足（繰り返し、言語混在、think タグ漏れ）
- ❌ GitHub 通知の summary は通知自体が既に要約的で LLM の付加価値が薄い
- ✅ 短文サマリー（1-2文）は「おまけ」レベルで使える（中国語混入あり）

**設計判断**:
- Bonsai は「ラベリングエンジン」に特化（urgency + category の enum 分類）
- 掲示板の統計行・アイテム一覧はテンプレートエンジンが生成
- 掲示板のサマリー文のみ Bonsai フリー生成（品質不安定、失敗時は非表示）
- 通知系は urgency + category ラベリング、サーベイ系は category ラベリングのみ（マージ済 PR に urgency をつけるのは不適切）

**成果物**: `poc/prompts/classify.gbnf`, `poc/prompts/classify_prompt.txt`

## Step 3: コア実装 ✅ 完了

アーキテクチャ: Ollama パターン（単一バイナリ + REST API）
技術スタック: cobra + viper, modernc.org/sqlite, slog
モジュール: `github.com/senna-lang/sentei`

5. ✅ Go プロジェクト構造作成 (`go mod init github.com/senna-lang/sentei`, cmd/sentei + internal/ 構成)
6. ✅ cobra セットアップ（`sentei serve` / `sentei status` サブコマンド）
7. ✅ Plugin インターフェース定義 (`Plugin`, `Core`, `Item`, `Label`, `LabeledItem` 型)
8. ✅ SQLite スキーマ作成とストレージ層実装（modernc.org/sqlite, items テーブル + UPSERT + フィルタクエリ）
9. ✅ Bonsai ラベリングクライアント実装（`/completion` + GBNF grammar + `n_predict: 150`, 同期呼び出し, 指数バックオフリトライ）
10. ✅ プラグインごとの GBNF grammar 管理（GrammarProvider interface, Git 用 category enum）
11. ✅ Submit パイプライン実装（バリデーション → 冪等性チェック → Bonsai ラベリング → SQLite 保存）
12. ✅ REST API ハンドラ実装（`GET /api/items`, `GET /api/board`, `GET /api/board/{repo}`, `GET /api/status`）
13. ✅ 掲示板テンプレートエンジン実装（統計行 + Bonsai サマリー + category 別グループ化）
14. ✅ コア単体テスト（plugin 3件, storage 7件, bonsai 7件, board 7件 = 全 pass）

## Step 4: Git プラグイン - 通知トリガー型 ✅ 完了

15. ✅ GitHub 通知取得（`gh api /notifications` 経由）
16. ✅ 1 分間隔ポーリング + 差分検出（seenIDs による重複スキップ）
17. ✅ Item 正規化（metadata: repo, notification_type, subject_type, author）
18. ✅ 通知タイプフィルタリング
19. ✅ API URL → HTML URL 変換
20. ✅ Git プラグインテスト（7件 pass）

**動作確認**: 実通知 19 件をラベリング。urgency/category 分類が正確に動作。
release PR の分類改善（プロンプトに「title に release → category: release」ルール追加）。

## Step 5: Git プラグイン - バッチサーベイ型 ✅ 完了

21. ✅ リポジトリ状況収集（REST API: マージ済 PR + 新規 Issue）
22. ✅ サーベイ結果の各アイテムを Item として Submit → Bonsai ラベリング
23. ✅ 掲示板テンプレートエンジン改修（統計行テンプレート + Bonsai サマリー文 + category 別一覧）
24. ✅ Bonsai サマリー生成（`/completion` フリー生成、cleanSummary でゴミ除去）
25. ✅ サーベイの定期実行ループ（デフォルト 1 時間間隔）

**動作確認**: arxiv-compass のマージ済 PR 4 件をサーベイ取得・ラベリング。
掲示板出力: 統計行「マージ 3件 / あなた担当 3件」+ Bonsai サマリー + category 別一覧。
4 リポジトリの掲示板を同時生成。

## Step 6: CLI + 動作確認 (3-4 日)

CLI は REST API クライアント。`sentei serve` 未起動時は自動起動する（Ollama パターン）。

26. serve コマンド実装（HTTP サーバー + プラグイン起動） ← 実装済み、Step 6 では config 読み込み統合
27. init コマンド実装（`~/.config/sentei/` 作成 + デフォルト config.toml）
28. list コマンド実装（`GET /api/items` + --urgency, --source, --category フィルタ + 色付き出力）
29. board コマンド実装（`GET /api/board` + テンプレート出力）
30. status / stop / plugin list コマンド実装
31. macOS 通知実装（urgency == urgent で osascript 通知）
32. CLI 出力フォーマット（色付き、urgency 色分け）
33. config.toml からの設定読み込み（viper 統合、ハードコード値を config に移行）
34. LaunchAgent plist 生成（`sentei init` で作成）
35. README 作成

---

**補足**:
- Step 1-5 完了済み。残りは Step 6 のみ
- Bonsai の役割: ラベリングエンジン（urgency + category enum 分類）+ 掲示板サマリー文（おまけ）
- 掲示板は「テンプレート統計行 + Bonsai サマリー + category 別アイテム一覧」の 3 層構造
- 通知系は urgency ラベリングあり、サーベイ系は urgency なし（マージ済 PR に優先度は不要）
- アーキテクチャ: Ollama パターン（単一バイナリ `sentei`、`serve` でデーモン、他コマンドは API クライアント）
- config: TOML (`~/.config/sentei/config.toml`)
