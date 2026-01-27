package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBorder     = lipgloss.Color("#874B07")
	ColorSuccess    = lipgloss.Color("#4CD964")
	ColorError      = lipgloss.Color("#FF6B6B")
	ColorText       = lipgloss.Color("#D4D4D4")
	ColorMuted      = lipgloss.Color("#888888")
	ColorVeryMuted  = lipgloss.Color("#666666")
	ColorBackground = lipgloss.Color("#3C3C3C")
	ColorHighlight  = lipgloss.Color("#5C5CFF")
	ColorWarning    = lipgloss.Color("#FFB347")
	ColorGold       = lipgloss.Color("#FFD700")
)

type HelpItem struct {
	Key   string
	Label string
}

func RenderHelp(items []HelpItem) string {
	if len(items) == 0 {
		return ""
	}
	var parts []string
	for _, item := range items {
		if item.Label != "" {
			parts = append(parts, "["+item.Key+"] "+item.Label)
		} else {
			parts = append(parts, "["+item.Key+"]")
		}
	}
	return lipgloss.NewStyle().Foreground(ColorVeryMuted).Render(strings.Join(parts, "  "))
}

var (
	HelpDismiss   = RenderHelp([]HelpItem{{"Esc/q/Enter", "Dismiss"}})
	HelpNavigate  = RenderHelp([]HelpItem{{"↑↓/j/k", "Navigate"}, {"Enter", "Select"}, {"Esc", "Cancel"}})
	HelpInput     = RenderHelp([]HelpItem{{"Enter", "Confirm"}, {"Tab", "Browse"}, {"Esc", "Cancel"}})
	HelpDashboard = RenderHelp([]HelpItem{{"c", "Create"}, {"Enter", "Menu"}, {"l", "Filter"}, {"?", "Help"}, {"q", "Quit"}})
)
