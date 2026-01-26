//go:build integration
// +build integration

package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/git"
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

	t.Run("should support --repos alias for --repo flag", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "alias-test", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Alias test", "--repos", repoURL})

		if env.ExitCalled() {
			t.Fatalf("Create with --repos alias exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should mention workspace created, got: %s", output)
		}
		if !strings.Contains(output, "alias-test") {
			t.Errorf("Create output should contain 'alias-test' repo name, got: %s", output)
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

func TestMockGitIntegration(t *testing.T) {
	t.Run("should list workspaces created via mocked git operations", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		mockGit := &git.MockGit{}
		mockGit.SetCurrentBranchResult("main")

		store, err := workspace.NewFSStore(env.TempDir, mockGit)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		env.Runner().Store = store

		ctx := context.Background()

		repoURL := filepath.Join(env.TempDir, "mock-repo")
		if err := os.MkdirAll(filepath.Join(repoURL, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create mock repo dir: %v", err)
		}

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Mocked git test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: ""},
			},
		})
		if err != nil {
			t.Fatalf("Create with mocked git failed: %v", err)
		}

		if ws == nil {
			t.Fatal("Expected workspace to be created with mocked git")
		}

		calls := mockGit.GetCurrentBranchCalls()
		if len(calls) == 0 {
			t.Error("Expected CurrentBranch to be called for local repo without ref")
		}

		env.ResetBuffers()
		env.Runner().List([]string{})

		output := env.Output()
		if !strings.Contains(output, "Mocked git test workspace") {
			t.Errorf("List output should contain mocked workspace purpose, got: %s", output)
		}
	})

	t.Run("should handle checkout errors gracefully", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		mockGit := &git.MockGit{}
		mockGit.SetCheckoutErr(&git.GitError{Operation: "checkout", Hint: "ref not found", Details: "reference not found"})

		store, err := workspace.NewFSStore(env.TempDir, mockGit)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "error-test", map[string]string{"file.txt": "content"})

		_, err = store.Create(ctx, workspace.CreateOptions{
			Purpose: "Error test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "nonexistent-branch"},
			},
		})
		if err == nil {
			t.Error("Create should fail when checkout fails")
		}

		output := err.Error()
		if !strings.Contains(output, "checkout") && !strings.Contains(output, "ref not found") {
			t.Errorf("Error should mention checkout failure, got: %s", output)
		}
	})

	t.Run("should track clone calls for verification", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		mockGit := &git.MockGit{}
		mockGit.SetCurrentBranchResult("main")

		store, err := workspace.NewFSStore(env.TempDir, mockGit)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "clone-test", map[string]string{"file.txt": "content"})

		_, err = store.Create(ctx, workspace.CreateOptions{
			Purpose: "Clone tracking test",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		calls := mockGit.GetCloneCalls()
		if len(calls) != 1 {
			t.Errorf("Expected 1 clone call, got: %d", len(calls))
		}
		if len(calls) > 0 && calls[0].URL != repoURL {
			t.Errorf("Expected clone URL %s, got: %s", repoURL, calls[0].URL)
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

	t.Run("should remove workspace with confirmation 'y'", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace to remove",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("y\n")
		env.ResetBuffers()
		env.Runner().Remove([]string{ws.Handle})

		if env.ExitCalled() {
			t.Fatalf("Remove exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "removed") {
			t.Errorf("Remove output should mention workspace removed, got: %s", output)
		}

		_, err = store.Get(ctx, ws.Handle)
		if err == nil {
			t.Error("Workspace should have been removed")
		}
	})

	t.Run("should not remove workspace with confirmation 'n'", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace to keep",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("n\n")
		env.ResetBuffers()
		env.Runner().Remove([]string{ws.Handle})

		if env.ExitCalled() {
			t.Fatalf("Remove exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "cancelled") {
			t.Errorf("Remove output should mention operation cancelled, got: %s", output)
		}

		_, err = store.Get(ctx, ws.Handle)
		if err != nil {
			t.Error("Workspace should not have been removed")
		}
	})

	t.Run("should not remove workspace with confirmation 'yes' (lowercase)", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace with yes confirmation",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("yes\n")
		env.ResetBuffers()
		env.Runner().Remove([]string{ws.Handle})

		if env.ExitCalled() {
			t.Fatalf("Remove exited unexpectedly: %s", env.ErrorOutput())
		}

		_, err = store.Get(ctx, ws.Handle)
		if err == nil {
			t.Error("Workspace should have been removed")
		}
	})

	t.Run("should handle empty stdin gracefully", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace no stdin",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("")
		env.ResetBuffers()
		env.Runner().Remove([]string{ws.Handle})

		if !env.ExitCalled() {
			t.Error("Remove should exit when stdin read fails")
		}

		output := env.Output()
		if !strings.Contains(output, "input") && !strings.Contains(output, "read") {
			t.Errorf("Remove output should mention input read error, got: %s", output)
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

func TestUpdate(t *testing.T) {
	t.Run("should fail cleanly for nonexistent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Update([]string{"--purpose", "New purpose", "nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Update should exit with error for nonexistent workspace")
		}

		output := env.Output()
		if !strings.Contains(output, "workspace") && !strings.Contains(output, "purpose") {
			t.Errorf("Update output should mention workspace and purpose, got: %s", output)
		}
	})

	t.Run("should fail when --purpose flag is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Original purpose",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Update([]string{ws.Handle})

		if !env.ExitCalled() {
			t.Error("Update should exit with error when --purpose is missing")
		}

		output := env.Output()
		if !strings.Contains(output, "purpose") || !strings.Contains(output, "flag") {
			t.Errorf("Update output should mention missing --purpose flag, got: %s", output)
		}
	})

	t.Run("should successfully update workspace purpose", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Original purpose",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Update([]string{"--purpose", "Updated purpose", ws.Handle})

		if env.ExitCalled() {
			t.Fatalf("Update exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "purpose updated") {
			t.Errorf("Update output should mention purpose updated, got: %s", output)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if retrieved.Purpose != "Updated purpose" {
			t.Errorf("Expected purpose 'Updated purpose', got: %s", retrieved.Purpose)
		}
	})

	t.Run("should handle handle-less update by finding workspace in current directory", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Original purpose",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		if err := os.Chdir(ws.Path); err != nil {
			t.Fatalf("Failed to change to workspace directory: %v", err)
		}
		defer os.Chdir("/")

		env.ResetBuffers()
		env.Runner().Update([]string{"--purpose", "Auto-found workspace purpose"})

		if env.ExitCalled() {
			t.Fatalf("Update exited unexpectedly: %s", env.ErrorOutput())
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if retrieved.Purpose != "Auto-found workspace purpose" {
			t.Errorf("Expected purpose 'Auto-found workspace purpose', got: %s", retrieved.Purpose)
		}
	})

	t.Run("should fail with empty purpose", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Original purpose",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Update([]string{"--purpose", "", ws.Handle})

		if !env.ExitCalled() {
			t.Error("Update should exit with error for empty purpose")
		}
	})
}

