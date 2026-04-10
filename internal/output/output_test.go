package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jaeyeom/gh-repox/internal/config"
	"github.com/jaeyeom/gh-repox/internal/diff"
)

func TestPrintCreateHuman(t *testing.T) {
	var buf bytes.Buffer
	r := &CreateResult{
		URL:     "https://github.com/test/repo",
		Applied: map[string]any{"private": true},
		Clone:   CloneResult{Requested: false},
	}
	PrintCreateHuman(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "https://github.com/test/repo") {
		t.Error("should contain repo URL")
	}
	if !strings.Contains(out, "skipped") {
		t.Error("should say clone skipped")
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	r := &CreateResult{Command: "create", Repo: "test/repo"}
	if err := PrintJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"command": "create"`) {
		t.Error("JSON should contain command field")
	}
}

func TestPrintConfigExplain(t *testing.T) {
	var buf bytes.Buffer
	entries := []config.Entry{
		{Key: "private", Value: true, Source: config.SourceDefault},
		{Key: "host", Value: "github.com", Source: config.SourceConfig},
	}
	PrintConfigExplain(&buf, entries)
	out := buf.String()
	if !strings.Contains(out, "source=default") {
		t.Error("should show source=default")
	}
	if !strings.Contains(out, "source=config") {
		t.Error("should show source=config")
	}
}

func TestPrintDiffHuman(t *testing.T) {
	var buf bytes.Buffer
	r := &DiffResult{
		Repo: "test/repo",
		Differences: []diff.Entry{
			{Field: "has_wiki", Current: true, Desired: false, DesiredSource: "default", Status: diff.StatusDifferent},
		},
	}
	PrintDiffHuman(&buf, r)
	if !strings.Contains(buf.String(), "has_wiki") {
		t.Error("should contain has_wiki")
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "''"},
		{"simple", "simple"},
		{"--flag=value", "--flag=value"},
		{"owner/repo", "owner/repo"},
		{"worker service", "'worker service'"},
		{"it's", `'it'\''s'`},
		{"hello world", "'hello world'"},
		{"has\ttab", "'has\ttab'"},
		{"a;b", "'a;b'"},
	}
	for _, tt := range tests {
		if got := ShellQuote(tt.in); got != tt.want {
			t.Errorf("ShellQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatCommand(t *testing.T) {
	got := FormatCommand("gh", "repo", "create", "test/repo")
	if got != "gh repo create test/repo" {
		t.Errorf("got %q", got)
	}
}

func TestFormatCommandQuotesSpaces(t *testing.T) {
	got := FormatCommand("gh", "repo", "create", "owner/repo", "--description", "worker service")
	want := "gh repo create owner/repo --description 'worker service'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
