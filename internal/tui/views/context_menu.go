package views

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/tui/components"
)

type MenuItem struct {
	key      string
	label    string
	desc     string
	selected bool
}

func (m MenuItem) Title() string       { return "[" + m.key + "] " + m.label }
func (m MenuItem) Description() string { return m.desc }
func (m MenuItem) FilterValue() string { return m.key + " " + m.label }

const ContextMenuWidth = 40

type ContextMenuView struct {
	store  store.Store
	ctx    context.Context
	handle string
	list   list.Model
	stale  bool
}

func NewContextMenuView(s store.Store, ctx context.Context, handle string) ContextMenuView {
	items := []list.Item{
		MenuItem{key: "i", label: "Inspect", desc: "Show workspace details", selected: false},
		MenuItem{key: "p", label: "Path", desc: "Copy path to clipboard", selected: false},
		MenuItem{key: "e", label: "Exec", desc: "Run command in repositories", selected: false},
		MenuItem{key: "u", label: "Update", desc: "Change workspace purpose", selected: false},
		MenuItem{key: "r", label: "Remove", desc: "Delete workspace (confirm)", selected: false},
	}

	l := list.New(items, list.NewDefaultDelegate(), ContextMenuWidth, 20)
	l.Title = "Actions for \"" + handle + "\""
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.Title = l.Styles.Title.Width(ContextMenuWidth)

	return ContextMenuView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		list:   l,
	}
}

func (v *ContextMenuView) Init() tea.Cmd {
	return nil
}

func (v *ContextMenuView) OnPush() {}
func (v *ContextMenuView) OnResume() {
	_, err := v.store.Get(v.ctx, v.handle)
	if err != nil {
		v.stale = true
	}
}

func (v *ContextMenuView) IsLoading() bool {
	return false
}

func (v *ContextMenuView) Cancel() {}

func (v *ContextMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.stale {
		return ViewResult{Action: StackPop{}}, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.list.SetSize(ContextMenuWidth, 20)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			selected := v.list.SelectedItem()
			if selected != nil {
				if mi, ok := selected.(MenuItem); ok {
					return v.handleMenuAction(mi.key), nil
				}
			}
			return ViewResult{}, nil
		}
		key := msg.String()
		if len(key) == 1 {
			for _, item := range v.list.Items() {
				if mi, ok := item.(MenuItem); ok && mi.key == key {
					return v.handleMenuAction(mi.key), nil
				}
			}
		}
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return ViewResult{}, cmd
}

func (v ContextMenuView) handleMenuAction(key string) ViewResult {
	switch key {
	case "i":
		inspectView := NewInspectView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: inspectView}
	case "p":
		pathView := NewPathView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: pathView}
	case "e":
		execView := NewExecView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: execView}
	case "u":
		updateView := NewUpdateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: updateView}
	case "r":
		removeView := NewRemoveView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: removeView}
	}
	return ViewResult{}
}

func (v ContextMenuView) View() string {
	frameStyle := ModalFrame()

	return frameStyle.Render(
		v.list.View() + "\n" +
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				MarginTop(1).
				Render("[↑↓/j/k] Navigate  [Enter] Select  [Esc/q/Ctrl+C] Cancel"),
	)
}

type ContextMenuViewSnapshot struct {
	Type        string
	Handle      string
	Items       []MenuItemInfo
	SelectedKey string
	Stale       bool
}

type MenuItemInfo struct {
	Key   string
	Label string
	Desc  string
}

func (v *ContextMenuView) Snapshot() interface{} {
	items := make([]MenuItemInfo, len(v.list.Items()))
	for i, item := range v.list.Items() {
		mi := item.(MenuItem)
		items[i] = MenuItemInfo{Key: mi.key, Label: mi.label, Desc: mi.desc}
	}
	var selectedKey string
	if selected := v.list.SelectedItem(); selected != nil {
		if mi, ok := selected.(MenuItem); ok {
			selectedKey = mi.key
		}
	}
	return ContextMenuViewSnapshot{
		Type:        "ContextMenuView",
		Handle:      v.handle,
		Items:       items,
		SelectedKey: selectedKey,
		Stale:       v.stale,
	}
}
