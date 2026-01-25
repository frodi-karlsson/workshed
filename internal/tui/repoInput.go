package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type recentRepoItem struct {
	url  string
	ref  string
	name string
}

func (r recentRepoItem) Title() string {
	if r.ref != "" {
		return fmt.Sprintf("%s @ %s", r.name, r.ref)
	}
	return r.name
}

func (r recentRepoItem) Description() string {
	return r.url
}

func (r recentRepoItem) FilterValue() string {
	return r.url + " " + r.name
}

type repoStepModel struct {
	purpose      string
	urlInput     textinput.Model
	refInput     textinput.Model
	recentList   list.Model
	mode         inputMode
	focusedField int
	repositories []workspace.RepositoryOption
	recentRepos  []recentRepoItem
	adding       bool
	done         bool
}

func newRepoStepModel(workspaces []*workspace.Workspace, purpose string) repoStepModel {
	urlInput := textinput.New()
	urlInput.Placeholder = "git@github.com:org/repo, https://..., or local path (., /usr/repo)"
	urlInput.CharLimit = 200
	urlInput.Prompt = ""

	refInput := textinput.New()
	refInput.Placeholder = "main (optional)"
	refInput.CharLimit = 100
	refInput.Prompt = ""

	recentRepos := extractRecentRepos(workspaces)
	items := make([]list.Item, len(recentRepos))
	for i, repo := range recentRepos {
		items[i] = repo
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	l := list.New(items, delegate, 50, maxListHeight)
	l.SetShowTitle(false)
	applyCommonListStyles(&l)

	return repoStepModel{
		purpose:      purpose,
		urlInput:     urlInput,
		refInput:     refInput,
		recentList:   l,
		mode:         modeTyping,
		focusedField: 0,
		recentRepos:  recentRepos,
		repositories: []workspace.RepositoryOption{},
	}
}

func extractRecentRepos(workspaces []*workspace.Workspace) []recentRepoItem {
	seen := make(map[string]bool)
	var repos []recentRepoItem

	for _, ws := range workspaces {
		for _, repo := range ws.Repositories {
			key := repo.URL
			if repo.Ref != "" {
				key = fmt.Sprintf("%s@%s", repo.URL, repo.Ref)
			}

			if !seen[key] {
				seen[key] = true
				repos = append(repos, recentRepoItem{
					url:  repo.URL,
					ref:  repo.Ref,
					name: repo.Name,
				})
			}
		}
	}

	return repos
}

func (m repoStepModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m repoStepModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.adding {
			return m.updateAdding(msg)
		}

		switch msg.String() {
		case "a":
			m.adding = true
			m.mode = modeTyping
			m.focusedField = 0
			m.urlInput.Focus()
			m.refInput.Blur()
			m.urlInput.SetValue("")
			m.refInput.SetValue("")
			return m, textinput.Blink
		case "d":
			if len(m.repositories) > 0 {
				m.repositories = m.repositories[:len(m.repositories)-1]
			}
			return m, nil
		case "enter":
			m.done = true
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.recentList.SetSize(msg.Width-h, min(msg.Height-v-10, maxListHeight))
	}

	return m, tea.Batch(cmds...)
}

func (m repoStepModel) updateAdding(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.adding = false
			m.urlInput.Blur()
			m.refInput.Blur()
			return m, nil
		case tea.KeyTab:
			if m.mode == modeTyping {
				if m.focusedField == 0 {
					m.focusedField = 1
					m.urlInput.Blur()
					m.refInput.Focus()
				} else if len(m.recentRepos) > 0 {
					m.mode = modeSelecting
					m.refInput.Blur()
				}
			} else {
				m.mode = modeTyping
				m.focusedField = 0
				m.urlInput.Focus()
			}
			return m, nil
		case tea.KeyEnter:
			if m.mode == modeTyping {
				url := strings.TrimSpace(m.urlInput.Value())
				if url == "" {
					return m, nil
				}
				ref := strings.TrimSpace(m.refInput.Value())

				m.repositories = append(m.repositories, workspace.RepositoryOption{
					URL: url,
					Ref: ref,
				})

				m.adding = false
				m.urlInput.Blur()
				m.refInput.Blur()
				return m, nil
			} else {
				if len(m.recentList.Items()) > 0 {
					selected := m.recentList.SelectedItem()
					if selected != nil {
						if item, ok := selected.(recentRepoItem); ok {
							m.repositories = append(m.repositories, workspace.RepositoryOption{
								URL: item.url,
								Ref: item.ref,
							})

							m.adding = false
							return m, nil
						}
					}
				}
			}
		case tea.KeyUp, tea.KeyDown:
			if len(m.recentRepos) > 0 {
				m.mode = modeSelecting
				m.urlInput.Blur()
				m.refInput.Blur()
			}
		case tea.KeyRunes:
			switch msg.String() {
			case "q":
				m.adding = false
				m.urlInput.Blur()
				m.refInput.Blur()
				return m, nil
			}
		}

		if m.mode == modeTyping {
			switch msg.String() {
			case "shift+tab":
				if m.focusedField == 1 {
					m.focusedField = 0
					m.urlInput.Focus()
					m.refInput.Blur()
				}
			}

			if m.focusedField == 0 {
				updatedInput, cmd := m.urlInput.Update(msg)
				m.urlInput = updatedInput
				cmds = append(cmds, cmd)
			} else {
				updatedInput, cmd := m.refInput.Update(msg)
				m.refInput = updatedInput
				cmds = append(cmds, cmd)
			}
		}

		if m.mode == modeSelecting {
			updatedList, cmd := m.recentList.Update(msg)
			m.recentList = updatedList
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.recentList.SetSize(msg.Width-h, min(msg.Height-v-10, maxListHeight))
	}

	return m, tea.Batch(cmds...)
}

