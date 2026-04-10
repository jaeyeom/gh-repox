package github_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jaeyeom/gh-repox/internal/exec"
	ghclient "github.com/jaeyeom/gh-repox/internal/github"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

// TestIntegrationCreateAndDelete exercises the real gh CLI against GitHub.
// It is skipped unless GH_REPOX_INTEGRATION=1 is set.
//
// Prerequisites:
//   - gh auth login (the authenticated user must be able to create/delete repos)
func TestIntegrationCreateAndDelete(t *testing.T) {
	if os.Getenv("GH_REPOX_INTEGRATION") != "1" {
		t.Skip("skipping integration test; set GH_REPOX_INTEGRATION=1 to run")
	}

	ctx := context.Background()
	runner := &exec.RealRunner{}
	client := ghclient.NewClient(runner, "github.com")

	// Discover authenticated user.
	login, err := client.GetAuthenticatedUser(ctx)
	if err != nil {
		t.Fatalf("GetAuthenticatedUser: %v", err)
	}

	repoName := fmt.Sprintf("gh-repox-smoke-%d", time.Now().UnixMilli())
	fullName := login + "/" + repoName

	// Ensure cleanup even on failure.
	t.Cleanup(func() {
		// gh repo delete requires --yes to skip confirmation.
		_, _, _ = runner.Run(ctx, "gh", "repo", "delete", fullName, "--yes")
	})

	// 1. Repo should not exist yet.
	exists, err := client.RepoExists(ctx, fullName)
	if err != nil {
		t.Fatalf("RepoExists (before create): %v", err)
	}
	if exists {
		t.Fatalf("repo %s already exists before test", fullName)
	}

	// 2. Create the repo.
	p := &policy.DesiredPolicy{
		Owner:               login,
		Repo:                repoName,
		Private:             true,
		HasIssues:           true,
		HasWiki:             false,
		AllowSquashMerge:    true,
		AllowMergeCommit:    false,
		AllowRebaseMerge:    false,
		AllowAutoMerge:      false,
		DeleteBranchOnMerge: true,
	}

	url, err := client.CreateRepo(ctx, p)
	if err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	if url == "" {
		t.Fatal("CreateRepo returned empty URL")
	}
	t.Logf("created repo: %s (%s)", fullName, url)

	// 3. Repo should exist now (allow brief propagation delay).
	var found bool
	for attempt := 0; attempt < 5; attempt++ {
		exists, err = client.RepoExists(ctx, fullName)
		if err != nil {
			t.Fatalf("RepoExists (after create): %v", err)
		}
		if exists {
			found = true
			break
		}
		t.Logf("repo not yet visible, retrying (%d/5)...", attempt+1)
		time.Sleep(2 * time.Second)
	}
	if !found {
		t.Fatalf("repo %s does not exist after creation (waited 10s)", fullName)
	}

	// 4. Edit the repo (the post-create step that Finding 1 is about).
	if err := client.EditRepo(ctx, fullName, p); err != nil {
		t.Fatalf("EditRepo: %v", err)
	}

	// 5. Fetch state and verify key settings roundtripped.
	state, err := client.FetchRepoState(ctx, fullName)
	if err != nil {
		t.Fatalf("FetchRepoState: %v", err)
	}
	if !state.Private {
		t.Error("expected private=true")
	}
	if !state.AllowSquashMerge {
		t.Error("expected allow_squash_merge=true")
	}
	if state.AllowMergeCommit {
		t.Error("expected allow_merge_commit=false")
	}
	if state.DeleteBranchOnMerge != true {
		t.Error("expected delete_branch_on_merge=true")
	}

	t.Logf("integration smoke test passed for %s", fullName)
	// Cleanup runs via t.Cleanup above.
}
