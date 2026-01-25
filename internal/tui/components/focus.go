package components

import "github.com/charmbracelet/lipgloss"

// FocusState represents which UI element currently has focus.
// This is used for keyboard navigation and focus management.
type FocusState int

const (
	// FocusNone indicates no element has focus.
	FocusNone FocusState = iota

	// FocusInput indicates a text input field has focus.
	FocusInput

	// FocusList indicates a list view has focus.
	FocusList

	// FocusButton indicates a button has focus.
	FocusButton

	// FocusCustom indicates a custom element has focus.
	FocusCustom
)

// FocusManager coordinates focus navigation between different UI elements.
// It maintains a ordered list of focusable elements and handles traversal.
type FocusManager struct {
	currentFocus  FocusState
	focusOrder    []FocusState
	OnFocusChange func(oldState, newState FocusState)
}

func NewFocusManager() *FocusManager {
	fm := &FocusManager{}
	fm.focusOrder = []FocusState{FocusInput, FocusList, FocusButton}
	fm.currentFocus = FocusInput
	return fm
}

func (fm *FocusManager) SetFocusOrder(order []FocusState) {
	fm.focusOrder = order
	if fm.currentFocus >= FocusState(len(order)) {
		fm.currentFocus = FocusNone
	}
}

func (fm *FocusManager) Next() FocusState {
	if fm.currentFocus == FocusNone {
		if len(fm.focusOrder) > 0 {
			fm.currentFocus = fm.focusOrder[0]
		}
	} else {
		found := false
		for i, focus := range fm.focusOrder {
			if focus == fm.currentFocus {
				found = true
				if i+1 < len(fm.focusOrder) {
					fm.currentFocus = fm.focusOrder[i+1]
				} else {
					fm.currentFocus = fm.focusOrder[0]
				}
				break
			}
		}
		if !found {
			fm.currentFocus = FocusNone
		}
	}

	if fm.OnFocusChange != nil {
		fm.OnFocusChange(FocusNone, fm.currentFocus)
	}

	return fm.currentFocus
}

func (fm *FocusManager) Previous() FocusState {
	if fm.currentFocus == FocusNone {
		if len(fm.focusOrder) > 0 {
			fm.currentFocus = fm.focusOrder[len(fm.focusOrder)-1]
		}
	} else {
		found := false
		for i, focus := range fm.focusOrder {
			if focus == fm.currentFocus {
				found = true
				if i > 0 {
					fm.currentFocus = fm.focusOrder[i-1]
				} else {
					fm.currentFocus = fm.focusOrder[len(fm.focusOrder)-1]
				}
				break
			}
		}
		if !found {
			fm.currentFocus = FocusNone
		}
	}

	if fm.OnFocusChange != nil {
		fm.OnFocusChange(FocusNone, fm.currentFocus)
	}

	return fm.currentFocus
}

func (fm *FocusManager) SetFocus(focus FocusState) bool {
	for _, f := range fm.focusOrder {
		if f == focus {
			fm.currentFocus = focus
			if fm.OnFocusChange != nil {
				fm.OnFocusChange(FocusNone, fm.currentFocus)
			}
			return true
		}
	}
	return false
}

func (fm *FocusManager) GetCurrentFocus() FocusState {
	return fm.currentFocus
}

func (fm *FocusManager) IsFocused(focus FocusState) bool {
	return fm.currentFocus == focus
}

func (fm *FocusManager) GetFocusIndex(focus FocusState) int {
	for i, f := range fm.focusOrder {
		if f == focus {
			return i
		}
	}
	return -1
}

func (fm *FocusManager) IsInputFocused() bool {
	return fm.currentFocus == FocusInput
}

func (fm *FocusManager) IsListFocused() bool {
	return fm.currentFocus == FocusList
}

func (fm *FocusManager) IsButtonFocused() bool {
	return fm.currentFocus == FocusButton
}

// FocusIndicator renders a visual cue showing which element has focus.
// It displays different characters/styles based on focus state.
type FocusIndicator struct {
	Style        lipgloss.Style
	FocusedStyle lipgloss.Style
	Char         string
	FocusedChar  string
}

func NewFocusIndicator() FocusIndicator {
	return FocusIndicator{
		Style: lipgloss.NewStyle().
			Foreground(ColorMuted),
		FocusedStyle: lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true),
		Char:        " ",
		FocusedChar: "▶",
	}
}

func (fi FocusIndicator) Render(focused bool) string {
	if focused {
		return fi.FocusedStyle.Render(fi.FocusedChar)
	}
	return fi.Style.Render(fi.Char)
}

func (fi *FocusIndicator) SetFocused(focused bool) *FocusIndicator {
	if focused {
		fi.Style = fi.Style.Foreground(ColorHighlight)
	} else {
		fi.Style = fi.Style.Foreground(ColorMuted)
	}
	return fi
}

// Focusable is the interface for UI elements that can receive focus.
// Implementations track and report their focus state.
type Focusable interface {
	// GetFocusState returns the current focus state of the element.
	GetFocusState() FocusState

	// SetFocus sets the focus state of the element.
	// The focused parameter indicates whether the element should be focused.
	SetFocus(focused bool)
}

// FocusableGroup manages a collection of Focusable elements.
// It allows focusing individual elements and clearing all focus at once.
type FocusableGroup struct {
	items []Focusable
}

func NewFocusableGroup() *FocusableGroup {
	return &FocusableGroup{items: []Focusable{}}
}

func (g *FocusableGroup) Add(item Focusable) {
	g.items = append(g.items, item)
}

func (g *FocusableGroup) GetItem(index int) Focusable {
	if index >= 0 && index < len(g.items) {
		return g.items[index]
	}
	return nil
}

func (g *FocusableGroup) Len() int {
	return len(g.items)
}

func (g *FocusableGroup) FocusIndex(index int) {
	for i, item := range g.items {
		item.SetFocus(i == index)
	}
}

func (g *FocusableGroup) ClearFocus() {
	for _, item := range g.items {
		item.SetFocus(false)
	}
}

// FocusableItem is a simple wrapper that treats a FocusState as a Focusable.
// It can be used to include focus states directly in a FocusableGroup.
type FocusableItem struct {
	state FocusState
}

func NewFocusableItem(state FocusState) *FocusableItem {
	return &FocusableItem{state: state}
}

func (i FocusableItem) GetFocusState() FocusState {
	return i.state
}

func (i *FocusableItem) SetFocus(focused bool) {
	if focused {
		i.state = FocusCustom
	}
}

// FocusStyle specifies how focus is visually indicated.
// Different styles provide various visual cues for the focused element.
type FocusStyle int

const (
	// FocusStyleNoIndicator hides the focus indicator.
	FocusStyleNone FocusStyle = iota

	// FocusStyleUnderline renders an underline for the focused element.
	FocusStyleUnderline

	// FocusStyleBorder renders a border around the focused element.
	FocusStyleBorder

	// FocusStyleHighlight renders highlighted background for the focused element.
	FocusStyleHighlight

	// FocusStyleReverse inverts foreground and background colors.
	FocusStyleReverse
)

func GetFocusStyle(style FocusStyle) lipgloss.Style {
	switch style {
	case FocusStyleUnderline:
		return lipgloss.NewStyle().
			Underline(true).
			Foreground(ColorHighlight)
	case FocusStyleBorder:
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorHighlight).
			Padding(0, 1)
	case FocusStyleHighlight:
		return lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorHighlight)
	case FocusStyleReverse:
		return lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorText)
	default:
		return lipgloss.NewStyle()
	}
}
