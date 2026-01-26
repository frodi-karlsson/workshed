package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/components"
	"github.com/frodi/workshed/internal/workspace"
)

func ShowInspectModal(ctx context.Context, s workspace.Store, handle string) error {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	content := buildWorkspaceDetailContent(ws)
	return ShowAlertModal(content)
}

func buildWorkspaceDetailContent(ws *workspace.Workspace) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorText)

	purposeStyle := lipgloss.NewStyle().
		Foreground(components.ColorSuccess)

	var repoLines []string
	for _, repo := range ws.Repositories {
		if repo.Ref != "" {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s @ %s", repo.Name, repo.URL, repo.Ref))
		} else {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s", repo.Name, repo.URL))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Workspace: ")+ws.Handle+"\n",
		purposeStyle.Render("Purpose: ")+ws.Purpose+"\n",
		headerStyle.Render("Path: ")+ws.Path+"\n",
		headerStyle.Render("Created: ")+ws.CreatedAt.Format("Jan 2, 2006")+"\n",
		"\n",
		headerStyle.Render("Repositories:")+"\n",
		lipgloss.JoinVertical(lipgloss.Left, repoLines...),
	)
}
