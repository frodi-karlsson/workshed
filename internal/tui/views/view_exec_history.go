package views

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type ExecHistoryView struct {
	store      workspace.Store
	ctx        context.Context
	handle     string
	executions []workspace.ExecutionRecord
	loading    bool
	selected   int
	size       measure.Window
}

func NewExecHistoryView(s workspace.Store, ctx context.Context, handle string) *ExecHistoryView {
	return &ExecHistoryView{
		store:    s,
		ctx:      ctx,
		handle:   handle,
		selected: -1,
	}
}

func (v *ExecHistoryView) Init() tea.Cmd { return nil }

func (v *ExecHistoryView) SetSize(size measure.Window) {
	v.size = size
}

func (v *ExecHistoryView) OnPush()   {}
func (v *ExecHistoryView) OnResume() {}
func (v *ExecHistoryView) IsLoading() bool {
	return v.loading
}

func (v *ExecHistoryView) Cancel() {
	v.loading = false
}

func (v *ExecHistoryView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if key.IsDown(msg) {
		if v.selected < len(v.executions)-1 {
			v.selected++
		}
	} else if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) == 1 {
		if string(km.Runes[0]) == "j" && v.selected < len(v.executions)-1 {
			v.selected++
		}
	}

	if key.IsUp(msg) {
		if v.selected > 0 {
			v.selected--
		}
	} else if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) == 1 {
		if string(km.Runes[0]) == "k" && v.selected > 0 {
			v.selected--
		}
	}

	if v.loading {
		return ViewResult{}, nil
	}

	if len(v.executions) == 0 {
		v.loading = true
		execs, err := v.store.ListExecutions(v.ctx, v.handle, workspace.ListExecutionsOptions{Limit: 20})
		v.loading = false
		if err == nil {
			v.executions = execs
			if len(execs) > 0 {
				v.selected = 0
			}
		}
	}

	return ViewResult{}, nil
}

func (v *ExecHistoryView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	selectedStyle := lipgloss.NewStyle().Foreground(components.ColorHighlight).Bold(true)

	if v.loading && len(v.executions) == 0 {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Execution History"),
				"",
				subStyle.Render("Loading..."),
			),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Execution History"),
		"",
	)

	if len(v.executions) == 0 {
		content += subStyle.Render("No executions recorded")
		content += "\n\n" + dimStyle.Render("[Esc] Dismiss")
		return ModalFrame(v.size).Render(content)
	}

	content += subStyle.Render(string(rune('0'+len(v.executions))) + " recent executions")
	content += "\n\n"

	for i, exec := range v.executions {
		isSelected := i == v.selected
		var lineStyle lipgloss.Style
		if exec.ExitCode != 0 {
			lineStyle = errorStyle
		} else {
			lineStyle = successStyle
		}

		if isSelected {
			lineStyle = selectedStyle
		}

		statusIcon := "✓"
		if exec.ExitCode != 0 {
			statusIcon = "✗"
		}

		timestamp := exec.Timestamp.Format("Jan 02 15:04")
		cmdStr := joinStrings(exec.Command, " ")

		if len(cmdStr) > 30 {
			cmdStr = cmdStr[:30] + "..."
		}

		line := lineStyle.Render(statusIcon+" ") + timestamp
		if isSelected {
			line += " "
		} else {
			line += "  "
		}
		line += cmdStr

		if isSelected {
			line += "\n" + dimStyle.Render("     Exit: "+string(rune('0'+exec.ExitCode))+"  Duration: "+formatDuration(exec.Duration))
		}

		content += line + "\n"
	}

	content += "\n" + dimStyle.Render("[↑↓/j/k] Navigate  [Enter] Details  [Esc] Dismiss")

	return ModalFrame(v.size).Render(content)
}

type ExecHistoryViewSnapshot struct {
	Type       string
	Handle     string
	ExecCount  int
	Selected   int
	FailureCnt int
}

func (v *ExecHistoryView) Snapshot() interface{} {
	failureCnt := 0
	for _, e := range v.executions {
		if e.ExitCode != 0 {
			failureCnt++
		}
	}
	return ExecHistoryViewSnapshot{
		Type:       "ExecHistoryView",
		Handle:     v.handle,
		ExecCount:  len(v.executions),
		Selected:   v.selected,
		FailureCnt: failureCnt,
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return string(rune('0'+ms)) + "ms"
	}
	s := ms / 1000
	if s < 60 {
		return string(rune('0'+s)) + "s"
	}
	m := s / 60
	s = s % 60
	return string(rune('0'+m)) + "m" + string(rune('0'+s)) + "s"
}
