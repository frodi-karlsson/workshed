package tui

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

var ErrCancelled = errors.New("wizard cancelled")

type wizardDoneMsg struct {
	result    *WizardResult
	cancelled bool
	err       error
}

type wizardStep int

const (
	stepPurpose wizardStep = iota
	stepRepositories
)

type WizardResult struct {
	Purpose      string
	Repositories []workspace.RepositoryOption
}

type wizardModel struct {
	store        store
	ctx          context.Context
	currentStep  wizardStep
	purposeModel *purposeStepModel
	repoModel    *repoStepModel
	purpose      string
	repositories []workspace.RepositoryOption
	done         bool
	cancelled    bool
	err          error
}

func newWizardModel(ctx context.Context, s store) (wizardModel, error) {
	workspaces, err := s.List(ctx, workspace.ListOptions{})
	if err != nil {
		return wizardModel{}, err
	}

	purposeModel := newPurposeStepModel(workspaces)

	return wizardModel{
		store:        s,
		ctx:          ctx,
		currentStep:  stepPurpose,
		purposeModel: &purposeModel,
	}, nil
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, func() tea.Msg {
			return wizardDoneMsg{err: m.err}
		}
	}

	switch msg := msg.(type) {
	case wizardDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, func() tea.Msg {
				return wizardDoneMsg{cancelled: true}
			}
		case tea.KeyRunes:
			if msg.String() == "q" {
				m.cancelled = true
				return m, func() tea.Msg {
					return wizardDoneMsg{cancelled: true}
				}
			}
		}
	}

	switch m.currentStep {
	case stepPurpose:
		return m.updatePurposeStep(msg)
	case stepRepositories:
		return m.updateRepoStep(msg)
	}

	return m, nil
}

func (m wizardModel) updatePurposeStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.purposeModel == nil {
		return m, nil
	}

	updated, cmd := m.purposeModel.Update(msg)
	if updated, ok := updated.(purposeStepModel); ok {
		*m.purposeModel = updated
	}

	if m.purposeModel.done {
		m.purpose = m.purposeModel.resultValue

		workspaces, err := m.store.List(m.ctx, workspace.ListOptions{})
		if err != nil {
			m.err = err
			return m, func() tea.Msg {
				return wizardDoneMsg{err: m.err}
			}
		}

		repoModel := newRepoStepModel(workspaces, m.purpose)
		m.repoModel = &repoModel
		m.currentStep = stepRepositories
		return m, textinput.Blink
	}

	if m.purposeModel.cancelled {
		m.cancelled = true
		return m, func() tea.Msg {
			return wizardDoneMsg{cancelled: true}
		}
	}

	return m, cmd
}

func (m wizardModel) updateRepoStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.repoModel == nil {
		return m, nil
	}

	updated, cmd := m.repoModel.Update(msg)
	if updated, ok := updated.(repoStepModel); ok {
		*m.repoModel = updated
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc && !m.repoModel.adding {
			m.currentStep = stepPurpose
			m.purposeModel.done = false
			return m, nil
		}
	}

	if m.repoModel.done {
		m.repositories = m.repoModel.repositories
		m.done = true
		return m, func() tea.Msg {
			return wizardDoneMsg{
				result: &WizardResult{
					Purpose:      m.purpose,
					Repositories: m.repositories,
				},
			}
		}
	}

	return m, cmd
}

func (m wizardModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(colorError).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.currentStep {
	case stepPurpose:
		if m.purposeModel != nil {
			return m.purposeModel.View()
		}
	case stepRepositories:
		if m.repoModel != nil {
			return m.repoModel.View()
		}
	}

	return ""
}

func RunCreateWizard(ctx context.Context, store *workspace.FSStore) (*WizardResult, error) {
	m, err := newWizardModel(ctx, store)
	if err != nil {
		return nil, fmt.Errorf("initializing wizard: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running wizard: %w", err)
	}

	if fm, ok := finalModel.(wizardModel); ok {
		if fm.err != nil {
			return nil, fm.err
		}
	}

	m2, ok := finalModel.(wizardModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	if m2.err != nil {
		return nil, m2.err
	}
	if m2.cancelled {
		return nil, ErrCancelled
	}
	if m2.done {
		return &WizardResult{
			Purpose:      m2.purpose,
			Repositories: m2.repositories,
		}, nil
	}

	return nil, fmt.Errorf("wizard did not complete")
}
