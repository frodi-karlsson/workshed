//go:build integration
// +build integration

package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationCreate(t *testing.T) {
	t.Run("should maintain atomic behavior during creation", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"README.md": "# Test Repo"})

		t.Run("successful creation leaves no temp dirs", func(t *testing.T) {
			ctx := context.Background()
			opts := CreateOptions{
				Purpose: "Test workspace",
				Repositories: []RepositoryOption{
					{URL: repoURL, Ref: "main"},
				},
			}

			ws, err := store.Create(ctx, opts)
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}

			MustNotHaveTempDirs(t, root)
			MustHaveFile(t, ws.Path)
			MustHaveFile(t, filepath.Join(ws.Path, metadataFileName))
		})

		t.Run("invalid repo URL leaves no artifacts", func(t *testing.T) {
			beforeEntries, err := os.ReadDir(root)
			if err != nil {
				t.Fatalf("ReadDir failed: %v", err)
			}

			ctx := context.Background()
			opts := CreateOptions{
				Purpose: "Test workspace",
				Repositories: []RepositoryOption{
					{URL: "invalid-scheme://example.com/repo"},
				},
			}

			_, err = store.Create(ctx, opts)
			if err == nil {
				t.Fatal("Expected error for invalid repo URL")
			}

			afterEntries, err := os.ReadDir(root)
			if err != nil {
				t.Fatalf("ReadDir failed: %v", err)
			}

			if len(afterEntries) != len(beforeEntries) {
				t.Errorf("Directory count changed: before=%d, after=%d", len(beforeEntries), len(afterEntries))
			}

			MustNotHaveTempDirs(t, root)
		})

		t.Run("validation happens before filesystem operations", func(t *testing.T) {
			beforeEntries, err := os.ReadDir(root)
			if err != nil {
				t.Fatalf("ReadDir failed: %v", err)
			}

			ctx := context.Background()
			opts := CreateOptions{
				Purpose: "Test workspace",
				Repositories: []RepositoryOption{
					{URL: "ftp://invalid.com/repo"},
				},
			}

			_, err = store.Create(ctx, opts)
			if err == nil {
				t.Fatal("Expected error for unsupported URL scheme")
			}
			if !strings.Contains(err.Error(), "invalid repository URL") {
				t.Errorf("Expected 'invalid repository URL' error, got: %v", err)
			}

			afterEntries, err := os.ReadDir(root)
			if err != nil {
				t.Fatalf("ReadDir failed: %v", err)
			}

			if len(afterEntries) != len(beforeEntries) {
				t.Error("Directories created despite validation failure")
			}
		})
	})

	t.Run("should handle special characters in purpose", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"README.md": "# Test"})

		testCases := []struct {
			name    string
			purpose string
		}{
			{"should handle unicode characters", "Debug: payment flow with café and naïve users"},
			{"should handle numbers and symbols", "Test workspace #123 @home & away"},
			{"should handle quotes", `Workspace with "double" and 'single' quotes`},
			{"should handle slashes", "Path/with/slashes and back\\too"},
			{"should handle spaces and tabs", "Lots of    spaces   and	tabs"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ws, err := store.Create(ctx, CreateOptions{
					Purpose: tc.purpose,
					Repositories: []RepositoryOption{
						{URL: repoURL, Ref: "main"},
					},
				})
				if err != nil {
					t.Fatalf("Create failed for purpose %q: %v", tc.purpose, err)
				}

				retrieved, err := store.Get(ctx, ws.Handle)
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}

				if retrieved.Purpose != tc.purpose {
					t.Errorf("Purpose mismatch: got %q, want %q", retrieved.Purpose, tc.purpose)
				}
			})
		}
	})

	t.Run("should avoid collisions during concurrent creation", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"README.md": "# Test"})

		numWorkspaces := 10

		type result struct {
			ws  *Workspace
			err error
			idx int
		}

		results := make(chan result, numWorkspaces)

		for i := 0; i < numWorkspaces; i++ {
			go func(idx int) {
				ws, err := store.Create(ctx, CreateOptions{
					Purpose: "Concurrent workspace",
					Repositories: []RepositoryOption{
						{URL: repoURL, Ref: "main"},
					},
				})
				results <- result{ws, err, idx}
			}(i)
		}

		handles := make(map[string]int)
		for i := 0; i < numWorkspaces; i++ {
			r := <-results
			if r.err != nil {
				t.Errorf("Create failed on iteration %d: %v", r.idx, r.err)
				continue
			}

			if r.ws == nil {
				t.Errorf("Nil workspace on iteration %d", r.idx)
				continue
			}

			if handles[r.ws.Handle] > 0 {
				t.Errorf("Duplicate handle generated: %s", r.ws.Handle)
			}
			handles[r.ws.Handle]++
		}

		if len(handles) != numWorkspaces {
			t.Errorf("Expected %d unique handles, got %d", numWorkspaces, len(handles))
		}
	})

	t.Run("should create workspace with local repository", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Local repo test",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})

		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if ws == nil {
			t.Fatal("Expected workspace to be created")
		}

		if len(ws.Repositories) != 1 {
			t.Errorf("Expected 1 repository, got %d", len(ws.Repositories))
		}
	})
}

