# Architecture Overview

This document describes the overall architecture of workshed, explaining the guiding philosophy, module organization, and how components interconnect.

## Guiding Philosophy

**Simplicity births reliability.** The codebase favors clear, explicit code over clever abstractions. Complex behaviors emerge from simple parts working together. When in doubt, simplify.

**Composition over configuration.** Features compose from primitives rather than being configured into existence. This makes behavior predictable and testing straightforward.

**Explicit state over hidden state.** All state lives in fields, never in globals or closures. This eliminates implicit dependencies and makes the system easier to reason about.

**Single responsibility.** Each package does one thing well. The CLI parses commands; the TUI handles interaction; the workspace package manages persistence. Clear boundaries reduce coupling and simplify testing.

## Module Organization

```
workshed/
├── cmd/workshed/           # Entry point, command routing
├── internal/
│   ├── cli/               # CLI command handlers
│   ├── tui/               # Terminal user interface
│   ├── workspace/         # Domain logic and persistence
│   ├── logger/            # Structured logging
│   ├── handle/            # Handle generation and validation
│   └── ...
└── docs/architecture/     # This documentation
```

### Module Boundaries

**cmd/workshed**: The thinnest possible layer. Parses `os.Args` and routes to command functions. No business logic here.

**internal/cli**: CLI command implementations. Parses flags, validates inputs, calls workspace operations, formats output. Each command is a standalone function with no shared state.

**internal/tui**: Interactive terminal interface using Bubbletea. Self-contained; doesn't know about CLI parsing. Can be invoked from CLI commands when interactivity is beneficial.

**internal/workspace**: Domain logic. Knows about workspaces, repositories, operations. Nothing about presentation (CLI output or TUI rendering).

**internal/logger**: Structured logging with human/JSON/raw modes. Used by all modules.

## How Modules Connect

### CLI to Workspace

CLI commands inject dependencies and call workspace operations:

```
CLI (internal/cli) ──> Workspace (internal/workspace)
    GetOrCreateStore(l)      NewFSStore(path)
    store.Create(ctx, opts)  FSStore.Create(ctx, opts)
```

The CLI never imports TUI packages. When interactivity is needed, it falls back to a separate TUI invocation.

### CLI to TUI

The CLI can invoke the TUI as a separate program when interactivity helps:

```
CLI ──> tui.RunCreateWizard(ctx, store) ──> New bubble tea program
CLI ──> tui.TrySelectWorkspace(ctx, store, err, l) ──> User selection
```

The TUI runs in its own process with its own event loop. Results are passed back through return values.

### TUI to Workspace

The TUI uses the same store interface as CLI commands:

```
TUI (internal/tui) ──> Workspace (internal/workspace)
    StackModel.store        FSStore (interface)
    store.List(ctx, opts)   FSStore.List(ctx, opts)
```

Both layers depend on the same interface. This ensures consistent behavior regardless of how the user interacts.

### Workspace to Storage

The workspace package abstracts storage behind an interface:

```
Workspace (internal/workspace) ──> Storage
    FSStore                  Filesystem-based implementation
    Store interface          Abstracts persistence
```

Tests use in-memory or mock stores. The production implementation uses the filesystem.

## Data Flow

### Command Execution

1. User runs `workshed <command> [flags]`
2. `main.go` routes to command function
3. Command parses flags with `spf13/pflag`
4. Command validates required inputs
5. Command gets store via `GetOrCreateStore`
6. Command calls workspace operation
7. Workspace performs business logic
8. Workspace persists to storage
9. Command formats and outputs result

### Interactive Session

1. User runs `workshed` (no command)
2. `main.go` calls `RunStackModel`
3. TUI creates stack model with store
4. Bubble tea program runs event loop
5. User navigates, selects, creates
6. Views call store operations
7. View updates automatically
8. On exit, control returns to shell

### Workspace Discovery

When a command needs a workspace handle:

1. User provides handle explicitly, or
2. CLI attempts auto-discovery from current directory, or
3. CLI falls back to TUI selection (human mode only)

```go
handle := ResolveHandle(ctx, store, providedHandle, l)
// Returns providedHandle, auto-discovered handle, or exits
```

## Shared Abstractions

### Store Interface

Both CLI and TUI use the same store interface:

```go
type store interface {
    List(ctx context.Context, opts ListOptions) ([]*Workspace, error)
    Get(ctx context.Context, handle string) (*Workspace, error)
    Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
    UpdatePurpose(ctx context.Context, handle string, purpose string) error
    Remove(ctx context.Context, handle string) error
    Exec(ctx context.Context, handle string, opts ExecOptions) ([]ExecResult, error)
}
```

This interface is defined in `internal/workspace` and is imported by both CLI and TUI packages.

### Logger Interface

All modules log through `internal/logger`:

- `DEBUG` - Diagnostic information
- `INFO` - General information
- `HELP` - User guidance
- `SUCCESS` - Successful completion
- `ERROR` - Error conditions

Output format adapts to mode (human, JSON, raw).

### Context Passing

All operations accept `context.Context` for cancellation and timeouts:

```go
store.Create(ctx, opts)  // Respects ctx cancellation
```

This enables timeout handling for long-running operations like cloning repositories.

## Configuration

Configuration flows through environment variables:

- `WORKSHED_ROOT`: Root directory for workspaces (default: `~/.workshed/workspaces`)
- `WORKSHED_LOG_FORMAT`: Output format (`human`, `json`, `raw`)

No config files. No complex configuration. Environment variables compose naturally.

## Error Handling

Errors follow a consistent pattern:

1. Log with context using structured fields
2. Exit with appropriate code (0 for success, 1 for error)
3. Provide actionable messages

```go
if err != nil {
    l.Error("operation failed", "context", value, "error", err)
    exitFunc(1)
}
```

The TUI handles errors differently, showing them in a modal with recovery options.

## Testing Strategy

### Layer Boundaries

| Layer | Tests | Mocking |
|-------|-------|---------|
| CLI | Unit + integration | Mock store, captured output |
| TUI | Unit + e2e (teatest) | Mock store |
| Workspace | Unit + integration | In-memory store |

### Integration Points

- `cli_integration_test.go`: Complete CLI workflows
- `internal/tui/snapshot/`: Snapshot tests for TUI views
- `workspace/...`: Tests for persistence logic

### Test Isolation

Each test:
- Uses temporary directories
- Captures output
- Overrides exit function
- Cleans up after itself

## Design Smells to Avoid

- Global state
- Implicit dependencies
- Feature creep in established modules
- Coupling between presentation and domain layers
- Over-abstraction (creating interfaces before understanding the shape)

When these appear, the design has drifted. Refactor toward simplicity.

## Further Reading

- [CLI Architecture](cli.md) - Command structure, patterns, and conventions
- [TUI Architecture](tui.md) - View design, state management, and interaction patterns
- [Testing Architecture](testing.md) - Snapshot testing and test patterns
- [Workspace Package](../workspace.md) - Domain model and persistence
- [Logger Package](../logger.md) - Logging patterns and usage (if exists)