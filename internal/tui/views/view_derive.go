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

type view_DeriveView struct {
	store   workspace.Store
	ctx     context.Context
	handle  string
	context *workspace.WorkspaceContext
	copied  bool
	err     error
	size    measure.Window
}

func NewDeriveView(s workspace.Store, ctx context.Context, handle string) *view_DeriveView {
	return &view_DeriveView{
		store:  s,
		ctx:    ctx,
		handle: handle,
	}
}

func (v *view_DeriveView) Init() tea.Cmd {
	ctx, err := v.store.DeriveContext(v.ctx, v.handle)
	if err == nil {
		v.context = ctx
	} else {
		v.err = err
	}
	return nil
}

func (v *view_DeriveView) SetSize(size measure.Window) {
	v.size = size
}

func (v *view_DeriveView) OnPush()   {}
func (v *view_DeriveView) OnResume() {}
func (v *view_DeriveView) IsLoading() bool {
	return false
}

func (v *view_DeriveView) Cancel() {}

func (v *view_DeriveView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if key.IsEnter(msg) {
		if v.context != nil {
			v.copied = true
		}
		return ViewResult{}, nil
	}

	return ViewResult{}, nil
}

func (v *view_DeriveView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	subStyle := lipgloss.NewStyle().Foreground(components.ColorMuted)
	dimStyle := lipgloss.NewStyle().Foreground(components.ColorVeryMuted)

	if v.context == nil {
		errMsg := "could not derive context"
		if v.err != nil {
			errMsg += ": " + v.err.Error()
		}
		return ModalFrame(v.size).Render("Error: " + errMsg)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Workspace Context"),
		"",
	)

	content += subStyle.Render("Handle: ")
	content += lipgloss.NewStyle().Foreground(components.ColorText).Render(v.context.Handle)
	content += "\n"

	content += subStyle.Render("Purpose: ")
	content += lipgloss.NewStyle().Foreground(components.ColorText).Render(v.context.Purpose)
	content += "\n\n"

	content += subStyle.Render("Repositories (" + string(rune('0'+len(v.context.Repositories))) + "):")
	if len(v.context.Repositories) == 0 {
		content += " " + subStyle.Render("[empty]")
	}
	content += "\n"
	for _, repo := range v.context.Repositories {
		content += "  - " + repo.Name + "\n"
		content += "    " + dimStyle.Render(repo.URL)
		content += "\n"
	}

	content += "\n" + subStyle.Render("Captures ("+string(rune('0'+len(v.context.Captures)))+"):")
	if len(v.context.Captures) == 0 {
		content += " " + subStyle.Render("[none]")
	}
	content += "\n"
	for _, cap := range v.context.Captures {
		content += "  - " + cap.Name + " (" + cap.Kind + ")\n"
	}

	content += "\n" + subStyle.Render("Metadata:")
	content += "\n  Workshed: " + v.context.Metadata.WorkshedVersion
	content += "\n  Executions: " + string(rune('0'+v.context.Metadata.ExecutionsCount))
	content += "\n  Captures: " + string(rune('0'+v.context.Metadata.CapturesCount))
	if v.context.Metadata.LastExecutedAt != nil {
		content += "\n  Last executed: " + v.context.Metadata.LastExecutedAt.Format("2006-01-02 15:04")
	}
	if v.context.Metadata.LastCapturedAt != nil {
		content += "\n  Last captured: " + v.context.Metadata.LastCapturedAt.Format("2006-01-02 15:04")
	}

	if v.copied {
		content += "\n\n" + lipgloss.NewStyle().Foreground(components.ColorSuccess).Render("Context JSON copied to clipboard!")
	}

	content += "\n\n" + dimStyle.Render("[Enter] Copy JSON  [Esc] Close")

	return ModalFrame(v.size).Render(content)
}

type DeriveViewSnapshot struct {
	Type         string
	Handle       string
	Purpose      string
	RepoCount    int
	CaptureCount int
	ExecCount    int
	Version      int
}

func (v *view_DeriveView) Snapshot() interface{} {
	repoCount := 0
	captureCount := 0
	execCount := 0
	purpose := ""
	version := 0
	if v.context != nil {
		repoCount = len(v.context.Repositories)
		captureCount = len(v.context.Captures)
		execCount = v.context.Metadata.ExecutionsCount
		purpose = v.context.Purpose
		version = v.context.Version
	}
	return DeriveViewSnapshot{
		Type:         "DeriveView",
		Handle:       v.handle,
		Purpose:      purpose,
		RepoCount:    repoCount,
		CaptureCount: captureCount,
		ExecCount:    execCount,
		Version:      version,
	}
}
