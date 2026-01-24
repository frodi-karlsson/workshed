# Workshed
**Intent-scoped local workspaces for single- and multi-repo work**
---
## What Is This?
Workshed is an alternative to `git worktree` designed for intent-driven, multi-repository work.
Where `git worktree` gives you *another checkout of a repo*, Workshed gives you a **workspace**:
- Explicitly created for a purpose
- Able to contain **multiple repositories**
- Isolated on disk
- Safe for humans *and* agentic tools

Workshed is local-first, filesystem-backed, and deliberately boring.

---
## Why Not `git worktree`?
`git worktree` is single-repo, branch-centric, and silent about intent.
Workshed is **task-centric**, **multi-repo**, globally discoverable, and explicit about purpose.

If `git worktree` is "another branch checkout,"  
Workshed is **"open an investigation."**

---
## Example
```bash
workshed create \
  --purpose "Debug: Payment timeout across services" \
  --template ai-debug \
  --repo api=git@github.com:org/api@main \
  --repo billing=git@github.com:org/billing@v2.3.1 \
  --repo frontend=git@github.com:org/web@bugfix/timeout
```
This creates a **single workspace directory** containing multiple independent Git repositories, template files (`.claude/`, `.mcp.json`, notes), and minimal metadata.

The directory name is a random, human-readable handle (e.g. `aquatic-fish-motion`).

---
## Core Principles
**Intent is first-class**: Every workspace requires `--purpose`. Intent is stored, visible, and human-authored.

**Multi-repo by design**: A workspace may contain one or many repositories, co-located but not orchestrated.

**Agent-friendly by default**: No daemons, no hidden state. Safe, isolated sandboxes ideal for LLM experimentation.

**Filesystem-backed**: One directory = one workspace. No database. If it exists on disk, it exists.

**Workspace Handles**: Each workspace is identified by a short, human-readable handle generated from random words (e.g. aquatic-fish-motion). Handles are opaque identifiers, not descriptions. Meaning lives in --purpose, not the name.

---
## Templates
Templates are **plain directories** copied verbatim into new workspaces. No variables, no logic—just files.
```bash
workshed template create
# Populate ~/.workshed/templates/some-random-id/ with standard files

workshed template path some-random-id
# Print template directory path for inspection/editing
```

Templates standardize *environment*, not behavior.

---
## CLI
```bash
create      # Create workspace (errors on name collision)
list        # List all workspaces with fuzzy filtering
            # --purpose <substring>, --name <substring>, --repo <repo-url>
            # If used without flags, it will fuzzy-search names, repos and purposes
inspect     # Show metadata and repos
path        # Print workspace path (use: cd $(workshed path <name>))
            # Also supports fuzzy filtering like `list`, with
            # all the same flags, but will only match exact
            # handles without flags.
update      # Edit purpose
remove      # Delete workspace

template create ai-debug    # Create new template
template list               # List available templates
template path <name>        # Print template directory path
template inspect <name>     # Show template contents
template remove <name>      # Delete template
```

---
## What Workshed Is NOT
❌ Git wrapper or branch manager  
❌ CI/CD system or orchestration tool  
❌ Background service  
❌ Repository orchestrator (no cross-repo command execution)  
✅ Local infrastructure for **intent-scoped, multi-repo work**

---
## Roadmap

### v0: Foundation
- Single-repo workspaces
- Intent (`--purpose`) required
- Basic lifecycle: `create`, `list`, `inspect`, `path`, `remove`
- Filesystem-backed storage
- Name collision errors

### v1: Multi-Repo + Templates
- Multi-repo support (`--repo` flag)
- Template system (copy-only, no logic)
- `template` subcommands
- Enhanced `list` with filtering (purpose, repo, age)

### v1.x: Polish
- Shallow clone support
- Performance improvements for large workspace counts
- TUI for workspace browsing and management when no handle is provided to commands like `path` and `inspect`

---
## The Pitch
Workshed is boring filesystem + Git plumbing for modern workflows: parallel, exploratory, and agent-assisted.

It doesn't replace Git.  
It replaces chaos.

**Ship small. Learn. Iterate.**