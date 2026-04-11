package validate

import (
	"testing"

	"github.com/jaeyeom/gh-repox/internal/policy"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		policy  *policy.DesiredPolicy
		wantErr bool
	}{
		{
			name:    "valid",
			policy:  &policy.DesiredPolicy{Owner: "user", Repo: "repo"},
			wantErr: false,
		},
		{
			name:    "missing owner",
			policy:  &policy.DesiredPolicy{Owner: "", Repo: "repo"},
			wantErr: true,
		},
		{
			name:    "missing repo",
			policy:  &policy.DesiredPolicy{Owner: "user", Repo: ""},
			wantErr: true,
		},
		{
			name:    "repo with slash",
			policy:  &policy.DesiredPolicy{Owner: "user", Repo: "owner/repo"},
			wantErr: true,
		},
		{
			name:    "owner with shell metacharacters",
			policy:  &policy.DesiredPolicy{Owner: "bad;touch /tmp/pwn", Repo: "repo"},
			wantErr: true,
		},
		{
			name:    "repo with shell metacharacters",
			policy:  &policy.DesiredPolicy{Owner: "user", Repo: "repo;rm -rf /"},
			wantErr: true,
		},
		{
			name:    "owner starting with hyphen",
			policy:  &policy.DesiredPolicy{Owner: "-badowner", Repo: "repo"},
			wantErr: true,
		},
		{
			name:    "template with auto_init",
			policy:  &policy.DesiredPolicy{Owner: "user", Repo: "repo", Template: "tpl/repo", AutoInit: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Create(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOwner(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		wantErr bool
	}{
		{"valid simple", "octocat", false},
		{"valid with hyphen", "my-org", false},
		{"valid single char", "x", false},
		{"empty", "", true},
		{"starts with hyphen", "-bad", true},
		{"ends with hyphen", "bad-", true},
		{"contains space", "bad owner", true},
		{"contains semicolon", "bad;rm", true},
		{"contains single quote", "bad'owner", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Owner(tt.owner)
			if (err != nil) != tt.wantErr {
				t.Errorf("Owner(%q) error = %v, wantErr %v", tt.owner, err, tt.wantErr)
			}
		})
	}
}

func TestRepo(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{"valid simple", "my-repo", false},
		{"valid with dot", "repo.go", false},
		{"valid with underscore", "my_repo", false},
		{"empty", "", true},
		{"contains slash", "owner/repo", true},
		{"contains semicolon", "repo;rm", true},
		{"contains space", "my repo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Repo(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo(%q) error = %v, wantErr %v", tt.repo, err, tt.wantErr)
			}
		})
	}
}

func TestApply(t *testing.T) {
	if err := Apply("owner", "repo"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := Apply("", "repo"); err == nil {
		t.Error("expected error for empty owner")
	}
	if err := Apply("bad;owner", "repo"); err == nil {
		t.Error("expected error for invalid owner")
	}
	if err := Apply("owner", "bad;repo"); err == nil {
		t.Error("expected error for invalid repo")
	}
}

func TestParseOwnerRepo(t *testing.T) {
	owner, repo, err := ParseOwnerRepo("octocat/my-repo")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "octocat" || repo != "my-repo" {
		t.Errorf("got %s/%s", owner, repo)
	}

	_, _, err = ParseOwnerRepo("invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
	_, _, err = ParseOwnerRepo("/repo")
	if err == nil {
		t.Error("expected error for empty owner")
	}
}
