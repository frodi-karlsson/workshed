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

type CapturesMenuView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	list     list.Model
	captures []workspace.Capture
	size     measure.Window
}

type CaptureMenuItem struct {
	key         string
	title       string
	description string
}

func (i CaptureMenuItem) Title() string { return i.title }

func (i CaptureMenuItem) Description() string {
	return i.description
}

func (i CaptureMenuItem) FilterValue() string {
	return i.title
}

func NewCapturesMenuView(s workspace.Store, ctx context.Context, handle string) CapturesMenuView {
	items := []list.Item{
		CaptureMenuItem{key: "n", title: "[n] New Capture", description: "Create a new state capture"},
		CaptureMenuItem{key: "l", title: "[l] List Captures", description: "View existing captures"},
		CaptureMenuItem{key: "a", title: "[a] Apply Capture", description: "Restore a captured state"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Captures"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return CapturesMenuView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		list:     l,
		captures: []workspace.Capture{},
	}
}

func (v *CapturesMenuView) OnPush() {
	captures, _ := v.store.ListCaptures(v.ctx, v.handle)
	v.captures = captures
}

func (v *CapturesMenuView) OnResume() {}

func (v *CapturesMenuView) IsLoading() bool { return false }

func (v *CapturesMenuView) Cancel() {}

func (v *CapturesMenuView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *CapturesMenuView) Init() tea.Cmd { return nil }

func (v *CapturesMenuView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent}},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *CapturesMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ci, ok := selected.(CaptureMenuItem); ok {
		return v.handleCaptureAction(ci), nil
	}
	return ViewResult{}, nil
}

func (v *CapturesMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *CapturesMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ci, ok := item.(CaptureMenuItem); ok && ci.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					return v.handleCaptureAction(ci), nil
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ci, ok := selected.(CaptureMenuItem); ok {
				return v.handleCaptureAction(ci), nil
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

func (v *CapturesMenuView) handleCaptureAction(item CaptureMenuItem) ViewResult {
	switch item.key {
	case "n":
		createView := NewCaptureCreateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: createView}
	case "l":
		listView := NewCaptureListView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: listView}
	case "a":
		applyView := NewApplyCaptureView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: &applyView}
	}
	return ViewResult{}
}

func (v *CapturesMenuView) View() string {
	helpText := GenerateHelp(v.KeyBindings())
	helpHint := lipgloss.NewStyle().
		Foreground(components.ColorMuted).
		MarginTop(1).
		Render(helpText)

	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			v.list.View(),
			helpHint,
		),
	)
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
