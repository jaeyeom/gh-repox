package policy

import (
	"testing"

	"github.com/jaeyeom/gh-repox/internal/config"
)

func TestFromConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Owner.Set("testuser", config.SourceInferred)
	cfg.Description.Set("test repo", config.SourceFlag)

	p := FromConfig(cfg, "my-repo")
	if p.Owner != "testuser" {
		t.Errorf("got owner=%q", p.Owner)
	}
	if p.Repo != "my-repo" {
		t.Errorf("got repo=%q", p.Repo)
	}
	if p.FullName() != "testuser/my-repo" {
		t.Errorf("got fullname=%q", p.FullName())
	}
	if !p.Private {
		t.Error("should be private by default")
	}
	if p.Description != "test repo" {
		t.Errorf("got description=%q", p.Description)
	}
}

func TestFromConfig_OrgOverridesOwner(t *testing.T) {
	cfg := config.Defaults()
	cfg.Owner.Set("testuser", config.SourceInferred)
	cfg.Org.Set("acme", config.SourceFlag)

	p := FromConfig(cfg, "my-repo")
	if p.Owner != "acme" {
		t.Errorf("org should override owner: got %q", p.Owner)
	}
	if p.FullName() != "acme/my-repo" {
		t.Errorf("got fullname=%q", p.FullName())
	}
}
