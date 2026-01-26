package snapshot

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/tui"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/gkampitakis/go-snaps/snaps"
)

const (
	testTermWidth  = 80
	testTermHeight = 24
)

// SyncTestHarness provides synchronous command execution for testing Bubble Tea models
type SyncTestHarness struct {
	model       tui.StackModel
	ctx         context.Context
	t           *testing.T
	pendingCmds []tea.Cmd
}

func NewSyncTestHarness(t *testing.T, model tui.StackModel, ctx context.Context) *SyncTestHarness {
	return &SyncTestHarness{model: model, ctx: ctx, t: t}
}

func (h *SyncTestHarness) Send(msg tea.Msg) {
	newModel, cmd := h.model.Update(msg)
	h.model = newModel.(tui.StackModel)
	if cmd != nil {
		h.pendingCmds = append(h.pendingCmds, cmd)
	}
}

func (h *SyncTestHarness) Pump() {
	iterations := 0
	for len(h.pendingCmds) > 0 && iterations < 100 {
		iterations++
		cmds := h.pendingCmds
		h.pendingCmds = nil

		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}

			// Execute command with timeout to skip timer-based visual effects
			msgChan := make(chan tea.Msg, 1)
			go func() {
				msgChan <- cmd()
			}()

			select {
			case msg := <-msgChan:
				if msg != nil {
					h.Send(msg)
				}
			case <-time.After(10 * time.Millisecond):
				// Skip slow commands (timer-based visual effects like cursor blink)
				continue
			}
		}
	}

	if iterations >= 100 {
		h.t.Fatal("Pump exceeded maximum iterations - possible infinite command loop")
	}
}

type StepType string

const (
	StepKey   StepType = "key"
	StepInput StepType = "input"
	StepEnter StepType = "enter"
)

type Step struct {
	Type        StepType
	Key         string
	Text        string
	Description string
}

type GitOption func(*git.MockGit)

func WithGitRemoteURL(url string) GitOption {
	return func(m *git.MockGit) {
		m.SetGetRemoteURLResult(url)
	}
}

func WithGitCloneError(err error) GitOption {
	return func(m *git.MockGit) {
		m.SetCloneErr(err)
	}
}

func WithGitCheckoutError(err error) GitOption {
	return func(m *git.MockGit) {
		m.SetCheckoutErr(err)
	}
}

func WithGitRemoteError(err error) GitOption {
	return func(m *git.MockGit) {
		m.SetGetRemoteErr(err)
	}
}

type Scenario struct {
	t       *testing.T
	mockGit *git.MockGit
	steps   []Step
	harness *SyncTestHarness
	ctx     context.Context
	store   *mockStore
}

type StoreOption func(*mockStore)

func WithCreateError(err error) StoreOption {
	return func(s *mockStore) {
		s.createErr = err
	}
}

func WithCreateDelay(duration time.Duration) StoreOption {
	return func(s *mockStore) {
		s.createDelay = duration
	}
}

func WithWorkspaces(workspaces []*workspace.Workspace) StoreOption {
	return func(s *mockStore) {
		s.workspaces = workspaces
	}
}

func WithListDelay(duration time.Duration) StoreOption {
	return func(s *mockStore) {
		s.listDelay = duration
	}
}

func WithListError(err error) StoreOption {
	return func(s *mockStore) {
		s.listErr = err
	}
}

type mockStore struct {
	mockGit     git.Git
	workspaces  []*workspace.Workspace
	createErr   error
	createDelay time.Duration
	listDelay   time.Duration
	listErr     error
}

func (s *mockStore) Create(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error) {
	if s.createErr != nil {
		err := s.createErr
		s.createErr = nil
		return nil, err
	}
	if s.createDelay > 0 {
		time.Sleep(s.createDelay)
	}
	ws := &workspace.Workspace{
		Handle:       "test-workspace",
		Purpose:      opts.Purpose,
		Repositories: make([]workspace.Repository, len(opts.Repositories)),
	}
	for i, r := range opts.Repositories {
		ws.Repositories[i] = workspace.Repository{
			URL:  r.URL,
			Ref:  r.Ref,
			Name: "repo-" + string(rune('1'+i)),
		}
	}
	return ws, nil
}

