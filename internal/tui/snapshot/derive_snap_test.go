package snapshot_test

import (
	"errors"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestDeriveView_ContextPreview(t *testing.T) {
	now := time.Now()
	ctx := &workspace.WorkspaceContext{
		Version:     1,
		GeneratedAt: now,
		Handle:      "test-ws",
		Purpose:     "API migration project",
		Repositories: []workspace.ContextRepo{
			{Name: "backend", Path: "/workspaces/test-ws/backend", URL: "https://github.com/org/backend", RootPath: "/workspaces/test-ws/backend"},
			{Name: "frontend", Path: "/workspaces/test-ws/frontend", URL: "https://github.com/org/frontend", RootPath: "/workspaces/test-ws/frontend"},
			{Name: "shared", Path: "/workspaces/test-ws/shared", URL: "https://github.com/org/shared", RootPath: "/workspaces/test-ws/shared"},
		},
		Captures: []workspace.ContextCapture{
			{ID: "cap-1", Timestamp: now.Add(-1 * time.Hour), Name: "Before refactor", Kind: "manual", RepoCount: 3},
			{ID: "cap-2", Timestamp: now.Add(-2 * time.Hour), Name: "Initial state", Kind: "manual", RepoCount: 3},
		},
		Metadata: workspace.ContextMetadata{
			WorkshedVersion: "1.0.0",
			ExecutionsCount: 12,
			CapturesCount:   2,
			LastExecutedAt:  &now,
			LastCapturedAt:  &now,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "API migration project",
				CreatedAt: now.Add(-7 * 24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "backend", URL: "https://github.com/org/backend"},
					{Name: "frontend", URL: "https://github.com/org/frontend"},
					{Name: "shared", URL: "https://github.com/org/shared"},
				},
			},
		}),
		snapshot.WithCaptures([]workspace.Capture{
			{ID: "cap-1", Timestamp: now.Add(-1 * time.Hour), Handle: "test-ws", Name: "Before refactor", Kind: "manual"},
			{ID: "cap-2", Timestamp: now.Add(-2 * time.Hour), Handle: "test-ws", Name: "Initial state", Kind: "manual"},
		}),
		snapshot.WithContext(ctx),
	})
	scenario.Enter("Open context menu")
	scenario.Key("D", "Navigate to derive")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestDeriveView_EmptyWorkspace(t *testing.T) {
	ctx := &workspace.WorkspaceContext{
		Version:      1,
		GeneratedAt:  time.Now(),
		Handle:       "empty-ws",
		Purpose:      "Empty workspace",
		Repositories: []workspace.ContextRepo{},
		Captures:     []workspace.ContextCapture{},
		Metadata: workspace.ContextMetadata{
			WorkshedVersion: "1.0.0",
			ExecutionsCount: 0,
			CapturesCount:   0,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "empty-ws",
				Purpose:      "Empty workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithContext(ctx),
	})
	scenario.Enter("Open context menu")
	scenario.Key("D", "Navigate to derive")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestDeriveView_WithExecutions(t *testing.T) {
	now := time.Now()
	executions := []workspace.ExecutionRecord{
		{ID: "exec-1", Timestamp: now.Add(-30 * time.Minute), Handle: "test-ws", Command: []string{"npm", "test"}, ExitCode: 0},
		{ID: "exec-2", Timestamp: now.Add(-1 * time.Hour), Handle: "test-ws", Command: []string{"go", "build"}, ExitCode: 0},
		{ID: "exec-3", Timestamp: now.Add(-2 * time.Hour), Handle: "test-ws", Command: []string{"git", "status"}, ExitCode: 1},
	}
	ctx := &workspace.WorkspaceContext{
		Version:     1,
		GeneratedAt: now,
		Handle:      "test-ws",
		Purpose:     "Project with history",
		Repositories: []workspace.ContextRepo{
			{Name: "main", Path: "/ws/main", URL: "https://github.com/org/main", RootPath: "/ws/main"},
		},
		Captures: []workspace.ContextCapture{},
		Metadata: workspace.ContextMetadata{
			WorkshedVersion: "1.0.0",
			ExecutionsCount: 3,
			CapturesCount:   0,
			LastExecutedAt:  &now,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Project with history",
				CreatedAt: now.Add(-24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "main", URL: "https://github.com/org/main"},
				},
			},
		}),
		snapshot.WithExecutions(executions),
		snapshot.WithContext(ctx),
	})
	scenario.Enter("Open context menu")
	scenario.Key("D", "Navigate to derive")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestDeriveView_DeriveError(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithDeriveError(errors.New("workspace not found")),
	})
	scenario.Enter("Open context menu")
	scenario.Key("D", "Navigate to derive")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestDeriveView_AfterCopyToClipboard(t *testing.T) {
	ctx := &workspace.WorkspaceContext{
		Version:     1,
		GeneratedAt: time.Now(),
		Handle:      "test-ws",
		Purpose:     "Test workspace",
		Repositories: []workspace.ContextRepo{
			{Name: "repo1", Path: "/ws/repo1", URL: "https://github.com/org/repo1", RootPath: "/ws/repo1"},
		},
		Captures: []workspace.ContextCapture{},
		Metadata: workspace.ContextMetadata{
			WorkshedVersion: "1.0.0",
			ExecutionsCount: 0,
			CapturesCount:   0,
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "repo1", URL: "https://github.com/org/repo1"},
				},
			},
		}),
		snapshot.WithContext(ctx),
	})
	scenario.Enter("Open context menu")
	scenario.Key("D", "Navigate to derive")
	scenario.Enter("Copy JSON to clipboard")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}
