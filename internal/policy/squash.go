package policy

import (
	"github.com/jaeyeom/gh-repox/internal/config"
)

// squashCommitAPI maps a UI/config preset to GitHub REST API fields.
var squashCommitAPI = map[string][2]string{
	config.SquashCommitMessageDefault:            {"COMMIT_OR_PR_TITLE", "COMMIT_MESSAGES"},
	config.SquashCommitMessagePRTitle:            {"PR_TITLE", "BLANK"},
	config.SquashCommitMessagePRTitleCommits:     {"PR_TITLE", "COMMIT_MESSAGES"},
	config.SquashCommitMessagePRTitleDescription: {"PR_TITLE", "PR_BODY"},
}

// SquashCommitAPIFields returns the REST API title and message values for a
// squash merge commit message preset. ok is false if the preset is unknown.
func SquashCommitAPIFields(preset string) (title, message string, ok bool) {
	fields, ok := squashCommitAPI[preset]
	if !ok {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// SquashCommitPreset returns the UI/config preset for a pair of REST API
// squash merge commit title and message values. ok is false if the pair is
// not a known GitHub UI option.
func SquashCommitPreset(title, message string) (preset string, ok bool) {
	for preset, fields := range squashCommitAPI {
		if fields[0] == title && fields[1] == message {
			return preset, true
		}
	}
	return "", false
}

// ValidSquashCommitMessage reports whether preset is a known squash merge
// commit message option.
func ValidSquashCommitMessage(preset string) bool {
	_, ok := squashCommitAPI[preset]
	return ok
}