func TestRepoAdd(t *testing.T) {
	t.Run("should add repository to existing workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL1 := workspace.CreateLocalGitRepo(t, "existing-repo", map[string]string{"file.txt": "content"})
		repoURL2 := workspace.CreateLocalGitRepo(t, "new-repo", map[string]string{"new.txt": "new content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL1, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{ws.Handle, "--repo", repoURL2})

		if env.ExitCalled() {
			t.Fatalf("RepoAdd exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "repository added") {
			t.Errorf("RepoAdd output should contain 'repository added', got: %s", output)
		}
		if !strings.Contains(output, "new-repo") {
			t.Errorf("RepoAdd output should contain repo name, got: %s", output)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 2 {
			t.Errorf("Expected 2 repositories, got: %d", len(retrieved.Repositories))
		}
		found := false
		for _, repo := range retrieved.Repositories {
			if repo.Name == "new-repo" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Added repository not found in workspace")
		}
	})

	t.Run("should add multiple repositories at once", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL1 := workspace.CreateLocalGitRepo(t, "repo1", map[string]string{"file.txt": "content"})
		repoURL2 := workspace.CreateLocalGitRepo(t, "repo2", map[string]string{"file.txt": "content"})
		repoURL3 := workspace.CreateLocalGitRepo(t, "repo3", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Multi-add test",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL1, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{ws.Handle, "--repo", repoURL2, "--repo", repoURL3})

		if env.ExitCalled() {
			t.Fatalf("RepoAdd exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "repositories added") {
			t.Errorf("RepoAdd output should contain 'repositories added', got: %s", output)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 3 {
			t.Errorf("Expected 3 repositories, got: %d", len(retrieved.Repositories))
		}
	})

	t.Run("should fail with missing workspace handle", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{"--repo", repoURL})

		if !env.ExitCalled() {
			t.Error("RepoAdd should exit with error when workspace handle is missing")
		}
	})

	t.Run("should fail with missing --repo flag", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{ws.Handle})

		if !env.ExitCalled() {
			t.Error("RepoAdd should exit with error when --repo is missing")
		}
	})

	t.Run("should fail with duplicate repository", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{ws.Handle, "--repo", repoURL})

		if !env.ExitCalled() {
			t.Error("RepoAdd should exit with error for duplicate repository")
		}

		output := env.Output()
		if !strings.Contains(output, "failed to add repository") {
			t.Errorf("RepoAdd output should mention failure, got: %s", output)
		}
	})

	t.Run("should fail for non-existent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().RepoAdd([]string{"non-existent-workspace", "--repo", repoURL})

		if !env.ExitCalled() {
			t.Error("RepoAdd should exit with error for non-existent workspace")
		}

		output := env.Output()
		if !strings.Contains(output, "failed to add repository") {
			t.Errorf("RepoAdd output should mention failure, got: %s", output)
		}
	})
}

