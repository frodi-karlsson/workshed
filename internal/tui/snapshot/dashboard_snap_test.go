package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestDashboardView_InitialStates(t *testing.T) {
	t.Run("empty_list", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("with_workspaces", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{
					Handle:    "api-migration",
					Purpose:   "API migration project",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "backend", URL: "https://github.com/org/backend"},
						{Name: "frontend", URL: "https://github.com/org/frontend"},
					},
				},
				{
					Handle:    "feature-dev",
					Purpose:   "Feature development",
					CreatedAt: time.Now(),
					Repositories: []workspace.Repository{
						{Name: "api", URL: "https://github.com/org/api"},
					},
				},
			}),
		})
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestDashboardView_Navigation(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{Handle: "ws-1", Purpose: "First workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			{Handle: "ws-2", Purpose: "Second workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			{Handle: "ws-3", Purpose: "Third workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
		}),
	})

	output := scenario.Record()
	snapshot.Match(t, "first_selected", output)

	scenario.Key("j", "Navigate down")
	output = scenario.Record()
	snapshot.Match(t, "second_selected", output)

	scenario.Key("j", "Navigate down again")
	output = scenario.Record()
	snapshot.Match(t, "third_selected", output)
}

func TestDashboardView_Overlays(t *testing.T) {
	t.Run("wizard", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)
		scenario.Key("c", "Open wizard")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("keyboard_navigation", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "First", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-2", Purpose: "Second", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("j", "Navigate with j")
		scenario.Key("k", "Navigate back with k")
		scenario.Key("↑", "Navigate with arrow up")
		scenario.Key("↓", "Navigate with arrow down")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestDashboardView_FilterMode(t *testing.T) {
	t.Run("enter_filter", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "First", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-2", Purpose: "Second", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("filter_typing_c", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "cache-ws", Purpose: "Cache project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "api-ws", Purpose: "API project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Type("c", "Type 'c' in filter")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("filter_navigation", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "First", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-2", Purpose: "Second", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-3", Purpose: "Third", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("↓", "Navigate down in filter")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("filter_enter_opens_menu", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Enter("Press enter to open context menu")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("filter_and_cancel", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "API project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("esc", "Cancel filter")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_created_desc", func(t *testing.T) {
		now := time.Now()
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "old-ws", Purpose: "Old workspace", CreatedAt: now.Add(-2 * time.Hour), Repositories: []workspace.Repository{}},
				{Handle: "new-ws", Purpose: "New workspace", CreatedAt: now, Repositories: []workspace.Repository{}},
				{Handle: "mid-ws", Purpose: "Middle workspace", CreatedAt: now.Add(-1 * time.Hour), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_created_asc", func(t *testing.T) {
		now := time.Now()
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "old-ws", Purpose: "Old workspace", CreatedAt: now.Add(-2 * time.Hour), Repositories: []workspace.Repository{}},
				{Handle: "new-ws", Purpose: "New workspace", CreatedAt: now, Repositories: []workspace.Repository{}},
				{Handle: "mid-ws", Purpose: "Middle workspace", CreatedAt: now.Add(-1 * time.Hour), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("s", "Cycle to created asc")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_purpose_asc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "Zebra project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-2", Purpose: "Alpha project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-3", Purpose: "Beta project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("s", "Cycle to created asc")
		scenario.Key("s", "Cycle to purpose asc")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_purpose_desc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "ws-1", Purpose: "Alpha project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-2", Purpose: "Zebra project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "ws-3", Purpose: "Beta project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("s", "Cycle to created asc")
		scenario.Key("s", "Cycle to purpose asc")
		scenario.Key("s", "Cycle to purpose desc")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_handle_asc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "zebra-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "alpha-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "beta-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("s", "Cycle to created asc")
		scenario.Key("s", "Cycle to purpose asc")
		scenario.Key("s", "Cycle to purpose desc")
		scenario.Key("s", "Cycle to handle asc")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("sort_handle_desc", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "alpha-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "zebra-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
				{Handle: "beta-ws", Purpose: "Project", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Key("l", "Enter filter mode")
		scenario.Key("s", "Cycle to created asc")
		scenario.Key("s", "Cycle to purpose asc")
		scenario.Key("s", "Cycle to purpose desc")
		scenario.Key("s", "Cycle to handle asc")
		scenario.Key("s", "Cycle to handle desc")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestDashboardView_ContextMenu(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:    "test-ws",
				Purpose:   "Test workspace",
				Path:      "/test/path/to/workspace",
				CreatedAt: time.Now(),
				Repositories: []workspace.Repository{
					{Name: "repo1", URL: "https://github.com/org/repo1"},
				},
			},
		}),
	})

	scenario.Enter("Open context menu")
	output := scenario.Record()
	snapshot.Match(t, "context_menu_open", output)

	scenario.Key("i", "Select inspect")
	output = scenario.Record()
	snapshot.Match(t, "inspect_view", output)

	scenario.Key("esc", "Dismiss inspect")
	output = scenario.Record()
	snapshot.Match(t, "inspect_dismissed", output)
}

func TestDashboardView_ContextMenuActions(t *testing.T) {
	t.Run("path", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", Path: "/test/path/to/workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("p", "Select path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("exec", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{{Name: "repo1", URL: "https://github.com/org/repo1"}}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("e", "Select exec")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("update_purpose", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("u", "Select update purpose")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("remove_confirm", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("r", "Select remove repo")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("capture", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("c", "Select capture")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("pagination", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		output := scenario.Record()
		snapshot.Match(t, "initial", output)

		scenario.Key("down", "Move down to second item")
		output = scenario.Record()
		snapshot.Match(t, "moved_down", output)

		scenario.Key("tab", "Go to next page")
		output = scenario.Record()
		snapshot.Match(t, "next_page", output)
	})
}

func TestDashboardView_Modals(t *testing.T) {
	t.Run("remove_dismiss", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("r", "Select remove")
		scenario.Key("n", "Dismiss removal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("remove_confirm", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Test workspace", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("r", "Select remove")
		scenario.Key("y", "Confirm removal")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("update_purpose_success", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
			snapshot.WithWorkspaces([]*workspace.Workspace{
				{Handle: "test-ws", Purpose: "Old purpose", CreatedAt: time.Now(), Repositories: []workspace.Repository{}},
			}),
		})
		scenario.Enter("Open context menu")
		scenario.Key("u", "Select update")
		scenario.Type("New purpose", "Enter new purpose")
		scenario.Enter("Save purpose")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}
