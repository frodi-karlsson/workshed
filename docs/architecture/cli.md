# CLI Architecture

Design principles and architectural guidance for the workshed command-line interface.

## Guiding Philosophy

The CLI prioritizes scriptability, clarity, and composition. It should work well in automation pipelines while remaining approachable for interactive use. Simple operations should stay simple; complex operations should have clear paths.

## Core Architecture

### Command Pattern

Each command follows a consistent structure:

```
Create logger -> Parse flags -> Validate inputs -> Execute -> Handle errors
```

Commands are functions in `internal/cli/`:
- `Create`, `List`, `Inspect`, `Path`, `Exec`, `Remove`, `Update`, `Repos`
- `Capture`, `Apply`, `Export`, `Captures`, `Health`

### Dependencies

Commands access functionality through thin wrappers:

```go
store := GetOrCreateStore(l)
ctx := context.Background()
```

This makes testing straightforward and behavior predictable.

## Design Principles

### Explicit Over Implicit

Flags are explicit. Arguments follow conventions. There are no hidden behaviors.

### Fail Fast

Validate inputs early. Exit with clear error messages and appropriate codes.

```go
if *purpose == "" {
    l.Error("missing required flag", "flag", "--purpose")
    fs.Usage()
    exitFunc(1)
}
```

### Composition

Commands compose primitive operations from the workspace package. No business logic in CLI handlers.

## Command Structure

### Standard Pattern

Each command function follows:

1. Create logger with command name
2. Initialize flag set with command-specific options
3. Define usage function
4. Parse arguments
5. Validate required flags
6. Get/create store
7. Execute business logic
8. Handle errors

### Flag Handling

Using `spf13/pflag` for POSIX-compliant parsing:

```go
fs := flag.NewFlagSet("create", flag.ExitOnError)
purpose := fs.String("purpose", "", "Purpose of the workspace (required)")
fs.StringSliceVarP(&repoFlags, "repo", "r", nil, "Repository URL")
```

### Argument Validation

Validation occurs at multiple levels:
- Flag parsing validation
- Business logic validation
- Repository URL validation

### Repository Management Commands

The `repos` command manages repositories within existing workspaces:

```
workshed repos add <handle> --repo url[@ref]...   # Add repositories
workshed repos remove <handle> --repo <name>      # Remove repository
```

The `Repo` function dispatches to subcommands:

```go
func (r *Runner) Repo(args []string) {
    subcommand := args[0]
    switch subcommand {
    case "add":
        r.RepoAdd(args[1:])
    case "remove":
        r.RepoRemove(args[1:])
    }
}
```

Both subcommands follow the standard pattern but add validation for repository uniqueness (no duplicate URLs or names) and clone/remove operations.

### State Management Commands

Workshed provides commands for state management:

- **capture**: Creates descriptive snapshots of git state. Captures are not authoritative state records; they document the current state of repositories at a point in time.
- **captures**: Lists captures for a workspace, optionally output as JSON or reversed order.
- **apply**: Restores git state from a capture. Includes preflight validation to detect conditions that would block safe restoration.

  Syntax:
  ```
  workshed apply [<handle>] <capture-id>
  workshed apply [<handle>] --name <capture-name>
  ```

  Examples:
  ```
  workshed apply 01HVABCDEFG
  workshed apply --name "Before refactor"
  workshed apply my-workspace 01HVABCDEFG
  workshed apply my-workspace --name "Starting point"
  ```

- **export**: Generates workspace configuration as JSON, including repository URLs, refs, and purpose.
- **health**: Checks workspace health, reporting issues like stale executions (>30 days).

These commands compose primitive operations from the workspace package.

## Error Handling and Logging

### Preflight Validation

Apply operations use non-interactive preflight validation to detect blocking conditions before attempting state restoration:

- Dirty working trees (would lose uncommitted changes)
- Missing repositories (capture references repos not in workspace)
- HEAD mismatches (current state differs from capture)

Preflight failures return structured errors describing what would block a safe apply, without prompting the user.

### Structured Logger

The `internal/logger` package provides dual-mode logging:

- **Human mode**: Formatted, readable output for interactive use
- **JSON mode**: Structured output for automation
- **Raw mode**: Minimal output for scripts

### Log Levels

- `DEBUG` - Diagnostic information
- `INFO` - General information
- `HELP` - User guidance
- `SUCCESS` - Successful completion
- `ERROR` - Error conditions

### Exit Patterns

```go
l.Error("workspace creation failed", "purpose", opts.Purpose, "error", err)
exitFunc(1)
```

Errors include context. Exit codes indicate success/failure.

## CLI-TUI Integration

### Dual Interface

The CLI provides both automation-friendly and interactive interfaces:

- **CLI flags**: For scripts and automation
- **TUI dashboard**: For interactive use - run `workshed` to open the dashboard, then press 'c' to create a workspace

### Fallback Mechanism

When a workspace handle is not provided:

1. Auto-discover from current directory
2. In human mode: Offer TUI selection via `tui.TrySelectWorkspace`
3. In automated mode: Return error

### Shared Logic

Both paths converge to the same business logic in `internal/workspace`:

```go
ws, err := store.Create(ctx, workspace.CreateOptions{
    Purpose:      purpose,
    Repositories: repos,
})
```

## Testing Strategy

### Test Organization

| Category | Location | Coverage |
|----------|----------|----------|
| Unit | `*_test.go` | Command logic, helpers |
| Integration | `*_integration_test.go` | Complete workflows |

### Mocking Patterns

Override `exitFunc`, `outWriter`, `errWriter` for testing:

```go
exitFunc = func(code int) { /* track exit */ }
outWriter = bytes.NewBuffer(nil)
```

### Common Tests

- Flag parsing and validation
- Error handling and exit codes
- Output format verification
- Edge cases (empty inputs, invalid values)

## Development Guidelines

### Adding a Command

1. Create `internal/cli/<command>.go` with a `<Command>` function
2. Follow the standard pattern: logger, flags, validate, execute, error
3. Wire in `cmd/workshed/main.go` routing
4. Add tests for success and error paths
5. Update usage text in `root.go` if needed

### Flag Conventions

- Use short and long flags: `-r, --repo`
- Provide sensible defaults
- Include clear descriptions

### Output Conventions

- Use logger for status messages
- Use `Output` struct with `OutputRenderer` for tabular/JSON output
- Commands should accept `--format table|json` flag
- The `OutputRenderer` dispatches to `FlexTableRenderer` (table) or `JSONRenderer` (json)

### Error Messages

Be specific and actionable:

```
Error: missing required flag --purpose
Usage: workshed create --purpose <purpose> [--repo url@ref]... [--template <dir>] [--map key=value]...
```

### Template Support

The `create` command supports template directories with variable substitution:

```
workshed create --purpose "Task name" --template /path/to/template --map key=value
```

- `--template` copies a directory into the workspace
- `--map` provides variables for substitution (`{{key}}` in filenames and content is replaced with value)
- Multiple `--map` flags can be provided for multiple variables

The workspace store handles template copying in `internal/workspace/store.go` via `Create` method.

## Aspirations

The CLI aims to:

- Work reliably in scripts and pipelines
- Provide clear feedback at each step
- Support both human and automated use cases
- Handle errors gracefully with actionable messages
- Remain lightweight with minimal dependencies

Implementation should move toward these goals incrementally.

## References

| Topic | Location |
|-------|----------|
| Command routing | `cmd/workshed/main.go` |
| Shared utilities | `internal/cli/root.go` |
| Workspace operations | `internal/workspace/` |
| Logging | `internal/logger/` |
| Integration tests | `internal/cli/cli_integration_test.go` |