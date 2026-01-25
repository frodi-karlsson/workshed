package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type PathModal struct {
	workspace *workspace.Workspace
	onDismiss func()
	dismissed bool
}

func NewPathModal(ws *workspace.Workspace, onDismiss func()) PathModal {
	return PathModal{
		workspace: ws,
		onDismiss: onDismiss,
		dismissed: false,
	}
}

func (m PathModal) Update(msg tea.Msg) (PathModal, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		case tea.KeyRunes:
			if msg.String() == "q" {
				m.dismissed = true
				if m.onDismiss != nil {
					m.onDismiss()
				}
				return m, true
			}
		}
	}
	return m, m.dismissed
}

func (m PathModal) View() string {
	return modalFrame().
		BorderForeground(colorSuccess).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Render("Workspace Path"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorMuted).
					Render(m.workspace.Path),
				"\n\n",
				lipgloss.NewStyle().
					Foreground(colorSuccess).
					Render("Path copied to clipboard!"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					Render("[Esc/q/Enter] Dismiss"),
			),
		)
}

func (m PathModal) Dismissed() bool {
	return m.dismissed
}
