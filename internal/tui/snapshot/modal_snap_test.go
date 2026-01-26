package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestModal_Inspect(t *testing.T) {
	t.Run("workspace", func(t *testing.T) {
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
		scenario.Key("i", "Navigate to inspect")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("single_repo", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:    "single-repo",
					Purpose:   "Single repository project",
					Path:      "/path/to/workspace",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "main-repo", URL: "https://github.com/org/main-repo", Ref: "develop"},
					},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("i", "Navigate to inspect")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("no_repos", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:       "no-repos",
					Purpose:      "Workspace without repos",
					Path:         "/path/to/no-repos",
					CreatedAt:    time.Now(),
					Repositories: []workspace.Repository{},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("i", "Navigate to inspect")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("dismiss_esc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:    "test-ws",
					Purpose:   "Test workspace",
					Path:      "/test/path",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "repo1", URL: "https://github.com/org/repo1"},
					},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("i", "Navigate to inspect")
		scenario.Key("esc", "Dismiss inspect modal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestModal_Path(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:       "test-ws",
					Purpose:      "Test workspace",
					Path:         "/Users/testuser/projects/test-workspace",
					CreatedAt:    time.Now(),
					Repositories: []workspace.Repository{},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("p", "Navigate to path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("dismiss_esc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:       "test-ws",
					Purpose:      "Test workspace",
					Path:         "/Users/test/path",
					CreatedAt:    time.Now(),
					Repositories: []workspace.Repository{},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("p", "Navigate to path")
		scenario.Key("esc", "Dismiss path modal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestModal_Remove(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "delete-me",
				Purpose:      "To be deleted",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
	})
	scenario.Enter("Open resource menu")
	scenario.Key("x", "Navigate to remove")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestModal_UpdatePurpose(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:       "old-purpose-ws",
					Purpose:      "Old purpose text",
					CreatedAt:    time.Now(),
					Repositories: []workspace.Repository{},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("u", "Navigate to update")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("dismiss_esc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:       "test-ws",
					Purpose:      "Old purpose",
					CreatedAt:    time.Now(),
					Repositories: []workspace.Repository{},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("u", "Navigate to update")
		scenario.Key("esc", "Dismiss update modal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestModal_Exec(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:    "multi-repo-ws",
					Purpose:   "Multi repository workspace",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "api", URL: "https://github.com/org/api"},
						{Name: "web", URL: "https://github.com/org/web"},
					},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("e", "Navigate to exec menu")
		scenario.Key("enter", "Open exec view")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("dismiss_esc", func(t *testing.T) {
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
		})
		scenario.Enter("Open resource menu")
		scenario.Key("e", "Navigate to exec menu")
		scenario.Key("enter", "Open exec view")
		scenario.Key("esc", "Dismiss exec modal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}
