package workspace

import (
	"context"
	"strings"
	"testing"
)

func TestCreateValidation(t *testing.T) {
	t.Run("should return error when purpose is empty", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		opts := CreateOptions{}

		ctx := context.Background()
		_, err = store.Create(ctx, opts)
		if err == nil {
			t.Error("Expected error when purpose is empty")
		}
		if err.Error() != "purpose is required" {
			t.Errorf("Expected 'purpose is required' error, got: %v", err)
		}
	})

	t.Run("should return error for invalid repository URL", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		opts := CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: "ftp://invalid.com/repo"},
			},
		}

		ctx := context.Background()
		_, err = store.Create(ctx, opts)
		if err == nil {
			t.Error("Expected error for invalid URL scheme")
		}
		if !strings.Contains(err.Error(), "unsupported URL scheme") {
			t.Errorf("Expected URL scheme error, got: %v", err)
		}
	})

	t.Run("should return error for duplicate repository URLs", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		opts := CreateOptions{
			Purpose: "Test workspace",
			Repositories: []RepositoryOption{
				{URL: "https://github.com/test/repo"},
				{URL: "https://github.com/test/repo"},
			},
		}

		ctx := context.Background()
		_, err = store.Create(ctx, opts)
		if err == nil {
			t.Error("Expected error for duplicate repository URLs")
		}
		if !strings.Contains(err.Error(), "duplicate repository URL") {
			t.Errorf("Expected duplicate URL error, got: %v", err)
		}
	})
}

func TestExtractRepoName(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"https://github.com/org/repo", "repo"},
		{"https://github.com/org/repo.git", "repo"},
		{"git@github.com:org/repo.git", "repo"},
		{"git@github.com:org/repo", "repo"},
		{"ssh://git@github.com/org/repo", "repo"},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			name := extractRepoName(tc.url)
			if name != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, name)
			}
		})
	}
}

func TestValidateRepoURL(t *testing.T) {
	validURLs := []string{
		"https://github.com/org/repo",
		"http://github.com/org/repo",
		"git://github.com/org/repo",
		"ssh://git@github.com/org/repo",
		"git@github.com:org/repo",
	}

	for _, url := range validURLs {
		t.Run(url, func(t *testing.T) {
			err := validateRepoURL(url)
			if err != nil {
				t.Errorf("Expected valid URL, got error: %v", err)
			}
		})
	}

	invalidURLs := []string{
		"",
		"ftp://github.com/org/repo",
		"git@github.com", // missing colon
	}

	for _, url := range invalidURLs {
		t.Run(url, func(t *testing.T) {
			err := validateRepoURL(url)
			if err == nil {
				t.Error("Expected error for invalid URL")
			}
		})
	}
}

func TestWorkspaceGetRepositoryByName(t *testing.T) {
	ws := &Workspace{
		Repositories: []Repository{
			{Name: "backend", URL: "https://github.com/org/backend"},
			{Name: "frontend", URL: "https://github.com/org/frontend"},
		},
	}

	t.Run("should find existing repository", func(t *testing.T) {
		repo := ws.GetRepositoryByName("backend")
		if repo == nil {
			t.Fatal("Expected to find backend repository")
		}
		if repo.URL != "https://github.com/org/backend" {
			t.Errorf("Expected backend URL, got: %s", repo.URL)
		}
	})

	t.Run("should return nil for nonexistent repository", func(t *testing.T) {
		repo := ws.GetRepositoryByName("nonexistent")
		if repo != nil {
			t.Error("Expected nil for nonexistent repository")
		}
	})
}

func TestExec(t *testing.T) {
	t.Run("should return error for nonexistent workspace", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		_, err = store.Exec(ctx, "nonexistent-handle", ExecOptions{
			Target:  "all",
			Command: []string{"echo", "hello"},
		})
		if err == nil {
			t.Error("Expected error for nonexistent workspace")
		}
		if !strings.Contains(err.Error(), "workspace not found") {
			t.Errorf("Expected 'workspace not found' error, got: %v", err)
		}
	})

	t.Run("should execute command in workspace root", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		results, err := store.Exec(ctx, ws.Handle, ExecOptions{
			Target:  "root",
			Command: []string{"pwd"},
		})
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Repository != "root" {
			t.Errorf("Expected repository 'root', got: %s", results[0].Repository)
		}
		if results[0].ExitCode != 0 {
			t.Errorf("Expected exit code 0, got: %d", results[0].ExitCode)
		}
	})

	t.Run("should execute command and return exit code on failure", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		_, err = store.Exec(ctx, ws.Handle, ExecOptions{
			Target:  "root",
			Command: []string{"sh", "-c", "exit 42"},
		})
		if err == nil {
			t.Error("Expected error for failed command")
		}
		if !strings.Contains(err.Error(), "exit code 42") {
			t.Errorf("Expected 'exit code 42' in error, got: %v", err)
		}
	})
}

func TestExecInRepository(t *testing.T) {
	t.Run("should return error for missing directory", func(t *testing.T) {
		root := t.TempDir()
		store, err := NewFSStore(root)
		if err != nil {
			t.Fatalf("NewFSStore failed: %v", err)
		}

		ctx := context.Background()
		ws, err := store.Create(ctx, CreateOptions{
			Purpose: "Test workspace",
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		repo := Repository{Name: "nonexistent", URL: "https://github.com/test/repo"}
		result, err := store.execInRepository(ctx, repo, ws.Path, []string{"echo", "hello"})
		if err == nil {
			t.Error("Expected error for missing directory")
		}
		if result.ExitCode == 0 {
			t.Error("Expected non-zero exit code")
		}
	})
}
