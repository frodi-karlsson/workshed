package views

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
)

const MaxListHeight = 10

func ModalFrame(size measure.Window) lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, size.ModalMargin()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.ColorBorder).
		Padding(1).
		Width(size.ModalWidth()).
		Height(size.ModalHeight())
}

func ErrorView(err error, size measure.Window) string {
	return lipgloss.NewStyle().
		Foreground(components.ColorError).
		Padding(1, 2).
		Width(size.ModalWidth()).
		Height(size.ModalHeight()).
		Render("Error: " + err.Error())
}
