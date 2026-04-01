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

func TestApply(t *testing.T) {
	if err := Apply("owner", "repo"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := Apply("", "repo"); err == nil {
		t.Error("expected error for empty owner")
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
