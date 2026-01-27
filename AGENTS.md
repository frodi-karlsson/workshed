# AGENTS.md

## Running

Use the Makefile. `make check` before committing.

---

## CLI Commands

Commands are in `cmd/workshed/<command>/<command>.go`. Tests in `<command>_test.go`.

### Backward Compatibility

Keep flag shorthands for existing flags:
| Flag | Shorthand |
|------|-----------|
| `--all` | `-a` |
| `--yes` | `-y` |

### TUI Behavior

Running `workshed` without arguments opens the TUI dashboard. This requires a terminal (human mode).

---

## Philosophy

**Simplicity births reliability.** Clear code over clever code. No shortcuts.

---

## Code Guidelines

* Clear names over comments
* Small functions, single responsibility
* No hidden state
* Make invalid states unrepresentable

---

## Testing

Tests for confidence, not coverage. Easy to read, easy to delete.

---

## Design Smells

* Abstraction without need
* Complex test setups
* Comments explaining bad code
* Over-eager deduplication