package views

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/key"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_UpdateView struct {
	store  workspace.Store
	ctx    context.Context
	handle string
	input  textinput.Model
	size   measure.Window
}

func NewUpdateView(s workspace.Store, ctx context.Context, handle string) *modal_UpdateView {
	ti := textinput.New()
	ti.Placeholder = "New purpose..."
	ti.Focus()
	return &modal_UpdateView{
		store:  s,
		ctx:    ctx,
		handle: handle,
		input:  ti,
	}
}

func (v *modal_UpdateView) Init() tea.Cmd { return textinput.Blink }

func (v *modal_UpdateView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_UpdateView) OnPush()   {}
func (v *modal_UpdateView) OnResume() {}
func (v *modal_UpdateView) IsLoading() bool {
	return false
}

func (v *modal_UpdateView) Cancel() {}

func (v *modal_UpdateView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}

	if key.IsEnter(msg) {
		if err := v.store.UpdatePurpose(v.ctx, v.handle, v.input.Value()); err != nil {
			errView := NewErrorView(err)
			return ViewResult{NextView: errView}, nil
		}
		return ViewResult{Action: StackPop{}}, nil
	}

	updatedInput, cmd := v.input.Update(msg)
	v.input = updatedInput
	return ViewResult{}, cmd
}

func (v *modal_UpdateView) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Update Purpose"), "\n", "\n",
			v.input.View(), "\n", "\n",
			lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render("[Enter] Save  [Esc] Cancel"),
		),
	)
}

type UpdateViewSnapshot struct {
	Type    string
	Handle  string
	Purpose string
}

func (v *modal_UpdateView) Snapshot() interface{} {
	return UpdateViewSnapshot{
		Type:    "UpdateView",
		Handle:  v.handle,
		Purpose: v.input.Value(),
	}
}
