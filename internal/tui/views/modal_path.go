package views

import (
	"context"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_PathView struct {
	store     store.Store
	ctx       context.Context
	handle    string
	workspace *workspace.Workspace
	copied    bool
}

func NewPathView(s store.Store, ctx context.Context, handle string) *modal_PathView {
	ws, _ := s.Get(ctx, handle)
	return &modal_PathView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
	}
}

func (v *modal_PathView) Init() tea.Cmd {
	if v.workspace != nil {
		_ = clipboard.WriteAll(v.workspace.Path)
		v.copied = true
	}
	return nil
}

func (v *modal_PathView) OnPush()   {}
func (v *modal_PathView) OnResume() {}
func (v *modal_PathView) IsLoading() bool {
	return false
}

func (v *modal_PathView) Cancel() {}

func (v *modal_PathView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
			return ViewResult{Action: StackPop{}}, nil
		case tea.KeyRunes:
			if msg.String() == "q" {
				return ViewResult{Action: StackPop{}}, nil
			}
		}
	}
	return ViewResult{}, nil
}

func (v *modal_PathView) View() string {
	if v.workspace == nil {
		return ModalFrame().Render("Loading...")
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)

	statusMsg := "Copied to clipboard!"
	if !v.copied {
		statusMsg = "Unable to copy to clipboard"
	}

	statusStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess).Render(statusMsg)

	return ModalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Path:"), "\n",
			v.workspace.Path, "\n", "\n",
			statusStyle, "\n", "\n",
			lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Esc/q/Enter] Dismiss"),
		),
	)
}

type PathViewSnapshot struct {
	Type   string
	Handle string
	Path   string
	Copied bool
}

func (v *modal_PathView) Snapshot() interface{} {
	return PathViewSnapshot{
		Type:   "PathView",
		Handle: v.handle,
		Path:   v.workspace.Path,
		Copied: v.copied,
	}
}
