# 実装タスク

## Phase 0: 前準備

1. 先行 change `redesign-theme-bonsai` をアーカイブ (`openspec archive redesign-theme-bonsai`) して spec を最新状態にする

## Phase 1: Go (バックエンド)

2. `internal/board/` → `internal/summary/` ディレクトリ rename + パッケージ宣言を `package summary` へ
3. `internal/summary/board.go` → `internal/summary/summary.go` (型 `Board` → `Summary`、関数 `RenderBoard*` → `RenderSummary*`)
4. `internal/summary/board_test.go` → `internal/summary/summary_test.go` (テストケース・assertion を新名へ)
5. `cmd/sentei/board_cmd.go` → `cmd/sentei/summary_cmd.go` (cobra コマンド `Use: "board"` → `Use: "summary"`、ヘルプ文言を更新)
6. `cmd/sentei/main.go` の cobra 登録を新コマンドに差し替え
7. `internal/server/server.go`: ルート `/api/board` → `/api/summary`、ハンドラ関数名・レスポンス JSON フィールドを更新
8. `go build ./...` と `go test ./...` を通す

## Phase 2: Swift (フロントエンド)

9. `Sources/Models/Board.swift` → `Sources/Models/Summary.swift` (型 `Board` → `Summary`、`Codable` キーを `summary` ベースに)
10. `Sources/Views/Dashboard/BoardView.swift` → `Sources/Views/Dashboard/SummaryView.swift` (View 名・参照変数名・空状態文言)
11. `Sources/Views/Dashboard/DashboardSelection.swift`: `case board(repo:)` → `case summary(repo:)`
12. `Sources/Services/APIClient.swift`: `fetchBoards()` → `fetchSummaries()`、URL `/api/board` → `/api/summary`
13. `Sources/ViewModels/AppViewModel.swift`: `boards` → `summaries`、`refreshBoards()` → `refreshSummaries()`、`repos` 算出ロジックの参照名差し替え
14. `Sources/Views/MenuBar/PopoverView.swift` / `SidebarView.swift` / `DashboardView.swift`: 参照とサイドバーセクション見出し「掲示板」→「サマリー」
15. `Tests/ModelTests.swift`: `Board` 関連テストを `Summary` に追従

## Phase 3: ビルド + 検証

16. `swift build` を通す (Sentei パッケージ)
17. `go test ./... && swift test --package-path app/Sentei` (両言語のテスト)
18. `swift run` で起動 → サイドバーに「サマリー」、各 repo 選択で SummaryView が表示されることを確認
19. CLI: `sentei summary` で従来 `sentei board` と同等の出力が出ることを確認

## Phase 4: 仕様 / ドキュメント

20. `spec/specs/core/spec.md` の Requirement「掲示板サマリー（テンプレート生成）」を rename し、本文中の `掲示板` → `サマリー`
21. `spec/specs/macos-app/spec.md` の関連 Requirement (「ダッシュボードウィンドウ」「アイテム一覧と掲示板の棲み分け」) を更新、`/api/board` → `/api/summary`
22. README にコマンド例があれば `sentei board` → `sentei summary` に書き換え (任意)

---

**メモ**:
- アーカイブ済 `spec/archive/2026-04-18-add-macos-app/` と `poc/` は履歴として触らない
- Go パッケージ rename はファイル移動 + 全 import 書き換えで通る (cgo なし、go モジュール内のみ)
- Swift の `Codable` キーが JSON フィールド名に直結するため、`/api/summary` のレスポンス形と Swift の `CodingKeys` が一致しているか必ず確認
