package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
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
	Exec(ctx context.Context, handle string, opts workspace.ExecOptions) ([]workspace.ExecResult, error)
}

type viewState int

const (
	viewDashboard viewState = iota
	viewContextMenu
	viewInspectModal
	viewPathModal
	viewExecModal
	viewExecResult
	viewUpdateModal
	viewRemoveModal
	viewCreateWizard
	viewHelpModal
	viewFilterInput
)

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
	selectedHandle string

	currentView     viewState
	modalWorkspace  *workspace.Workspace
	modalErr        error
	execResults     []workspace.ExecResult
	execCommand     string
	wizardModel     *wizardModel
	wizardResult    *WizardResult
	wizardErr       error
	contextResult   string
	updatePurpose   string
	removeConfirm   bool
	contextMenuList list.Model
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

func (m *dashboardModel) initContextMenuList(handle string) {
	items := []list.Item{
		menuItem{key: "i", label: "Inspect", desc: "Show workspace details", selected: false},
		menuItem{key: "p", label: "Path", desc: "Copy path to clipboard", selected: false},
		menuItem{key: "e", label: "Exec", desc: "Run command in repositories", selected: false},
		menuItem{key: "u", label: "Update", desc: "Change workspace purpose", selected: false},
		menuItem{key: "r", label: "Remove", desc: "Delete workspace (confirm)", selected: false},
	}

	l := list.New(items, list.NewDefaultDelegate(), contextMenuWidth, maxListHeight)
	l.Title = "Actions for \"" + handle + "\""
	l.SetShowTitle(true)
	applyCommonListStyles(&l)
	applyTitleStyle(&l)
	l.Styles.Title = l.Styles.Title.Width(contextMenuWidth)

	m.contextMenuList = l
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
	m.currentView = viewDashboard
	return m.loadWorkspacesCmd
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", " ":
				m.err = nil
				return m, m.loadWorkspacesCmd
			case "q", "ctrl+c", "esc":
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch m.currentView {
	case viewDashboard:
		return m.updateDashboard(msg)
	case viewContextMenu:
		return m.updateContextMenu(msg)
	case viewInspectModal:
		return m.updateInspectModal(msg)
	case viewPathModal:
		return m.updatePathModal(msg)
	case viewExecModal:
		return m.updateExecModal(msg)
	case viewExecResult:
		return m.updateExecResult(msg)
	case viewUpdateModal:
		return m.updateUpdateModal(msg)
	case viewRemoveModal:
		return m.updateRemoveModal(msg)
	case viewCreateWizard:
		return m.updateWizard(msg)
	case viewHelpModal:
		return m.updateHelpModal(msg)
	case viewFilterInput:
		return m.updateFilterInput(msg)
	}

	return m, nil
}

