package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputMode int

const (
	ModeTyping InputMode = iota
	ModeSelecting
)

type ValidationResult struct {
	Valid bool
	Error string
}

type InputValidator func(value string) ValidationResult

type UnifiedInput struct {
	textInput   textinput.Model
	list        list.Model
	mode        InputMode
	validator   InputValidator
	placeholder string
	hasResult   bool
	resultValue string
}

func NewUnifiedInput() UnifiedInput {
	ti := textinput.New()
	ti.CharLimit = 100
	ti.Prompt = ""

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 6)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return UnifiedInput{
		textInput: ti,
		list:      l,
		mode:      ModeTyping,
	}
}

func (i *UnifiedInput) SetPlaceholder(placeholder string) *UnifiedInput {
	i.placeholder = placeholder
	i.textInput.Placeholder = placeholder
	return i
}

func (i *UnifiedInput) SetValidator(validator InputValidator) *UnifiedInput {
	i.validator = validator
	return i
}

func (i *UnifiedInput) Focus() {
	i.textInput.Focus()
	i.mode = ModeTyping
}

func (i *UnifiedInput) HasResult() bool {
	return i.hasResult
}

func (i *UnifiedInput) GetResult() string {
	return i.resultValue
}

func (i *UnifiedInput) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if i.mode == ModeTyping {
				i.mode = ModeSelecting
				i.textInput.Blur()
			} else {
				i.mode = ModeTyping
				i.textInput.Focus()
			}
			return true, nil
		case tea.KeyEnter:
			if i.mode == ModeTyping {
				value := strings.TrimSpace(i.textInput.Value())
				if value == "" && i.placeholder != "" {
					value = i.placeholder
				}
				if value != "" {
					if i.validator != nil {
						result := i.validator(value)
						if !result.Valid {
							return true, nil
						}
					}
					i.hasResult = true
					i.resultValue = value
					return true, tea.Quit
				}
			} else if selected := i.list.SelectedItem(); selected != nil {
				i.hasResult = true
				i.resultValue = selected.FilterValue()
				return true, tea.Quit
			}
		case tea.KeyEsc:
			i.hasResult = false
			i.resultValue = ""
			return true, tea.Quit
		case tea.KeyUp, tea.KeyDown:
			if i.mode == ModeTyping {
				i.mode = ModeSelecting
				i.textInput.Blur()
			}
		}
	}

	var cmds []tea.Cmd
	if i.mode == ModeTyping {
		updated, cmd := i.textInput.Update(msg)
		i.textInput = updated
		cmds = append(cmds, cmd)
	}
	if i.mode == ModeSelecting {
		updated, cmd := i.list.Update(msg)
		i.list = updated
		cmds = append(cmds, cmd)
	}
	return false, tea.Batch(cmds...)
}

func (i *UnifiedInput) View() string {
	modeIndicator := "[TYPING]"
	helpText := "[Enter] Confirm  [Tab] Browse  [Esc] Cancel"
	if i.mode == ModeSelecting {
		modeIndicator = "[SELECTING]"
		helpText = "[Enter] Select  [Tab] Type  [↑↓] Navigate  [Esc] Cancel"
	}

	modeStyle := lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(ColorMuted).MarginTop(1)

	inputView := i.textInput.View()
	listView := i.list.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		inputView+"  "+modeStyle.Render(modeIndicator),
		"\n",
		"Suggestions:",
		listView,
		"\n"+helpStyle.Render(helpText),
	)
}

type MultiFieldInput struct {
	fields     []textinput.Model
	focusedIdx int
	mode       InputMode
	validator  func(values []string) ValidationResult
	results    []string
	done       bool
	cancelled  bool
}

func NewMultiFieldInput(fieldCount int) MultiFieldInput {
	fields := make([]textinput.Model, fieldCount)
	for i := 0; i < fieldCount; i++ {
		fields[i] = textinput.New()
		fields[i].CharLimit = 100
		fields[i].Prompt = ""
	}
	return MultiFieldInput{
		fields:     fields,
		focusedIdx: 0,
		mode:       ModeTyping,
	}
}

func (m *MultiFieldInput) SetPlaceholder(idx int, placeholder string) {
	if idx >= 0 && idx < len(m.fields) {
		m.fields[idx].Placeholder = placeholder
	}
}

func (m *MultiFieldInput) SetValidator(validator func(values []string) ValidationResult) {
	m.validator = validator
}

func (m *MultiFieldInput) GetValues() []string {
	values := make([]string, len(m.fields))
	for i, f := range m.fields {
		values[i] = f.Value()
	}
	return values
}

func (m *MultiFieldInput) IsDone() bool         { return m.done }
func (m *MultiFieldInput) IsCancelled() bool    { return m.cancelled }
func (m *MultiFieldInput) GetResults() []string { return m.results }

func (m *MultiFieldInput) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab:
			dir := 1
			if msg.Type == tea.KeyShiftTab {
				dir = -1
			}
			m.focusedIdx = (m.focusedIdx + dir + len(m.fields)) % len(m.fields)
			for i := range m.fields {
				if i == m.focusedIdx {
					m.fields[i].Focus()
				} else {
					m.fields[i].Blur()
				}
			}
			return true, nil
		case tea.KeyEnter:
			values := m.GetValues()
			if m.validator != nil {
				result := m.validator(values)
				if !result.Valid {
					return true, nil
				}
			}
			m.results = values
			m.done = true
			return true, tea.Quit
		case tea.KeyEsc:
			m.cancelled = true
			return true, tea.Quit
		}
	}

	for i := range m.fields {
		updated, cmd := m.fields[i].Update(msg)
		m.fields[i] = updated
		if cmd != nil {
			return true, cmd
		}
	}
	return false, nil
}

func (m MultiFieldInput) View() string {
	var content []string
	for i, f := range m.fields {
		prefix := "  "
		if i == m.focusedIdx {
			prefix = "▶ "
		}
		content = append(content, prefix+"Field "+string(rune('1'+i))+": "+f.View())
	}
	helpStyle := lipgloss.NewStyle().Foreground(ColorMuted).MarginTop(1)
	content = append(content, helpStyle.Render("[Tab] Next  [Enter] Submit  [Esc] Cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}
