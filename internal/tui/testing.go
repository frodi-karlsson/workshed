package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tealist "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

const testTermWidth = 80
const testTermHeight = 24
const testTimeout = 2 * time.Second

var errWorkspaceNotFound = errors.New("workspace not found")

type mockStore struct {
	workspaces []*workspace.Workspace
	getErr     error
	listErr    error
	createErr  error
	updateErr  error
	removeErr  error
}

func newMockStore(t *testing.T, workspaces []*workspace.Workspace) *mockStore {
	return &mockStore{
		workspaces: workspaces,
	}
}

func (m *mockStore) List(ctx context.Context, opts workspace.ListOptions) ([]*workspace.Workspace, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	if opts.PurposeFilter == "" {
		return m.workspaces, nil
	}

	var result []*workspace.Workspace
	filter := strings.ToLower(opts.PurposeFilter)
	for _, ws := range m.workspaces {
		if strings.Contains(strings.ToLower(ws.Purpose), filter) {
			result = append(result, ws)
		}
	}
	return result, nil
}

func (m *mockStore) Get(ctx context.Context, handle string) (*workspace.Workspace, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, ws := range m.workspaces {
		if ws.Handle == handle {
			return ws, nil
		}
	}
	return nil, errWorkspaceNotFound
}

func (m *mockStore) Create(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	ws := &workspace.Workspace{
		Handle:       generateTestHandle(opts.Purpose),
		Purpose:      opts.Purpose,
		Path:         "/test/workspaces/" + generateTestHandle(opts.Purpose),
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}
	m.workspaces = append(m.workspaces, ws)
	return ws, nil
}

func (m *mockStore) UpdatePurpose(ctx context.Context, handle string, purpose string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for _, ws := range m.workspaces {
		if ws.Handle == handle {
			ws.Purpose = purpose
			return nil
		}
	}
	return errWorkspaceNotFound
}

func (m *mockStore) Remove(ctx context.Context, handle string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	for i, ws := range m.workspaces {
		if ws.Handle == handle {
			m.workspaces = append(m.workspaces[:i], m.workspaces[i+1:]...)
			return nil
		}
	}
	return errWorkspaceNotFound
}

func (m *mockStore) Exec(ctx context.Context, handle string, opts workspace.ExecOptions) ([]workspace.ExecResult, error) {
	return nil, nil
}

func (m *mockStore) FindWorkspace(ctx context.Context, dir string) (*workspace.Workspace, error) {
	return nil, errWorkspaceNotFound
}

func generateTestHandle(purpose string) string {
	if len(purpose) == 0 {
		return "test-workspace"
	}
	words := strings.Fields(purpose)
	if len(words) == 0 {
		return "test-workspace"
	}
	result := strings.ToLower(words[0])
	for i := 1; i < len(words) && i < 3; i++ {
		result += "-" + strings.ToLower(words[i])
	}
	return result
}

func newTestDashboardModel(t *testing.T, store store) dashboardModel {
	return dashboardModel{
		list:       newTestListModel(),
		textInput:  newTestTextInputModel(),
		store:      store,
		ctx:        context.Background(),
		workspaces: nil,
		filterMode: false,
		quitting:   false,
		showHelp:   false,
		err:        nil,

		currentView:    viewDashboard,
		modalWorkspace: nil,
		modalErr:       nil,
		execResults:    nil,
		execCommand:    "",
		wizardResult:   nil,
		wizardErr:      nil,
		contextResult:  "",
		updatePurpose:  "",
		removeConfirm:  false,
	}
}

func newTestListModel() tealist.Model {
	return tealist.New([]tealist.Item{}, tealist.NewDefaultDelegate(), testTermWidth, testTermHeight-5)
}

func newTestTextInputModel() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Prompt = ""
	return ti
}

func newTestContextMenuModel(handle string) contextMenuModel {
	items := []tealist.Item{
		menuItem{key: "i", label: "Inspect", desc: "Show workspace details", selected: false},
		menuItem{key: "p", label: "Path", desc: "Copy path to clipboard", selected: false},
		menuItem{key: "e", label: "Exec", desc: "Run command in repositories", selected: false},
		menuItem{key: "u", label: "Update", desc: "Change workspace purpose", selected: false},
		menuItem{key: "r", label: "Remove", desc: "Delete workspace (confirm)", selected: false},
	}

	l := tealist.New(items, tealist.NewDefaultDelegate(), 30, 8)
	l.Title = "Actions for \"" + handle + "\""
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Background(colorBackground).
		Padding(0, 1)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(colorMuted)
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(colorMuted)

	return contextMenuModel{
		list:   l,
		done:   false,
		quit:   false,
		result: "",
	}
}

func pressKey(tm *teatest.TestModel, key string) {
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func pressEnter(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
}

func pressEsc(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
}

func pressCtrlC(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
}

func navigateDown(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
}

func navigateUp(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
}

func requireOutputContains(t *testing.T, tm *teatest.TestModel, substr string) {
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), substr)
	}, teatest.WithCheckInterval(50*time.Millisecond), teatest.WithDuration(testTimeout))
}

func requireOutputContainsWithTimeout(t *testing.T, tm *teatest.TestModel, substr string, timeout time.Duration) {
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), substr)
	}, teatest.WithCheckInterval(50*time.Millisecond), teatest.WithDuration(timeout))
}

func quitAndWait(tm *teatest.TestModel, t *testing.T) {
	if err := tm.Quit(); err != nil {
		t.Logf("Quit error: %v", err)
	}
	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func sendKey(m dashboardModel, key string) (dashboardModel, tea.Cmd) {
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	updatedModel, cmd := m.Update(keyMsg)
	return updatedModel.(dashboardModel), cmd
}

func newTestExecModel(ws *workspace.Workspace) execModel {
	repoItems := make([]tealist.Item, len(ws.Repositories))
	for i, repo := range ws.Repositories {
		repoItems[i] = repoItem{repo: &repo, selected: true}
	}

	ti := textinput.New()
	ti.CharLimit = 200
	ti.Prompt = ""

	l := tealist.New(repoItems, tealist.NewDefaultDelegate(), 30, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return execModel{
		textInput: ti,
		list:      l,
		workspace: ws,
		focus:     focusInput,
		done:      false,
		quit:      false,
	}
}

func newTestPurposeStepModel(workspaces []*workspace.Workspace) purposeStepModel {
	return newPurposeStepModel(workspaces)
}

func newTestRepoStepModel(workspaces []*workspace.Workspace, purpose string) repoStepModel {
	return newRepoStepModel(workspaces, purpose)
}

func pressTab(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
}

func typeText(tm *teatest.TestModel, text string) {
	for _, r := range text {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}
