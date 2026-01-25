package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type inspectModel struct {
	workspace *workspace.Workspace
	quitting  bool
}

func (m inspectModel) Init() tea.Cmd {
	return nil
}

func (m inspectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyRunes:
			switch msg.String() {
			case "q", " ":
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m inspectModel) View() string {
	ws := m.workspace

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText)

	purposeStyle := lipgloss.NewStyle().
		Foreground(colorSuccess)

	var repoLines []string
	for _, repo := range ws.Repositories {
		if repo.Ref != "" {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s @ %s", repo.Name, repo.URL, repo.Ref))
		} else {
			repoLines = append(repoLines, fmt.Sprintf("  • %s\t%s", repo.Name, repo.URL))
		}
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(colorVeryMuted)

	return modalFrame().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Workspace: ")+ws.Handle+"\n",
			purposeStyle.Render("Purpose: ")+ws.Purpose+"\n",
			headerStyle.Render("Path: ")+ws.Path+"\n",
			headerStyle.Render("Created: ")+ws.CreatedAt.Format("Jan 2, 2006")+"\n",
			"\n",
			headerStyle.Render("Repositories:")+"\n",
			lipgloss.JoinVertical(lipgloss.Left, repoLines...),
			"\n",
			helpStyle.Render("[Press any key to return]"),
		),
	)
}

func ShowInspectModal(ctx context.Context, store *workspace.FSStore, handle string) error {
	ws, err := store.Get(ctx, handle)
	if err != nil {
		return err
	}

	m := inspectModel{workspace: ws}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running inspect modal: %w", err)
	}

	return nil
}
