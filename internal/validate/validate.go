package validate

import (
	"fmt"
	"strings"

	"github.com/jaeyeom/gh-repox/internal/policy"
)

// ValidateCreate validates a desired policy for repository creation.
func ValidateCreate(p *policy.DesiredPolicy) error {
	if p.Owner == "" {
		return fmt.Errorf("owner is required: run `gh auth login`, or pass --owner or --org")
	}
	if p.Repo == "" {
		return fmt.Errorf("repository name is required")
	}
	if strings.Contains(p.Repo, "/") {
		return fmt.Errorf("repository name should not contain '/': got %q", p.Repo)
	}
	if p.Template != "" && (p.AutoInit || p.Gitignore != "" || p.License != "") {
		return fmt.Errorf("cannot use --template with --add-readme, --gitignore, or --license")
	}
	return nil
}

// ValidateApply validates a desired policy for applying to an existing repo.
func ValidateApply(owner, repo string) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner/repo is required (e.g., octocat/my-repo)")
	}
	return nil
}

// ParseOwnerRepo splits an "owner/repo" string.
func ParseOwnerRepo(fullName string) (string, string, error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo format %q: expected owner/repo", fullName)
	}
	return parts[0], parts[1], nil
}
