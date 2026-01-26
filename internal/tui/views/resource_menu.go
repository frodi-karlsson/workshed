package views

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ResourceMenuView struct {
	store         workspace.Store
	ctx           context.Context
	handle        string
	list          list.Model
	size          measure.Window
	invocationCtx workspace.InvocationContext
	subMenus      map[string]func() View
}

type ResourceItem struct {
	key         string
	title       string
	description string
	section     string
}

func (i ResourceItem) Title() string { return i.title }

func (i ResourceItem) Description() string {
	return i.description
}

func (i ResourceItem) FilterValue() string {
	return i.title
}

func NewResourceMenuView(s workspace.Store, ctx context.Context, handle string, invocationCtx workspace.InvocationContext) ResourceMenuView {
	items := []list.Item{
		ResourceItem{key: "i", title: "[i] Info", description: "View, copy path, update, or check health", section: "Info"},
		ResourceItem{key: "e", title: "[e] Exec", description: "Run command or view history", section: "Actions"},
		ResourceItem{key: "r", title: "[r] Repositories", description: "Manage repositories", section: "Repositories"},
		ResourceItem{key: "c", title: "[c] Captures", description: "State captures and apply", section: "Captures"},
		ResourceItem{key: "x", title: "[x] Remove", description: "Delete this workspace", section: "System"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Actions"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return ResourceMenuView{
		store:         s,
		ctx:           ctx,
		handle:        handle,
		list:          l,
		invocationCtx: invocationCtx,
		subMenus: map[string]func() View{
			"i": func() View { infoMenu := NewInfoMenuView(s, ctx, handle); return &infoMenu },
		},
	}
}

func (v *ResourceMenuView) OnPush() {}

func (v *ResourceMenuView) OnResume() {
	_, err := v.store.Get(v.ctx, v.handle)
	if err != nil {
		v.list.SetItems([]list.Item{})
	}
}

func (v *ResourceMenuView) IsLoading() bool { return false }

func (v *ResourceMenuView) Cancel() {}

func (v *ResourceMenuView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *ResourceMenuView) Init() tea.Cmd { return nil }

func (v *ResourceMenuView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent},
		{Key: "esc", Help: "[Esc] Back", Action: v.goBack},
		{Key: "ctrl+c", Help: "[Ctrl+C] Back", Action: v.goBack},
	}
}

func (v *ResourceMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ri, ok := selected.(ResourceItem); ok {
		return v.handleResourceAction(ri), nil
	}
	return ViewResult{}, nil
}

func (v *ResourceMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ResourceMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ri, ok := item.(ResourceItem); ok && ri.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					if subMenu, exists := v.subMenus[key]; exists {
						return ViewResult{NextView: subMenu()}, nil
					}
					result := v.handleResourceAction(ri)
					if result.NextView != nil {
						return result, nil
					}
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ri, ok := selected.(ResourceItem); ok {
				if subMenu, exists := v.subMenus[ri.key]; exists {
					return ViewResult{NextView: subMenu()}, nil
				}
				return v.handleResourceAction(ri), nil
			}
		}
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, cmd
}

func (v *ResourceMenuView) handleResourceAction(item ResourceItem) ViewResult {
	switch item.key {
	case "e":
		execMenu := NewExecMenuView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: &execMenu}
	case "r":
		reposMenu := NewReposMenuView(v.store, v.ctx, v.handle, v.invocationCtx)
		return ViewResult{NextView: &reposMenu}
	case "c":
		capturesMenu := NewCapturesMenuView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: &capturesMenu}
	case "x":
		removeView := NewRemoveConfirmView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: &removeView}
	}
	return ViewResult{}
}

func (v *ResourceMenuView) View() string {
	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(v.list.View())
}

type ResourceMenuViewSnapshot struct {
	Type        string
	Handle      string
	ItemCount   int
	SelectedIdx int
}

func (v *ResourceMenuView) Snapshot() interface{} {
	return ResourceMenuViewSnapshot{
		Type:        "ResourceMenuView",
		Handle:      v.handle,
		ItemCount:   len(v.list.Items()),
		SelectedIdx: v.list.Index(),
	}
}
