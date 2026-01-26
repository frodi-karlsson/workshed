package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestHealthView_DetectsIssues(t *testing.T) {
	executions := []workspace.ExecutionRecord{
		{ID: "exec-1", Timestamp: time.Now().Add(-35 * 24 * time.Hour), Handle: "test-ws", Command: []string{"old"}, ExitCode: 0},
		{ID: "exec-2", Timestamp: time.Now().Add(-40 * 24 * time.Hour), Handle: "test-ws", Command: []string{"older"}, ExitCode: 0},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "missing-repo", URL: "https://github.com/org/missing-repo"},
				},
			},
		}),
		snapshot.WithExecutions(executions),
	})
	scenario.Enter("Open context menu")
	scenario.Key("k", "Navigate to health")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestHealthView_DryRun(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions([]workspace.ExecutionRecord{
			{ID: "exec-1", Timestamp: time.Now().Add(-35 * 24 * time.Hour), Handle: "test-ws", Command: []string{"stale"}, ExitCode: 0},
		}),
	})
	scenario.Enter("Open context menu")
	scenario.Key("k", "Navigate to health")
	scenario.Key("tab", "Toggle dry-run on")
	scenario.Enter("Preview cleanup actions")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestHealthView_HealthyWorkspace(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "healthy-ws",
				Purpose:   "Healthy workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "main", URL: "https://github.com/org/main"},
				},
			},
		}),
		snapshot.WithExecutions([]workspace.ExecutionRecord{
			{ID: "exec-1", Timestamp: time.Now(), Handle: "healthy-ws", Command: []string{"recent"}, ExitCode: 0},
		}),
	})
	scenario.Enter("Open context menu")
	scenario.Key("k", "Navigate to health")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestHealthView_AfterCleanup(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions([]workspace.ExecutionRecord{
			{ID: "exec-1", Timestamp: time.Now().Add(-35 * 24 * time.Hour), Handle: "test-ws", Command: []string{"stale"}, ExitCode: 0},
		}),
	})
	scenario.Enter("Open context menu")
	scenario.Key("k", "Navigate to health")
	scenario.Key("tab", "Toggle dry-run on")
	scenario.Enter("Preview cleanup actions")
	scenario.Enter("Dismiss after dry-run")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestHealthView_ToggleDryRun(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions([]workspace.ExecutionRecord{
			{ID: "exec-1", Timestamp: time.Now().Add(-31 * 24 * time.Hour), Handle: "test-ws", Command: []string{"old"}, ExitCode: 0},
		}),
	})
	scenario.Enter("Open context menu")
	scenario.Key("k", "Navigate to health")
	scenario.Key("tab", "Toggle dry-run ON")
	output := scenario.Record()
	snapshot.Match(t, "dry_run_on", output)

	scenario.Key("tab", "Toggle dry-run OFF")
	output = scenario.Record()
	snapshot.Match(t, "dry_run_off", output)
}
