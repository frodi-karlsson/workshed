package views

import (
	"context"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_RemoveView struct {
	store   workspace.Store
	ctx     context.Context
	handle  string
	done    bool
	confirm bool
	size    measure.Window
}

func NewRemoveView(s workspace.Store, ctx context.Context, handle string) *modal_RemoveView {
	return &modal_RemoveView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		done:   false,
	}
}

func (v *modal_RemoveView) Init() tea.Cmd { return nil }

func (v *modal_RemoveView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_RemoveView) OnPush()   {}
func (v *modal_RemoveView) OnResume() {}
func (v *modal_RemoveView) IsLoading() bool {
	return false
}

func (v *modal_RemoveView) Cancel() {}

func (v *modal_RemoveView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if !v.done {
				if err := v.store.Remove(v.ctx, v.handle); err != nil {
					errView := NewErrorView(err)
					return ViewResult{NextView: errView}, nil
				}
				v.done = true
				v.confirm = true
			}
			return ViewResult{Action: StackPopUntilType{reflect.TypeOf(&DashboardView{})}}, nil
		case "n", "N", "esc", "q", "ctrl+c":
			return ViewResult{Action: StackPop{}}, nil
		case "enter":
			if v.done {
				return ViewResult{Action: StackPopUntilType{reflect.TypeOf(&DashboardView{})}}, nil
			}
		}
	}
	return ViewResult{}, nil
}

func (v *modal_RemoveView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorError)

	if v.done && v.confirm {
		return ModalFrame(v.size).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Workspace removed!"), "\n",
				lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Dismiss"),
			),
		)
	}

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Remove Workspace?"), "\n",
			"Handle: "+v.handle, "\n", "\n",
			lipgloss.NewStyle().Foreground(components.ColorWarning).Render("[y] Yes  [n] No"), "\n",
			lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Esc/q] Cancel"),
		),
	)
}

type RemoveViewSnapshot struct {
	Type    string
	Handle  string
	Done    bool
	Confirm bool
}

func (v *modal_RemoveView) Snapshot() interface{} {
	return RemoveViewSnapshot{
		Type:    "RemoveView",
		Handle:  v.handle,
		Done:    v.done,
		Confirm: v.confirm,
	}
}
