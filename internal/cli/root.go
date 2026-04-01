package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagConfig  string
	flagHost    string
	flagOwner   string
	flagOrg     string
	flagJSON    bool
	flagVerbose bool
	flagDryRun  bool
	flagStrict  bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gh-repox",
		Short:         "Manage repository policy with opinionated defaults",
		Long:          "gh-repox creates and manages GitHub repositories with consistent, opinionated defaults.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := cmd.PersistentFlags()
	pf.StringVar(&flagConfig, "config", "", "Config file path")
	pf.StringVar(&flagHost, "host", "", "GitHub hostname")
	pf.StringVar(&flagOwner, "owner", "", "Personal owner override")
	pf.StringVar(&flagOrg, "org", "", "Organization override")
	pf.BoolVar(&flagJSON, "json", false, "Machine-readable JSON output")
	pf.BoolVar(&flagVerbose, "verbose", false, "Verbose logs")
	pf.BoolVar(&flagDryRun, "dry-run", false, "Print plan without making changes")
	pf.BoolVar(&flagStrict, "strict", false, "Fail on any post-create/apply setting failure")

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newApplyCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
