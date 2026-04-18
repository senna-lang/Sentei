# Proposal: macOS ネイティブアプリ (SwiftUI)

## Why

**Context**:
- sentei の REST API (localhost:7890) は Phase 1 で完成済み（アイテム一覧、掲示板、ステータス）
- 現在の UI は CLI のみ。ターミナルを開いて能動的にコマンドを打つ必要がある
- urgent 通知は osascript で飛ぶが、それ以外の情報は CLI でしか確認できない
- RSS / arxiv プラグイン追加で情報量が増える予定。CLI だけでは一覧性が不足する

**Current state**: `sentei list` / `sentei board` のターミナル出力 + osascript による urgent 通知

**Desired state**: macOS メニューバーに常駐し、ワンクリックで状況を把握できるネイティブアプリ。
メニューバーポップオーバーで直近の通知を確認、ダッシュボードウィンドウで掲示板含む全体を俯瞰。
通知はアプリの UserNotifications に一本化し、Go 側の osascript 通知は廃止する。

## What Changes

- SwiftUI macOS アプリ新規作成（`app/` ディレクトリ）
- メニューバー常駐（`MenuBarExtra`）+ ポップオーバー（直近 10 件、全 urgency）
- ダッシュボードウィンドウ（`NavigationSplitView` サイドバー + アイテム一覧 + 掲示板）
- REST API クライアント（URLSession, 動的ポーリング）
- ダーク寄りモダンデザイン（Raycast/Linear 風、カスタムカラー）
- アプリがデーモン（`sentei serve`）を自動起動・管理（Ollama パターン）
- アプリの UserNotifications で urgent 通知（Go 側の osascript 通知は廃止）
- Go 側に `DELETE /api/items/{source}/{source_id}` エンドポイント追加

## Impact

### Affected Specifications
- 新規仕様追加: macOS アプリ
- 既存仕様変更: コアデーモンの通知方式（osascript → アプリに委譲）

### Affected Code
- `app/` — 新規 Xcode プロジェクト（SwiftUI）
- `internal/server/server.go` — `DELETE /api/items/{source}/{source_id}` 追加
- `internal/storage/storage.go` — `DeleteItem(source, sourceID)` 追加
- `cmd/sentei/serve.go` — osascript 通知の OnSubmit hook 削除

### User Impact
- メニューバーから即座に通知・掲示板を確認可能に
- ターミナル不要で sentei の情報にアクセスできる
- アイテムのチェック（対応済み）で削除可能

### API Changes
- 追加: `DELETE /api/items/{source}/{source_id}` — アイテムの物理削除

### Migration Required
- [ ] Go 側の osascript 通知コード削除
- [ ] Documentation updates（README にアプリのスクリーンショット追加）

## アーキテクチャ

```
┌──────────────────┐  spawn  ┌──────────────┐
│  Sentei.app      │ ──────→ │ sentei serve │
│  (SwiftUI)       │         │ :7890        │
│                  │  HTTP   │              │
│  MenuBarExtra    │ ←─────→ │  Core Engine │
│  + WindowGroup   │  JSON   │  + Plugins   │
│  + Notifications │         │              │
└──────────────────┘         └──────────────┘
     ↑ 終了時に停止
```

- アプリが全体のエントリポイント（LaunchAgent はアプリのみ起動）
- アプリ起動時にデーモンを spawn、アプリ終了時にデーモンも停止
- アプリはドメインロジックを持たない。表示層 + 通知のみ

## 技術スタック

| 項目 | 選択 |
|------|------|
| フレームワーク | SwiftUI (macOS 14.0+) |
| メニューバー | `MenuBarExtra` (macOS 13+) |
| 状態管理 | `@Observable` (macOS 14+) |
| HTTP | `URLSession` (標準ライブラリ) |
| 通知 | `UserNotifications` (標準フレームワーク) |
| プロセス管理 | `Process` (Foundation) |
| デザイン | ダーク寄りモダン、カスタムカラーパレット |
| 配置 | `app/` ディレクトリ |

## プロジェクト構成

