package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type AddRepoView struct {
	store         workspace.Store
	ctx           context.Context
	handle        string
	input         textinput.Model
	repos         []workspace.RepositoryOption
	err           error
	stale         bool
	cancelled     bool
	size          measure.Window
	invocationCtx workspace.InvocationContext
}

func NewAddRepoView(s workspace.Store, ctx context.Context, handle string, invocationCtx workspace.InvocationContext) AddRepoView {
	ti := textinput.New()
	ti.Placeholder = "Repository URL (e.g., https://github.com/org/repo@branch)"
	ti.CharLimit = 500
	ti.Prompt = ""
	ti.Focus()

	return AddRepoView{
		store:         s,
		ctx:           ctx,
		handle:        handle,
		input:         ti,
		repos:         []workspace.RepositoryOption{},
		invocationCtx: invocationCtx,
	}
}

func (v *AddRepoView) Init() tea.Cmd {
	return textinput.Blink
}

func (v *AddRepoView) SetSize(size measure.Window) {
	v.size = size
}

func (v *AddRepoView) OnPush() {}

func (v *AddRepoView) OnResume() {
	_, err := v.store.Get(v.ctx, v.handle)
	if err != nil {
		v.stale = true
	}
}

func (v *AddRepoView) IsLoading() bool {
	return false
}

func (v *AddRepoView) Cancel() {
	v.cancelled = true
}

func (v *AddRepoView) KeyBindings() []KeyBinding {
	if v.err != nil {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Retry", Action: v.retry},
			{Key: "r", Help: "[r] Retry", Action: v.retry},
			{Key: "esc", Help: "[Esc] Cancel", Action: v.cancel},
		}
	}
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Add", Action: v.addRepo},
		{Key: "esc", Help: "[Esc] Done", Action: v.done},
	}
}

func (v *AddRepoView) retry() (ViewResult, tea.Cmd) {
	v.err = nil
	v.input.Reset()
	return ViewResult{}, textinput.Blink
}

func (v *AddRepoView) addRepo() (ViewResult, tea.Cmd) {
	url := strings.TrimSpace(v.input.Value())
	if url == "" {
		if len(v.repos) > 0 {
			return v.confirmAndAdd()
		}
		return ViewResult{}, nil
	}
	url, ref := workspace.ParseRepoFlag(url)
	v.repos = append(v.repos, workspace.RepositoryOption{
		URL: url,
		Ref: ref,
	})
	v.input.Reset()
	return ViewResult{}, textinput.Blink
}

func (v *AddRepoView) confirmAndAdd() (ViewResult, tea.Cmd) {
	if len(v.repos) == 0 {
		return ViewResult{Action: StackPop{}}, nil
	}
	err := v.store.AddRepositories(v.ctx, v.handle, v.repos, v.invocationCtx.GetInvocationCWD())
	if err != nil {
		v.err = err
		return ViewResult{}, nil
	}
	return ViewResult{Action: StackPopCount{Count: 2}}, nil
}

func (v *AddRepoView) done() (ViewResult, tea.Cmd) {
	if len(v.repos) > 0 {
		return v.confirmAndAdd()
	}
	return ViewResult{Action: StackPop{}}, nil
}

func (v *AddRepoView) cancel() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *AddRepoView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.stale || v.cancelled {
		return ViewResult{Action: StackPop{}}, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	if v.err == nil {
		updatedInput, cmd := v.input.Update(msg)
		v.input = updatedInput
		return ViewResult{}, cmd
	}
	return ViewResult{}, nil
}

func (v *AddRepoView) View() string {
	if v.err != nil {
		return ErrorView(v.err, v.size)
	}

	var repoLines []string
	for i, repo := range v.repos {
		if repo.Ref != "" {
			repoLines = append(repoLines, lipgloss.NewStyle().
				Foreground(components.ColorSuccess).
				Render(fmt.Sprintf("  %d. %s @ %s", i+1, repo.URL, repo.Ref)))
		} else {
			repoLines = append(repoLines, lipgloss.NewStyle().
				Foreground(components.ColorSuccess).
				Render(fmt.Sprintf("  %d. %s", i+1, repo.URL)))
		}
	}

	helpText := "[Enter] Add  [Esc] Done/Cancel"
	if len(v.repos) > 0 {
		helpText = "[Enter] Add more  [Esc] Done"
	}

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(components.ColorText).
				Render("Add Repository to \""+v.handle+"\""),
			"\n",
			v.input.View(),
			"\n",
			lipgloss.NewStyle().
				Foreground(components.ColorVeryMuted).
				Render("Enter repository URL with optional @ref (branch or tag)"),
			"\n",
		) + "\n" +
			lipgloss.NewStyle().
				Foreground(components.ColorText).
				Render("Pending repositories:") + "\n" +
			lipgloss.JoinVertical(lipgloss.Left, repoLines...) + "\n\n" +
			lipgloss.NewStyle().
				Foreground(components.ColorMuted).
				Render(helpText),
	)
}

func (v *AddRepoView) Snapshot() interface{} {
	return AddRepoViewSnapshot{
		Type:        "AddRepoView",
		Handle:      v.handle,
		InputValue:  v.input.Value(),
		RepoCount:   len(v.repos),
		HasError:    v.err != nil,
		ErrorString: errorToString(v.err),
	}
}

type AddRepoViewSnapshot struct {
	Type        string
	Handle      string
	InputValue  string
	RepoCount   int
	HasError    bool
	ErrorString string
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
