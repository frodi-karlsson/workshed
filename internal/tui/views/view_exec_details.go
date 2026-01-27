package views

import (
	"context"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ExecDetailsView struct {
	store       workspace.Store
	ctx         context.Context
	handle      string
	execID      string
	exec        *workspace.ExecutionRecord
	outputs     map[string]string
	selectedTab int
	vp          components.Scrollable
	size        measure.Window
}

func NewExecDetailsView(s workspace.Store, ctx context.Context, handle, execID string) *ExecDetailsView {
	exec, _ := s.GetExecution(ctx, handle, execID)

	outputs := make(map[string]string)
	if exec != nil {
		wsPath, _ := s.Path(ctx, handle)
		execDir := filepath.Join(wsPath, ".workshed", "executions", execID)

		for _, result := range exec.Results {
			var combinedOutput []byte

			stdoutPath := filepath.Join(execDir, "stdout", result.Repository+".txt")
			if data, err := os.ReadFile(stdoutPath); err == nil {
				combinedOutput = append(combinedOutput, data...)
			}

			stderrPath := filepath.Join(execDir, "stderr", result.Repository+".txt")
			if data, err := os.ReadFile(stderrPath); err == nil {
				if len(combinedOutput) > 0 {
					combinedOutput = append(combinedOutput, '\n')
				}
				combinedOutput = append(combinedOutput, data...)
			}

			if len(combinedOutput) == 0 {
				outputs[result.Repository] = "(no output)"
			} else {
				outputs[result.Repository] = string(combinedOutput)
			}
		}

		if len(exec.Results) == 0 {
			outputs["root"] = "(no repository results)"
		}
	}

	vp := components.NewScrollable(80, 20)

	return &ExecDetailsView{
		store:       s,
		ctx:         ctx,
		handle:      handle,
		execID:      execID,
		exec:        exec,
		outputs:     outputs,
		selectedTab: 0,
		vp:          vp,
	}
}

func (v *ExecDetailsView) Init() tea.Cmd { return nil }

func (v *ExecDetailsView) SetSize(size measure.Window) {
	v.size = size
	v.vp.SetSize(size.ModalWidth()-4, size.ModalHeight()-18)
	v.updateViewportContent()
}

func (v *ExecDetailsView) OnPush()   {}
func (v *ExecDetailsView) OnResume() {}
func (v *ExecDetailsView) IsLoading() bool {
	return false
}

func (v *ExecDetailsView) Cancel() {}

func (v *ExecDetailsView) KeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "tab", Help: "[Tab] Next repo", Action: v.nextTab},
		{Key: "shift+tab", Help: "[Shift+Tab] Prev repo", Action: v.prevTab},
		{Key: "up", Help: "[↑] Scroll up", Action: v.scrollUp},
		{Key: "down", Help: "[↓] Scroll down", Action: v.scrollDown},
		{Key: "q", Help: "[q] Back", Action: v.dismiss},
		{Key: "esc", Help: "[Esc] Back", Action: v.dismiss},
		{Key: "ctrl+c", Help: "[Ctrl+C] Back", Action: v.dismiss},
	}
}

func (v *ExecDetailsView) nextTab() (ViewResult, tea.Cmd) {
	repos := v.repos()
	if len(repos) <= 1 {
		return ViewResult{}, nil
	}
	v.selectedTab = (v.selectedTab + 1) % len(repos)
	v.updateViewportContent()
	return ViewResult{}, nil
}

func (v *ExecDetailsView) prevTab() (ViewResult, tea.Cmd) {
	repos := v.repos()
	if len(repos) <= 1 {
		return ViewResult{}, nil
	}
	v.selectedTab--
	if v.selectedTab < 0 {
		v.selectedTab = len(repos) - 1
	}
	v.updateViewportContent()
	return ViewResult{}, nil
}

func (v *ExecDetailsView) scrollUp() (ViewResult, tea.Cmd) {
	v.vp.LineUp()
	return ViewResult{}, nil
}

func (v *ExecDetailsView) scrollDown() (ViewResult, tea.Cmd) {
	v.vp.LineDown()
	return ViewResult{}, nil
}

func (v *ExecDetailsView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *ExecDetailsView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *ExecDetailsView) repos() []string {
	if v.exec == nil || v.exec.Results == nil {
		return []string{"root"}
	}
	repos := make([]string, 0, len(v.exec.Results))
	for _, r := range v.exec.Results {
		repos = append(repos, r.Repository)
	}
	return repos
}

func (v *ExecDetailsView) updateViewportContent() {
	repos := v.repos()
	if len(repos) == 0 {
		v.vp.SetContent("(no output)")
		return
	}

	if v.selectedTab >= len(repos) {
		v.selectedTab = 0
	}

	repo := repos[v.selectedTab]
	output, ok := v.outputs[repo]
	if !ok {
		output = "(no output)"
	}
	v.vp.SetContent(output)
}

func (v *ExecDetailsView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	if v.exec == nil {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Execution Details"),
				"",
				subStyle.Render("Loading..."),
			),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Execution Details"),
		"",
	)

	timestamp := v.exec.Timestamp.Format("Jan 02 15:04")
	cmdStr := joinStrings(v.exec.Command, " ")
	content += subStyle.Render("Command: ") + cmdStr + "\n"
	content += subStyle.Render("Time: ") + timestamp + "\n"

	exitLabel := "Success"
	exitColor := successStyle
	if v.exec.ExitCode != 0 {
		exitLabel = "Failed"
		exitColor = errorStyle
	}
	content += subStyle.Render("Status: ") + exitColor.Render(exitLabel) + "\n"
	content += subStyle.Render("Duration: ") + formatDuration(v.exec.Duration) + "\n"

	content += "\n"
	repos := v.repos()
	for i, repo := range repos {
		if i == v.selectedTab {
			content += lipgloss.NewStyle().
				Foreground(components.ColorHighlight).
				Bold(true).
				Render("[" + repo + "]")
		} else {
			content += subStyle.Render(" " + repo + " ")
		}
		content += " "
	}
	content += "\n\n"

	v.updateViewportContent()

	content += v.vp.View()
	content += "\n" + dimStyle.Render(GenerateHelp(v.KeyBindings()))

	return ModalFrame(v.size).Render(content)
}

type ExecDetailsViewSnapshot struct {
	Type      string
	Handle    string
	ExecID    string
	Command   string
	ExitCode  int
	RepoCount int
	Selected  int
	HasOutput bool
}

func (v *ExecDetailsView) Snapshot() interface{} {
	repos := v.repos()
	hasOutput := false
	for _, output := range v.outputs {
		if output != "(no output)" && output != "(no repository results)" {
			hasOutput = true
			break
		}
	}
	return ExecDetailsViewSnapshot{
		Type:      "ExecDetailsView",
		Handle:    v.handle,
		ExecID:    v.execID,
		Command:   joinStrings(v.exec.Command, " "),
		ExitCode:  v.exec.ExitCode,
		RepoCount: len(repos),
		Selected:  v.selectedTab,
		HasOutput: hasOutput,
	}
}
