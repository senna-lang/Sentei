# Implementation Tasks

## Step 0: Go 側の変更（API 追加 + 通知廃止）

- [x] 1. `internal/storage/storage.go` に `DeleteItem(source, sourceID)` メソッド追加 + テスト
- [x] 2. `internal/server/server.go` に `DELETE /api/items/{source}/{source_id}` エンドポイント追加
- [x] 3. `cmd/sentei/serve.go` から osascript 通知の OnSubmit hook を削除
- [x] 4. `go test ./...` で全テスト pass 確認

## Step 1: プロジェクトセットアップ + Models

- [x] 5. Xcode プロジェクト作成（macOS App, SwiftUI, Swift）`app/Sentei/` (SwiftPM で実装)
- [x] 6. Info.plist 設定（`LSUIElement = true`, Deployment Target macOS 14.0）
- [x] 7. Codable モデル定義（SenteiItem, Board, Status — Go の型に対応）
- [x] 8. JSON デコードテスト（モックレスポンスからのパース確認）

## Step 2: APIClient + ViewModel + DaemonManager

- [x] 9. URLSession ベース REST クライアント実装（fetchItems, fetchBoards, fetchStatus, deleteItem）
- [x] 10. DaemonManager 実装（Process で `sentei serve` を spawn、アプリ終了時に停止）
- [x] 11. NotificationService 実装（UserNotifications で urgent アイテムを通知、bundle ガード付き）
- [x] 12. AppViewModel 実装（@Observable, 動的ポーリング: バックグラウンド 60 秒 / フォアグラウンド 15 秒 / 表示時即時 fetch、接続状態管理）
- [x] 13. APIClient テスト（モックサーバーでの正常系・エラー系）

## Step 3: メニューバー + ポップオーバー

- [x] 14. SenteiApp.swift に MenuBarExtra + Window 定義（WindowGroup → Window に変更: 単一インスタンス保証）
- [x] 15. PopoverView 実装（接続状態ヘッダー + 直近 10 件 + フッター「ダッシュボードを開く / 終了」、固定高さ 440px）
- [x] 16. ItemRow 共有コンポーネント実装（チェックボタン + urgency バッジ + category アイコン + title、行クリックで URL を開く）
- [x] 17. UrgencyBadge / CategoryIcon コンポーネント実装
- [x] 18. メニューバーアイコン 3 状態実装（通常 / urgent バッジ付き / グレー未接続）

## Step 4: ダッシュボードウィンドウ

- [x] 19. DashboardView 実装（NavigationSplitView: サイドバー + コンテンツ、× はウィンドウを隠すだけ）
- [x] 20. SidebarView 実装（アイテム urgency 別カウント / 掲示板リポジトリ個別ネスト / ステータス）
- [x] 21. ItemListView 実装（urgency / source / category フィルタ Picker + ItemRow 一覧、survey アイテムは除外）
- [x] 22. BoardView 実装（monospace プレーンテキストでリポジトリ別掲示板表示）

## Step 5: テーマ + 仕上げ

- [x] 23. SenteiTheme 定義（カスタムカラーパレット、urgency 色、背景、カード）
- [x] 24. ダークモード適用（`.preferredColorScheme(.dark)`）
- [x] 25. ウィンドウサイズ・位置の記憶（`@SceneStorage` でサイドバー選択状態 + Window 自動永続化）
- [x] 26. メニューバーアイコン作成（SF Symbol `leaf.fill`）
- [x] 27. ビルド確認 + 動作テスト（アプリ起動 → デーモン自動起動 → ポップオーバー → ダッシュボード）

## Step 6: 追加改善 (2026-04-18)

- [x] 28. アイテム一覧を自分宛通知のみに絞り込み（AppViewModel で `surveyType == nil` フィルタ）
- [x] 29. 掲示板を今日分のみに絞り込み（server `filterTodaySurvey`）
- [x] 30. Git サーベイを拡張: 既存 `merged_pr` / `new_issue` に加え `open_pr` / `release` を追加
- [x] 31. 掲示板サマリーを LLM フリー生成から決定的テンプレート生成に差し替え（`bonsai.BuildTemplateSummary`）
- [x] 32. 過去に survey されたレポジトリは今日ゼロ活動でも掲示板枠を表示

---

**Notes**:
- Step 0 は Go 側の変更。Step 1〜5 は Swift 側。Step 6 は両方にまたがる追加改善。
- ItemRow はポップオーバーとダッシュボードで共有
- アプリがデーモンのライフサイクルを管理（Ollama パターン）
- 通知はアプリの UserNotifications に一本化（Go 側 osascript 廃止）
- 掲示板はプレーンテキスト monospace 表示
- ポーリングは動的（バックグラウンド 60 秒 / フォアグラウンド 15 秒 / 表示時即時 fetch）
- アイテム一覧 = 自分宛通知 / 掲示板 = 今日のレポジトリ活動（棲み分けを明確化）
