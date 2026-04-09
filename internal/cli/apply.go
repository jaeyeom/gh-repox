package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaeyeom/gh-repox/internal/config"
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
				return exitErrorf(ExitInvalidInput, "parse repo: %w", err)
			}
			if err := validate.Apply(owner, repo); err != nil {
				return exitErrorf(ExitInvalidInput, "validate apply: %w", err)
			}

			cfg, err := resolveConfig()
			if err != nil {
				return exitErrorf(ExitInvalidInput, "config error: %w", err)
			}

			p := policy.FromConfig(cfg, repo)
			p.Owner = owner
			fullName := p.FullName()

			runner := &exec.RealRunner{}
			client := ghclient.NewClient(runner, cfg.Host.Value)

			// Dry run
			if cfg.DryRun.Value {
				return printApplyDryRun(fullName, p, cfg)
			}

			// Check repo exists
			exists, err := client.RepoExists(ctx, fullName)
			if err != nil {
				return fmt.Errorf("check repo exists: %w", err)
			}
			if !exists {
				return exitErrorf(ExitInvalidInput, "repository %s does not exist", fullName)
			}

			result := &output.ApplyResult{
				Command: "apply",
				Repo:    fullName,
			}

			// Apply settings
			if err := client.EditRepo(ctx, fullName, p); err != nil {
				if cfg.Strict.Value {
					return exitErrorf(ExitStrictFailed, "edit repo: %w", err)
				}
				result.Warnings = append(result.Warnings, err.Error())
			} else {
				visibility := "public"
				if p.Private {
					visibility = "private"
				}
				result.Applied = append(result.Applied,
					"visibility: "+visibility,
					"description: "+p.Description,
					"homepage: "+p.Homepage,
					"issues: "+enabledStr(p.HasIssues),
					"wiki: "+enabledStr(p.HasWiki),
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
				return exitErrorf(ExitStrictFailed, "strict mode: %d warning(s)", len(result.Warnings))
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

func printApplyDryRun(fullName string, p *policy.DesiredPolicy, cfg *config.Config) error {
	editArgs := ghclient.EditRepoArgs(fullName, p)
	editArgs = append(editArgs, ghclient.HostArgs(cfg.Host.Value)...)
	cmds := []string{
		output.FormatCommand("gh", editArgs...),
	}
	cmds = append(cmds, ghclient.PlannedSecurityCommands(fullName, p, cfg.Host.Value)...)
	if flagJSON {
		result := &output.DryRunResult{
			Command:  "apply",
			DryRun:   true,
			Repo:     fullName,
			Commands: cmds,
		}
		if err := output.PrintJSON(os.Stdout, result); err != nil {
			return fmt.Errorf("print JSON: %w", err)
		}
		return nil
	}
	header := fmt.Sprintf("Dry run: gh repox apply %s\n\nWould apply resolved policy to %s", fullName, fullName)
	output.PrintDryRun(os.Stdout, header, cmds)
	return nil
}

func enabledStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
