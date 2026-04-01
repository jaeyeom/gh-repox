# CLAUDE.md

## Project Overview

gh-repox is a precompiled Go GitHub CLI extension that manages repository policy with opinionated defaults. Users invoke it as `gh repox <subcommand>`.

## Build & Test Commands

- `make all` — format, fix, test, build (full local workflow)
- `make check` — CI-safe checks (no mutation)
- `make build` — compile binary to `./gh-repox`
- `make test` — run all tests
- `make coverage` — generate coverage profile
- `make format` — auto-format with gofmt
- `make lint` — run go vet
- `go mod tidy` — tidy dependencies

## Project Structure

- `cmd/gh-repox/` — main entry point
- `internal/cli/` — Cobra commands and flag wiring (root, create, apply, diff, config)
- `internal/config/` — YAML config loading, env vars, defaults, source tracking with `Field[T]` generic
- `internal/policy/` — `DesiredPolicy` and `ActualState` structs, policy builder
- `internal/github/` — `gh` CLI command construction, repo CRUD, security settings
- `internal/exec/` — `Runner` interface for subprocess execution (real + mock)
- `internal/diff/` — compare desired vs actual repo state
- `internal/output/` — human-readable and JSON output formatters
- `internal/validate/` — semantic validation (ParseOwnerRepo, ValidateCreate, ValidateApply)

## Architecture Notes

- Config precedence: defaults < YAML file < env vars < CLI flags (tracked via `config.Source`)
- All `gh` commands go through the `exec.Runner` interface for testability
- The `MockRunner` in `internal/exec` is used across all test files
- `policy.DesiredPolicy` is the shared model used by create, apply, and diff commands
- Config file discovery: `--config` flag > `./.repox.yaml` > `./repox.yaml` > `~/.config/gh-repox/config.yaml`

## Conventions

- Go 1.26+, uses generics (`Field[T]`)
- Dependencies: cobra, yaml.v3 (kept minimal)
- Tests use table-driven patterns and `t.TempDir()` / `t.Setenv()`
- No hardcoded usernames; owner is inferred from `gh api user --jq .login`
