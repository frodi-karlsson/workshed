package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type execResultModel struct {
	results    []workspace.ExecResult
	command    string
	currentIdx int
	showFull   bool
	quitting   bool
	success    bool
	duration   time.Duration
	truncated  bool
}

func (m execResultModel) Init() tea.Cmd {
	return nil
}

func (m execResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "enter":
			m.quitting = true
			return m, tea.Quit
		case "f":
			m.showFull = !m.showFull
			return m, nil
		case "n", "j":
			if m.currentIdx < len(m.results)-1 {
				m.currentIdx++
			}
			return m, nil
		case "p", "k":
			if m.currentIdx > 0 {
				m.currentIdx--
			}
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m execResultModel) View() string {
	borderColor := colorSuccess
	if !m.success {
		borderColor = colorError
	}

	frameStyle := modalFrame().BorderForeground(borderColor)

	titleText := "Command Execution Results"
	if len(m.results) > 1 {
		repoName := m.results[m.currentIdx].Repository
		titleText = fmt.Sprintf("Command Execution Results (%d/%d: %s)", m.currentIdx+1, len(m.results), repoName)
	}

	statusText := "Success"
	statusColor := colorSuccess
	if !m.success {
		statusText = "Failed"
		statusColor = colorError
	}

	status := lipgloss.NewStyle().
		Foreground(statusColor).
		Render("[" + statusText + "]")

	commandLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Command:")

	commandValue := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(m.command)

	repoLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Repository:")

	repoName := m.results[m.currentIdx].Repository
	if repoName == "root" {
		repoName = "(workspace root)"
	}
	repoValue := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(repoName)

	exitLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Exit Code:")

	exitCode := m.results[m.currentIdx].ExitCode
	exitColor := colorSuccess
	if exitCode != 0 {
		exitColor = colorError
	}
	exitValue := lipgloss.NewStyle().
		Foreground(exitColor).
		Render(fmt.Sprintf("%d", exitCode))

	durationLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Duration:")

	durationValue := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(formatDuration(m.duration))

	outputLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Output:")

	truncatedStr, isTruncated := truncateOutput(string(m.results[m.currentIdx].Output), 30)
	m.truncated = isTruncated

	outputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(60).
		Render(truncatedStr)

	var toggleHint string
	if m.truncated {
		toggleHint = lipgloss.NewStyle().
			Foreground(colorWarning).
			Render("[f] Show full output")
	} else {
		toggleHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("[f] Toggle output")
	}

	var navHint string
	if len(m.results) > 1 {
		navHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("[n/p or j/k] Next/Previous repository  ")
	}

	helpHint := lipgloss.NewStyle().
		Foreground(colorVeryMuted).
		MarginTop(1).
		Render(navHint + "[Enter/Esc/q] Close")

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Render(titleText),
			"",
			lipgloss.JoinHorizontal(lipgloss.Left, status, "  ", commandLabel, " ", commandValue),
			lipgloss.JoinHorizontal(lipgloss.Left, repoLabel, " ", repoValue, "  ", exitLabel, " ", exitValue, "  ", durationLabel, " ", durationValue),
			"",
			outputLabel,
			"",
			outputBox,
			"",
			lipgloss.JoinHorizontal(lipgloss.Left, toggleHint, "  ", helpHint),
		),
	)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}

func truncateOutput(output string, maxLines int) (string, bool) {
	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output, false
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (output truncated, press [f] to view full output)", true
}

func ShowExecResultModal(results []workspace.ExecResult, command string) error {
	if len(results) == 0 {
		return nil
	}

	var totalDuration time.Duration
	allSuccess := true
	for _, result := range results {
		totalDuration += result.Duration
		if result.ExitCode != 0 {
			allSuccess = false
		}
	}

	m := execResultModel{
		results:    results,
		command:    command,
		currentIdx: 0,
		showFull:   false,
		quitting:   false,
		success:    allSuccess,
		duration:   totalDuration,
		truncated:  false,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running exec result modal: %w", err)
	}

	return nil
}