func TestRepoRemove(t *testing.T) {
	t.Run("should remove repository from workspace with --force", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL1 := workspace.CreateLocalGitRepo(t, "repo-to-keep", map[string]string{"keep.txt": "keep"})
		repoURL2 := workspace.CreateLocalGitRepo(t, "repo-to-remove", map[string]string{"remove.txt": "remove"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL1, Ref: "main"},
				{URL: repoURL2, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--repo", "repo-to-remove", "--force"})

		if env.ExitCalled() {
			t.Fatalf("RepoRemove exited unexpectedly: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "repository removed") {
			t.Errorf("RepoRemove output should contain 'repository removed', got: %s", output)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 1 {
			t.Errorf("Expected 1 repository, got: %d", len(retrieved.Repositories))
		}
		found := false
		for _, repo := range retrieved.Repositories {
			if repo.Name == "repo-to-remove" {
				t.Error("Removed repository still exists in workspace")
			}
			if repo.Name == "repo-to-keep" {
				found = true
			}
		}
		if !found {
			t.Error("Kept repository not found in workspace")
		}
	})

	t.Run("should require --force and confirm with y", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "repo-to-remove", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("y\n")
		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--repo", "repo-to-remove"})

		if env.ExitCalled() {
			t.Fatalf("RepoRemove exited unexpectedly: %s", env.ErrorOutput())
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 0 {
			t.Errorf("Expected 0 repositories, got: %d", len(retrieved.Repositories))
		}
	})

	t.Run("should cancel when user declines confirmation", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "protected-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.SetStdin("n\n")
		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--repo", "protected-repo"})

		if env.ExitCalled() {
			t.Fatalf("RepoRemove exited unexpectedly: %s", env.ErrorOutput())
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 1 {
			t.Errorf("Repository should not have been removed, got: %d", len(retrieved.Repositories))
		}

		output := env.Output()
		if !strings.Contains(output, "cancelled") {
			t.Errorf("RepoRemove output should mention 'cancelled', got: %s", output)
		}
	})

	t.Run("should fail with missing workspace handle", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{"--repo", "some-repo", "--force"})

		if !env.ExitCalled() {
			t.Error("RepoRemove should exit with error when workspace handle is missing")
		}
	})

	t.Run("should fail with missing --repo flag", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--force"})

		if !env.ExitCalled() {
			t.Error("RepoRemove should exit with error when --repo is missing")
		}
	})

	t.Run("should fail for non-existent repository", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Test workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--repo", "non-existent-repo", "--force"})

		if !env.ExitCalled() {
			t.Error("RepoRemove should exit with error for non-existent repository")
		}

		output := env.Output()
		if !strings.Contains(output, "repository not found") {
			t.Errorf("RepoRemove output should mention 'repository not found', got: %s", output)
		}
	})

	t.Run("should fail for non-existent workspace", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{"non-existent-workspace", "--repo", "some-repo", "--force"})

		if !env.ExitCalled() {
			t.Error("RepoRemove should exit with error for non-existent workspace")
		}
	})

	t.Run("should remove last repository", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		repoURL := workspace.CreateLocalGitRepo(t, "only-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, workspace.CreateOptions{
			Purpose: "Single repo workspace",
			Repositories: []workspace.RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create workspace: %v", err)
		}

		env.ResetBuffers()
		env.Runner().RepoRemove([]string{ws.Handle, "--repo", "only-repo", "--force"})

		if env.ExitCalled() {
			t.Fatalf("RepoRemove exited unexpectedly: %s", env.ErrorOutput())
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}
		if len(retrieved.Repositories) != 0 {
			t.Errorf("Expected 0 repositories, got: %d", len(retrieved.Repositories))
		}
	})
}

