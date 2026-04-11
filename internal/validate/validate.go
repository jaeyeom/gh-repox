package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jaeyeom/gh-repox/internal/policy"
)

// ownerRe matches valid GitHub usernames and org names: alphanumeric and
// single hyphens, not starting or ending with a hyphen.
var ownerRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

// repoRe matches valid GitHub repository names: alphanumeric, hyphens,
// underscores, and dots.
var repoRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Owner validates a GitHub owner (user or organization) name.
func Owner(owner string) error {
	if owner == "" {
		return fmt.Errorf("owner is required: run `gh auth login`, or pass --owner or --org")
	}
	if !ownerRe.MatchString(owner) {
		return fmt.Errorf("invalid owner %q: must contain only alphanumeric characters or hyphens, and cannot start or end with a hyphen", owner)
	}
	return nil
}

// Repo validates a GitHub repository name.
func Repo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}
	if !repoRe.MatchString(repo) {
		return fmt.Errorf("invalid repository name %q: must contain only alphanumeric characters, hyphens, underscores, or dots", repo)
	}
	return nil
}

// Create validates a desired policy for repository creation.
func Create(p *policy.DesiredPolicy) error {
	if err := Owner(p.Owner); err != nil {
		return err
	}
	if err := Repo(p.Repo); err != nil {
		return err
	}
	if p.Template != "" && (p.AutoInit || p.Gitignore != "" || p.License != "") {
		return fmt.Errorf("cannot use --template with --add-readme, --gitignore, or --license")
	}
	return nil
}

// Apply validates a desired policy for applying to an existing repo.
func Apply(owner, repo string) error {
	if err := Owner(owner); err != nil {
		return err
	}
	if err := Repo(repo); err != nil {
		return err
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
