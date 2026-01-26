package views

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
)

type PathInputView struct {
	textInput   textinput.Model
	value       string
	placeholder string
	prompt      string
	onSelect    func(path string)
	onCancel    func()
	focused     bool
}

func NewPathInputView(placeholder string, prompt string, onSelect func(path string), onCancel func()) PathInputView {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = prompt
	ti.Focus()
	ti.CharLimit = 300

	return PathInputView{
		textInput:   ti,
		placeholder: placeholder,
		prompt:      prompt,
		onSelect:    onSelect,
		onCancel:    onCancel,
		focused:     true,
	}
}

func (v *PathInputView) Init() tea.Cmd {
	return textinput.Blink
}

func (v *PathInputView) OnPush()   {}
func (v *PathInputView) OnResume() {}
func (v *PathInputView) IsLoading() bool {
	return false
}
func (v *PathInputView) Cancel() {}

func (v *PathInputView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			if v.onCancel != nil {
				v.onCancel()
			}
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			value := strings.TrimSpace(v.textInput.Value())
			if value != "" {
				absPath, _ := filepath.Abs(value)
				if v.onSelect != nil {
					v.onSelect(absPath)
				}
				return ViewResult{Action: StackPop{}}, nil
			}
		}
	}

	updatedInput, cmd := v.textInput.Update(msg)
	v.textInput = updatedInput
	v.value = v.textInput.Value()

	return ViewResult{}, cmd
}

func (v *PathInputView) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText)

	borderStyle := ModalFrame()

	return borderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Select Path"), "\n", "\n",
			v.textInput.View(), "\n", "\n",
			lipgloss.NewStyle().
				Foreground(components.ColorMuted).
				Render("Type a path"),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				Render("[Enter] Select  [Tab] Complete  [Esc] Cancel"),
		),
	)
}

type PathInputViewSnapshot struct {
	Type  string
	Value string
}

func (v *PathInputView) Snapshot() interface{} {
	return PathInputViewSnapshot{
		Type:  "PathInputView",
		Value: v.textInput.Value(),
	}
}
