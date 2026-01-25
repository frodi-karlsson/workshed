package wizard

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/workspace"
)

var ErrCancelled = errors.New("wizard cancelled")

type WizardDoneMsg struct {
	Result    *WizardResult
	Cancelled bool
	Err       error
}

type WizardResult struct {
	Purpose      string
	Repositories []workspace.RepositoryOption
}

type wizardStep int

const (
	stepPurpose wizardStep = iota
	stepRepositories
)

type Step interface {
	Update(msg tea.Msg) (Step, tea.Cmd)
	View() string
	Init() tea.Cmd
	IsDone() bool
	IsCancelled() bool
	GetResult() interface{}
}

type store interface {
	List(ctx context.Context, opts workspace.ListOptions) ([]*workspace.Workspace, error)
}
