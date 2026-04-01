package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	c := Defaults()
	if c.Private.Value != true {
		t.Error("default private should be true")
	}
	if c.Private.Source != SourceDefault {
		t.Error("default private source should be 'default'")
	}
	if c.AllowSquashMerge.Value != true {
		t.Error("default allow_squash_merge should be true")
	}
	if c.AllowMergeCommit.Value != false {
		t.Error("default allow_merge_commit should be false")
	}
	if c.HasWiki.Value != false {
		t.Error("default has_wiki should be false")
	}
	if c.Host.Value != "github.com" {
		t.Error("default host should be github.com")
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
github:
  host: github.example.com
  owner: testuser
repo:
  private: false
  has_wiki: true
merge:
  allow_merge_commit: true
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	c := Defaults()
	if err := c.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if c.Host.Value != "github.example.com" {
		t.Errorf("got host=%q, want github.example.com", c.Host.Value)
	}
	if c.Host.Source != SourceConfig {
		t.Errorf("got host source=%q, want config", c.Host.Source)
	}
	if c.Private.Value != false {
		t.Error("loaded private should be false")
	}
	if c.HasWiki.Value != true {
		t.Error("loaded has_wiki should be true")
	}
	if c.AllowMergeCommit.Value != true {
		t.Error("loaded allow_merge_commit should be true")
	}
	if c.AllowSquashMerge.Source != SourceDefault {
		t.Error("unset allow_squash_merge source should remain default")
	}
}

func TestLoadEnv(t *testing.T) {
	c := Defaults()
	t.Setenv("REPOX_HOST", "ghe.corp.com")
	t.Setenv("REPOX_OWNER", "envuser")
	t.Setenv("REPOX_PRIVATE", "false")
	t.Setenv("REPOX_STRICT", "true")
	c.LoadEnv()
	if c.Host.Value != "ghe.corp.com" {
		t.Errorf("got host=%q", c.Host.Value)
	}
	if c.Host.Source != SourceEnv {
		t.Errorf("got host source=%q", c.Host.Source)
	}
	if c.Owner.Value != "envuser" {
		t.Errorf("got owner=%q", c.Owner.Value)
	}
	if c.Private.Value != false {
		t.Error("env private should be false")
	}
	if c.Strict.Value != true {
		t.Error("env strict should be true")
	}
}

func TestPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
github:
  host: config-host.com
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	c := Defaults()
	if err := c.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REPOX_HOST", "env-host.com")
	c.LoadEnv()
	if c.Host.Value != "env-host.com" {
		t.Errorf("env should override config: got %q", c.Host.Value)
	}
	if c.Host.Source != SourceEnv {
		t.Errorf("source should be env: got %q", c.Host.Source)
	}
}

func TestFindConfigFile_Explicit(t *testing.T) {
	got := FindConfigFile("/tmp/explicit.yaml")
	if got != "/tmp/explicit.yaml" {
		t.Errorf("explicit path: got %q", got)
	}
}

func TestEntries(t *testing.T) {
	c := Defaults()
	entries := c.Entries()
	if len(entries) == 0 {
		t.Fatal("entries should not be empty")
	}
	found := false
	for _, e := range entries {
		if e.Key == "private" {
			found = true
			if e.Value != true {
				t.Error("private entry value should be true")
			}
		}
	}
	if !found {
		t.Error("should have 'private' entry")
	}
}
