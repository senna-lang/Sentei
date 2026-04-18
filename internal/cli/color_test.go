package cli

import (
	"os"
	"testing"

	"github.com/senna-lang/sentei/internal/plugin"
)

func TestUrgencyColor_Mapping(t *testing.T) {
	tests := []struct {
		urgency plugin.Urgency
		want    Color
	}{
		{plugin.UrgencyUrgent, ColorRed},
		{plugin.UrgencyShouldCheck, ColorYellow},
		{plugin.UrgencyCanWait, ColorReset},
		{plugin.UrgencyIgnore, ColorGray},
		{plugin.Urgency("unknown"), ColorReset},
	}

	for _, tt := range tests {
		got := UrgencyColor(tt.urgency)
		if got != tt.want {
			t.Errorf("UrgencyColor(%q) = %q, want %q", tt.urgency, got, tt.want)
		}
	}
}

func TestColorize_WithColor(t *testing.T) {
	// NO_COLOR が未設定の場合、色付きになる
	os.Unsetenv("NO_COLOR")

	got := Colorize("test", ColorRed)
	want := "\033[31mtest\033[0m"
	if got != want {
		t.Errorf("Colorize = %q, want %q", got, want)
	}
}

func TestColorize_NoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	got := Colorize("test", ColorRed)
	if got != "test" {
		t.Errorf("with NO_COLOR, Colorize should return plain text, got %q", got)
	}
}

func TestFormatUrgency_Urgent(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	got := FormatUrgency(plugin.UrgencyUrgent)
	if got == "urgent" {
		t.Error("FormatUrgency should add ANSI codes when NO_COLOR is unset")
	}
}

func TestSuccess_Green(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	got := Success("ok")
	want := "\033[32mok\033[0m"
	if got != want {
		t.Errorf("Success = %q, want %q", got, want)
	}
}

func TestError_Red(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	got := Error("fail")
	want := "\033[31mfail\033[0m"
	if got != want {
		t.Errorf("Error = %q, want %q", got, want)
	}
}
