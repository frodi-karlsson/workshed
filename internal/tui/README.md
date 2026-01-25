# TUI Components

This directory contains the Text User Interface components for the Workshed application.

## Architecture

The TUI is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework and follows a component-based architecture.

### Components

#### Styled Components (`components/styled.go`)

Provides reusable UI components with consistent styling:

- `StyleSet`: Centralized style definitions for colors, text, buttons, etc.
- `Modal`: Standard modal dialog wrapper
- `InputField`: Styled text input with focus indicators
- `ListItem`: Styled list item with selection highlighting
- `Button`: Styled button with active states
- `ProgressBar`: Visual progress indicator
- `HelpText`: Standardized help text display

#### Focus Management (`components/focus.go`)

Manages focus state across different UI elements:

- `FocusManager`: Tracks and manages focus between different UI elements
- `FocusIndicator`: Visual indicator for focused elements
- `Focusable`: Interface for focusable components
- `FocusStyle`: Different focus visualization styles (underline, border, highlight)

#### Error Handling (`components/error.go`)

Comprehensive error display and management:

- `UIError`: Structured error with severity, category, and recovery options
- `ErrorDisplay`: Renders errors with consistent styling
- `ErrorManager`: Manages error queue and display
- `RecoveryOption`: User-selectable recovery actions

#### Input Components (`components/input.go`)

Unified input handling for various input modes:

- `UnifiedInput`: Combines text input with suggestion list
- `MultiFieldInput`: Handles multiple field inputs with validation
- `InputMode`: Typing vs Selecting mode support
- `InputValidator`: Custom validation with error messages

### Keyboard Handling (`keyboard/handlers.go`)

Reusable keyboard handling utilities:

- `KeyHandler`: Interface for key event handlers
- `NavigationHandler`: Handles list navigation (arrows, j/k, pgup/pgdn)
- `QuitHandler`: Consistent quit key handling (Ctrl+C, Esc, q)
- `ConfirmHandler`: Enter key handling for confirmations
- `CancelHandler`: Cancel/Esc key handling
- `FocusHandler`: Tab-based focus switching
- `SelectionHandler`: Multi-select support with Space key
- `CompositeHandler`: Combines multiple handlers

### State Management (`state/manager.go`)

Application state tracking and modal management:

- `StateManager`: Tracks application state with history
- `StateTransition`: Defines valid state transitions
- `ModalStack`: Manages modal dialog stack
- `ModalInfo`: Individual modal metadata and callbacks

## Usage

### Using Styled Components

```go
import "github.com/frodi/workshed/internal/tui/components"

// Create styled components
modal := components.NewModal()
modal.SetBorderColor(components.ColorSuccess)
modal.SetSize(60, 20)

button := components.NewButton("OK")
button.SetActive(true)

input := components.NewInputField()
input.SetPlaceholder("Enter value...")
```

### Using Focus Management

```go
focusMgr := components.NewFocusManager()
focusMgr.SetFocusOrder([]components.FocusState{
    components.FocusInput,
    components.FocusList,
    components.FocusButton,
})
focusMgr.Next() // Cycle to next focus
focusMgr.IsFocused(components.FocusInput)
```

### Using Error Display

```go
errDisplay := components.NewErrorDisplay()
uiErr := components.NewUIError("Failed to load workspace").
    WithSeverity(components.SeverityError).
    WithCategory(components.CategoryIO).
    WithRecovery(
        components.RecoveryOption{Label: "Retry", Action: "retry"},
    )
```

### Using Keyboard Handlers

```go
import "github.com/frodi/workshed/internal/tui/keyboard"

navHandler := keyboard.NewNavigationHandler(&list)
quitHandler := keyboard.NewQuitHandler()

composite := keyboard.NewCompositeHandler()
composite.Add(navHandler)
composite.Add(quitHandler)

handled, cmd := composite.HandleKey(msg)
```

## Key Bindings

The TUI follows these standard key bindings:

| Key | Action |
|-----|--------|
| Enter | Confirm / Select |
| Tab | Switch focus |
| ↑↓/j/k | Navigate lists |
| Space | Toggle selection |
| Home/End | Jump to start/end |
| PgUp/PgDown | Page navigation |
| Esc/q/Ctrl+C | Cancel / Quit |

## Color Palette

| Constant | Color | Usage |
|----------|-------|-------|
| ColorBorder | #874B07 | Default border color |
| ColorSuccess | #4CD964 | Success states |
| ColorError | #FF6B6B | Error states |
| ColorText | #D4D4D4 | Default text |
| ColorMuted | #888888 | Muted text |
| ColorVeryMuted | #666666 | Very muted text |
| ColorBackground | #3C3C3C | Background |
| ColorHighlight | #5C5CFF | Focus/selection |
| ColorWarning | #FFB347 | Warning states |

## Best Practices

1. **Use shared components**: Always use the styled components from `components/` for consistent UI
2. **Follow key binding standards**: Use the keyboard handlers for consistent input handling
3. **Manage focus properly**: Use FocusManager for complex views with multiple input areas
4. **Handle errors gracefully**: Use ErrorManager for consistent error display
5. **Test interactions**: Write tests for keyboard navigation and focus management