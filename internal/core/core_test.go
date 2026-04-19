/**
 * コアエンジンのテスト
 * 主に metadata["urgency_floor"] による post-process のカバレッジ
 */
package core

import (
	"testing"

	"github.com/senna-lang/sentei/internal/plugin"
)

func TestApplyUrgencyFloor_GradesUp(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyCanWait, Category: "llm_news"}
	meta := map[string]string{"urgency_floor": "should_check"}

	got := applyUrgencyFloor(label, meta)

	if got.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("urgency = %q, want %q", got.Urgency, plugin.UrgencyShouldCheck)
	}
}

func TestApplyUrgencyFloor_AlreadyAboveFloor_NoOp(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyShouldCheck, Category: "llm_news"}
	meta := map[string]string{"urgency_floor": "can_wait"}

	got := applyUrgencyFloor(label, meta)

	if got.Urgency != plugin.UrgencyShouldCheck {
		t.Errorf("urgency = %q, want should_check (unchanged)", got.Urgency)
	}
}

func TestApplyUrgencyFloor_NoMetadata_NoOp(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyIgnore, Category: "other"}

	got := applyUrgencyFloor(label, map[string]string{})

	if got.Urgency != plugin.UrgencyIgnore {
		t.Errorf("urgency = %q, want ignore (unchanged, metadata 無し)", got.Urgency)
	}
}

func TestApplyUrgencyFloor_EmptyFloor_NoOp(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyIgnore, Category: "other"}
	meta := map[string]string{"urgency_floor": ""}

	got := applyUrgencyFloor(label, meta)

	if got.Urgency != plugin.UrgencyIgnore {
		t.Errorf("urgency = %q, want ignore (unchanged, 空文字列)", got.Urgency)
	}
}

func TestApplyUrgencyFloor_InvalidFloor_NoOp(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyCanWait, Category: "other"}
	meta := map[string]string{"urgency_floor": "super_urgent"}

	got := applyUrgencyFloor(label, meta)

	if got.Urgency != plugin.UrgencyCanWait {
		t.Errorf("urgency = %q, want can_wait (不正値は無視)", got.Urgency)
	}
}

func TestApplyUrgencyFloor_UrgentStaysUrgent(t *testing.T) {
	label := plugin.Label{Urgency: plugin.UrgencyUrgent, Category: "llm_news"}
	meta := map[string]string{"urgency_floor": "should_check"}

	got := applyUrgencyFloor(label, meta)

	if got.Urgency != plugin.UrgencyUrgent {
		t.Errorf("urgency = %q, want urgent (最上位は維持)", got.Urgency)
	}
}