```
app/
├── Sentei.xcodeproj/
├── Sentei/
│   ├── SenteiApp.swift              # @main: MenuBarExtra + WindowGroup
│   ├── Models/
│   │   ├── SenteiItem.swift         # Codable: LabeledItem
│   │   ├── Board.swift              # Codable: Board レスポンス
│   │   └── Status.swift             # Codable: Status レスポンス
│   ├── Services/
│   │   ├── APIClient.swift          # URLSession REST クライアント
│   │   ├── DaemonManager.swift      # sentei serve の起動・停止管理
│   │   └── NotificationService.swift # UserNotifications による urgent 通知
│   ├── ViewModels/
│   │   └── AppViewModel.swift       # @Observable: 状態管理 + ポーリング
│   ├── Views/
│   │   ├── MenuBar/
│   │   │   └── PopoverView.swift    # ポップオーバー（直近 10 件）
│   │   ├── Dashboard/
│   │   │   ├── DashboardView.swift  # メインウィンドウ
│   │   │   ├── SidebarView.swift    # サイドバー（リポジトリ個別ネスト）
│   │   │   ├── ItemListView.swift   # アイテム一覧
│   │   │   └── BoardView.swift      # 掲示板表示（monospace プレーンテキスト）
│   │   └── Components/
│   │       ├── ItemRow.swift        # カード風アイテム行（共有コンポーネント）
│   │       ├── UrgencyBadge.swift   # urgency バッジ
│   │       └── CategoryIcon.swift   # category アイコン
│   └── Theme/
│       └── SenteiTheme.swift        # カラー・フォント定義
└── SenteiTests/
    ├── APIClientTests.swift
    └── ModelTests.swift
```

## 設計判断（Grill 結果）

| # | 判断 | 決定 |
|---|------|------|
| 1 | アプリ終了 | ポップオーバーフッターに「ダッシュボードを開く」+「終了」。× はウィンドウを隠すだけ |
| 2 | デーモン自動起動 | アプリ起動時に `sentei serve` を spawn（Ollama パターン） |
| 3 | 通知の一本化 | Go 側 osascript 廃止 → アプリの UserNotifications に統一 |
| 4 | メニューバーアイコン | 3状態: 通常 / urgent バッジ付き / グレー（未接続） |
| 5 | ポーリング | 動的: バックグラウンド 60 秒、フォアグラウンド 15 秒、表示時即時 fetch |
| 6 | 掲示板レンダリング | プレーンテキスト monospace 表示 |
| 7 | 掲示板ナビ | サイドバーにリポジトリ個別ネスト |
| 8 | アイテム操作 | チェックボタンで物理削除（DELETE API）、行クリックで URL をブラウザで開く |
| 9 | ポップオーバーレイアウト | 接続状態ヘッダー + 10 件リスト + フッター（ダッシュボード / 終了） |
| 10 | 起動順序 | LaunchAgent → アプリ → デーモン spawn。アプリ終了でデーモンも停止 |
| 11 | ItemRow 共有 | ポップオーバーとダッシュボードで同一コンポーネント |

## デザイン方針

- **テーマ**: ダーク寄りモダン（Raycast/Linear 風）
- **urgency カラー**: urgent=#FF4444, should_check=#FFB020, can_wait=#888, ignore=#555
- **背景**: ダークグラデーション
- **カード**: 角丸、微妙なボーダー
- **Dock**: 非表示（`LSUIElement = true`）— メニューバーのみ
- **メニューバーアイコン**: 3状態（通常 / urgent バッジ / グレー）

## ポップオーバーレイアウト

```
┌─────────────────────────────────────┐
│ ● 接続中  |  urgent: 2             │
├─────────────────────────────────────┤
│ [✓] 🔴 🔀 PR レビュー依頼 (@mentor)│
│ [✓] 🟡 📝 SPECTER2対応 (#46)       │
│ [✓] ⚪ 🔀 マージ済: 丸め誤差修正    │
│ ...                                 │
├─────────────────────────────────────┤
│ [ダッシュボードを開く]  [終了]       │
└─────────────────────────────────────┘
```

- 左端: チェックボタン（✓ で DELETE → アイテム消える）
- urgency バッジ（色丸）+ category アイコン + title + author
- 行クリック → ブラウザで URL を開く

## Timeline Estimate

5〜7 日（Step 0〜5）

## Risks

- **Swift 未経験** (可能性: 高 / 影響: 中): SwiftUI は宣言的で React に近い。表示層のみなので学習コストは限定的。Claude が Swift コードを生成可能
- **MenuBarExtra のポップオーバーサイズ制限** (可能性: 中 / 影響: 低): macOS のポップオーバーは高さ制限あり。10 件に絞ることで対応済み
- **API レスポンス形式の不一致** (可能性: 低 / 影響: 中): Go の JSON と Swift Codable の型マッピング。テストで早期検証
- **プロセス管理** (可能性: 中 / 影響: 中): アプリからの `sentei serve` spawn + 終了時の停止。Process のライフサイクル管理が必要
