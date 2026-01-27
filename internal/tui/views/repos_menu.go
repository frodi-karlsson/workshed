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

type ReposMenuView struct {
	store         workspace.Store
	ctx           context.Context
	handle        string
	list          list.Model
	size          measure.Window
	invocationCtx workspace.InvocationContext
}

type RepoMenuItem struct {
	key         string
	title       string
	description string
}

func (i RepoMenuItem) Title() string { return i.title }

func (i RepoMenuItem) Description() string {
	return i.description
}

func (i RepoMenuItem) FilterValue() string {
	return i.title
}

func NewReposMenuView(s workspace.Store, ctx context.Context, handle string, invocationCtx workspace.InvocationContext) ReposMenuView {
	items := []list.Item{
		RepoMenuItem{key: "a", title: "[a] Add Repository", description: "Add a new repository"},
		RepoMenuItem{key: "l", title: "[l] List Repositories", description: "List and manage repositories"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Repositories"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Padding(0, 1)

	return ReposMenuView{
		store:         s,
		ctx:           ctx,
		handle:        handle,
		list:          l,
		invocationCtx: invocationCtx,
	}
}

func (v *ReposMenuView) OnPush() {}

func (v *ReposMenuView) OnResume() {}

func (v *ReposMenuView) IsLoading() bool { return false }

func (v *ReposMenuView) Cancel() {}

func (v *ReposMenuView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *ReposMenuView) Init() tea.Cmd { return nil }

func (v *ReposMenuView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Select", Action: v.selectCurrent}},
		GetDismissKeyBindings(v.goBack, "Back")...,
	)
}

func (v *ReposMenuView) selectCurrent() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if ri, ok := selected.(RepoMenuItem); ok {
		return v.handleRepoAction(ri), nil
	}
	return ViewResult{}, nil
}

func (v *ReposMenuView) goBack() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ReposMenuView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.Type == tea.KeyRunes && len(km.Runes) == 1 {
			key := string(km.Runes[0])
			for i := 0; i < len(v.list.Items()); i++ {
				item := v.list.Items()[i]
				if ri, ok := item.(RepoMenuItem); ok && ri.key == key {
					v.list.CursorDown()
					v.list.CursorUp()
					return v.handleRepoAction(ri), nil
				}
			}
		}
		if km.Type == tea.KeyEnter {
			selected := v.list.SelectedItem()
			if ri, ok := selected.(RepoMenuItem); ok {
				return v.handleRepoAction(ri), nil
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

func (v *ReposMenuView) handleRepoAction(item RepoMenuItem) ViewResult {
	switch item.key {
	case "a":
		addRepoView := NewAddRepoView(v.store, v.ctx, v.handle, v.invocationCtx)
		return ViewResult{NextView: &addRepoView}
	case "l":
		removeRepoView := NewRemoveRepoView(v.store, v.ctx, v.handle)
		return ViewResult{NextView: removeRepoView}
	}
	return ViewResult{}
}

func (v *ReposMenuView) View() string {
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

type ReposMenuViewSnapshot struct {
	Type        string
	Handle      string
	ItemCount   int
	SelectedIdx int
}

func (v *ReposMenuView) Snapshot() interface{} {
	return ReposMenuViewSnapshot{
		Type:        "ReposMenuView",
		Handle:      v.handle,
		ItemCount:   len(v.list.Items()),
		SelectedIdx: v.list.Index(),
	}
}
