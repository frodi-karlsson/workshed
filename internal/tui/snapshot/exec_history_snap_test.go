package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestExecHistoryView_RecentList(t *testing.T) {
	now := time.Now()
	executions := []workspace.ExecutionRecord{
		{
			ID:          "exec-1",
			Timestamp:   now,
			Handle:      "test-ws",
			Command:     []string{"npm", "test"},
			ExitCode:    0,
			StartedAt:   now.Add(-5 * time.Second),
			CompletedAt: now,
			Duration:    5000,
		},
		{
			ID:          "exec-2",
			Timestamp:   now.Add(-10 * time.Minute),
			Handle:      "test-ws",
			Command:     []string{"go", "build", "./..."},
			ExitCode:    0,
			StartedAt:   now.Add(-10*time.Minute - 30*time.Second),
			CompletedAt: now.Add(-10 * time.Minute),
			Duration:    30000,
		},
		{
			ID:          "exec-3",
			Timestamp:   now.Add(-30 * time.Minute),
			Handle:      "test-ws",
			Command:     []string{"git", "push", "origin", "main"},
			ExitCode:    1,
			StartedAt:   now.Add(-30*time.Minute - 2*time.Second),
			CompletedAt: now.Add(-30 * time.Minute),
			Duration:    2000,
			Results: []workspace.ExecutionRepoResult{
				{Repository: "main-repo", ExitCode: 1, Error: "permission denied"},
			},
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "main-repo", URL: "https://github.com/org/main-repo"},
				},
			},
		}),
		snapshot.WithExecutions(executions),
	})
	scenario.Enter("Open context menu")
	scenario.Key("h", "Navigate to history")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestExecHistoryView_Navigation(t *testing.T) {
	now := time.Now()
	executions := []workspace.ExecutionRecord{
		{
			ID:        "exec-1",
			Timestamp: now,
			Handle:    "test-ws",
			Command:   []string{"cmd1"},
			ExitCode:  0,
		},
		{
			ID:        "exec-2",
			Timestamp: now.Add(-5 * time.Minute),
			Handle:    "test-ws",
			Command:   []string{"cmd2"},
			ExitCode:  0,
		},
		{
			ID:        "exec-3",
			Timestamp: now.Add(-10 * time.Minute),
			Handle:    "test-ws",
			Command:   []string{"cmd3"},
			ExitCode:  1,
		},
		{
			ID:        "exec-4",
			Timestamp: now.Add(-15 * time.Minute),
			Handle:    "test-ws",
			Command:   []string{"cmd4"},
			ExitCode:  0,
		},
		{
			ID:        "exec-5",
			Timestamp: now.Add(-20 * time.Minute),
			Handle:    "test-ws",
			Command:   []string{"cmd5"},
			ExitCode:  0,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions(executions),
	})
	scenario.Enter("Open context menu")
	scenario.Key("h", "Navigate to history")
	output := scenario.Record()
	snapshot.Match(t, "first_selected", output)

	scenario.Key("j", "Navigate down")
	output = scenario.Record()
	snapshot.Match(t, "second_selected", output)

	scenario.Key("j", "Navigate down again")
	output = scenario.Record()
	snapshot.Match(t, "third_selected", output)
}

func TestExecHistoryView_Empty(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions([]workspace.ExecutionRecord{}),
	})
	scenario.Enter("Open context menu")
	scenario.Key("h", "Navigate to history")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestExecHistoryView_AllFailures(t *testing.T) {
	now := time.Now()
	executions := []workspace.ExecutionRecord{
		{
			ID:        "exec-1",
			Timestamp: now,
			Handle:    "test-ws",
			Command:   []string{"npm", "run", "lint"},
			ExitCode:  1,
		},
		{
			ID:        "exec-2",
			Timestamp: now.Add(-1 * time.Hour),
			Handle:    "test-ws",
			Command:   []string{"go", "test"},
			ExitCode:  1,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Failing workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithExecutions(executions),
	})
	scenario.Enter("Open context menu")
	scenario.Key("h", "Navigate to history")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}
