package components

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PathCompleter struct {
	textInput   textinput.Model
	list        list.Model
	matches     []string
	selected    int
	visible     bool
	dir         string
	prefix      string
	onComplete  func(path string)
	onCancel    func()
	placeholder string
	prompt      string
}

type pathCompleterItem struct {
	name    string
	display string
}

func (i pathCompleterItem) Title() string       { return i.display }
func (i pathCompleterItem) Description() string { return "" }
func (i pathCompleterItem) FilterValue() string { return i.name }

type pathCompleterDelegate struct{}

func (d pathCompleterDelegate) Height() int  { return 1 }
func (d pathCompleterDelegate) Spacing() int { return 0 }

func (d pathCompleterDelegate) Update(msg tea.Msg, l *list.Model) tea.Cmd { return nil }

func (d pathCompleterDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	pi, ok := item.(pathCompleterItem)
	if !ok {
		return
	}

	prefix := "  "
	if index == m.Index() {
		prefix = "▶ "
	}

	//nolint:errcheck
	// Render cannot handle write errors - no recovery mechanism exists
	fmt.Fprint(w, prefix, pi.display)
}

func NewPathCompleter() PathCompleter {
	ti := textinput.New()
	ti.CharLimit = 300
	ti.Prompt = "> "

	l := list.New([]list.Item{}, pathCompleterDelegate{}, 50, 4)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return PathCompleter{
		textInput: ti,
		list:      l,
		selected:  0,
		visible:   false,
	}
}

func (c *PathCompleter) SetPlaceholder(placeholder string) *PathCompleter {
	c.placeholder = placeholder
	c.textInput.Placeholder = placeholder
	return c
}

func (c *PathCompleter) SetPrompt(prompt string) *PathCompleter {
	c.prompt = prompt
	c.textInput.Prompt = prompt
	return c
}

func (c *PathCompleter) SetOnComplete(fn func(path string)) *PathCompleter {
	c.onComplete = fn
	return c
}

func (c *PathCompleter) SetOnCancel(fn func()) *PathCompleter {
	c.onCancel = fn
	return c
}

func (c *PathCompleter) Focus() {
	c.textInput.Focus()
}

func (c *PathCompleter) Blur() {
	c.textInput.Blur()
}

func (c *PathCompleter) Value() string {
	return c.textInput.Value()
}

func (c *PathCompleter) SetValue(value string) {
	c.textInput.SetValue(value)
	c.hideCompletion()
}

func (c *PathCompleter) IsVisible() bool {
	return c.visible
}

func (c *PathCompleter) hideCompletion() {
	c.visible = false
	c.matches = nil
	c.selected = 0
}

func (c *PathCompleter) findMatches() {
	current := c.textInput.Value()
	if current == "" {
		c.hideCompletion()
		return
	}

	dir := "."
	prefix := current

	if strings.Contains(prefix, "/") || strings.HasPrefix(prefix, "./") || strings.HasPrefix(prefix, "../") {
		dir = filepath.Dir(prefix)
		if dir == "." {
			dir = ""
		}
		prefix = filepath.Base(prefix)
	}

	home, _ := os.UserHomeDir()
	if strings.HasPrefix(dir, "~") {
		dir = filepath.Join(home, dir[1:])
	}

	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		c.hideCompletion()
		return
	}

	var items []list.Item
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			display := name
			if entry.IsDir() {
				display = name + "/"
			}
			items = append(items, pathCompleterItem{name: name, display: display})
		}
	}

	if len(items) == 0 {
		c.hideCompletion()
		return
	}

	c.matches = make([]string, len(items))
	for i, item := range items {
		c.matches[i] = item.(pathCompleterItem).name
	}

	c.list.SetItems(items)
	c.dir = dir
	c.prefix = prefix
	c.visible = true
	c.selected = 0
	c.list.Select(0)
}

func (c *PathCompleter) selectCurrent() {
	if len(c.matches) == 0 || !c.visible {
		return
	}

	selected := c.matches[c.selected]
	newValue := filepath.Join(c.dir, selected)
	if !strings.HasPrefix(c.textInput.Value(), "./") &&
		!strings.HasPrefix(c.textInput.Value(), "../") &&
		c.dir == "." {
		newValue = selected
	}

	c.textInput.SetValue(newValue)
	c.textInput.CursorEnd()
	c.hideCompletion()

	if c.onComplete != nil {
		c.onComplete(newValue)
	}
}

func (c *PathCompleter) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if c.visible {
				c.selectCurrent()
				return true, nil
			}
			c.findMatches()
			if c.visible {
				return true, nil
			}
			return false, nil
		case tea.KeyEnter:
			if c.visible {
				c.hideCompletion()
				return false, nil
			}
			return false, nil
		case tea.KeyUp:
			if c.visible && len(c.matches) > 0 {
				c.selected--
				if c.selected < 0 {
					c.selected = len(c.matches) - 1
				}
				c.list.Select(c.selected)
				return true, nil
			}
		case tea.KeyDown:
			if c.visible && len(c.matches) > 0 {
				c.selected = (c.selected + 1) % len(c.matches)
				c.list.Select(c.selected)
				return true, nil
			}
		case tea.KeyEsc:
			if c.visible {
				c.hideCompletion()
				if c.onCancel != nil {
					c.onCancel()
				}
				return true, nil
			}
		case tea.KeyCtrlC:
			if c.visible {
				c.hideCompletion()
			}
			if c.onCancel != nil {
				c.onCancel()
			}
			return false, nil
		default:
			if c.visible {
				inputVal := c.textInput.Value()
				if inputVal == "" || (!strings.Contains(inputVal, "/") &&
					!strings.HasPrefix(inputVal, "./") &&
					!strings.HasPrefix(inputVal, "../")) {
					c.hideCompletion()
				} else {
					c.findMatches()
				}
			}
		}
	}

	var cmds []tea.Cmd

	updated, cmd := c.textInput.Update(msg)
	c.textInput = updated
	cmds = append(cmds, cmd)

	if c.visible {
		updated, cmd := c.list.Update(msg)
		c.list = updated
		cmds = append(cmds, cmd)
	}

	return false, tea.Batch(cmds...)
}

func (c *PathCompleter) View() string {
	inputView := c.textInput.View()

	if !c.visible {
		return inputView
	}

	itemsView := c.list.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		"\n",
		itemsView,
		helpStyle.Render(" [↑↓] Navigate  [Tab] Complete  [Esc] Dismiss"),
	)
}

func (c *PathCompleter) Snapshot() interface{} {
	return struct {
		Type     string
		Value    string
		Visible  bool
		Count    int
		Selected int
	}{
		Type:     "PathCompleter",
		Value:    c.textInput.Value(),
		Visible:  c.visible,
		Count:    len(c.matches),
		Selected: c.selected,
	}
}
