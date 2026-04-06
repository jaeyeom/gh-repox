package github

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jaeyeom/gh-repox/internal/exec"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

func TestGetAuthenticatedUser(t *testing.T) {
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stdout: "octocat\n"},
		},
	}
	c := NewClient(mock, "")
	user, err := c.GetAuthenticatedUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user != "octocat" {
		t.Errorf("got %q", user)
	}
}

func TestGetAuthenticatedUser_Error(t *testing.T) {
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stderr: "not logged in", Err: fmt.Errorf("exit 1")},
		},
	}
	c := NewClient(mock, "")
	_, err := c.GetAuthenticatedUser(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateRepoArgs(t *testing.T) {
	p := &policy.DesiredPolicy{
		Owner:   "user",
		Repo:    "my-repo",
		Private: true,
	}
	args := CreateRepoArgs(p)
	if args[0] != "repo" || args[1] != "create" || args[2] != "user/my-repo" {
		t.Errorf("unexpected args: %v", args)
	}
	found := false
	for _, a := range args {
		if a == "--private" {
			found = true
		}
	}
	if !found {
		t.Error("should have --private flag")
	}
}

func TestCreateRepoArgs_Public(t *testing.T) {
	p := &policy.DesiredPolicy{
		Owner:   "user",
		Repo:    "my-repo",
		Private: false,
	}
	args := CreateRepoArgs(p)
	found := false
	for _, a := range args {
		if a == "--public" {
			found = true
		}
	}
	if !found {
		t.Error("should have --public flag")
	}
}

func TestCreateRepoArgs_Template(t *testing.T) {
	p := &policy.DesiredPolicy{
		Owner:    "user",
		Repo:     "my-repo",
		Private:  true,
		Template: "acme/go-template",
	}
	args := CreateRepoArgs(p)
	found := false
	for i, a := range args {
		if a == "--template" && i+1 < len(args) && args[i+1] == "acme/go-template" {
			found = true
		}
	}
	if !found {
		t.Error("should have --template flag")
	}
}

func TestEditRepoArgs(t *testing.T) {
	p := &policy.DesiredPolicy{
		AllowSquashMerge:    true,
		AllowMergeCommit:    false,
		AllowRebaseMerge:    false,
		AllowAutoMerge:      true,
		DeleteBranchOnMerge: true,
	}
	args := EditRepoArgs("user/repo", p)
	if args[0] != "repo" || args[1] != "edit" || args[2] != "user/repo" {
		t.Errorf("unexpected prefix: %v", args)
	}
}

func TestCloneRepoArgs(t *testing.T) {
	args := CloneRepoArgs("user/repo", "", nil)
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d: %v", len(args), args)
	}

	args = CloneRepoArgs("user/repo", "/tmp/dest", []string{"--depth=1"})
	if args[3] != "/tmp/dest" {
		t.Errorf("expected dir at index 3: %v", args)
	}
	found := false
	for _, a := range args {
		if a == "--depth=1" {
			found = true
		}
	}
	if !found {
		t.Error("should have extra args")
	}
}

func TestRepoExists(t *testing.T) {
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stdout: "repo info"},
		},
	}
	c := NewClient(mock, "")
	exists, err := c.RepoExists(context.Background(), "user/repo")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("should exist")
	}
}

func TestRepoExists_NotFound(t *testing.T) {
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stderr: "GraphQL: Could not resolve to a Repository", Err: fmt.Errorf("exit 1")},
		},
	}
	c := NewClient(mock, "")
	exists, err := c.RepoExists(context.Background(), "user/repo")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("should not exist")
	}
}

func TestRepoExists_Error(t *testing.T) {
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stderr: "authentication required", Err: fmt.Errorf("exit 1")},
		},
	}
	c := NewClient(mock, "")
	_, err := c.RepoExists(context.Background(), "user/repo")
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
}

func TestPlannedCommands(t *testing.T) {
	p := &policy.DesiredPolicy{
		Owner:            "user",
		Repo:             "my-repo",
		Private:          true,
		AllowSquashMerge: true,
		DependabotAlerts: true,
		CloneAfterCreate: true,
	}
	cmds := PlannedCommands(p, "")
	if len(cmds) < 3 {
		t.Errorf("expected at least 3 commands, got %d", len(cmds))
	}
}

func TestPlannedSecurityCommands(t *testing.T) {
	tests := []struct {
		name            string
		policy          *policy.DesiredPolicy
		wantSubstrings  []string
		wantNoSubstring []string
	}{
		{
			name: "all enabled",
			policy: &policy.DesiredPolicy{
				DependencyGraph:           true,
				DependabotAlerts:          true,
				DependabotSecurityUpdates: true,
			},
			wantSubstrings: []string{
				`echo '{"security_and_analysis":{"dependency_graph":{"status":"enabled"}}}' | gh api --method PATCH /repos/owner/repo --input -`,
				"gh api --method PUT /repos/owner/repo/vulnerability-alerts",
				"gh api --method PUT /repos/owner/repo/automated-security-fixes",
			},
		},
		{
			name: "all disabled",
			policy: &policy.DesiredPolicy{
				DependencyGraph:           false,
				DependabotAlerts:          false,
				DependabotSecurityUpdates: false,
			},
			wantSubstrings: []string{
				`echo '{"security_and_analysis":{"dependency_graph":{"status":"disabled"}}}' | gh api --method PATCH /repos/owner/repo --input -`,
				"gh api --method DELETE /repos/owner/repo/vulnerability-alerts",
				"gh api --method DELETE /repos/owner/repo/automated-security-fixes",
			},
		},
		{
			name: "commands are valid shell invocations",
			policy: &policy.DesiredPolicy{
				DependencyGraph: true,
			},
			wantNoSubstring: []string{
				"--input <{",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmds := PlannedSecurityCommands("owner/repo", tt.policy, "")
			if len(cmds) != 3 {
				t.Fatalf("expected 3 commands, got %d", len(cmds))
			}
			joined := strings.Join(cmds, "\n")
			for _, want := range tt.wantSubstrings {
				if !strings.Contains(joined, want) {
					t.Errorf("expected command output to contain %q, got:\n%s", want, joined)
				}
			}
			for _, noWant := range tt.wantNoSubstring {
				if strings.Contains(joined, noWant) {
					t.Errorf("command output should not contain %q, got:\n%s", noWant, joined)
				}
			}
		})
	}
}

func TestFetchRepoState(t *testing.T) {
	jsonResp := `{
		"isPrivate": true,
		"description": "test",
		"homepageUrl": "",
		"hasIssuesEnabled": true,
		"hasWikiEnabled": false,
		"hasProjectsEnabled": false,
		"squashMergeAllowed": true,
		"mergeCommitAllowed": true,
		"rebaseMergeAllowed": false,
		"autoMergeAllowed": false,
		"deleteBranchOnMerge": false
	}`
	mock := &exec.MockRunner{
		Responses: []exec.MockCall{
			{Stdout: jsonResp},
		},
	}
	c := NewClient(mock, "")
	state, err := c.FetchRepoState(context.Background(), "user/repo")
	if err != nil {
		t.Fatal(err)
	}
	if !state.Private {
		t.Error("should be private")
	}
	if state.Description != "test" {
		t.Errorf("got description=%q", state.Description)
	}
	if !state.AllowMergeCommit {
		t.Error("merge commit should be allowed")
	}
}