func TestCreateWithTemplate(t *testing.T) {
	t.Run("should create workspace with template", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		templateDir := filepath.Join(env.TempDir, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "config.json")
		if err := os.WriteFile(templateFile, []byte(`{"key": "value"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Template test", "--template", templateDir, "--repo", repoURL})
		if env.ExitCalled() {
			t.Fatalf("Create exited: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}

		handle := ExtractHandleFromLog(t, output)

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Get(ctx, handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}

		configPath := filepath.Join(ws.Path, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Expected config.json from template to exist")
		}
	})

	t.Run("should create workspace with template and variables", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		templateDir := filepath.Join(env.TempDir, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "{{env}}.json")
		if err := os.WriteFile(templateFile, []byte(`{"env": "{{env}}"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Create([]string{
			"--purpose", "Template var test",
			"--template", templateDir,
			"--map", "env=production",
			"--repo", repoURL,
		})
		if env.ExitCalled() {
			t.Fatalf("Create exited: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}

		handle := ExtractHandleFromLog(t, output)

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Get(ctx, handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}

		expectedFile := filepath.Join(ws.Path, "production.json")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Error("Expected production.json from variable substitution")
		}
	})

	t.Run("should support multiple --map flags", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		templateDir := filepath.Join(env.TempDir, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "{{env}}-{{region}}.txt")
		if err := os.WriteFile(templateFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Create([]string{
			"--purpose", "Multi-var test",
			"--template", templateDir,
			"--map", "env=staging",
			"--map", "region=us-west",
			"--repo", repoURL,
		})
		if env.ExitCalled() {
			t.Fatalf("Create exited: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}

		handle := ExtractHandleFromLog(t, output)

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Get(ctx, handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}

		expectedFile := filepath.Join(ws.Path, "staging-us-west.txt")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Error("Expected staging-us-west.txt from variable substitution")
		}
	})

	t.Run("should fail for non-existent template", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().Create([]string{
			"--purpose", "Invalid template test",
			"--template", "/nonexistent/path",
			"--repo", repoURL,
		})

		if !env.ExitCalled() {
			t.Error("Create should have exited with error")
		}

		errOutput := env.Output()
		if !strings.Contains(errOutput, "workspace creation failed") {
			t.Errorf("Error output should mention 'workspace creation failed', got: %s", errOutput)
		}
	})

	t.Run("should fail for invalid map format", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		templateDir := filepath.Join(env.TempDir, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		env.ResetBuffers()
		env.Runner().Create([]string{
			"--purpose", "Invalid map test",
			"--template", templateDir,
			"--map", "novalue",
			"--repo", repoURL,
		})

		if !env.ExitCalled() {
			t.Error("Create should have exited with error")
		}

		errOutput := env.Output()
		if !strings.Contains(errOutput, "invalid template variable") {
			t.Errorf("Error output should mention 'invalid template variable', got: %s", errOutput)
		}
	})

	t.Run("should create workspace with template and repos", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		templateDir := filepath.Join(env.TempDir, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "README.md")
		if err := os.WriteFile(templateFile, []byte("# Template README"), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		repoURL := workspace.CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		env.ResetBuffers()
		env.Runner().Create([]string{
			"--purpose", "Template and repo test",
			"--template", templateDir,
			"--repo", repoURL,
		})
		if env.ExitCalled() {
			t.Fatalf("Create exited: %s", env.ErrorOutput())
		}

		output := env.Output()
		if !strings.Contains(output, "workspace created") {
			t.Errorf("Create output should contain 'workspace created', got: %s", output)
		}

		handle := ExtractHandleFromLog(t, output)

		store, err := workspace.NewFSStore(env.TempDir)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Get(ctx, handle)
		if err != nil {
			t.Fatalf("Failed to get workspace: %v", err)
		}

		if len(ws.Repositories) != 1 {
			t.Errorf("Expected 1 repository, got: %d", len(ws.Repositories))
		}

		readmePath := filepath.Join(ws.Path, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Error("Expected README.md from template to exist")
		}

		repoPath := filepath.Join(ws.Path, "test-repo")
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			t.Error("Expected test-repo to exist")
		}
	})
}
