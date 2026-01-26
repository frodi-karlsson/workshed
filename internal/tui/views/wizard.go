package views

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

var ErrNoRepositories = errors.New("at least one repository is required")

type GitProvider interface {
	GetGit() git.Git
}

type currentRepoResultMsg struct {
	url string
	err error
}

type workspaceCreateResultMsg struct {
	ws  *workspace.Workspace
	err error
}

type WizardView struct {
	store        store.Store
	ctx          context.Context
	git          git.Git
	step         int
	purpose      string
	template     string
	templateVars map[string]string
	input        textinput.Model
	repoInput    components.PathCompleter
	repos        []workspace.RepositoryOption
	done         bool
	loadingType  string
	finishMode   bool
	size         measure.Window
}

func NewWizardView(ctx context.Context, s store.Store, g ...git.Git) WizardView {
	ti := textinput.New()
	ti.Placeholder = "What is this workspace for?"
	ti.Prompt = "> "
	ti.Focus()

	repoInput := components.NewPathCompleter()
	repoInput.SetPlaceholder("github.com/user/repo, user/repo@branch, or ./path")
	repoInput.SetPrompt("> ")

	var gitClient git.Git
	if len(g) > 0 && g[0] != nil {
		gitClient = g[0]
	} else if gp, ok := s.(GitProvider); ok {
		gitClient = gp.GetGit()
	}

	if gitClient == nil {
		gitClient = git.RealGit{}
	}

	return WizardView{
		store:        s,
		ctx:          ctx,
		git:          gitClient,
		step:         0,
		input:        ti,
		repoInput:    repoInput,
		templateVars: make(map[string]string),
		finishMode:   false,
	}
}

func (v *WizardView) Init() tea.Cmd {
	return nil
}

func (v *WizardView) SetSize(size measure.Window) {
	v.size = size
	v.repoInput.SetWidth(size.ContentWidth())
}

func (v *WizardView) OnPush()   {}
func (v *WizardView) OnResume() {}
func (v *WizardView) IsLoading() bool {
	return v.loadingType != ""
}
func (v *WizardView) Cancel() {
	v.loadingType = ""
}

func parseRepoFlag(repo string) (url, ref string) {
	if strings.HasPrefix(repo, "git@") {
		colonIdx := strings.Index(repo, ":")
		if colonIdx != -1 {
			atIdx := strings.LastIndex(repo[colonIdx:], "@")
			if atIdx != -1 {
				actualIdx := colonIdx + atIdx
				url = repo[:actualIdx]
				ref = repo[actualIdx+1:]
				return url, ref
			}
		}
		return repo, ""
	}

	atIdx := strings.LastIndex(repo, "@")
	if atIdx != -1 {
		url = repo[:atIdx]
		ref = repo[atIdx+1:]
	} else {
		url = repo
	}

	return url, ref
}

func (v *WizardView) detectCurrentRepoCmd() tea.Cmd {
	return func() tea.Msg {
		url, err := v.git.GetRemoteURL(v.ctx, ".")
		if err != nil {
			return currentRepoResultMsg{err: err}
		}
		return currentRepoResultMsg{url: url}
	}
}

func createWorkspaceCmd(ctx context.Context, s store.Store, purpose string, template string, templateVars map[string]string, repos []workspace.RepositoryOption) tea.Cmd {
	return func() tea.Msg {
		ws, err := s.Create(ctx, workspace.CreateOptions{
			Purpose:      purpose,
			Template:     template,
			TemplateVars: templateVars,
			Repositories: repos,
		})
		if err != nil {
			return workspaceCreateResultMsg{err: err}
		}
		return workspaceCreateResultMsg{ws: ws}
	}
}

