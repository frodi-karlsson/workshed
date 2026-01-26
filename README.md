# Workshed

Workshed is a tool for creating temporary, intent-scoped workspaces that bundle one or more Git repositories into a single directory.

One recommended use case is creating ad-hoc monorepos for focused work that spans multiple repositories.

Another is creating a clean, disposable workspace for a specific task without cluttering your main development directories. This is useful for exploratory work, refactors, and agent-driven development, where isolation and clear task boundaries matter more than long-lived branches.


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

```bash
# Create a workspace with purpose and repositories
workshed create --purpose "Debug payment timeout" --repo github.com/org/service@main --repo ./local-lib

# Create from template (copies template dir with variable substitution)
workshed create --purpose "New feature" --template ~/templates/node --map env=prod --map version=1.0

# Run commands in all repositories
workshed exec -- make test

# Add/remove repositories
workshed repo add my-workspace --repo https://github.com/org/new-repo@main
workshed repo remove my-workspace --repo new-repo

# List and manage workspaces
workshed list
workshed inspect
workshed path
workshed update --purpose "New focus"
workshed remove

# Or use the interactive dashboard
workshed
```

---

## Key Concepts

### Purpose
A workspace must have a purpose. This is stored as metadata and used for listing and discovery. The handle (e.g. `aquatic-fish-motion`) is randomly generated and not meaningful.

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

## Commands

| Command | Description |
|---------|-------------|
| `workshed` | Open interactive dashboard |
| `workshed create` | Create a new workspace |
| `workshed list` | List workspaces, filter by purpose |
| `workshed inspect` | Show workspace details |
| `workshed path` | Print workspace path |
| `workshed exec -- <cmd>` | Run command in repositories |
| `workshed repo add` | Add repository to workspace |
| `workshed repo remove` | Remove repository from workspace |
| `workshed update` | Update workspace purpose |
| `workshed remove` | Delete a workspace |
| `workshed --version` | Show version |

Run `workshed <command> --help` for detailed usage.

---

## Environment Variables

- `WORKSHED_ROOT` — workspace directory (default: `~/.workshed/workspaces`)
- `WORKSHED_LOG_FORMAT` — output format: `human`, `json`, or `raw` (default: `human`)

---

## Terminal UI (TUI)

Workshed includes optional interactive UI for workspace selection and purpose input. Active by default in human mode.

### When TUI Is Enabled

Default behavior when `WORKSHED_LOG_FORMAT` is unset or `human`. Set to `json` to disable.

### Features

- **Workspace selector** — Interactive list when auto-discovery fails (arrow keys/`j`/`k` to navigate, `Enter` to select)
- **Purpose autocomplete** — Suggestions from existing purposes as you type
- **Path completion** — Tab completes file/directory paths; ↑/↓ navigate suggestions; Esc dismisses
- **Dashboard** — Full interactive workspace management
- **Template support** — Configure template directory and variable substitution (`{{key}}` → value)

### Non-Interactive Fallback

When TUI is disabled (CI, piped output), Workshed falls back to plain text prompts and error messages.

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