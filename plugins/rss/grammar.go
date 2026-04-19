/**
 * RSS プラグイン用の GBNF grammar とプロンプトテンプレート。
 * urgency は 3 値 (urgent は含まない)、category は 5 値。
 */
package rss

// Grammar は RSS エントリのラベリングに使う GBNF grammar。
// urgency から "urgent" を意図的に除外 (grammar 制約で生成不能化)。
const Grammar = `root     ::= "{" ws "\"urgency\":" ws urgency "," ws "\"category\":" ws category "," ws "\"summary\":" ws summary "}" ws
urgency  ::= "\"should_check\"" | "\"can_wait\"" | "\"ignore\""
category ::= "\"llm_research\"" | "\"llm_news\"" | "\"dev_tools\"" | "\"swe\"" | "\"other\""
summary  ::= "\"" char char char char char+ "\""
char     ::= [^"\\\n]
ws       ::= [ \t\n]*`

// PromptTemplate は RSS エントリのラベリング用プロンプト。
// buildPrompt (internal/bonsai) が {notification_json} をエントリ JSON に置換する。
const PromptTemplate = `/no_think
You are classifying an RSS feed entry for an LLM/AI/SWE-focused engineer.

User context: Japanese software engineer tracking LLM releases, ML research, and software engineering practices. Interests: AI, LLM, Claude Code, TypeScript, React.

Category must be one of: llm_research, llm_news, dev_tools, swe, other
Urgency must be one of: should_check, can_wait, ignore

Classification priority (apply in order):
1. Research / paper deep-dive (architecture, training, evaluation methods) -> llm_research
2. LLM product announcements, releases, or benchmarks -> llm_news
3. Specific tool / library / CLI / editor extension review -> dev_tools
4. Language / framework / design pattern / engineering practice (tool-agnostic) -> swe
5. None of the above -> other

LLM-related entries prefer llm_* categories. However, if the article's main theme is "how I used tool X" (tooling / workflow knowledge), choose dev_tools.

Urgency heuristics:
- should_check: high-relevance to user's interests (AI/LLM/SWE core topics)
- can_wait: adjacent / learning value but not pressing
- ignore: weakly related or noise

Examples:
Input: {"title":"Announcing Claude 3.5 Sonnet","metadata":{"feed_name":"Anthropic News"},"content":"Claude 3.5 Sonnet is available today..."}
Output: {"urgency":"should_check","category":"llm_news","summary":"Anthropic released Claude 3.5 Sonnet with improved benchmarks"}

Input: {"title":"RLHF から DPO への移行","metadata":{"feed_name":"Lil'Log (Lilian Weng)"},"content":"The shift from RLHF to Direct Preference Optimization..."}
Output: {"urgency":"should_check","category":"llm_research","summary":"RLHF から DPO への移行の理論的背景"}

Input: {"title":"Claude Code で TypeScript プロジェクトを refactor する","metadata":{"feed_name":"Zenn - Claude Code"},"content":"Claude Code を使って既存の TypeScript コードベースを..."}
Output: {"urgency":"should_check","category":"dev_tools","summary":"Claude Code による TypeScript refactor の実例"}

Input: {"title":"Next.js 15 の App Router 仕様変更","metadata":{"feed_name":"Zenn - React"},"content":"Next.js 15 で App Router の behavior が..."}
Output: {"urgency":"should_check","category":"swe","summary":"Next.js 15 の App Router 仕様変更まとめ"}

Input: {"title":"初心者が Docker 入門した記録","metadata":{"feed_name":"Qiita - TypeScript"},"content":"Docker 触ったことなかったので..."}
Output: {"urgency":"ignore","category":"other","summary":""}

Now classify this entry. Output JSON only.

Entry:
{notification_json}`
