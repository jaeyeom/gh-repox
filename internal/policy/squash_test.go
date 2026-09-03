package policy

import (
	"testing"

	"github.com/jaeyeom/gh-repox/internal/config"
)

func TestSquashCommitAPIFields(t *testing.T) {
	tests := []struct {
		preset  string
		title   string
		message string
		ok      bool
	}{
		{config.SquashCommitMessageDefault, "COMMIT_OR_PR_TITLE", "COMMIT_MESSAGES", true},
		{config.SquashCommitMessagePRTitle, "PR_TITLE", "BLANK", true},
		{config.SquashCommitMessagePRTitleCommits, "PR_TITLE", "COMMIT_MESSAGES", true},
		{config.SquashCommitMessagePRTitleDescription, "PR_TITLE", "PR_BODY", true},
		{"", "", "", false},
		{"bogus", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			title, message, ok := SquashCommitAPIFields(tt.preset)
			if ok != tt.ok {
				t.Fatalf("ok=%v, want %v", ok, tt.ok)
			}
			if title != tt.title || message != tt.message {
				t.Errorf("got title=%q message=%q, want title=%q message=%q", title, message, tt.title, tt.message)
			}
		})
	}
}

func TestSquashCommitPreset(t *testing.T) {
	tests := []struct {
		title   string
		message string
		preset  string
		ok      bool
	}{
		{"COMMIT_OR_PR_TITLE", "COMMIT_MESSAGES", config.SquashCommitMessageDefault, true},
		{"PR_TITLE", "BLANK", config.SquashCommitMessagePRTitle, true},
		{"PR_TITLE", "COMMIT_MESSAGES", config.SquashCommitMessagePRTitleCommits, true},
		{"PR_TITLE", "PR_BODY", config.SquashCommitMessagePRTitleDescription, true},
		{"", "PR_BODY", "", false},
		{"PR_TITLE", "", "", false},
		{"OTHER", "PR_BODY", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.title+"/"+tt.message, func(t *testing.T) {
			preset, ok := SquashCommitPreset(tt.title, tt.message)
			if ok != tt.ok {
				t.Fatalf("ok=%v, want %v", ok, tt.ok)
			}
			if preset != tt.preset {
				t.Errorf("got preset=%q, want %q", preset, tt.preset)
			}
		})
	}
}
