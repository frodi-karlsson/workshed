package snapshot_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestImportWizardView_Wizard(t *testing.T) {
	t.Run("initial_state", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)
		scenario.Key("i", "Navigate to import wizard")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("preview_step", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
				{Name: "repo2", URL: "https://github.com/org/repo2"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "export.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("export.json", "Enter file path")
		scenario.Enter("Confirm path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("empty_path_validation", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)
		scenario.Key("i", "Navigate to import wizard")
		scenario.Enter("Try empty path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("back_from_preview", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("esc", "Go back to file path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("back_from_options", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("esc", "Go back to preview")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestImportWizardView_Options(t *testing.T) {
	t.Run("options_step", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("toggle_preserve_handle", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("p", "Toggle preserve handle")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("toggle_force", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("f", "Toggle force")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("both_options_toggled", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("p", "Toggle preserve handle")
		scenario.Key("f", "Toggle force")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestImportWizardView_Success(t *testing.T) {
	t.Run("import_completes", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("enter", "Import workspace")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})

	t.Run("dismiss_success", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)

		ctx := &workspace.WorkspaceContext{
			Version:     1,
			Handle:      "imported-ws",
			Purpose:     "Imported workspace",
			GeneratedAt: time.Now(),
			Repositories: []workspace.ContextRepo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		}
		data, _ := json.MarshalIndent(ctx, "", "  ")
		cwd := scenario.GetInvocationCWD()
		if err := os.WriteFile(filepath.Join(cwd, "file.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("file.json", "Enter file path")
		scenario.Enter("Confirm path")
		scenario.Key("enter", "Continue to options")
		scenario.Key("enter", "Import workspace")
		scenario.Enter("Dismiss success")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}

func TestImportWizardView_Error(t *testing.T) {
	t.Run("file_not_found", func(t *testing.T) {
		scenario := snapshot.NewScenario(t, nil, nil)
		scenario.Key("i", "Navigate to import wizard")
		scenario.Type("nonexistent.json", "Enter file path")
		scenario.Enter("Confirm path")
		output := scenario.Record()
		snapshot.Match(t, t.Name(), output)
	})
}
