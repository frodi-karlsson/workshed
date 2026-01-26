package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/workspace"
)

type ExecResultModal struct {
	command   string
	results   []workspace.ExecResult
	onDismiss func()
	dismissed bool
}

func NewExecResultModal(command string, results []workspace.ExecResult, onDismiss func()) ExecResultModal {
	return ExecResultModal{
		command:   command,
		results:   results,
		onDismiss: onDismiss,
		dismissed: false,
	}
}

func (m ExecResultModal) Update(msg tea.Msg) (ExecResultModal, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "enter":
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		}
	}
	return m, m.dismissed
}

func (m ExecResultModal) View() string {
	if len(m.results) == 0 {
		return modalFrame().Render("No results")
	}

	borderColor := components.ColorSuccess
	allSuccess := true
	for _, result := range m.results {
		if result.ExitCode != 0 {
			allSuccess = false
		}
	}
	if !allSuccess {
		borderColor = components.ColorError
	}

	statusText := "Success"
	if !allSuccess {
		statusText = "Failed"
	}

	status := lipgloss.NewStyle().
		Foreground(borderColor).
		Render("[" + statusText + "]")

	return modalFrame().BorderForeground(borderColor).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(components.ColorText).
					Render("Command Execution Results"),
				"",
				lipgloss.JoinHorizontal(lipgloss.Left, status, "  ", m.command),
				"\n",
				lipgloss.NewStyle().
					Foreground(components.ColorVeryMuted).
					MarginTop(1).
					Render("[Enter/Esc/q] Close"),
			),
		)
}

func (m ExecResultModal) Dismissed() bool {
	return m.dismissed
}
