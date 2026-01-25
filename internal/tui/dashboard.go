package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type actionType int

const (
	actionNone actionType = iota
	actionCreate
	actionInspect
	actionPath
	actionExec
	actionUpdate
	actionRemove
)

type workspacesLoadedMsg struct {
	workspaces []*workspace.Workspace
	err        error
}

type store interface {
	List(ctx context.Context, opts workspace.ListOptions) ([]*workspace.Workspace, error)
	Get(ctx context.Context, handle string) (*workspace.Workspace, error)
	Create(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error)
	UpdatePurpose(ctx context.Context, handle string, purpose string) error
	Remove(ctx context.Context, handle string) error
}

type dashboardModel struct {
	list           list.Model
	textInput      textinput.Model
	store          store
	ctx            context.Context
	workspaces     []*workspace.Workspace
	filterMode     bool
	quitting       bool
	showHelp       bool
	err            error
	pendingAction  actionType
	selectedHandle string
}

func NewDashboardModel(ctx context.Context, store *workspace.FSStore) dashboardModel {
	ti := textinput.New()
	ti.Placeholder = "Filter workspaces..."
	ti.CharLimit = 100
	ti.Prompt = ""

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 30, maxListHeight)
	l.Title = "Workspaces"
	l.SetShowTitle(true)
	applyCommonListStyles(&l)
	applyTitleStyle(&l)

	return dashboardModel{
		list:       l,
		textInput:  ti,
		store:      store,
		ctx:        ctx,
		filterMode: false,
	}
}

func (m dashboardModel) loadWorkspacesCmd() tea.Msg {
	workspaces, err := m.store.List(m.ctx, workspace.ListOptions{})
	return workspacesLoadedMsg{workspaces: workspaces, err: err}
}

func (m *dashboardModel) loadWorkspaces() error {
	workspaces, err := m.store.List(m.ctx, workspace.ListOptions{})
	if err != nil {
		return err
	}
	m.workspaces = workspaces

	filterQuery := ""
	if m.filterMode {
		filterQuery = m.textInput.Value()
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

	m.list.SetItems(items)

	return nil
}

func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text)

	for i, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLine.WriteString(word)
		}

		if i == len(words)-1 {
			result.WriteString(currentLine.String())
		}
	}

	return result.String()
}

func (m dashboardModel) Init() tea.Cmd {
	return m.loadWorkspacesCmd
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", " ":
				m.err = nil
				return m, m.loadWorkspacesCmd
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case workspacesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.workspaces = msg.workspaces

		filterQuery := ""
		if m.filterMode {
			filterQuery = m.textInput.Value()
		}

		items := make([]list.Item, 0, len(msg.workspaces))
		for _, ws := range msg.workspaces {
			if filterQuery == "" {
				items = append(items, WorkspaceItem{workspace: ws})
			} else {
				filterVal := ws.Handle + " " + ws.Purpose
				if containsCaseInsensitive(filterVal, filterQuery) {
					items = append(items, WorkspaceItem{workspace: ws})
				}
			}
		}

		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.filterMode {
				m.filterMode = false
				m.textInput.Blur()
				if err := m.loadWorkspaces(); err != nil {
					m.err = err
				}
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "c":
			m.pendingAction = actionCreate
			return m, tea.Quit
		case "l":
			if !m.filterMode {
				m.filterMode = true
				m.textInput.Focus()
				return m, textinput.Blink
			}
		case "?":
			m.showHelp = !m.showHelp
		}

		if m.filterMode {
			switch msg.String() {
			case "enter":
				m.filterMode = false
				m.textInput.Blur()
				if err := m.loadWorkspaces(); err != nil {
					m.err = err
				}
			default:
				updatedInput, cmd := m.textInput.Update(msg)
				m.textInput = updatedInput
				cmds = append(cmds, cmd)
				if err := m.loadWorkspaces(); err != nil {
					m.err = err
				}
			}
		} else {
			switch msg.String() {
			case "enter":
				selected := m.list.SelectedItem()
				if selected != nil {
					if wi, ok := selected.(WorkspaceItem); ok {
						m.selectedHandle = wi.workspace.Handle
						result, err := ShowContextMenu(wi.workspace.Handle)
						if err != nil {
							m.err = err
						} else if result != "" {
							switch result {
							case "i":
								m.pendingAction = actionInspect
							case "p":
								m.pendingAction = actionPath
							case "e":
								m.pendingAction = actionExec
							case "u":
								m.pendingAction = actionUpdate
							case "r":
								m.pendingAction = actionRemove
							}
						}
						return m, tea.Quit
					}
				}
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m dashboardModel) View() string {
	if m.err != nil {
		return m.errorView()
	}

	if m.showHelp {
		return m.helpView()
	}

	var content []string

	if m.filterMode {
		filterHint := lipgloss.NewStyle().
			Foreground(colorSuccess).
			Render("[FILTER MODE] ") +
			m.textInput.View() +
			" (Enter to apply, Esc to cancel)"
		content = append(content, filterHint)
	} else {
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Render("Workshed Dashboard")
		content = append(content, header)
	}

	workspaceCount := len(m.list.Items())
	if workspaceCount == 0 {
		content = append(content, "\nNo workspaces found. Press 'c' to create one.")
	} else {
		content = append(content, m.list.View())
	}

	helpHint := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1).
		Render("[c] Create  [Enter] Menu  [l] Filter  [?] Help  [q/Esc] Quit")

	if m.filterMode {
		helpHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1).
			Render("[Enter] Apply filter  [Esc] Cancel filter")
	}

	content = append(content, helpHint)

	frameStyle := modalFrame()
	if m.quitting {
		frameStyle = frameStyle.BorderForeground(colorError)
	}
	return frameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m dashboardModel) helpView() string {
	helpText := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Keyboard Shortcuts") + "\n\n" +
		"[c] Create workspace\n" +
		"[Enter] Open action menu for selected workspace\n" +
		"[l] Filter workspaces by purpose or handle\n" +
		"[↑/↓/j/k] Navigate workspaces\n" +
		"[?] Toggle this help\n" +
		"[q/Esc] Quit\n\n" +
		lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("Press any key to return")

	return modalFrame().Render(helpText)
}

