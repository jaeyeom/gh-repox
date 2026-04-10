# Design Doc: `gh-repox`

## Overview

Build a GitHub CLI extension named `gh-repox` that manages repository policy with opinionated defaults.

The extension should support:
- creating new repositories with team defaults
- reconciling existing repositories to those defaults
- showing drift between current repo settings and desired defaults
- inspecting resolved configuration and where each value came from

This is a GitHub CLI extension, so users will invoke it as:

```bash
gh repox <subcommand> ...
```

Examples:
```bash
gh repox create my-repo
gh repox apply owner/repo
gh repox diff owner/repo
gh repox config show
gh repox config explain
```

The extension should be implemented as a precompiled Go GitHub CLI extension.[web:82][web:90]

---

## Product goals

1. Make repository setup consistent and fast.
2. Encode team defaults once and reuse them everywhere.
3. Avoid hardcoded owner values; derive the authenticated user dynamically.
4. Allow overrides through YAML config and CLI flags.
5. Support both creating new repos and enforcing policy on existing repos.
6. Keep output understandable, concise, and scriptable.

---

## Non-goals

- Replacing `gh` authentication.
- Managing every possible GitHub repository setting in v1.
- Full branch protection and ruleset management in v1.
- GitHub Enterprise Server-specific feature parity in v1 beyond host selection.
- Local git init / first commit workflows in v1.

---

## Extension identity

Repository name:
```text
gh-repox
```

Binary name:
```text
gh-repox
```

Installed invocation:
```bash
gh repox
```

This extension should be structured as a multi-command CLI with a root command and subcommands.[web:82][web:84]

---

## Core commands

### `gh repox create <name>`

Create a new GitHub repository using resolved defaults plus any overrides.

Responsibilities:
- resolve target owner or org
- create an empty repository by default
- apply post-creation settings
- optionally clone the repo locally after configuration is complete

Examples:
```bash
gh repox create my-service
gh repox create my-service --org acme
gh repox create my-service --public
gh repox create my-service --clone
gh repox create my-service --description "internal API service"
```

### `gh repox apply <owner/repo>`

Apply the resolved policy to an existing repository.

Responsibilities:
- read current repository settings
- reconcile desired settings
- apply supported changes
- report what changed and what could not be changed

Examples:
```bash
gh repox apply octocat/my-service
gh repox apply acme/platform --strict
```

### `gh repox diff <owner/repo>`

Show the difference between the repository’s current settings and the resolved desired policy.

Responsibilities:
- resolve desired policy
- fetch current settings
- render a human-readable diff
- optionally render JSON for automation

Examples:
```bash
gh repox diff octocat/my-service
gh repox diff acme/platform --json
```

### `gh repox config show`

Show the fully resolved effective configuration.

Examples:
```bash
gh repox config show
gh repox config show --json
```

### `gh repox config explain`

Show the effective configuration and the source of each value.

Examples:
```bash
gh repox config explain
gh repox config explain --org acme
```

---

## Opinionated defaults

These are the built-in defaults.

### Ownership and visibility

- owner: current authenticated GitHub user
- visibility: private

### Repository initialization

Default behavior: create an **empty repository**, with no README, no `.gitignore`, and no license unless explicitly requested.[web:54]

This is important:
- do not initialize with README by default
- do not add `.gitignore` by default
- do not add a license by default

### Merge policy

- allow_squash_merge: true
- allow_merge_commit: false
- allow_rebase_merge: false
- allow_auto_merge: true
- delete_branch_on_merge: true

Interpretation:
- squash merge should be the only enabled merge method by default

### Repository features

- has_issues: true
- has_wiki: false
- has_projects: false

### Security defaults

- dependency_graph: true
- dependabot_alerts: true
- dependabot_security_updates: false by default in v1

### Clone behavior

- clone_after_create: false by default
- if enabled, clone only after create + configuration succeeds or finishes best-effort

---

## Design principles

1. **Safe by default**: private repo, squash-only merges, no extra repo surface area.
2. **Explicit overrides**: users can override defaults in config or flags.
3. **No hardcoded username**: owner is inferred from authenticated `gh` session unless explicitly set.
4. **Empty repo first**: no bootstrap files unless requested.
5. **Best-effort configuration**: apply all supported settings and report partial failures clearly.
6. **Dry-run friendly**: show exactly what would happen before mutation.
7. **Composable**: commands should share the same config resolution and repo-policy model.

