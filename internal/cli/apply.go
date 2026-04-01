package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaeyeom/gh-repox/internal/exec"
	ghclient "github.com/jaeyeom/gh-repox/internal/github"
	"github.com/jaeyeom/gh-repox/internal/output"
	"github.com/jaeyeom/gh-repox/internal/policy"
	"github.com/jaeyeom/gh-repox/internal/validate"
)

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <owner/repo>",
		Short: "Apply resolved policy to an existing repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()

			owner, repo, err := validate.ParseOwnerRepo(args[0])
			if err != nil {
				return fmt.Errorf("parse repo: %w", err)
			}
			if err := validate.Apply(owner, repo); err != nil {
				return fmt.Errorf("validate apply: %w", err)
			}

			cfg, err := resolveConfig()
			if err != nil {
				return fmt.Errorf("config error: %w", err)
			}

			p := policy.FromConfig(cfg, repo)
			p.Owner = owner
			fullName := p.FullName()

			runner := &exec.RealRunner{}
			client := ghclient.NewClient(runner, cfg.Host.Value)

			// Dry run
			if cfg.DryRun.Value {
				header := fmt.Sprintf("Dry run: gh repox apply %s\n\nWould apply resolved policy to %s", fullName, fullName)
				cmds := []string{
					output.FormatCommand("gh", ghclient.EditRepoArgs(fullName, p)...),
				}
				if p.DependabotAlerts {
					cmds = append(cmds, fmt.Sprintf("gh api --method PUT /repos/%s/vulnerability-alerts", fullName))
				}
				output.PrintDryRun(os.Stdout, header, cmds)
				return nil
			}

			// Check repo exists
			exists, err := client.RepoExists(ctx, fullName)
			if err != nil {
				return fmt.Errorf("check repo exists: %w", err)
			}
			if !exists {
				return fmt.Errorf("repository %s does not exist", fullName)
			}

			result := &output.ApplyResult{
				Command: "apply",
				Repo:    fullName,
			}

			// Apply settings
			if err := client.EditRepo(ctx, fullName, p); err != nil {
				if cfg.Strict.Value {
					return fmt.Errorf("edit repo: %w", err)
				}
				result.Warnings = append(result.Warnings, err.Error())
			} else {
				result.Applied = append(result.Applied,
					"squash merge: "+enabledStr(p.AllowSquashMerge),
					"merge commits: "+enabledStr(p.AllowMergeCommit),
					"rebase merge: "+enabledStr(p.AllowRebaseMerge),
					"auto-merge: "+enabledStr(p.AllowAutoMerge),
					"delete branch on merge: "+enabledStr(p.DeleteBranchOnMerge),
					"projects: "+enabledStr(p.HasProjects),
				)
			}

			// Security settings
			secApplied, secWarnings := client.ApplySecuritySettings(ctx, fullName, p, cfg.Strict.Value)
			result.Applied = append(result.Applied, secApplied...)
			result.Warnings = append(result.Warnings, secWarnings...)

			if cfg.Strict.Value && len(result.Warnings) > 0 {
				if flagJSON {
					_ = output.PrintJSON(os.Stdout, result)
				} else {
					output.PrintApplyHuman(os.Stdout, result)
				}
				return fmt.Errorf("strict mode: %d warning(s)", len(result.Warnings))
			}

			if flagJSON {
				return output.PrintJSON(os.Stdout, result)
			}
			output.PrintApplyHuman(os.Stdout, result)
			return nil
		},
	}
	return cmd
}

func enabledStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
