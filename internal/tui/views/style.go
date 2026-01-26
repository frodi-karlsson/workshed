package views

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

const MaxListHeight = 10

func ModalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.ColorBorder).
		Padding(1).
		Width(64)
}

func ErrorView(err error) string {
	return lipgloss.NewStyle().
		Foreground(components.ColorError).
		Padding(1, 2).
		Render("Error: " + err.Error())
}
