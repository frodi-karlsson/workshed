package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestCaptureCreateView_Success(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "my-workspace",
				Purpose:   "My project purpose",
				Path:      "/Users/dev/workspaces/my-workspace",
				CreatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
				Repositories: []workspace.Repository{
					{Name: "backend", URL: "https://github.com/org/backend", Ref: "main"},
					{Name: "frontend", URL: "https://github.com/org/frontend"},
				},
			},
		}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("n", "Create new capture")
	scenario.Type("Before migration", "Enter capture name")
	scenario.Enter("Confirm name")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureCreateView_WithDirtyRepo(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "clean-repo", URL: "https://github.org/clean"},
					{Name: "dirty-repo", URL: "https://github.org/dirty"},
				},
			},
		}),
		snapshot.WithDirtyRepo("dirty-repo"),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("n", "Create new capture")
	scenario.Type("State with changes", "Enter capture name")
	scenario.Enter("Confirm name")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureCreateView_MultipleReposMixed(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "mixed-state-ws",
				Purpose:   "Mixed clean/dirty workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "backend", URL: "https://github.com/org/backend", Ref: "main"},
					{Name: "frontend", URL: "https://github.com/org/frontend", Ref: "develop"},
					{Name: "shared", URL: "https://github.com/org/shared", Ref: "main"},
				},
			},
		}),
		snapshot.WithDirtyRepo("backend"),
		snapshot.WithDirtyRepo("shared"),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("n", "Create new capture")
	scenario.Type("Mixed state capture", "Enter capture name")
	scenario.Enter("Confirm name")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureCreateView_EmptyDescription(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "simple-ws",
				Purpose:   "Simple workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "main-repo", URL: "https://github.com/org/main-repo"},
				},
			},
		}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("n", "Create new capture")
	scenario.Type("Quick capture", "Enter capture name")
	scenario.Enter("Confirm name")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureListView_MultipleCaptures(t *testing.T) {
	now := time.Now()
	captures := []workspace.Capture{
		{
			ID:        "cap-1",
			Timestamp: now,
			Handle:    "test-ws",
			Name:      "Before refactor",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo1", Branch: "main", Commit: "abc123", Dirty: false}},
		},
		{
			ID:        "cap-2",
			Timestamp: now.Add(-1 * time.Hour),
			Handle:    "test-ws",
			Name:      "After initial cleanup",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo1", Branch: "develop", Commit: "def456", Dirty: false}},
		},
		{
			ID:        "cap-3",
			Timestamp: now.Add(-2 * time.Hour),
			Handle:    "test-ws",
			Name:      "Initial state",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo1", Branch: "main", Commit: "ghi789", Dirty: true}},
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: now.Add(-24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "repo1", URL: "https://github.com/org/repo1"},
				},
			},
		}),
		snapshot.WithCaptures(captures),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureListView_Empty(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "empty-ws",
				Purpose:      "Empty workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureDetailsView_Success(t *testing.T) {
	capture := workspace.Capture{
		ID:        "cap-ready",
		Timestamp: time.Now(),
		Handle:    "test-ws",
		Name:      "Clean snapshot",
		Kind:      workspace.CaptureKindManual,
		GitState: []workspace.GitRef{
			{Repository: "backend", Branch: "main", Commit: "abc123def", Dirty: false},
		},
	}
	preflight := workspace.ApplyPreflightResult{
		Valid: true,
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "backend", URL: "https://github.com/org/backend", Ref: "main"},
				},
			},
		}),
		snapshot.WithCaptures([]workspace.Capture{capture}),
		snapshot.WithPreflightResult(preflight),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Enter("Open capture details")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureDetailsView_ApplyConfirm(t *testing.T) {
	capture := workspace.Capture{
		ID:        "cap-123",
		Timestamp: time.Now(),
		Handle:    "test-ws",
		Name:      "State to apply",
		Kind:      workspace.CaptureKindManual,
		GitState: []workspace.GitRef{
			{Repository: "repo1", Branch: "feature", Commit: "xyz789", Dirty: false},
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
		snapshot.WithCaptures([]workspace.Capture{capture}),
		snapshot.WithPreflightResult(workspace.ApplyPreflightResult{Valid: true}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Enter("Open capture details")
	scenario.Enter("Confirm apply")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureDetailsView_PreflightBlocked(t *testing.T) {
	capture := workspace.Capture{
		ID:        "cap-blocked",
		Timestamp: time.Now(),
		Handle:    "test-ws",
		Name:      "Blocked capture",
		Kind:      workspace.CaptureKindManual,
		GitState: []workspace.GitRef{
			{Repository: "backend", Branch: "main", Commit: "abc123", Dirty: false},
			{Repository: "frontend", Branch: "develop", Commit: "def456", Dirty: false},
		},
	}
	preflight := workspace.ApplyPreflightResult{
		Valid: false,
		Errors: []workspace.ApplyPreflightError{
			{
				Repository: "backend",
				Reason:     workspace.ReasonDirtyWorkingTree,
				Details:    "Uncommitted changes in backend/",
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
					{Name: "backend", URL: "https://github.com/org/backend"},
					{Name: "frontend", URL: "https://github.com/org/frontend"},
				},
			},
		}),
		snapshot.WithCaptures([]workspace.Capture{capture}),
		snapshot.WithPreflightResult(preflight),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Enter("Open capture details")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCapturesMenuView_EmptyState(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "empty-ws",
				Purpose:      "Empty workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCapturesMenuView_WithCaptures(t *testing.T) {
	captures := []workspace.Capture{
		{
			ID:        "cap-1",
			Timestamp: time.Now(),
			Handle:    "test-ws",
			Name:      "First capture",
			Kind:      workspace.CaptureKindManual,
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
		snapshot.WithCaptures(captures),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureListView_WithFilter(t *testing.T) {
	now := time.Now()
	captures := []workspace.Capture{
		{
			ID:        "cap-1",
			Timestamp: now,
			Handle:    "test-ws",
			Name:      "Backend refactor",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "backend", Branch: "main", Commit: "abc123", Dirty: false}},
		},
		{
			ID:        "cap-2",
			Timestamp: now.Add(-1 * time.Hour),
			Handle:    "test-ws",
			Name:      "Frontend changes",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "frontend", Branch: "develop", Commit: "def456", Dirty: false}},
		},
		{
			ID:        "cap-3",
			Timestamp: now.Add(-2 * time.Hour),
			Handle:    "test-ws",
			Name:      "Backend initial",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "backend", Branch: "feature", Commit: "ghi789", Dirty: false}},
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: now.Add(-24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "backend", URL: "https://github.com/org/backend"},
					{Name: "frontend", URL: "https://github.com/org/frontend"},
				},
			},
		}),
		snapshot.WithCaptures(captures),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Key("/", "Enter filter mode")
	scenario.Type("backend", "Filter by backend")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureListView_FilterByBranch(t *testing.T) {
	now := time.Now()
	captures := []workspace.Capture{
		{
			ID:        "cap-1",
			Timestamp: now,
			Handle:    "test-ws",
			Name:      "Main branch capture",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo", Branch: "main", Commit: "abc123", Dirty: false}},
		},
		{
			ID:        "cap-2",
			Timestamp: now.Add(-1 * time.Hour),
			Handle:    "test-ws",
			Name:      "Dev branch capture",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo", Branch: "develop", Commit: "def456", Dirty: false}},
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: now.Add(-24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "repo", URL: "https://github.com/org/repo"},
				},
			},
		}),
		snapshot.WithCaptures(captures),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Key("/", "Enter filter mode")
	scenario.Type("main", "Filter by main branch")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestCaptureListView_FilterNoMatch(t *testing.T) {
	now := time.Now()
	captures := []workspace.Capture{
		{
			ID:        "cap-1",
			Timestamp: now,
			Handle:    "test-ws",
			Name:      "Some capture",
			Kind:      workspace.CaptureKindManual,
			GitState:  []workspace.GitRef{{Repository: "repo1", Branch: "main", Commit: "abc123", Dirty: false}},
		},
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				CreatedAt: now.Add(-24 * time.Hour),
				Repositories: []workspace.Repository{
					{Name: "repo1", URL: "https://github.com/org/repo1"},
				},
			},
		}),
		snapshot.WithCaptures(captures),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("c", "Open captures menu")
	scenario.Key("l", "Open captures list")
	scenario.Key("/", "Enter filter mode")
	scenario.Type("nonexistent", "Filter with no matches")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}
