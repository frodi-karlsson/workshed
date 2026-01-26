# Testing Architecture

Design principles and architectural guidance for Workshed's testing approach.

## Guiding Philosophy

Tests exist to increase confidence, not to inflate numbers. Each test should verify one clear behavior. High coverage is a goal, but simplicity is the higher priority.

## Test Organization

### Layer Boundaries

| Layer | Tests | Mocking |
|-------|-------|---------|
| CLI | Unit + integration | Mock store, captured output |
| TUI | Unit + e2e (snapshot) | Mock store |
| Workspace | Unit + integration | In-memory store |

### Test Structure

```
internal/
├── cli/
│   ├── *_test.go           # Unit tests
│   └── cli_integration_test.go  # Integration tests
├── tui/
│   ├── snapshot/
│   │   ├── snaplib.go      # Test utilities and helpers
│   │   ├── *_snap_test.go  # Snapshot tests
│   │   └── __snapshots__/  # Generated snapshot files
│   └── *_test.go           # Unit tests
└── workspace/
    ├── *_test.go           # Unit tests
    └── integration_test.go # Integration tests
```

## Snapshot Testing

### Framework

The TUI uses snapshot testing via `go-snaps` to verify view state:

- **Framework**: `github.com/gkampitakis/go-snaps`
- **Helper library**: `internal/tui/snapshot/snaplib.go`
- **Test pattern**: Record view state, compare against stored snapshot

### Key Components

**`snaplib.go`** provides the testing DSL:

```go
// Create a scenario with mock store
scenario := snapshot.NewScenario(t, gitOpts, storeOpts)

// Add steps
scenario.Key("c", "Open create wizard")
scenario.Type("My project", "Enter purpose")
scenario.Enter("Confirm purpose")

// Record and match
output := scenario.Record()
snapshot.Match(t, "test_name", output)
```

**`Match(t, name, snapshot)`** writes to a file named after the test function. Multiple tests in the same file share a snapshot file with named sections:

```
[TestDashboardView_Navigation - 1]
first_selected
{ ... snapshot data ... }
---

[TestDashboardView_Navigation - 2]
second_selected
{ ... snapshot data ... }
```

### Scenario Configuration

**Store options**:
- `WithWorkspaces([]*Workspace)` - Pre-populate workspace list
- `WithCreateError(error)` - Simulate creation failure
- `WithCreateDelay(duration)` - Add delay for testing loading states
- `WithListError(error)` - Simulate list loading failure

**Git options**:
- `WithGitRemoteURL(url)` - Mock detected remote URL
- `WithGitRemoteError(error)` - Simulate git detection failure

### Running Tests

```bash
# Run all snapshot tests
go test ./internal/tui/snapshot/...

# Update snapshots after UI changes
UPDATE_SNAPS=true go test ./internal/tui/snapshot/...

# Run specific test
go test ./internal/tui/snapshot/... -run "TestDashboardView_Navigation"
```

### Adding a Snapshot Test

1. Create test in appropriate `*_snap_test.go` file
2. Set up scenario with store options
3. Add key/type/enter steps to reach desired state
4. Call `scenario.Record()` and `snapshot.Match(t, name, output)`

```go
func TestDashboardView_WithWorkspaces(t *testing.T) {
    scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
        snapshot.WithWorkspaces([]*workspace.Workspace{
            {Handle: "ws-1", Purpose: "First", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
        }),
    })

    output := scenario.Record()
    snapshot.Match(t, t.Name(), output)
}
```

### Snapshot Files

Snapshot files live in `__snapshots__/` directory:

- Named after test function: `TestDashboardView_Navigation.snap`
- Each file contains multiple named snapshots for related states
- Use subtest names for grouping: `TestDashboardView_Navigation/first_selected`

Format:
```
[TestName - N]
snapshot_name
{ json data }
---
```

### Best Practices

1. **One concept per test file**: Group related view states
2. **Descriptive snapshot names**: Use meaningful names, not just "view1", "view2"
3. **Test edge cases**: Empty states, error states, navigation flows
4. **Update intentionally**: Always review snapshot changes before committing

### Common Patterns

**Navigation testing**:
```go
output := scenario.Record()
snapshot.Match(t, "first_selected", output)

scenario.Key("j", "Navigate down")
output = scenario.Record()
snapshot.Match(t, "second_selected", output)
```

**Modal workflows**:
```go
scenario.Enter("Open context menu")
scenario.Key("i", "Select inspect")
output := scenario.Record()
snapshot.Match(t, "inspect_view", output)
```

**Error states**:
```go
scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
    snapshot.WithCreateError(errors.New("failed")),
})
```

## Unit Testing

### CLI Tests

Override dependencies for isolation:

```go
exitFunc = func(code int) { /* track exit */ }
outWriter = bytes.NewBuffer(nil)
store := NewMockStore()
```

### Workspace Tests

Use temporary directories and real operations:

```go
env := NewTestEnvironment(t)
store := env.CreateStore()
ws, err := store.Create(ctx, opts)
```

### TUI Unit Tests

Test view logic in isolation with mock store and captured output.

## Integration Testing

### CLI Integration

Test complete workflows with temporary filesystem:

```go
tmpDir := t.TempDir()
root := filepath.Join(tmpDir, "workspaces")
// Run commands, verify output
```

### Test Isolation

Each test:
- Uses temporary directories
- Captures output
- Overrides exit function
- Cleans up after itself

## References

| Topic | Location |
|-------|----------|
| Snapshot helpers | `internal/tui/snapshot/snaplib.go` |
| Snapshot tests | `internal/tui/snapshot/*_snap_test.go` |
| Test examples | `internal/tui/snapshot/dashboard_snap_test.go` |
| CLI tests | `internal/cli/*_test.go` |
| Workspace tests | `internal/workspace/*_test.go` |