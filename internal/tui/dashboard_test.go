//go:build !integration
// +build !integration

package tui

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

func TestDashboardInitialRender(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	requireOutputContains(t, tm, "Workshed Dashboard")

	quitAndWait(tm, t)
}

func TestDashboardWithWorkspaces(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:    "test-workspace",
			Purpose:   "Test purpose",
			Path:      "/test/workspaces/test-workspace",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	time.Sleep(50 * time.Millisecond)

	output, _ := io.ReadAll(tm.Output())
	outputStr := string(output)

	if !strings.Contains(outputStr, "test-workspace") {
		t.Errorf("Output should contain 'test-workspace'")
	}
	if !strings.Contains(outputStr, "Test purpose") {
		t.Errorf("Output should contain 'Test purpose'")
	}
	if !strings.Contains(outputStr, "1 repo") {
		t.Errorf("Output should contain '1 repo'")
	}

	quitAndWait(tm, t)
}

func TestDashboardNavigate(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:       "workspace-1",
			Purpose:      "First workspace",
			Path:         "/test/workspaces/workspace-1",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "workspace-2",
			Purpose:      "Second workspace",
			Path:         "/test/workspaces/workspace-2",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	requireOutputContains(t, tm, "workspace-1")

	navigateDown(tm)

	requireOutputContains(t, tm, "workspace-2")

	navigateUp(tm)

	requireOutputContains(t, tm, "workspace-1")

	quitAndWait(tm, t)
}

func TestDashboardHelp(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	m1, _ := sendKey(m, "?")
	if m1.currentView != viewHelpModal {
		t.Errorf("Expected currentView to be viewHelpModal after '?' key, got %v", m1.currentView)
	}

	// Toggle it back by pressing any key
	m2, _ := sendKey(m1, "q")
	if m2.currentView != viewDashboard {
		t.Errorf("Expected currentView to be viewDashboard after toggle back, got %v", m2.currentView)
	}
}

func TestDashboardQuit(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressCtrlC(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestDashboardQuitEsc(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressEsc(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestDashboardQuitQ(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	updatedModel, cmd := sendKey(m, "q")

	if !updatedModel.quitting {
		t.Error("Expected quitting to be true after 'q' key")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}
}

func TestDashboardMultipleWorkspaces(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:       "alpha-workspace",
			Purpose:      "Alpha purpose",
			Path:         "/test/workspaces/alpha",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "beta-workspace",
			Purpose:      "Beta purpose",
			Path:         "/test/workspaces/beta",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "gamma-workspace",
			Purpose:      "Gamma purpose",
			Path:         "/test/workspaces/gamma",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	time.Sleep(50 * time.Millisecond)

	output, _ := io.ReadAll(tm.Output())
	outputStr := string(output)

	if !strings.Contains(outputStr, "alpha-workspace") {
		t.Errorf("Output should contain 'alpha-workspace'")
	}
	if !strings.Contains(outputStr, "beta-workspace") {
		t.Errorf("Output should contain 'beta-workspace'")
	}
	if !strings.Contains(outputStr, "gamma-workspace") {
		t.Errorf("Output should contain 'gamma-workspace'")
	}

	navigateDown(tm)
	navigateDown(tm)

	quitAndWait(tm, t)
}

func TestDashboardWorkspaceWithMultipleRepos(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:    "multi-repo",
			Purpose:   "Multi repo workspace",
			Path:      "/test/workspaces/multi-repo",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{Name: "frontend", URL: "https://github.com/org/frontend"},
				{Name: "backend", URL: "https://github.com/org/backend"},
				{Name: "api", URL: "https://github.com/org/api"},
			},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestDashboardEmptyState(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	time.Sleep(50 * time.Millisecond)

	output, _ := io.ReadAll(tm.Output())
	outputStr := string(output)

	if !strings.Contains(outputStr, "No workspaces found") {
		t.Errorf("Output should contain 'No workspaces found'")
	}
	if !strings.Contains(outputStr, "Press 'c' to create") {
		t.Errorf("Output should contain 'Press 'c' to create'")
	}

	quitAndWait(tm, t)
}

func TestDashboardCreateWorkspace(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	updatedModel, cmd := sendKey(m, "c")

	if updatedModel.currentView != viewCreateWizard {
		t.Errorf("Expected currentView to be viewCreateWizard after 'c' key, got %v", updatedModel.currentView)
	}
	if cmd != nil {
		t.Error("Expected no command for view switch")
	}
}

func TestDashboardNavigationWithJAndK(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:       "workspace-one",
			Purpose:      "First",
			Path:         "/test/ws/one",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "workspace-two",
			Purpose:      "Second",
			Path:         "/test/ws/two",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	// Wait for initial render before navigating
	requireOutputContains(t, tm, "workspace-one")

	pressKey(tm, "j")

	requireOutputContainsWithTimeout(t, tm, "workspace-two", 3*time.Second)

	pressKey(tm, "k")

	requireOutputContainsWithTimeout(t, tm, "workspace-one", 3*time.Second)

	quitAndWait(tm, t)
}

func TestDashboardTabNavigation(t *testing.T) {
	store := newMockStore(t, []*workspace.Workspace{})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressKey(tm, "\t")

	pressEsc(tm)

	quitAndWait(tm, t)
}

func TestDashboardShowsCreationDate(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	store := newMockStore(t, []*workspace.Workspace{
		{
			Handle:       "dated-workspace",
			Purpose:      "Dated purpose",
			Path:         "/test/workspaces/dated",
			CreatedAt:    created,
			Repositories: []workspace.Repository{},
		},
	})
	m := newTestDashboardModel(t, store)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	time.Sleep(50 * time.Millisecond)

	output, _ := io.ReadAll(tm.Output())
	outputStr := string(output)

	if !strings.Contains(outputStr, "dated-workspace") {
		t.Errorf("Output should contain 'dated-workspace'")
	}
	if !strings.Contains(outputStr, "Jan 15") {
		t.Errorf("Output should contain 'Jan 15'")
	}

	quitAndWait(tm, t)
}
