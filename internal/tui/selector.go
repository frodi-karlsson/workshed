package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/workspace"
)

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

type model struct {
	list           list.Model
	workspaces     []*workspace.Workspace
	hasSelected    bool
	selectedHandle string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if len(m.list.Items()) == 0 {
				return m, tea.Quit
			}
			selected := m.list.SelectedItem()
			if selected == nil {
				return m, tea.Quit
			}
			if item, ok := selected.(WorkspaceItem); ok {
				m.hasSelected = true
				m.selectedHandle = item.workspace.Handle
			}
			return m, tea.Quit
		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, min(msg.Height-v, maxListHeight))
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if len(m.workspaces) == 0 {
		return noWorkspacesView
	}
	return docStyle.Render(m.list.View())
}

var (
	noWorkspacesView = fmt.Sprintf(`%s

No workspaces found. Create one with:

    workshed create

`, lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorError).
		Render("No workspaces available"))

	docStyle = modalFrame()
)

type Selector struct {
	store workspace.Store
}

func NewSelector(s workspace.Store) *Selector {
	return &Selector{store: s}
}

func (s *Selector) Run(ctx context.Context) (string, error) {
	workspaces, err := s.store.List(ctx, workspace.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("listing workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces available")
	}

	items := make([]list.Item, len(workspaces))
	for i, ws := range workspaces {
		items[i] = WorkspaceItem{workspace: ws}
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, maxListHeight)
	l.Title = "Select workspace"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	applyTitleStyle(&l)

	m := model{
		list:       l,
		workspaces: workspaces,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("running TUI: %w", err)
	}

	if finalM, ok := finalModel.(model); ok && finalM.hasSelected {
		return finalM.selectedHandle, nil
	}

	return "", fmt.Errorf("selection cancelled")
}
