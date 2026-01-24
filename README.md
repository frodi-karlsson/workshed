# Workshed

Workshed is a tool for creating temporary, intent-scoped workspaces that bundle one or more Git repositories into a single directory.

You can think of it as a way to create ad-hoc monorepos for focused work that spans more than one repository.

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
# (multi-repo support coming in v0.2)
workshed create \
  --purpose "Debug payment timeout across services" \
  --repo api=git@github.com:org/api@main

workshed list
workshed list --purpose payment

cd $(workshed path aquatic-fish-motion)

workshed exec -- git status
workshed exec -- git commit -m "WIP"
workshed exec -- git push

workshed remove aquatic-fish-motion
```

This creates a directory containing the cloned repository and a small metadata file describing the workspace.

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

### Batch execution (`exec`)
`workshed exec` runs a command in each repository directory:

- Commands run sequentially
- Repositories are processed in creation order
- Each command runs from the repository root
- Output is streamed
- Execution stops on the first non-zero exit code

There is no rollback, retry logic, or interpretation of results.

---

## Commands

- `create` — Create a workspace (requires `--purpose`)
- `list` — List workspaces with optional filtering
- `inspect` — Show workspace metadata and repositories
- `path` — Print the workspace path (for `cd $(...)`)
- `exec` — Run a command in each repository
- `remove` — Delete a workspace
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

### Implemented (v0.1)
- Workspace creation with required purpose
- Single-repository workspaces
- Listing and filtering by purpose
- Workspace inspection and path lookup
- Filesystem-backed storage

### Planned

**v0.2**
- Multiple repositories per workspace
- `exec` for sequential batch commands
- Updating workspace purpose
- Improved output and errors

**Later**
- Templates for workspace setup
- Additional filtering and discovery
- Optional concurrent execution

---

## Summary

Workshed is a small tool for organizing temporary, multi-repository work. It’s useful when a task spans several repositories and you want a clean, disposable workspace to work in.

It doesn’t replace Git. It just makes this kind of work easier to manage.
