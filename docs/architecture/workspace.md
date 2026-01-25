# Workspace Architecture

Design principles and architectural guidance for the workshed workspace layer.

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

### Store Interface

The `Store` interface abstracts persistence operations:

```go
type Store interface {
    Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
    Get(ctx context.Context, handle string) (*Workspace, error)
    List(ctx context.Context, opts ListOptions) ([]*Workspace, error)
    Remove(ctx context.Context, handle string) error
    Path(ctx context.Context, handle string) (string, error)
    UpdatePurpose(ctx context.Context, handle string, purpose string) error
    FindWorkspace(ctx context.Context, dir string) (*Workspace, error)
    Exec(ctx context.Context, handle string, opts ExecOptions) ([]ExecResult, error)
}
```

Implementations:
- `FSStore`: Filesystem-based store (primary implementation)
- Tests use in-memory implementations

### Handle Generation

Handles are unique, generated identifiers for workspaces:

- Generated using `handle.NewGenerator()`
- 8-character base36 encoding (alphanumeric)
- Collision-avoidance through existence checking
- User-friendly and script-friendly (no special characters)

## Module Organization

```
internal/workspace/
├── workspace.go       # Core types (Workspace, Repository, Store interface)
├── store.go          # FSStore implementation
├── store_test.go     # Unit tests
├── integration_test.go # Integration tests
└── testutil.go       # Test utilities
```

## Operational Guarantees

### Atomic Creation

Workspace creation is atomic:

1. Validate inputs (purpose, repositories)
2. Generate unique handle
3. Create temporary directory
4. Write metadata to temp directory
5. Clone repositories into temp directory
6. Rename temp directory to final location

On failure at any step:
- Cleanup is automatic via `defer`
- No artifacts remain in workspace root
- Verified by integration tests

### Repository Management

Repositories are cloned during workspace creation:

- Local paths: Copied via git clone (not symlinked)
- Remote URLs: Cloned with specified ref (default: main)
- Repository name: Derived from URL (last path component, stripped .git)
- Validation: URL format, local path existence, git repository check

### Cleanup on Remove

`Remove()` deletes the entire workspace directory:

- Uses `os.RemoveAll()` on workspace path
- All cloned repositories are removed
- No separate cleanup of metadata and repos needed

## Error Handling

### Error Patterns

Errors follow consistent patterns:

```go
// Context-wrapped errors for operation failures
return fmt.Errorf("reading metadata: %w", err)

// Sentinel-style errors for missing resources
return fmt.Errorf("workspace not found: %s", handle)

// Validation errors (early returns)
if purpose == "" {
    return errors.New("purpose is required")
}
```

### Error Categories

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

All operations accept `context.Context` for:

- Cancellation (Ctrl+C during long operations)
- Timeouts (clone timeout in CLI, configurable)
- Future: Resource limits, deadlines

```go
createCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()
ws, err := s.Create(createCtx, opts)
```

## Dependencies

The workspace package depends on:

- `internal/handle`: Handle generation and validation
- `internal/logger`: Structured logging (used sparingly)
- Standard library: `context`, `os`, `path/filepath`, `encoding/json`, `exec`

No CLI or TUI imports. The workspace layer is presentation-agnostic.

## CLI Integration

CLI commands use the workspace package through the store interface:

```go
// From internal/cli/create.go
opts := workspace.CreateOptions{
    Purpose:      purpose,
    Repositories: repos,
}
ws, err := s.Create(createCtx, opts)
```

CLI handles:
- Flag parsing and validation
- Interactive TUI fallback when purpose is missing
- Output formatting and exit codes

Workspace handles:
- Business logic (validation, cloning, persistence)
- Error semantics

## TUI Integration

TUI uses the same store interface as CLI:

```go
// From internal/tui/dashboard.go
func NewDashboardModel(ctx context.Context, store store.Store) dashboardModel
```

Both layers call identical methods:
- `store.List()` for workspace enumeration
- `store.Get()` for details
- `store.Create()` for wizard completion
- `store.Exec()` for command execution

## Testing Strategy

### Unit Tests

| Coverage | Location |
|----------|----------|
| Validation logic | `store_test.go` |
| URL parsing | `store_test.go` |
| Edge cases | `store_test.go` |

Unit tests use temporary directories and mock-free testing.

### Integration Tests

| Coverage | Location |
|----------|----------|
| Atomic creation | `integration_test.go` |
| Concurrent creation | `integration_test.go` |
| Special characters | `integration_test.go` |
| Cleanup verification | `integration_test.go` |

Integration tests verify real filesystem behavior.

### Test Utilities

`testutil.go` provides helpers:

- `WorkspaceTestEnvironment`: Full test setup
- `CreateTestStore()`: Store with temp root
- `CreateLocalGitRepo()`: Git repo for testing
- `MustHaveFile()`, `MustNotHaveTempDirs()`: Assertions

## Development Guidelines

### Adding an Operation

1. Define request/response types in `workspace.go`
2. Add method to `Store` interface
3. Implement in `FSStore`
4. Add unit tests
5. Update this document

### Repository Validation

When adding repository sources:

1. Update `isLocalPath()` detection
2. Add URL validation in `validateRepoURL()`
3. Add extraction logic in `extractRepoName()`
4. Update cloning logic if different handling needed

### Error Messages

Follow existing patterns:

- Be specific (include handle, path, URL)
- Use wrapped errors for operation failures
- Provide actionable context for git errors

## Aspirations

The workspace layer aims to:

- Remain free of CLI or TUI concerns
- Support multiple store implementations (SQLite, cloud storage)
- Provide richer workspace metadata (tags, last accessed, etc.)
- Enable workspace templates and snapshots

Implementation moves toward these goals incrementally.

## References

| Topic | Location |
|-------|----------|
| Core types | `internal/workspace/workspace.go` |
| Store implementation | `internal/workspace/store.go` |
| Unit tests | `internal/workspace/store_test.go` |
| Integration tests | `internal/workspace/integration_test.go` |
| Handle generation | `internal/handle/` |
| Architecture overview | `index.md` |