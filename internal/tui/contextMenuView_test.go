//go:build !integration
// +build !integration

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestContextMenuView_Initialization(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	if view.handle != "test-workspace" {
		t.Errorf("Expected handle 'test-workspace', got '%s'", view.handle)
	}

	if view.selectedAction != "" {
		t.Errorf("Expected empty selectedAction initially, got '%s'", view.selectedAction)
	}

	if view.cancelled {
		t.Error("Expected cancelled to be false initially")
	}

	if view.list.Title != "Actions for \"test-workspace\"" {
		t.Errorf("Unexpected list title: %s", view.list.Title)
	}

	items := view.list.Items()
	if len(items) != 5 {
		t.Errorf("Expected 5 menu items, got %d", len(items))
	}
}

func TestContextMenuView_Selection(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view.list.Select(0)
	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if view.selectedAction != "i" {
		t.Errorf("Expected action 'i' for first item, got '%s'", view.selectedAction)
	}
}

func TestContextMenuView_CancelEsc(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !view.cancelled {
		t.Error("Expected cancelled to be true on ESC")
	}
}

func TestContextMenuView_CancelCtrlC(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !view.cancelled {
		t.Error("Expected cancelled to be true on Ctrl+C")
	}
}

func TestContextMenuView_Navigation(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	initialIndex := view.list.Index()

	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyDown})
	newIndex := view.list.Index()

	if newIndex != initialIndex+1 {
		t.Errorf("Expected index to increase by 1, got %d (was %d)", newIndex, initialIndex)
	}

	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyUp})
	newIndex = view.list.Index()

	if newIndex != initialIndex {
		t.Errorf("Expected index to return to %d, got %d", initialIndex, newIndex)
	}
}

func TestContextMenuView_JKeyNavigation(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	initialIndex := view.list.Index()
	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	newIndex := view.list.Index()

	if newIndex != initialIndex+1 {
		t.Errorf("Expected 'j' to navigate down, index changed from %d to %d", initialIndex, newIndex)
	}
}

func TestContextMenuView_KKeyNavigation(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view.list.Select(2)
	initialIndex := view.list.Index()
	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	newIndex := view.list.Index()

	if newIndex != initialIndex-1 {
		t.Errorf("Expected 'k' to navigate up, index changed from %d to %d", initialIndex, newIndex)
	}
}

func TestContextMenuView_ViewOutput(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	output := view.View()

	if output == "" {
		t.Error("Expected non-empty view output")
	}

	if !contains(output, "test-workspace") {
		t.Error("View output should contain workspace handle")
	}

	if !contains(output, "Navigate") {
		t.Error("View output should contain navigation help text")
	}

	if !contains(output, "Inspect") {
		t.Error("View output should contain Inspect menu item")
	}
}

func TestContextMenuView_Init(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	cmd := view.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestContextMenuView_NoSelectionOnEmptyList(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view.list.Select(100)
	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if view.selectedAction != "" {
		t.Errorf("Expected no selection when list has invalid index, got '%s'", view.selectedAction)
	}
}

func TestContextMenuView_AllMenuOptions(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	actions := []string{"i", "p", "e", "u", "r"}

	for i, action := range actions {
		view.list.Select(i)
		view, _ = view.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if view.selectedAction != action {
			t.Errorf("Expected action '%s' at index %d, got '%s'", action, i, view.selectedAction)
		}

		view = NewContextMenuView("test-workspace")
	}
}

func TestContextMenuView_EscCancelsAndDoesNotSetAction(t *testing.T) {
	view := NewContextMenuView("test-workspace")

	view.list.Select(0)
	view, _ = view.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !view.cancelled {
		t.Error("Expected cancelled to be true after ESC")
	}

	if view.selectedAction != "" {
		t.Errorf("Expected no selection when cancelled, got '%s'", view.selectedAction)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
