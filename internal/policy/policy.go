package policy

import (
	"github.com/jaeyeom/gh-repox/internal/config"
)

// DesiredPolicy represents the desired state for a repository.
type DesiredPolicy struct {
	Host  string
	Owner string
	Repo  string

	Private     bool
	Description string
	Homepage    string
	HasIssues   bool
	HasWiki     bool
	HasProjects bool

	AutoInit  bool
	Gitignore string
	License   string
	Template  string

	AllowSquashMerge    bool
	AllowMergeCommit    bool
	AllowRebaseMerge    bool
	AllowAutoMerge      bool
	DeleteBranchOnMerge bool

	DependencyGraph           bool
	DependabotAlerts          bool
	DependabotSecurityUpdates bool

	CloneAfterCreate bool
	CloneDirectory   string
	CloneExtraArgs   []string
}

// ActualState represents the current state of a repository fetched from GitHub.
type ActualState struct {
	Private     bool
	Description string
	Homepage    string
	HasIssues   bool
	HasWiki     bool
	HasProjects bool

	AllowSquashMerge    bool
	AllowMergeCommit    bool
	AllowRebaseMerge    bool
	AllowAutoMerge      *bool
	DeleteBranchOnMerge bool

	DependencyGraph           *bool
	DependabotAlerts          *bool
	DependabotSecurityUpdates *bool
}

// FromConfig builds a DesiredPolicy from a resolved Config.
func FromConfig(cfg *config.Config, repoName string) *DesiredPolicy {
	owner := cfg.Owner.Value
	if cfg.Org.Value != "" {
		owner = cfg.Org.Value
	}

	return &DesiredPolicy{
		Host:  cfg.Host.Value,
		Owner: owner,
		Repo:  repoName,

		Private:     cfg.Private.Value,
		Description: cfg.Description.Value,
		Homepage:    cfg.Homepage.Value,
		HasIssues:   cfg.HasIssues.Value,
		HasWiki:     cfg.HasWiki.Value,
		HasProjects: cfg.HasProjects.Value,

		AutoInit:  cfg.AutoInit.Value,
		Gitignore: cfg.Gitignore.Value,
		License:   cfg.License.Value,
		Template:  cfg.Template.Value,

		AllowSquashMerge:    cfg.AllowSquashMerge.Value,
		AllowMergeCommit:    cfg.AllowMergeCommit.Value,
		AllowRebaseMerge:    cfg.AllowRebaseMerge.Value,
		AllowAutoMerge:      cfg.AllowAutoMerge.Value,
		DeleteBranchOnMerge: cfg.DeleteBranchOnMerge.Value,

		DependencyGraph:           cfg.DependencyGraph.Value,
		DependabotAlerts:          cfg.DependabotAlerts.Value,
		DependabotSecurityUpdates: cfg.DependabotSecurityUpdates.Value,

		CloneAfterCreate: cfg.CloneAfterCreate.Value,
		CloneDirectory:   cfg.CloneDirectory.Value,
		CloneExtraArgs:   cfg.CloneExtraArgs.Value,
	}
}

// FullName returns owner/repo.
func (p *DesiredPolicy) FullName() string {
	return p.Owner + "/" + p.Repo
}
