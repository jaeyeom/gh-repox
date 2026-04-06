# gh-repox

Manage GitHub repositories with opinionated defaults. This is a GitHub CLI extension that creates and configures repositories consistently across your organization.

## What It Does

gh-repox is a GitHub CLI extension that manages repository policy with sensible, opinionated defaults. It provides commands to:

- **Create** new repositories with your team's defaults applied automatically
- **Apply** policy to existing repositories to reconcile them with your standards
- **Diff** repositories to see what settings don't match your policy
- **Inspect** resolved configuration and trace where each value comes from

By default, gh-repox creates private, empty repositories with squash-only merges enabled and extra features like wiki and projects disabled. All defaults are configurable.

## Installation

Install as a GitHub CLI extension:

```bash
gh extension install jaeyeom/gh-repox
```

To verify installation:

```bash
gh repox --help
```

## Quick Start

Create a new private repository with opinionated defaults:

```bash
gh repox create my-service
```

That's it. The repository is created as private with squash merges enabled and merge commits disabled.

Apply policy to an existing repository:

```bash
gh repox apply owner/repo
```

See what settings need to change:

```bash
gh repox diff owner/repo
```

Inspect your configuration and where each value came from:

```bash
gh repox config explain
```

## Commands

### create

Create a new repository with resolved defaults.

```bash
gh repox create <name> [flags]
```

Examples:

```bash
gh repox create my-service
gh repox create my-service --public
gh repox create my-service --clone
gh repox create my-service --org acme
gh repox create my-service --description "worker service"
gh repox create my-service --template acme/go-template
gh repox create my-service --clone --clone-dir ~/src
```

Flags:

```
--private                  Force private repo
--public                   Force public repo
--description <text>       Repo description
--homepage <url>           Homepage URL
--enable-issues            Enable issues
--disable-issues           Disable issues
--enable-wiki              Enable wiki
--disable-wiki             Disable wiki
--enable-projects          Enable projects
--disable-projects         Disable projects
--add-readme               Initialize with README
--gitignore <template>     Add gitignore template
--license <key>            Add license template
--template <owner/name>    Create from template repository
--clone                    Clone after creation/configuration
--clone-dir <path>         Clone destination directory
--clone-arg <arg>          Extra clone arg (repeatable, passed after --)
--open                     Open repo in browser after success
```

By default, `create` builds an empty repository with no README, no `.gitignore`, and no license unless explicitly requested via flags.

### apply

Apply resolved policy to an existing repository.

```bash
gh repox apply <owner/repo> [flags]
```

Examples:

```bash
gh repox apply octocat/my-service
gh repox apply acme/platform --strict
```

The `apply` command reads current repository settings, compares them to your resolved policy, and applies changes. Failures in non-critical settings are reported as warnings unless strict mode is enabled.

### diff

Show the difference between a repository's current settings and your resolved policy.

```bash
gh repox diff <owner/repo> [flags]
```

Examples:

```bash
gh repox diff octocat/my-service
gh repox diff acme/platform --json
```

Output shows current values, desired values, and the source of each desired setting (default, config, env, or flag).

### config show

Display the fully resolved effective configuration.

```bash
gh repox config show [flags]
```

Shows final configuration with no source metadata.

### config explain

Display the fully resolved configuration with the source of each value.

```bash
gh repox config explain [flags]
```

Explains where each setting comes from: default, config file, environment variable, flag, or inferred (like owner from authenticated user).

## Opinionated Defaults

gh-repox comes with sensible defaults designed for safe, maintainable repositories:

### Visibility and Ownership

- **private**: true (repositories are private by default)
- **owner**: current authenticated GitHub user (never hardcoded)

### Repository Initialization

- **auto_init**: false (create empty repositories)
- **gitignore**: none (add explicitly via `--gitignore`)
- **license**: none (add explicitly via `--license`)
- **readme**: none (add explicitly via `--add-readme`)

Empty repositories give you full control over how projects are initialized.

