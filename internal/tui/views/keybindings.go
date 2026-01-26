package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type KeyAction func() (ViewResult, tea.Cmd)

type KeyBinding struct {
	Key      string
	Help     string
	Action   KeyAction
	When     func() bool
	Disabled bool
}

type KeyBinder interface {
	KeyBindings() []KeyBinding
}

func matchesKey(msg tea.KeyMsg, key string) bool {
	switch key {
	case "enter":
		return msg.Type == tea.KeyEnter
	case "esc":
		return msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC
	case "ctrl+c":
		return msg.Type == tea.KeyCtrlC
	case "tab":
		return msg.Type == tea.KeyTab
	case "shift+tab":
		return msg.String() == "shift+tab"
	case "up":
		return msg.Type == tea.KeyUp || msg.String() == "k"
	case "down":
		return msg.Type == tea.KeyDown || msg.String() == "j"
	case "left":
		return msg.Type == tea.KeyLeft
	case "right":
		return msg.Type == tea.KeyRight
	case "pgup":
		return msg.Type == tea.KeyPgUp
	case "pgdown":
		return msg.Type == tea.KeyPgDown
	case "home":
		return msg.Type == tea.KeyHome
	case "end":
		return msg.Type == tea.KeyEnd
	case "space":
		return msg.Type == tea.KeySpace
	case "backspace":
		return msg.Type == tea.KeyBackspace
	case "delete":
		return msg.Type == tea.KeyDelete
	default:
		if len(key) == 1 {
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				return string(msg.Runes[0]) == key
			}
			return msg.String() == key
		}
		return false
	}
}

func GenerateHelp(bindings []KeyBinding) string {
	type HelpGroup struct {
		keys       []string
		normalized []string
	}

	groups := make(map[string]*HelpGroup)
	var order []string

	for _, b := range bindings {
		if b.Disabled {
			continue
		}
		if b.When != nil && !b.When() {
			continue
		}

		idx := strings.Index(b.Help, "] ")
		if idx == -1 {
			continue
		}
		desc := b.Help[idx+2:]

		if groups[desc] == nil {
			groups[desc] = &HelpGroup{}
			order = append(order, desc)
		}
		groups[desc].keys = append(groups[desc].keys, b.Key)
	}

	var parts []string
	for _, desc := range order {
		g := groups[desc]
		for _, key := range g.keys {
			g.normalized = append(g.normalized, normalizeKey(key))
		}
		parts = append(parts, "["+strings.Join(g.normalized, "/")+"] "+desc)
	}
	return strings.Join(parts, "  ")
}

func HandleKey(bindings []KeyBinding, msg tea.KeyMsg) (ViewResult, tea.Cmd, bool) {
	for _, b := range bindings {
		if b.Disabled {
			continue
		}
		if b.When != nil && !b.When() {
			continue
		}
		if matchesKey(msg, b.Key) {
			if b.Action != nil {
				result, cmd := b.Action()
				return result, cmd, true
			}
			return ViewResult{}, nil, true
		}
	}
	return ViewResult{}, nil, false
}

func BindingsHelpText(bindings []KeyBinding, condition func() bool, enabledText, disabledText string) string {
	for _, b := range bindings {
		if b.Disabled {
			return disabledText
		}
		if condition != nil && !condition() {
			return disabledText
		}
	}
	return enabledText
}

func GetDismissKeyBindings(dismiss func() (ViewResult, tea.Cmd), description string) []KeyBinding {
	return []KeyBinding{
		{Key: "q", Help: "[q] " + description, Action: dismiss},
		{Key: "esc", Help: "[Esc] " + description, Action: dismiss},
		{Key: "ctrl+c", Help: "[Ctrl+C] " + description, Action: dismiss},
	}
}

func normalizeKey(key string) string {
	switch key {
	case "ctrl+c":
		return "Ctrl+C"
	case "enter":
		return "Enter"
	case "tab":
		return "Tab"
	case "shift+tab":
		return "Shift+Tab"
	case "pgup":
		return "PgUp"
	case "pgdown":
		return "PgDown"
	case "backspace":
		return "Backspace"
	case "delete":
		return "Delete"
	case "home":
		return "Home"
	case "end":
		return "End"
	case "space":
		return "Space"
	case "esc":
		return "Esc"
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	default:
		return strings.ToUpper(key[:1]) + key[1:]
	}
}
