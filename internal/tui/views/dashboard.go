package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type SortOrder int

const (
	SortCreatedDesc SortOrder = iota
	SortCreatedAsc
	SortPurposeAsc
	SortPurposeDesc
	SortHandleAsc
	SortHandleDesc
)

func (s SortOrder) String() string {
	switch s {
	case SortCreatedDesc:
		return "Created ↓"
	case SortCreatedAsc:
		return "Created ↑"
	case SortPurposeAsc:
		return "Purpose ↑"
	case SortPurposeDesc:
		return "Purpose ↓"
	case SortHandleAsc:
		return "Handle ↑"
	case SortHandleDesc:
		return "Handle ↓"
	default:
		return "Created ↓"
	}
}

func (s SortOrder) Next() SortOrder {
	return (s + 1) % 6
}

type DashboardView struct {
	store         workspace.Store
	ctx           context.Context
	list          list.Model
	textInput     textinput.Model
	filterMode    bool
	sortOrder     SortOrder
	err           error
	size          measure.Window
	invocationCtx workspace.InvocationContext
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

func NewDashboardView(ctx context.Context, s workspace.Store, invocationCtx workspace.InvocationContext) DashboardView {
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
		store:         s,
		ctx:           ctx,
		list:          l,
		textInput:     ti,
		filterMode:    false,
		invocationCtx: invocationCtx,
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

	sort.Slice(items, func(i, j int) bool {
		wi := items[i].(WorkspaceItem)
		wj := items[j].(WorkspaceItem)

		switch v.sortOrder {
		case SortCreatedDesc:
			return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
		case SortCreatedAsc:
			return wi.workspace.CreatedAt.Before(wj.workspace.CreatedAt)
		case SortPurposeAsc:
			if wi.workspace.Purpose == wj.workspace.Purpose {
				return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
			}
			return strings.ToLower(wi.workspace.Purpose) < strings.ToLower(wj.workspace.Purpose)
		case SortPurposeDesc:
			if wi.workspace.Purpose == wj.workspace.Purpose {
				return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
			}
			return strings.ToLower(wi.workspace.Purpose) > strings.ToLower(wj.workspace.Purpose)
		case SortHandleAsc:
			if wi.workspace.Handle == wj.workspace.Handle {
				return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
			}
			return strings.ToLower(wi.workspace.Handle) < strings.ToLower(wj.workspace.Handle)
		case SortHandleDesc:
			if wi.workspace.Handle == wj.workspace.Handle {
				return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
			}
			return strings.ToLower(wi.workspace.Handle) > strings.ToLower(wj.workspace.Handle)
		default:
			return wi.workspace.CreatedAt.After(wj.workspace.CreatedAt)
		}
	})

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

func (v *DashboardView) KeyBindings() []KeyBinding {
	if v.err != nil {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Retry", Action: v.retry},
			{Key: "esc", Help: "[Esc] Quit", Action: v.quit},
			{Key: "ctrl+c", Help: "[Ctrl+C] Quit", Action: v.quit},
		}
	}
	if v.filterMode {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Menu", Action: v.openMenu},
			{Key: "up", Help: "[↑] Navigate", Action: nil},
			{Key: "down", Help: "[↓] Navigate", Action: nil},
			{Key: "k", Help: "[k] Navigate", Action: nil},
			{Key: "j", Help: "[j] Navigate", Action: nil},
			{Key: "s", Help: "[s] Sort", Action: v.cycleSort},
			{Key: "esc", Help: "[Esc] Cancel", Action: v.cancelFilter},
		}
	}
	return []KeyBinding{
		{Key: "c", Help: "[c] Create", Action: v.createWorkspace},
		{Key: "enter", Help: "[Enter] Menu", Action: v.openMenu},
		{Key: "l", Help: "[l] Filter", Action: v.enableFilter},
		{Key: "q", Help: "[q] Quit", Action: v.quit},
		{Key: "esc", Help: "[Esc] Quit", Action: v.quit},
		{Key: "ctrl+c", Help: "[Ctrl+C] Quit", Action: v.quit},
	}
}

func (v *DashboardView) createWorkspace() (ViewResult, tea.Cmd) {
	wizardView := NewWizardView(v.ctx, v.store)
	return ViewResult{NextView: &wizardView}, nil
}

func (v *DashboardView) openMenu() (ViewResult, tea.Cmd) {
	selected := v.list.SelectedItem()
	if selected != nil {
		if wi, ok := selected.(WorkspaceItem); ok {
			resourceMenuView := NewResourceMenuView(v.store, v.ctx, wi.workspace.Handle, v.invocationCtx)
			return ViewResult{NextView: &resourceMenuView}, nil
		}
	}
	return ViewResult{}, nil
}

func (v *DashboardView) enableFilter() (ViewResult, tea.Cmd) {
	v.filterMode = true
	v.textInput.Focus()
	return ViewResult{}, textinput.Blink
}

func (v *DashboardView) cycleSort() (ViewResult, tea.Cmd) {
	v.sortOrder = v.sortOrder.Next()
	_ = v.refreshWorkspaces()
	return ViewResult{}, nil
}

func (v *DashboardView) cancelFilter() (ViewResult, tea.Cmd) {
	v.filterMode = false
	v.textInput.Blur()
	v.textInput.SetValue("")
	_ = v.refreshWorkspaces()
	return ViewResult{}, nil
}

func (v *DashboardView) quit() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackDismissAll{}}, nil
}

func (v *DashboardView) retry() (ViewResult, tea.Cmd) {
	v.err = nil
	_ = v.refreshWorkspaces()
	return ViewResult{}, nil
}

func (v *DashboardView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.err != nil {
		if km, ok := msg.(tea.KeyMsg); ok {
			if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
				return result, nil
			}
		}
		return ViewResult{}, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	if v.filterMode {
		updatedInput, cmd := v.textInput.Update(msg)
		v.textInput = updatedInput
		_ = v.refreshWorkspaces()
		return ViewResult{}, cmd
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return ViewResult{}, cmd
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
		content = append(content, "\nNo workspaces found. You can create one with:\n\n    workshed create\n")
	} else {
		content = append(content, v.list.View())
	}

	helpText := GenerateHelp(v.KeyBindings())

	helpHint := lipgloss.NewStyle().
		Foreground(components.ColorMuted).
		MarginTop(1).
		Render(helpText)

	content = append(content, helpHint)

	if v.filterMode {
		sortInfo := lipgloss.NewStyle().
			Foreground(components.ColorVeryMuted).
			Render(fmt.Sprintf("Sort: %s", v.sortOrder.String()))
		content = append(content, sortInfo)
	}

	frameStyle := ModalFrame(v.size)
	return frameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

type DashboardViewSnapshot struct {
	Type          string
	FilterMode    bool
	FilterQuery   string
	SortOrder     string
	ItemCount     int
	SelectedIndex int
	HasError      bool
}

func (v *DashboardView) Snapshot() interface{} {
	return DashboardViewSnapshot{
		Type:          "DashboardView",
		FilterMode:    v.filterMode,
		FilterQuery:   v.textInput.Value(),
		SortOrder:     v.sortOrder.String(),
		ItemCount:     len(v.list.Items()),
		SelectedIndex: v.list.Index(),
		HasError:      v.err != nil,
	}
}
