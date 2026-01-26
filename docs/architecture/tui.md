# TUI Architecture

Design principles for the workshed terminal user interface.

## Guiding Philosophy

The TUI prioritizes clarity and discoverability. Users should always understand where they are and what they can do. The interface should feel consistent across different views.

## Core Framework

Built on **Bubbletea** following the Elm architecture:

- **Model**: Explicit state in fields
- **Update**: State transitions via messages
- **View**: Render state to string

Simple patterns produce predictable behavior.

## View Stack Architecture

### StackModel

The TUI manages views as a stack:

```
StackModel
├── stack []View    # Views pushed/popped during navigation
└── store Store     # Shared data access
```

The top view handles messages. It can:
- Push a new view (`ViewResult.NextView`)
- Perform stack action (`ViewResult.Action`: pop, dismiss, etc.)

### View Interface

All views implement a consistent interface:

```go
type View interface {
    Init() tea.Cmd           // Initial command
    Update(msg) (ViewResult, tea.Cmd)
    View() string
    OnPush()                 // Called when pushed
    OnResume()               // Called when returned to
    IsLoading() bool         // Show spinner
    Cancel()                 // Cancel ongoing work
    Snapshot() interface{}   // Capture state
}
```

This pattern keeps views decoupled and testable.

## View Types

### DashboardView

The entry point. Shows workspaces in a list with filtering and navigation.

### WizardView

Multi-step creation flow. Handles user input, git detection, and repository cloning.

### ContextMenuView

Actions menu for a selected workspace: inspect, path, exec, update, remove.

### Modal Views

Focused overlays: Inspect, Path, Exec, Update, Remove, Error.

Each modal is self-contained with its own state and rendering.

## Key Characteristics

### Composition Over Inheritance

Views compose from simpler components. The stack manages navigation; individual views handle their domain logic.

### Explicit State

All state lives in fields. No globals, no hidden variables. This makes behavior traceable and testing straightforward.

### Consistent Patterns

Similar behaviors use similar patterns. Lists navigate the same way. Modals dismiss the same way. This reduces cognitive load for users and developers.

## Key Bindings

### Dashboard
- `c` - Create workspace
- `?` - Help
- `l` - Filter
- Navigation: arrows or `j`/`k`
- `Enter` - Open context menu

### Context Menu
- `i/p/e/u/r` - Actions
- `Esc` - Dismiss

### Modals
- `Enter` - Confirm
- `Esc/Ctrl+C` - Cancel

## Testing

The TUI uses snapshot testing to verify view states:

- Framework: `go-snaps` with custom DSL in `internal/tui/snapshot/snaplib.go`
- Tests in `internal/tui/snapshot/`
- Each test records view state and compares against stored snapshot

See [Testing Architecture](testing.md) for details.

## Development Principles

### Adding a View

1. Implement the `View` interface
2. Keep state explicit in fields
3. Return `ViewResult` for navigation
4. Add snapshot test

### Styling

- Use semantic colors
- Apply consistent spacing
- Follow help text format: `[Key] Action`

### Keep It Simple

If a view becomes complex, split it. If behavior isn't clear, simplify the design.

## Aspirations

The TUI aims to feel:
- **Responsive** - Quick to start, snappy navigation
- **Discoverable** - Clear help, consistent shortcuts
- **Reliable** - Errors show clearly, recovery is obvious
- **Approachable** - Simple for new users, efficient for power users

Implementation moves toward these goals incrementally.

## References

| Topic | Location |
|-------|----------|
| Stack model | `internal/tui/stack.go` |
| View interface | `internal/tui/views/view.go` |
| Views | `internal/tui/views/` |
| Snapshot testing | `internal/tui/snapshot/` |
| Testing guide | [Testing Architecture](testing.md) |