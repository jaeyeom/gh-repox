package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
}
