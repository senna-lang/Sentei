# Spec Delta: core

This file contains specification changes for `spec/specs/core/spec.md`.

## MODIFIED Requirements

### Requirement: サマリー（テンプレート生成）
**Previous**: 「掲示板サマリー（テンプレート生成）」。`掲示板` という呼称で、レンダリング対象を表現していた。

WHEN サマリーがレンダリングされる場合、
システムは `survey_type` / カテゴリ / urgency / author から決定的テンプレートでサマリー文を生成しなければならない (SHALL)。
LLM フリー生成は品質不安定なため使用しない。

#### Scenario: テンプレートサマリーの生成
GIVEN 今日の活動が「マージ 2 件 / 新規 Issue 1 件 / @alice と @bob が関与」である
WHEN サマリーが生成される
THEN 例: `PR が 2 件マージされました、Issue が 1 件起票。 関与: @alice, @bob。`
AND LLM 呼び出しは発生しない

#### Scenario: 活動ゼロ時
GIVEN 今日の活動がゼロ
WHEN サマリーが生成される
THEN サマリーは空文字列を返し、サマリー出力にはサマリー行が出力されない

---

## Notes

- core spec の冒頭概要文（「掲示板レンダリングを含む」）も「サマリーレンダリングを含む」に置き換える（Requirement 外の本文編集）
