# Spec Delta: macos-app

This file contains specification changes for `spec/specs/macos-app/spec.md`.

## MODIFIED Requirements

### Requirement: ダッシュボードウィンドウ
**Previous**: サイドバー付きのメインウィンドウでアイテム一覧と「掲示板」を表示する。

WHEN ユーザーがダッシュボードウィンドウを開いた場合、
システムはサイドバー付きのメインウィンドウでアイテム一覧とサマリーを表示しなければならない (SHALL)。
ウィンドウの閉じるボタン（×）はウィンドウを隠すだけで、アプリは終了しない。

#### Scenario: サイドバーナビゲーション
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーがサイドバーを操作する
THEN 以下のセクションが選択可能である:
- 「アイテム」セクション（urgency 別のカウントバッジ付き: urgent / should_check / 全件）
- 「サマリー」セクション（リポジトリ名が個別にネスト表示）
- 「ステータス」セクション

#### Scenario: アイテム一覧の表示
GIVEN ダッシュボードで「アイテム」が選択されている
AND 20 件のアイテムがある
WHEN アイテム一覧が表示される
THEN urgency 順でアイテムがカード形式で表示される
AND 各カードにチェックボタン、urgency 色帯、category アイコン、title、author、時刻が含まれる
AND ポップオーバーと同じ ItemRow コンポーネントが使用される

#### Scenario: アイテム一覧のフィルタリング
GIVEN ダッシュボードで「アイテム」が選択されている
WHEN ユーザーが urgency フィルタで "urgent" を選択する
THEN urgency が "urgent" のアイテムのみが表示される

#### Scenario: アイテムのチェックボタン
GIVEN ダッシュボードのアイテム一覧にアイテムが表示されている
WHEN ユーザーがチェックボタンをクリックする
THEN DELETE /api/items/{source}/{source_id} が呼ばれる
AND アイテムがリストから消える

#### Scenario: サマリーの表示
GIVEN ダッシュボードで特定リポジトリのサマリーが選択されている
WHEN サマリーが表示される
THEN サマリーテキストが monospace プレーンテキストで表示される

#### Scenario: ウィンドウを閉じる
GIVEN ダッシュボードウィンドウが表示されている
WHEN ユーザーが閉じるボタン（×）をクリックする
THEN ウィンドウが非表示になる
AND アプリはメニューバーに常駐し続ける

---

### Requirement: アイテム一覧とサマリーの棲み分け
**Previous**: 「アイテム一覧と掲示板の棲み分け」。`掲示板` 呼称、エンドポイントは `GET /api/board`。

WHEN macOS アプリがデータを表示する場合、
システムは「アイテム一覧」と「サマリー」で異なる責務のデータを表示しなければならない (SHALL)。

- **アイテム一覧** = 自分宛の GitHub 通知のみ（`survey_type` メタデータを持たないアイテム）
- **サマリー** = 監視対象リポジトリの**今日**の活動（`survey_type` を持ち、`timestamp >= startOfToday` のアイテム）

#### Scenario: アイテム一覧は自分宛のみ
GIVEN ストレージに通知由来 (survey_type なし) とサーベイ由来 (survey_type: merged_pr など) が混在する
WHEN アイテム一覧がレンダリングされる
THEN 通知由来のアイテムだけが表示される
AND urgentCount / recentItems / サイドバーの urgency 別カウントも通知のみを対象とする

#### Scenario: サマリーは今日の活動のみ
GIVEN サーベイ由来アイテムが過去 30 日分ストレージに存在する
WHEN `GET /api/summary` が呼ばれる
THEN `timestamp >= 今日の 00:00` のサーベイアイテムだけが各リポジトリのサマリーに含まれる
AND 過去に一度でもサーベイされたリポジトリは、今日の活動がゼロでもサマリー枠が返る（本文は「特に動きはありません」）

#### Scenario: 通知はサマリーに混ざらない
GIVEN `ncdcdev/foo` から通知だけ来ており、survey 対象には入っていない
WHEN サマリー API が呼ばれる
THEN `ncdcdev/foo` のサマリーは返らない（survey_type のあるアイテムが存在しないため）

---

## Notes

- Requirement「ダッシュボードウィンドウの単一インスタンス」は board / 掲示板 を含まないため変更なし
- 本 delta では Requirement 名「アイテム一覧と掲示板の棲み分け」自体を「アイテム一覧とサマリーの棲み分け」に rename している (要件名変更を含む MODIFIED)