func (m dashboardModel) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
				m.currentView = viewDashboard
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
			wm, err := newWizardModel(m.ctx, m.store)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.wizardModel = &wm
			m.currentView = viewCreateWizard
			return m, textinput.Blink
		case "l":
			if !m.filterMode {
				m.filterMode = true
				m.currentView = viewFilterInput
				m.textInput.Focus()
				return m, textinput.Blink
			}
		case "?":
			m.currentView = viewHelpModal
			return m, nil
		}

		if m.filterMode {
			switch msg.String() {
			case "enter":
				m.filterMode = false
				m.textInput.Blur()
				m.currentView = viewDashboard
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
						ws, err := m.store.Get(m.ctx, wi.workspace.Handle)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.modalWorkspace = ws
						m.initContextMenuList(wi.workspace.Handle)
						m.currentView = viewContextMenu
						return m, nil
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

	switch m.currentView {
	case viewDashboard:
		return m.viewDashboard()
	case viewContextMenu:
		return m.viewContextMenu()
	case viewInspectModal:
		return m.viewInspectModal()
	case viewPathModal:
		return m.viewPathModal()
	case viewExecModal:
		return m.viewExecModal()
	case viewExecResult:
		return m.viewExecResult()
	case viewUpdateModal:
		return m.viewUpdateModal()
	case viewRemoveModal:
		return m.viewRemoveModal()
	case viewCreateWizard:
		return m.viewWizard()
	case viewHelpModal:
		return m.viewHelpModal()
	case viewFilterInput:
		return m.viewFilterInput()
	}

	return m.viewDashboard()
}

func (m dashboardModel) errorView() string {
	const maxWidth = 60
	errorMsg := m.err.Error()

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

func (m dashboardModel) updateContextMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.currentView = viewDashboard
			return m, nil
		case tea.KeyEnter:
			selected := m.contextMenuList.SelectedItem()
			if selected != nil {
				if mi, ok := selected.(menuItem); ok {
					m.contextResult = mi.key
					switch mi.key {
					case "i":
						m.currentView = viewInspectModal
					case "p":
						m.currentView = viewPathModal
					case "e":
						m.currentView = viewExecModal
					case "u":
						m.currentView = viewUpdateModal
					case "r":
						m.currentView = viewRemoveModal
					}
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.contextMenuList, cmd = m.contextMenuList.Update(msg)
	return m, cmd
}

func (m dashboardModel) viewContextMenu() string {
	frameStyle := modalFrame()

	return frameStyle.Render(
		m.contextMenuList.View() + "\n" +
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				MarginTop(1).
				Render("[↑↓/j/k] Navigate  [Enter] Select  [Esc/q/Ctrl+C] Cancel"),
	)
}

func (m dashboardModel) updateInspectModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	alertModal := NewEmbeddableAlertModal(buildWorkspaceDetailContent(m.modalWorkspace))
	if alertModal.Update(msg) {
		m.currentView = viewDashboard
	}
	return m, nil
}

func (m dashboardModel) viewInspectModal() string {
	alertModal := NewEmbeddableAlertModal(buildWorkspaceDetailContent(m.modalWorkspace))
	return alertModal.View()
}

func (m dashboardModel) updatePathModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	alertModal := NewEmbeddableAlertModal("")
	if alertModal.Update(msg) {
		m.currentView = viewDashboard
	}
	return m, nil
}

func (m dashboardModel) viewPathModal() string {
	ws := m.modalWorkspace

	helpStyle := lipgloss.NewStyle().
		Foreground(colorVeryMuted)

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
					Render(ws.Path),
				"\n\n",
				lipgloss.NewStyle().
					Foreground(colorSuccess).
					Render("Path copied to clipboard!"),
				"\n",
				helpStyle.Render("[Esc/q/Enter] Dismiss"),
			),
		)
}

func (m dashboardModel) updateExecModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.currentView = viewDashboard
			return m, nil
		case tea.KeyRunes:
			switch msg.String() {
			case "q":
				m.currentView = viewDashboard
				return m, nil
			}
		case tea.KeyEnter:
			m.currentView = viewDashboard
			return m, nil
		case tea.KeyTab:
			return m, nil
		case tea.KeySpace:
			return m, nil
		}
	}
	return m, nil
}

func (m dashboardModel) viewExecModal() string {
	return modalFrame().
		BorderForeground(colorSuccess).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Render("Command to run in \""+m.modalWorkspace.Handle+"\""),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					Render("Exec modal placeholder"),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					MarginTop(1).
					Render("[Tab] Switch  [Space] Toggle  [Enter] Run  [Esc] Cancel"),
			),
		)
}

func (m dashboardModel) updateExecResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "enter":
			m.currentView = viewDashboard
			return m, nil
		}
	}
	return m, nil
}

func (m dashboardModel) viewExecResult() string {
	if len(m.execResults) == 0 {
		return modalFrame().Render("No results")
	}

	borderColor := colorSuccess
	allSuccess := true
	for _, result := range m.execResults {
		if result.ExitCode != 0 {
			allSuccess = false
		}
	}
	if !allSuccess {
		borderColor = colorError
	}

	statusText := "Success"
	if !allSuccess {
		statusText = "Failed"
	}

	status := lipgloss.NewStyle().
		Foreground(borderColor).
		Render("[" + statusText + "]")

	return modalFrame().BorderForeground(borderColor).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().
					Bold(true).
					Foreground(colorText).
					Render("Command Execution Results"),
				"",
				lipgloss.JoinHorizontal(lipgloss.Left, status, "  ", m.execCommand),
				"\n",
				lipgloss.NewStyle().
					Foreground(colorVeryMuted).
					MarginTop(1).
					Render("[Enter/Esc/q] Close"),
			),
		)
}

func (m dashboardModel) updateUpdateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.currentView = viewDashboard
			return m, nil
		case "enter":
			p := strings.TrimSpace(m.textInput.Value())
			if p != "" {
				m.updatePurpose = p
				err := m.store.UpdatePurpose(m.ctx, m.selectedHandle, p)
				if err != nil {
					m.err = err
				}
			}
			m.currentView = viewDashboard
			return m, nil
		}
	}

	updatedInput, cmd := m.textInput.Update(msg)
	m.textInput = updatedInput
	return m, cmd
}

