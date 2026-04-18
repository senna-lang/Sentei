# Git プラグイン仕様

sentei の Git プラグインの振る舞いに関する仕様。
GitHub 通知の監視と、監視対象リポジトリの定期サーベイを扱う。

## Requirements

### Requirement: Git サーベイの対象種別
WHEN Git プラグインがサーベイを実行する場合、
システムは以下 4 種のレポジトリ活動を取得して `survey_type` メタデータで識別しなければならない (SHALL)。

- `merged_pr`: 過去 30 日以内にマージされた PR
- `open_pr`: state=open かつ直近 30 日以内に更新された PR
- `new_issue`: 直近 30 日以内に作成された Issue
- `release`: 直近 30 日以内の published リリース（Draft 除外）

#### Scenario: 4 種別の収集
GIVEN サーベイ対象リポジトリに関連 PR / Issue / リリースがある
WHEN `surv111eyRepo` が呼ばれる
THEN 各項目がそれぞれ対応する `survey_type` メタデータ付きで Submit される

#### Scenario: Draft リリースの除外
GIVEN リポジトリに Draft リリースが存在する
WHEN サーベイが実行される
THEN Draft は Submit されない
