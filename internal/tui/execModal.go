package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frodi/workshed/internal/workspace"
)

type repoItem struct {
	repo     *workspace.Repository
	selected bool
}

func (r repoItem) Title() string {
	if r.selected {
		return "[x] " + r.repo.Name
	}
	return "[ ] " + r.repo.Name
}

func (r repoItem) Description() string {
	if r.repo.Ref != "" {
		return r.repo.URL + " @ " + r.repo.Ref
	}
	return r.repo.URL
}

func (r repoItem) FilterValue() string {
	return r.repo.Name
}

type execModel struct {
	textInput textinput.Model
	list      list.Model
	workspace *workspace.Workspace
	focus     string
	done      bool
	quit      bool
}

const (
	focusInput = "input"
	focusList  = "list"
)

func (m execModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m execModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.focus == focusInput {
				m.focus = focusList
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case tea.KeyTab:
			if m.focus == focusInput {
				m.focus = focusList
			} else {
				m.focus = focusInput
			}
			return m, nil
		}

		if m.focus == focusInput {
			updatedInput, cmd := m.textInput.Update(msg)
			m.textInput = updatedInput
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		if m.focus == focusList {
			switch msg.Type {
			case tea.KeySpace:
				idx := m.list.Index()
				items := m.list.Items()
				if idx >= 0 && idx < len(items) {
					if ri, ok := items[idx].(repoItem); ok {
						newItems := make([]list.Item, len(items))
						copy(newItems, items)
						ri.selected = !ri.selected
						newItems[idx] = ri
						m.list.SetItems(newItems)
					}
				}
				return m, nil
			}
		}
	}

	if m.focus == focusInput {
		updatedInput, cmd := m.textInput.Update(msg)
		m.textInput = updatedInput
		cmds = append(cmds, cmd)
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = updatedList
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m execModel) View() string {
	inputView := m.textInput.View()
	repoView := m.list.View()

	frameStyle := modalFrame()
	if m.done {
		frameStyle = frameStyle.BorderForeground(colorSuccess)
	}

	focusIndicator := ""
	if m.focus == focusInput {
		focusIndicator = "[Input]"
	} else {
		focusIndicator = "[List]"
	}

	return frameStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Render("Command to run in \""+m.workspace.Handle+"\": "+focusIndicator),
			"\n",
			inputView,
			"\n\n",
			lipgloss.NewStyle().
				Foreground(colorText).
				Render("Repositories (space to toggle):"),
			"\n",
			repoView,
			"\n",
			lipgloss.NewStyle().
				Foreground(colorVeryMuted).
				MarginTop(1).
				Render("[Tab] Switch  [Space] Toggle  [Enter] Run  [Esc] Cancel"),
		),
	)
}

type ExecResult struct {
	Command   string
	RepoNames []string
}

func ShowExecModal(ctx context.Context, store *workspace.FSStore, handle string) (*ExecResult, error) {
	ws, err := store.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	repoItems := make([]list.Item, len(ws.Repositories))
	for i, repo := range ws.Repositories {
		repoItems[i] = repoItem{repo: &repo, selected: true}
	}

	ti := textinput.New()
	ti.Placeholder = "e.g., make test, go build, npm run lint..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Prompt = ""

	l := list.New(repoItems, list.NewDefaultDelegate(), 30, maxListHeight)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	applyCommonListStyles(&l)

	m := execModel{
		textInput: ti,
		list:      l,
		workspace: ws,
		focus:     focusInput,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running exec modal: %w", err)
	}

	if fm, ok := finalModel.(execModel); ok {
		if fm.quit {
			return nil, nil
		}
		if fm.done {
			command := strings.TrimSpace(fm.textInput.Value())
			if command == "" {
				return nil, nil
			}

			var repoNames []string
			for _, item := range fm.list.Items() {
				if ri, ok := item.(repoItem); ok && ri.selected {
					repoNames = append(repoNames, ri.repo.Name)
				}
			}

			return &ExecResult{
				Command:   command,
				RepoNames: repoNames,
			}, nil
		}
	}

	return nil, nil
}
