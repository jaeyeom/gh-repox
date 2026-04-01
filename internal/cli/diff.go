package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaeyeom/gh-repox/internal/diff"
	"github.com/jaeyeom/gh-repox/internal/exec"
	ghclient "github.com/jaeyeom/gh-repox/internal/github"
	"github.com/jaeyeom/gh-repox/internal/output"
	"github.com/jaeyeom/gh-repox/internal/policy"
	"github.com/jaeyeom/gh-repox/internal/validate"
)

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <owner/repo>",
		Short: "Show drift between current repo settings and desired policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()

			owner, repo, err := validate.ParseOwnerRepo(args[0])
			if err != nil {
				return fmt.Errorf("parse repo: %w", err)
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

			actual, err := client.FetchRepoState(ctx, fullName)
			if err != nil {
				return fmt.Errorf("fetch repo state: %w", err)
			}

			entries := diff.Compare(p, actual, cfg)
			result := &output.DiffResult{
				Command:     "diff",
				Repo:        fullName,
				Differences: entries,
			}

			if flagJSON {
				return output.PrintJSON(os.Stdout, result)
			}
			output.PrintDiffHuman(os.Stdout, result)
			return nil
		},
	}
	return cmd
}