func (s *mockStore) Get(ctx context.Context, handle string) (*workspace.Workspace, error) {
	for _, ws := range s.workspaces {
		if ws.Handle == handle {
			return ws, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *mockStore) List(ctx context.Context, opts workspace.ListOptions) ([]*workspace.Workspace, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	if s.listDelay > 0 {
		time.Sleep(s.listDelay)
	}
	return s.workspaces, nil
}

func (s *mockStore) Remove(ctx context.Context, handle string) error {
	return nil
}

func (s *mockStore) Path(ctx context.Context, handle string) (string, error) {
	return "", nil
}

func (s *mockStore) UpdatePurpose(ctx context.Context, handle string, purpose string) error {
	return nil
}

func (s *mockStore) FindWorkspace(ctx context.Context, dir string) (*workspace.Workspace, error) {
	return nil, nil
}

func (s *mockStore) Exec(ctx context.Context, handle string, opts workspace.ExecOptions) ([]workspace.ExecResult, error) {
	return nil, nil
}

func (s *mockStore) AddRepository(ctx context.Context, handle string, repo workspace.RepositoryOption) error {
	return nil
}

func (s *mockStore) AddRepositories(ctx context.Context, handle string, repos []workspace.RepositoryOption) error {
	return nil
}

func (s *mockStore) RemoveRepository(ctx context.Context, handle string, repoName string) error {
	return nil
}

func (s *mockStore) GetGit() git.Git {
	return s.mockGit
}

func NewScenario(t *testing.T, gitOpts []GitOption, storeOpts []StoreOption) *Scenario {
	mockGit := &git.MockGit{}
	for _, opt := range gitOpts {
		opt(mockGit)
	}

	store := &mockStore{
		mockGit:    mockGit,
		workspaces: []*workspace.Workspace{},
	}
	for _, opt := range storeOpts {
		opt(store)
	}

	return &Scenario{
		t:       t,
		mockGit: mockGit,
		steps:   make([]Step, 0),
		store:   store,
	}
}

func (s *Scenario) Key(key, description string) *Scenario {
	s.steps = append(s.steps, Step{
		Type:        StepKey,
		Key:         key,
		Description: description,
	})
	return s
}

func (s *Scenario) Type(text, description string) *Scenario {
	s.steps = append(s.steps, Step{
		Type:        StepInput,
		Text:        text,
		Description: description,
	})
	return s
}

func (s *Scenario) Enter(description string) *Scenario {
	s.steps = append(s.steps, Step{
		Type:        StepEnter,
		Description: description,
	})
	return s
}

func (s *Scenario) Record() tui.StackSnapshot {
	s.t.Helper()

	s.ctx = context.Background()
	stackModel := tui.NewStackModel(s.ctx, s.store)
	s.harness = NewSyncTestHarness(s.t, stackModel, s.ctx)

	// Initialize and set window size
	if initCmd := s.harness.model.Init(); initCmd != nil {
		s.harness.pendingCmds = append(s.harness.pendingCmds, initCmd)
		s.harness.Pump()
	}
	s.harness.Send(tea.WindowSizeMsg{Width: testTermWidth, Height: testTermHeight})

	for _, step := range s.steps {
		s.executeStep(step)
	}

	// Pump any remaining commands
	s.harness.Pump()

	return s.harness.model.Snapshot()
}

func (s *Scenario) WaitForIdle() {
	// No-op: commands are processed synchronously
}

func (s *Scenario) executeStep(step Step) {
	switch step.Type {
	case StepKey:
		s.sendKey(step.Key)
	case StepInput:
		s.sendText(step.Text)
	case StepEnter:
		s.sendEnter()
	}
}

func (s *Scenario) sendKey(key string) {
	switch key {
	case "esc":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyEsc})
	case "enter":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyEnter})
	case "up":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyUp})
	case "down":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyDown})
	case "tab":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyTab})
	case "ctrl+c":
		s.harness.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	default:
		s.harness.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(key),
		})
	}
}

func (s *Scenario) sendText(text string) {
	for _, r := range text {
		s.harness.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{r},
		})
	}
}

func (s *Scenario) sendEnter() {
	s.harness.Send(tea.KeyMsg{Type: tea.KeyEnter})
}

func Match(t *testing.T, name string, snapshot interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal snapshot: %v", err)
	}

	baseName := t.Name()
	if idx := strings.LastIndex(baseName, "/"); idx != -1 {
		baseName = baseName[:idx]
	}

	snaps.WithConfig(snaps.Filename(baseName)).MatchSnapshot(t, name, string(jsonData))
}
