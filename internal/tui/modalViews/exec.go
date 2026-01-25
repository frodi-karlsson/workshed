package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type ExecModal struct {
	workspace *workspace.Workspace
	onDismiss func()
	dismissed bool
}

func NewExecModal(ws *workspace.Workspace, onDismiss func()) ExecModal {
	return ExecModal{
		workspace: ws,
		onDismiss: onDismiss,
		dismissed: false,
	}
}

func (m ExecModal) Update(msg tea.Msg) (ExecModal, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		case tea.KeyRunes:
			switch msg.String() {
			case "q":
				m.dismissed = true
				if m.onDismiss != nil {
					m.onDismiss()
				}
				return m, true
			}
		case tea.KeyEnter:
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		case tea.KeyTab:
			return m, false
		case tea.KeySpace:
			return m, false
		}
	}
	return m, m.dismissed
}

func (m ExecModal) View() string {
	return modalFrame().
		BorderForeground(colorSuccess).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Render("Command to run in \""+m.workspace.Handle+"\""),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					Render("Exec modal placeholder"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					MarginTop(1).
					Render("[Tab] Switch  [Space] Toggle  [Enter] Run  [Esc] Cancel"),
			),
		)
}

func (m ExecModal) Dismissed() bool {
	return m.dismissed
}
