/**
 * Git プラグイン用の GBNF grammar とプロンプトテンプレート
 * POC で検証済みの定義をそのまま組み込む
 */
package git

// GitGrammar は Git プラグイン用の GBNF grammar
// category: pr, issue, ci, release, discussion, other
const GitGrammar = `root     ::= "{" ws "\"urgency\":" ws urgency "," ws "\"category\":" ws category "," ws "\"summary\":" ws summary "}" ws
urgency  ::= "\"urgent\"" | "\"should_check\"" | "\"can_wait\"" | "\"ignore\""
category ::= "\"pr\"" | "\"issue\"" | "\"ci\"" | "\"release\"" | "\"discussion\"" | "\"other\""
summary  ::= "\"" char char char char char+ "\""
char     ::= [^"\\\n]
ws       ::= [ \t\n]*`

// GitPromptTemplate は Git 通知のラベリング用プロンプトテンプレート
const GitPromptTemplate = `/no_think
You are classifying a GitHub notification for a software engineer.

User context: Japanese software engineer working on LLM infrastructure, local AI tools, and web development projects.

Category must be one of: pr, issue, ci, release, discussion, other
Urgency must be one of: urgent, should_check, can_wait, ignore

Rules:
- If the title contains "release" → category is ALWAYS "release", regardless of notification type
- review_requested from a mentor or important collaborator → urgent, category: pr
- mentioned in a bug issue → urgent or should_check, category: issue
- CI failure → should_check, category: ci
- CI success → can_wait or ignore, category: ci
- subscribed-only updates → can_wait or ignore
- PR merge notifications where user is not involved → can_wait, category: pr

Examples:
Input: {"type": "review_requested", "repo": "main-project", "title": "fix: score calculation bug"}
Output: {"urgency": "urgent", "category": "pr", "summary": "Review requested on PR fix: score calculation bug"}

Input: {"type": "review_requested", "repo": "app", "title": "release(electron): v0.0.27 to stg"}
Output: {"urgency": "should_check", "category": "release", "summary": "Release v0.0.27 to staging review requested"}

Input: {"type": "ci_activity", "repo": "web-app", "title": "CI workflow run failed for main branch"}
Output: {"urgency": "should_check", "category": "ci", "summary": "CI failed on main branch of web-app"}

Input: {"type": "mention", "repo": "web-app", "title": "Bug: login page crash on Safari"}
Output: {"urgency": "should_check", "category": "issue", "summary": "Mentioned in Safari login crash bug report"}

Now classify this notification. Output JSON only.

Notification:
{notification_json}`
