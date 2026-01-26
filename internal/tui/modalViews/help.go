package modalViews

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

type HelpModal struct {
	onDismiss func()
	dismissed bool
}

func NewHelpModal(onDismiss func()) HelpModal {
	return HelpModal{
		onDismiss: onDismiss,
		dismissed: false,
	}
}

func (m HelpModal) Update(msg tea.Msg) (HelpModal, bool) {
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

func (m HelpModal) View() string {
	helpText := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Render("Keyboard Shortcuts") + "\n\n" +
		"[c] Create workspace\n" +
		"[Enter] Open action menu for selected workspace\n" +
		"[l] Filter workspaces by purpose or handle\n" +
		"[↑/↓/j/k] Navigate workspaces\n" +
		"[?] Toggle this help\n" +
		"[q/Esc] Quit"

	helpStyle := lipgloss.NewStyle().
		Foreground(components.ColorVeryMuted)

	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			"\n",
			helpStyle.Render("[Esc/q/Enter] Dismiss"),
		),
	)
}

func (m HelpModal) Dismissed() bool {
	return m.dismissed
}
