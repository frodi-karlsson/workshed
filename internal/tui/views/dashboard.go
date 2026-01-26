package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type DashboardView struct {
	store      store.Store
	ctx        context.Context
	list       list.Model
	textInput  textinput.Model
	filterMode bool
	err        error
	size       measure.Window
}

type WorkspaceItem struct {
	workspace *workspace.Workspace
}

func (w WorkspaceItem) Title() string { return w.workspace.Handle }

func (w WorkspaceItem) Description() string {
	repoCount := len(w.workspace.Repositories)
	repoStr := "repos"
	if repoCount == 1 {
		repoStr = "repo"
	}
	return fmt.Sprintf("%s • %d %s • %s",
		w.workspace.Purpose,
		repoCount,
		repoStr,
		w.workspace.CreatedAt.Format("Jan 2"),
	)
}

func (w WorkspaceItem) FilterValue() string {
	return fmt.Sprintf("%s %s", w.workspace.Handle, w.workspace.Purpose)
}

func NewDashboardView(ctx context.Context, s store.Store) DashboardView {
	ti := textinput.New()
	ti.Placeholder = "Filter workspaces..."
	ti.CharLimit = 100
	ti.Prompt = ""

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 30, MaxListHeight)
	l.Title = "Workspaces"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Background(components.ColorBackground).
		Padding(0, 1)

	v := DashboardView{
		store:      s,
		ctx:        ctx,
		list:       l,
		textInput:  ti,
		filterMode: false,
	}
	_ = v.refreshWorkspaces()
	return v
}

func (v *DashboardView) OnPush() {
	_ = v.refreshWorkspaces()
}

func (v *DashboardView) OnResume() {
	_ = v.refreshWorkspaces()
}

func (v *DashboardView) IsLoading() bool {
	return false
}

func (v *DashboardView) Cancel() {}

func (v *DashboardView) refreshWorkspaces() error {
	workspaces, err := v.store.List(v.ctx, workspace.ListOptions{})
	if err != nil {
		v.err = err
		return err
	}

	filterQuery := ""
	if v.filterMode {
		filterQuery = v.textInput.Value()
	}

	items := make([]list.Item, 0, len(workspaces))
	for _, ws := range workspaces {
		if filterQuery == "" {
			items = append(items, WorkspaceItem{workspace: ws})
		} else {
			filterVal := ws.Handle + " " + ws.Purpose
			if containsCaseInsensitive(filterVal, filterQuery) {
				items = append(items, WorkspaceItem{workspace: ws})
			}
		}
	}

	v.list.SetItems(items)
	v.err = nil
	return nil
}

func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func (v *DashboardView) Init() tea.Cmd {
	return nil
}

func (v *DashboardView) SetSize(size measure.Window) {
	v.size = size
	v.list.SetSize(size.ListWidth(), size.ListHeight())
}

func (v *DashboardView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.err != nil {
		if key.IsEnter(msg) {
			v.err = nil
			_ = v.refreshWorkspaces()
			return ViewResult{}, nil
		}
		if key.IsCancel(msg) {
			return ViewResult{Action: StackDismissAll{}}, nil
		}
		return ViewResult{}, nil
	}

	if key.IsCancel(msg) {
		if v.filterMode {
			v.filterMode = false
			v.textInput.Blur()
			v.textInput.SetValue("")
			_ = v.refreshWorkspaces()
			return ViewResult{}, nil
		}
		return ViewResult{Action: StackPop{}}, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "c":
			wizardView := NewWizardView(v.ctx, v.store)
			return ViewResult{NextView: &wizardView}, nil
		case "l":
			if !v.filterMode {
				v.filterMode = true
				v.textInput.Focus()
				return ViewResult{}, textinput.Blink
			}
		case "?":
			helpView := NewHelpView()
			return ViewResult{NextView: &helpView}, nil
		}

		if v.filterMode {
			if key.IsEnter(msg) {
				v.filterMode = false
				v.textInput.Blur()
				_ = v.refreshWorkspaces()
				return ViewResult{}, nil
			}
			updatedInput, cmd := v.textInput.Update(msg)
			v.textInput = updatedInput
			_ = v.refreshWorkspaces()
			return ViewResult{}, cmd
		}

		if key.IsEnter(msg) {
			selected := v.list.SelectedItem()
			if selected != nil {
				if wi, ok := selected.(WorkspaceItem); ok {
					contextMenuView := NewContextMenuView(v.store, v.ctx, wi.workspace.Handle)
					return ViewResult{NextView: &contextMenuView}, nil
				}
			}
		}

		var cmd tea.Cmd
		v.list, cmd = v.list.Update(msg)
		return ViewResult{}, cmd
	}

	return ViewResult{}, nil
}

func (v *DashboardView) View() string {
	if v.err != nil {
		return ErrorView(v.err, v.size)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText).
		Render("Workshed Dashboard")

	content := []string{header}

	workspaceCount := len(v.list.Items())
	if workspaceCount == 0 {
		content = append(content, "\nNo workspaces found. Press 'c' to create one.")
	} else {
		content = append(content, v.list.View())
	}

	helpText := "[c] Create  [Enter] Menu  [l] Filter  [?] Help  [q/Esc] Quit"
	if v.filterMode {
		helpText = "[Enter] Apply  [Esc] Cancel"
	}

	helpHint := lipgloss.NewStyle().
		Foreground(components.ColorMuted).
		MarginTop(1).
		Render(helpText)

	content = append(content, helpHint)

	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

type DashboardViewSnapshot struct {
	Type          string
	FilterMode    bool
	FilterQuery   string
	ItemCount     int
	SelectedIndex int
	HasError      bool
}

func (v *DashboardView) Snapshot() interface{} {
	return DashboardViewSnapshot{
		Type:          "DashboardView",
		FilterMode:    v.filterMode,
		FilterQuery:   v.textInput.Value(),
		ItemCount:     len(v.list.Items()),
		SelectedIndex: v.list.Index(),
		HasError:      v.err != nil,
	}
}
