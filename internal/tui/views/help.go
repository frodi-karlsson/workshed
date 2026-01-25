package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpView struct{}

func NewHelpView() HelpView {
	return HelpView{}
}

func (v *HelpView) Init() tea.Cmd { return nil }

func (v *HelpView) OnPush()   {}
func (v *HelpView) OnResume() {}
func (v *HelpView) IsLoading() bool {
	return false
}

func (v *HelpView) Cancel() {}

func (v *HelpView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v HelpView) View() string {
	return ModalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(ColorText).Render("Help"),
			"\n",
			"[c] Create workspace",
			"[Enter] Open menu for selected workspace",
			"[l] Filter workspaces",
			"[?] Show this help",
			"[q/Esc] Quit",
			"\n",
			"In menu:",
			"[i] Inspect workspace details",
			"[p] Copy path to clipboard",
			"[e] Execute command in repositories",
			"[u] Update workspace purpose",
			"[r] Remove workspace",
			"\n",
			"[Esc/q] Back",
		),
	)
}

type HelpViewSnapshot struct {
	Type string
}

func (v *HelpView) Snapshot() interface{} {
	return HelpViewSnapshot{
		Type: "HelpView",
	}
}
