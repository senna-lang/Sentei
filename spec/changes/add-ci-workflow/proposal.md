# 提案: GitHub Actions による CI ワークフロー追加

## なぜ

**背景**:
- 試作段階のまま `github.com/senna-lang/Sentei` に初回 push を済ませたが、品質ゲートが無い状態。手元で `go test` を忘れると壊れたまま main に乗る余地がある
- 先日の初回 push では `.gitignore` のミス (`sentei` ルールが `cmd/sentei/` ディレクトリまで ignore) を push 後まで気付けず、実質ビルド不能な main を一時的に残した。CI があれば即座に検出できた
- 今後プラグインを追加したり contributor を受け入れる前に、baseline として最低限のテスト実行環境を整えておきたい

**現状**: CI ワークフロー無し。`go test ./...` と `swift build` は手動のみ。

**目指す状態**: main への push と PR で Go 側の build / vet / test が自動実行され、少なくとも「ビルドが通ってテストが緑」を保証する。macOS アプリの Swift build は Phase 2 で追加する。

## 方針

- 試作段階なので**最小構成から始める**。後で必要に応じて拡張する
- Ubuntu runner (無料枠潤沢) で Go 側のみ。macOS runner (Swift 用) は Phase 2
- Go のバージョンは `go.mod` の `go 1.XX` に従う (setup-go の `go-version-file` で追随)

## 変更内容

### Phase 1: Go CI (必須)
- `.github/workflows/ci.yml` を新設
- trigger: `push` to main / `pull_request` to main
- jobs: `go-test` (ubuntu-latest)
  - `actions/checkout@v4`
  - `actions/setup-go@v5` with `go-version-file: go.mod`
  - `go build ./...`
  - `go vet ./...`
  - `go test -race ./...`
- PR テンプレートは今回入れない (試作段階につき不要)

### Phase 2: Swift build (後続、必要になったら)
- 同 workflow に `swift-build` ジョブ (macos-latest) を追加
- `cd app/Sentei && swift build`
- 別 workflow に分けるか同一にするかは実装時に判断

## 影響範囲

### 影響する仕様
- なし (インフラ変更、機能仕様に影響なし)

### 影響するコード / 設定
- `.github/workflows/ci.yml` を新規作成
- 既存コードは無変更

### ユーザー影響
- なし (開発体験のみの改善)

### API 変更
- なし

### マイグレーション
- 不要

## 規模見積り

- Phase 1: Extra Small (1 時間以内)
- Phase 2: Small (半日、macOS runner 初見対応込み)

## リスク

- **macOS runner のコスト**: 無料枠で月 2000 分 (Linux 換算 10×)。Phase 2 では消費が早いので PR trigger 頻度を絞るか、PR ラベル制にする
  - 緩和: Phase 1 は Linux のみで始める。Phase 2 着手時に予算確認
- **Flaky test**: `go test -race` で race 検知を有効にすると CI だけ失敗することがある
  - 緩和: 失敗したら root cause を直す。`-race` を外すのは最終手段
- **push 時の即時フィードバック遅延**: Actions の起動は数秒〜。手元での `go test` を置き換えない
  - 緩和: CLAUDE.md に従い手元でも毎回 `go test` を走らせる運用は維持

## メモ

試作段階のため、proposal だけ残して実装は後回し。実装着手時は Phase 1 から。
