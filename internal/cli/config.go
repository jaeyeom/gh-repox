package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaeyeom/gh-repox/internal/output"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and inspect resolved configuration",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigExplainCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the fully resolved effective configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := resolveConfig()
			if err != nil {
				return exitErrorf(ExitInvalidInput, "config error: %w", err)
			}

			// Resolve owner for display
			if cfg.Owner.Value == "" && cfg.Org.Value == "" {
				if err := resolveOwner(context.Background(), cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve owner: %v\n", err)
				}
			}

			entries := cfg.Entries()
			if flagJSON {
				m := make(map[string]any)
				for _, e := range entries {
					m[e.Key] = e.Value
				}
				return output.PrintJSON(os.Stdout, m)
			}
			output.PrintConfigShow(os.Stdout, entries)
			return nil
		},
	}
}

func newConfigExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: "Show effective configuration with the source of each value",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := resolveConfig()
			if err != nil {
				return exitErrorf(ExitInvalidInput, "config error: %w", err)
			}

			// Resolve owner for display
			if cfg.Owner.Value == "" && cfg.Org.Value == "" {
				if err := resolveOwner(context.Background(), cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve owner: %v\n", err)
				}
			}

			entries := cfg.Entries()
			if flagJSON {
				type explainEntry struct {
					Key    string `json:"key"`
					Value  any    `json:"value"`
					Source string `json:"source"`
				}
				var result []explainEntry
				for _, e := range entries {
					result = append(result, explainEntry{
						Key:    e.Key,
						Value:  e.Value,
						Source: string(e.Source),
					})
				}
				return output.PrintJSON(os.Stdout, result)
			}
			output.PrintConfigExplain(os.Stdout, entries)
			return nil
		},
	}
}
