package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jaeyeom/gh-repox/internal/output"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

// runCmd executes the root command with the given args, returning the error.
// It resets package-level flags before each run.
func runCmd(args ...string) error {
	// Reset global flags to defaults before each test invocation.
	flagConfig = ""
	flagHost = ""
	flagOwner = ""
	flagOrg = ""
	flagJSON = false
	flagDryRun = false
	flagStrict = false

	cmd := newRootCmd()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

// assertExitCode checks that err is an *ExitError with the expected code.
func assertExitCode(t *testing.T, err error, wantCode int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected ExitError with code %d, got nil", wantCode)
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != wantCode {
		t.Errorf("exit code = %d, want %d (err: %v)", exitErr.Code, wantCode, exitErr)
	}
}

func TestDiffInvalidRepo_ExitCode2(t *testing.T) {
	err := runCmd("diff", "badformat")
	assertExitCode(t, err, ExitInvalidInput)
}

func TestDiffBadConfig_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/bad.yaml"
	if err := os.WriteFile(tmpFile, []byte("{{invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("diff", "owner/repo", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestApplyInvalidRepo_ExitCode2(t *testing.T) {
	err := runCmd("apply", "badformat")
	assertExitCode(t, err, ExitInvalidInput)
}

func TestApplyBadConfig_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/bad.yaml"
	if err := os.WriteFile(tmpFile, []byte("{{invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("apply", "owner/repo", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestConfigShowBadConfig_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/bad.yaml"
	if err := os.WriteFile(tmpFile, []byte("{{invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("config", "show", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestConfigExplainBadConfig_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/bad.yaml"
	if err := os.WriteFile(tmpFile, []byte("{{invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("config", "explain", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestCreateInvalidSquashMergeCommitMessage_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/repox.yaml"
	if err := os.WriteFile(tmpFile, []byte("merge:\n  squash_merge_commit_message: bogus\n"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("create", "test-repo", "--dry-run", "--owner", "testowner", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestApplyInvalidSquashMergeCommitMessage_ExitCode2(t *testing.T) {
	tmpFile := t.TempDir() + "/repox.yaml"
	if err := os.WriteFile(tmpFile, []byte("merge:\n  squash_merge_commit_message: bogus\n"), 0600); err != nil {
		t.Fatal(err)
	}
	err := runCmd("apply", "testowner/test-repo", "--dry-run", "--config", tmpFile)
	assertExitCode(t, err, ExitInvalidInput)
}

func TestCreateConflictingFlags_ExitCode2(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
	}{
		{"private and public", []string{"--private", "--public"}},
		{"enable and disable issues", []string{"--enable-issues", "--disable-issues"}},
		{"enable and disable wiki", []string{"--enable-wiki", "--disable-wiki"}},
		{"enable and disable projects", []string{"--enable-projects", "--disable-projects"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"create", "test-repo", "--dry-run", "--owner", "testowner"}, tt.flags...)
			err := runCmd(args...)
			assertExitCode(t, err, ExitInvalidInput)
		})
	}
}

func TestCreateDryRunJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCmd("create", "test-repo", "--dry-run", "--json", "--owner", "testowner")

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.DryRunResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result.Command != "create" {
		t.Errorf("command = %q, want %q", result.Command, "create")
	}
	if !result.DryRun {
		t.Error("dryRun should be true")
	}
	if result.Repo != "testowner/test-repo" {
		t.Errorf("repo = %q, want %q", result.Repo, "testowner/test-repo")
	}
	if len(result.Commands) == 0 {
		t.Error("expected non-empty commands list")
	}
	foundSquash := false
	for _, cmd := range result.Commands {
		if strings.Contains(cmd, `"squash_merge_commit_title":"PR_TITLE"`) && strings.Contains(cmd, `"squash_merge_commit_message":"PR_BODY"`) {
			foundSquash = true
			break
		}
	}
	if !foundSquash {
		t.Errorf("create dry-run should plan squash commit message PATCH, got: %v", result.Commands)
	}
}

func TestApplyDryRunJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCmd("apply", "testowner/test-repo", "--dry-run", "--json")

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result output.DryRunResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result.Command != "apply" {
		t.Errorf("command = %q, want %q", result.Command, "apply")
	}
	if !result.DryRun {
		t.Error("dryRun should be true")
	}
	if result.Repo != "testowner/test-repo" {
		t.Errorf("repo = %q, want %q", result.Repo, "testowner/test-repo")
	}
	foundSquash := false
	for _, cmd := range result.Commands {
		if strings.Contains(cmd, `"squash_merge_commit_title":"PR_TITLE"`) && strings.Contains(cmd, `"squash_merge_commit_message":"PR_BODY"`) {
			foundSquash = true
			break
		}
	}
	if !foundSquash {
		t.Errorf("apply dry-run should plan squash commit message PATCH, got: %v", result.Commands)
	}
}

func runCmdCapture(t *testing.T, args ...string) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	cmdErr := runCmd(args...)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.String(), cmdErr
}

func decodeCreateDryRun(t *testing.T, out string) output.DryRunResult {
	t.Helper()
	var result output.DryRunResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	return result
}

func TestCreateOwnerRepoDryRunJSON(t *testing.T) {
	out, err := runCmdCapture(t, "create", "jaeyeom/agentic-factory", "--dry-run", "--json", "--clone")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	result := decodeCreateDryRun(t, out)
	if result.Repo != "jaeyeom/agentic-factory" {
		t.Errorf("repo = %q, want %q", result.Repo, "jaeyeom/agentic-factory")
	}
	var cloneCmd string
	for _, cmd := range result.Commands {
		if strings.Contains(cmd, "repo clone") {
			cloneCmd = cmd
			break
		}
	}
	if cloneCmd != "gh repo clone jaeyeom/agentic-factory" {
		t.Errorf("clone command = %q, want destination to stay the repository name", cloneCmd)
	}
}

func TestCreateOwnerRepoIgnoresConfigAndEnvOwner(t *testing.T) {
	cfgPath := t.TempDir() + "/repox.yaml"
	body := "github:\n  owner: configowner\n  org: configorg\n"
	if err := os.WriteFile(cfgPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REPOX_OWNER", "envowner")
	t.Setenv("REPOX_ORG", "envorg")

	out, err := runCmdCapture(t, "create", "jaeyeom/agentic-factory", "--dry-run", "--json", "--config", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	result := decodeCreateDryRun(t, out)
	if result.Repo != "jaeyeom/agentic-factory" {
		t.Errorf("repo = %q, want positional owner to win over config and env", result.Repo)
	}
}

func TestCreateBareNameOrgOverridesOwner(t *testing.T) {
	out, err := runCmdCapture(t, "create", "my-service", "--dry-run", "--json", "--org", "acme", "--owner", "someone")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	result := decodeCreateDryRun(t, out)
	if result.Repo != "acme/my-service" {
		t.Errorf("repo = %q, want %q", result.Repo, "acme/my-service")
	}
}

func TestCreateOwnerRepoFlagConflict(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{
			name:    "org differs",
			args:    []string{"create", "jaeyeom/agentic-factory", "--dry-run", "--org", "other"},
			wantMsg: `conflicts with --org "other"`,
		},
		{
			name:    "owner differs",
			args:    []string{"create", "jaeyeom/agentic-factory", "--dry-run", "--owner", "other"},
			wantMsg: `conflicts with --owner "other"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCmd(tt.args...)
			assertExitCode(t, err, ExitInvalidInput)
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %v, want substring %q", err, tt.wantMsg)
			}
		})
	}
}

func TestCreateOwnerRepoMatchingFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "org matches", args: []string{"create", "jaeyeom/agentic-factory", "--dry-run", "--json", "--org", "jaeyeom"}},
		{name: "owner matches", args: []string{"create", "jaeyeom/agentic-factory", "--dry-run", "--json", "--owner", "jaeyeom"}},
		{name: "both match", args: []string{"create", "jaeyeom/agentic-factory", "--dry-run", "--json", "--org", "jaeyeom", "--owner", "jaeyeom"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCmdCapture(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v\noutput: %s", err, out)
			}
			result := decodeCreateDryRun(t, out)
			if result.Repo != "jaeyeom/agentic-factory" {
				t.Errorf("repo = %q, want %q", result.Repo, "jaeyeom/agentic-factory")
			}
		})
	}
}

func TestCreateInvalidOwnerRepo(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantMsg string
	}{
		{name: "empty repo", arg: "owner/", wantMsg: "invalid repo format"},
		{name: "empty owner", arg: "/repo", wantMsg: "invalid repo format"},
		{name: "extra slash", arg: "a/b/c", wantMsg: `invalid repository name "b/c"`},
		{name: "invalid owner", arg: "bad;owner/repo", wantMsg: "invalid owner"},
		{name: "invalid repo", arg: "owner/bad;repo", wantMsg: "invalid repository name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCmd("create", tt.arg, "--dry-run")
			assertExitCode(t, err, ExitInvalidInput)
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %v, want substring %q", err, tt.wantMsg)
			}
		})
	}
}

func TestCreateUsageDocumentsOwnerRepo(t *testing.T) {
	cmd := newCreateCmd()
	if cmd.Use != "create <name> | <owner/repo>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <name> | <owner/repo>")
	}
}

func TestBuildAppliedMapNotPrePopulated(t *testing.T) {
	p := &policy.DesiredPolicy{
		Private:             true,
		AllowSquashMerge:    true,
		AllowMergeCommit:    false,
		AllowRebaseMerge:    false,
		AllowAutoMerge:      true,
		DeleteBranchOnMerge: true,
		HasIssues:           true,
		HasWiki:             false,
		HasProjects:         false,
	}
	m := buildAppliedMap(p)
	if len(m) == 0 {
		t.Fatal("buildAppliedMap should return non-empty map")
	}
	if m["private"] != true {
		t.Errorf("private = %v, want true", m["private"])
	}
	if _, ok := m["squash_merge_commit_message"]; ok {
		t.Error("squash_merge_commit_message is applied via REST PATCH, not buildAppliedMap")
	}
}
