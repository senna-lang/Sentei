# 仕様差分: コアデーモン

対象仕様: `spec/specs/core/spec.md`

本変更では RSS プラグインの正式追加に伴い、以下を更新する:
- category enum の「RSS 用を予約」という暫定定義を、「RSS プラグイン正式追加による採用」へ格上げする

---

## MODIFIED 要件

### 要件: ラベリング category の source 別 enum

従来 `spec/archive/2026-04-19-add-core-and-git-plugin/specs/core/spec-delta.md` で定めていた:

> RSS: `"llm_research"` | `"llm_news"` | `"dev_tools"` | `"other"` (予約)

を、RSS プラグインの実装に伴い「予約」の但し書きを外す。enum 値・意味は変更しない。

#### シナリオ: RSS アイテムの有効な category
GIVEN RSS プラグインがラベリングを実行する
WHEN Bonsai がエントリに category を付与する
THEN `category` は `"llm_research"` / `"llm_news"` / `"dev_tools"` / `"other"` のいずれか
AND 他の文字列は GBNF grammar により生成不能

#### シナリオ: source 別 enum の独立性
GIVEN 同じアイテムタイトルでも source によって category enum は異なる (例: git の `"pr"` と rss の `"llm_research"`)
WHEN `sentei list --source rss --category llm_research` が実行される
THEN RSS source かつ `llm_research` category のアイテムのみが返る
AND git source のアイテムは含まれない
