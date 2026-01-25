//go:build !integration
// +build !integration

package modalViews

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/workspace"
)

func TestPathModal_Initialization(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	onDismiss := func() {}

	modal := NewPathModal(ws, onDismiss)

	if modal.workspace != ws {
		t.Error("Expected workspace to be set")
	}

	if modal.onDismiss == nil {
		t.Error("Expected onDismiss callback to be set")
	}

	if modal.dismissed {
		t.Error("Expected dismissed to be false initially")
	}
}

func TestPathModal_UpdateDismissEsc(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewPathModal(ws, func() { dismissedCalled = true })

	updatedModal, dismissed := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true on ESC")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
	}

	if !dismissedCalled {
		t.Error("Expected onDismiss callback to be invoked")
	}
}

func TestPathModal_UpdateDismissCtrlC(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewPathModal(ws, func() { dismissedCalled = true })

	updatedModal, dismissed := modal.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true on Ctrl+C")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
	}

	if !dismissedCalled {
		t.Error("Expected onDismiss callback to be invoked")
	}
}

func TestPathModal_UpdateDismissEnter(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewPathModal(ws, func() { dismissedCalled = true })

	updatedModal, dismissed := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true on Enter")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
	}

	if !dismissedCalled {
		t.Error("Expected onDismiss callback to be invoked")
	}
}

func TestPathModal_UpdateDismissQ(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewPathModal(ws, func() { dismissedCalled = true })

	updatedModal, dismissed := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true on 'q'")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
	}

	if !dismissedCalled {
		t.Error("Expected onDismiss callback to be invoked")
	}
}

func TestPathModal_UpdateNoDismissOnOtherKeys(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewPathModal(ws, func() { dismissedCalled = true })

	updatedModal, dismissed := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	if dismissed {
		t.Error("Expected Update to return dismissed=false on other keys")
	}

	if updatedModal.dismissed {
		t.Error("Expected dismissed flag to remain false")
	}

	if dismissedCalled {
		t.Error("Expected onDismiss callback NOT to be invoked on other keys")
	}
}

func TestPathModal_DismissedMethod(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	m := NewPathModal(ws, func() {})

	if m.Dismissed() {
		t.Error("Expected Dismissed() to return false initially")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !m.Dismissed() {
		t.Error("Expected Dismissed() to return true after ESC")
	}
}

func TestPathModal_NoCallbackOnInit(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	NewPathModal(ws, func() { dismissedCalled = true })

	if dismissedCalled {
		t.Error("Expected onDismiss callback NOT to be invoked during initialization")
	}
}

func TestPathModal_CallbackWithNil(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	m := NewPathModal(ws, nil)

	updatedModal, dismissed := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
	}
}
