//go:build !integration
// +build !integration

package wizard

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/workspace"
)

func TestPurposeStep_Initialization(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Existing purpose 1",
			Path:         "/test/ws1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws2",
			Purpose:      "Existing purpose 2",
			Path:         "/test/ws2",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	step := NewPurposeStep(workspaces)

	if step.done {
		t.Error("Expected done to be false initially")
	}

	if step.cancelled {
		t.Error("Expected cancelled to be false initially")
	}

	if step.mode != modeTyping {
		t.Error("Expected initial mode to be modeTyping")
	}

	if !step.textInput.Focused() {
		t.Error("Expected textInput to be focused initially")
	}
}

func TestPurposeStep_InitializationEmptyWorkspaces(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	if step.done {
		t.Error("Expected done to be false with empty workspaces")
	}

	items := step.list.Items()
	if len(items) != 0 {
		t.Error("Expected no suggestion items with empty workspaces")
	}
}

func TestPurposeStep_UpdateCancelEsc(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	updatedStep, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !updatedStep.cancelled {
		t.Error("Expected cancelled to be true after ESC")
	}

	if cmd == nil {
		t.Error("Expected non-nil command on ESC")
	}
}

func TestPurposeStep_UpdateCancelCtrlC(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	updatedStep, cmd := step.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !updatedStep.cancelled {
		t.Error("Expected cancelled to be true after Ctrl+C")
	}

	if cmd == nil {
		t.Error("Expected non-nil command on Ctrl+C")
	}
}

func TestPurposeStep_UpdateCancelQ(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	updatedStep, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if !updatedStep.cancelled {
		t.Error("Expected cancelled to be true after 'q'")
	}

	if cmd == nil {
		t.Error("Expected non-nil command on 'q'")
	}
}

func TestPurposeStep_ToggleMode(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	if step.mode != modeTyping {
		t.Error("Expected initial mode to be modeTyping")
	}

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if updatedStep.mode != modeSelecting {
		t.Error("Expected mode to toggle to modeSelecting after Tab")
	}

	updatedStep, _ = updatedStep.Update(tea.KeyMsg{Type: tea.KeyTab})

	if updatedStep.mode != modeTyping {
		t.Error("Expected mode to toggle back to modeTyping after second Tab")
	}
}

func TestPurposeStep_NavigationSwitchesToSelectingMode(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	if step.mode != modeTyping {
		t.Error("Expected initial mode to be modeTyping")
	}

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyUp})

	if updatedStep.mode != modeSelecting {
		t.Error("Expected mode to switch to modeSelecting on KeyUp")
	}

	updatedStep, _ = updatedStep.Update(tea.KeyMsg{Type: tea.KeyDown})

	if updatedStep.mode != modeSelecting {
		t.Error("Expected mode to remain modeSelecting on KeyDown")
	}
}

func TestPurposeStep_EnterWithEmptyInput(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if updatedStep.done {
		t.Error("Expected done to be false when entering empty purpose in typing mode")
	}
}

func TestPurposeStep_EnterWithValidInput(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	step.textInput.SetValue("New purpose")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !updatedStep.done {
		t.Error("Expected done to be true after entering valid purpose")
	}

	if updatedStep.resultValue != "New purpose" {
		t.Errorf("Expected resultValue 'New purpose', got '%s'", updatedStep.resultValue)
	}
}

func TestPurposeStep_SelectFromList(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Existing purpose",
			Path:         "/test/ws1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	step := NewPurposeStep(workspaces)

	step.mode = modeSelecting
	step.list.Select(0)

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !updatedStep.done {
		t.Error("Expected done to be true after selecting from list")
	}

	if updatedStep.resultValue != "Existing purpose" {
		t.Errorf("Expected resultValue 'Existing purpose', got '%s'", updatedStep.resultValue)
	}
}

func TestPurposeStep_TextInputFiltersSuggestions(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "OpenCode development",
			Path:         "/test/ws1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws2",
			Purpose:      "API migration",
			Path:         "/test/ws2",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	step := NewPurposeStep(workspaces)

	for _, r := range "open" {
		step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	items := step.list.Items()
	if len(items) != 1 {
		t.Errorf("Expected 1 filtered item, got %d", len(items))
	}
}

func TestPurposeStep_IsDone(t *testing.T) {
	step := NewPurposeStep(nil)

	if step.IsDone() {
		t.Error("Expected IsDone() to return false initially")
	}

	step.done = true

	if !step.IsDone() {
		t.Error("Expected IsDone() to return true after done is set")
	}
}

func TestPurposeStep_IsCancelled(t *testing.T) {
	step := NewPurposeStep(nil)

	if step.IsCancelled() {
		t.Error("Expected IsCancelled() to return false initially")
	}

	step.cancelled = true

	if !step.IsCancelled() {
		t.Error("Expected IsCancelled() to return true after cancelled is set")
	}
}

func TestPurposeStep_GetResult(t *testing.T) {
	step := NewPurposeStep(nil)

	if step.GetResult() != "" {
		t.Error("Expected GetResult() to return empty string initially")
	}

	step.resultValue = "Test purpose"

	if step.GetResult() != "Test purpose" {
		t.Errorf("Expected GetResult() to return 'Test purpose', got '%s'", step.GetResult())
	}
}

func TestPurposeStep_GetPurpose(t *testing.T) {
	step := NewPurposeStep(nil)

	if step.GetPurpose() != "" {
		t.Error("Expected GetPurpose() to return empty string initially")
	}

	step.resultValue = "Test purpose"

	if step.GetPurpose() != "Test purpose" {
		t.Errorf("Expected GetPurpose() to return 'Test purpose', got '%s'", step.GetPurpose())
	}
}

func TestPurposeStep_Init(t *testing.T) {
	step := NewPurposeStep(nil)

	cmd := step.Init()

	if cmd == nil {
		t.Error("Expected Init to return non-nil command (textinput.Blink)")
	}
}

func TestPurposeStep_View(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewPurposeStep(workspaces)

	output := step.View()

	if output == "" {
		t.Error("Expected non-empty view output")
	}

	if !contains(output, "Create Workspace") {
		t.Error("View output should contain 'Create Workspace' title")
	}

	if !contains(output, "Purpose:") {
		t.Error("View output should contain 'Purpose:' label")
	}

	if !contains(output, "Enter") {
		t.Error("View output should contain Enter help text")
	}
}

func TestPurposeStep_ViewWithWorkspaces(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Existing purpose",
			Path:         "/test/ws1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	step := NewPurposeStep(workspaces)

	output := step.View()

	if !contains(output, "Suggestions") {
		t.Error("View output should contain 'Suggestions' when workspaces exist")
	}
}

func TestPurposeStep_ViewWithModeIndicator(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Existing purpose",
			Path:         "/test/ws1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	step := NewPurposeStep(workspaces)

	output := step.View()

	t.Logf("TYPING output:\n%s", output)

	if !contains(output, "TYPING") {
		t.Error("View output should contain 'TYPING' mode indicator")
	}

	step.mode = modeSelecting
	output = step.View()

	t.Logf("SELECTING output:\n%s", output)

	if !contains(output, "SELECTING") {
		t.Error("View output should contain 'SELECTING' mode indicator")
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
