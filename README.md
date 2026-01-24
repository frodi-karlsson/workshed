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

## Example

```bash
# Create a workspace for a specific task
workshed create \
  --purpose "Debug payment timeout across services" \
  --repo git@github.com:org/api@main \
  --repo git@github.com:org/worker@develop

# Or use local repositories with direct paths
workshed create \
  --purpose "Local development" \
  --repo /Users/dev/my-api \
  --repo /Users/dev/my-worker

# Or create a workspace using current directory (no cloning)
cd /path/to/your/project
workshed create --purpose "Quick exploration"

# Commands can use current directory to discover workspace
workshed exec -- make test          # Runs in workspace (auto-discovered)
workshed inspect                    # Shows workspace details
workshed path                       # Prints workspace path
workshed update --purpose "New focus"  # Updates workspace purpose
workshed remove                     # Removes workspace (with confirmation)

workshed list
```

This creates a directory containing cloned repositories and a small metadata file describing the workspace.

---

## How It Works

### Purpose
A workspace must be created with `--purpose`. This is stored as metadata and used for listing and discovery. The handle itself is not intended to be meaningful.

### Workspace layout
A workspace is a directory containing:
- A metadata file
- One subdirectory per repository

If the directory exists, the workspace exists.

### Multiple repositories
Repositories are cloned into subdirectories of the workspace. They are not coupled beyond being colocated.

### Current directory workspace
If no `--repo` is provided when creating a workspace, the current directory is used as the repository. This is useful for exploring existing projects without cloning.

### Auto-discovery
Most commands (`exec`, `inspect`, `path`, `update`, `remove`) can automatically find the workspace containing the current directory when no handle is provided. This means you can run commands directly from within a workspace without needing to remember or look up the handle.

```bash
# From within a workspace directory
workshed exec -- make test
workshed inspect
workshed path
workshed update --purpose "New focus"
```

### Batch execution (`exec`)
`workshed exec` runs a command in each repository:

- Commands run sequentially by default
- Repositories are processed in creation order
- Each command runs from the repository root
- Use `--repo <name>` to target a specific repository
- Use `--repo all` to run in all repositories (default)
- Use `--repo root` to run in workspace root
- Output is streamed with repository headers
- Execution stops on the first non-zero exit code
- If no handle is provided, uses the workspace containing the current directory

```bash
# Run in all repositories (auto-discovers workspace from current directory)
workshed exec -- make test

# Run in specific repository
workshed exec --repo api -- make build

# Run in workspace root
workshed exec --repo root -- make setup
```

There is no rollback, retry logic, or interpretation of results.

---

## Commands

- `create` — Create a workspace (requires `--purpose`, repos optional, defaults to current directory)
- `list` — List workspaces with optional filtering
- `inspect` — Show workspace details (handle optional, auto-discovers from current directory)
- `path` — Print the workspace path (handle optional, auto-discovers from current directory)
- `exec` — Run a command in repositories (handle optional, auto-discovers from current directory)
- `remove` — Delete a workspace (handle optional, auto-discovers from current directory)
- `update` — Update workspace purpose (handle optional, auto-discovers from current directory)
- `version` / `--version` — Show version information

---

## What Workshed Is Not

- A permanent monorepo solution
- A build or dependency management tool
- A CI/CD system
- A Git wrapper or branch manager
- A background service

Workshed does not try to understand your code or your workflow. It only manages directories and runs commands you explicitly ask for.

---

## Development Status

### Implemented (v0.2.0)
- Workspace creation with required purpose
- Multiple repositories per workspace
- `exec` for sequential batch commands
- Listing and filtering by purpose
- Workspace inspection and path lookup
- Filesystem-backed storage
- Updating workspace purpose

### Planned
- Templates for workspace setup
- Additional filtering and discovery
- Optional concurrent execution

---

## Development

### Build
```bash
make build
./bin/workshed --help
```

### Test
```bash
make test          # unit tests
make test-integration      # integration tests
make test-all    # all tests
make check       # lint + all tests
```

### Lint
```bash
make lint         # check
make lint-fix     # auto-fix issues
```

### Environment Variables
- `WORKSHED_ROOT` — workspace directory (default: `~/.workshed/workspaces`)
- `WORKSHED_LOG_FORMAT` — output format: `human`, `json`, or `raw` (default: `human`)

### Metadata
Workspaces are stored as directories containing a `.workshed.json` metadata file with workspace metadata (handle, purpose, list of repositories with URL/ref/name, creation time).

---

## Summary

Workshed is a small tool for organizing temporary, multi-repository work. It’s useful when a task spans several repositories and you want a clean, disposable workspace to work in.

It doesn’t replace Git. It just makes this kind of work easier to manage.
