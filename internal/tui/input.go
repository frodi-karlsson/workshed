package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/workspace"
)

type purposeItem struct {
	purpose string
}

func (p purposeItem) Title() string       { return p.purpose }
func (p purposeItem) Description() string { return "" }
func (p purposeItem) FilterValue() string { return p.purpose }

type inputMode int

const (
	modeTyping inputMode = iota
	modeSelecting
)

type inputModel struct {
	textInput   textinput.Model
	list        list.Model
	workspaces  []*workspace.Workspace
	hasResult   bool
	resultValue string
	mode        inputMode
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			if m.mode == modeTyping {
				m.mode = modeSelecting
			} else {
				m.mode = modeTyping
			}
			return m, nil
		case tea.KeyEnter:
			if m.mode == modeTyping {
				value := strings.TrimSpace(m.textInput.Value())
				if value == "" {
					return m, nil
				}
				m.hasResult = true
				m.resultValue = value
				return m, tea.Quit
			} else {
				if len(m.list.Items()) > 0 {
					selected := m.list.SelectedItem()
					if selected != nil {
						if item, ok := selected.(purposeItem); ok {
							m.hasResult = true
							m.resultValue = item.purpose
							return m, tea.Quit
						}
					}
				}
			}
		case tea.KeyUp, tea.KeyDown:
			m.mode = modeSelecting
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, min(msg.Height-v-5, maxListHeight))
	}

	if m.mode == modeTyping {
		updatedInput, cmd := m.textInput.Update(msg)
		m.textInput = updatedInput
		cmds = append(cmds, cmd)

		query := strings.ToLower(m.textInput.Value())
		filteredItems := filterPurposeItems(m.workspaces, query)
		m.list.SetItems(filteredItems)
	}

	if m.mode == modeSelecting {
		updatedList, cmd := m.list.Update(msg)
		m.list = updatedList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func filterPurposeItems(workspaces []*workspace.Workspace, query string) []list.Item {
	purposeSet := make(map[string]bool)
	var items []list.Item

	lowerQuery := strings.ToLower(query)
	for _, ws := range workspaces {
		if !purposeSet[ws.Purpose] {
			purposeSet[ws.Purpose] = true
			if query == "" || strings.Contains(strings.ToLower(ws.Purpose), lowerQuery) {
				items = append(items, purposeItem{purpose: ws.Purpose})
			}
		}
	}

	return items
}

func (m inputModel) View() string {
	if len(m.workspaces) == 0 {
		return noPurposesView
	}

	modeIndicator := "[TYPING]"
	helpText := "[Enter] Next  [Tab] Browse  [Esc] Cancel"
	if m.mode == modeSelecting {
		modeIndicator = "[SELECTING]"
		helpText = "[Enter] Select  [Tab] Type  [↑↓] Navigate  [Esc] Cancel"
	}

	inputView := m.textInput.View()
	listView := m.list.View()

	modeStyle := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		inputStyle.Render("Purpose: ")+inputView+"  "+modeStyle.Render(modeIndicator)+"\n",
		suggestionHeader.Render("Suggestions (Tab to browse):")+"\n",
		docStyle.Render(listView),
		"\n"+helpStyle.Render(helpText),
	)
}

var (
	noPurposesView = fmt.Sprintf(`%s

No workspaces exist yet. Type a purpose and press Enter to create your first workspace.

`, lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSuccess).
		Render("No existing purposes"))

	inputStyle = lipgloss.NewStyle().
			Foreground(colorText)

	suggestionHeader = lipgloss.NewStyle().
				Foreground(colorMuted).
				MarginTop(1)
)

type PurposeInput struct {
	store store.Store
}

func NewPurposeInput(store store.Store) *PurposeInput {
	return &PurposeInput{store: store}
}

