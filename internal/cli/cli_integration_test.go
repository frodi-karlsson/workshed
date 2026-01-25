//go:build integration
// +build integration

package cli

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/workspace"
)

func TestCreate(t *testing.T) {
	t.Run("should execute complete workspace lifecycle", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Test workspace", "--repo", repoURL})
		if env.ExitCalled() {
			t.Fatalf("Create exited: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}

		handle := ExtractHandleFromLog(t, output)

		env.ResetBuffers()
		env.Runner().List([]string{})
		if env.ExitCalled() {
			t.Fatalf("List exited: %s", env.ErrorOutput())
		}

		output = env.Output()
		if !strings.Contains(output, handle) {
			t.Errorf("List output should contain handle %s, got: %s", handle, output)
		}
		if !strings.Contains(output, "Test workspace") {
			t.Errorf("List output should contain purpose, got: %s", output)
		}

		env.ResetBuffers()
		env.Runner().Inspect([]string{handle})
		if env.ExitCalled() {
			t.Fatalf("Inspect exited: %s", env.ErrorOutput())
		}

		output = env.Output()
		if !strings.Contains(output, handle) {
			t.Errorf("Inspect output should contain handle, got: %s", output)
		}
		if !strings.Contains(output, "Test workspace") {
			t.Errorf("Inspect output should contain purpose, got: %s", output)
		}
		if !strings.Contains(output, "test-repo") {
			t.Errorf("Inspect output should contain repository name, got: %s", output)
		}

		env.ResetBuffers()
		env.Runner().Remove([]string{"--force", handle})
		if env.ExitCalled() {
			t.Fatalf("Remove exited: %s", env.ErrorOutput())
		}

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.Get(ctx, handle)
		if err == nil {
			t.Error("Workspace should have been removed")
		}
	})

	t.Run("should exit with error for invalid repo URL", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Test", "--repo", "/nonexistent/local/repo"})

		if !env.ExitCalled() {
			t.Error("Create should have exited with error")
		}

		errOutput := env.Output()
		if !strings.Contains(errOutput, "workspace creation failed") {
			t.Errorf("Error output should mention workspace creation failed, got: %s", errOutput)
		}
	})

	t.Run("should create workspace with local repository", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "local-test", map[string]string{"README.md": "# Test"})

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Local repo test", "--repo", repoURL})

		if env.ExitCalled() {
			t.Fatalf("Create exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}
		if !strings.Contains(output, "local-test") {
			t.Errorf("Create output should contain repository name, got: %s", output)
		}
	})

	t.Run("should exit with error in read-only directory", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		// Make directory read-only
		if err := os.Chmod(env.TempDir, 0555); err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}
		// Restore permissions on cleanup so t.TempDir() can remove it
		defer os.Chmod(env.TempDir, 0755)

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Test workspace", "--repo", repoURL})

		if !env.ExitCalled() {
			t.Error("Create should exit with error in read-only directory")
		}
	})

	t.Run("should handle special characters in purpose", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		purpose := "Debug: payment flow with café and naïve users"
		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", purpose, "--repo", repoURL})

		if env.ExitCalled() {
			t.Fatalf("Create exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should mention workspace created, got: %s", output)
		}

		env.ResetBuffers()
		env.Runner().List([]string{})

		listOutput := env.Output()
		if !strings.Contains(listOutput, purpose) {
			t.Errorf("List output should contain purpose with special chars, got: %s", listOutput)
		}
	})

	t.Run("should create workspace with multiple repositories", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL1 := workspace.CreateLocalGitRepo(t, "api", map[string]string{"file.txt": "api content"})
		repoURL2 := workspace.CreateLocalGitRepo(t, "worker", map[string]string{"file.txt": "worker content"})

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Multi-repo test", "--repo", repoURL1, "--repo", repoURL2})

		if env.ExitCalled() {
			t.Fatalf("Create exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should mention workspace created, got: %s", output)
		}
		if !strings.Contains(output, "api") {
			t.Errorf("Create output should contain 'api' repo name, got: %s", output)
		}
		if !strings.Contains(output, "worker") {
			t.Errorf("Create output should contain 'worker' repo name, got: %s", output)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("should filter workspaces by purpose", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		_, err = store.Create(ctx, workspace.CreateOptions{
			Purpose: "Debug payment flow",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		_, err = store.Create(ctx, workspace.CreateOptions{
			Purpose: "Add login feature",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		_, err = store.Create(ctx, workspace.CreateOptions{
			Purpose: "Debug checkout bug",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().List([]string{})
		output := env.Output()
		if strings.Count(output, "Debug") != 2 {
			t.Errorf("Should show 2 debug workspaces, got: %s", output)
		}

		env.ResetBuffers()
		env.Runner().List([]string{"--purpose", "debug"})
		output = env.Output()
		if !strings.Contains(output, "payment") {
			t.Errorf("Filtered list should contain 'payment', got: %s", output)
		}
		if !strings.Contains(output, "checkout") {
			t.Errorf("Filtered list should contain 'checkout', got: %s", output)
		}
		if strings.Contains(output, "login") {
			t.Errorf("Filtered list should not contain 'login', got: %s", output)
		}
	})

	t.Run("should handle empty directory gracefully", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().List([]string{})

		output := env.Output()
		if !strings.Contains(output, "no workspaces") {
			t.Errorf("List should mention no workspaces found, got: %s", output)
		}
	})
}

func TestRemove(t *testing.T) {
	t.Run("should exit with error for nonexistent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Remove([]string{"--force", "nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Remove should exit with error for nonexistent workspace")
		}

		output := env.Output()
		if !strings.Contains(output, "not found") && !strings.Contains(output, "failed to get") {
			t.Errorf("Output should mention workspace not found or failed to get, got: %s", output)
		}
	})
}

func TestInspect(t *testing.T) {
	t.Run("should fail cleanly for nonexistent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Inspect([]string{"nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Inspect should exit with error for nonexistent workspace")
		}

		output := env.Output()
		if !strings.Contains(output, "not found") && !strings.Contains(output, "failed to get") {
			t.Errorf("Inspect output should mention workspace not found or failed to get, got: %s", output)
		}
	})
}

func TestPath(t *testing.T) {
	t.Run("should fail cleanly for nonexistent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Path([]string{"nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Path should exit with error for nonexistent workspace")
		}

		output := env.Output()
		if !strings.Contains(output, "not found") && !strings.Contains(output, "failed to get") {
			t.Errorf("Path output should mention workspace not found or failed to get, got: %s", output)
		}
	})
}
