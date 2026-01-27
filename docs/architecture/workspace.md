# Workspace Architecture

Design principles for the workshed workspace layer.

## Guiding Philosophy

The workspace package embodies the core workshed principles:

- **Simplicity births reliability.** The workspace layer manages persistence with minimal abstraction.
- **Explicit state over hidden state.** Workspace metadata and repository configurations are stored explicitly.
- **Single responsibility.** The package handles workspace lifecycle and repository management only; presentation is delegated to CLI and TUI.

## Core Architecture

### Data Model

```
Workspace
├── Handle        (unique identifier, auto-generated)
├── Purpose       (user-defined description)
├── Repositories  (list of Repository entries)
│   └── Repository
│       ├── URL   (git URL or local path)
│       ├── Ref   (branch/tag/ref, optional)
│       └── Name  (derived from URL)
├── Version       (metadata schema version)
└── CreatedAt     (timestamp)

Workspace metadata stored in: <workspace>/.workshed.json
```

### Artifacts

Workshed maintains several artifact types under `.workshed/`:

| Artifact | Purpose |
|----------|---------|
| Executions | Durable records of command executions |
| Captures | Descriptive snapshots of git state |

### Store Interface

The Store interface abstracts persistence operations. It defines methods for workspace lifecycle (Create, Get, List, Remove), repository management (AddRepository, RemoveRepository), execution (Exec, RecordExecution), and state management (CaptureState, ApplyCapture, ListCaptures).

Context is derived dynamically and not stored persistently.

### Handle Generation

Handles are unique, generated identifiers:
- Generated using handle package
- 8-character base36 encoding (alphanumeric)
- Collision-avoidance through existence checking
- User-friendly and script-friendly (no special characters)

## Operational Guarantees

### Atomic Creation

Workspace creation is atomic:
1. Validate inputs (purpose, repositories, template)
2. Generate unique handle
3. Create temporary directory
4. Write metadata to temp directory
5. Copy template directory (if provided)
6. Clone repositories into temp directory
7. Rename temp directory to final location

On failure at any step, cleanup is automatic and no artifacts remain in workspace root.

### Template Support

Workspaces can include a template directory that is copied into the workspace:
- **Template directory**: Contents are copied to workspace root
- **Variable substitution**: Use `{{key}}` in file/directory names, substitute with `--map key=value`
- **Merge behavior**: Repository clones take precedence over template files

### Repository Management

Repositories are typically cloned during workspace creation:
- Local paths: Cloned via git clone (not symlinked)
- Remote URLs: Cloned with specified ref (default: main)
- Repository name: Derived from URL (last path component, stripped .git)

Adding repositories to existing workspaces:
1. Validate repository URL and uniqueness
2. Clone repository into workspace directory
3. Append to workspace metadata

Removing repositories:
1. Remove repository directory from filesystem
2. Remove repository from workspace metadata
3. Handles orphaned directories gracefully

### Execution Records

Execution records are durable artifacts created during `exec` operations. They include command executed, per-repository exit codes and duration, timestamps, and links to output files.

### State Capture

Captures create descriptive snapshots of git state: commit hash, current branch, working tree status. Captures are **descriptive, not authoritative**. They document state without guaranteeing restoration will produce identical results.

### Apply with Preflight

ApplyCapture restores git state with safety checks. Preflight detects conditions that would block safe restoration: dirty working trees, missing repositories, HEAD mismatches, git failures.

## Error Handling

### Error Patterns

| Category | Pattern | Example |
|----------|---------|---------|
| Validation | Direct error return | `errors.New("purpose is required")` |
| Not found | Formatted message | `fmt.Errorf("workspace not found: %s", handle)` |
| Operation | Wrapped error | `fmt.Errorf("reading metadata: %w", err)` |
| Git | Classified with hint | `classifyGitError()` |

### Git Error Classification

The `classifyGitError()` function provides actionable error messages for common git failures:
- `repository not found`: URL incorrect or inaccessible
- `authentication failed`: Permission issues
- `network error`: Connectivity problems
- `ref not found`: Branch/tag does not exist

## Context Usage

All operations accept `context.Context` for cancellation and timeouts. This enables handling Ctrl+C during long operations and configurable timeouts for clone operations.

## Dependencies

The workspace package depends on:
- `internal/handle`: Handle generation and validation
- `internal/logger`: Structured logging (used sparingly)
- Standard library: `context`, `os`, `path/filepath`, `encoding/json`, `exec`

No CLI or TUI imports. The workspace layer is presentation-agnostic.

## CLI Integration

CLI commands use the workspace package through the store interface. CLI handles flag parsing, validation, errors, output formatting, and exit codes. Workspace handles business logic (validation, cloning, persistence) and error semantics.

## TUI Integration

TUI uses the same store interface as CLI. Both layers call identical methods for workspace enumeration, details, creation, and execution.

## Testing Strategy

### Unit Tests

Focus on validation logic, URL parsing, and edge cases. Use temporary directories and mock-free testing.

### Integration Tests

Verify real filesystem behavior: atomic creation, concurrent creation, special characters, cleanup verification.

### Test Utilities

Helpers provide test setup, store creation, git repo creation for testing, and assertions.

## Design Decisions

### Why Filesystem-Based Storage?

Simplicity. The filesystem is the simplest possible storage layer with no external dependencies and no migration concerns.

### Why Descriptive Captures, Not Authoritative?

Captures document what was, without guaranteeing what will be. Git operations are complex and restoration may fail or produce unexpected results. Users should understand captures are informational, not a backup mechanism.

### Why No Hidden State?

Explicit state makes the system easier to reason about, test, and debug. No globals or closures means no implicit dependencies between operations.

## Development Guidelines

### Adding an Operation

1. Define request/response types
2. Add method to Store interface
3. Implement in FSStore
4. Add unit tests

### Repository Validation

When adding repository sources:
1. Update path/URL detection
2. Add URL validation
3. Add extraction logic for repository name
4. Update cloning logic if different handling needed

### Error Messages

Follow existing patterns: be specific (include handle, path, URL), use wrapped errors for operation failures, provide actionable context for git errors.