package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

const maxListHeight = 10

var (
	colorBorder     = components.ColorBorder
	colorSuccess    = components.ColorSuccess
	colorError      = components.ColorError
	colorText       = components.ColorText
	colorMuted      = components.ColorMuted
	colorVeryMuted  = components.ColorVeryMuted
	colorBackground = components.ColorBackground
	colorWarning    = components.ColorWarning
)

func applyCommonListStyles(l *list.Model) {
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(colorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(colorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(colorMuted)
}

func applyTitleStyle(l *list.Model) {
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Background(colorBackground).
		Padding(0, 1)
}

func modalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
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
