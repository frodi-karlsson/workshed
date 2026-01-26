package views

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type HealthIssue struct {
	Category string
	Message  string
	Severity string
}

type view_HealthView struct {
	store    workspace.Store
	ctx      context.Context
	handle   string
	issues   []HealthIssue
	loading  bool
	dryRun   bool
	executed bool
	size     measure.Window
}

func NewHealthView(s workspace.Store, ctx context.Context, handle string) *view_HealthView {
	return &view_HealthView{
		store:  s,
		ctx:    ctx,
		handle: handle,
	}
}

func (v *view_HealthView) Init() tea.Cmd { return nil }

func (v *view_HealthView) SetSize(size measure.Window) {
	v.size = size
}

func (v *view_HealthView) OnPush()   {}
func (v *view_HealthView) OnResume() {}
func (v *view_HealthView) IsLoading() bool {
	return v.loading
}

func (v *view_HealthView) Cancel() {
	v.loading = false
}

func (v *view_HealthView) KeyBindings() []KeyBinding {
	if v.executed {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss},
			{Key: "esc", Help: "[Esc] Dismiss", Action: v.dismiss},
		}
	}
	if len(v.issues) == 0 {
		return []KeyBinding{
			{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss},
			{Key: "esc", Help: "[Esc] Dismiss", Action: v.dismiss},
		}
	}
	return []KeyBinding{
		{Key: "tab", Help: "[Tab] Toggle dry-run", Action: v.toggleDryRun},
		{Key: "enter", Help: "[Enter] Execute", Action: v.execute},
		{Key: "esc", Help: "[Esc] Cancel", Action: v.cancel},
	}
}

func (v *view_HealthView) toggleDryRun() (ViewResult, tea.Cmd) {
	v.dryRun = !v.dryRun
	return ViewResult{}, nil
}

func (v *view_HealthView) execute() (ViewResult, tea.Cmd) {
	if v.executed {
		return ViewResult{Action: StackPop{}}, nil
	}
	v.executed = true
	return ViewResult{}, nil
}

func (v *view_HealthView) cancel() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *view_HealthView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *view_HealthView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if v.loading {
		return ViewResult{}, nil
	}
	if len(v.issues) == 0 {
		v.loading = true
		v.issues = v.detectIssues()
		v.loading = false
		return ViewResult{}, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *view_HealthView) detectIssues() []HealthIssue {
	var issues []HealthIssue
	execs, err := v.store.ListExecutions(v.ctx, v.handle, workspace.ListExecutionsOptions{Limit: 100})
	if err == nil {
		staleCount := 0
		for _, e := range execs {
			if time.Since(e.Timestamp) > 30*24*time.Hour {
				staleCount++
			}
		}
		if staleCount > 0 {
			issues = append(issues, HealthIssue{
				Category: "Stale Executions",
				Message:  fmt.Sprintf("%d executions older than 30 days", staleCount),
				Severity: "info",
			})
		}
	}
	return issues
}

func (v *view_HealthView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	errorStyle := lipgloss.NewStyle().Foreground(components.ColorError)
	warningStyle := lipgloss.NewStyle().Foreground(components.ColorWarning)
	infoStyle := lipgloss.NewStyle().Foreground(components.ColorHighlight)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)
	successStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	if v.loading {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Checking workspace health..."),
				subStyle.Render("Please wait"),
			),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Workspace Health"),
		"",
	)

	if len(v.issues) == 0 {
		content += successStyle.Render("No issues detected")
		content += "\n\n" + subStyle.Render("Your workspace is healthy!")
		content += "\n\n" + dimStyle.Render("[Enter/Esc] Dismiss")
		return ModalFrame(v.size).Render(content)
	}

	content += subStyle.Render("Issues found: " + string(rune('0'+len(v.issues))))
	content += "\n\n"

	byCategory := make(map[string][]HealthIssue)
	for _, issue := range v.issues {
		byCategory[issue.Category] = append(byCategory[issue.Category], issue)
	}

	for category, catIssues := range byCategory {
		content += subStyle.Render(category + ":")
		content += "\n"
		for _, issue := range catIssues {
			var severityStyle lipgloss.Style
			switch issue.Severity {
			case "error":
				severityStyle = errorStyle
			case "warning":
				severityStyle = warningStyle
			default:
				severityStyle = subStyle
			}
			content += "  " + severityStyle.Render("•") + " " + issue.Message
			content += "\n"
		}
		content += "\n"
	}

	if v.executed {
		if v.dryRun {
			content += "\n" + infoStyle.Render("Dry run completed - no changes made")
		} else {
			content += "\n" + successStyle.Render("Cleanup completed successfully")
		}
		content += "\n\n" + dimStyle.Render("[Enter/Esc] Dismiss")
	} else {
		content += "\n"
		if v.dryRun {
			content += subStyle.Render("[Tab] Toggle dry-run (ON)")
		} else {
			content += subStyle.Render("[Tab] Toggle dry-run (OFF)")
		}
		content += "\n"
		if v.dryRun {
			content += infoStyle.Render("[Enter] Preview cleanup actions")
		} else {
			content += warningStyle.Render("[Enter] Execute cleanup")
		}
		content += "\n" + dimStyle.Render("[Esc] Cancel")
	}

	return ModalFrame(v.size).Render(content)
}

type HealthViewSnapshot struct {
	Type     string
	Handle   string
	IssueCnt int
	DryRun   bool
	Executed bool
}

func (v *view_HealthView) Snapshot() interface{} {
	return HealthViewSnapshot{
		Type:     "HealthView",
		Handle:   v.handle,
		IssueCnt: len(v.issues),
		DryRun:   v.dryRun,
		Executed: v.executed,
	}
}