func (m dashboardModel) errorView() string {
	const maxWidth = 60
	errorMsg := m.err.Error()

	// Wrap error message if it's too long
	wrappedMsg := wrapText(errorMsg, maxWidth)

	return modalFrame().
		BorderForeground(colorError).
		Width(maxWidth).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Foreground(colorError).Render("Error"),
				"\n",
				lipgloss.NewStyle().Foreground(colorText).Render(wrappedMsg),
				"\n",
				lipgloss.NewStyle().Foreground(colorVeryMuted).MarginTop(1).Render("[Enter] Dismiss  [q] Quit"),
			),
		)
}

func RunDashboard(ctx context.Context, store *workspace.FSStore) error {
	var pendingError error

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m := NewDashboardModel(ctx, store)
		if pendingError != nil {
			m.err = pendingError
			pendingError = nil
		}

		p := tea.NewProgram(m, tea.WithAltScreen())

		ctxCancel, cancel := context.WithCancel(ctx)
		go func() {
			<-ctxCancel.Done()
			p.Quit()
		}()

		finalModel, err := p.Run()
		cancel()
		if err != nil {
			return fmt.Errorf("running dashboard: %w", err)
		}

		if fm, ok := finalModel.(dashboardModel); ok {
			m = fm
		}

		if m.err != nil {
			return m.err
		}

		if m.quitting {
			return nil
		}

		switch m.pendingAction {
		case actionCreate:
			result, err := RunCreateWizard(ctx, store)
			if err != nil {
				if !errors.Is(err, ErrCancelled) {
					pendingError = err
				}
				continue
			}
			if result != nil {
				_, err := store.Create(ctx, workspace.CreateOptions{
					Purpose:      result.Purpose,
					Repositories: result.Repositories,
				})
				if err != nil {
					pendingError = fmt.Errorf("creating workspace: %w", err)
					continue
				}
			}
		case actionInspect:
			if err := ShowInspectModal(ctx, store, m.selectedHandle); err != nil {
				pendingError = err
				continue
			}
		case actionPath:
			if err := ShowPathModal(ctx, store, m.selectedHandle); err != nil {
				pendingError = err
				continue
			}
		case actionExec:
			execResult, err := ShowExecModal(ctx, store, m.selectedHandle)
			if err != nil {
				pendingError = err
				continue
			}
			if execResult != nil {
				cmdParts := strings.Fields(execResult.Command)
				if len(cmdParts) > 0 {
					execOpts := workspace.ExecOptions{
						Command: cmdParts,
					}

					var allResults []workspace.ExecResult
					var execErr error

					if len(execResult.RepoNames) > 0 {
						for _, repoName := range execResult.RepoNames {
							execOpts.Target = repoName
							results, err := store.Exec(ctx, m.selectedHandle, execOpts)
							if err != nil {
								execErr = err
								break
							}
							allResults = append(allResults, results...)
						}
					} else {
						results, err := store.Exec(ctx, m.selectedHandle, execOpts)
						if err != nil {
							execErr = err
						} else {
							allResults = append(allResults, results...)
						}
					}

					if len(allResults) > 0 {
						showErr := ShowExecResultModal(allResults, execResult.Command)
						if showErr != nil {
							pendingError = showErr
						}
					} else if execErr != nil {
						pendingError = fmt.Errorf("exec: %w", execErr)
					}
				}
			}
		case actionUpdate:
			if err := ShowUpdateModal(ctx, store, m.selectedHandle); err != nil {
				pendingError = err
				continue
			}
		case actionRemove:
			if err := ShowRemoveModal(ctx, store, m.selectedHandle); err != nil {
				pendingError = err
				continue
			}
		}
	}
}

