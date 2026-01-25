# TUI Architecture

Design principles and architectural guidance for the workshed terminal user interface.

## Guiding Philosophy

The TUI prioritizes clarity, consistency, and discoverability. Users should always understand where they are, what they can do, and how to do it. The interface should feel familiar across different views while accommodating domain-specific interactions.

## Core Framework

Built on the **Charm ecosystem** (Bubbletea, Bubbles, Lipgloss) following the Elm architecture pattern:

- **Model**: Holds all state explicitly
- **Update**: Pure state transitions in response to messages
- **View**: Renders state as a string
- **Message**: User input or async events

This pattern ensures predictability and testability.

## Architectural Layers

### Views (State Machine)

The TUI operates as a state machine where `currentView` determines behavior:

- **Dashboard**: Primary workspace list and navigation
- **Modals**: Focused overlays for specific tasks (inspect, exec, update)
- **Wizard**: Multi-step workflows (create workspace)
- **Input modes**: Filter, context menus, help

Each view has dedicated `update*` and `view*` methods, keeping concerns separated.

### Components

Components fall into three categories:

**Containers**: Large, stateful views that own their sub-components (dashboard, wizard)
**Modals**: Self-contained overlays with limited lifetime
**Primitives**: Reusable UI elements (UnifiedInput, styled components)

### Styling

Styles are centralized and semantic:

- **Colors**: Named by meaning (success, error, muted) not appearance
- **Frames**: Consistent modal borders, padding, and margins
- **Help text**: Standardized formats for discoverability

This ensures visual consistency and simplifies theme changes.

## Design Principles

### Explicit State

All state lives in model fields. No globals, no hidden variables. This makes behavior predictable and testing straightforward.

### Composition

Complex views compose simpler ones. The wizard embeds step models; the dashboard embeds the wizard. Avoid deep inheritance hierarchies.

### Message Passing

Components communicate through messages, not direct method calls. This decouples components and enables testability.

### Consistent Patterns

Similar behaviors use similar patterns. Modals share help text formats. Lists share navigation keys. Input modes switch with Tab. This reduces cognitive load for users and developers.

## User Experience Guidelines

### Discoverability

- Always show available actions
- Standardize help text format: `[Key] Action`
- Use consistent keyboard shortcuts across views

### Feedback

- Visual indication of focus and selection
- Clear confirmation of actions
- Error messages explain the problem and next steps

### Mode Awareness

When modes exist (typing vs. selecting), indicate the current mode and how to switch. Arrow keys should auto-switch when appropriate.

## Development Guidelines

### Adding a New View

1. Add a `viewState` constant
2. Add `updateView()` and `viewView()` methods
3. Wire in the main `Update()` and `View()` switches
4. Follow existing modal patterns for consistency
5. Add keyboard handling consistent with other modals

### Adding a New Component

1. Decide: container, modal, or primitive?
2. Follow the Elm model pattern (Init, Update, View)
3. Use centralized colors from `components/styled.go`
4. Apply common list/input styles if applicable
5. Add tests using the mock store pattern

### Keyboard Handling

- `Enter`: Confirm/Select
- `Esc`: Cancel/Dismiss
- `q`: Alternative quit (for user preference)
- `↑/↓/j/k`: Navigation
- `Tab`: Mode switch / Next field

Deviations should have clear rationale.

### Styling

- Use semantic color names from `components/styled.go`
- Apply `modalFrame()` for consistent overlays
- Follow help text conventions
- Keep styling minimal; let content shine

## Testing Strategy

- **Unit tests**: Model behavior in isolation with mock dependencies
- **Integration tests**: Real filesystem operations
- **E2E tests**: User interaction flows with `teatest`

Each level catches different classes of bugs. Prior unit tests for complex logic, E2E for critical paths.

## Aspirations

The TUI aims to:

- Feel responsive and polished
- Support efficient keyboard-only workflows
- Handle errors gracefully with clear recovery paths
- Scale to many workspaces without performance degradation
- Remain approachable for new users whilepowerful for frequent users

Implementation should move toward these goals incrementally.

## References

| Topic | Location |
|-------|----------|
| Main model | `internal/tui/dashboard.go` |
| Styling constants | `internal/tui/components/styled.go` |
| Reusable components | `internal/tui/components/` |
| Testing utilities | `internal/tui/testing.go` |