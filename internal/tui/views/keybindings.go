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
	var parts []string
	for _, b := range bindings {
		if b.Disabled {
			continue
		}
		if b.When != nil && !b.When() {
			continue
		}
		parts = append(parts, b.Help)
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
