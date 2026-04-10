package diff

import (
	"testing"

	"github.com/jaeyeom/gh-repox/internal/config"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

func boolPtr(b bool) *bool { return &b }

func TestCompare(t *testing.T) {
	cfg := config.Defaults()
	desired := &policy.DesiredPolicy{
		Private:             true,
		HasIssues:           true,
		HasWiki:             false,
		HasProjects:         false,
		AllowSquashMerge:    true,
		AllowMergeCommit:    false,
		AllowRebaseMerge:    false,
		AllowAutoMerge:      true,
		DeleteBranchOnMerge: true,
		DependencyGraph:     true,
		DependabotAlerts:    true,
	}
	actual := &policy.ActualState{
		Private:             true,
		HasIssues:           true,
		HasWiki:             true,
		HasProjects:         false,
		AllowSquashMerge:    true,
		AllowMergeCommit:    true,
		AllowRebaseMerge:    false,
		AllowAutoMerge:      boolPtr(false),
		DeleteBranchOnMerge: true,
		DependencyGraph:     boolPtr(true),
		DependabotAlerts:    nil,
	}

	entries := Compare(desired, actual, cfg)
	if !HasDifferences(entries) {
		t.Fatal("should have differences")
	}

	diffs := map[string]Status{}
	for _, e := range entries {
		diffs[e.Field] = e.Status
	}

	if diffs["has_wiki"] != StatusDifferent {
		t.Error("has_wiki should be different")
	}
	if diffs["allow_merge_commit"] != StatusDifferent {
		t.Error("allow_merge_commit should be different")
	}
	if diffs["allow_auto_merge"] != StatusDifferent {
		t.Error("allow_auto_merge should be different")
	}
	if diffs["private"] != StatusSame {
		t.Error("private should be same")
	}
	if diffs["dependabot_alerts"] != StatusUnknown {
		t.Error("dependabot_alerts should be unknown")
	}
}

func TestFormatHuman_NoDiffs(t *testing.T) {
	entries := []Entry{
		{Field: "private", Current: true, Desired: true, DesiredSource: "default", Status: StatusSame},
	}
	out := FormatHuman(entries)
	if out != "No differences found.\n" {
		t.Error("should show no differences")
	}
}
