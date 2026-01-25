//go:build !integration
// +build !integration

package modalViews

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/workspace"
)

func TestInspectModal_Initialization(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	onDismiss := func() {}

	modal := NewInspectModal(ws, onDismiss)

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

func TestInspectModal_UpdateDismissEsc(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewInspectModal(ws, func() { dismissedCalled = true })

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

func TestInspectModal_UpdateDismissCtrlC(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewInspectModal(ws, func() { dismissedCalled = true })

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

func TestInspectModal_UpdateDismissEnter(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewInspectModal(ws, func() { dismissedCalled = true })

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

func TestInspectModal_UpdateDismissQ(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewInspectModal(ws, func() { dismissedCalled = true })

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

func TestInspectModal_UpdateNoDismissOnOtherKeys(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	modal := NewInspectModal(ws, func() { dismissedCalled = true })

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

func TestInspectModal_ViewOutput(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	modal := NewInspectModal(ws, func() {})

	output := modal.View()

	if output == "" {
		t.Error("Expected non-empty view output")
	}

	if !contains(output, "Workspace:") {
		t.Error("View output should contain 'Workspace:' label")
	}

	if !contains(output, "test-ws") {
		t.Error("View output should contain workspace handle")
	}

	if !contains(output, "Purpose:") {
		t.Error("View output should contain 'Purpose:' label")
	}

	if !contains(output, "Test purpose") {
		t.Error("View output should contain workspace purpose")
	}

	if !contains(output, "Path:") {
		t.Error("View output should contain 'Path:' label")
	}

	if !contains(output, "/test/path") {
		t.Error("View output should contain workspace path")
	}

	if !contains(output, "Dismiss") {
		t.Error("View output should contain dismiss help text")
	}
}

func TestInspectModal_ViewWithRepositories(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "test-ws",
		Purpose:   "Test purpose",
		Path:      "/test/path",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			{URL: "https://github.com/org/repo2", Ref: "", Name: "repo2"},
		},
	}

	modal := NewInspectModal(ws, func() {})

	output := modal.View()

	if !contains(output, "Repositories:") {
		t.Error("View output should contain 'Repositories:' label")
	}

	if !contains(output, "repo1") {
		t.Error("View output should contain first repository name")
	}

	if !contains(output, "repo2") {
		t.Error("View output should contain second repository name")
	}

	if !contains(output, "https://github.com/org/repo1") {
		t.Error("View output should contain first repository URL")
	}

	if !contains(output, "main") {
		t.Error("View output should contain ref for first repository")
	}
}

func TestInspectModal_DismissedMethod(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	m := NewInspectModal(ws, func() {})

	if m.Dismissed() {
		t.Error("Expected Dismissed() to return false initially")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !m.Dismissed() {
		t.Error("Expected Dismissed() to return true after ESC")
	}
}

func TestInspectModal_NoCallbackOnInit(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	var dismissedCalled bool

	NewInspectModal(ws, func() { dismissedCalled = true })

	if dismissedCalled {
		t.Error("Expected onDismiss callback NOT to be invoked during initialization")
	}
}

func TestInspectModal_CallbackWithNil(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Test purpose",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	m := NewInspectModal(ws, nil)

	updatedModal, dismissed := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !dismissed {
		t.Error("Expected Update to return dismissed=true")
	}

	if !updatedModal.dismissed {
		t.Error("Expected dismissed flag to be set")
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
