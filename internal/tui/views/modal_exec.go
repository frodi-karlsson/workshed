package views

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_ExecView struct {
	store     workspace.Store
	ctx       context.Context
	handle    string
	workspace *workspace.Workspace
	input     textinput.Model
	result    []workspace.ExecResult
	done      bool
	size      measure.Window
}

func NewExecView(s workspace.Store, ctx context.Context, handle string) *modal_ExecView {
	ws, _ := s.Get(ctx, handle)
	ti := textinput.New()
	ti.Placeholder = "Enter command (e.g., git status)"
	ti.Prompt = "> "
	ti.Focus()
	return &modal_ExecView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
		input:     ti,
	}
}

func (v *modal_ExecView) Init() tea.Cmd { return textinput.Blink }

func (v *modal_ExecView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_ExecView) OnPush()   {}
func (v *modal_ExecView) OnResume() {}
func (v *modal_ExecView) IsLoading() bool {
	return false
}

func (v *modal_ExecView) Cancel() {}

func (v *modal_ExecView) KeyBindings() []KeyBinding {
	if v.done {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss},
			{Key: "esc", Help: "[Esc] Dismiss", Action: v.dismiss},
		}
	}
	return []KeyBinding{
		{Key: "enter", Help: "[Enter] Run", Action: v.run},
		{Key: "esc", Help: "[Esc] Cancel", Action: v.cancel},
		{Key: "ctrl+c", Help: "[Ctrl+C] Cancel", Action: v.cancel},
	}
}

func (v *modal_ExecView) run() (ViewResult, tea.Cmd) {
	results, err := v.store.Exec(v.ctx, v.handle, workspace.ExecOptions{
		Command: strings.Fields(v.input.Value()),
	})
	if err != nil {
		errView := NewErrorView(err)
		return ViewResult{NextView: errView}, nil
	}
	v.result = results
	v.done = true
	return ViewResult{}, nil
}

func (v *modal_ExecView) cancel() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *modal_ExecView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *modal_ExecView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	if !v.done {
		updatedInput, cmd := v.input.Update(msg)
		v.input = updatedInput
		return ViewResult{}, cmd
	}
	return ViewResult{}, nil
}

func (v *modal_ExecView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)

	if v.done {
		var resultLines []string
		for _, r := range v.result {
			resultLines = append(resultLines, "["+r.Repository+"] "+string(r.Output))
		}

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Command Output:"),
			lipgloss.JoinVertical(lipgloss.Left, resultLines...),
		)

		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left, content, "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter/Esc] Dismiss"),
			),
		)
	}

	if v.workspace == nil {
		return ModalFrame(v.size).Render("Loading...")
	}

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Command:"), "\n",
		) + "\n" + v.input.View() + "\n" +
			lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Run  [Esc] Cancel"),
	)
}

type ExecViewSnapshot struct {
	Type      string
	Handle    string
	Command   string
	Done      bool
	RepoCount int
}

func (v *modal_ExecView) Snapshot() interface{} {
	repoCount := 0
	if v.workspace != nil {
		repoCount = len(v.workspace.Repositories)
	}
	return ExecViewSnapshot{
		Type:      "ExecView",
		Handle:    v.handle,
		Command:   v.input.Value(),
		Done:      v.done,
		RepoCount: repoCount,
	}
}