func (m dashboardModel) viewUpdateModal() string {
	frameStyle := modalFrame()

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Render("Update Purpose for \""+m.selectedHandle+"\""),
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

func (m dashboardModel) updateRemoveModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "n", "N":
			m.currentView = viewDashboard
			return m, nil
		case "y", "Y":
			m.removeConfirm = true
			err := m.store.Remove(m.ctx, m.selectedHandle)
			if err != nil {
				m.err = err
			}
			m.currentView = viewDashboard
			return m, nil
		}
	}
	return m, nil
}

func (m dashboardModel) viewRemoveModal() string {
	frameStyle := modalFrame().BorderForeground(colorError)
	if m.removeConfirm {
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
				Render(m.selectedHandle),
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

func (m dashboardModel) updateWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.wizardModel == nil {
		m.currentView = viewDashboard
		return m, nil
	}

	switch msg := msg.(type) {
	case wizardDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.currentView = viewDashboard
			return m, nil
		}
		if msg.cancelled {
			m.currentView = viewDashboard
			return m, nil
		}
		if msg.result != nil {
			m.wizardResult = msg.result
			ws, err := m.store.Create(m.ctx, workspace.CreateOptions{
				Purpose:      msg.result.Purpose,
				Repositories: msg.result.Repositories,
			})
			if err != nil {
				m.err = err
			} else {
				m.modalWorkspace = ws
			}
			m.currentView = viewDashboard
			if err := m.loadWorkspaces(); err != nil {
				m.err = err
			}
			return m, nil
		}
		m.currentView = viewDashboard
		return m, nil
	}

	updated, cmd := m.wizardModel.Update(msg)
	if updated, ok := updated.(wizardModel); ok {
		*m.wizardModel = updated
	}
	return m, cmd
}

func (m dashboardModel) viewWizard() string {
	if m.wizardModel == nil {
		return modalFrame().Render("Loading wizard...")
	}
	return m.wizardModel.View()
}

func (m dashboardModel) updateHelpModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	alertModal := NewEmbeddableAlertModal("")
	if alertModal.Update(msg) {
		m.currentView = viewDashboard
	}
	return m, nil
}

func (m dashboardModel) viewHelpModal() string {
	helpText := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Keyboard Shortcuts") + "\n\n" +
		"[c] Create workspace\n" +
		"[Enter] Open action menu for selected workspace\n" +
		"[l] Filter workspaces by purpose or handle\n" +
		"[↑/↓/j/k] Navigate workspaces\n" +
		"[?] Toggle this help\n" +
		"[q/Esc] Quit"

	helpStyle := lipgloss.NewStyle().
		Foreground(colorVeryMuted)

	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			"\n",
			helpStyle.Render("[Esc/q/Enter] Dismiss"),
		),
	)
}

func (m dashboardModel) updateFilterInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.filterMode = false
			m.textInput.Blur()
			m.currentView = viewDashboard
			if err := m.loadWorkspaces(); err != nil {
				m.err = err
			}
			return m, nil
		case "esc":
			m.filterMode = false
			m.textInput.Blur()
			m.currentView = viewDashboard
			return m, nil
		}
	}

	updatedInput, cmd := m.textInput.Update(msg)
	m.textInput = updatedInput
	if err := m.loadWorkspaces(); err != nil {
		m.err = err
	}
	return m, cmd
}

func (m dashboardModel) viewFilterInput() string {
	filterHint := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Render("[FILTER MODE] ") +
		m.textInput.View() +
		" (Enter to apply, Esc to cancel)"

	content := []string{filterHint}

	workspaceCount := len(m.list.Items())
	if workspaceCount == 0 {
		content = append(content, "\nNo workspaces found.")
	} else {
		content = append(content, m.list.View())
	}

	helpHint := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1).
		Render("[Enter] Apply filter  [Esc] Cancel filter")
	content = append(content, helpHint)

	return modalFrame().Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m dashboardModel) viewDashboard() string {
	var content []string

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Workshed Dashboard")
	content = append(content, header)

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

	content = append(content, helpHint)

	frameStyle := modalFrame()
	if m.quitting {
		frameStyle = frameStyle.BorderForeground(colorError)
	}
	return frameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func RunDashboard(ctx context.Context, store *workspace.FSStore) error {
	m := NewDashboardModel(ctx, store)

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
		if fm.err != nil {
			return fm.err
		}
	}

	return nil
}
