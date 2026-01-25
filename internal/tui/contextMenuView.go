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

func (m menuItem) Title() string       { return "[" + m.key + "] " + m.label }
func (m menuItem) Description() string { return m.desc }
func (m menuItem) FilterValue() string { return m.key + " " + m.label }

const contextMenuWidth = 40

type contextMenuView struct {
	list           list.Model
	handle         string
	selectedAction string
	cancelled      bool
}

func NewContextMenuView(handle string) contextMenuView {
	items := []list.Item{
		menuItem{key: "i", label: "Inspect", desc: "Show workspace details", selected: false},
		menuItem{key: "p", label: "Path", desc: "Copy path to clipboard", selected: false},
		menuItem{key: "e", label: "Exec", desc: "Run command in repositories", selected: false},
		menuItem{key: "u", label: "Update", desc: "Change workspace purpose", selected: false},
		menuItem{key: "r", label: "Remove", desc: "Delete workspace (confirm)", selected: false},
	}

	l := list.New(items, list.NewDefaultDelegate(), contextMenuWidth, 20)
	l.Title = "Actions for \"" + handle + "\""
	l.SetShowTitle(true)
	applyCommonListStyles(&l)
	applyTitleStyle(&l)
	l.Styles.Title = l.Styles.Title.Width(contextMenuWidth)

	return contextMenuView{
		list:   l,
		handle: handle,
	}
}

func (v contextMenuView) Init() tea.Cmd {
	return nil
}

func (v contextMenuView) Update(msg tea.Msg) (contextMenuView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.list.SetSize(contextMenuWidth, 20)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			v.cancelled = true
			return v, nil
		case tea.KeyEnter:
			selected := v.list.SelectedItem()
			if selected != nil {
				if mi, ok := selected.(menuItem); ok {
					v.selectedAction = mi.key
				}
			}
			return v, nil
		}
		key := msg.String()
		if len(key) == 1 {
			for _, item := range v.list.Items() {
				if mi, ok := item.(menuItem); ok && mi.key == key {
					v.selectedAction = mi.key
					return v, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v contextMenuView) View() string {
	frameStyle := modalFrame()

	return frameStyle.Render(
		v.list.View() + "\n" +
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				MarginTop(1).
				Render("[↑↓/j/k] Navigate  [Enter] Select  [Esc/q/Ctrl+C] Cancel"),
	)
}
