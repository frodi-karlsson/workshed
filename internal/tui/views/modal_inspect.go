package views

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/workspace"
)

type modal_InspectView struct {
	store     store.Store
	ctx       context.Context
	handle    string
	workspace *workspace.Workspace
}

func NewInspectView(s store.Store, ctx context.Context, handle string) *modal_InspectView {
	ws, _ := s.Get(ctx, handle)
	return &modal_InspectView{
		store:     s,
		ctx:       ctx,
		handle:    handle,
		workspace: ws,
	}
}

func (v *modal_InspectView) Init() tea.Cmd { return nil }

func (v *modal_InspectView) OnPush()   {}
func (v *modal_InspectView) OnResume() {}
func (v *modal_InspectView) IsLoading() bool {
	return false
}

func (v *modal_InspectView) Cancel() {}

func (v *modal_InspectView) Update(msg tea.Msg) (ViewResult, tea.Cmd) {
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

func (v *modal_InspectView) View() string {
	if v.workspace == nil {
		return ModalFrame().Render("Loading...")
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorText)
	purposeStyle := lipgloss.NewStyle().Foreground(ColorSuccess)

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

	return ModalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			content, "\n",
			lipgloss.NewStyle().Foreground(ColorVeryMuted).Render("[Esc/q/Enter] Dismiss"),
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