---

## Configuration precedence

Resolved configuration should follow this order, lowest to highest precedence:

1. built-in defaults
2. YAML config file
3. environment variables
4. CLI flags

Highest precedence wins.

This precedence should apply consistently across all subcommands.

---

## Config file discovery

Support these config sources:

1. `--config <path>` if provided
2. repo-local config: `./.repox.yaml` or `./repox.yaml`
3. user config: `~/.config/gh-repox/config.yaml`

For v1:
- load only one config file
- precedence among file locations:
  1. explicit `--config`
  2. repo-local
  3. user config

Future enhancement:
- merge user config + repo-local config

---

## YAML schema

### Example config

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

### Field notes

- `github.owner`: explicit personal owner override
- `github.org`: explicit default org override
- `repo.auto_init`: whether to initialize with README; default false
- `repo.template`: optional template repository in `owner/name` form
- `clone.after_create`: if true, run `gh repo clone` after creation/configuration
- `clone.directory`: optional local destination
- `clone.extra_args`: pass-through arguments after `--` to `git clone` via `gh repo clone`.[web:117]

---

## CLI interface

## Root command

```bash
gh repox
```

### Persistent flags

Available to all subcommands:

```text
--config <path>              Config file path
--host <hostname>            GitHub hostname, default github.com
--owner <owner>              Personal owner override
--org <org>                  Organization override
--json                       Machine-readable output
--dry-run                    Print plan without making changes
--strict                     Fail on any requested post-create/apply setting failure
```

---

## `create` command

### Usage

```bash
gh repox create <name> [flags]
```

### Flags

```text
--private                    Force private repo
--public                     Force public repo
--description <text>         Repo description
--homepage <url>             Homepage URL
--enable-issues              Enable issues
--disable-issues             Disable issues
--enable-wiki                Enable wiki
--disable-wiki               Disable wiki
--enable-projects            Enable projects
--disable-projects           Disable projects

--add-readme                 Initialize with README
--gitignore <template>       Add gitignore template
--license <key>              Add license template
--template <owner/name>      Create from template repository

--clone                      Clone after creation/configuration
--clone-dir <path>           Clone destination directory
--clone-arg <arg>            Extra clone arg, repeatable
--open                       Open repo in browser after success
```

### Notes

- By default, `create` should create an empty repo: no README, no `.gitignore`, no license.[web:54]
- If `--clone` is set, the tool should run `gh repo clone OWNER/REPO` after configuration is finished.[web:117]
- If `--clone-dir` is provided, clone into that directory.
- If extra clone args are provided, pass them after `--` to `gh repo clone`.[web:117]

### Example invocations

```bash
gh repox create my-service
gh repox create my-service --clone
gh repox create my-service --clone --clone-dir ~/src
gh repox create my-service --org acme --public
gh repox create my-service --description "worker service"
gh repox create my-service --template acme/go-template
```

---

## `apply` command

### Usage

```bash
gh repox apply <owner/repo> [flags]
```

### Behavior

- apply resolved policy to an existing repository
- do not create a repo
- do not clone anything
- produce a report of changed settings and warnings

### Example invocations

```bash
gh repox apply octocat/my-service
gh repox apply acme/platform --strict
```

---

## `diff` command

### Usage

```bash
gh repox diff <owner/repo> [flags]
```

### Behavior

- fetch current repo settings
- compare against resolved desired policy
- display differences as:
  - current value
  - desired value
  - source of desired value

### Example invocations

```bash
gh repox diff octocat/my-service
gh repox diff acme/platform --json
```

---

## `config show` command

### Usage

```bash
gh repox config show [flags]
```

### Behavior

- print fully resolved config
- no source metadata, just final effective values

---

## `config explain` command

### Usage

```bash
gh repox config explain [flags]
```

### Behavior

- print fully resolved config
- show source of each field:
  - default
  - config
  - env
  - flag
  - inferred

This is especially important for debugging precedence.

---

## Owner and org resolution

### Requirement

Never hardcode a username.

### Resolution order

Use this order to resolve target owner:

