# 提案: デザインテーマを盆栽 / 剪定モチーフの緑基調へ刷新

## なぜ

**背景**:
- sentei というプロダクト名は「剪定（盆栽の手入れ）」から来ており、「通知・情報の剪定」というコンセプトを体現している
- 現状の UI テーマは「Raycast/Linear 風」の汎用的なダーク紺色（`#1A1A2E` / `#16213E` / `#222244`）で、プロダクトのコンセプトと結びついていない
- urgency の色も一般的な赤/黄/灰のシグナル配色で、盆栽・自然というテーマと乖離している

**現状**: `SenteiTheme.swift` に紺色ベースの汎用ダークテーマ。urgency は赤 (`#FF4444`) / 橙 (`#FFB020`) / 灰 (`#888888` / `#555555`)。

**目指す状態**: 苔・常緑樹・墨・和紙・柿渋（古銅）をモチーフにした、緑基調で侘び寂びを感じさせる唯一無二のダークテーマ。背景は純黒ではなく緑寄りの墨色。urgent だけ暖色（柿渋）で緑の調和を一点だけ破ることで、視線誘導と美観を両立する。

## コンセプト

**苔庭 × 墨 × 古銅**。

- **墨 (sumi)**: 背景は純黒ではなく、ごく僅かに緑を差した墨色で統一
- **苔 (koke)**: 主要アクセントは苔色の低彩度グリーン。接続状態・focus・primary action に適用
- **和紙 (washi)**: テキストは純白を避け、緑と黄みを微かに含む和紙色
- **柿渋 (kakishibu)**: urgent のみ古銅（柿渋）色。緑の調和を一点で破る侘び寂びの手法
- **自然系 urgency**: shouldCheck → 苔黄、canWait → 竹鼠、ignore → 涸

## 変更内容

- `SenteiTheme.swift` のカラーパレットを全面的に入れ替え（18 色を再定義）
  - background primary/secondary/card/border
  - text primary/secondary/tertiary
  - accent: koke / jōroku / wakaba
  - urgency: urgent (kakishibu) / shouldCheck / canWait / ignore
  - divider / borderSubtle / focusRing
- 主要ビューへの適用と最終調整
  - `DashboardView` / `SidebarView` / `ItemListView` / `BoardView`
  - `PopoverView`
  - コンポーネント: `ItemRow` / `UrgencyBadge` / `CategoryIcon`
- spec 更新: `spec/specs/macos-app/spec.md` の「デザインテーマ」Requirement を緑基調へ修正

## 影響範囲

### 影響する仕様
- `spec/specs/macos-app/spec.md` - Requirement「デザインテーマ」を MODIFIED（urgency 色値の更新 + 緑基調の規定追加）

### 影響するコード
- `app/Sentei/Sources/Theme/SenteiTheme.swift` - パレット全面刷新
- `app/Sentei/Sources/Views/Dashboard/*.swift` - 背景・テキスト・アクセント参照の再確認
- `app/Sentei/Sources/Views/MenuBar/PopoverView.swift` - 同上
- `app/Sentei/Sources/Views/Components/*.swift` - UrgencyBadge / CategoryIcon / ItemRow の配色確認

### ユーザー影響
- 見た目が大きく変わる（ダーク紺 → ダーク墨緑）
- 既存ユーザーはアップデート後、urgency の色が変わることに気付く（赤 → 古銅 など）。意味（urgent が一番目立つ）は維持される

### API 変更
- なし（ビジュアルのみ）

### マイグレーション
- [ ] 設定ファイル / DB 変更なし
- [ ] ユーザーへの告知は不要（視覚的変化は起動時に自明）
- [ ] ドキュメント更新: README のスクリーンショットがあれば差し替え

## 規模見積り

Small（半日程度）。パレット置換は機械的、主要ビューはパレット参照のみで色ハードコードがないため、SenteiTheme.swift への集中修正で完結する。

## リスク

- **コントラスト不足**: 緑基調で彩度を落とすと可読性が落ちる可能性
  - 緩和: 提案パレットは事前に contrast ratio を確認済み（textPrimary on bg = 14:1, urgent on bg = 6:1）
- **urgency の意味が伝わりにくくなる**: 赤 → 柿渋（古銅）への変更で「危険度」の直感が薄れる恐れ
  - 緩和: urgent だけ唯一の暖色に割り当てることで、全体の中で一点だけ目立つ構造を維持。色以外（位置・フォント・通知）でも urgent を伝える
- **spec とコードのずれ**: spec の urgency 色値と実装がズレる
  - 緩和: 本提案で spec を同時更新する