func TestIntegrationGet(t *testing.T) {
	t.Run("should handle corrupted metadata gracefully", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"README.md": "# Test"})

		ws := store.CreateMust(ctx, "Test workspace", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		metaPath := filepath.Join(ws.Path, metadataFileName)
		if err := os.WriteFile(metaPath, []byte("invalid json {{{"), 0644); err != nil {
			t.Fatalf("Failed to corrupt metadata: %v", err)
		}

		_, err := store.Get(ctx, ws.Handle)
		if err == nil {
			t.Error("Expected error when reading corrupted metadata")
		}
	})

	t.Run("should detect malformed workspace structure", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"README.md": "# Test"})

		ws := store.CreateMust(ctx, "Test workspace", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		if err := os.RemoveAll(ws.Path); err != nil {
			t.Fatalf("Failed to remove workspace directory: %v", err)
		}

		_, err := store.Get(ctx, ws.Handle)
		if err == nil {
			t.Error("Expected error when workspace directory is missing")
		}
	})
}

func TestIntegrationClone(t *testing.T) {
	t.Run("should clone local repository successfully", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "test content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Local clone test",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})

		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if ws == nil {
			t.Fatal("Expected workspace to be created")
		}

		repoPath, err := store.GetRepositoryPath(ctx, ws.Handle, "test-repo")
		if err != nil {
			t.Fatalf("GetRepositoryPath failed: %v", err)
		}
		if !FileExists(repoPath) {
			t.Errorf("Expected repository to be cloned at %s", repoPath)
		}

		expectedFile := filepath.Join(repoPath, "file.txt")
		if !FileExists(expectedFile) {
			t.Errorf("Expected cloned file to exist at %s", expectedFile)
		}
	})
}

func (s *FSStore) CreateMust(ctx context.Context, purpose string, repos []RepositoryOption) *Workspace {
	ws, err := s.Create(ctx, CreateOptions{
		Purpose:      purpose,
		Repositories: repos,
	})
	if err != nil {
		panic(err)
	}
	return ws
}

func TestIntegrationExec(t *testing.T) {
	t.Run("should execute command in all repositories", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo1", map[string]string{"file.txt": "content"})

		ws := store.CreateMust(ctx, "Test exec", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		results, err := store.Exec(ctx, ws.Handle, ExecOptions{
			Target:  "all",
			Command: []string{"pwd"},
		})
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("should execute command in specific repository", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo1", map[string]string{"file.txt": "content"})

		ws := store.CreateMust(ctx, "Test exec", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		results, err := store.Exec(ctx, ws.Handle, ExecOptions{
			Target:  "repo1",
			Command: []string{"pwd"},
		})
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Repository != "repo1" {
			t.Errorf("Expected repository 'repo1', got: %s", results[0].Repository)
		}
	})

	t.Run("should return error for nonexistent repository", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo1", map[string]string{"file.txt": "content"})

		ws := store.CreateMust(ctx, "Test exec", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		_, err := store.Exec(ctx, ws.Handle, ExecOptions{
			Target:  "nonexistent-repo",
			Command: []string{"pwd"},
		})
		if err == nil {
			t.Error("Expected error for nonexistent repository")
		}
		if !strings.Contains(err.Error(), "repository not found") {
			t.Errorf("Expected 'repository not found' error, got: %v", err)
		}
	})
}

func TestIntegrationRemove(t *testing.T) {
	t.Run("should remove workspace directory completely", func(t *testing.T) {
		store, root := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"file.txt": "content"})

		ws := store.CreateMust(ctx, "Test remove", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		workspacePath := filepath.Join(root, ws.Handle)
		if !FileExists(workspacePath) {
			t.Fatalf("Expected workspace directory to exist at %s", workspacePath)
		}

		err := store.Remove(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if FileExists(workspacePath) {
			t.Errorf("Expected workspace directory to be removed from %s", workspacePath)
		}
	})

	t.Run("should clean up cloned repositories", func(t *testing.T) {
		store, root := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "my-repo", map[string]string{"test.txt": "data"})

		ws := store.CreateMust(ctx, "Test repo cleanup", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		repoPath := filepath.Join(ws.Path, "my-repo")
		if !FileExists(repoPath) {
			t.Fatalf("Expected cloned repository at %s", repoPath)
		}

		err := store.Remove(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		workspacePath := filepath.Join(root, ws.Handle)
		if FileExists(workspacePath) {
			t.Errorf("Expected workspace directory to be removed")
		}
	})

	t.Run("should handle remove of nonexistent workspace", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		err := store.Remove(ctx, "nonexistent-workspace")
		if err == nil {
			t.Error("Expected error for nonexistent workspace")
		}
		if !strings.Contains(err.Error(), "workspace not found") {
			t.Errorf("Expected 'workspace not found' error, got: %v", err)
		}
	})

	t.Run("should handle remove of corrupted workspace directory", func(t *testing.T) {
		store, _ := CreateTestStore(t)
		ctx := context.Background()

		repoURL := CreateLocalGitRepo(t, "repo", map[string]string{"file.txt": "content"})

		ws := store.CreateMust(ctx, "Test corrupted remove", []RepositoryOption{
			{URL: repoURL, Ref: "main"},
		})

		if err := os.RemoveAll(ws.Path); err != nil {
			t.Fatalf("Failed to manually remove workspace: %v", err)
		}

		err := store.Remove(ctx, ws.Handle)
		if err == nil {
			t.Error("Expected error when workspace directory is missing")
		}
	})
}