func (v *WizardView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case currentRepoResultMsg:
		if v.loadingType != "git" {
			return ViewResult{}, nil
		}
		v.loadingType = ""

		if msg.err != nil {
			errView := NewErrorView(ErrNoRepositories)
			return ViewResult{NextView: errView}, nil
		}

		v.repos = append(v.repos, workspace.RepositoryOption{URL: msg.url, Ref: ""})
		return ViewResult{}, textinput.Blink

	case workspaceCreateResultMsg:
		if v.loadingType != "create" {
			return ViewResult{}, nil
		}
		v.loadingType = ""

		if msg.err != nil {
			errView := NewErrorView(msg.err)
			return ViewResult{NextView: errView}, nil
		}

		v.done = true
		return ViewResult{}, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if v.loadingType != "" {
				v.loadingType = ""
				return ViewResult{}, nil
			}
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyRight:
			if v.step == 1 && len(v.repos) > 0 && !v.finishMode {
				v.finishMode = true
				v.input.Blur()
				return ViewResult{}, textinput.Blink
			}
		case tea.KeyLeft:
			if v.step == 1 && v.finishMode {
				v.finishMode = false
				v.input.Focus()
				return ViewResult{}, textinput.Blink
			}
		case tea.KeyEnter:
			if v.done {
				return ViewResult{Action: StackPop{}}, nil
			}
			if v.step == 0 {
				purpose := v.input.Value()
				if purpose == "" {
					return ViewResult{}, nil
				}
				v.purpose = purpose
				v.step = 1
				v.repoInput.Focus()
				return ViewResult{}, textinput.Blink
			} else {
				if v.finishMode {
					if v.loadingType != "" {
						return ViewResult{}, nil
					}
					if len(v.repos) == 0 {
						v.loadingType = "git"
						return ViewResult{}, v.detectCurrentRepoCmd()
					}
					v.loadingType = "create"
					return ViewResult{}, createWorkspaceCmd(v.ctx, v.store, v.purpose, v.template, v.templateVars, v.repos)
				}

				repoInput := strings.TrimSpace(v.repoInput.Value())
				if repoInput == "" {
					if v.loadingType != "" {
						return ViewResult{}, nil
					}
					if len(v.repos) == 0 {
						v.loadingType = "git"
						return ViewResult{}, v.detectCurrentRepoCmd()
					}
					v.loadingType = "create"
					return ViewResult{}, createWorkspaceCmd(v.ctx, v.store, v.purpose, v.template, v.templateVars, v.repos)
				}

				url, ref := parseRepoFlag(repoInput)
				v.repos = append(v.repos, workspace.RepositoryOption{URL: url, Ref: ref})
				v.repoInput.SetValue("")
				return ViewResult{}, textinput.Blink
			}
		case tea.KeyRunes:
			if v.finishMode && msg.String() == "t" {
				templateView := NewTemplateConfigView(v.ctx, v.store, v.template, v.templateVars)
				return ViewResult{NextView: &templateView}, nil
			}
		}
	}

	if v.step == 0 {
		updatedInput, inputCmd := v.input.Update(msg)
		v.input = updatedInput
		cmd = inputCmd
	} else {
		_, inputCmd := v.repoInput.Update(msg)
		cmd = inputCmd
	}

	return ViewResult{}, cmd
}

func (v *WizardView) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText)

	if v.done {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Workspace created!"), "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Dismiss"),
			),
		)
	}

	if v.step == 0 {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Create Workspace"), "\n", "\n",
				"Purpose:", "\n",
				v.input.View(), "\n", "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Next  [Esc] Cancel"),
			),
		)
	}

	var repoLines []string
	for _, r := range v.repos {
		if r.Ref != "" {
			repoLines = append(repoLines, "  • "+r.URL+"@"+r.Ref)
		} else {
			repoLines = append(repoLines, "  • "+r.URL)
		}
	}

	var reposContent string
	if len(repoLines) > 0 {
		reposContent = lipgloss.JoinVertical(lipgloss.Left, repoLines...)
	} else {
		reposContent = lipgloss.NewStyle().Foreground(components.ColorMuted).Render("  No repositories added yet")
	}

	var helpText string
	if v.finishMode {
		helpText = "[Enter] Create workspace  [←] Add more  [t] Template  [Esc] Cancel"
	} else if len(v.repos) > 0 {
		helpText = "[Enter] Add repo  [→] Finish  [Tab] Complete path  [Esc] Cancel"
	} else {
		helpText = "[Enter] Add repo  [Tab] Complete path  [Esc] Cancel"
	}

	var templateInfo string
	if v.template != "" {
		templateInfo = "Template: " + v.template + "\n"
		if len(v.templateVars) > 0 {
			templateInfo += "Variables: " + fmt.Sprintf("%d configured", len(v.templateVars)) + "\n"
		}
		templateInfo += "\n"
	}

	repoInputView := v.repoInput.View()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Add Repositories"), "\n", "\n",
		"Purpose: "+v.purpose, "\n",
		templateInfo,
		"Repositories:", "\n",
		reposContent, "\n",
		repoInputView, "\n",
		lipgloss.NewStyle().Foreground(components.ColorMuted).Render("  e.g. github.com/user/repo, github.com/user/repo@branch, ./repo, ~/repo"), "\n",
		lipgloss.NewStyle().Foreground(components.ColorMuted).Render("  (default branch used if @branch omitted; current repo used if empty)"), "\n", "\n",
		lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render(helpText),
	)

	return ModalFrame(v.size).Render(content)
}

type WizardViewSnapshot struct {
	Type             string
	Step             int
	Purpose          string
	Template         string
	TemplateVarCount int
	RepoCount        int
	FinishMode       bool
	Loading          bool
	Done             bool
}

func (v *WizardView) Snapshot() interface{} {
	return WizardViewSnapshot{
		Type:             "WizardView",
		Step:             v.step,
		Purpose:          v.purpose,
		Template:         v.template,
		TemplateVarCount: len(v.templateVars),
		RepoCount:        len(v.repos),
		FinishMode:       v.finishMode,
		Loading:          v.loadingType != "",
		Done:             v.done,
	}
}