type pathModel struct {
	path     string
	quitting bool
}

func (m pathModel) Init() tea.Cmd {
	return nil
}

func (m pathModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m pathModel) View() string {
	return modalFrame().
		BorderForeground(colorSuccess).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Render("Workspace Path"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorMuted).
					Render(m.path),
				"\n\n",
				lipgloss.NewStyle().
					Foreground(colorSuccess).
					Render("Path copied to clipboard!"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					MarginTop(1).
					Render("[Press any key to return]"),
			),
		)
}

func ShowPathModal(ctx context.Context, store *workspace.FSStore, handle string) error {
	ws, err := store.Get(ctx, handle)
	if err != nil {
		return err
	}

	m := pathModel{path: ws.Path}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running path modal: %w", err)
	}

	_ = os.WriteFile("/tmp/workshed_path", []byte(ws.Path), 0644)

	return nil
}

type updateModel struct {
	textInput textinput.Model
	purpose   string
	handle    string
	quitting  bool
	done      bool
}

func (m updateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m updateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			p := strings.TrimSpace(m.textInput.Value())
			if p != "" {
				m.purpose = p
				m.done = true
			}
			return m, tea.Quit
		}
	}

	updatedInput, cmd := m.textInput.Update(msg)
	m.textInput = updatedInput
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m updateModel) View() string {
	frameStyle := modalFrame()
	if m.done {
		frameStyle = frameStyle.BorderForeground(colorSuccess)
	}

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Render("Update Purpose for \""+m.handle+"\""),
			"\n",
			m.textInput.View(),
			"\n",
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				MarginTop(1).
				Render("[Enter] Save  [Esc] Cancel"),
		),
	)
}

func ShowUpdateModal(ctx context.Context, store *workspace.FSStore, handle string) error {
	ws, err := store.Get(ctx, handle)
	if err != nil {
		return err
	}

	ti := textinput.New()
	ti.Placeholder = "New purpose..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Prompt = ""
	ti.SetValue(ws.Purpose)

	m := updateModel{
		textInput: ti,
		handle:    handle,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running update modal: %w", err)
	}

	if fm, ok := finalModel.(updateModel); ok {
		if !fm.quitting && fm.done && fm.purpose != "" {
			if err := store.UpdatePurpose(ctx, handle, fm.purpose); err != nil {
				return fmt.Errorf("updating purpose: %w", err)
			}
		}
	}

	return nil
}

type removeModel struct {
	handle    string
	purpose   string
	confirmed bool
	quitting  bool
}

func (m removeModel) Init() tea.Cmd {
	return nil
}

func (m removeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "n", "N":
			m.quitting = true
			return m, tea.Quit
		case "y", "Y":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m removeModel) View() string {
	frameStyle := modalFrame().BorderForeground(colorError)
	if m.confirmed {
		frameStyle = frameStyle.BorderForeground(colorSuccess)
	}

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorError).
				Render("Remove Workspace?"),
			"\n",
			lipgloss.NewStyle().
				Foreground(colorText).
				Render(m.handle+" - "+m.purpose),
			"\n\n",
			lipgloss.NewStyle().
				Foreground(colorMuted).
				Render("This will delete the workspace directory."),
			"\n\n",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Render("[y] Yes  [n] No"),
		),
	)
}

func ShowRemoveModal(ctx context.Context, store *workspace.FSStore, handle string) error {
	ws, err := store.Get(ctx, handle)
	if err != nil {
		return err
	}

	m := removeModel{
		handle:  handle,
		purpose: ws.Purpose,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running remove modal: %w", err)
	}

	return nil
}
