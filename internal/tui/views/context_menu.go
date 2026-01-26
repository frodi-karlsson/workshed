package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ContextMenuView struct {
	store         workspace.Store
	ctx           context.Context
	handle        string
	menu          components.MenuModel
	stale         bool
	size          measure.Window
	invocationCtx workspace.InvocationContext
}

func NewContextMenuView(s workspace.Store, ctx context.Context, handle string, invocationCtx workspace.InvocationContext) ContextMenuView {
	sections := []components.MenuSection{
		{
			Name: "Info",
			Items: []components.MenuItem{
				{Key: "i", Label: "Inspect", Desc: "View workspace details", Section: "Info"},
				{Key: "p", Label: "Path", Desc: "Copy path to clipboard", Section: "Info"},
				{Key: "u", Label: "Update", Desc: "Edit purpose", Section: "Info"},
			},
		},
		{
			Name: "Actions",
			Items: []components.MenuItem{
				{Key: "e", Label: "Exec", Desc: "Run command", Section: "Actions"},
				{Key: "h", Label: "History", Desc: "Execution history", Section: "Actions"},
			},
		},
		{
			Name: "Repositories",
			Items: []components.MenuItem{
				{Key: "a", Label: "Add", Desc: "Add repository", Section: "Repositories"},
				{Key: "r", Label: "Remove", Desc: "Remove repository", Section: "Repositories"},
			},
		},
		{
			Name: "State",
			Items: []components.MenuItem{
				{Key: "c", Label: "Capture", Desc: "Create or view captures", Section: "State"},
			},
		},
		{
			Name: "System",
			Items: []components.MenuItem{
				{Key: "D", Label: "Derive", Desc: "Generate context JSON", Section: "System"},
				{Key: "v", Label: "Validate", Desc: "Check AGENTS.md", Section: "System"},
				{Key: "k", Label: "Health", Desc: "Check workspace health", Section: "System"},
			},
		},
	}

	menu := components.NewMenuModel()
	menu.SetSections(sections)

	return ContextMenuView{
		store:         s,
		ctx:           ctx,
		handle:        handle,
		menu:          menu,
		stale:         false,
		invocationCtx: invocationCtx,
	}
}

func (v *ContextMenuView) Init() tea.Cmd {
	return nil
}

func (v *ContextMenuView) SetSize(size measure.Window) {
	v.size = size
	v.menu.SetSize(size.ListWidth(), size.ListHeight())
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

func (v *ContextMenuView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent},
		{Key: "esc", Help: "[Esc] Back", Action: v.goBack},
		{Key: "ctrl+c", Help: "[Ctrl+C] Back", Action: v.goBack},
	}
}

func (v *ContextMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.menu.SelectedItem()
	if selected != nil {
		return v.handleMenuAction(selected.Key), nil
	}
	return ViewResult{}, nil
}

func (v *ContextMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ContextMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.stale {
		return ViewResult{Action: StackPop{}}, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			key := string(msg.Runes[0])
			selected := v.menu.SelectByKey(key)
			if selected != nil {
				return v.handleMenuAction(key), nil
			}
		}
		if msg.Type == tea.KeyEnter {
			selected := v.menu.SelectedItem()
			if selected != nil {
				return v.handleMenuAction(selected.Key), nil
			}
			return ViewResult{}, nil
		}
		v.menu.Update(msg)

		if result, _, handled := HandleKey(v.KeyBindings(), msg); handled {
			return result, nil
		}
	}

	return ViewResult{}, nil
}

func (v ContextMenuView) handleMenuAction(key string) ViewResult {
	switch key {
	case "i":
		inspectView := NewInspectView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: inspectView}
	case "p":
		pathView := NewPathView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: pathView}
	case "u":
		updateView := NewUpdateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: updateView}
	case "e":
		execView := NewExecView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: execView}
	case "h":
		historyView := NewExecHistoryView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: historyView}
	case "a":
		addRepoView := NewAddRepoView(v.store, v.ctx, v.handle, v.invocationCtx)
		return ViewResult{NextView: &addRepoView}
	case "r":
		removeRepoView := NewRemoveRepoView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: removeRepoView}
	case "c":
		capturesMenu := NewCapturesMenuView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: &capturesMenu}
	case "D":
		deriveView := NewDeriveView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: deriveView}
	case "v":
		validateView := NewValidateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: validateView}
	case "k":
		healthView := NewHealthView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: healthView}
	}
	return ViewResult{}
}

func (v ContextMenuView) View() string {
	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(v.menu.View())
}

type ContextMenuViewSnapshot struct {
	Type        string         `json:",omitempty"`
	Handle      string         `json:",omitempty"`
	Items       []MenuItemInfo `json:",omitempty"`
	SelectedKey string         `json:",omitempty"`
	Stale       bool           `json:",omitempty"`
	CurrentPage int            `json:",omitempty"`
	PageCount   int            `json:",omitempty"`
}

type MenuItemInfo struct {
	Key   string
	Label string
	Desc  string
}

func (v *ContextMenuView) Snapshot() interface{} {
	sections := v.menu.Sections
	items := []MenuItemInfo{}
	for _, section := range sections {
		for _, item := range section.Items {
			items = append(items, MenuItemInfo{
				Key:   item.Key,
				Label: item.Label,
				Desc:  item.Desc,
			})
		}
	}

	selectedItem := v.menu.SelectedItem()
	selectedKey := ""
	if selectedItem != nil {
		selectedKey = selectedItem.Key
	}

	return ContextMenuViewSnapshot{
		Type:        "ContextMenuView",
		Handle:      v.handle,
		Items:       items,
		SelectedKey: selectedKey,
		Stale:       v.stale,
		CurrentPage: v.menu.CurrentPage(),
		PageCount:   v.menu.PageCount(),
	}

}
