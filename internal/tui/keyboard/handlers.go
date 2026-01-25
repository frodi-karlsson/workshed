package keyboard

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyHandler processes key press events and returns whether the event was handled.
// Implementations translate key presses into application actions.
type KeyHandler interface {
	// HandleKey processes a key message.
	// Returns true if the key was handled, along with an optional command to execute.
	HandleKey(msg tea.KeyMsg) (bool, tea.Cmd)
}

type KeyHandlerFunc func(msg tea.KeyMsg) (bool, tea.Cmd)

func (f KeyHandlerFunc) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	return f(msg)
}

type NavigationHandler struct {
	list *list.Model
}

func NewNavigationHandler(l *list.Model) NavigationHandler {
	return NavigationHandler{list: l}
}

func (h NavigationHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		h.list.CursorUp()
		return true, nil
	case tea.KeyDown, tea.KeyCtrlN:
		h.list.CursorDown()
		return true, nil
	case tea.KeyRunes:
		switch msg.String() {
		case "k":
			h.list.CursorUp()
			return true, nil
		case "j":
			h.list.CursorDown()
			return true, nil
		}
	case tea.KeyPgUp:
		for i := 0; i < 5; i++ {
			h.list.CursorUp()
		}
		return true, nil
	case tea.KeyPgDown:
		for i := 0; i < 5; i++ {
			h.list.CursorDown()
		}
		return true, nil
	case tea.KeyHome:
		items := h.list.Items()
		for i := 0; i < len(items); i++ {
			h.list.CursorUp()
		}
		return true, nil
	case tea.KeyEnd:
		items := h.list.Items()
		for i := 0; i < len(items); i++ {
			h.list.CursorDown()
		}
		return true, nil
	}
	return false, nil
}

type QuitHandler struct{}

func NewQuitHandler() QuitHandler {
	return QuitHandler{}
}

func (h QuitHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return true, tea.Quit
	case tea.KeyRunes:
		if msg.String() == "q" {
			return true, tea.Quit
		}
	}
	return false, nil
}

type ConfirmHandler struct {
	OnConfirm func() (bool, tea.Cmd)
}

func NewConfirmHandler(onConfirm func() (bool, tea.Cmd)) ConfirmHandler {
	return ConfirmHandler{OnConfirm: onConfirm}
}

func (h ConfirmHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		if h.OnConfirm != nil {
			return h.OnConfirm()
		}
		return true, nil
	}
	return false, nil
}

type CancelHandler struct {
	OnCancel func() (bool, tea.Cmd)
}

func NewCancelHandler(onCancel func() (bool, tea.Cmd)) CancelHandler {
	return CancelHandler{OnCancel: onCancel}
}

func (h CancelHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		if h.OnCancel != nil {
			return h.OnCancel()
		}
		return true, tea.Quit
	case tea.KeyRunes:
		if msg.String() == "n" || msg.String() == "N" {
			if h.OnCancel != nil {
				return h.OnCancel()
			}
			return true, tea.Quit
		}
	}
	return false, nil
}

type FocusHandler struct {
	focusIndex int
	focusCount int
	OnFocus    func(index int) (bool, tea.Cmd)
}

func NewFocusHandler(focusCount int) FocusHandler {
	return FocusHandler{
		focusIndex: 0,
		focusCount: focusCount,
	}
}

func (h *FocusHandler) SetFocusCount(count int) {
	h.focusCount = count
	if h.focusIndex >= count {
		h.focusIndex = count - 1
	}
}

func (h FocusHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.Type == tea.KeyTab || msg.String() == "shift+tab" {
		if msg.String() == "shift+tab" {
			h.focusIndex--
			if h.focusIndex < 0 {
				h.focusIndex = h.focusCount - 1
			}
		} else {
			h.focusIndex++
			if h.focusIndex >= h.focusCount {
				h.focusIndex = 0
			}
		}
		if h.OnFocus != nil {
			return h.OnFocus(h.focusIndex)
		}
		return true, nil
	}
	return false, nil
}

func (h FocusHandler) GetFocusIndex() int {
	return h.focusIndex
}

type SelectionHandler struct {
	list        *list.Model
	OnSelect    func(index int) (bool, tea.Cmd)
	multiSelect bool
	selected    map[int]bool
}

func NewSelectionHandler(l *list.Model) SelectionHandler {
	return SelectionHandler{
		list:        l,
		multiSelect: false,
		selected:    make(map[int]bool),
	}
}

func (h SelectionHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeySpace:
		if h.multiSelect {
			idx := h.list.Index()
			h.selected[idx] = !h.selected[idx]
			return true, nil
		}
	case tea.KeyEnter:
		if h.OnSelect != nil {
			return h.OnSelect(h.list.Index())
		}
	}
	return false, nil
}

func (h SelectionHandler) GetSelected() []int {
	var selected []int
	for idx := range h.selected {
		if h.selected[idx] {
			selected = append(selected, idx)
		}
	}
	return selected
}

type CompositeHandler struct {
	handlers []KeyHandler
}

func NewCompositeHandler() CompositeHandler {
	return CompositeHandler{handlers: []KeyHandler{}}
}

func (h *CompositeHandler) Add(handler KeyHandler) {
	h.handlers = append(h.handlers, handler)
}

func (h CompositeHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	for _, handler := range h.handlers {
		if handled, cmd := handler.HandleKey(msg); handled {
			return handled, cmd
		}
	}
	return false, nil
}

type InputHandler struct {
}

func NewInputHandler() InputHandler {
	return InputHandler{}
}

func (h InputHandler) HandleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		return true, nil
	case tea.KeyBackspace:
		return true, nil
	case tea.KeyDelete:
		return true, nil
	case tea.KeyLeft:
		return true, nil
	case tea.KeyRight:
		return true, nil
	case tea.KeyHome:
		return true, nil
	case tea.KeyEnd:
		return true, nil
	}
	return false, nil
}
