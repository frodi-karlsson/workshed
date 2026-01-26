package components

import "github.com/charmbracelet/lipgloss"

type FocusIndicator struct {
	Style        lipgloss.Style
	FocusedStyle lipgloss.Style
	Char         string
	FocusedChar  string
}

func NewFocusIndicator() FocusIndicator {
	return FocusIndicator{
		Style: lipgloss.NewStyle().
			Foreground(ColorMuted),
		FocusedStyle: lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true),
		Char:        " ",
		FocusedChar: "▶",
	}
}

func (fi FocusIndicator) Render(focused bool) string {
	if focused {
		return fi.FocusedStyle.Render(fi.FocusedChar)
	}
	return fi.Style.Render(fi.Char)
}

func (fi *FocusIndicator) SetFocused(focused bool) *FocusIndicator {
	if focused {
		fi.Style = fi.Style.Foreground(ColorHighlight)
	} else {
		fi.Style = fi.Style.Foreground(ColorMuted)
	}
	return fi
}

type FocusStyle int

const (
	FocusStyleNone FocusStyle = iota

	FocusStyleUnderline

	FocusStyleBorder

	FocusStyleHighlight

	FocusStyleReverse
)

func GetFocusStyle(style FocusStyle) lipgloss.Style {
	switch style {
	case FocusStyleUnderline:
		return lipgloss.NewStyle().
			Underline(true).
			Foreground(ColorHighlight)
	case FocusStyleBorder:
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorHighlight).
			Padding(0, 1)
	case FocusStyleHighlight:
		return lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorHighlight)
	case FocusStyleReverse:
		return lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorText)
	default:
		return lipgloss.NewStyle()
	}
}