### Merge Policy

- **allow_squash_merge**: true (squash is the primary merge method)
- **allow_merge_commit**: false
- **allow_rebase_merge**: false
- **allow_auto_merge**: true (enable auto-merge when branch is ready)
- **delete_branch_on_merge**: true (keep repository clean)

Squash-only merges encourage clean, linear history.

### Repository Features

- **has_issues**: true (issues are useful for most projects)
- **has_wiki**: false (keep documentation centralized)
- **has_projects**: false (use external project management)

### Security

- **dependency_graph**: true (enable dependency tracking)
- **dependabot_alerts**: true (alert on vulnerable dependencies)
- **dependabot_security_updates**: false (review updates manually)

## Configuration

Configuration is resolved with the following precedence (lowest to highest):

1. Built-in defaults
2. YAML config file
3. Environment variables
4. CLI flags

The highest precedence source wins.

### Config File Locations

gh-repox searches for configuration in this order:

1. `--config <path>` (explicit flag)
2. `.repox.yaml` or `repox.yaml` (current directory)
3. `~/.config/gh-repox/config.yaml` (user home)

Only the first found file is loaded.

### YAML Configuration

Create a `.repox.yaml` or `~/.config/gh-repox/config.yaml` file:

```yaml
github:
  host: github.com
  owner: null
  org: null

repo:
  private: true
  description: ""
  homepage: ""
  has_issues: true
  has_wiki: false
  has_projects: false
  auto_init: false
  gitignore: ""
  license: ""
  template: ""

merge:
  allow_squash_merge: true
  allow_merge_commit: false
  allow_rebase_merge: false
  allow_auto_merge: true
  delete_branch_on_merge: true

security:
  dependency_graph: true
  dependabot_alerts: true
  dependabot_security_updates: false

clone:
  after_create: false
  directory: ""
  extra_args: []

behavior:
  dry_run: false
  verbose: false
  strict: false
  open_repo: false
```

Field notes:

- `github.owner`: override personal owner (leave null to infer from authenticated user)
- `github.org`: default organization for new repositories
- `repo.auto_init`: initialize with README if true
- `repo.template`: template repository in `owner/name` format
- `clone.after_create`: automatically clone repo after creation
- `clone.directory`: directory to clone into (defaults to repo name)
- `clone.extra_args`: additional arguments passed to `git clone` after `--`

### Environment Variables

Supported environment variables:

```
REPOX_HOST                 GitHub hostname
REPOX_OWNER                Personal owner override
REPOX_ORG                  Organization override
REPOX_PRIVATE              true/false for repo visibility
REPOX_VERBOSE              true/false for verbose output
REPOX_DRY_RUN              true/false for dry-run mode
REPOX_STRICT               true/false for strict mode
REPOX_CLONE_AFTER_CREATE   true/false for clone behavior
```

### Persistent Flags

Available on all commands:

```
--config <path>      Config file path
--host <hostname>    GitHub hostname (default github.com)
--owner <owner>      Personal owner override
--org <org>          Organization override
--json               Machine-readable JSON output
--verbose            Verbose logs
--dry-run            Print plan without making changes
--strict             Fail on any post-create/apply setting failure
```

## Dry-Run Mode

Use `--dry-run` with `create` or `apply` to preview changes without making them:

```bash
gh repox create my-service --dry-run
```

Output shows:

- Resolved target owner and repository name
- Owner source (inferred or from config/flag)
- Visibility and initialization details
- Whether clone will run
- Exact `gh` commands that would be executed

## Strict Mode

By default, gh-repox attempts all changes and reports warnings for failures in non-critical settings.

Enable strict mode with `--strict` to fail if any requested setting fails:

```bash
gh repox create my-service --strict
```

Strict mode returns non-zero exit code on any failure, making it suitable for CI/CD pipelines.

## JSON Output

Use `--json` for machine-readable output suitable for automation:

```bash
gh repox create my-service --json
gh repox diff owner/repo --json
```

