/**
 * RSS プラグイン用の GBNF grammar とプロンプトテンプレート。
 * urgency は 3 値 (urgent は含まない)、category は 5 値。
 */
package rss

// Grammar は RSS エントリのラベリングに使う GBNF grammar。
// urgency は含まない — RSS は category 分類のみで運用する (優先度は Bonsai に難しいため廃止)。
const Grammar = `root     ::= "{" ws "\"category\":" ws category "," ws "\"summary\":" ws summary "}" ws
category ::= "\"llm_research\"" | "\"llm_news\"" | "\"dev_tools\"" | "\"swe\"" | "\"other\""
summary  ::= "\"" char char char char char+ "\""
char     ::= [^"\\\n]
ws       ::= [ \t\n]*`

// PromptTemplate は RSS エントリのラベリング用プロンプト。
// buildPrompt (internal/bonsai) が {notification_json} をエントリ JSON に置換する。
// 優先度 (urgency) の判断は Bonsai には困難なため廃止し、category 分類のみを要求する。
const PromptTemplate = `/no_think
You are classifying an RSS feed entry. Pick the most specific category. Avoid "other" unless nothing else fits.

Categories (in decision order — pick the FIRST that matches):

1. llm_research
   - LLM research, papers, architectures, training methods, evaluation methodologies, alignment
   - Keywords: RLHF, DPO, attention, transformer, benchmark, fine-tuning, embedding, retrieval
   - "なぜ X は Y するのか" / "X の理論" / "X の原理" style deep-dives about LLM/ML

2. llm_news
   - Official LLM product announcements, releases, company news, pricing, benchmarks results
   - Keywords: 発表, リリース, 公開, announcing, available, launch, ships
   - Claude/GPT/Gemini/Llama/etc model updates

3. dev_tools
   - Using / reviewing / trying a specific tool, library, CLI, IDE extension, framework
   - Article's main point is "how tool X works" or "I tried X" or "X feature explained"
   - Examples: Claude Code usage, MCP servers, VSCode extensions, CLI utilities, SDKs
   - AI agents / AI workflow articles that focus on TOOLING also go here
   - Keywords: 試してみた, 使ってみた, で〜した, を動かす, を実装する, SDK, API, CLI

4. swe
   - Software engineering concepts tied to a language / framework / pattern (not tool-specific)
   - TypeScript / React / Go / Rust / Next.js / Rails features, APIs, design patterns
   - Architecture, algorithms, data structures, system design
   - Even if title mentions a beginner-sounding topic, if it's about a language/framework feature → swe

5. other
   - ONLY if truly none of the above apply.
   - Personal diaries, self-introductions, essays about AI industry / career
   - Off-topic (not tech), clickbait with no technical content

Disambiguation rules:
- "AI で X した" with specific tool → dev_tools, without tool → other
- TypeScript/React/Go/Next.js の機能/仕様/パターン → swe
- ツール/ライブラリ/エージェントを使った話 → dev_tools
- LLM 研究・理論 → llm_research
- モデルリリース/発表 → llm_news
- エッセイ・日記・自己紹介 → other

Summary: one short sentence (日本語可, 30 字以内目安).

Examples:

Input: {"title":"Announcing Claude 4","metadata":{"feed_name":"Anthropic News"},"content":"Claude 4 is available today..."}
Output: {"category":"llm_news","summary":"Anthropic が Claude 4 を公開"}

Input: {"title":"RLHF から DPO への理論的背景","metadata":{"feed_name":"Lil'Log"},"content":"The shift from RLHF to DPO..."}
Output: {"category":"llm_research","summary":"RLHF から DPO への理論的移行"}

Input: {"title":"GitHub Copilot SDK 完全入門","metadata":{"feed_name":"Zenn - AI"},"content":"Copilot SDK を使って自分のエージェントを作る..."}
Output: {"category":"dev_tools","summary":"Copilot SDK でエージェント作成"}

Input: {"title":"TypeScript 6.0 は新機能追加より移行準備版として見るべき","metadata":{"feed_name":"Zenn - TypeScript"},"content":"TypeScript 6.0 の新機能..."}
Output: {"category":"swe","summary":"TypeScript 6.0 の位置づけ解説"}

Input: {"title":"Claude Code で TypeScript 型定義を生成してみた","metadata":{"feed_name":"Zenn - Claude Code"},"content":"Claude Code を触ってみたので..."}
Output: {"category":"dev_tools","summary":"Claude Code で TS 型生成を試した記録"}

Input: {"title":"React の useTransition で宣言的ローディング UI","metadata":{"feed_name":"Zenn - React"},"content":"React 18 の useTransition を使って..."}
Output: {"category":"swe","summary":"useTransition の実装パターン"}

Input: {"title":"Mastra で作る AIエージェント(25) Mastra Platform","metadata":{"feed_name":"Zenn - AI"},"content":"Mastra Platform を使ってみる..."}
Output: {"category":"dev_tools","summary":"Mastra Platform 入門記"}

Input: {"title":"AI時代の仕事設計","metadata":{"feed_name":"Zenn - AI"},"content":"AI が普及する時代に..."}
Output: {"category":"other","summary":"AI 時代のエンジニア論"}

Input: {"title":"個人開発と「壁打ち」を記録するために Zenn を始めました","metadata":{"feed_name":"Zenn - AI"},"content":"こんにちは..."}
Output: {"category":"other","summary":"Zenn 開始の自己紹介"}

Now classify. Pick the most specific category. Avoid "other" unless nothing else fits. Output JSON only.

Entry:
{notification_json}`
