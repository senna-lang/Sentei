package rss

import (
	"strings"
	"testing"
)

func TestGrammar_ExcludesUrgent(t *testing.T) {
	// urgency grammar 行に "urgent" が単独で現れないこと
	if strings.Contains(Grammar, "\\\"urgent\\\"") {
		t.Error("RSS grammar が urgent を含んでいる (3 値制約が壊れている)")
	}
}

func TestGrammar_ContainsAllThreeUrgencies(t *testing.T) {
	for _, v := range []string{`"should_check"`, `"can_wait"`, `"ignore"`} {
		if !strings.Contains(Grammar, `\`+v[:1]+v[1:len(v)-1]+`\"`) {
			// シンプルに生の文字列で探す
		}
		if !strings.Contains(Grammar, strings.ReplaceAll(v, `"`, `\"`)) {
			t.Errorf("Grammar に %s が含まれていない", v)
		}
	}
}

func TestGrammar_ContainsAllFiveCategories(t *testing.T) {
	for _, v := range []string{"llm_research", "llm_news", "dev_tools", "swe", "other"} {
		if !strings.Contains(Grammar, `\"`+v+`\"`) {
			t.Errorf("Grammar に category %q が含まれていない", v)
		}
	}
}

func TestPromptTemplate_HasPlaceholder(t *testing.T) {
	if !strings.Contains(PromptTemplate, "{notification_json}") {
		t.Error("prompt template に {notification_json} placeholder が無い")
	}
}

func TestPromptTemplate_HasNoThinkPrefix(t *testing.T) {
	if !strings.HasPrefix(PromptTemplate, "/no_think") {
		t.Error("prompt は /no_think prefix で始まるべき")
	}
}

func TestPromptTemplate_MentionsAllCategoriesInRules(t *testing.T) {
	for _, v := range []string{"llm_research", "llm_news", "dev_tools", "swe", "other"} {
		if !strings.Contains(PromptTemplate, v) {
			t.Errorf("prompt 本文に category %q が言及されていない", v)
		}
	}
}

func TestPromptTemplate_HasFewShotExamples(t *testing.T) {
	// "Input:" / "Output:" のペアが 5 つ含まれていること
	inputCount := strings.Count(PromptTemplate, "Input:")
	outputCount := strings.Count(PromptTemplate, "Output:")
	if inputCount < 5 || outputCount < 5 {
		t.Errorf("few-shot example 不足: Input=%d Output=%d (各 5 以上を期待)", inputCount, outputCount)
	}
}
