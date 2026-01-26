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

type RepoItem struct {
	repo *workspace.Repository
}

func (r RepoItem) Title() string {
	return r.repo.Name
}

func (r RepoItem) Description() string {
	if r.repo.Ref != "" {
		return fmt.Sprintf("%s @ %s", r.repo.URL, r.repo.Ref)
	}
	return r.repo.URL
}

func (r RepoItem) FilterValue() string {
	return r.repo.Name + " " + r.repo.URL
}

type RemoveRepoView struct {
	store     workspace.Store
	ctx       context.Context
	handle    string
	list      list.Model
	workspace *workspace.Workspace
	err       error
	stale     bool
	size      measure.Window
}

func NewRemoveRepoView(s workspace.Store, ctx context.Context, handle string) *RemoveRepoView {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 15)
	l.Title = "Select repository to remove from \"" + handle + "\""
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)

	v := &RemoveRepoView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		list:   l,
	}
	_ = v.refreshWorkspace()
	return v
}

func (v *RemoveRepoView) Init() tea.Cmd {
	return nil
}

func (v *RemoveRepoView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *RemoveRepoView) OnPush() {
	_ = v.refreshWorkspace()
}

func (v *RemoveRepoView) OnResume() {
	_ = v.refreshWorkspace()
}

func (v *RemoveRepoView) IsLoading() bool {
	return false
}

func (v *RemoveRepoView) Cancel() {}

func (v *RemoveRepoView) refreshWorkspace() error {
	ws, err := v.store.Get(v.ctx, v.handle)
	if err != nil {
		v.err = err
		return err
	}
	v.workspace = ws

	items := make([]list.Item, len(ws.Repositories))
	for i, repo := range ws.Repositories {
		items[i] = RepoItem{repo: &repo}
	}
	v.list.SetItems(items)
	v.err = nil
	return nil
}

func (v *RemoveRepoView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.stale {
		return ViewResult{Action: StackPop{}}, nil
	}

	if v.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				v.err = nil
				_ = v.refreshWorkspace()
				return ViewResult{}, nil
			}
			if msg.Type == tea.KeyEsc {
				return ViewResult{Action: StackPop{}}, nil
			}
		}
		return ViewResult{}, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.list.SetSize(40, 15)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			selected := v.list.SelectedItem()
			if selected != nil {
				if ri, ok := selected.(RepoItem); ok {
					return v.confirmRemove(ri.repo), nil
				}
			}
			return ViewResult{}, nil
		}
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return ViewResult{}, cmd
}

func (v *RemoveRepoView) confirmRemove(repo *workspace.Repository) ViewResult {
	return ViewResult{
		NextView: NewRemoveRepoConfirmView(v.store, v.ctx, v.handle, repo.Name),
	}
}

func (v *RemoveRepoView) View() string {
	if v.err != nil {
		return ErrorView(v.err, v.size)
	}

	if len(v.workspace.Repositories) == 0 {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(components.ColorText).
					Render("No repositories in workspace \""+v.handle+"\""),
				"\n",
				lipgloss.NewStyle().
					Foreground(components.ColorVeryMuted).
					MarginTop(1).
					Render("Press [Esc] to go back"),
			),
		)
	}

	return ModalFrame(v.size).Render(
		v.list.View() + "\n" +
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				MarginTop(1).
				Render("[↑↓/j/k] Navigate  [Enter] Select to remove  [Esc] Cancel"),
	)
}

func (v *RemoveRepoView) Snapshot() interface{} {
	return RemoveRepoViewSnapshot{
		Type:      "RemoveRepoView",
		Handle:    v.handle,
		RepoCount: len(v.workspace.Repositories),
		HasError:  v.err != nil,
	}
}

type RemoveRepoViewSnapshot struct {
	Type      string
	Handle    string
	RepoCount int
	HasError  bool
}

type RemoveRepoConfirmView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	repoName string
	err      error
	size     measure.Window
}

func NewRemoveRepoConfirmView(s workspace.Store, ctx context.Context, handle, repoName string) *RemoveRepoConfirmView {
	return &RemoveRepoConfirmView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		repoName: repoName,
	}
}

func (v *RemoveRepoConfirmView) Init() tea.Cmd {
	return nil
}

func (v *RemoveRepoConfirmView) SetSize(size measure.Window) {
	v.size = size
}

func (v *RemoveRepoConfirmView) OnPush() {}

func (v *RemoveRepoConfirmView) OnResume() {}

func (v *RemoveRepoConfirmView) IsLoading() bool {
	return false
}

func (v *RemoveRepoConfirmView) Cancel() {}

func (v *RemoveRepoConfirmView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyEnter:
			err := v.store.RemoveRepository(v.ctx, v.handle, v.repoName)
			if err != nil {
				v.err = err
				return ViewResult{}, nil
			}
			return ViewResult{Action: StackPopCount{Count: 2}}, nil
		}
	}
	return ViewResult{}, nil
}

func (v *RemoveRepoConfirmView) View() string {
	if v.err != nil {
		return ErrorView(v.err, v.size)
	}

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(components.ColorError).
				Render("Remove Repository?"),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorText).
				Render(fmt.Sprintf("Remove %q from workspace %q?", v.repoName, v.handle)),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				MarginTop(1).
				Render("[Enter] Confirm  [Esc] Cancel"),
		),
	)
}

func (v *RemoveRepoConfirmView) Snapshot() interface{} {
	return RemoveRepoConfirmViewSnapshot{
		Type:     "RemoveRepoConfirmView",
		Handle:   v.handle,
		RepoName: v.repoName,
		HasError: v.err != nil,
	}
}

type RemoveRepoConfirmViewSnapshot struct {
	Type     string
	Handle   string
	RepoName string
	HasError bool
}
