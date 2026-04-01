package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaeyeom/gh-repox/internal/exec"
	"github.com/jaeyeom/gh-repox/internal/policy"
)

// Client wraps a Runner to execute GitHub CLI commands.
type Client struct {
	Runner exec.Runner
	Host   string
}

// NewClient creates a new GitHub CLI client.
func NewClient(runner exec.Runner, host string) *Client {
	return &Client{Runner: runner, Host: host}
}

func (c *Client) hostArgs() []string {
	if c.Host != "" && c.Host != "github.com" {
		return []string{"--hostname", c.Host}
	}
	return nil
}

// GetAuthenticatedUser returns the current authenticated user login.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	args := []string{"api", "user", "--jq", ".login"}
	args = append(args, c.hostArgs()...)
	stdout, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return "", fmt.Errorf("could not determine authenticated GitHub user: %s: %w", strings.TrimSpace(stderr), err)
	}
	login := strings.TrimSpace(stdout)
	if login == "" {
		return "", fmt.Errorf("could not determine authenticated GitHub user: empty response")
	}
	return login, nil
}

// RepoExists checks if a repository exists.
func (c *Client) RepoExists(ctx context.Context, fullName string) (bool, error) {
	args := []string{"repo", "view", fullName}
	args = append(args, c.hostArgs()...)
	_, _, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CreateRepoArgs returns the arguments for creating a repository.
func CreateRepoArgs(p *policy.DesiredPolicy) []string {
	args := []string{"repo", "create", p.FullName()}

	if p.Private {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	if p.Description != "" {
		args = append(args, "--description", p.Description)
	}
	if p.Homepage != "" {
		args = append(args, "--homepage", p.Homepage)
	}

	if p.Template != "" {
		args = append(args, "--template", p.Template)
	} else if p.AutoInit {
		args = append(args, "--add-readme")
	}

	if p.Gitignore != "" {
		args = append(args, "--gitignore", p.Gitignore)
	}
	if p.License != "" {
		args = append(args, "--license", p.License)
	}

	if !p.HasIssues {
		args = append(args, "--disable-issues")
	}
	if p.HasWiki {
		args = append(args, "--enable-wiki")
	} else {
		args = append(args, "--disable-wiki")
	}

	return args
}

// CreateRepo creates a new repository.
func (c *Client) CreateRepo(ctx context.Context, p *policy.DesiredPolicy) (string, error) {
	args := CreateRepoArgs(p)
	args = append(args, c.hostArgs()...)
	stdout, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return "", fmt.Errorf("failed to create repository: %s: %w", strings.TrimSpace(stderr), err)
	}
	return strings.TrimSpace(stdout), nil
}

// EditRepoArgs returns the arguments for editing repo settings post-creation.
func EditRepoArgs(fullName string, p *policy.DesiredPolicy) []string {
	args := []string{"repo", "edit", fullName}

	if p.AllowSquashMerge {
		args = append(args, "--enable-squash-merge")
	} else {
		args = append(args, "--disable-squash-merge")
	}
	if p.AllowMergeCommit {
		args = append(args, "--enable-merge-commit")
	} else {
		args = append(args, "--disable-merge-commit")
	}
	if p.AllowRebaseMerge {
		args = append(args, "--enable-rebase-merge")
	} else {
		args = append(args, "--disable-rebase-merge")
	}
	if p.AllowAutoMerge {
		args = append(args, "--enable-auto-merge")
	} else {
		args = append(args, "--disable-auto-merge")
	}
	if p.DeleteBranchOnMerge {
		args = append(args, "--delete-branch-on-merge")
	}
	if p.HasProjects {
		args = append(args, "--enable-projects")
	} else {
		args = append(args, "--disable-projects")
	}

	return args
}

// EditRepo applies post-creation settings.
func (c *Client) EditRepo(ctx context.Context, fullName string, p *policy.DesiredPolicy) error {
	args := EditRepoArgs(fullName, p)
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("failed to edit repository: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// EnableVulnerabilityAlerts enables Dependabot vulnerability alerts.
func (c *Client) EnableVulnerabilityAlerts(ctx context.Context, fullName string) error {
	args := []string{"api", "--method", "PUT", fmt.Sprintf("/repos/%s/vulnerability-alerts", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("could not enable Dependabot alerts: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// DisableVulnerabilityAlerts disables Dependabot vulnerability alerts.
func (c *Client) DisableVulnerabilityAlerts(ctx context.Context, fullName string) error {
	args := []string{"api", "--method", "DELETE", fmt.Sprintf("/repos/%s/vulnerability-alerts", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("could not disable Dependabot alerts: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// EnableDependencyGraph enables the dependency graph.
func (c *Client) EnableDependencyGraph(ctx context.Context, fullName string) error {
	args := []string{"api", "--method", "PUT", fmt.Sprintf("/repos/%s/automated-security-fixes", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("could not enable dependency graph: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// CloneRepoArgs returns the arguments for cloning a repository.
func CloneRepoArgs(fullName string, dir string, extraArgs []string) []string {
	args := []string{"repo", "clone", fullName}
	if dir != "" {
		args = append(args, dir)
	}
	if len(extraArgs) > 0 {
		args = append(args, "--")
		args = append(args, extraArgs...)
	}
	return args
}

// CloneRepo clones a repository.
func (c *Client) CloneRepo(ctx context.Context, fullName string, dir string, extraArgs []string) error {
	args := CloneRepoArgs(fullName, dir, extraArgs)
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("clone failed: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// OpenInBrowser opens the repository in the default browser.
func (c *Client) OpenInBrowser(ctx context.Context, fullName string) error {
	args := []string{"repo", "view", fullName, "--web"}
	args = append(args, c.hostArgs()...)
	_, _, err := c.Runner.Run(ctx, "gh", args...)
	return err
}

// repoJSON is the structure returned by gh repo view --json.
type repoJSON struct {
	IsPrivate           bool   `json:"isPrivate"`
	Description         string `json:"description"`
	HomepageURL         string `json:"homepageUrl"`
	HasIssuesEnabled    bool   `json:"hasIssuesEnabled"`
	HasWikiEnabled      bool   `json:"hasWikiEnabled"`
	HasProjectsEnabled  bool   `json:"hasProjectsEnabled"`
	SquashMergeAllowed  bool   `json:"squashMergeAllowed"`
	MergeCommitAllowed  bool   `json:"mergeCommitAllowed"`
	RebaseMergeAllowed  bool   `json:"rebaseMergeAllowed"`
	AutoMergeAllowed    bool   `json:"autoMergeAllowed"`
	DeleteBranchOnMerge bool   `json:"deleteBranchOnMerge"`
}

// FetchRepoState fetches the current repository state.
func (c *Client) FetchRepoState(ctx context.Context, fullName string) (*policy.ActualState, error) {
	fields := "isPrivate,description,homepageUrl,hasIssuesEnabled,hasWikiEnabled,hasProjectsEnabled,squashMergeAllowed,mergeCommitAllowed,rebaseMergeAllowed,autoMergeAllowed,deleteBranchOnMerge"
	args := []string{"repo", "view", fullName, "--json", fields}
	args = append(args, c.hostArgs()...)
	stdout, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo state: %s: %w", strings.TrimSpace(stderr), err)
	}

	var rj repoJSON
	if err := json.Unmarshal([]byte(stdout), &rj); err != nil {
		return nil, fmt.Errorf("failed to parse repo state: %w", err)
	}

	return &policy.ActualState{
		Private:             rj.IsPrivate,
		Description:         rj.Description,
		Homepage:            rj.HomepageURL,
		HasIssues:           rj.HasIssuesEnabled,
		HasWiki:             rj.HasWikiEnabled,
		HasProjects:         rj.HasProjectsEnabled,
		AllowSquashMerge:    rj.SquashMergeAllowed,
		AllowMergeCommit:    rj.MergeCommitAllowed,
		AllowRebaseMerge:    rj.RebaseMergeAllowed,
		AllowAutoMerge:      rj.AutoMergeAllowed,
		DeleteBranchOnMerge: rj.DeleteBranchOnMerge,
	}, nil
}

// ApplySecuritySettings applies security-related settings.
func (c *Client) ApplySecuritySettings(ctx context.Context, fullName string, p *policy.DesiredPolicy, strict bool) (applied []string, warnings []string) {
	if p.DependabotAlerts {
		if err := c.EnableVulnerabilityAlerts(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "Dependabot alerts enabled")
		}
	}
	return applied, warnings
}

// PlannedCommands returns the list of commands that would be executed.
func PlannedCommands(p *policy.DesiredPolicy) []string {
	var cmds []string
	cmds = append(cmds, "gh "+strings.Join(CreateRepoArgs(p), " "))
	cmds = append(cmds, "gh "+strings.Join(EditRepoArgs(p.FullName(), p), " "))
	if p.DependabotAlerts {
		cmds = append(cmds, fmt.Sprintf("gh api --method PUT /repos/%s/vulnerability-alerts", p.FullName()))
	}
	if p.CloneAfterCreate {
		cmds = append(cmds, "gh "+strings.Join(CloneRepoArgs(p.FullName(), p.CloneDirectory, p.CloneExtraArgs), " "))
	}
	return cmds
}
