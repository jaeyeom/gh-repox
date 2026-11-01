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

// TestIntegrationCreateAndVerify exercises create, edit, and the squash merge
// commit message against GitHub. It is skipped unless GH_REPOX_INTEGRATION=1
// is set. The repository is left in place; deleting it does not test gh-repox.
//
// Prerequisites:
//   - gh auth login with the repo scope
func TestIntegrationCreateAndVerify(t *testing.T) {
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
		Owner:                    login,
		Repo:                     repoName,
		Private:                  true,
		HasIssues:                true,
		HasWiki:                  false,
		AllowSquashMerge:         true,
		AllowMergeCommit:         false,
		AllowRebaseMerge:         false,
		AllowAutoMerge:           false,
		DeleteBranchOnMerge:      true,
		SquashMergeCommitMessage: "pr-title-description",
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

	// 5. Set the squash merge commit message the same way create does.
	// EditRepo intentionally leaves this to a separate REST PATCH.
	if err := ghclient.RetryOnNotFound(ctx, ghclient.DefaultRetryAttempts, ghclient.DefaultRetryDelay, func() error {
		return client.SetSquashMergeCommitMessage(ctx, fullName, p.SquashMergeCommitMessage)
	}); err != nil {
		t.Fatalf("SetSquashMergeCommitMessage: %v", err)
	}

	// 6. Fetch state and verify key settings roundtripped.
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
	if state.SquashMergeCommitMessage == nil || *state.SquashMergeCommitMessage != "pr-title-description" {
		got := "<nil>"
		if state.SquashMergeCommitMessage != nil {
			got = *state.SquashMergeCommitMessage
		}
		t.Errorf("expected squash_merge_commit_message=pr-title-description, got %s", got)
	}

	t.Logf("integration smoke test passed for %s; left in place for manual deletion", fullName)
}
