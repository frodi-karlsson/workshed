package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

type AlertModal struct {
	content  string
	quitting bool
}

func NewAlertModal(content string) *AlertModal {
	return &AlertModal{content: content}
}

func (m *AlertModal) Init() tea.Cmd {
	return nil
}

func (m *AlertModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AlertModal) View() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(components.ColorVeryMuted)

	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.content,
			"\n",
			helpStyle.Render("[Esc/Enter] Dismiss"),
		),
	)
}

func ShowAlertModal(content string) error {
	m := NewAlertModal(content)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running alert modal: %w", err)
	}
	return nil
}
