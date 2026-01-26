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

type ExecMenuView struct {
	store  workspace.Store
	ctx    context.Context
	handle string
	list   list.Model
	size   measure.Window
}

type ExecMenuItem struct {
	key         string
	title       string
	description string
}

func (i ExecMenuItem) Title() string { return i.title }

func (i ExecMenuItem) Description() string {
	return i.description
}

func (i ExecMenuItem) FilterValue() string {
	return i.title
}

func NewExecMenuView(s workspace.Store, ctx context.Context, handle string) ExecMenuView {
	items := []list.Item{
		ExecMenuItem{key: "r", title: "[r] Run", description: "Execute a command"},
		ExecMenuItem{key: "h", title: "[h] History", description: "View execution history"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Exec"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return ExecMenuView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		list:   l,
	}
}

func (v *ExecMenuView) OnPush() {}

func (v *ExecMenuView) OnResume() {}

func (v *ExecMenuView) IsLoading() bool { return false }

func (v *ExecMenuView) Cancel() {}

func (v *ExecMenuView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *ExecMenuView) Init() tea.Cmd { return nil }

func (v *ExecMenuView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent}},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *ExecMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ei, ok := selected.(ExecMenuItem); ok {
		return v.handleExecAction(ei), nil
	}
	return ViewResult{}, nil
}

func (v *ExecMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ExecMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ei, ok := item.(ExecMenuItem); ok && ei.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					return v.handleExecAction(ei), nil
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ei, ok := selected.(ExecMenuItem); ok {
				return v.handleExecAction(ei), nil
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

func (v *ExecMenuView) handleExecAction(item ExecMenuItem) ViewResult {
	switch item.key {
	case "r":
		execView := NewExecView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: execView}
	case "h":
		historyView := NewExecHistoryView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: historyView}
	}
	return ViewResult{}
}

func (v *ExecMenuView) View() string {
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

type ExecMenuViewSnapshot struct {
	Type        string
	Handle      string
	ItemCount   int
	SelectedIdx int
}

func (v *ExecMenuView) Snapshot() interface{} {
	return ExecMenuViewSnapshot{
		Type:        "ExecMenuView",
		Handle:      v.handle,
		ItemCount:   len(v.list.Items()),
		SelectedIdx: v.list.Index(),
	}
}
