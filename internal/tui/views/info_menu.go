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

type InfoMenuView struct {
	store  workspace.Store
	ctx    context.Context
	handle string
	list   list.Model
	size   measure.Window
}

type InfoMenuItem struct {
	key         string
	title       string
	description string
}

func (i InfoMenuItem) Title() string { return i.title }

func (i InfoMenuItem) Description() string {
	return i.description
}

func (i InfoMenuItem) FilterValue() string {
	return i.title
}

func NewInfoMenuView(s workspace.Store, ctx context.Context, handle string) InfoMenuView {
	items := []list.Item{
		InfoMenuItem{key: "i", title: "[i] Inspect", description: "View workspace details"},
		InfoMenuItem{key: "p", title: "[p] Path", description: "Copy path to clipboard"},
		InfoMenuItem{key: "u", title: "[u] Update", description: "Edit purpose"},
		InfoMenuItem{key: "h", title: "[h] Health", description: "Check workspace health"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Info"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return InfoMenuView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		list:   l,
	}
}

func (v *InfoMenuView) OnPush() {}

func (v *InfoMenuView) OnResume() {}

func (v *InfoMenuView) IsLoading() bool { return false }

func (v *InfoMenuView) Cancel() {}

func (v *InfoMenuView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *InfoMenuView) Init() tea.Cmd { return nil }

func (v *InfoMenuView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent}},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *InfoMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ii, ok := selected.(InfoMenuItem); ok {
		return v.handleInfoAction(ii), nil
	}
	return ViewResult{}, nil
}

func (v *InfoMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *InfoMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ii, ok := item.(InfoMenuItem); ok && ii.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					return v.handleInfoAction(ii), nil
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ii, ok := selected.(InfoMenuItem); ok {
				return v.handleInfoAction(ii), nil
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

func (v *InfoMenuView) handleInfoAction(item InfoMenuItem) ViewResult {
	switch item.key {
	case "i":
		inspectView := NewInspectView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: inspectView}
	case "p":
		pathView := NewPathView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: pathView}
	case "u":
		updateView := NewUpdateView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: updateView}
	case "h":
		healthView := NewHealthView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: healthView}
	}
	return ViewResult{}
}

func (v *InfoMenuView) View() string {
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

type InfoMenuViewSnapshot struct {
	Type        string
	Handle      string
	ItemCount   int
	SelectedIdx int
}

func (v *InfoMenuView) Snapshot() interface{} {
	return InfoMenuViewSnapshot{
		Type:        "InfoMenuView",
		Handle:      v.handle,
		ItemCount:   len(v.list.Items()),
		SelectedIdx: v.list.Index(),
	}
}