1. `--org`
2. `--owner`
3. `github.org` from config
4. `github.owner` from config
5. inferred authenticated user from `gh`

If an org is selected, repo target becomes `ORG/REPO`.
If no org is selected, target becomes `OWNER/REPO`.

### Authenticated user detection

Preferred approach:
```bash
gh api user --jq .login
```

Diagnostics / fallback:
```bash
gh auth status --active --json hosts
```

If detection fails:
- return a clear error telling the user to run `gh auth login` or pass `--owner` / `--org`.[web:67][web:78]

### Host awareness

If `--host` or config host is set:
- use that host consistently in all `gh` and `gh api` calls

---

## Repository creation flow

### High-level steps for `create`

1. parse args and flags
2. load config
3. merge defaults + config + env + flags
4. validate config
5. resolve owner/org
6. verify `gh` installation and auth
7. verify target repo does not already exist
8. create repository
9. apply post-creation settings
10. optionally clone
11. optionally open repo
12. print summary

### Repo creation details

Primary command:
```bash
gh repo create OWNER/REPO ...
```

Default creation mode:
- empty repo
- private unless overridden
- no README
- no gitignore
- no license

Do not pass init flags unless explicitly requested.[web:54]

### Create command logic

If no template:
- use `gh repo create`

If template is specified:
- use the best supported template flow
- then apply the same post-create settings

### Repo existence check

Check before creating:
```bash
gh repo view OWNER/REPO
```

If repo exists:
- fail clearly with non-zero exit code
- future enhancement: `--adopt` or `--apply-if-exists`

---

## Post-creation configuration flow

After creation, configure the repository to match resolved policy.

### Merge and repo settings

Primary path:
```bash
gh repo edit OWNER/REPO ...
```

Apply these where supported:
- squash merge enabled
- merge commits disabled
- rebase merge disabled
- auto-merge enabled
- delete branch on merge enabled
- wiki enabled/disabled
- issues enabled/disabled if needed
- projects enabled/disabled if supported

### Security settings

Use `gh api` for settings not exposed through `gh repo edit`.

Examples:
- vulnerability / Dependabot alerts endpoints
- other repo patch endpoints as needed

Behavior:
- attempt each requested setting independently
- collect failures individually
- in non-strict mode, continue and report warnings
- in strict mode, exit non-zero if any requested setting failed

### Stricter org policy handling

If org policy already enforces a stricter setting:
- do not attempt to loosen it
- treat it as compliant
- mention it in verbose or diff output if relevant

---

## Clone-after-create behavior

### Goal

Support optionally cloning the newly created repo after all configuration is complete.

### Why after configuration

Run clone **after** creation and post-create reconciliation so:
- the created repo URL is stable
- failures in config are known before local checkout
- clone behavior is isolated from create behavior

### Command form

Default:
```bash
gh repo clone OWNER/REPO
```

If `--clone-dir` is set:
- run clone in that parent directory or target destination as appropriate

If extra clone args are set:
```bash
gh repo clone OWNER/REPO -- <extra git clone args>
```

### Examples

```bash
gh repox create my-service --clone
gh repox create my-service --clone --clone-dir ~/src
gh repox create my-service --clone --clone-arg=--depth=1
```

### Failure behavior

If create/config succeeded but clone fails:
- repository creation remains success
- output should indicate:
  - repo created successfully
  - clone failed separately
- in `--strict` mode:
  - strict should apply to requested workflow as a whole
  - clone failure should produce non-zero exit

---

## Internal policy model

Create a single internal “desired repo policy” model used by `create`, `apply`, and `diff`.

Example shape:

```go
type DesiredPolicy struct {
    Host   string
    Owner  string
    Repo   string

    Private      bool
    Description  string
    Homepage     string
    HasIssues    bool
    HasWiki      bool
    HasProjects  bool

    AutoInit     bool
    Gitignore    string
    License      string
    Template     string

    AllowSquashMerge    bool
    AllowMergeCommit    bool
    AllowRebaseMerge    bool
    AllowAutoMerge      bool
    DeleteBranchOnMerge bool

    DependencyGraph           bool
    DependabotAlerts          bool
    DependabotSecurityUpdates bool

    CloneAfterCreate bool
    CloneDirectory   string
    CloneExtraArgs   []string
}
```

This shared model should be the output of config resolution.

