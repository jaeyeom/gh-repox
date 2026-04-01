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

func newCreateCmd() *cobra.Command {
	var (
		flagPrivate         bool
		flagPublic          bool
		flagDescription     string
		flagHomepage        string
		flagEnableIssues    bool
		flagDisableIssues   bool
		flagEnableWiki      bool
		flagDisableWiki     bool
		flagEnableProjects  bool
		flagDisableProjects bool
		flagAddReadme       bool
		flagGitignore       string
		flagLicense         string
		flagTemplate        string
		flagClone           bool
		flagCloneDir        string
		flagCloneArgs       []string
		flagOpen            bool
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new repository with opinionated defaults",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			repoName := args[0]

			cfg, err := resolveConfig()
			if err != nil {
				return fmt.Errorf("config error: %w", err)
			}

			// Apply create-specific flags
			if flagPrivate {
				cfg.Private.Set(true, config.SourceFlag)
			}
			if flagPublic {
				cfg.Private.Set(false, config.SourceFlag)
			}
			if flagDescription != "" {
				cfg.Description.Set(flagDescription, config.SourceFlag)
			}
			if flagHomepage != "" {
				cfg.Homepage.Set(flagHomepage, config.SourceFlag)
			}
			if flagEnableIssues {
				cfg.HasIssues.Set(true, config.SourceFlag)
			}
			if flagDisableIssues {
				cfg.HasIssues.Set(false, config.SourceFlag)
			}
			if flagEnableWiki {
				cfg.HasWiki.Set(true, config.SourceFlag)
			}
			if flagDisableWiki {
				cfg.HasWiki.Set(false, config.SourceFlag)
			}
			if flagEnableProjects {
				cfg.HasProjects.Set(true, config.SourceFlag)
			}
			if flagDisableProjects {
				cfg.HasProjects.Set(false, config.SourceFlag)
			}
			if flagAddReadme {
				cfg.AutoInit.Set(true, config.SourceFlag)
			}
			if flagGitignore != "" {
				cfg.Gitignore.Set(flagGitignore, config.SourceFlag)
			}
			if flagLicense != "" {
				cfg.License.Set(flagLicense, config.SourceFlag)
			}
			if flagTemplate != "" {
				cfg.Template.Set(flagTemplate, config.SourceFlag)
			}
			if flagClone {
				cfg.CloneAfterCreate.Set(true, config.SourceFlag)
			}
			if flagCloneDir != "" {
				cfg.CloneDirectory.Set(flagCloneDir, config.SourceFlag)
			}
			if len(flagCloneArgs) > 0 {
				cfg.CloneExtraArgs.Set(flagCloneArgs, config.SourceFlag)
			}
			if flagOpen {
				cfg.OpenRepo.Set(true, config.SourceFlag)
			}

			// Resolve owner
			if err := resolveOwner(ctx, cfg); err != nil {
				return err
			}

			// Build policy
			p := policy.FromConfig(cfg, repoName)

			// Validate
			if err := validate.ValidateCreate(p); err != nil {
				return err
			}

			runner := &exec.RealRunner{}
			client := ghclient.NewClient(runner, cfg.Host.Value)

			// Dry run
			if cfg.DryRun.Value {
				header := fmt.Sprintf("Dry run: gh repox create %s\n\nResolved target:\n- repo: %s\n- owner source: %s\n- visibility: %s\n- init: %s\n- clone after create: %v",
					repoName,
					p.FullName(),
					cfg.Owner.Source,
					visibilityStr(p.Private),
					initStr(p),
					p.CloneAfterCreate,
				)
				cmds := ghclient.PlannedCommands(p)
				output.PrintDryRun(os.Stdout, header, cmds)
				return nil
			}

			// Check if repo exists
			exists, err := client.RepoExists(ctx, p.FullName())
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("repository %s already exists", p.FullName())
			}

			// Create
			url, err := client.CreateRepo(ctx, p)
			if err != nil {
				return err
			}

			result := &output.CreateResult{
				Command:     "create",
				Repo:        p.FullName(),
				URL:         url,
				Created:     true,
				OwnerSource: string(cfg.Owner.Source),
				Applied:     buildAppliedMap(p),
				Warnings:    []string{},
			}

			// Post-create settings
			if err := client.EditRepo(ctx, p.FullName(), p); err != nil {
				if cfg.Strict.Value {
					return err
				}
				result.Warnings = append(result.Warnings, err.Error())
			}

			// Security settings
			secApplied, secWarnings := client.ApplySecuritySettings(ctx, p.FullName(), p, cfg.Strict.Value)
			_ = secApplied
			result.Warnings = append(result.Warnings, secWarnings...)

			// Clone
			result.Clone.Requested = p.CloneAfterCreate
			if p.CloneAfterCreate {
				if err := client.CloneRepo(ctx, p.FullName(), p.CloneDirectory, p.CloneExtraArgs); err != nil {
					result.Clone.Completed = false
					if cfg.Strict.Value {
						return err
					}
					result.Warnings = append(result.Warnings, err.Error())
				} else {
					result.Clone.Completed = true
					result.Clone.Directory = p.CloneDirectory
					if result.Clone.Directory == "" {
						result.Clone.Directory = p.Repo
					}
				}
			}

			// Open in browser
			if cfg.OpenRepo.Value {
				_ = client.OpenInBrowser(ctx, p.FullName())
			}

			// Output
			if flagJSON {
				return output.PrintJSON(os.Stdout, result)
			}
			output.PrintCreateHuman(os.Stdout, result)
			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVar(&flagPrivate, "private", false, "Force private repo")
	f.BoolVar(&flagPublic, "public", false, "Force public repo")
	f.StringVar(&flagDescription, "description", "", "Repo description")
	f.StringVar(&flagHomepage, "homepage", "", "Homepage URL")
	f.BoolVar(&flagEnableIssues, "enable-issues", false, "Enable issues")
	f.BoolVar(&flagDisableIssues, "disable-issues", false, "Disable issues")
	f.BoolVar(&flagEnableWiki, "enable-wiki", false, "Enable wiki")
	f.BoolVar(&flagDisableWiki, "disable-wiki", false, "Disable wiki")
	f.BoolVar(&flagEnableProjects, "enable-projects", false, "Enable projects")
	f.BoolVar(&flagDisableProjects, "disable-projects", false, "Disable projects")
	f.BoolVar(&flagAddReadme, "add-readme", false, "Initialize with README")
	f.StringVar(&flagGitignore, "gitignore", "", "Add gitignore template")
	f.StringVar(&flagLicense, "license", "", "Add license template")
	f.StringVar(&flagTemplate, "template", "", "Create from template repository")
	f.BoolVar(&flagClone, "clone", false, "Clone after creation/configuration")
	f.StringVar(&flagCloneDir, "clone-dir", "", "Clone destination directory")
	f.StringArrayVar(&flagCloneArgs, "clone-arg", nil, "Extra clone arg (repeatable)")
	f.BoolVar(&flagOpen, "open", false, "Open repo in browser after success")

	return cmd
}

func visibilityStr(private bool) string {
	if private {
		return "private"
	}
	return "public"
}

func initStr(p *policy.DesiredPolicy) string {
	if p.Template != "" {
		return "from template " + p.Template
	}
	if p.AutoInit {
		return "with README"
	}
	return "empty repository"
}

func buildAppliedMap(p *policy.DesiredPolicy) map[string]any {
	return map[string]any{
		"private":                p.Private,
		"allow_squash_merge":     p.AllowSquashMerge,
		"allow_merge_commit":     p.AllowMergeCommit,
		"allow_rebase_merge":     p.AllowRebaseMerge,
		"allow_auto_merge":       p.AllowAutoMerge,
		"delete_branch_on_merge": p.DeleteBranchOnMerge,
		"has_issues":             p.HasIssues,
		"has_wiki":               p.HasWiki,
		"has_projects":           p.HasProjects,
	}
}
