package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/key"
)

type modal_ErrorView struct {
	err error
}

func NewErrorView(err error) *modal_ErrorView {
	return &modal_ErrorView{err: err}
}

func (v *modal_ErrorView) Init() tea.Cmd { return nil }

func (v *modal_ErrorView) OnPush()   {}
func (v *modal_ErrorView) OnResume() {}
func (v *modal_ErrorView) IsLoading() bool {
	return false
}

func (v *modal_ErrorView) Cancel() {}

func (v *modal_ErrorView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if key.IsCancel(msg) || key.IsEnter(msg) {
		return ViewResult{Action: StackPop{}}, nil
	}
	return ViewResult{}, nil
}

func (v *modal_ErrorView) View() string {
	return ErrorView(v.err)
}

type ErrorViewSnapshot struct {
	Type  string
	Error string
}

func (v *modal_ErrorView) Snapshot() interface{} {
	return ErrorViewSnapshot{
		Type:  "ErrorView",
		Error: v.err.Error(),
	}
}