---

## Internal architecture

### Recommended language

Go.

### Recommended command framework

Cobra.

Reason:
- natural fit for root command + multiple subcommands
- good flag handling
- good help generation
- shell completion support
- clean project organization for a growing extension

### Suggested package layout

```text
cmd/gh-repox/
internal/cli/
internal/config/
internal/policy/
internal/github/
internal/exec/
internal/diff/
internal/output/
internal/validate/
```

### Package responsibilities

- `cli`: cobra commands and flag wiring
- `config`: load YAML, env, and defaults; resolve precedence
- `policy`: build final desired policy model
- `github`: `gh`/`gh api` command construction and GitHub-specific logic
- `exec`: subprocess runner abstraction
- `diff`: compare actual vs desired settings
- `output`: human and JSON output
- `validate`: semantic validation

---

## Execution abstraction

Wrap subprocess execution behind an interface so commands can be tested.

```go
type Runner interface {
    Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error)
}
```

Implementation notes:
- do not use shell string interpolation
- pass args explicitly
- capture stdout/stderr separately

---

## Fetching current repo state

Needed for:
- `apply`
- `diff`
- post-create verification

Data sources:
- `gh repo view`
- `gh api repos/{owner}/{repo}`
- other `gh api` endpoints for security settings as needed

The fetched state should be normalized into a single internal “actual repo state” struct for comparison.

Example:
```go
type ActualState struct {
    Private      bool
    HasIssues    bool
    HasWiki      bool
    HasProjects  bool

    AllowSquashMerge    bool
    AllowMergeCommit    bool
    AllowRebaseMerge    bool
    AllowAutoMerge      bool
    DeleteBranchOnMerge bool

    DependencyGraph           *bool
    DependabotAlerts          *bool
    DependabotSecurityUpdates *bool
}
```

Use pointer booleans where the API may not expose a value reliably.

---

## Diff model

`diff` should compare desired policy to actual state and produce entries like:

```go
type DiffEntry struct {
    Field        string
    Current      any
    Desired      any
    DesiredSource string
    Status       string // same, different, unknown, unsupported
}
```

Human output example:
```text
allow_merge_commit:
  current: true
  desired: false
  source: default
  status: different
```

JSON output should include the same information in machine-readable form.

---

## Config precedence and explainability

Track source metadata for every resolved field.

Example sources:
- default
- config
- env
- flag
- inferred

This metadata should be available to:
- `config explain`
- `diff`
- verbose logs
- dry-run output

Example:
```text
owner=octocat         source=inferred
private=true          source=default
has_wiki=false        source=config
clone_after_create=true source=flag
```

---

## Environment variable support

Support a small, explicit set of environment variables.

Suggested variables:
```text
REPOX_HOST
REPOX_OWNER
REPOX_ORG
REPOX_PRIVATE
REPOX_VERBOSE
REPOX_DRY_RUN
REPOX_STRICT
REPOX_CLONE_AFTER_CREATE
```

Do not make env handling too magical in v1.

---

## Error handling

### Exit codes

- `0`: success
- `1`: runtime error
- `2`: invalid input or config
- `3`: `gh` missing or auth unavailable
- `4`: repository creation failed
- `5`: post-create or apply reconciliation failed in strict mode
- `6`: clone failed in strict mode

### Behavior principles

- fail fast on invalid input
- do best-effort on non-critical post-create settings unless strict mode is enabled
- do not hide partial success
- keep messages actionable

### Examples

Auth failure:
```text
Could not determine authenticated GitHub user.
Run `gh auth login`, or pass --owner or --org.
```

Partial configuration failure:
```text
Repository created: https://github.com/acme/my-service

Applied:
- private
- squash merge enabled
- merge commits disabled
- rebase merge disabled
- auto-merge enabled

Warnings:
- could not enable Dependabot alerts: insufficient permission
- could not disable projects: unsupported by current host or API
```

Create succeeded, clone failed:
```text
Repository created and configured: https://github.com/acme/my-service

Warnings:
- clone failed: destination path already exists
```

---

## Output design

### Default human output

Keep concise.

For create:
```text
Created repository: https://github.com/OWNER/REPO

Applied settings:
- private
- squash-only merges
- auto-merge enabled
- delete branch on merge enabled
- wiki disabled
- dependency graph enabled
- Dependabot alerts enabled

Clone:
- skipped
```

