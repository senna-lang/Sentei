# 実装タスク

## Phase 1: Go CI (必須)

1. `.github/workflows/ci.yml` を新規作成
   - `on: push:` branches main / `on: pull_request:` branches main
   - job `go-test` on `ubuntu-latest`
2. `actions/checkout@v4` でチェックアウト
3. `actions/setup-go@v5` を `go-version-file: go.mod` で設定
4. `go build ./...` を実行
5. `go vet ./...` を実行
6. `go test -race ./...` を実行
7. ローカルで `act` 等で試運転 (任意) → main に push して初回 run 成功を確認
8. バッジを README に貼る (README が未作成のためスキップ可、後続タスクで対応)

## Phase 2: Swift build (後続・必要性が出たら)

9. 同 workflow に job `swift-build` on `macos-latest` を追加
10. `cd app/Sentei && swift build` の CI 実行
11. 消費 minutes のモニタリング (無料枠 2000 min/月)、必要なら trigger を PR ラベル制に絞る

## Phase 3: 仕上げ

12. PR に CI 必須チェックを設定 (branch protection rule、GitHub 上で手動設定)
13. 落ちた時のデバッグ手順を記録 (必要になったら `spec/specs/ci/spec.md` を新設するか判断)

---

**メモ**: 本提案は試作段階のため実装保留。着手する際は Phase 1 のみでも十分に価値がある。
