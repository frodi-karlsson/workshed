package views

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type RemoveConfirmView struct {
	store  workspace.Store
	ctx    context.Context
	handle string
	list   list.Model
	size   measure.Window
}

type ConfirmItem struct {
	key   string
	title string
}

func (i ConfirmItem) Title() string { return i.title }

func (i ConfirmItem) Description() string { return "" }

func (i ConfirmItem) FilterValue() string { return i.title }

func NewRemoveConfirmView(s workspace.Store, ctx context.Context, handle string) RemoveConfirmView {
	items := []list.Item{
		ConfirmItem{key: "y", title: "Yes, delete workspace"},
		ConfirmItem{key: "n", title: "No, cancel"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = fmt.Sprintf("Delete %s?", handle)
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return RemoveConfirmView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		list:   l,
	}
}

func (v *RemoveConfirmView) OnPush() {}

func (v *RemoveConfirmView) OnResume() {}

func (v *RemoveConfirmView) IsLoading() bool { return false }

func (v *RemoveConfirmView) Cancel() {}

func (v *RemoveConfirmView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *RemoveConfirmView) Init() tea.Cmd { return nil }

func (v *RemoveConfirmView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent},
		{Key: "esc", Help: "[Esc] Back", Action: v.goBack},
		{Key: "ctrl+c", Help: "[Ctrl+C] Back", Action: v.goBack},
	}
}

func (v *RemoveConfirmView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ci, ok := selected.(ConfirmItem); ok {
		return v.handleConfirmAction(ci), nil
	}
	return ViewResult{}, nil
}

func (v *RemoveConfirmView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *RemoveConfirmView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			if key == "y" {
				if err := v.store.Remove(v.ctx, v.handle); err != nil {
					return ViewResult{Action: StackPop{}}, nil
				}
				return ViewResult{Action: StackPopCount{Count: 2}}, nil
			}
			if key == "n" {
				return ViewResult{Action: StackPop{}}, nil
			}
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ci, ok := item.(ConfirmItem); ok && ci.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					return v.handleConfirmAction(ci), nil
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ci, ok := selected.(ConfirmItem); ok {
				return v.handleConfirmAction(ci), nil
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

func (v *RemoveConfirmView) handleConfirmAction(item ConfirmItem) ViewResult {
	switch item.key {
	case "y":
		if err := v.store.Remove(v.ctx, v.handle); err != nil {
			return ViewResult{Action: StackPop{}}
		}
		return ViewResult{Action: StackPopCount{Count: 2}}
	case "n":
		return ViewResult{Action: StackPop{}}
	}
	return ViewResult{}
}

func (v *RemoveConfirmView) View() string {
	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(v.list.View())
}

type RemoveConfirmViewSnapshot struct {
	Type   string
	Handle string
}

func (v *RemoveConfirmView) Snapshot() interface{} {
	return RemoveConfirmViewSnapshot{
		Type:   "RemoveConfirmView",
		Handle: v.handle,
	}
}
