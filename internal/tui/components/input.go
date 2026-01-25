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
	textInput       textinput.Model
	list            list.Model
	workspaces      interface{}
	mode            InputMode
	validator       InputValidator
	placeholder     string
	hasResult       bool
	resultValue     string
	focusedField    int
	fieldCount      int
	showSuggestions bool
	filterFunc      func(workspaces interface{}, query string) []list.Item
}

func NewUnifiedInput() UnifiedInput {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.Focus()

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 6)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return UnifiedInput{
		textInput:       ti,
		list:            l,
		mode:            ModeTyping,
		focusedField:    0,
		fieldCount:      1,
		showSuggestions: true,
	}
}

func (i *UnifiedInput) SetPlaceholder(placeholder string) *UnifiedInput {
	i.placeholder = placeholder
	i.textInput.Placeholder = placeholder
	return i
}

func (i *UnifiedInput) SetCharLimit(limit int) *UnifiedInput {
	i.textInput.CharLimit = limit
	return i
}

func (i *UnifiedInput) SetValidator(validator InputValidator) *UnifiedInput {
	i.validator = validator
	return i
}

func (i *UnifiedInput) SetFilterFunc(filterFunc func(workspaces interface{}, query string) []list.Item) *UnifiedInput {
	i.filterFunc = filterFunc
	return i
}

func (i *UnifiedInput) SetWorkspaces(workspaces interface{}) *UnifiedInput {
	i.workspaces = workspaces
	i.updateSuggestions()
	return i
}

func (i *UnifiedInput) SetFieldCount(count int) *UnifiedInput {
	i.fieldCount = count
	return i
}

func (i *UnifiedInput) Focus() {
	i.textInput.Focus()
}

func (i *UnifiedInput) Blur() {
	i.textInput.Blur()
}

func (i *UnifiedInput) SetValue(value string) {
	i.textInput.SetValue(value)
}

func (i *UnifiedInput) GetValue() string {
	return i.textInput.Value()
}

func (i *UnifiedInput) Mode() InputMode {
	return i.mode
}

func (i *UnifiedInput) HasResult() bool {
	return i.hasResult
}

func (i *UnifiedInput) GetResult() string {
	return i.resultValue
}

func (i *UnifiedInput) updateSuggestions() {
	if i.filterFunc == nil || i.workspaces == nil {
		return
	}

	query := strings.ToLower(i.textInput.Value())
	items := i.filterFunc(i.workspaces, query)
	i.list.SetItems(items)
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

				if i.validator != nil {
					result := i.validator(value)
					if !result.Valid {
						return true, nil
					}
				}

				if value == "" && i.placeholder == "" {
					return true, nil
				}

				i.hasResult = true
				if value == "" && i.placeholder != "" {
					i.resultValue = i.placeholder
				} else {
					i.resultValue = value
				}
				return true, tea.Quit
			} else {
				if len(i.list.Items()) > 0 {
					selected := i.list.SelectedItem()
					if selected != nil {
						item := selected
						i.hasResult = true
						i.resultValue = item.FilterValue()
						return true, tea.Quit
					}
				}
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
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		i.list.SetSize(msg.Width-h, min(msg.Height-v-5, 6))
	}

	var cmds []tea.Cmd

	if i.mode == ModeTyping {
		updatedInput, cmd := i.textInput.Update(msg)
		i.textInput = updatedInput
		cmds = append(cmds, cmd)
		i.updateSuggestions()
	}

	if i.mode == ModeSelecting {
		updatedList, cmd := i.list.Update(msg)
		i.list = updatedList
		cmds = append(cmds, cmd)
	}

	return false, tea.Batch(cmds...)
}

func (i UnifiedInput) View() string {
	modeIndicator := "[TYPING]"
	helpText := "[Enter] Confirm  [Tab] Browse  [Esc] Cancel"

	if i.mode == ModeSelecting {
		modeIndicator = "[SELECTING]"
		helpText = "[Enter] Select  [Tab] Type  [↑↓] Navigate  [Esc] Cancel"
	}

	modeStyle := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)

	suggestionStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)

	inputView := i.textInput.View()

	var content []string
	content = append(content, inputStyle.Render("Value: ")+inputView+"  "+modeStyle.Render(modeIndicator))

	if i.showSuggestions && len(i.list.Items()) > 0 {
		content = append(content, suggestionStyle.Render("Suggestions:"))
		content = append(content, i.list.View())
	}

	content = append(content, helpStyle.Render(helpText))

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

type MultiFieldInput struct {
	fields     []textinput.Model
	focusedIdx int
	fieldCount int
	mode       InputMode
	validator  func(values []string) ValidationResult
	results    []string
	done       bool
	cancelled  bool
}

func NewMultiFieldInput(fieldCount int) MultiFieldInput {
	fields := make([]textinput.Model, fieldCount)
	for i := 0; i < fieldCount; i++ {
		ti := textinput.New()
		ti.CharLimit = 100
		ti.Prompt = ""
		fields[i] = ti
	}

	return MultiFieldInput{
		fields:     fields,
		focusedIdx: 0,
		fieldCount: fieldCount,
		mode:       ModeTyping,
	}
}

func (m *MultiFieldInput) SetPlaceholder(idx int, placeholder string) {
	if idx >= 0 && idx < len(m.fields) {
		m.fields[idx].Placeholder = placeholder
	}
}

func (m *MultiFieldInput) GetValues() []string {
	values := make([]string, len(m.fields))
	for i, f := range m.fields {
		values[i] = f.Value()
	}
	return values
}

func (m *MultiFieldInput) SetValidator(validator func(values []string) ValidationResult) {
	m.validator = validator
}

func (m *MultiFieldInput) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab:
			dir := 1
			if msg.Type == tea.KeyShiftTab {
				dir = -1
			}
			m.focusedIdx += dir
			if m.focusedIdx >= m.fieldCount {
				m.focusedIdx = 0
			}
			if m.focusedIdx < 0 {
				m.focusedIdx = m.fieldCount - 1
			}
			for i, f := range m.fields {
				if i == m.focusedIdx {
					f.Focus()
				} else {
					f.Blur()
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
		content = append(content, lipgloss.NewStyle().Foreground(ColorText).Render(prefix+"Field "+string(rune('1'+i))+": "+f.View()))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)
	content = append(content, helpStyle.Render("[Tab] Next field  [Enter] Submit  [Esc] Cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m *MultiFieldInput) IsDone() bool {
	return m.done
}

func (m *MultiFieldInput) IsCancelled() bool {
	return m.cancelled
}

func (m *MultiFieldInput) GetResults() []string {
	return m.results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
