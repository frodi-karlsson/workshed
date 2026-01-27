package clitest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frodi/workshed/cmd/workshed/apply"
	"github.com/frodi/workshed/cmd/workshed/capture"
	"github.com/frodi/workshed/cmd/workshed/captures"
	"github.com/frodi/workshed/cmd/workshed/create"
	"github.com/frodi/workshed/cmd/workshed/export"
	"github.com/frodi/workshed/cmd/workshed/health"
	"github.com/frodi/workshed/cmd/workshed/importcmd"
	"github.com/frodi/workshed/cmd/workshed/inspect"
	"github.com/frodi/workshed/cmd/workshed/list"
	"github.com/frodi/workshed/cmd/workshed/path"
	"github.com/frodi/workshed/cmd/workshed/remove"
	"github.com/frodi/workshed/cmd/workshed/repos"
	"github.com/frodi/workshed/cmd/workshed/update"
	"github.com/frodi/workshed/internal/workspace"
)

func TestCapturesCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle", func(t *testing.T) {
		err := env.Run(captures.Command(), []string{ws.Handle})
		if err != nil {
			t.Errorf("captures with valid handle should succeed, got error: %v", err)
		}
		if strings.Contains(env.ErrorOutput(), "unknown command") {
			t.Errorf("captures should not produce unknown command error, stderr: %s", env.ErrorOutput())
		}
	})

	t.Run("with handle and filter", func(t *testing.T) {
		err := env.Run(captures.Command(), []string{ws.Handle, "--filter", "test"})
		if err != nil {
			t.Errorf("captures with filter should succeed, got error: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(captures.Command(), []string{"nonexistent-handle"})
		if err == nil {
			t.Error("captures with invalid handle should fail")
		}
		if !strings.Contains(env.ErrorOutput(), "workspace") {
			t.Errorf("captures with invalid handle should mention workspace, stderr: %s", env.ErrorOutput())
		}
	})
}

func TestPathCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle returns path", func(t *testing.T) {
		err := env.Run(path.Command(), []string{ws.Handle})
		if err != nil {
			t.Errorf("path with valid handle should succeed: %v", err)
		}
		if !strings.Contains(env.Output(), ws.Handle) {
			t.Errorf("path output should contain handle, got: %s", env.Output())
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(path.Command(), []string{"nonexistent"})
		if err == nil {
			t.Error("path with invalid handle should fail")
		}
		if !strings.Contains(env.ErrorOutput(), "workspace") {
			t.Errorf("path with invalid handle should mention workspace, stderr: %s", env.ErrorOutput())
		}
	})

	t.Run("format json", func(t *testing.T) {
		err := env.Run(path.Command(), []string{ws.Handle, "--format", "json"})
		if err != nil {
			t.Errorf("path --format json should work: %v", err)
		}
	})

	t.Run("format raw", func(t *testing.T) {
		err := env.Run(path.Command(), []string{ws.Handle, "--format", "raw"})
		if err != nil {
			t.Errorf("path --format raw should work: %v", err)
		}
		if !strings.Contains(env.Output(), ws.Path) {
			t.Errorf("raw format should contain path, got: %s", env.Output())
		}
	})
}

func TestInspectCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle", func(t *testing.T) {
		err := env.Run(inspect.Command(), []string{ws.Handle})
		if err != nil {
			t.Errorf("inspect with valid handle should succeed: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(inspect.Command(), []string{"nonexistent"})
		if err == nil {
			t.Error("inspect with invalid handle should fail")
		}
	})
}

func TestHealthCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle", func(t *testing.T) {
		err := env.Run(health.Command(), []string{ws.Handle})
		if err != nil {
			t.Errorf("health with valid handle should succeed: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(health.Command(), []string{"nonexistent"})
		if err == nil {
			t.Error("health with invalid handle should fail")
		}
	})
}

func TestExportCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle", func(t *testing.T) {
		err := env.Run(export.Command(), []string{ws.Handle})
		if err != nil {
			t.Errorf("export with valid handle should succeed: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(export.Command(), []string{"nonexistent"})
		if err == nil {
			t.Error("export with invalid handle should fail")
		}
	})

	t.Run("format json", func(t *testing.T) {
		err := env.Run(export.Command(), []string{ws.Handle, "--format", "json"})
		if err != nil {
			t.Errorf("export --format json should work: %v", err)
		}
		if !strings.Contains(env.Output(), `"handle"`) {
			t.Errorf("export json should contain handle, got: %s", env.Output())
		}
	})
}

func TestRemoveCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle and -y flag", func(t *testing.T) {
		err := env.Run(remove.Command(), []string{"-y", ws.Handle})
		if err != nil {
			t.Errorf("remove with valid handle should succeed: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(remove.Command(), []string{"-y", "nonexistent"})
		if err == nil {
			t.Error("remove with invalid handle should fail")
		}
	})

	t.Run("dry-run flag", func(t *testing.T) {
		ws := env.CreateWorkspace("dry-run test", nil)
		err := env.Run(remove.Command(), []string{"--dry-run", ws.Handle})
		if err != nil {
			t.Errorf("remove --dry-run should work: %v", err)
		}
	})
}

func TestUpdateCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("original purpose", nil)

	t.Run("with valid handle and --purpose", func(t *testing.T) {
		err := env.Run(update.Command(), []string{"--purpose", "new purpose", ws.Handle})
		if err != nil {
			t.Errorf("update with valid handle should succeed: %v", err)
		}
	})

	t.Run("with missing --purpose", func(t *testing.T) {
		err := env.Run(update.Command(), []string{ws.Handle})
		if err == nil {
			t.Error("update without --purpose should fail")
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(update.Command(), []string{"--purpose", "x", "nonexistent"})
		if err == nil {
			t.Error("update with invalid handle should fail")
		}
	})
}

func TestReposListCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle", func(t *testing.T) {
		err := env.Run(repos.ListCommand(), []string{ws.Handle})
		if err != nil {
			t.Errorf("repos list with valid handle should succeed: %v", err)
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(repos.ListCommand(), []string{"nonexistent"})
		if err == nil {
			t.Error("repos list with invalid handle should fail")
		}
	})
}

func TestReposAddCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle and --repo", func(t *testing.T) {
		newRepoDir := workspace.CreateLocalGitRepo(t, "newrepo", map[string]string{"file.txt": "content"})
		err := env.Run(repos.AddCommand(), []string{"--repo", newRepoDir, ws.Handle})
		if err != nil {
			t.Errorf("repos add should work: %v", err)
		}
	})

	t.Run("with missing --repo", func(t *testing.T) {
		err := env.Run(repos.AddCommand(), []string{ws.Handle})
		if err == nil {
			t.Error("repos add without --repo should fail")
		}
	})
}

func TestReposRemoveCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle and --repo", func(t *testing.T) {
		err := env.Run(repos.RemoveCommand(), []string{"--repo", "testrepo", ws.Handle})
		if err != nil {
			t.Errorf("repos remove should work: %v", err)
		}
	})

	t.Run("with missing --repo", func(t *testing.T) {
		err := env.Run(repos.RemoveCommand(), []string{ws.Handle})
		if err == nil {
			t.Error("repos remove without --repo should fail")
		}
	})
}

func TestApplyCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with valid handle and capture-id", func(t *testing.T) {
		err := env.Run(apply.Command(), []string{ws.Handle, "ws_nonexistent"})
		if err == nil {
			t.Error("apply with nonexistent capture should fail")
		}
		if !strings.Contains(env.ErrorOutput(), "not found") {
			t.Errorf("apply should mention capture not found, stderr: %s", env.ErrorOutput())
		}
	})

	t.Run("with invalid handle", func(t *testing.T) {
		err := env.Run(apply.Command(), []string{"nonexistent", "ws_id"})
		if err == nil {
			t.Error("apply with invalid handle should fail")
		}
	})
}

func TestCreateCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	t.Run("with --purpose", func(t *testing.T) {
		err := env.Run(create.Command(), []string{"--purpose", "my purpose"})
		if err != nil {
			t.Errorf("create with --purpose should work: %v", err)
		}
		output := env.Output()
		if output == "" {
			t.Errorf("create output should not be empty")
		}
	})

	t.Run("format json", func(t *testing.T) {
		err := env.Run(create.Command(), []string{"--purpose", "json format test", "--format", "json"})
		if err != nil {
			t.Errorf("create --format json should work: %v", err)
		}
		output := env.Output()
		if !strings.Contains(output, `"handle"`) {
			t.Errorf("create json should contain handle, got: %s", output)
		}
		if !strings.Contains(output, `"purpose"`) {
			t.Errorf("create json should contain purpose, got: %s", output)
		}
	})
}

func TestListCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	env.CreateWorkspace("test purpose", nil)
	env.CreateWorkspace("another purpose", nil)

	t.Run("list all workspaces", func(t *testing.T) {
		err := env.Run(list.Command(), []string{})
		if err != nil {
			t.Errorf("list should work: %v", err)
		}
	})

	t.Run("format json", func(t *testing.T) {
		err := env.Run(list.Command(), []string{"--format", "json"})
		if err != nil {
			t.Errorf("list --format json should work: %v", err)
		}
	})
}

func TestCaptureCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("with --name", func(t *testing.T) {
		err := env.Run(capture.Command(), []string{"--name", "test capture", ws.Handle})
		if err != nil {
			t.Errorf("capture with --name should work: %v", err)
		}
		output := env.Output()
		if output == "" {
			t.Errorf("capture output should not be empty")
		}
		if !strings.Contains(output, "test capture") {
			t.Errorf("capture output should contain name, got: %s", output)
		}
	})

	t.Run("format json", func(t *testing.T) {
		err := env.Run(capture.Command(), []string{"--name", "json capture", "--format", "json", ws.Handle})
		if err != nil {
			t.Errorf("capture --format json should work: %v", err)
		}
	})
}

func TestImportCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	t.Run("with --file flag", func(t *testing.T) {
		ws := env.CreateWorkspace("source workspace", nil)
		exportData, err := env.Store.ExportContext(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		tmpFile := filepath.Join(env.Root, "import.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err = env.Run(importcmd.Command(), []string{"--file", tmpFile})
		if err != nil {
			t.Errorf("import with --file should work: %v", err)
		}
	})

	t.Run("with positional arg", func(t *testing.T) {
		ws := env.CreateWorkspace("another workspace", nil)
		exportData, err := env.Store.ExportContext(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		tmpFile := filepath.Join(env.Root, "import2.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err = env.Run(importcmd.Command(), []string{tmpFile})
		if err != nil {
			t.Errorf("import with positional arg should work: %v", err)
		}
	})

	t.Run("with --preserve-handle", func(t *testing.T) {
		ws := env.CreateWorkspace("preserve handle test", nil)
		exportData, err := env.Store.ExportContext(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		tmpFile := filepath.Join(env.Root, "import_preserve.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err = env.Store.Remove(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		err = env.Run(importcmd.Command(), []string{"--file", tmpFile, "--preserve-handle"})
		if err != nil {
			t.Errorf("import with --preserve-handle should work: %v", err)
		}

		output := env.Output()
		if output == "" {
			t.Errorf("import output should not be empty")
		}
		if !strings.Contains(output, "handle") {
			t.Errorf("import should mention handle, got: %s", output)
		}
	})

	t.Run("with --force", func(t *testing.T) {
		ws := env.CreateWorkspace("force overwrite test", nil)
		exportData, err := env.Store.ExportContext(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		tmpFile := filepath.Join(env.Root, "import_force.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err = env.Run(importcmd.Command(), []string{"--file", tmpFile, "--force"})
		if err != nil {
			t.Errorf("import with --force should work: %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		err := env.Run(importcmd.Command(), []string{"--file", "/nonexistent/file.json"})
		if err == nil {
			t.Error("import with nonexistent file should fail")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpFile := filepath.Join(env.Root, "invalid.json")
		if err := os.WriteFile(tmpFile, []byte("not json {{{"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err := env.Run(importcmd.Command(), []string{"--file", tmpFile})
		if err == nil {
			t.Error("import with invalid json should fail")
		}
	})

	t.Run("missing purpose in json", func(t *testing.T) {
		ctx := &workspace.WorkspaceContext{
			Version:      1,
			Handle:       "test-handle",
			Purpose:      "",
			Repositories: []workspace.ContextRepo{{Name: "repo", URL: "https://example.com/repo"}},
		}
		jsonData, _ := json.MarshalIndent(ctx, "", "  ")
		tmpFile := filepath.Join(env.Root, "missing_purpose.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err := env.Run(importcmd.Command(), []string{"--file", tmpFile})
		if err == nil {
			t.Error("import with missing purpose should fail")
		}
	})

	t.Run("format json", func(t *testing.T) {
		ws := env.CreateWorkspace("format test", nil)
		exportData, err := env.Store.ExportContext(env.Ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		tmpFile := filepath.Join(env.Root, "format_json.json")
		if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		err = env.Run(importcmd.Command(), []string{"--file", tmpFile, "--format", "json"})
		if err != nil {
			t.Errorf("import --format json should work: %v", err)
		}
		output := env.Output()
		if !strings.Contains(output, `"handle"`) {
			t.Errorf("import json should contain handle, got: %s", output)
		}
	})
}
