# Workshed

Workshed is a tool for creating temporary, intent-scoped workspaces that bundle one or more Git repositories into a single directory.

One recommended use case is creating ad-hoc monorepos for focused work that spans multiple repositories.

Another is creating a clean, disposable workspace for a specific task without cluttering your main development directories. This is useful for exploratory work, refactors, and agent-driven development, where isolation and clear task boundaries matter more than long-lived branches.


---

## Installation

```bash
brew tap frodi-karlsson/homebrew-tap
brew install workshed
```

---

## What Is This?

Workshed helps organize work that touches multiple repositories. Instead of juggling multiple clones or remembering which repos belong to which task, you create a workspace with a clear purpose and put everything related to that task in one place.

Each workspace:
- Has a required, human-written purpose
- Can contain one or more Git repositories
- Lives in its own directory on disk
- Is identified by a random, human-readable handle (e.g. `aquatic-fish-motion`)
- Can be operated on as a unit via batch commands
- Is disposable — delete it when the task is complete

There is no daemon and no database. State is the filesystem.

---

## Why Not `git worktree`?

`git worktree` is designed for working with multiple branches of a single repository.

Workshed is designed for tasks that involve multiple repositories at the same time.

If `git worktree` is about checking out another branch, Workshed is about grouping related repositories for the duration of a task.

---

## Quick Start

Run `workshed` to open the interactive dashboard, or run commands directly:

```bash
# Create a workspace with purpose and repositories
workshed create --purpose "Debug payment timeout" --repo github.com/org/service@main --repo ./local-lib

# Create from template (copies template dir with variable substitution)
workshed create --purpose "New feature" --template ~/templates/node --map env=prod --map version=1.0

# Run commands in all repositories
workshed exec -- make test

# Capture state with intent before making changes
workshed capture --name "Starting point" --description "Initial workspace state"

# Add/remove repositories
workshed repos add my-workspace --repo https://github.com/org/new-repo@main
workshed repos remove my-workspace --repo new-repo
workshed repos remove my-workspace --repo new-repo --dry-run  # Preview removal

# Output formats: --format table|json (default varies by command)
workshed list --format json
workshed captures --format json
workshed export --format json

# List and manage workspaces
workshed list
workshed inspect
workshed path
workshed update --purpose "New focus"
workshed remove
workshed remove --dry-run  # Preview deletion
```

---

## Key Concepts

### Purpose
A workspace must have a purpose. This is stored as metadata and used for listing and discovery.

### Handle
Each workspace is identified by an auto-generated, random handle (e.g. `aquatic-fish-motion`). Handles are not meaningful — they simply provide a stable identifier for the workspace. The handle is displayed in command output and used for CLI commands that require a workspace reference.

### Layout
A workspace is a directory containing:
- A metadata file (`.workshed.json`)
- One subdirectory per cloned repository

If the directory exists, the workspace exists.

### Current Directory Mode
If no `--repo` is provided when creating, the current directory becomes the repository. No cloning needed for local projects.

### Auto-Discovery
Most commands work from within a workspace without specifying a handle — Workshed finds the enclosing workspace automatically.

---

## Use Cases

### Cross-Repository Refactoring
When a feature spans multiple services (e.g., updating an API contract shared across frontend and backend):

```bash
workshed create --purpose "Update user API contract" \
  --repo git@github.com:org/frontend.git \
  --repo git@github.com:org/backend.git \
  --repo git@github.com:org/api-client.git
```

### Agent-Driven Development
For Claude/AI agents working in isolated environments:

```bash
# Create clean workspace per task
workshed create --purpose "Fix auth bug in payment flow" --repo ./current-project

# Agent works in workspace, can be discarded after
workshed exec -- "npm test && git diff"
workshed remove
```

### Exploratory Work
Trying changes without affecting your main development setup:

```bash
workshed create --purpose "Experiment with new auth library" --repo ./my-app
cd $(workshed path)
# ... experiment ...
workshed remove  # Clean slate when done
```

### Monorepo Without the Monorepo
Group related repositories for focused work:

```bash
workshed create --purpose "Q3 OKR delivery" \
  --repo git@github.com:org/orders.git@feature/q3 \
  --repo git@github.com:org/inventory.git@feature/q3 \
  --repo git@github.com:org/billing.git@feature/q3
```

### Templates for Standard Setups
Create workspaces from templates with variable substitution:

```bash
# Template directory: ~/templates/react-app
#   {{name}}/package.json -> my-app/package.json
workshed create --purpose "New SPA" \
  --template ~/templates/react-app \
  --map name=my-app \
  --map env=production
```

## State & Context Management

Workshed provides lightweight primitives for capturing and deriving workspace context. These features are designed to support exploratory, parallel, and agent-assisted workflows — not to guarantee perfect reproducibility.

### Captures

A capture records observed git state (commit, branch, dirty status) for all repositories in a workspace. Captures are **descriptive snapshots, not authoritative checkpoints**. They document what was, without guaranteeing what will be.

