package views

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/tui/measure"
)

// View is the interface all TUI views must implement.
// It extends tea.Model with view lifecycle and snapshot capabilities.
type View interface {
	// Init returns the first command to run after the view is created.
	// Returns nil if no initialization is needed.
	Init() tea.Cmd

	// Update processes incoming messages and returns a result with optional command.
	// The result may indicate a new view should be pushed or an action should be taken.
	Update(msg tea.Msg) (ViewResult, tea.Cmd)

	// View renders the view's current state as a string.
	View() string

	// SetSize updates the view's dimensions based on the terminal window size.
	// Called when the terminal is resized.
	SetSize(size measure.Window)

	// OnPush is called when the view is pushed onto the stack.
	// Use this to refresh data when the view becomes visible.
	OnPush()

	// OnResume is called when returning to this view from another.
	// Use this to reload data that may have changed.
	OnResume()

	// IsLoading indicates whether the view is performing background work.
	// When true, the StackModel renders a loading spinner.
	IsLoading() bool

	// Cancel halts any ongoing background operations.
	Cancel()

	// Snapshot captures the view's state for testing or serialization.
	Snapshot() interface{}
}

// ViewResult contains the outcome of a view update.
// Either NextView or Action may be set, but not both.
type ViewResult struct {
	// NextView specifies a view to push onto the stack.
	NextView View

	// Action specifies a stack operation to perform.
	Action StackAction
}

// StackAction is the interface for stack manipulation operations.
// Implementations are private types that satisfy this interface.
type StackAction interface {
	isStackAction()
}

type StackPop struct{}

func (StackPop) isStackAction() {}

type StackPopUntilType struct {
	Type reflect.Type
}

func (StackPopUntilType) isStackAction() {}

type StackPopCount struct {
	Count int
}

func (StackPopCount) isStackAction() {}

type StackDismissAll struct{}

func (StackDismissAll) isStackAction() {}
