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

### Artifacts

Workshed maintains several artifact types under `.workshed/`:

| Artifact | Location | Purpose |
|----------|----------|---------|
| Executions | `.workshed/executions/<ulid>/` | Durable records of command executions |
| Captures | `.workshed/captures/<ulid>/` | Descriptive snapshots of git state |
| Context | `.workshed/context.json` | Derived workspace metadata |

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
    AddRepository(ctx context.Context, handle string, repo RepositoryOption, invocationCWD string) error
    AddRepositories(ctx context.Context, handle string, repos []RepositoryOption, invocationCWD string) error
    RemoveRepository(ctx context.Context, handle string, repoName string) error

    // Execution records
    RecordExecution(ctx context.Context, handle string, record ExecutionRecord) error
    GetExecution(ctx context.Context, handle, execID string) (*ExecutionRecord, error)
    ListExecutions(ctx context.Context, handle string, opts ListExecutionsOptions) ([]ExecutionRecord, error)

    // State capture and apply
    CaptureState(ctx context.Context, handle string, opts CaptureOptions) (*Capture, error)
    ApplyCapture(ctx context.Context, handle string, captureID string) error
    PreflightApply(ctx context.Context, handle string, captureID string) (ApplyPreflightResult, error)
    GetCapture(ctx context.Context, handle, captureID string) (*Capture, error)
    ListCaptures(ctx context.Context, handle string) ([]Capture, error)

// Context derivation
    DeriveContext(ctx context.Context, handle string) (*WorkspaceContext, error)
 }
```

### Execution Records

Execution records are durable artifacts created during `exec` operations:

```
.workshed/executions/<ulid>/
├── record.json       # ExecutionRecord with command, results, timing
├── stdout/           # Per-repository stdout files
└── stderr/           # Per-repository stderr files
```

The record includes:
- Command executed and target scope
- Per-repository exit codes and duration
- Timestamps for start/completion
- Links to output files

### Captures

Captures are **descriptive snapshots, not authoritative state**. They document git state at a point in time without implying a restoration contract.

```
.workshed/captures/<ulid>/
└── capture.json      # GitRef list with commit, branch, status
```

Each capture includes:
- **Kind**: Describes why the capture exists (manual, execution, checkpoint)
- **SourceExecutionID**: Optional link to originating execution record
- **GitState**: Per-repository commit hash, branch, and dirty status
- **Metadata**: User-provided description, tags, custom fields

Captures are **descriptive snapshots, not authoritative state**. They document git state at a point in time without implying restoration will produce identical results.

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
├── types.go           # Extended types (ExecutionRecord, Capture, WorkspaceContext)
├── store.go          # FSStore implementation
├── store_test.go     # Unit tests
├── integration_test.go # Integration tests
└── testutil.go       # Test utilities
```

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

On failure at any step:
- Cleanup is automatic via `defer`
- No artifacts remain in workspace root
- Verified by integration tests

### Template Support

Workspaces can include a template directory that is copied into the workspace:

- **Template directory**: Contents are copied to workspace root
- **Variable substitution**: Use `{{key}}` in file/directory names, substitute with `--map key=value`
- **Merge behavior**: Repository clones take precedence over template files (repos overwrite conflicts)
- **Validation**: Template path must exist and be a directory (if provided)

### Repository Management

Repositories are typically cloned during workspace creation:

- Local paths: Cloned via git clone (not symlinked)
- Remote URLs: Cloned with specified ref (default: main)
- Repository name: Derived from URL (last path component, stripped .git)
- Validation: URL format, local path existence, git repository check

#### Adding Repositories to Existing Workspaces

`AddRepository()` and `AddRepositories()` add repositories to existing workspaces:

1. Validate repository URL and uniqueness (no duplicate URLs or names)
2. Clone repository into workspace directory
3. Append to workspace metadata

#### Removing Repositories

`RemoveRepository()` removes a repository from a workspace:

1. Remove repository directory from filesystem
2. Remove repository from workspace metadata
3. Handles orphaned directories gracefully (succeeds even if directory is already gone)

### Cleanup on Remove

`Remove()` deletes the entire workspace directory:

- Uses `os.RemoveAll()` on workspace path
- All cloned repositories are removed
- No separate cleanup of metadata and repos needed

### Execution Records

`RecordExecution()` creates durable artifacts for command invocations:

1. Create `.workshed/executions/<ulid>/` directory
2. Write `record.json` with execution metadata
3. Create `stdout/` and `stderr/` subdirectories
4. On completion, update record with exit codes and duration

Execution records are never deleted by workshed. They serve as an audit trail.

### State Capture

`CaptureState()` creates descriptive snapshots of git state:

1. Create `.workshed/captures/<ulid>/` directory
2. For each repository, record:
   - Commit hash (`git rev-parse HEAD`)
   - Current branch (`git branch --show-current`)
   - Working tree status (`git status --porcelain`)
3. Write `capture.json` with all GitRef data

Captures are **descriptive, not authoritative**. They document state without guaranteeing restoration will produce identical results.

### Apply with Preflight

`ApplyCapture()` restores git state with safety checks:

1. Call `PreflightApply()` to validate workspace state
2. Fail fast if preflight returns errors (non-interactive)
3. For each repository: `git checkout <commit>`

Preflight detects conditions that would block safe restoration:
- Dirty working trees (would lose uncommitted changes)
- Missing repositories (capture references nonexistent repos)
- HEAD mismatches (current state differs from capture)
- Git failures (checkout operation errors)

Preflight errors are structured for programmatic handling but do not prompt the user.

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
- Errors if --purpose is missing 
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