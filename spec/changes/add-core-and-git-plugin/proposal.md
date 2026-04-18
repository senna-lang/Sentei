# 提案: Phase 1 - コアデーモン + Git プラグイン

## コンセプト

**常駐型の優先度ラベリングツール**。
大量の通知、情報収集疲れの日々を、Bonsai（1-bit LLM）による自動ラベリングで「今何を見るべきか」を明確にする。

## なぜ

**背景**:
- エンジニアは複数の GitHub リポジトリからの通知が多すぎて、本当に対応すべきものを見失う
- RSS、arxiv、Slack など情報源が多すぎて追いきれない
- Bonsai-8B (1-bit LLM, 1.15GB) をローカル常駐させれば、プライバシーを保ちつつ即時にラベリング可能
- プラグイン型にすることで、情報源を段階的に追加できる

**POC で判明したこと**:
- Bonsai は GBNF grammar 制約下での enum 分類（urgency, category）は動作する
- 自由テキスト生成（掲示板要約等）は品質不足（繰り返し、言語混在、think タグ漏れ）
- GitHub 通知自体が既に要約的で、通知単体の LLM 要約は価値が薄い
- **Bonsai の役割は「ラベリングエンジン」に特化すべき**

**Bonsai の設計方針**:
- Bonsai は全プラグイン共通の「ラベリングエンジン」として機能する
- 各アイテムに `{urgency, category}` をラベル付け（GBNF で制約）
- summary は「おまけ」（使えたら表示、使えなくても title で代替）
- 表示（掲示板・リスト）はテンプレートエンジンが構造化データから生成

**現状**: POC 完了（Step 1-2）。Bonsai Daemon 稼働中、判定能力検証済み。

**目指す状態**: Bonsai がバックグラウンド常駐し、Git プラグインが 2 つのモードで動作する最小動作状態。
- **通知トリガー型（リアルタイム）**: GitHub 通知を検知 → Bonsai で urgency + category ラベル付け → CLI / macOS 通知
- **バッチサーベイ型（定期）**: リポジトリの活動を収集 → 各アイテムに Bonsai でラベル付け → テンプレートで掲示板生成

## アーキテクチャ: Ollama パターン

単一バイナリ (`sentei`) が CLI とデーモンの両方を兼ねる。
`sentei serve` でデーモン起動、他のサブコマンドは REST API クライアントとして動作。

```
sentei serve   → HTTP サーバー (localhost:7890) + プラグイン実行
sentei list    → GET /api/items（API クライアント）
sentei board   → GET /api/board（API クライアント）
```

将来の Web UI やメニューバーアプリも同じ REST API を叩くだけで接続可能。

## 技術スタック

| 項目 | 選択 |
|---|---|
| 言語 | Go (`github.com/senna-lang/sentei`) |
| CLI フレームワーク | cobra + viper |
| SQLite | modernc.org/sqlite (pure Go, CGO 不要) |
| API | REST (JSON), localhost:7890 |
| config | TOML, `~/.config/sentei/config.toml` |
| デーモン化 | LaunchAgent (macOS) に委譲 |
| プラグイン | コンパイル時組み込み |
| Bonsai 呼び出し | 同期（Submit → ラベリング完了待ち → 保存） |
| ログ | slog (Go 標準) |

## 変更内容

- Go プロジェクト構造の新規作成（cmd/sentei + internal/ 構成）
- コアデーモン実装（HTTP サーバー、プラグインマネージャー、Bonsai ラベリングクライアント、SQLite ストレージ）
- プラグインインターフェース定義（コンパイル時組み込み）
- Git プラグイン実装
  - 通知トリガー型: GitHub 通知取得 → Bonsai ラベル付け（urgency + category）
  - バッチサーベイ型: リポジトリ活動収集 → 各アイテムに Bonsai ラベル付け → テンプレートで掲示板
- CLI コマンド群（serve, list, board, status, init, stop, plugin list）
- LaunchAgent による自動起動設定
- macOS 通知（urgent アイテム）

## 影響範囲

### コードへの影響
```
sentei/
├── cmd/sentei/main.go          # エントリポイント + cobra 設定
├── internal/
│   ├── server/                  # HTTP サーバー + API ハンドラ
│   ├── core/                    # Submit パイプライン、プラグインマネージャー
│   ├── bonsai/                  # Bonsai ラベリングクライアント
│   ├── storage/                 # SQLite ストレージ
│   ├── plugin/                  # Plugin インターフェース定義
│   ├── board/                   # 掲示板テンプレートエンジン
│   └── notify/                  # macOS 通知
├── plugins/
│   └── git/                     # Git プラグイン実装
├── poc/                         # POC スクリプト（完了済み）
├── spec/                        # OpenSpec 仕様
└── docs/                        # ドキュメント
```

### ユーザーへの影響
- 新規ツール。brew install または go install で導入
- GitHub Personal Access Token の設定が必要
- `list` で通知ベースのアイテム確認、`board` でリポジトリ別状況確認

### マイグレーション
- [x] ドキュメント更新（README 作成）
- 新規プロジェクトのため、DB マイグレーション・API バージョン管理は不要

## 工数見積もり

3-4 週間（4 ステップ。Step 1-2 は POC で完了済み）

## リスク

- **Bonsai の判定品質** (可能性: 中 / 影響: 中): POC で urgency 分類は動作確認済み。category（Git 固有: pr/issue/ci/release 等）は追加検証が必要
- **GitHub API レートリミット** (可能性: 中 / 影響: 中): 通知 1 分ポーリング + サーベイ 1 時間で REST/GraphQL 併用。POC で GraphQL が REST の約半分の時間と確認
- **summary の品質** (可能性: 高 / 影響: 低): POC で自由テキスト生成は弱いと判明。title をフォールバックとして使い、summary は「おまけ」扱い
