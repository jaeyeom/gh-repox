package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Source indicates where a config value came from.
type Source string

const (
	SourceDefault  Source = "default"
	SourceConfig   Source = "config"
	SourceEnv      Source = "env"
	SourceFlag     Source = "flag"
	SourceInferred Source = "inferred"
)

// Field holds a value and its source.
type Field[T any] struct {
	Value  T
	Source Source
}

// Set updates the field value and source.
func (f *Field[T]) Set(val T, src Source) {
	f.Value = val
	f.Source = src
}

// Config holds all resolved configuration with source tracking.
type Config struct {
	// GitHub settings
	Host  Field[string]
	Owner Field[string]
	Org   Field[string]

	// Repo settings
	Private     Field[bool]
	Description Field[string]
	Homepage    Field[string]
	HasIssues   Field[bool]
	HasWiki     Field[bool]
	HasProjects Field[bool]

	// Init settings
	AutoInit  Field[bool]
	Gitignore Field[string]
	License   Field[string]
	Template  Field[string]

	// Merge settings
	AllowSquashMerge    Field[bool]
	AllowMergeCommit    Field[bool]
	AllowRebaseMerge    Field[bool]
	AllowAutoMerge      Field[bool]
	DeleteBranchOnMerge Field[bool]

	// Security settings
	DependencyGraph           Field[bool]
	DependabotAlerts          Field[bool]
	DependabotSecurityUpdates Field[bool]

	// Clone settings
	CloneAfterCreate Field[bool]
	CloneDirectory   Field[string]
	CloneExtraArgs   Field[[]string]

	// Behavior settings
	DryRun   Field[bool]
	Verbose  Field[bool]
	Strict   Field[bool]
	OpenRepo Field[bool]
}

// yamlConfig is the YAML file structure.
type yamlConfig struct {
	GitHub struct {
		Host  *string `yaml:"host"`
		Owner *string `yaml:"owner"`
		Org   *string `yaml:"org"`
	} `yaml:"github"`
	Repo struct {
		Private     *bool   `yaml:"private"`
		Description *string `yaml:"description"`
		Homepage    *string `yaml:"homepage"`
		HasIssues   *bool   `yaml:"has_issues"`
		HasWiki     *bool   `yaml:"has_wiki"`
		HasProjects *bool   `yaml:"has_projects"`
		AutoInit    *bool   `yaml:"auto_init"`
		Gitignore   *string `yaml:"gitignore"`
		License     *string `yaml:"license"`
		Template    *string `yaml:"template"`
	} `yaml:"repo"`
	Merge struct {
		AllowSquashMerge    *bool `yaml:"allow_squash_merge"`
		AllowMergeCommit    *bool `yaml:"allow_merge_commit"`
		AllowRebaseMerge    *bool `yaml:"allow_rebase_merge"`
		AllowAutoMerge      *bool `yaml:"allow_auto_merge"`
		DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge"`
	} `yaml:"merge"`
	Security struct {
		DependencyGraph           *bool `yaml:"dependency_graph"`
		DependabotAlerts          *bool `yaml:"dependabot_alerts"`
		DependabotSecurityUpdates *bool `yaml:"dependabot_security_updates"`
	} `yaml:"security"`
	Clone struct {
		AfterCreate *bool    `yaml:"after_create"`
		Directory   *string  `yaml:"directory"`
		ExtraArgs   []string `yaml:"extra_args"`
	} `yaml:"clone"`
	Behavior struct {
		DryRun   *bool `yaml:"dry_run"`
		Verbose  *bool `yaml:"verbose"`
		Strict   *bool `yaml:"strict"`
		OpenRepo *bool `yaml:"open_repo"`
	} `yaml:"behavior"`
}

