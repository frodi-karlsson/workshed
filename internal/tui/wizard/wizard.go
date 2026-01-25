package wizard

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/workspace"
)

type Wizard struct {
	store        StoreList
	ctx          context.Context
	currentStep  wizardStep
	purposeStep  *PurposeStep
	repoStep     *RepoStep
	purpose      string
	repositories []workspace.RepositoryOption
	done         bool
	cancelled    bool
	err          error
}

func NewWizard(ctx context.Context, s StoreList) (*Wizard, error) {
	workspaces, err := s.List(ctx, workspace.ListOptions{})
	if err != nil {
		return nil, err
	}

	purposeStep := NewPurposeStep(workspaces)

	return &Wizard{
		store:       s,
		ctx:         ctx,
		currentStep: stepPurpose,
		purposeStep: &purposeStep,
	}, nil
}

func (m *Wizard) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, func() tea.Msg {
			return WizardDoneMsg{Err: m.err}
		}
	}

	switch msg := msg.(type) {
	case WizardDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, func() tea.Msg {
				return WizardDoneMsg{Cancelled: true}
			}
		case tea.KeyRunes:
			if msg.String() == "q" {
				m.cancelled = true
				return m, func() tea.Msg {
					return WizardDoneMsg{Cancelled: true}
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

func (m *Wizard) updatePurposeStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.purposeStep == nil {
		return m, nil
	}

	updatedPurposeStep, cmd := m.purposeStep.Update(msg)
	*m.purposeStep = updatedPurposeStep

	if m.purposeStep.done {
		m.purpose = m.purposeStep.resultValue

		workspaces, err := m.store.List(m.ctx, workspace.ListOptions{})
		if err != nil {
			m.err = err
			return m, func() tea.Msg {
				return WizardDoneMsg{Err: m.err}
			}
		}

		repoStep := NewRepoStep(workspaces, m.purpose)
		m.repoStep = &repoStep
		m.currentStep = stepRepositories
		return m, textinput.Blink
	}

	if m.purposeStep.cancelled {
		m.cancelled = true
		return m, func() tea.Msg {
			return WizardDoneMsg{Cancelled: true}
		}
	}

	return m, cmd
}

func (m *Wizard) updateRepoStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.repoStep == nil {
		return m, nil
	}

	updatedRepoStep, cmd := m.repoStep.Update(msg)
	*m.repoStep = updatedRepoStep

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc && !m.repoStep.adding {
			m.currentStep = stepPurpose
			m.purposeStep.done = false
			return m, nil
		}
	}

	if m.repoStep.done {
		m.repositories = m.repoStep.repositories
		m.done = true
		return m, func() tea.Msg {
			return WizardDoneMsg{
				Result: &WizardResult{
					Purpose:      m.purpose,
					Repositories: m.repositories,
				},
			}
		}
	}

	return m, cmd
}

func (m *Wizard) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.currentStep {
	case stepPurpose:
		if m.purposeStep != nil {
			return m.purposeStep.View()
		}
	case stepRepositories:
		if m.repoStep != nil {
			return m.repoStep.View()
		}
	}

	return ""
}

func RunCreateWizard(ctx context.Context, s store.Store) (*WizardResult, error) {
	m, err := NewWizard(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("initializing wizard: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running wizard: %w", err)
	}

	if fm, ok := finalModel.(*Wizard); ok {
		if fm.err != nil {
			return nil, fm.err
		}
	}

	m2, ok := finalModel.(*Wizard)
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
