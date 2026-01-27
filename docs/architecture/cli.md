# CLI Architecture

Commands are implemented in `cmd/workshed/<command>/` using Cobra.

| Command | Location |
|---------|----------|
| create, list, inspect, path | `cmd/workshed/<command>/<command>.go` |
| repos (add, list, remove) | `cmd/workshed/repos/*.go` |
| capture, captures, apply | `cmd/workshed/<command>/<command>.go` |
| exec, export, health | `cmd/workshed/<command>/<command>.go` |
| import | `cmd/workshed/importcmd/` |
| remove, update, completion | `cmd/workshed/<command>/<command>.go` |

All commands use `internal/cli/runner.go` for shared functionality (store, logger, handle resolution).

See `AGENTS.md` for CLI development guidelines. |