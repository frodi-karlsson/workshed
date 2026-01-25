//go:build !integration
// +build !integration

package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

func TestPurposeInput_TypingMode(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "api-migration",
			Purpose:      "API migration",
			Path:         "/test/workspaces/api-migration",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	t.Run("should use typed value when pressing Enter in typing mode", func(t *testing.T) {
		m := newTestPurposeStepModel(workspaces)
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		typeText(tm, "new feature")
		pressEnter(tm)

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		pm := finalModel.(purposeStepModel)

		if !pm.done {
			t.Error("Expected model to be done")
		}
		if pm.resultValue != "new feature" {
			t.Errorf("Expected result value to be 'new feature', got %q", pm.resultValue)
		}
	})

	t.Run("should filter suggestions while typing", func(t *testing.T) {
		m := newTestPurposeStepModel(workspaces)
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		typeText(tm, "API")

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		pm := finalModel.(purposeStepModel)

		// Check that the list has been filtered to show "API migration"
		found := false
		for _, item := range pm.list.Items() {
			if pItem, ok := item.(purposeItem); ok {
				if pItem.purpose == "API migration" {
					found = true
					break
				}
			}
		}

		if !found {
			t.Error("Expected filtered suggestions to contain 'API migration'")
		}
	})
}

func TestPurposeInput_SelectingMode(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "api-migration",
			Purpose:      "API migration",
			Path:         "/test/workspaces/api-migration",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "feature-dev",
			Purpose:      "Feature development",
			Path:         "/test/workspaces/feature-dev",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	t.Run("should use selected item when pressing Enter in selecting mode", func(t *testing.T) {
		m := newTestPurposeStepModel(workspaces)
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressTab(tm)
		pressEnter(tm)

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		pm := finalModel.(purposeStepModel)

		if !pm.done {
			t.Error("Expected model to be done")
		}
		if pm.resultValue != "API migration" && pm.resultValue != "Feature development" {
			t.Errorf("Expected result value to be one of the suggestions, got %q", pm.resultValue)
		}
	})

	t.Run("should auto-switch to selecting mode on arrow keys", func(t *testing.T) {
		m := newTestPurposeStepModel(workspaces)
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		navigateDown(tm)

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		pm := finalModel.(purposeStepModel)

		if pm.mode != modeSelecting {
			t.Errorf("Expected mode to be modeSelecting, got %v", pm.mode)
		}
	})
}

func TestPurposeInput_ModeToggling(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "api-migration",
			Purpose:      "API migration",
			Path:         "/test/workspaces/api-migration",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	t.Run("should toggle between typing and selecting modes on Tab", func(t *testing.T) {
		m := newTestPurposeStepModel(workspaces)
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		time.Sleep(50 * time.Millisecond)

		// Press Tab twice to toggle modes
		pressTab(tm)
		time.Sleep(50 * time.Millisecond)
		pressTab(tm)
		time.Sleep(50 * time.Millisecond)

		quitAndWait(tm, t)

		// Check final mode is back to typing
		finalModel := tm.FinalModel(t)
		pm := finalModel.(purposeStepModel)

		if pm.mode != modeTyping {
			t.Errorf("Expected final mode to be modeTyping after toggling, got %v", pm.mode)
		}
	})
}
