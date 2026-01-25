package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/tui/modalViews"
	"github.com/frodi/workshed/internal/tui/wizard"
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
	wizardModel     *wizard.Wizard
	wizardResult    *wizard.WizardResult
	wizardErr       error
	contextResult   string
	updatePurpose   string
	removeConfirm   bool
	contextMenuView *contextMenuView

	inspectModal    *modalViews.InspectModal
	pathModal       *modalViews.PathModal
	execModal       *modalViews.ExecModal
	execResultModal *modalViews.ExecResultModal
	updateModal     *modalViews.UpdateModal
	removeModal     *modalViews.RemoveModal
	helpModal       *modalViews.HelpModal
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
			wm, err := wizard.NewWizard(m.ctx, m.store)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.wizardModel = wm
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
			onDismiss := func() {
				m.currentView = viewDashboard
			}
			helpModal := modalViews.NewHelpModal(onDismiss)
			m.helpModal = &helpModal
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

						onDismiss := func() {
							m.currentView = viewDashboard
						}

						onConfirmRemove := func() {
							err := m.store.Remove(m.ctx, m.selectedHandle)
							if err != nil {
								m.err = err
							}
						}

						onConfirmUpdate := func(purpose string) {
							err := m.store.UpdatePurpose(m.ctx, m.selectedHandle, purpose)
							if err != nil {
								m.err = err
							}
						}

						inspectModal := modalViews.NewInspectModal(ws, onDismiss)
						pathModal := modalViews.NewPathModal(ws, onDismiss)
						execModal := modalViews.NewExecModal(ws, onDismiss)
						execResultModal := modalViews.NewExecResultModal(m.execCommand, m.execResults, onDismiss)
						updateModal := modalViews.NewUpdateModal(m.selectedHandle, onDismiss, onConfirmUpdate)
						removeModal := modalViews.NewRemoveModal(m.selectedHandle, onDismiss, onConfirmRemove)
						helpModal := modalViews.NewHelpModal(onDismiss)

						m.inspectModal = &inspectModal
						m.pathModal = &pathModal
						m.execModal = &execModal
						m.execResultModal = &execResultModal
						m.updateModal = &updateModal
						m.removeModal = &removeModal
						m.helpModal = &helpModal

						cm := NewContextMenuView(wi.workspace.Handle)
						m.contextMenuView = &cm
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
		return RenderDashboard(m.list, "[c] Create  [Enter] Menu  [l] Filter  [?] Help  [q/Esc] Quit", m.quitting)
	case viewContextMenu:
		return ViewContextMenu(m.contextMenuView)
	case viewInspectModal:
		return RenderInspectModal(m.inspectModal)
	case viewPathModal:
		return RenderPathModal(m.pathModal)
	case viewExecModal:
		return RenderExecModal(m.execModal)
	case viewExecResult:
		return RenderExecResultModal(m.execResultModal)
	case viewUpdateModal:
		return RenderUpdateModal(m.updateModal)
	case viewRemoveModal:
		return RenderRemoveModal(m.removeModal)
	case viewCreateWizard:
		return RenderWizard(m.wizardModel)
	case viewHelpModal:
		return RenderHelpModal(m.helpModal)
	case viewFilterInput:
		return RenderFilterInput(m.textInput, len(m.list.Items()))
	}

	return RenderDashboard(m.list, "[c] Create  [Enter] Menu  [l] Filter  [?] Help  [q/Esc] Quit", m.quitting)
}

func (m dashboardModel) errorView() string {
	return ErrorView(m.err)
}

func (m dashboardModel) updateContextMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.contextMenuView == nil {
		m.currentView = viewDashboard
		return m, nil
	}

	updated, cmd := m.contextMenuView.Update(msg)
	*m.contextMenuView = updated

	if m.contextMenuView.cancelled {
		m.contextMenuView = nil
		m.currentView = viewDashboard
		return m, nil
	}

	if m.contextMenuView.selectedAction != "" {
		switch m.contextMenuView.selectedAction {
		case "i":
			m.currentView = viewInspectModal
		case "p":
			ws, err := m.store.Get(m.ctx, m.selectedHandle)
			if err == nil {
				_ = clipboard.WriteAll(ws.Path)
			}
			m.currentView = viewPathModal
		case "e":
			m.currentView = viewExecModal
		case "u":
			m.currentView = viewUpdateModal
		case "r":
			m.currentView = viewRemoveModal
		}
		m.contextMenuView = nil
		return m, nil
	}

	return m, cmd
}

func (m dashboardModel) updateInspectModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.inspectModal == nil, viewInspectModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.inspectModal.Update(msg)
		*mPtr.inspectModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updatePathModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.pathModal == nil, viewPathModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.pathModal.Update(msg)
		*mPtr.pathModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updateExecModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.execModal == nil, viewExecModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.execModal.Update(msg)
		*mPtr.execModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updateExecResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.execResultModal == nil, viewExecResult, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.execResultModal.Update(msg)
		*mPtr.execResultModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updateUpdateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.updateModal == nil, viewUpdateModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.updateModal.Update(msg)
		*mPtr.updateModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updateRemoveModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.removeModal == nil, viewRemoveModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.removeModal.Update(msg)
		*mPtr.removeModal = updated
		return dismissed
	})
	return m, cmd
}

func (m dashboardModel) updateWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.wizardModel == nil {
		m.currentView = viewDashboard
		return m, nil
	}

	switch msg := msg.(type) {
	case wizard.WizardDoneMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.currentView = viewDashboard
			return m, nil
		}
		if msg.Cancelled {
			m.currentView = viewDashboard
			return m, nil
		}
		if msg.Result != nil {
			m.wizardResult = msg.Result
			ws, err := m.store.Create(m.ctx, workspace.CreateOptions{
				Purpose:      msg.Result.Purpose,
				Repositories: msg.Result.Repositories,
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

	_, cmd := m.wizardModel.Update(msg)
	return m, cmd
}

func (m dashboardModel) updateHelpModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := updateDismissableModal(m, m.helpModal == nil, viewHelpModal, msg, func(mPtr *dashboardModel) bool {
		updated, dismissed := mPtr.helpModal.Update(msg)
		*mPtr.helpModal = updated
		return dismissed
	})
	return m, cmd
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