func (m repoStepModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText)

	borderStyle := modalFrame()

	if m.adding {
		return m.viewAdding(titleStyle, borderStyle)
	}

	repoList := ""
	if len(m.repositories) == 0 {
		repoList = lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("  (none - will use current directory)")
	} else {
		var items []string
		for _, repo := range m.repositories {
			if repo.Ref != "" {
				items = append(items, fmt.Sprintf("  • %s @ %s", repo.URL, repo.Ref))
			} else {
				items = append(items, fmt.Sprintf("  • %s", repo.URL))
			}
		}
		repoList = strings.Join(items, "\n")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1)

	return borderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("Add Repositories"),
			"\n",
			lipgloss.NewStyle().
				Foreground(colorMuted).
				Render("Purpose: \""+m.purpose+"\""),
			"\n",
			lipgloss.NewStyle().
				Foreground(colorText).
				Render("Repositories:"),
			"\n",
			repoList,
			"\n",
			helpStyle.Render("[a] Add repository  [d] Remove last  [Enter] Create workspace"),
			helpStyle.Render("[Esc] Back"),
		),
	)
}

func (m repoStepModel) viewAdding(titleStyle, borderStyle lipgloss.Style) string {
	modeIndicator := "[TYPING]"
	var helpText string

	if m.mode == modeSelecting {
		modeIndicator = "[SELECTING]"
		helpText = "[Enter] Select  [Tab] Type  [↑↓] Navigate  [Esc] Back"
	} else if m.focusedField == 0 {
		if len(m.recentRepos) > 0 {
			helpText = "[Enter] Add  [Tab] Next field  [Esc] Back"
		} else {
			helpText = "[Enter] Add  [Tab] Next field  [Esc] Back"
		}
	} else {
		if len(m.recentRepos) > 0 {
			helpText = "[Enter] Add  [Tab] Select recent  [Shift+Tab] Prev field  [Esc] Back"
		} else {
			helpText = "[Enter] Add  [Shift+Tab] Prev field  [Esc] Back"
		}
	}

	modeStyle := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1)

	urlFocusIndicator := ""
	refFocusIndicator := ""
	if m.focusedField == 0 {
		urlFocusIndicator = " ←"
	} else {
		refFocusIndicator = " ←"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Add Repository")+"  "+modeStyle.Render(modeIndicator),
		"\n",
		lipgloss.NewStyle().Foreground(colorText).Render("URL: ")+m.urlInput.View()+urlFocusIndicator,
		lipgloss.NewStyle().Foreground(colorText).Render("Ref: ")+m.refInput.View()+refFocusIndicator,
		"\n",
	)

	if len(m.recentRepos) > 0 {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			suggestionHeader.Render("Recent repos:"),
			"\n",
			docStyle.Render(m.recentList.View()),
			"\n",
		)
	}

	content = lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		helpStyle.Render(helpText),
	)

	return borderStyle.Render(content)
}
