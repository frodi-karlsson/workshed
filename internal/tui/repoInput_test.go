//go:build !integration
// +build !integration

package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

func TestRepoInput_AddRepository(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	t.Run("should allow adding repo when entering URL", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(50 * time.Millisecond)

		typeText(tm, "https://github.com/org/repo")
		pressEnter(tm)

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if len(rm.repositories) != 1 {
			t.Errorf("Expected 1 repository, got %d", len(rm.repositories))
		}
		if len(rm.repositories) > 0 && rm.repositories[0].URL != "https://github.com/org/repo" {
			t.Errorf("Expected repo URL to be 'https://github.com/org/repo', got %q", rm.repositories[0].URL)
		}
	})

	t.Run("should show mode indicator when adding repo", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if !rm.adding {
			t.Error("Expected model to be in adding mode")
		}
		if rm.mode != modeTyping {
			t.Errorf("Expected mode to be modeTyping, got %v", rm.mode)
		}
	})
}

func TestRepoInput_RemoveRepository(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	t.Run("should remove last repo when pressing d", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		m.repositories = []workspace.RepositoryOption{
			{URL: "https://github.com/org/repo1", Ref: "main"},
			{URL: "https://github.com/org/repo2", Ref: "develop"},
		}

		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "d")
		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if len(rm.repositories) != 1 {
			t.Errorf("Expected 1 repository after removal, got %d", len(rm.repositories))
		}
		if len(rm.repositories) > 0 && rm.repositories[0].URL != "https://github.com/org/repo1" {
			t.Errorf("Expected remaining repo to be repo1, got %q", rm.repositories[0].URL)
		}
	})
}

func TestRepoInput_CreateWorkspace(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	t.Run("should mark as done when pressing Enter", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressEnter(tm)

		time.Sleep(100 * time.Millisecond) // Allow model to process

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if !rm.done {
			t.Error("Expected model to be done")
		}
	})
}

func TestRepoInput_RecentRepos(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:  "ws1",
			Purpose: "Test 1",
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	t.Run("should show recent repos when available", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(100 * time.Millisecond)

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if !rm.adding {
			t.Error("Expected model to be in adding mode")
		}
		if len(rm.recentRepos) == 0 {
			t.Error("Expected recent repos to be available")
		}
	})
}

func TestRepoInput_TabBetweenFields(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:  "ws1",
			Purpose: "Test 1",
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	t.Run("should tab from URL to Ref field", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(50 * time.Millisecond)

		pressTab(tm)
		time.Sleep(50 * time.Millisecond)

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if rm.focusedField != 1 {
			t.Errorf("Expected focusedField to be 1 (Ref field), got %d", rm.focusedField)
		}
		if rm.mode != modeTyping {
			t.Errorf("Expected mode to be modeTyping, got %v", rm.mode)
		}
	})

	t.Run("should shift-tab from Ref back to URL field", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(50 * time.Millisecond)

		pressTab(tm)
		time.Sleep(50 * time.Millisecond)

		pressKey(tm, "shift+tab")
		time.Sleep(50 * time.Millisecond)

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if rm.focusedField != 0 {
			t.Errorf("Expected focusedField to be 0 (URL field), got %d", rm.focusedField)
		}
	})
}

func TestRepoInput_EscapeFromAddingMode(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	t.Run("should exit adding mode on Escape", func(t *testing.T) {
		m := newTestRepoStepModel(workspaces, "test purpose")
		tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

		pressKey(tm, "a")
		time.Sleep(50 * time.Millisecond)

		pressEsc(tm)
		time.Sleep(50 * time.Millisecond)

		quitAndWait(tm, t)

		finalModel := tm.FinalModel(t)
		rm := finalModel.(repoStepModel)

		if rm.adding {
			t.Error("Expected adding to be false after pressing Escape")
		}
	})
}