JSON output includes:

- Command executed
- Repository identifier
- Applied settings
- Any warnings or failures
- Clone results

## Exit Codes

gh-repox uses specific exit codes to signal different failure modes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (execution failed) |
| 2 | Invalid input or configuration |
| 3 | GitHub authentication unavailable |
| 4 | Repository creation failed |
| 5 | Post-create or apply settings failed in strict mode |
| 6 | Clone failed in strict mode |

## Examples

### Create a private repo and clone it

```bash
gh repox create my-api --clone --clone-dir ~/projects
```

### Apply policy to all repos in an organization

```bash
for repo in $(gh repo list myorg --jq '.[].name' -L 100); do
  gh repox apply "myorg/$repo" --strict
done
```

### Check what needs to change

```bash
gh repox diff myorg/platform --json | jq '.differences[] | select(.status == "different")'
```

### Create with a custom config file

```bash
gh repox create my-service --config ./team-defaults.yaml
```

### Explain where config values come from

```bash
gh repox config explain --org myorg
```

## Development

### Prerequisites

- Go 1.21 or later
- GNU Make

### Build

Build the binary for local development:

```bash
make build
```

This produces a local `gh-repox` binary for testing during development. This
binary is git-ignored and is **not** the artifact used for public installation.
Public users install precompiled binaries attached to GitHub releases (see
[Releasing](#releasing)).

### Test

Run all unit tests:

```bash
make test
```

Run tests with coverage:

```bash
make coverage
```

View HTML coverage report:

```bash
make coverage-html
```

### Format and Lint

Auto-format all Go source files:

```bash
make format
```

Run Go vet:

```bash
make vet
```

Full check (no mutations, safe for CI):

```bash
make check
```

Full workflow (format, lint, test, build):

```bash
make all
```

### Install Locally

Install the binary to `$GOPATH/bin`:

```bash
make install
```

Then use it as a GitHub CLI extension:

```bash
gh repox --help
```

## Releasing

### Release-readiness check

Before tagging a release, verify that the project cross-compiles for all
supported platforms:

```bash
make release-check
```

This builds binaries for macOS (amd64/arm64), Linux (amd64/arm64), and Windows
(amd64) into the `dist/` directory — the same platforms targeted by the release
workflow. If any platform fails to compile, fix the issue before publishing.

### Publishing a release

To publish a new release, push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The [release workflow](.github/workflows/release.yml) automatically
cross-compiles binaries and attaches them to the GitHub release. Once published,
users can install or upgrade with:

```bash
gh extension install jaeyeom/gh-repox
gh extension upgrade jaeyeom/gh-repox
```

### Local builds vs public installation

| Path | Artifact | Purpose |
|------|----------|---------|
| `make build` | `./gh-repox` (git-ignored) | Local development and testing only |
| `make release-check` | `dist/*` (git-ignored) | Verify cross-compilation before tagging |
| `git push origin v*` | GitHub release binaries | Public installation via `gh extension install` |

A successful `make build` does **not** guarantee that the repository is publicly
installable. Always run `make release-check` before publishing a new tag.

## Architecture

gh-repox is organized into focused packages:

- `cmd/gh-repox`: Binary entry point
- `internal/cli`: Cobra command definitions and flag wiring
- `internal/config`: Configuration loading, merging, and source tracking
- `internal/policy`: DesiredPolicy and ActualState models
- `internal/github`: GitHub API interactions via `gh` CLI
- `internal/exec`: Subprocess execution abstraction
- `internal/diff`: Comparing actual vs desired settings
- `internal/output`: Human and JSON output formatting
- `internal/validate`: Semantic configuration validation

Configuration is resolved once and tracked with source metadata, allowing both `config explain` and `diff` to show where values come from.

## Limitations

Version 1 focuses on core functionality:

- No branch protection or ruleset management
- No team assignment for org repositories
- Single config file loading (future: merge user + repo-local)
- Limited security settings (future: expanded coverage)

## License

MIT