Every capture requires intent: provide `--kind`, `--description`, or `--tag` to clarify why the capture exists.

### Apply

The `apply` command attempts to restore git state from a capture. It uses non-interactive, fail-fast preflight checks to detect blocking conditions:

- **Dirty working trees** — would lose uncommitted changes
- **Missing repositories** — capture references repos not in the workspace

Apply does not enforce HEAD matching or historical correctness. It attempts the operation and reports success or failure.

### Export

The `export` command produces machine-readable workspace configuration as JSON. The output includes:

- Repository URLs and refs
- Workspace purpose
- Capture metadata (not git state)

This output is designed for sharing workspace configurations with teammates or recreating workspaces.

### Example Usage

```bash
# Export workspace configuration
workshed export > workspace.json

# Export and use with jq
workshed export | jq '.repositories'

# Import to create a new workspace
workshed import workspace.json

# Import with original handle
workshed import workspace.json --preserve-handle

# Capture current state with intent
workshed capture --name "Before refactor" \
  --description "API change point"

# List captures in JSON format for scripting
workshed captures --format json | jq '.[].name'

# Attempt to restore a previous capture
workshed apply 01HVABCDEFG
workshed apply --name "Before refactor"
workshed apply my-workspace --name "Starting point"

# List captures
workshed captures
```

---

## Commands

| Command | Description |
|---------|-------------|
| `workshed` | Open interactive dashboard |
| `workshed create` | Create a new workspace |
| `workshed list` | List workspaces, filter by purpose, paginate with --page and --page-size |
| `workshed inspect` | Show workspace details |
| `workshed path` | Print workspace path |
| `workshed exec -- <cmd>` | Run command in repositories (use -a flag for all repos) |
| `workshed capture` | Record a descriptive snapshot of git state |
| `workshed captures` | List captures for a workspace, filter by repository or branch with --filter |
| `workshed apply` | Attempt to restore git state from a capture |
| `workshed export` | Export workspace configuration |
| `workshed import` | Create workspace from exported JSON |
| `workshed health` | Check workspace health |
| `workshed repos list` | List repositories in a workspace |
| `workshed repos add` | Add repository to workspace |
| `workshed repos remove` | Remove repository from workspace |
| `workshed update` | Update workspace purpose |
| `workshed remove` | Delete a workspace |
| `workshed --version` | Show version |

Run `workshed <command> --help` for detailed usage.

---

## Environment Variables

- `WORKSHED_ROOT` — workspace directory (default: `~/.workshed/workspaces`)
- `WORKSHED_LOG_FORMAT` — logger format: `human`, `json`, or `raw` (default: `human`)

## Output Formats

Most commands support `--format table|json` for structured output:

| Command | Default | Notes |
|---------|---------|-------|
| `list` | table | Lists workspaces; supports table, json, raw |
| `captures` | table | Lists captures; supports table, json, raw |
| `capture` | table | Shows capture details |
| `apply` | table | Shows applied capture |
| `create` | table | Shows created workspace info |
| `inspect` | table | Shows workspace details; supports table, json, raw |
| `path` | raw | Shows workspace path; supports raw, table, json |
| `repos list` | table | Lists repositories; supports table, json, raw |
| `repos add` | table | Shows added repositories |
| `repos remove` | table | Shows removed repository |
| `update` | table | Shows updated purpose |
| `export` | table | Auto-detects from `--output` extension |
| `exec` | stream | Raw command output by default |

JSON output is designed for scripting and automation.

---

## Output

Workshed provides structured output for scripting:

- `--format table` - Human-readable table output (default)
- `--format json` - JSON for scripting and automation
- `--format raw` - Minimal output for parsing

JSON mode disables interactive prompts. Set `WORKSHED_LOG_FORMAT=json` for fully non-interactive use.

---

## What Workshed Is Not

- A permanent monorepo solution
- A build or dependency management tool
- A CI/CD system
- A Git wrapper or branch manager
- A background service

Workshed only manages directories and runs commands you explicitly provide.

---

## Development

See [docs/architecture/](docs/architecture/) for detailed documentation.

| Task | Command |
|------|---------|
| Build | `make build` |
| Unit tests | `make test` |
| Integration tests | `make test-integration` |
| TUI e2e tests | `make test-e2e` |
| Lint | `make lint` |
| Full check | `make check` |

---

## Architecture

For details on how Workshed is organized and why:

- [Architecture Overview](docs/architecture/index.md) — Guiding philosophy and module organization
- [CLI Architecture](docs/architecture/cli.md) — Command patterns and conventions
- [TUI Architecture](docs/architecture/tui.md) — View design and interaction patterns

---

## Summary

Workshed is a small tool for organizing temporary, multi-repository work. It's useful when a task spans several repositories and you want a clean, disposable workspace to work in.

It doesn't replace Git. It just makes this kind of work easier to manage.