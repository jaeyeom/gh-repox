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

			// Reject conflicting flag pairs
			conflicts := [][2]string{
				{"private", "public"},
				{"enable-issues", "disable-issues"},
				{"enable-wiki", "disable-wiki"},
				{"enable-projects", "disable-projects"},
			}
			for _, pair := range conflicts {
				if cmd.Flags().Changed(pair[0]) && cmd.Flags().Changed(pair[1]) {
					return exitErrorf(ExitInvalidInput, "conflicting flags: --%s and --%s", pair[0], pair[1])
				}
			}

			cfg, err := resolveConfig()
			if err != nil {
				return exitErrorf(ExitInvalidInput, "config error: %w", err)
			}

			// Apply create-specific flags
			applyCreateRepoFlags(cfg, flagPrivate, flagPublic, flagDescription, flagHomepage,
				flagEnableIssues, flagDisableIssues, flagEnableWiki, flagDisableWiki,
				flagEnableProjects, flagDisableProjects)
			applyCreateInitFlags(cfg, flagAddReadme, flagGitignore, flagLicense, flagTemplate,
				flagClone, flagCloneDir, flagCloneArgs, flagOpen)

			// Resolve owner
			if err := resolveOwner(ctx, cfg); err != nil {
				return exitErrorf(ExitNoAuth, "%s", err)
			}

			// Build policy
			p := policy.FromConfig(cfg, repoName)

			// Validate
			if err := validate.Create(p); err != nil {
				return exitErrorf(ExitInvalidInput, "validate create: %w", err)
			}

			runner := &exec.RealRunner{}
			client := ghclient.NewClient(runner, cfg.Host.Value)

			// Dry run
			if cfg.DryRun.Value {
				if flagJSON {
					return printCreateDryRunJSON(repoName, p, cfg)
				}
				printCreateDryRun(repoName, p, cfg)
				return nil
			}

			return executeCreate(ctx, client, cfg, p)
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

func applyCreateRepoFlags(
	cfg *config.Config,
	flagPrivate, flagPublic bool,
	flagDescription, flagHomepage string,
	flagEnableIssues, flagDisableIssues bool,
	flagEnableWiki, flagDisableWiki bool,
	flagEnableProjects, flagDisableProjects bool,
) {
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
}

func applyCreateInitFlags(
	cfg *config.Config,
	flagAddReadme bool,
	flagGitignore, flagLicense, flagTemplate string,
	flagClone bool,
	flagCloneDir string,
	flagCloneArgs []string,
	flagOpen bool,
) {
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
}

func printCreateDryRunJSON(_ string, p *policy.DesiredPolicy, cfg *config.Config) error {
	cmds := ghclient.PlannedCommands(p, cfg.Host.Value)
	result := &output.DryRunResult{
		Command:  "create",
		DryRun:   true,
		Repo:     p.FullName(),
		Commands: cmds,
	}
	if err := output.PrintJSON(os.Stdout, result); err != nil {
		return fmt.Errorf("print JSON: %w", err)
	}
	return nil
}

func printCreateDryRun(repoName string, p *policy.DesiredPolicy, cfg *config.Config) {
	header := fmt.Sprintf("Dry run: gh repox create %s\n\nResolved target:\n- repo: %s\n- owner source: %s\n- visibility: %s\n- init: %s\n- clone after create: %v",
		repoName, p.FullName(), cfg.Owner.Source, visibilityStr(p.Private), initStr(p), p.CloneAfterCreate)
	cmds := ghclient.PlannedCommands(p, cfg.Host.Value)
	output.PrintDryRun(os.Stdout, header, cmds)
}

func executeCreate(ctx context.Context, client *ghclient.Client, cfg *config.Config, p *policy.DesiredPolicy) error {
	exists, err := client.RepoExists(ctx, p.FullName())
	if err != nil {
		return fmt.Errorf("check repo exists: %w", err)
	}
	if exists {
		return exitErrorf(ExitInvalidInput, "repository %s already exists", p.FullName())
	}

	url, err := client.CreateRepo(ctx, p)
	if err != nil {
		return exitErrorf(ExitCreateFailed, "create repo: %w", err)
	}

	result := &output.CreateResult{
		Command:     "create",
		Repo:        p.FullName(),
		URL:         url,
		Created:     true,
		OwnerSource: string(cfg.Owner.Source),
		Applied:     map[string]any{},
		Warnings:    []string{},
	}

	if err := client.EditRepo(ctx, p.FullName(), p); err != nil {
		if cfg.Strict.Value {
			result.Warnings = append(result.Warnings, err.Error())
			if flagJSON {
				_ = output.PrintJSON(os.Stdout, result)
			} else {
				output.PrintCreateHuman(os.Stdout, result)
			}
			return exitErrorf(ExitStrictFailed, "edit repo: %w", err)
		}
		result.Warnings = append(result.Warnings, err.Error())
	} else {
		result.Applied = buildAppliedMap(p)
	}

	secApplied, secWarnings := client.ApplySecuritySettings(ctx, p.FullName(), p, cfg.Strict.Value)
	for _, s := range secApplied {
		result.Applied[s] = true
	}
	result.Warnings = append(result.Warnings, secWarnings...)

	if cfg.Strict.Value && len(result.Warnings) > 0 {
		if flagJSON {
			_ = output.PrintJSON(os.Stdout, result)
		} else {
			output.PrintCreateHuman(os.Stdout, result)
		}
		return exitErrorf(ExitStrictFailed, "strict mode: %d warning(s)", len(result.Warnings))
	}

	if err := handleClone(ctx, client, cfg, p, result); err != nil {
		return err
	}

	if cfg.OpenRepo.Value {
		_ = client.OpenInBrowser(ctx, p.FullName())
	}

	if flagJSON {
		if err := output.PrintJSON(os.Stdout, result); err != nil {
			return fmt.Errorf("print JSON: %w", err)
		}
		return nil
	}
	output.PrintCreateHuman(os.Stdout, result)
	return nil
}

func handleClone(ctx context.Context, client *ghclient.Client, cfg *config.Config, p *policy.DesiredPolicy, result *output.CreateResult) error {
	result.Clone.Requested = p.CloneAfterCreate
	if !p.CloneAfterCreate {
		return nil
	}
	if err := client.CloneRepo(ctx, p.FullName(), p.CloneDirectory, p.CloneExtraArgs); err != nil {
		result.Clone.Completed = false
		if cfg.Strict.Value {
			return exitErrorf(ExitCloneFailed, "clone repo: %w", err)
		}
		result.Warnings = append(result.Warnings, err.Error())
	} else {
		result.Clone.Completed = true
		result.Clone.Directory = p.CloneDirectory
		if result.Clone.Directory == "" {
			result.Clone.Directory = p.Repo
		}
	}
	return nil
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
