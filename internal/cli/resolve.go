package cli

import (
	"context"
	"fmt"

	"github.com/jaeyeom/gh-repox/internal/config"
	"github.com/jaeyeom/gh-repox/internal/exec"
	ghclient "github.com/jaeyeom/gh-repox/internal/github"
)

// resolveConfig loads config with full precedence: defaults < file < env < flags.
func resolveConfig() (*config.Config, error) {
	cfg := config.Defaults()

	// Load config file
	path := config.FindConfigFile(flagConfig)
	if err := cfg.LoadFile(path); err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}

	// Load env
	cfg.LoadEnv()

	// Apply CLI flags (highest precedence)
	applyFlags(cfg)

	return cfg, nil
}

// applyFlags applies CLI flag overrides to config.
func applyFlags(cfg *config.Config) {
	if flagHost != "" {
		cfg.Host.Set(flagHost, config.SourceFlag)
	}
	if flagOwner != "" {
		cfg.Owner.Set(flagOwner, config.SourceFlag)
	}
	if flagOrg != "" {
		cfg.Org.Set(flagOrg, config.SourceFlag)
	}
	if flagDryRun {
		cfg.DryRun.Set(true, config.SourceFlag)
	}
	if flagStrict {
		cfg.Strict.Set(true, config.SourceFlag)
	}
}

// resolveOwner resolves the target owner using the resolution order.
func resolveOwner(ctx context.Context, cfg *config.Config) error {
	// If org or owner already set by flag/config/env, we're done
	if cfg.Org.Value != "" || cfg.Owner.Value != "" {
		return nil
	}

	// Infer from authenticated user
	runner := &exec.RealRunner{}
	client := ghclient.NewClient(runner, cfg.Host.Value)
	login, err := client.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("get authenticated user: %w", err)
	}
	cfg.Owner.Set(login, config.SourceInferred)
	return nil
}
