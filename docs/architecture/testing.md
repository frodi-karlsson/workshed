# Testing Architecture

Design principles for Workshed's testing approach.

## Guiding Philosophy

Tests exist to increase confidence, not to inflate numbers. Each test should verify one clear behavior. High coverage is a goal, but simplicity is the higher priority.

## Test Organization

| Layer | Tests | Characteristics |
|-------|-------|-----------------|
| CLI | Unit + integration | Mock store, captured output |
| TUI | Unit + e2e (snapshot) | Mock store |
| Workspace | Unit + integration | In-memory store |

## Snapshot Testing

The TUI uses snapshot testing to verify view state. Tests record view state and compare against stored snapshots.

### Pattern

1. Create a scenario with mock store and git configuration
2. Add steps (key presses, text input) to reach desired state
3. Record view output and match against stored snapshot

### Running Tests

```bash
# Run all snapshot tests
go test -timeout=5s ./internal/tui/...

# Update snapshots after UI changes
UPDATE_SNAPS=true go test -timeout=5s ./internal/tui/...
```

### Best Practices

1. **One concept per test file**: Group related view states
2. **Descriptive snapshot names**: Use meaningful names
3. **Test edge cases**: Empty states, error states, navigation flows
4. **Update intentionally**: Always review snapshot changes before committing

## Unit Testing

### CLI Tests

Override dependencies for isolation (exit function, output writer, mock store).

### Workspace Tests

Use temporary directories and real operations.

### TUI Unit Tests

Test view logic in isolation with mock store and captured output.

## Integration Testing

Test complete workflows with temporary filesystem. Each test uses temporary directories, captures output, overrides exit function, and cleans up after itself.

## Characteristics of Good Tests

- Easy to read and understand
- Easy to delete when no longer needed
- Verify one clear behavior
- Run quickly
- Independent of each other