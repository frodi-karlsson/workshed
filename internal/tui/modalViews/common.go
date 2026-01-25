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

type ModalView interface {
	Update(msg tea.Msg) (ModalView, bool)
	View() string
	Dismissed() bool
}

type DismissableModal interface {
	ModalView
	Dismissed() bool
}

func modalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1)
}
