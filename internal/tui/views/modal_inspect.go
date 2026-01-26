package views

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/tui/measure"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_InspectView struct {
	store     workspace.Store
	ctx       context.Context
	handle    string
	workspace *workspace.Workspace
	size      measure.Window
}

func NewInspectView(s workspace.Store, ctx context.Context, handle string) *modal_InspectView {
	ws, _ := s.Get(ctx, handle)
	return &modal_InspectView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
	}
}

func (v *modal_InspectView) SetSize(size measure.Window) {
	v.size = size
}

func (v *modal_InspectView) Init() tea.Cmd { return nil }

func (v *modal_InspectView) OnPush()         {}
func (v *modal_InspectView) OnResume()       {}
func (v *modal_InspectView) IsLoading() bool { return false }
func (v *modal_InspectView) Cancel()         {}

func (v *modal_InspectView) KeyBindings() []KeyBinding {
	return append(
		[]KeyBinding{{Key: "enter", Help: "[Enter] Dismiss", Action: v.dismiss}},
		GetDismissKeyBindings(v.dismiss, "Dismiss")...,
	)
}

func (v *modal_InspectView) dismiss() (ViewResult, tea.Cmd) {
	return ViewResult{Action: StackPop{}}, nil
}

func (v *modal_InspectView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if result, _, handled := HandleKey(v.KeyBindings(), km); handled {
			return result, nil
		}
	}
	return ViewResult{}, nil
}

func (v *modal_InspectView) View() string {
	if v.workspace == nil {
		return ModalFrame(v.size).Render("Loading...")
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(components.ColorText)
	purposeStyle := lipgloss.NewStyle().Foreground(components.ColorSuccess)

	var repoLines []string
	for _, repo := range v.workspace.Repositories {
		if repo.Ref != "" {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s @ %s", repo.Name, repo.URL, repo.Ref))
		} else {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s", repo.Name, repo.URL))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Workspace: ")+v.workspace.Handle+"\n",
		purposeStyle.Render("Purpose: ")+v.workspace.Purpose+"\n",
		headerStyle.Render("Path: ")+v.workspace.Path+"\n",
		headerStyle.Render("Created: ")+v.workspace.CreatedAt.Format("Jan 2, 2006")+"\n",
		"\n",
		headerStyle.Render("Repositories:"),
		lipgloss.JoinVertical(lipgloss.Left, repoLines...),
	)

	helpText := GenerateHelp(v.KeyBindings())
	helpHint := lipgloss.NewStyle().Foreground(components.ColorVeryMuted).Render(helpText)

	return ModalFrame(v.size).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			content, "\n",
			helpHint,
		),
	)
}

type InspectViewSnapshot struct {
	Type        string
	Handle      string
	Purpose     string
	Path        string
	RepoCount   int
	CreatedDate string
}

func (v *modal_InspectView) Snapshot() interface{} {
	createdDate := ""
	if v.workspace != nil {
		createdDate = v.workspace.CreatedAt.Format("Jan 2, 2006")
	}
	return InspectViewSnapshot{
		Type:        "InspectView",
		Handle:      v.handle,
		Purpose:     v.workspace.Purpose,
		Path:        v.workspace.Path,
		RepoCount:   len(v.workspace.Repositories),
		CreatedDate: createdDate,
	}
}
