# Spec Delta: macos-app

This file contains specification changes for `spec/specs/macos-app/spec.md`.

## MODIFIED Requirements

### Requirement: デザインテーマ
**Previous**: ダーク寄りモダンデザイン（Raycast/Linear 風）。urgency 色は赤 (`#FF4444`) / 黄 (`#FFB020`) / 灰 (`#888888` / `#555555`)。

システムは盆栽 / 剪定モチーフの緑基調ダークテーマで表示しなければならない (SHALL)。
背景は純黒ではなく緑寄りの墨色を基調とし、アクセントには苔色の低彩度グリーンを用いる。
urgency の "urgent" のみ唯一の暖色（柿渋 / 古銅）を割り当て、他のラベルは自然系（苔黄・竹鼠・涸）でまとめる。
テキストは純白を避け、和紙色（`#E8EDE5`）を最高階層とする。

#### Scenario: ダークモード（緑基調墨色）
GIVEN macOS アプリが起動する
WHEN ウィンドウとポップオーバーが描画される
THEN ダークカラースキームが適用される
AND 背景は墨色 `#141815`（primary）/ `#1A1F1C`（secondary）/ `#222823`（card）で構成される
AND テキストは和紙色 `#E8EDE5`（primary）/ `#A8B0A4`（secondary）/ `#6B756A`（tertiary）の階層で描画される

#### Scenario: 主要アクセントは苔色
GIVEN 接続状態 / focus / primary action が描画される場面
WHEN アクセント色が必要になる
THEN 苔色 `#7BA05B`（accentPrimary）が用いられる
AND hover / セクションヘッダー等の二次アクセントは常緑 `#4A6B3D` を用いる

#### Scenario: urgency の自然色マッピング
GIVEN アイテムが表示される場面（ポップオーバーまたはダッシュボード）
WHEN urgency ラベルが描画される
THEN "urgent" は柿渋 `#C97E4A`（古銅、唯一の暖色）で表示される
AND "should_check" は苔黄 `#C9B560` で表示される
AND "can_wait" は竹鼠 `#7A8A78` で表示される
AND "ignore" は涸 `#4A524A` で表示される

#### Scenario: コントラスト確保
GIVEN 上記パレットでテキストと背景が組み合わされる
WHEN 主要テキストが描画される
THEN textPrimary on backgroundPrimary のコントラスト比は 4.5:1 以上（WCAG AA 準拠）を満たす
AND urgent on backgroundPrimary のコントラスト比は 3:1 以上を満たす（UI コンポーネント基準）

#### Scenario: 純黒・純白の不使用
GIVEN テーマパレットが定義される
WHEN 背景色および主要テキスト色が指定される
THEN 純黒 `#000000` と純白 `#FFFFFF` は背景・主要テキストいずれにも用いられない
