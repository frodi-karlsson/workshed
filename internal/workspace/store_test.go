package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFSStoreCreateShouldCreateWorkspaceWithValidOptions(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	opts := CreateOptions{
		Purpose: "Test workspace",
	}

	ctx := context.Background()
	ws, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if ws.Handle == "" {
		t.Error("Expected non-empty handle")
	}

	if ws.Purpose != opts.Purpose {
		t.Errorf("Expected purpose %q, got %q", opts.Purpose, ws.Purpose)
	}

	if ws.Version != CurrentMetadataVersion {
		t.Errorf("Expected version %d, got %d", CurrentMetadataVersion, ws.Version)
	}

	if !fileExists(filepath.Join(ws.Path, metadataFileName)) {
		t.Error("Metadata file not created")
	}
}

func TestFSStoreCreateShouldReturnErrorWhenPurposeIsEmpty(t *testing.T) {
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
}

func TestFSStoreGetShouldRetrieveExistingWorkspaceByHandle(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	opts := CreateOptions{
		Purpose: "Test workspace",
	}

	ctx := context.Background()
	created, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := store.Get(ctx, created.Handle)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Handle != created.Handle {
		t.Errorf("Expected handle %q, got %q", created.Handle, retrieved.Handle)
	}

	if retrieved.Purpose != created.Purpose {
		t.Errorf("Expected purpose %q, got %q", created.Purpose, retrieved.Purpose)
	}
}

func TestFSStoreGetShouldReturnErrorForNonexistentWorkspace(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.Get(ctx, "nonexistent-handle")
	if err == nil {
		t.Error("Expected error for nonexistent workspace")
	}
}

func TestFSStoreListShouldReturnAllWorkspacesWhenNoFilterProvided(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts1 := CreateOptions{Purpose: "Debug payment"}
	opts2 := CreateOptions{Purpose: "Add feature"}

	_, err = store.Create(ctx, opts1)
	if err != nil {
		t.Fatalf("Create 1 failed: %v", err)
	}

	_, err = store.Create(ctx, opts2)
	if err != nil {
		t.Fatalf("Create 2 failed: %v", err)
	}

	workspaces, err := store.List(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("Expected 2 workspaces, got %d", len(workspaces))
	}
}

