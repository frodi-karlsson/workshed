package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	key      string
	label    string
	desc     string
	selected bool
}

func (m menuItem) Title() string {
	if m.selected {
		return lipgloss.NewStyle().
			Foreground(colorSuccess).
			Render("[" + m.key + "] " + m.label)
	}
	return "[" + m.key + "] " + m.label
}

func (m menuItem) Description() string { return m.desc }
func (m menuItem) FilterValue() string { return m.key + " " + m.label }

const contextMenuWidth = 40

type contextMenuModel struct {
	list     list.Model
	selected int
	done     bool
	quit     bool
	result   string
}

func newContextMenuModel(handle string) contextMenuModel {
	items := []list.Item{
		menuItem{key: "i", label: "Inspect", desc: "Show workspace details", selected: false},
		menuItem{key: "p", label: "Path", desc: "Copy path to clipboard", selected: false},
		menuItem{key: "e", label: "Exec", desc: "Run command in repositories", selected: false},
		menuItem{key: "u", label: "Update", desc: "Change workspace purpose", selected: false},
		menuItem{key: "r", label: "Remove", desc: "Delete workspace (confirm)", selected: false},
	}

	l := list.New(items, list.NewDefaultDelegate(), contextMenuWidth, maxListHeight)
	l.Title = "Actions for \"" + handle + "\""
	l.SetShowTitle(true)
	applyCommonListStyles(&l)
	applyTitleStyle(&l)
	l.Styles.Title = l.Styles.Title.Width(contextMenuWidth)

	return contextMenuModel{
		list:     l,
		selected: 0,
		done:     false,
		quit:     false,
		result:   "",
	}
}

func (m contextMenuModel) Init() tea.Cmd { return nil }

func (m contextMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			return m, tea.Quit
		case tea.KeyEnter:
			items := m.list.Items()
			if m.selected >= 0 && m.selected < len(items) {
				if mi, ok := items[m.selected].(menuItem); ok {
					m.result = mi.key
				}
			}
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.selected = m.list.Index()
	return m, cmd
}

func (m contextMenuModel) View() string {
	frameStyle := modalFrame()
	if m.done {
		frameStyle = frameStyle.BorderForeground(colorSuccess)
	}
	return frameStyle.Render(
		m.list.View() + "\n\n" +
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				Render("[↑↓/j/k] Navigate  [Enter] Select  [Esc/q/Ctrl+C] Cancel"),
	)
}

func ShowContextMenu(handle string) (string, error) {
	m := newContextMenuModel(handle)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	if fm, ok := finalModel.(contextMenuModel); ok {
		if fm.quit {
			return "", nil
		}
		return fm.result, nil
	}
	return "", nil
}
