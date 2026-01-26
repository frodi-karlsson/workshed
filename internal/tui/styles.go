package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

const maxListHeight = 10

func applyCommonListStyles(l *list.Model) {
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
}

func applyTitleStyle(l *list.Model) {
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Background(components.ColorBackground).
		Padding(0, 1)
}

func modalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.ColorBorder).
		Padding(1)
}

const (
	KeyConfirm  = "enter"
	KeyCancel   = "esc"
	KeyQuit     = "q"
	KeyNavigate = "up/down/j/k"
	KeyTab      = "tab"
	KeySpace    = "space"
	KeyHome     = "home"
	KeyEnd      = "end"
	KeyPageUp   = "pgup"
	KeyPageDown = "pgdown"
)
