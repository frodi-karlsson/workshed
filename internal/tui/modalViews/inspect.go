package modalViews

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type InspectModal struct {
	workspace *workspace.Workspace
	onDismiss func()
	dismissed bool
}

func NewInspectModal(ws *workspace.Workspace, onDismiss func()) InspectModal {
	return InspectModal{
		workspace: ws,
		onDismiss: onDismiss,
		dismissed: false,
	}
}

func (m InspectModal) Update(msg tea.Msg) (InspectModal, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
			m.dismissed = true
			if m.onDismiss != nil {
				m.onDismiss()
			}
			return m, true
		case tea.KeyRunes:
			if msg.String() == "q" {
				m.dismissed = true
				if m.onDismiss != nil {
					m.onDismiss()
				}
				return m, true
			}
		}
	}
	return m, m.dismissed
}

func (m InspectModal) View() string {
	content := m.buildContent()

	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"\n",
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				Render("[Esc/q/Enter] Dismiss"),
		),
	)
}

func (m InspectModal) buildContent() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText)

	purposeStyle := lipgloss.NewStyle().
		Foreground(colorSuccess)

	var repoLines []string
	for _, repo := range m.workspace.Repositories {
		if repo.Ref != "" {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s @ %s", repo.Name, repo.URL, repo.Ref))
		} else {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s", repo.Name, repo.URL))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Workspace: ")+m.workspace.Handle+"\n",
		purposeStyle.Render("Purpose: ")+m.workspace.Purpose+"\n",
		headerStyle.Render("Path: ")+m.workspace.Path+"\n",
		headerStyle.Render("Created: ")+m.workspace.CreatedAt.Format("Jan 2, 2006")+"\n",
		"\n",
		headerStyle.Render("Repositories:")+"\n",
		lipgloss.JoinVertical(lipgloss.Left, repoLines...),
	)
}

func (m InspectModal) Dismissed() bool {
	return m.dismissed
}