If clone ran:
```text
Clone:
- completed at ~/src/my-service
```

### JSON output

For `create`:
```json
{
  "command": "create",
  "repo": "OWNER/REPO",
  "url": "https://github.com/OWNER/REPO",
  "created": true,
  "owner_source": "gh-auth",
  "applied": {
    "private": true,
    "allow_squash_merge": true,
    "allow_merge_commit": false,
    "allow_rebase_merge": false,
    "allow_auto_merge": true,
    "delete_branch_on_merge": true,
    "has_wiki": false,
    "has_projects": false,
    "dependency_graph": true,
    "dependabot_alerts": true
  },
  "clone": {
    "requested": true,
    "completed": true,
    "directory": "~/src/my-service"
  },
  "warnings": []
}
```

For `diff`:
```json
{
  "command": "diff",
  "repo": "OWNER/REPO",
  "differences": [
    {
      "field": "allow_merge_commit",
      "current": true,
      "desired": false,
      "desired_source": "default",
      "status": "different"
    }
  ]
}
```

---

## Dry-run behavior

`--dry-run` should work on `create` and `apply`.

### `create --dry-run`

Must show:
- final resolved owner and target repo
- resolved policy
- whether repo will be empty
- whether clone will run
- exact `gh` / `gh api` commands planned

Example:
```text
Dry run: gh repox create my-service

Resolved target:
- repo: octocat/my-service
- owner source: inferred
- visibility: private
- init: empty repository
- clone after create: true

Planned commands:
1. gh repo create octocat/my-service --private
2. gh repo edit octocat/my-service --enable-squash-merge --disable-merge-commit --disable-rebase-merge --enable-auto-merge --delete-branch-on-merge --disable-wiki
3. gh api --method PUT /repos/octocat/my-service/vulnerability-alerts
4. gh repo clone octocat/my-service
```

### `apply --dry-run`

Must show what would change without changing anything.

---

## Suggested implementation details

### Go libraries

Recommendation:
- Cobra for command structure
- `gopkg.in/yaml.v3` for YAML parsing

Avoid relying on heavy config magic in v1.
Keep precedence merging explicit in code.

### Why explicit config merge

This tool has:
- defaults
- inferred values
- YAML config
- env vars
- CLI flags

Explicit merge logic is easier to debug than abstract precedence layers.

---

## Testing strategy

### Unit tests

- config loading
- precedence merging
- owner resolution
- empty-repo create command generation
- clone command generation
- apply command generation
- diff generation
- source tracking
- strict vs non-strict behavior

### Integration tests

Use a mocked runner to simulate:
- successful auth
- failed auth
- create success
- repo already exists
- partial `gh repo edit` failure
- security API failure
- clone failure

### Manual tests

- `gh repox create my-repo`
- `gh repox create my-repo --clone`
- `gh repox create my-repo --org acme`
- `gh repox apply acme/platform`
- `gh repox diff acme/platform`
- `gh repox config explain`

---

## Milestones

### v1

- root command + subcommands
- `create`
- `apply`
- `diff`
- `config show`
- `config explain`
- owner inference from authenticated `gh`
- empty repo creation by default
- optional clone after create
- YAML config
- CLI flag overrides
- dry-run
- JSON output

### v1.1

- branch protection or ruleset support
- apply-if-exists mode
- multi-file config merge
- team assignment for org repos
- support for more security settings
- org-wide audit/reporting

---

## Acceptance criteria

The extension is acceptable for v1 if:

1. `gh repox create my-repo` creates a private, empty repo under the authenticated user with no hardcoded owner.
2. `gh repox create my-repo --clone` creates the repo, applies settings, then clones it locally.
3. `gh repox create my-repo --org acme` creates the repo under `acme`.
4. `gh repox apply owner/repo` reconciles an existing repo to the desired policy.
5. `gh repox diff owner/repo` shows drift between actual settings and desired defaults.
6. Config precedence works: defaults < config < env < flags.
7. `gh repox config explain` shows the source of each resolved value.
8. Empty repository creation is the default unless init flags are explicitly requested.
9. Partial failures are reported clearly.
10. Strict mode returns non-zero when requested operations fail.