// Defaults returns a Config with all built-in defaults.
func Defaults() *Config {
	c := &Config{}
	c.Host.Set("github.com", SourceDefault)
	c.Owner.Set("", SourceDefault)
	c.Org.Set("", SourceDefault)

	c.Private.Set(true, SourceDefault)
	c.Description.Set("", SourceDefault)
	c.Homepage.Set("", SourceDefault)
	c.HasIssues.Set(true, SourceDefault)
	c.HasWiki.Set(false, SourceDefault)
	c.HasProjects.Set(false, SourceDefault)

	c.AutoInit.Set(false, SourceDefault)
	c.Gitignore.Set("", SourceDefault)
	c.License.Set("", SourceDefault)
	c.Template.Set("", SourceDefault)

	c.AllowSquashMerge.Set(true, SourceDefault)
	c.AllowMergeCommit.Set(false, SourceDefault)
	c.AllowRebaseMerge.Set(false, SourceDefault)
	c.AllowAutoMerge.Set(true, SourceDefault)
	c.DeleteBranchOnMerge.Set(true, SourceDefault)

	c.DependencyGraph.Set(true, SourceDefault)
	c.DependabotAlerts.Set(true, SourceDefault)
	c.DependabotSecurityUpdates.Set(false, SourceDefault)

	c.CloneAfterCreate.Set(false, SourceDefault)
	c.CloneDirectory.Set("", SourceDefault)
	c.CloneExtraArgs.Set(nil, SourceDefault)

	c.DryRun.Set(false, SourceDefault)
	c.Verbose.Set(false, SourceDefault)
	c.Strict.Set(false, SourceDefault)
	c.OpenRepo.Set(false, SourceDefault)

	return c
}

// FindConfigFile discovers the config file path.
func FindConfigFile(explicit string) string {
	if explicit != "" {
		return explicit
	}
	candidates := []string{".repox.yaml", "repox.yaml"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".config", "gh-repox", "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// LoadFile loads a YAML config file and merges it into the Config.
func (c *Config) LoadFile(path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return err
	}
	if yc.GitHub.Host != nil {
		c.Host.Set(*yc.GitHub.Host, SourceConfig)
	}
	if yc.GitHub.Owner != nil {
		c.Owner.Set(*yc.GitHub.Owner, SourceConfig)
	}
	if yc.GitHub.Org != nil {
		c.Org.Set(*yc.GitHub.Org, SourceConfig)
	}
	if yc.Repo.Private != nil {
		c.Private.Set(*yc.Repo.Private, SourceConfig)
	}
	if yc.Repo.Description != nil {
		c.Description.Set(*yc.Repo.Description, SourceConfig)
	}
	if yc.Repo.Homepage != nil {
		c.Homepage.Set(*yc.Repo.Homepage, SourceConfig)
	}
	if yc.Repo.HasIssues != nil {
		c.HasIssues.Set(*yc.Repo.HasIssues, SourceConfig)
	}
	if yc.Repo.HasWiki != nil {
		c.HasWiki.Set(*yc.Repo.HasWiki, SourceConfig)
	}
	if yc.Repo.HasProjects != nil {
		c.HasProjects.Set(*yc.Repo.HasProjects, SourceConfig)
	}
	if yc.Repo.AutoInit != nil {
		c.AutoInit.Set(*yc.Repo.AutoInit, SourceConfig)
	}
	if yc.Repo.Gitignore != nil {
		c.Gitignore.Set(*yc.Repo.Gitignore, SourceConfig)
	}
	if yc.Repo.License != nil {
		c.License.Set(*yc.Repo.License, SourceConfig)
	}
	if yc.Repo.Template != nil {
		c.Template.Set(*yc.Repo.Template, SourceConfig)
	}
	if yc.Merge.AllowSquashMerge != nil {
		c.AllowSquashMerge.Set(*yc.Merge.AllowSquashMerge, SourceConfig)
	}
	if yc.Merge.AllowMergeCommit != nil {
		c.AllowMergeCommit.Set(*yc.Merge.AllowMergeCommit, SourceConfig)
	}
	if yc.Merge.AllowRebaseMerge != nil {
		c.AllowRebaseMerge.Set(*yc.Merge.AllowRebaseMerge, SourceConfig)
	}
	if yc.Merge.AllowAutoMerge != nil {
		c.AllowAutoMerge.Set(*yc.Merge.AllowAutoMerge, SourceConfig)
	}
	if yc.Merge.DeleteBranchOnMerge != nil {
		c.DeleteBranchOnMerge.Set(*yc.Merge.DeleteBranchOnMerge, SourceConfig)
	}
	if yc.Security.DependencyGraph != nil {
		c.DependencyGraph.Set(*yc.Security.DependencyGraph, SourceConfig)
	}
	if yc.Security.DependabotAlerts != nil {
		c.DependabotAlerts.Set(*yc.Security.DependabotAlerts, SourceConfig)
	}
	if yc.Security.DependabotSecurityUpdates != nil {
		c.DependabotSecurityUpdates.Set(*yc.Security.DependabotSecurityUpdates, SourceConfig)
	}
	if yc.Clone.AfterCreate != nil {
		c.CloneAfterCreate.Set(*yc.Clone.AfterCreate, SourceConfig)
	}
	if yc.Clone.Directory != nil {
		c.CloneDirectory.Set(*yc.Clone.Directory, SourceConfig)
	}
	if yc.Clone.ExtraArgs != nil {
		c.CloneExtraArgs.Set(yc.Clone.ExtraArgs, SourceConfig)
	}
	if yc.Behavior.DryRun != nil {
		c.DryRun.Set(*yc.Behavior.DryRun, SourceConfig)
	}
	if yc.Behavior.Verbose != nil {
		c.Verbose.Set(*yc.Behavior.Verbose, SourceConfig)
	}
	if yc.Behavior.Strict != nil {
		c.Strict.Set(*yc.Behavior.Strict, SourceConfig)
	}
	if yc.Behavior.OpenRepo != nil {
		c.OpenRepo.Set(*yc.Behavior.OpenRepo, SourceConfig)
	}
	return nil
}

// LoadEnv applies environment variable overrides.
func (c *Config) LoadEnv() {
	if v := os.Getenv("REPOX_HOST"); v != "" {
		c.Host.Set(v, SourceEnv)
	}
	if v := os.Getenv("REPOX_OWNER"); v != "" {
		c.Owner.Set(v, SourceEnv)
	}
	if v := os.Getenv("REPOX_ORG"); v != "" {
		c.Org.Set(v, SourceEnv)
	}
	if v := os.Getenv("REPOX_PRIVATE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Private.Set(b, SourceEnv)
		}
	}
	if v := os.Getenv("REPOX_VERBOSE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Verbose.Set(b, SourceEnv)
		}
	}
	if v := os.Getenv("REPOX_DRY_RUN"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.DryRun.Set(b, SourceEnv)
		}
	}
	if v := os.Getenv("REPOX_STRICT"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Strict.Set(b, SourceEnv)
		}
	}
	if v := os.Getenv("REPOX_CLONE_AFTER_CREATE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.CloneAfterCreate.Set(b, SourceEnv)
		}
	}
}

