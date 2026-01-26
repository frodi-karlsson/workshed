package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

type RemoveModal struct {
	handle       string
	confirmState bool
	onDismiss    func()
	onConfirm    func()
	dismissed    bool
}

func NewRemoveModal(handle string, onDismiss func(), onConfirm func()) RemoveModal {
	return RemoveModal{
		handle:       handle,
		confirmState: false,
		onDismiss:    onDismiss,
		onConfirm:    onConfirm,
		dismissed:    false,
	}
}

func (m RemoveModal) Update(msg tea.Msg) (RemoveModal, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "n", "N":
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		case "y", "Y":
			m.confirmState = true
			if m.onConfirm != nil {
				m.onConfirm()
			}
			m.dismissed = true
			return m, true
		}
	}
	return m, m.dismissed
}

func (m RemoveModal) View() string {
	frameStyle := modalFrame().BorderForeground(components.ColorError)
	if m.confirmState {
		frameStyle = frameStyle.BorderForeground(components.ColorSuccess)
	}

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(components.ColorError).
				Render("Remove Workspace?"),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorText).
				Render(m.handle),
			"\n\n",
			lipgloss.NewStyle().
				Foreground(components.ColorMuted).
				Render("This will delete the workspace directory."),
			"\n\n",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(components.ColorText).
				Render("[y] Yes  [n] No"),
		),
	)
}

func (m RemoveModal) Dismissed() bool {
	return m.dismissed
}
