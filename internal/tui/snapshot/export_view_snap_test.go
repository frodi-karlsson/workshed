package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestExportView_View(t *testing.T) {
	t.Run("view exports", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:    "test-ws",
					Purpose:   "Test workspace",
					Path:      "/test/path",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "repo1", URL: "https://github.com/org/repo1"},
						{Name: "repo2", URL: "https://github.com/org/repo2"},
					},
				},
			}),
		})
		scenario.Enter("Open resource menu")
		scenario.Key("x", "Navigate to export view")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("copy to clipboard", func(t *testing.T) {
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
		scenario.Key("x", "Navigate to export view")
		scenario.Key("c", "Copy to clipboard")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("save to file", func(t *testing.T) {
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
		scenario.Key("x", "Navigate to export view")
		scenario.Key("f", "Select save to file")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("back navigation", func(t *testing.T) {
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
		scenario.Key("x", "Navigate to export view")
		scenario.Key("esc", "Go back")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestExportView_InfoMenu(t *testing.T) {
	t.Run("export via info menu", func(t *testing.T) {
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
		scenario.Key("i", "Navigate to info menu")
		scenario.Key("e", "Navigate to export view")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}
