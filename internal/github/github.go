package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jaeyeom/gh-repox/internal/exec"
	"github.com/jaeyeom/gh-repox/internal/output"
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

// HostArgs returns the --hostname flag pair for a non-default GitHub host.
func HostArgs(host string) []string {
	if host != "" && host != "github.com" {
		return []string{"--hostname", host}
	}
	return nil
}

func (c *Client) hostArgs() []string {
	return HostArgs(c.Host)
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
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		lower := strings.ToLower(strings.TrimSpace(stderr))
		if strings.Contains(lower, "not found") || strings.Contains(lower, "could not resolve") {
			return false, nil
		}
		return false, fmt.Errorf("check repo exists: %s: %w", strings.TrimSpace(stderr), err)
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

	// Visibility
	if p.Private {
		args = append(args, "--visibility", "private")
	} else {
		args = append(args, "--visibility", "public")
	}
	args = append(args, "--accept-visibility-change-consequences")

	args = append(args, "--description", p.Description)
	args = append(args, "--homepage", p.Homepage)

	// Features — gh repo edit only has --enable-* flags; disable with =false.
	args = append(args, fmt.Sprintf("--enable-issues=%t", p.HasIssues))
	args = append(args, fmt.Sprintf("--enable-wiki=%t", p.HasWiki))

	// Merge settings
	args = append(args, fmt.Sprintf("--enable-squash-merge=%t", p.AllowSquashMerge))
	args = append(args, fmt.Sprintf("--enable-merge-commit=%t", p.AllowMergeCommit))
	args = append(args, fmt.Sprintf("--enable-rebase-merge=%t", p.AllowRebaseMerge))
	args = append(args, fmt.Sprintf("--enable-auto-merge=%t", p.AllowAutoMerge))
	args = append(args, fmt.Sprintf("--delete-branch-on-merge=%t", p.DeleteBranchOnMerge))
	args = append(args, fmt.Sprintf("--enable-projects=%t", p.HasProjects))

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

// patchSecurityAnalysis sends a PATCH to /repos/{repo} with a security_and_analysis body.
func (c *Client) patchSecurityAnalysis(ctx context.Context, fullName string, body string) error {
	tmpFile, err := os.CreateTemp("", "gh-repox-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	args := []string{"api", "--method", "PATCH", fmt.Sprintf("/repos/%s", fullName), "--input", tmpFile.Name()}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("PATCH security_and_analysis: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// EnableDependencyGraph enables the dependency graph via the security_and_analysis API.
func (c *Client) EnableDependencyGraph(ctx context.Context, fullName string) error {
	body := `{"security_and_analysis":{"dependency_graph":{"status":"enabled"}}}`
	if err := c.patchSecurityAnalysis(ctx, fullName, body); err != nil {
		return fmt.Errorf("could not enable dependency graph: %w", err)
	}
	return nil
}

// DisableDependencyGraph disables the dependency graph via the security_and_analysis API.
func (c *Client) DisableDependencyGraph(ctx context.Context, fullName string) error {
	body := `{"security_and_analysis":{"dependency_graph":{"status":"disabled"}}}`
	if err := c.patchSecurityAnalysis(ctx, fullName, body); err != nil {
		return fmt.Errorf("could not disable dependency graph: %w", err)
	}
	return nil
}

// EnableAutomatedSecurityFixes enables Dependabot security updates.
func (c *Client) EnableAutomatedSecurityFixes(ctx context.Context, fullName string) error {
	args := []string{"api", "--method", "PUT", fmt.Sprintf("/repos/%s/automated-security-fixes", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("could not enable Dependabot security updates: %s: %w", strings.TrimSpace(stderr), err)
	}
	return nil
}

// DisableAutomatedSecurityFixes disables Dependabot security updates.
func (c *Client) DisableAutomatedSecurityFixes(ctx context.Context, fullName string) error {
	args := []string{"api", "--method", "DELETE", fmt.Sprintf("/repos/%s/automated-security-fixes", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return fmt.Errorf("could not disable Dependabot security updates: %s: %w", strings.TrimSpace(stderr), err)
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
	args := []string{"repo", "clone", fullName}
	if dir != "" {
		args = append(args, dir)
	}
	args = append(args, c.hostArgs()...)
	if len(extraArgs) > 0 {
		args = append(args, "--")
		args = append(args, extraArgs...)
	}
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
	if err != nil {
		return fmt.Errorf("opening repository in browser: %w", err)
	}
	return nil
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
	DeleteBranchOnMerge bool   `json:"deleteBranchOnMerge"`
}

// FetchRepoState fetches the current repository state.
func (c *Client) FetchRepoState(ctx context.Context, fullName string) (*policy.ActualState, error) {
	fields := "isPrivate,description,homepageUrl,hasIssuesEnabled,hasWikiEnabled,hasProjectsEnabled,squashMergeAllowed,mergeCommitAllowed,rebaseMergeAllowed,deleteBranchOnMerge"
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

	state := &policy.ActualState{
		Private:             rj.IsPrivate,
		Description:         rj.Description,
		Homepage:            rj.HomepageURL,
		HasIssues:           rj.HasIssuesEnabled,
		HasWiki:             rj.HasWikiEnabled,
		HasProjects:         rj.HasProjectsEnabled,
		AllowSquashMerge:    rj.SquashMergeAllowed,
		AllowMergeCommit:    rj.MergeCommitAllowed,
		AllowRebaseMerge:    rj.RebaseMergeAllowed,
		AllowAutoMerge:      c.fetchAllowAutoMerge(ctx, fullName),
		DeleteBranchOnMerge: rj.DeleteBranchOnMerge,
	}

	// Fetch security settings via REST API (not available through gh repo view --json).
	state.DependabotAlerts = c.fetchVulnerabilityAlertsEnabled(ctx, fullName)
	state.DependencyGraph = c.fetchDependencyGraphEnabled(ctx, fullName)
	state.DependabotSecurityUpdates = c.fetchAutomatedSecurityFixesEnabled(ctx, fullName)

	return state, nil
}

// fetchAllowAutoMerge checks if auto-merge is allowed via the REST API.
// The autoMergeAllowed field is not available through gh repo view --json,
// so we fetch it from the REST API's allow_auto_merge field.
// Returns nil on error to indicate the value is unknown.
func (c *Client) fetchAllowAutoMerge(ctx context.Context, fullName string) *bool {
	args := []string{"api", fmt.Sprintf("/repos/%s", fullName), "--jq", ".allow_auto_merge"}
	args = append(args, c.hostArgs()...)
	stdout, _, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return nil
	}
	enabled := strings.TrimSpace(stdout) == "true"
	return &enabled
}

// fetchVulnerabilityAlertsEnabled checks if Dependabot alerts are enabled.
// The /vulnerability-alerts endpoint returns 204 if enabled, 404 if disabled.
// On other errors (permissions, network), returns nil to indicate unknown.
func (c *Client) fetchVulnerabilityAlertsEnabled(ctx context.Context, fullName string) *bool {
	args := []string{"api", fmt.Sprintf("/repos/%s/vulnerability-alerts", fullName)}
	args = append(args, c.hostArgs()...)
	_, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		// 404 means alerts are disabled; other errors are unknown.
		if strings.Contains(strings.ToLower(strings.TrimSpace(stderr)), "not found") {
			enabled := false
			return &enabled
		}
		return nil
	}
	enabled := true
	return &enabled
}

// fetchDependencyGraphEnabled checks if the dependency graph is enabled.
func (c *Client) fetchDependencyGraphEnabled(ctx context.Context, fullName string) *bool {
	args := []string{"api", fmt.Sprintf("/repos/%s", fullName), "--jq", ".security_and_analysis.dependency_graph.status"}
	args = append(args, c.hostArgs()...)
	stdout, _, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		return nil
	}
	status := strings.TrimSpace(stdout)
	if status == "" {
		return nil
	}
	enabled := status == "enabled"
	return &enabled
}

// fetchAutomatedSecurityFixesEnabled checks if Dependabot security updates are enabled.
// The endpoint returns 200 with JSON when accessible, 404 when not enabled.
// On other errors (permissions, network), returns nil to indicate unknown.
func (c *Client) fetchAutomatedSecurityFixesEnabled(ctx context.Context, fullName string) *bool {
	args := []string{"api", fmt.Sprintf("/repos/%s/automated-security-fixes", fullName)}
	args = append(args, c.hostArgs()...)
	stdout, stderr, err := c.Runner.Run(ctx, "gh", args...)
	if err != nil {
		// 404 means security fixes are not enabled; other errors are unknown.
		if strings.Contains(strings.ToLower(strings.TrimSpace(stderr)), "not found") {
			enabled := false
			return &enabled
		}
		return nil
	}
	var result struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil
	}
	return &result.Enabled
}

// ApplySecuritySettings applies security-related settings.
func (c *Client) ApplySecuritySettings(ctx context.Context, fullName string, p *policy.DesiredPolicy, _ bool) (applied []string, warnings []string) {
	// Dependency graph
	if p.DependencyGraph {
		if err := c.EnableDependencyGraph(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "dependency graph enabled")
		}
	} else {
		if err := c.DisableDependencyGraph(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "dependency graph disabled")
		}
	}

	// Dependabot alerts
	if p.DependabotAlerts {
		if err := c.EnableVulnerabilityAlerts(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "Dependabot alerts enabled")
		}
	} else {
		if err := c.DisableVulnerabilityAlerts(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "Dependabot alerts disabled")
		}
	}

	// Dependabot security updates
	if p.DependabotSecurityUpdates {
		if err := c.EnableAutomatedSecurityFixes(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "Dependabot security updates enabled")
		}
	} else {
		if err := c.DisableAutomatedSecurityFixes(ctx, fullName); err != nil {
			warnings = append(warnings, err.Error())
		} else {
			applied = append(applied, "Dependabot security updates disabled")
		}
	}

	return applied, warnings
}

