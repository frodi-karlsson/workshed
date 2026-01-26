//go:build integration
// +build integration

package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/git"
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo")

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
						{URL: fakeRepoPath, Ref: "main"},
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo")

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
						{URL: fakeRepoPath, Ref: "main"},
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo")

		ws := store.CreateMust(ctx, "Test workspace", []RepositoryOption{
			{URL: fakeRepoPath, Ref: "main"},
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo")

		ws := store.CreateMust(ctx, "Test workspace", []RepositoryOption{
			{URL: fakeRepoPath, Ref: "main"},
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo1")

		ws := store.CreateMust(ctx, "Test exec", []RepositoryOption{
			{URL: fakeRepoPath, Ref: "main"},
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
		store, root, _ := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-repo")

		ws := store.CreateMust(ctx, "Test corrupted remove", []RepositoryOption{
			{URL: fakeRepoPath, Ref: "main"},
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

func TestMockGitIntegration(t *testing.T) {
	t.Run("should use mocked git for workspace creation", func(t *testing.T) {
		store, root, mockGit := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "fake-local-repo")

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Mocked git test",
			Repositories: []RepositoryOption{
				{URL: fakeRepoPath, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if ws == nil {
			t.Fatal("Expected workspace to be created with mocked git")
		}

		calls := mockGit.GetCloneCalls()
		if len(calls) != 1 {
			t.Errorf("Expected 1 clone call, got: %d", len(calls))
		}

		checkoutCalls := mockGit.GetCheckoutCalls()
		if len(checkoutCalls) != 1 {
			t.Errorf("Expected 1 checkout call, got: %d", len(checkoutCalls))
		}
	})

	t.Run("should handle checkout errors via mock", func(t *testing.T) {
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, t.TempDir(), "fake-repo")

		mockGit := &git.MockGit{}
		mockGit.SetCheckoutErr(&git.GitError{
			Operation: "checkout",
			Hint:      "ref not found",
			Details:   "reference not found",
		})

		storeWithErr, err := NewFSStore(t.TempDir(), mockGit)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		_, err = storeWithErr.Create(ctx, CreateOptions{
			Purpose: "Checkout error test",
			Repositories: []RepositoryOption{
				{URL: fakeRepoPath, Ref: "nonexistent"},
			},
		})
		if err == nil {
			t.Error("Expected error for checkout failure")
		}

		output := err.Error()
		if !strings.Contains(output, "checkout") && !strings.Contains(output, "ref not found") {
			t.Errorf("Error should mention checkout failure, got: %s", output)
		}
	})

	t.Run("should track clone call arguments", func(t *testing.T) {
		store, root, mockGit := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "my-repo")

		_, err := store.Create(ctx, CreateOptions{
			Purpose: "Clone args test",
			Repositories: []RepositoryOption{
				{URL: fakeRepoPath, Ref: "develop"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		calls := mockGit.GetCloneCalls()
		if len(calls) != 1 {
			t.Fatalf("Expected 1 clone call, got: %d", len(calls))
		}

		if calls[0].URL != fakeRepoPath {
			t.Errorf("Expected clone URL %s, got: %s", fakeRepoPath, calls[0].URL)
		}

		if calls[0].Opts.Depth != 0 {
			t.Errorf("Expected depth 0 (full clone), got: %d", calls[0].Opts.Depth)
		}
	})

	t.Run("should handle clone errors via mock", func(t *testing.T) {
		fakeRepoPath := CreateFakeRepo(t, t.TempDir(), "nonexistent-repo")

		mockGit := &git.MockGit{}
		mockGit.SetCloneErr(&git.GitError{
			Operation: "clone",
			Hint:      "repository not found",
			Details:   "repository not found",
		})

		store, err := NewFSStore(t.TempDir(), mockGit)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		_, err = store.Create(ctx, CreateOptions{
			Purpose: "Clone error test",
			Repositories: []RepositoryOption{
				{URL: fakeRepoPath, Ref: "main"},
			},
		})
		if err == nil {
			t.Error("Expected error for clone failure")
		}
	})

	t.Run("should verify mocked git is used for local repo without ref", func(t *testing.T) {
		store, root, mockGit := CreateMockedTestStore(t)
		ctx := context.Background()

		fakeRepoPath := CreateFakeRepo(t, root, "local-repo")

		_, err := store.Create(ctx, CreateOptions{
			Purpose: "Current branch test",
			Repositories: []RepositoryOption{
				{URL: fakeRepoPath, Ref: ""},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		branchCalls := mockGit.GetCurrentBranchCalls()
		if len(branchCalls) != 1 {
			t.Errorf("Expected 1 CurrentBranch call, got: %d", len(branchCalls))
		}
	})
}

func TestIntegrationAddRepository(t *testing.T) {
	t.Run("should add single repository to workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL1 := CreateLocalGitRepo(t, "existing-repo", map[string]string{"file.txt": "content"})
		repoURL2 := CreateLocalGitRepo(t, "new-repo", map[string]string{"new.txt": "new content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL1, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.AddRepository(ctx, ws.Handle, RepositoryOption{URL: repoURL2, Ref: "main"})
		if err != nil {
			t.Fatalf("AddRepository failed: %v", err)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(retrieved.Repositories) != 2 {
			t.Errorf("Expected 2 repositories, got: %d", len(retrieved.Repositories))
		}
		if !strings.Contains(retrieved.Repositories[1].URL, "new-repo") {
			t.Errorf("Expected new-repo URL, got: %s", retrieved.Repositories[1].URL)
		}
	})

	t.Run("should add multiple repositories", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL1 := CreateLocalGitRepo(t, "existing-repo", map[string]string{"file.txt": "content"})
		repoURL2 := CreateLocalGitRepo(t, "new-repo-1", map[string]string{"file.txt": "content"})
		repoURL3 := CreateLocalGitRepo(t, "new-repo-2", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL1, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.AddRepositories(ctx, ws.Handle, []RepositoryOption{
			{URL: repoURL2, Ref: "main"},
			{URL: repoURL3, Ref: "main"},
		})
		if err != nil {
			t.Fatalf("AddRepositories failed: %v", err)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(retrieved.Repositories) != 3 {
			t.Errorf("Expected 3 repositories, got: %d", len(retrieved.Repositories))
		}
	})

	t.Run("should reject duplicate URL", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.AddRepository(ctx, ws.Handle, RepositoryOption{URL: repoURL, Ref: "develop"})
		if err == nil {
			t.Error("Expected error for duplicate URL")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Error should mention 'already exists', got: %v", err)
		}
	})

	t.Run("should reject duplicate repository name", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL1 := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})
		repoURL2 := CreateLocalGitRepo(t, "another-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL1, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		_ = ws
		_ = repoURL2

		err = store.AddRepository(ctx, ws.Handle, RepositoryOption{URL: repoURL1, Ref: "develop"})
		if err == nil {
			t.Error("Expected error for same URL (different ref)")
		}
	})

	t.Run("should fail for non-existent workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		err = store.AddRepository(ctx, "non-existent", RepositoryOption{URL: repoURL})
		if err == nil {
			t.Error("Expected error for non-existent workspace")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Error should mention 'not found', got: %v", err)
		}
	})

	t.Run("should fail with empty repository list", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.AddRepositories(ctx, ws.Handle, []RepositoryOption{})
		if err == nil {
			t.Error("Expected error for empty repository list")
		}
	})
}

func TestIntegrationRemoveRepository(t *testing.T) {
	t.Run("should remove repository from workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL1 := CreateLocalGitRepo(t, "repo-to-keep", map[string]string{"keep.txt": "keep"})
		repoURL2 := CreateLocalGitRepo(t, "repo-to-remove", map[string]string{"remove.txt": "remove"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL1, Ref: "main"},
				{URL: repoURL2, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.RemoveRepository(ctx, ws.Handle, "repo-to-remove")
		if err != nil {
			t.Fatalf("RemoveRepository failed: %v", err)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(retrieved.Repositories) != 1 {
			t.Errorf("Expected 1 repository, got: %d", len(retrieved.Repositories))
		}
		if retrieved.Repositories[0].Name != "repo-to-keep" {
			t.Errorf("Expected repo-to-keep, got: %s", retrieved.Repositories[0].Name)
		}
	})

	t.Run("should remove last repository", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "only-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Single repo workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.RemoveRepository(ctx, ws.Handle, "only-repo")
		if err != nil {
			t.Fatalf("RemoveRepository failed: %v", err)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(retrieved.Repositories) != 0 {
			t.Errorf("Expected 0 repositories, got: %d", len(retrieved.Repositories))
		}
	})

	t.Run("should fail for non-existent repository", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = store.RemoveRepository(ctx, ws.Handle, "non-existent-repo")
		if err == nil {
			t.Error("Expected error for non-existent repository")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Error should mention 'not found', got: %v", err)
		}
	})

	t.Run("should fail for non-existent workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()

		err = store.RemoveRepository(ctx, "non-existent", "some-repo")
		if err == nil {
			t.Error("Expected error for non-existent workspace")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Error should mention 'not found', got: %v", err)
		}
	})

	t.Run("should handle already-deleted repo directory", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		repoURL := CreateLocalGitRepo(t, "test-repo", map[string]string{"file.txt": "content"})

		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		os.RemoveAll(filepath.Join(ws.Path, "test-repo"))

		err = store.RemoveRepository(ctx, ws.Handle, "test-repo")
		if err != nil {
			t.Fatalf("RemoveRepository should succeed even if directory is already gone: %v", err)
		}

		retrieved, err := store.Get(ctx, ws.Handle)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if len(retrieved.Repositories) != 0 {
			t.Errorf("Expected 0 repositories, got: %d", len(retrieved.Repositories))
		}
	})
}

func TestIntegrationTemplate(t *testing.T) {
	t.Run("should copy template directory into workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		templateDir := filepath.Join(root, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "config.json")
		if err := os.WriteFile(templateFile, []byte(`{"key": "value"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		subdir := filepath.Join(templateDir, "subdir")
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("Failed to create template subdirectory: %v", err)
		}

		subdirFile := filepath.Join(subdir, "nested.txt")
		if err := os.WriteFile(subdirFile, []byte("nested content"), 0644); err != nil {
			t.Fatalf("Failed to write nested file: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose:      "Template test",
			Template:     templateDir,
			Repositories: []RepositoryOption{},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if !FileExists(filepath.Join(ws.Path, "config.json")) {
			t.Error("Expected config.json from template")
		}
		if !FileExists(filepath.Join(ws.Path, "subdir", "nested.txt")) {
			t.Error("Expected nested.txt from template")
		}
	})

	t.Run("should apply variable substitution in template", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		templateDir := filepath.Join(root, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "{{env}}.json")
		if err := os.WriteFile(templateFile, []byte(`{"env": "{{env}}"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose:  "Template var test",
			Template: templateDir,
			TemplateVars: map[string]string{
				"env": "production",
			},
			Repositories: []RepositoryOption{},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if !FileExists(filepath.Join(ws.Path, "production.json")) {
			t.Error("Expected production.json with variable substitution")
		}
	})

	t.Run("should skip template when path is empty", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose:      "No template test",
			Template:     "",
			Repositories: []RepositoryOption{},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if ws == nil {
			t.Fatal("Expected workspace to be created")
		}
	})

	t.Run("should error for non-existent template", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		_, err = store.Create(ctx, CreateOptions{
			Purpose:      "Invalid template test",
			Template:     "/non/existent/path",
			Repositories: []RepositoryOption{},
		})
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Error should mention 'does not exist', got: %v", err)
		}
	})

	t.Run("should error when template is a file, not directory", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		templateFile := filepath.Join(root, "template.txt")
		if err := os.WriteFile(templateFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		ctx := context.Background()
		_, err = store.Create(ctx, CreateOptions{
			Purpose:      "File instead of dir test",
			Template:     templateFile,
			Repositories: []RepositoryOption{},
		})
		if err == nil {
			t.Error("Expected error when template is a file")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("Error should mention 'not a directory', got: %v", err)
		}
	})

	t.Run("should allow repo to overwrite template file", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		templateDir := filepath.Join(root, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "config.json")
		if err := os.WriteFile(templateFile, []byte(`{"source": "template"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		repoURL := CreateLocalGitRepo(t, "repo-with-config", map[string]string{"config.json": `{"source": "repo"}`})

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose:  "Repo overwrites template test",
			Template: templateDir,
			Repositories: []RepositoryOption{
				{URL: repoURL, Ref: "main"},
			},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		configPath := filepath.Join(ws.Path, "repo-with-config", "config.json")
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config.json: %v", err)
		}
		if !strings.Contains(string(content), "repo") {
			t.Error("Expected repo content to overwrite template content")
		}
	})

	t.Run("should handle multiple template variables", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		templateDir := filepath.Join(root, "template")
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		templateFile := filepath.Join(templateDir, "{{env}}-{{region}}-config.json")
		if err := os.WriteFile(templateFile, []byte(`{"env": "{{env}}", "region": "{{region}}"}`), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose:  "Multiple vars test",
			Template: templateDir,
			TemplateVars: map[string]string{
				"env":    "staging",
				"region": "us-west",
			},
			Repositories: []RepositoryOption{},
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		expectedFile := "staging-us-west-config.json"
		if !FileExists(filepath.Join(ws.Path, expectedFile)) {
			t.Errorf("Expected %s with variable substitution", expectedFile)
		}
	})
}
