package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/tui/modalViews"
	"github.com/frodi/workshed/internal/tui/wizard"
)

func RenderDashboard(list list.Model, helpText string, quitting bool) string {
	var content []string

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Workshed Dashboard")
	content = append(content, header)

	workspaceCount := len(list.Items())
	if workspaceCount == 0 {
		content = append(content, "\nNo workspaces found. Press 'c' to create one.")
	} else {
		content = append(content, list.View())
	}

	helpHint := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1).
		Render(helpText)

	content = append(content, helpHint)

	frameStyle := modalFrame()
	if quitting {
		frameStyle = frameStyle.BorderForeground(colorError)
	}
	return frameStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func RenderHelpModal(modal *modalViews.HelpModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderInspectModal(modal *modalViews.InspectModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderPathModal(modal *modalViews.PathModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderExecModal(modal *modalViews.ExecModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderExecResultModal(modal *modalViews.ExecResultModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderUpdateModal(modal *modalViews.UpdateModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderRemoveModal(modal *modalViews.RemoveModal) string {
	if modal == nil {
		return modalFrame().Render("Loading...")
	}
	return modal.View()
}

func RenderWizard(w *wizard.Wizard) string {
	if w == nil {
		return modalFrame().Render("Loading wizard...")
	}
	return w.View()
}

func ViewContextMenu(m *contextMenuView) string {
	if m == nil {
		return modalFrame().Render("Loading...")
	}
	return m.View()
}

func updateDismissableModal(m dashboardModel, isNil bool, currentView viewState, msg tea.Msg, updateFn func(*dashboardModel) bool) (dashboardModel, tea.Cmd) {
	if isNil {
		m.currentView = viewDashboard
		return m, nil
	}

	dismissed := updateFn(&m)
	if dismissed {
		m.currentView = viewDashboard
	}
	return m, nil
}

func RenderFilterInput(textInput textinput.Model, workspaceCount int) string {
	filterHint := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Render("[FILTER MODE] ") +
		textInput.View() +
		" (Enter to apply, Esc to cancel)"

	content := []string{filterHint}

	if workspaceCount == 0 {
		content = append(content, "\nNo workspaces found.")
	} else {
		content = append(content, "\n")
	}

	helpHint := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1).
		Render("[Enter] Apply filter  [Esc] Cancel filter")
	content = append(content, helpHint)

	return modalFrame().Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}
