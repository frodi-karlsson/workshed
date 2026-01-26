package modalViews

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/tui/components"
)

type UpdateModal struct {
	textInput textinput.Model
	handle    string
	onDismiss func()
	onConfirm func(purpose string)
	dismissed bool
}

func NewUpdateModal(handle string, onDismiss func(), onConfirm func(purpose string)) UpdateModal {
	ti := textinput.New()
	ti.Placeholder = "Enter new purpose..."
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.Focus()

	return UpdateModal{
		textInput: ti,
		handle:    handle,
		onDismiss: onDismiss,
		onConfirm: onConfirm,
		dismissed: false,
	}
}

func (m UpdateModal) Update(msg tea.Msg) (UpdateModal, bool) {
	if key.IsCancel(msg) {
		m.dismissed = true
		if m.onDismiss != nil {
			m.onDismiss()
		}
		return m, true
	}

	if key.IsEnter(msg) {
		purpose := m.textInput.Value()
		if purpose != "" {
			if m.onConfirm != nil {
				m.onConfirm(purpose)
			}
		}
		m.dismissed = true
		return m, true
	}

	updatedInput, cmd := m.textInput.Update(msg)
	m.textInput = updatedInput
	return m, cmd == nil && m.dismissed
}

func (m UpdateModal) View() string {
	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(components.ColorText).
				Render("Update Purpose for \""+m.handle+"\""),
			"\n",
			m.textInput.View(),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				MarginTop(1).
				Render("[Enter] Save  [Esc] Cancel"),
		),
	)
}

func (m UpdateModal) Dismissed() bool {
	return m.dismissed
}
