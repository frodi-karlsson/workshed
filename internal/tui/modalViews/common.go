package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorBorder    = lipgloss.Color("#874B07")
	colorSuccess   = lipgloss.Color("#4CD964")
	colorError     = lipgloss.Color("#FF6B6B")
	colorText      = lipgloss.Color("#D4D4D4")
	colorMuted     = lipgloss.Color("#888888")
	colorVeryMuted = lipgloss.Color("#666666")
)

// ModalView defines the interface for modal dialogs.
// Modal views are self-contained dialogs that overlay the main view.
type ModalView interface {
	// Update processes messages and returns an updated modal view.
	// The boolean indicates whether the modal state changed.
	Update(msg tea.Msg) (ModalView, bool)

	// View renders the modal's current state.
	View() string

	// Dismissed returns true when the user has closed the modal.
	Dismissed() bool
}

// DismissableModal is a ModalView with explicit dismissal semantics.
// Implementations track user confirmation or cancellation state.
type DismissableModal interface {
	ModalView

	// Dismissed returns true when the user has closed the modal.
	Dismissed() bool
}

func modalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1)
}