func TestFSStoreListShouldFilterWorkspacesByPurposeWhenFilterProvided(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts1 := CreateOptions{Purpose: "Debug payment"}
	opts2 := CreateOptions{Purpose: "Add feature"}

	_, err = store.Create(ctx, opts1)
	if err != nil {
		t.Fatalf("Create 1 failed: %v", err)
	}

	_, err = store.Create(ctx, opts2)
	if err != nil {
		t.Fatalf("Create 2 failed: %v", err)
	}

	workspaces, err := store.List(ctx, ListOptions{PurposeFilter: "debug"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	if workspaces[0].Purpose != opts1.Purpose {
		t.Errorf("Expected purpose %q, got %q", opts1.Purpose, workspaces[0].Purpose)
	}
}

func TestFSStoreRemoveShouldDeleteWorkspaceDirectory(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	opts := CreateOptions{Purpose: "Test workspace"}

	ctx := context.Background()
	ws, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := store.Remove(ctx, ws.Handle); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if fileExists(ws.Path) {
		t.Error("Workspace directory still exists after removal")
	}
}

func TestFSStorePathShouldReturnCorrectWorkspacePath(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	opts := CreateOptions{Purpose: "Test workspace"}

	ctx := context.Background()
	ws, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	path, err := store.Path(ctx, ws.Handle)
	if err != nil {
		t.Fatalf("Path failed: %v", err)
	}

	if path != ws.Path {
		t.Errorf("Expected path %q, got %q", ws.Path, path)
	}
}

func TestExtractRepoNameShouldExtractRepositoryNameFromVariousURLFormats(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/repo", "repo"},
		{"https://github.com/user/repo.git", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
	}

	for _, tt := range tests {
		got := extractRepoName(tt.url)
		if got != tt.want {
			t.Errorf("extractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestValidateRepoURLShouldValidateRepositoryURLsCorrectly(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://github.com/user/repo", false},
		{"http://github.com/user/repo", false},
		{"git://github.com/user/repo", false},
		{"ssh://git@github.com/user/repo", false},
		{"git@github.com:user/repo.git", false},
		{"", true},
		{"ftp://github.com/user/repo", true},
		{"git@github.com", true},
		{"invalid-url", true},
	}

	for _, tt := range tests {
		err := validateRepoURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateRepoURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
		}
	}
}

func TestCreateAtomicCleanupShouldCleanUpTempDirectoriesOnFailure(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts := CreateOptions{
		Purpose: "Test workspace",
		RepoURL: "invalid-url",
	}

	_, err = store.Create(ctx, opts)
	if err == nil {
		t.Fatal("Expected error when creating workspace with invalid repo URL")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temp directory %q was not cleaned up after failure", entry.Name())
		}
	}
}

func TestGitErrorClassificationShouldClassifyGitErrorsAppropriately(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		err            error
		output         string
		wantHint       string
		wantErrContain string
	}{
		{
			name:           "repository not found",
			operation:      "clone",
			err:            errors.New("exit status 128"),
			output:         "remote: Repository not found.\nfatal: repository 'https://github.com/org/notfound' not found",
			wantHint:       "repository not found",
			wantErrContain: "git clone failed (repository not found)",
		},
		{
			name:           "authentication failed",
			operation:      "clone",
			err:            errors.New("exit status 128"),
			output:         "fatal: Authentication failed for 'https://github.com/org/repo.git/'",
			wantHint:       "authentication failed",
			wantErrContain: "git clone failed (authentication failed)",
		},
		{
			name:           "permission denied",
			operation:      "clone",
			err:            errors.New("exit status 128"),
			output:         "Permission denied (publickey).\nfatal: Could not read from remote repository.",
			wantHint:       "authentication failed",
			wantErrContain: "git clone failed (authentication failed)",
		},
		{
			name:           "network error - could not resolve host",
			operation:      "clone",
			err:            errors.New("exit status 128"),
			output:         "fatal: unable to access 'https://invalid.example.com/repo.git/': Could not resolve host: invalid.example.com",
			wantHint:       "network error",
			wantErrContain: "git clone failed (network error)",
		},
		{
			name:           "ref not found",
			operation:      "checkout",
			err:            errors.New("exit status 1"),
			output:         "error: pathspec 'nonexistent-branch' did not match any file(s) known to git",
			wantHint:       "ref not found",
			wantErrContain: "git checkout failed (ref not found)",
		},
		{
			name:           "unknown error",
			operation:      "clone",
			err:            errors.New("exit status 1"),
			output:         "some unknown error occurred",
			wantHint:       "",
			wantErrContain: "git clone failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyGitError(tt.operation, tt.err, []byte(tt.output))

			if err == nil {
				t.Fatal("classifyGitError should return an error")
			}

			errStr := err.Error()
			if !strings.Contains(errStr, tt.wantErrContain) {
				t.Errorf("Error should contain %q, got: %s", tt.wantErrContain, errStr)
			}

			// Verify original output is preserved
			if !strings.Contains(errStr, tt.output) {
				t.Errorf("Error should preserve original output, got: %s", errStr)
			}
		})
	}
}

func TestCloneWithContextTimeoutShouldRespectContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	tmpDir := t.TempDir()
	store, err := NewFSStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := CreateOptions{
		Purpose: "Test timeout",
		RepoURL: "https://github.com/torvalds/linux", // Large repo that won't clone in 100ms
		RepoRef: "master",
	}

	_, err = store.Create(ctx, opts)
	if err == nil {
		t.Error("Create should fail with timeout")
	}

	// Verify error mentions context or timeout
	if !strings.Contains(err.Error(), "context") &&
		!strings.Contains(err.Error(), "signal") &&
		!strings.Contains(err.Error(), "killed") {
		t.Logf("Expected context/signal/killed error, got: %v", err)
	}

	// Verify cleanup happened
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read tmpDir: %v", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temporary directory not cleaned up: %s", entry.Name())
		}
	}
}