// Entries returns all config fields as key-value-source triples for display.
func (c *Config) Entries() []Entry {
	return []Entry{
		{"host", c.Host.Value, c.Host.Source},
		{"owner", c.Owner.Value, c.Owner.Source},
		{"org", c.Org.Value, c.Org.Source},
		{"private", c.Private.Value, c.Private.Source},
		{"description", c.Description.Value, c.Description.Source},
		{"homepage", c.Homepage.Value, c.Homepage.Source},
		{"has_issues", c.HasIssues.Value, c.HasIssues.Source},
		{"has_wiki", c.HasWiki.Value, c.HasWiki.Source},
		{"has_projects", c.HasProjects.Value, c.HasProjects.Source},
		{"auto_init", c.AutoInit.Value, c.AutoInit.Source},
		{"gitignore", c.Gitignore.Value, c.Gitignore.Source},
		{"license", c.License.Value, c.License.Source},
		{"template", c.Template.Value, c.Template.Source},
		{"allow_squash_merge", c.AllowSquashMerge.Value, c.AllowSquashMerge.Source},
		{"allow_merge_commit", c.AllowMergeCommit.Value, c.AllowMergeCommit.Source},
		{"allow_rebase_merge", c.AllowRebaseMerge.Value, c.AllowRebaseMerge.Source},
		{"allow_auto_merge", c.AllowAutoMerge.Value, c.AllowAutoMerge.Source},
		{"delete_branch_on_merge", c.DeleteBranchOnMerge.Value, c.DeleteBranchOnMerge.Source},
		{"dependency_graph", c.DependencyGraph.Value, c.DependencyGraph.Source},
		{"dependabot_alerts", c.DependabotAlerts.Value, c.DependabotAlerts.Source},
		{"dependabot_security_updates", c.DependabotSecurityUpdates.Value, c.DependabotSecurityUpdates.Source},
		{"clone_after_create", c.CloneAfterCreate.Value, c.CloneAfterCreate.Source},
		{"clone_directory", c.CloneDirectory.Value, c.CloneDirectory.Source},
		{"clone_extra_args", c.CloneExtraArgs.Value, c.CloneExtraArgs.Source},
		{"dry_run", c.DryRun.Value, c.DryRun.Source},
		{"verbose", c.Verbose.Value, c.Verbose.Source},
		{"strict", c.Strict.Value, c.Strict.Source},
		{"open_repo", c.OpenRepo.Value, c.OpenRepo.Source},
	}
}

// Entry is a single config entry for display.
type Entry struct {
	Key    string
	Value  any
	Source Source
}