// PlannedSecurityCommands returns the security-related commands that would be executed.
func PlannedSecurityCommands(fullName string, p *policy.DesiredPolicy, host string) []string {
	ha := HostArgs(host)
	hostSuffix := ""
	for _, a := range ha {
		hostSuffix += " " + output.ShellQuote(a)
	}
	qn := output.ShellQuote("/repos/" + fullName)
	var cmds []string
	if p.DependencyGraph {
		cmds = append(cmds, fmt.Sprintf(`echo '{"security_and_analysis":{"dependency_graph":{"status":"enabled"}}}' | gh api --method PATCH %s --input -%s`, qn, hostSuffix))
	} else {
		cmds = append(cmds, fmt.Sprintf(`echo '{"security_and_analysis":{"dependency_graph":{"status":"disabled"}}}' | gh api --method PATCH %s --input -%s`, qn, hostSuffix))
	}
	if p.DependabotAlerts {
		cmds = append(cmds, fmt.Sprintf("gh api --method PUT %s/vulnerability-alerts%s", qn, hostSuffix))
	} else {
		cmds = append(cmds, fmt.Sprintf("gh api --method DELETE %s/vulnerability-alerts%s", qn, hostSuffix))
	}
	if p.DependabotSecurityUpdates {
		cmds = append(cmds, fmt.Sprintf("gh api --method PUT %s/automated-security-fixes%s", qn, hostSuffix))
	} else {
		cmds = append(cmds, fmt.Sprintf("gh api --method DELETE %s/automated-security-fixes%s", qn, hostSuffix))
	}
	return cmds
}

// PlannedCommands returns the list of commands that would be executed.
func PlannedCommands(p *policy.DesiredPolicy, host string) []string {
	ha := HostArgs(host)
	var cmds []string

	createArgs := CreateRepoArgs(p)
	createArgs = append(createArgs, ha...)
	cmds = append(cmds, output.FormatCommand("gh", createArgs...))

	editArgs := EditRepoArgs(p.FullName(), p)
	editArgs = append(editArgs, ha...)
	cmds = append(cmds, output.FormatCommand("gh", editArgs...))

	cmds = append(cmds, PlannedSecurityCommands(p.FullName(), p, host)...)

	if p.CloneAfterCreate {
		cloneArgs := []string{"repo", "clone", p.FullName()}
		if p.CloneDirectory != "" {
			cloneArgs = append(cloneArgs, p.CloneDirectory)
		}
		cloneArgs = append(cloneArgs, ha...)
		if len(p.CloneExtraArgs) > 0 {
			cloneArgs = append(cloneArgs, "--")
			cloneArgs = append(cloneArgs, p.CloneExtraArgs...)
		}
		cmds = append(cmds, output.FormatCommand("gh", cloneArgs...))
	}

	return cmds
}
