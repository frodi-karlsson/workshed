package views

import "github.com/charmbracelet/lipgloss"

var (
	ColorBorder     = lipgloss.Color("#874B07")
	ColorSuccess    = lipgloss.Color("#4CD964")
	ColorError      = lipgloss.Color("#FF6B6B")
	ColorText       = lipgloss.Color("#D4D4D4")
	ColorMuted      = lipgloss.Color("#888888")
	ColorVeryMuted  = lipgloss.Color("#666666")
	ColorBackground = lipgloss.Color("#3C3C3C")
	ColorWarning    = lipgloss.Color("#FFB347")
)

const MaxListHeight = 10

func ModalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1).
		Width(64)
}

func ErrorView(err error) string {
	return lipgloss.NewStyle().
		Foreground(ColorError).
		Padding(1, 2).
		Render("Error: " + err.Error())
}
