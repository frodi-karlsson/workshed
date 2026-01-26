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

func WithCaptures(captures []workspace.Capture) StoreOption {
	return func(s *mockStore) {
		s.captures = captures
	}
}

func WithExecutions(executions []workspace.ExecutionRecord) StoreOption {
	return func(s *mockStore) {
		s.executions = executions
	}
}

func WithValidationResult(result workspace.AgentsValidationResult) StoreOption {
	return func(s *mockStore) {
		s.validationResult = result
	}
}

func WithPreflightResult(result workspace.ApplyPreflightResult) StoreOption {
	return func(s *mockStore) {
		s.preflightResult = result
	}
}

func WithContext(ctx *workspace.WorkspaceContext) StoreOption {
	return func(s *mockStore) {
		s.context = ctx
	}
}

func WithApplyError(err error) StoreOption {
	return func(s *mockStore) {
		s.applyErr = err
	}
}

func WithCaptureError(err error) StoreOption {
	return func(s *mockStore) {
		s.captureErr = err
	}
}

func WithDeriveError(err error) StoreOption {
	return func(s *mockStore) {
		s.deriveErr = err
	}
}

func WithValidateError(err error) StoreOption {
	return func(s *mockStore) {
		s.validateErr = err
	}
}

func WithDirtyRepo(name string) StoreOption {
	return func(s *mockStore) {
		s.dirtyRepos = append(s.dirtyRepos, name)
	}
}

type mockStore struct {
	mockGit          git.Git
	workspaces       []*workspace.Workspace
	captures         []workspace.Capture
	executions       []workspace.ExecutionRecord
	validationResult workspace.AgentsValidationResult
	preflightResult  workspace.ApplyPreflightResult
	context          *workspace.WorkspaceContext
	createErr        error
	createDelay      time.Duration
	listDelay        time.Duration
	listErr          error
	applyErr         error
	captureErr       error
	deriveErr        error
	validateErr      error
	dirtyRepos       []string
	invocationCWD    string
}

func (s *mockStore) GetInvocationCWD() string {
	return s.invocationCWD
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

func (s *mockStore) AddRepository(ctx context.Context, handle string, repo workspace.RepositoryOption, invocationCWD string) error {
	return nil
}

func (s *mockStore) AddRepositories(ctx context.Context, handle string, repos []workspace.RepositoryOption, invocationCWD string) error {
	return nil
}

func (s *mockStore) RemoveRepository(ctx context.Context, handle string, repoName string) error {
	return nil
}

func (s *mockStore) RecordExecution(ctx context.Context, handle string, record workspace.ExecutionRecord) error {
	return nil
}

func (s *mockStore) GetExecution(ctx context.Context, handle, execID string) (*workspace.ExecutionRecord, error) {
	return nil, nil
}

func (s *mockStore) CaptureState(ctx context.Context, handle string, opts workspace.CaptureOptions) (*workspace.Capture, error) {
	if s.captureErr != nil {
		err := s.captureErr
		s.captureErr = nil
		return nil, err
	}
	capture := &workspace.Capture{
		ID:        "cap-" + handle,
		Handle:    handle,
		Name:      opts.Name,
		Kind:      opts.Kind,
		Timestamp: time.Now(),
		GitState:  []workspace.GitRef{},
		Metadata: workspace.CaptureMetadata{
			Description: opts.Description,
			Tags:        opts.Tags,
			Custom:      opts.Custom,
		},
	}
	for _, repo := range s.workspaces {
		for _, r := range repo.Repositories {
			isDirty := false
			for _, dirty := range s.dirtyRepos {
				if dirty == r.Name {
					isDirty = true
					break
				}
			}
			capture.GitState = append(capture.GitState, workspace.GitRef{
				Repository: r.Name,
				Branch:     r.Ref,
				Commit:     "abc123",
				Dirty:      isDirty,
				Status:     "",
			})
		}
	}
	return capture, nil
}

func (s *mockStore) ApplyCapture(ctx context.Context, handle string, captureID string) error {
	if s.applyErr != nil {
		err := s.applyErr
		s.applyErr = nil
		return err
	}
	return nil
}

func (s *mockStore) PreflightApply(ctx context.Context, handle string, captureID string) (workspace.ApplyPreflightResult, error) {
	return s.preflightResult, nil
}

func (s *mockStore) GetCapture(ctx context.Context, handle, captureID string) (*workspace.Capture, error) {
	for _, c := range s.captures {
		if c.ID == captureID {
			return &c, nil
		}
	}
	return nil, nil
}

func (s *mockStore) ListCaptures(ctx context.Context, handle string) ([]workspace.Capture, error) {
	return s.captures, nil
}

func (s *mockStore) DeriveContext(ctx context.Context, handle string) (*workspace.WorkspaceContext, error) {
	if s.deriveErr != nil {
		err := s.deriveErr
		s.deriveErr = nil
		return nil, err
	}
	if s.context != nil {
		return s.context, nil
	}
	ctxVal := &workspace.WorkspaceContext{
		Version:      1,
		GeneratedAt:  time.Now(),
		Handle:       handle,
		Purpose:      "Test workspace",
		Repositories: []workspace.ContextRepo{},
		Captures:     []workspace.ContextCapture{},
		Metadata: workspace.ContextMetadata{
			WorkshedVersion: "1.0.0",
			ExecutionsCount: len(s.executions),
			CapturesCount:   len(s.captures),
		},
	}
	for _, ws := range s.workspaces {
		if ws.Handle == handle {
			ctxVal.Purpose = ws.Purpose
			for _, r := range ws.Repositories {
				ctxVal.Repositories = append(ctxVal.Repositories, workspace.ContextRepo{
					Name:     r.Name,
					Path:     "/workspaces/" + handle + "/" + r.Name,
					URL:      r.URL,
					RootPath: "/workspaces/" + handle + "/" + r.Name,
				})
			}
			break
		}
	}
	for _, c := range s.captures {
		ctxVal.Captures = append(ctxVal.Captures, workspace.ContextCapture{
			ID:        c.ID,
			Timestamp: c.Timestamp,
			Name:      c.Name,
			Kind:      c.Kind,
			RepoCount: len(c.GitState),
		})
	}
	if len(s.executions) > 0 {
		lastExec := s.executions[len(s.executions)-1]
		ctxVal.Metadata.LastExecutedAt = &lastExec.Timestamp
	}
	if len(s.captures) > 0 {
		lastCap := s.captures[len(s.captures)-1]
		ctxVal.Metadata.LastCapturedAt = &lastCap.Timestamp
	}
	return ctxVal, nil
}

func (s *mockStore) ValidateAgents(ctx context.Context, handle string, agentsPath string) (workspace.AgentsValidationResult, error) {
	if s.validateErr != nil {
		err := s.validateErr
		s.validateErr = nil
		return workspace.AgentsValidationResult{}, err
	}
	return s.validationResult, nil
}

func (s *mockStore) ListExecutions(ctx context.Context, handle string, opts workspace.ListExecutionsOptions) ([]workspace.ExecutionRecord, error) {
	return s.executions, nil
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
		mockGit:       mockGit,
		workspaces:    []*workspace.Workspace{},
		invocationCWD: t.TempDir(),
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
	stackModel := tui.NewStackModel(s.ctx, s.store, s.store)
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
