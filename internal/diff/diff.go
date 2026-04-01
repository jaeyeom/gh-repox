package diff

import (
	"fmt"

	"github.com/jaeyeom/gh-repox/internal/config"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

// Status represents the status of a diff entry.
type Status string

const (
	StatusSame        Status = "same"
	StatusDifferent   Status = "different"
	StatusUnknown     Status = "unknown"
	StatusUnsupported Status = "unsupported"
)

// Entry represents a single field comparison.
type Entry struct {
	Field         string `json:"field"`
	Current       any    `json:"current"`
	Desired       any    `json:"desired"`
	DesiredSource string `json:"desired_source"`
	Status        Status `json:"status"`
}

// Compare compares desired policy against actual state and returns diff entries.
func Compare(desired *policy.DesiredPolicy, actual *policy.ActualState, cfg *config.Config) []Entry {
	var entries []Entry

	addBool := func(field string, current, desired bool, source config.Source) {
		status := StatusSame
		if current != desired {
			status = StatusDifferent
		}
		entries = append(entries, Entry{
			Field:         field,
			Current:       current,
			Desired:       desired,
			DesiredSource: string(source),
			Status:        status,
		})
	}

	addOptBool := func(field string, current *bool, desired bool, source config.Source) {
		if current == nil {
			entries = append(entries, Entry{
				Field:         field,
				Current:       nil,
				Desired:       desired,
				DesiredSource: string(source),
				Status:        StatusUnknown,
			})
			return
		}
		addBool(field, *current, desired, source)
	}

	addString := func(field string, current, desired string, source config.Source) {
		status := StatusSame
		if current != desired {
			status = StatusDifferent
		}
		entries = append(entries, Entry{
			Field:         field,
			Current:       current,
			Desired:       desired,
			DesiredSource: string(source),
			Status:        status,
		})
	}

	addBool("private", actual.Private, desired.Private, cfg.Private.Source)
	addString("description", actual.Description, desired.Description, cfg.Description.Source)
	addString("homepage", actual.Homepage, desired.Homepage, cfg.Homepage.Source)
	addBool("has_issues", actual.HasIssues, desired.HasIssues, cfg.HasIssues.Source)
	addBool("has_wiki", actual.HasWiki, desired.HasWiki, cfg.HasWiki.Source)
	addBool("has_projects", actual.HasProjects, desired.HasProjects, cfg.HasProjects.Source)
	addBool("allow_squash_merge", actual.AllowSquashMerge, desired.AllowSquashMerge, cfg.AllowSquashMerge.Source)
	addBool("allow_merge_commit", actual.AllowMergeCommit, desired.AllowMergeCommit, cfg.AllowMergeCommit.Source)
	addBool("allow_rebase_merge", actual.AllowRebaseMerge, desired.AllowRebaseMerge, cfg.AllowRebaseMerge.Source)
	addBool("allow_auto_merge", actual.AllowAutoMerge, desired.AllowAutoMerge, cfg.AllowAutoMerge.Source)
	addBool("delete_branch_on_merge", actual.DeleteBranchOnMerge, desired.DeleteBranchOnMerge, cfg.DeleteBranchOnMerge.Source)
	addOptBool("dependency_graph", actual.DependencyGraph, desired.DependencyGraph, cfg.DependencyGraph.Source)
	addOptBool("dependabot_alerts", actual.DependabotAlerts, desired.DependabotAlerts, cfg.DependabotAlerts.Source)
	addOptBool("dependabot_security_updates", actual.DependabotSecurityUpdates, desired.DependabotSecurityUpdates, cfg.DependabotSecurityUpdates.Source)

	return entries
}

// HasDifferences returns true if any entry has a non-same status.
func HasDifferences(entries []Entry) bool {
	for _, e := range entries {
		if e.Status == StatusDifferent {
			return true
		}
	}
	return false
}

// FormatHuman formats diff entries for human display.
func FormatHuman(entries []Entry) string {
	var s string
	for _, e := range entries {
		if e.Status == StatusSame {
			continue
		}
		s += fmt.Sprintf("%s:\n  current: %v\n  desired: %v\n  source: %s\n  status: %s\n", e.Field, e.Current, e.Desired, e.DesiredSource, e.Status)
	}
	if s == "" {
		return "No differences found.\n"
	}
	return s
}
