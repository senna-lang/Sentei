# 提案: 「掲示板 / board」を「サマリー / summary」へ全面 rename

## なぜ

**背景**:
- 現状の機能は「リポジトリの今日の活動を 1 つにまとめて見せる」もので、これは掲示板 (公開・告知) より **サマリー (要約・まとめ)** の方がコンセプトを正確に表す
- 「掲示板」はユーザーに「他人の発言が貼られる場」を連想させ、機能の実態 (自分の監視リポジトリの今日の動きの集約) と乖離する
- UI 文言だけでなくコード識別子 (`Board` / `internal/board` / `/api/board`) も同時に揃えることで、新規参加者の学習コストとドメイン用語のズレを防ぐ

**現状**: UI ラベル・Go パッケージ・REST API・Swift 型・CLI サブコマンド・spec すべてが `board` / `掲示板` で統一されている。

**目指す状態**: すべての層で `summary` / `サマリー` に統一。`board` という単語をコードベース全体から排除する (アーカイブ済 proposal を除く)。

## 変更内容

### Go 側
- `internal/board/` → `internal/summary/` (パッケージ名 `board` → `summary`)
- `internal/board/board.go` → `internal/summary/summary.go` (型 `Board` → `Summary`、関数名 `Render*Board*` → `Render*Summary*` 等)
- `internal/board/board_test.go` → `internal/summary/summary_test.go`
- `cmd/sentei/board_cmd.go` → `cmd/sentei/summary_cmd.go` (CLI サブコマンド `sentei board` → `sentei summary`)
- `internal/server/server.go`: `/api/board` → `/api/summary`、ハンドラ関数名・JSON フィールド・コメントを連動
- `cmd/sentei/main.go`: cobra コマンド登録の差し替え

### Swift 側
- `Sources/Models/Board.swift` → `Sources/Models/Summary.swift` (型 `Board` → `Summary`)
- `Sources/Views/Dashboard/BoardView.swift` → `Sources/Views/Dashboard/SummaryView.swift`
- `Sources/Views/Dashboard/DashboardSelection.swift`: `case board(repo:)` → `case summary(repo:)`
- `Sources/Services/APIClient.swift`: `fetchBoards()` → `fetchSummaries()`、エンドポイント `/api/board` → `/api/summary`
- `Sources/ViewModels/AppViewModel.swift`: `boards` プロパティ・`refreshBoards()` を `summaries` / `refreshSummaries()` へ
- `Sources/Views/MenuBar/PopoverView.swift` / `Sources/Views/Dashboard/SidebarView.swift` / `Sources/Views/Dashboard/DashboardView.swift`: 参照とラベルを差し替え
- `Tests/ModelTests.swift`: `Board` 関連テストを `Summary` へ

### UI 文言
- サイドバーのセクション見出し「掲示板」→「サマリー」
- BoardView 空状態「{repo} の掲示板はまだ生成されていません」→「{repo} のサマリーはまだ生成されていません」
- CLI ヘルプテキストも更新

### spec
- `spec/specs/core/spec.md`: 概要文・Requirement「掲示板サマリー（テンプレート生成）」を rename
- `spec/specs/macos-app/spec.md`: 「ダッシュボードウィンドウ」「アイテム一覧と掲示板の棲み分け」内の `掲示板` / `/api/board` を更新

### REST API (破壊的変更)
- `GET /api/board` → `GET /api/summary` (旧エンドポイントは削除、後方互換は持たない)
- レスポンス JSON のフィールド名 `board` → `summary`

## 影響範囲

### 影響する仕様
- `spec/specs/core/spec.md` - Requirement「掲示板サマリー（テンプレート生成）」を MODIFIED
- `spec/specs/macos-app/spec.md` - Requirement「ダッシュボードウィンドウ」「アイテム一覧と掲示板の棲み分け」を MODIFIED

### 影響するコード
```
Go:
- internal/board/  → internal/summary/  (パッケージ rename)
- cmd/sentei/board_cmd.go → cmd/sentei/summary_cmd.go
- internal/server/server.go (ルート + ハンドラ)
- cmd/sentei/main.go (cobra 登録)

Swift:
- app/Sentei/Sources/Models/Board.swift → Summary.swift
- app/Sentei/Sources/Views/Dashboard/BoardView.swift → SummaryView.swift
- app/Sentei/Sources/Views/Dashboard/DashboardSelection.swift
- app/Sentei/Sources/Services/APIClient.swift
- app/Sentei/Sources/ViewModels/AppViewModel.swift
- app/Sentei/Sources/Views/MenuBar/PopoverView.swift
- app/Sentei/Sources/Views/Dashboard/SidebarView.swift
- app/Sentei/Sources/Views/Dashboard/DashboardView.swift
- app/Sentei/Tests/ModelTests.swift
```

### ユーザー影響
- CLI: `sentei board` を叩いていたスクリプトは `sentei summary` への書き換えが必要
- API: `GET /api/board` を直接叩いていた外部クライアントはエンドポイント差し替えが必要 (ただし内部利用のみ想定)
- UI: ラベルが「掲示板」→「サマリー」に変わる

### API 変更
- 破壊的: `GET /api/board` → `GET /api/summary` (旧エンドポイント削除)
- 破壊的: レスポンス JSON フィールド `board` → `summary`

### マイグレーション
- [ ] DB 変更なし (ストレージスキーマには board / summary を含まない)
- [ ] config 変更なし
- [ ] 既存ユーザーは macOS アプリのアップデートで自動的に新 API を叩く (アプリと daemon が同一バイナリ系列のため版ずれリスクは低い)
- [ ] アーカイブ済 proposal (`spec/archive/2026-04-18-add-macos-app/`) は履歴として touchしない

## 規模見積り

Medium (1 日程度)。ファイル数は多いが機械的な rename。Go パッケージ rename は import path 連動で範囲は明確。

## リスク

- **既存 daemon との版ずれ**: 古い daemon (board API 提供) に新アプリ (summary API 期待) が接続するとデータが取れない
  - 緩和: macOS アプリはバイナリに sentei daemon を同梱して spawn するため、版ずれは起きにくい構造。それでも README にバージョン揃え推奨を明記
- **アーカイブ済 proposal 内の言及との不一致**: `spec/archive/2026-04-18-add-macos-app/` には「掲示板」が残る
  - 緩和: アーカイブは履歴扱いなので意図的に触らない。新 spec が真実
- **未アーカイブ change `redesign-theme-bonsai` との順序**: 先にこちらをアーカイブするか後にするかで spec の最終形が変わる
  - 緩和: 先に `redesign-theme-bonsai` をアーカイブしてから本提案に着手するのが安全
- **POC スクリプト (`poc/check_board.sh` 等)**: 履歴として残るが現在のフローでは使用されない
  - 緩和: 触らない (履歴扱い)。本提案のスコープ外
