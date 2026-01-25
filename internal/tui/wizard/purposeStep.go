package wizard

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

const maxListHeight = 6

type inputMode int

const (
	modeTyping inputMode = iota
	modeSelecting
)

type purposeItem struct {
	purpose string
}

func (p purposeItem) Title() string       { return p.purpose }
func (p purposeItem) Description() string { return "" }
func (p purposeItem) FilterValue() string { return p.purpose }

type PurposeStep struct {
	textInput   textinput.Model
	list        list.Model
	workspaces  []*workspace.Workspace
	done        bool
	cancelled   bool
	resultValue string
	mode        inputMode
}

func NewPurposeStep(workspaces []*workspace.Workspace) PurposeStep {
	ti := textinput.New()
	ti.Placeholder = "e.g., OpenCode development, API migration..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D4D4D4")).
		Bold(true)

	items := filterPurposeItems(workspaces, "")
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(items, delegate, 50, maxListHeight)
	l.SetShowTitle(false)
	applyCommonListStyles(&l)

	return PurposeStep{
		textInput:  ti,
		list:       l,
		workspaces: workspaces,
		mode:       modeTyping,
	}
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

func applyCommonListStyles(l *list.Model) {
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
}

func (m PurposeStep) Init() tea.Cmd {
	return textinput.Blink
}

func (m PurposeStep) Update(msg tea.Msg) (PurposeStep, tea.Cmd) {
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

func (m PurposeStep) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#D4D4D4"))

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
					Foreground(lipgloss.Color("#888888")).
					Render("No existing workspaces. Type a purpose and press Enter."),
				"\n",
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888")).
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
		Foreground(lipgloss.Color("#4CD964")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
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

func (m PurposeStep) IsDone() bool {
	return m.done
}

func (m PurposeStep) IsCancelled() bool {
	return m.cancelled
}

func (m PurposeStep) GetResult() interface{} {
	return m.resultValue
}

func (m PurposeStep) GetPurpose() string {
	return m.resultValue
}

var (
	docStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874B07")).
			Padding(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D4D4D4"))

	suggestionHeader = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				MarginTop(1)
)

func modalFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874B07")).
		Padding(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
