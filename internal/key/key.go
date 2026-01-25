package key

import tea "github.com/charmbracelet/bubbletea"

func IsCancel(msg tea.Msg) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return true
		case tea.KeyRunes:
			return km.String() == "q"
		}
	}
	return false
}

func IsEnter(msg tea.Msg) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		return km.Type == tea.KeyEnter
	}
	return false
}

func IsTab(msg tea.Msg) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		return km.Type == tea.KeyTab
	}
	return false
}

func IsUp(msg tea.Msg) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		return km.Type == tea.KeyUp || km.String() == "k"
	}
	return false
}

func IsDown(msg tea.Msg) bool {
	if km, ok := msg.(tea.KeyMsg); ok {
		return km.Type == tea.KeyDown || km.String() == "j"
	}
	return false
}
