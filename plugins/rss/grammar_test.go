package rss

import (
	"strings"
	"testing"
)

func TestGrammar_NoUrgencyField(t *testing.T) {
	// RSS は urgency を廃止したので grammar に含まれないこと
	if strings.Contains(Grammar, "urgency") {
		t.Error("RSS grammar に urgency が残っている (廃止済みのはず)")
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
