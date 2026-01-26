package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type CapturesMenuView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	menu     components.MenuModel
	captures []workspace.Capture
	loading  bool
	size     measure.Window
}

func NewCapturesMenuView(s workspace.Store, ctx context.Context, handle string) CapturesMenuView {
	sections := []components.MenuSection{
		{
			Name: "Captures",
			Items: []components.MenuItem{
				{Key: "n", Label: "New", Desc: "Create new capture", Section: "Captures"},
			},
		},
	}

	menu := components.NewMenuModel()
	menu.SetSections(sections)

	return CapturesMenuView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		menu:     menu,
		captures: []workspace.Capture{},
	}
}

func (v *CapturesMenuView) Init() tea.Cmd { return nil }

func (v *CapturesMenuView) SetSize(size measure.Window) {
	v.size = size
	v.menu.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *CapturesMenuView) OnPush() {
	captures, _ := v.store.ListCaptures(v.ctx, v.handle)
	v.captures = captures
	v.updateMenuForCaptures()
}

func (v *CapturesMenuView) OnResume() {}
func (v *CapturesMenuView) IsLoading() bool {
	return v.loading
}

func (v *CapturesMenuView) Cancel() {
	v.loading = false
}

func (v *CapturesMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if v.loading {
		return ViewResult{}, nil
	}

	if len(v.captures) == 0 {
		captures, err := v.store.ListCaptures(v.ctx, v.handle)
		if err == nil {
			v.captures = captures
			v.updateMenuForCaptures()
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			selected := v.menu.SelectedItem()
			if selected != nil {
				return v.handleMenuAction(selected.Key), nil
			}
			return ViewResult{}, nil
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				key := string(msg.Runes[0])
				selected := v.menu.SelectByKey(key)
				if selected != nil {
					return v.handleMenuAction(key), nil
				}
			}
		}
		v.menu.Update(msg)
	}

	return ViewResult{}, nil
}

func (v *CapturesMenuView) updateMenuForCaptures() {
	hasCaptures := len(v.captures) > 0

	sections := []components.MenuSection{
		{
			Name: "Captures",
			Items: []components.MenuItem{
				{Key: "n", Label: "New", Desc: "Create new capture", Section: "Captures"},
			},
		},
	}

	if hasCaptures {
		sections[0].Items = append(sections[0].Items, components.MenuItem{
			Key:   "l",
			Label: "List",
			Desc:  "View existing captures",
		})
	}

	menu := components.NewMenuModel()
	menu.SetSections(sections)
	v.menu = menu
}

func (v *CapturesMenuView) handleMenuAction(key string) ViewResult {
	switch key {
	case "n", "c":
		createView := NewCaptureCreateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: createView}
	case "l":
		listView := NewCaptureListView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: listView}
	}
	return ViewResult{}
}

func (v *CapturesMenuView) View() string {
	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(v.menu.View())
}

type CapturesMenuViewSnapshot struct {
	Type        string
	Handle      string
	HasCaptures bool
	CaptureCnt  int
}

func (v *CapturesMenuView) Snapshot() interface{} {
	return CapturesMenuViewSnapshot{
		Type:        "CapturesMenuView",
		Handle:      v.handle,
		HasCaptures: len(v.captures) > 0,
		CaptureCnt:  len(v.captures),
	}
}
