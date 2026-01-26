package views

import (
	"context"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_PathView struct {
	store        workspace.Store
	ctx          context.Context
	handle       string
	workspace    *workspace.Workspace
	copied       bool
	clipboardErr error
	size         measure.Window
}

func NewPathView(s workspace.Store, ctx context.Context, handle string) *modal_PathView {
	ws, _ := s.Get(ctx, handle)
	return &modal_PathView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
	}
}

func (v *modal_PathView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_PathView) Init() tea.Cmd {
	if v.workspace != nil {
		err := clipboard.WriteAll(v.workspace.Path)
		v.clipboardErr = err
		v.copied = err == nil
	}
	return nil
}

func (v *modal_PathView) OnPush()         {}
func (v *modal_PathView) OnResume()       {}
func (v *modal_PathView) IsLoading() bool { return false }
func (v *modal_PathView) Cancel()         {}

func (v *modal_PathView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss}},
		GetDismissKeyBindings(v.dismiss, "Dismiss")...,
	)
}

func (v *modal_PathView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *modal_PathView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *modal_PathView) View() string {
	if v.workspace == nil {
		return ModalFrame(v.size).Render("Loading...")
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)

	var statusLines []string
	if v.copied {
		statusLines = append(statusLines,
			lipgloss.NewStyle().Foreground(components.ColorSuccess).Render("Copied to clipboard!"),
		)
	} else {
		statusLines = append(statusLines,
			lipgloss.NewStyle().Foreground(components.ColorError).Render("Unable to copy to clipboard"),
		)
		if v.clipboardErr != nil {
			statusLines = append(statusLines,
				lipgloss.NewStyle().Foreground(components.ColorMuted).Render("Error: "+v.clipboardErr.Error()),
			)
		}
	}

	content := []string{
		headerStyle.Render("Path:"), "\n",
		v.workspace.Path, "\n", "\n",
	}
	content = append(content, statusLines...)
	helpText := GenerateHelp(v.KeyBindings())
	helpHint := lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render(helpText)
	content = append(content, "\n", "\n", helpHint)

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

type PathViewSnapshot struct {
	Type   string
	Handle string
	Path   string
}

func (v *modal_PathView) Snapshot() interface{} {
	return PathViewSnapshot{
		Type:   "PathView",
		Handle: v.handle,
		Path:   v.workspace.Path,
	}
}