func (p *PurposeInput) Run(ctx context.Context) (string, error) {
	workspaces, err := p.store.List(ctx, workspace.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("listing workspaces: %w", err)
	}

	ti := textinput.New()
	ti.Placeholder = "e.g., OpenCode development, API migration..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	items := filterPurposeItems(workspaces, "")
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(items, delegate, 50, maxListHeight)
	l.SetShowTitle(false)
	applyCommonListStyles(&l)

	m := inputModel{
		textInput:  ti,
		list:       l,
		workspaces: workspaces,
		mode:       modeTyping,
	}

	ptea := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := ptea.Run()
	if err != nil {
		return "", fmt.Errorf("running TUI: %w", err)
	}

	if finalM, ok := finalModel.(inputModel); ok && finalM.hasResult {
		return finalM.resultValue, nil
	}

	return "", fmt.Errorf("input cancelled")
}

func TryInputPurpose(ctx context.Context, s store.Store, l *logger.Logger) (string, bool) {
	if !IsHumanMode() {
		return "", false
	}

	input := NewPurposeInput(s)
	purpose, err := input.Run(ctx)
	if err != nil {
		l.Help("purpose input cancelled")
		return "", false
	}

	return purpose, true
}

type purposeStepModel struct {
	textInput   textinput.Model
	list        list.Model
	workspaces  []*workspace.Workspace
	done        bool
	cancelled   bool
	resultValue string
	mode        inputMode
}

func newPurposeStepModel(workspaces []*workspace.Workspace) purposeStepModel {
	ti := textinput.New()
	ti.Placeholder = "e.g., OpenCode development, API migration..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	items := filterPurposeItems(workspaces, "")
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(items, delegate, 50, maxListHeight)
	l.SetShowTitle(false)
	applyCommonListStyles(&l)

	return purposeStepModel{
		textInput:  ti,
		list:       l,
		workspaces: workspaces,
		mode:       modeTyping,
	}
}

func (m purposeStepModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m purposeStepModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyTab:
			if m.mode == modeTyping {
				m.mode = modeSelecting
			} else {
				m.mode = modeTyping
			}
			return m, nil
		case tea.KeyEnter:
			if m.mode == modeTyping {
				value := strings.TrimSpace(m.textInput.Value())
				if value == "" {
					return m, nil
				}
				m.done = true
				m.resultValue = value
				return m, nil
			} else {
				if len(m.list.Items()) > 0 {
					selected := m.list.SelectedItem()
					if selected != nil {
						if item, ok := selected.(purposeItem); ok {
							m.done = true
							m.resultValue = item.purpose
							return m, nil
						}
					}
				}
			}
		case tea.KeyUp, tea.KeyDown:
			m.mode = modeSelecting
		case tea.KeyRunes:
			switch msg.String() {
			case "q":
				m.cancelled = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, min(msg.Height-v-5, maxListHeight))
	}

	if m.mode == modeTyping {
		updatedInput, cmd := m.textInput.Update(msg)
		m.textInput = updatedInput
		cmds = append(cmds, cmd)

		query := strings.ToLower(m.textInput.Value())
		filteredItems := filterPurposeItems(m.workspaces, query)
		m.list.SetItems(filteredItems)
	}

	if m.mode == modeSelecting {
		updatedList, cmd := m.list.Update(msg)
		m.list = updatedList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m purposeStepModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText)

	borderStyle := modalFrame()

	if len(m.workspaces) == 0 {
		return borderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render("Create Workspace"),
				"\n",
				inputStyle.Render("Purpose: ")+m.textInput.View(),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorMuted).
					Render("No existing workspaces. Type a purpose and press Enter."),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorMuted).
					MarginTop(1).
					Render("[Enter] Next  [Esc] Cancel"),
			),
		)
	}

	modeIndicator := "[TYPING]"
	helpText := "[Enter] Next  [Tab] Browse  [Esc] Cancel"
	if m.mode == modeSelecting {
		modeIndicator = "[SELECTING]"
		helpText = "[Enter] Select  [Tab] Type  [↑↓] Navigate  [Esc] Cancel"
	}

	modeStyle := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1)

	return borderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("Create Workspace"),
			"\n",
			inputStyle.Render("Purpose: ")+m.textInput.View()+"  "+modeStyle.Render(modeIndicator),
			"\n",
			suggestionHeader.Render("Suggestions (Tab to browse):"),
			"\n",
			docStyle.Render(m.list.View()),
			"\n",
			helpStyle.Render(helpText),
		),
	)
}
