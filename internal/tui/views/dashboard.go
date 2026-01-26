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
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/store"
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
	store         store.Store
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

// NewDashboardView creates a new dashboard view.
func NewDashboardView(ctx context.Context, s store.Store, invocationCtx workspace.InvocationContext) DashboardView {
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

		if v.filterMode {
			if key.IsEnter(msg) {
				selected := v.list.SelectedItem()
				if selected != nil {
					if wi, ok := selected.(WorkspaceItem); ok {
						contextMenuView := NewContextMenuView(v.store, v.ctx, wi.workspace.Handle, v.invocationCtx)
						return ViewResult{NextView: &contextMenuView}, nil
					}
				}
				return ViewResult{}, nil
			}

			if key.IsUp(msg) || key.IsDown(msg) {
				var cmd tea.Cmd
				v.list, cmd = v.list.Update(msg)
				return ViewResult{}, cmd
			}

			if msg.String() == "s" {
				v.sortOrder = v.sortOrder.Next()
				_ = v.refreshWorkspaces()
				return ViewResult{}, nil
			}

			updatedInput, cmd := v.textInput.Update(msg)
			v.textInput = updatedInput
			_ = v.refreshWorkspaces()
			return ViewResult{}, cmd
		}

		switch msg.String() {
		case "c":
			wizardView := NewWizardView(v.ctx, v.store)
			return ViewResult{NextView: &wizardView}, nil
		case "l":
			v.filterMode = true
			v.textInput.Focus()
			return ViewResult{}, textinput.Blink
		}

		if key.IsEnter(msg) {
			selected := v.list.SelectedItem()
			if selected != nil {
				if wi, ok := selected.(WorkspaceItem); ok {
					contextMenuView := NewContextMenuView(v.store, v.ctx, wi.workspace.Handle, v.invocationCtx)
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
		content = append(content, "\nNo workspaces found. You can create one with:\n\n    workshed create\n")
	} else {
		content = append(content, v.list.View())
	}

	helpText := "[c] Create  [Enter] Menu  [l] Filter  [q/Esc] Quit"
	if v.filterMode {
		helpText = "[Enter] Menu  [↑/↓] Navigate  [s] Sort  [Esc] Cancel"
	}

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
