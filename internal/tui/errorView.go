package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func ErrorView(err error) string {
	const maxWidth = 60
	errorMsg := err.Error()

	wrappedMsg := wrapText(errorMsg, maxWidth)

	return modalFrame().
		BorderForeground(colorError).
		Width(maxWidth).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Foreground(colorError).Render("Error"),
				"\n",
				lipgloss.NewStyle().Foreground(colorText).Render(wrappedMsg),
				"\n",
				lipgloss.NewStyle().Foreground(colorVeryMuted).MarginTop(1).Render("[Enter] Dismiss  [q] Quit"),
			),
		)
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text)

	for i, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLine.WriteString(word)
		}

		if i == len(words)-1 {
			result.WriteString(currentLine.String())
		}
	}

	return result.String()
}
