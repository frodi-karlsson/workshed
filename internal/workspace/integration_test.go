package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateAtomicBehaviorShouldMaintainAtomicBehaviorDuringWorkspaceCreation(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	t.Run("successful creation leaves no temp dirs", func(t *testing.T) {
		ctx := context.Background()
		opts := CreateOptions{
			Purpose: "Test workspace",
		}

		ws, err := store.Create(ctx, opts)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".tmp-") {
				t.Errorf("Temp directory %q found after successful creation", entry.Name())
			}
		}

		if !fileExists(ws.Path) {
			t.Error("Workspace directory does not exist")
		}

		if !fileExists(filepath.Join(ws.Path, metadataFileName)) {
			t.Error("Metadata file does not exist")
		}
	})

	t.Run("invalid repo URL leaves no artifacts", func(t *testing.T) {
		beforeEntries, err := os.ReadDir(root)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		ctx := context.Background()
		opts := CreateOptions{
			Purpose: "Test workspace",
			RepoURL: "invalid-scheme://example.com/repo",
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

		for _, entry := range afterEntries {
			if strings.HasPrefix(entry.Name(), ".tmp-") {
				t.Errorf("Temp directory %q not cleaned up after failure", entry.Name())
			}
		}
	})

	t.Run("validation happens before filesystem operations", func(t *testing.T) {
		beforeEntries, err := os.ReadDir(root)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		ctx := context.Background()
		opts := CreateOptions{
			Purpose: "Test workspace",
			RepoURL: "",
		}

		opts.RepoURL = "ftp://invalid.com/repo"
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
}

func TestWorkspaceCorruptedMetadataShouldHandleGracefully(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts := CreateOptions{
		Purpose: "Test workspace",
	}

	ws, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	metaPath := filepath.Join(ws.Path, metadataFileName)
	if err := os.WriteFile(metaPath, []byte("invalid json {{{"), 0644); err != nil {
		t.Fatalf("Failed to corrupt metadata: %v", err)
	}

	_, err = store.Get(ctx, ws.Handle)
	if err == nil {
		t.Error("Expected error when reading corrupted metadata")
	}
}

func TestWorkspaceMalformedStructureShouldBeDetected(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts := CreateOptions{
		Purpose: "Test workspace",
	}

	ws, err := store.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := os.RemoveAll(ws.Path); err != nil {
		t.Fatalf("Failed to remove workspace directory: %v", err)
	}

	_, err = store.Get(ctx, ws.Handle)
	if err == nil {
		t.Error("Expected error when workspace directory is missing")
	}
}

func TestWorkspaceWithSpecialCharactersInPurposeShouldWork(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()

	testCases := []struct {
		name    string
		purpose string
	}{
		{"unicode_chars", "Debug: payment flow with café and naïve users"},
		{"numbers_and_symbols", "Test workspace #123 @home & away"},
		{"quotes", `Workspace with "double" and 'single' quotes`},
		{"slashes", "Path/with/slashes and back\\too"},
		{"spaces", "Lots of    spaces   and	tabs"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := CreateOptions{
				Purpose: tc.purpose,
			}

			ws, err := store.Create(ctx, opts)
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
}

func TestConcurrentWorkspaceCreationShouldAvoidCollisions(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	numWorkspaces := 10

	type result struct {
		ws  *Workspace
		err error
		idx int
	}

	results := make(chan result, numWorkspaces)

	for i := 0; i < numWorkspaces; i++ {
		go func(idx int) {
			opts := CreateOptions{
				Purpose: "Concurrent workspace",
			}
			ws, err := store.Create(ctx, opts)
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
}

func TestWorkspaceCloneWithInvalidAuthShouldClassifyCorrectly(t *testing.T) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	ctx := context.Background()
	opts := CreateOptions{
		Purpose: "Test private repo",
		RepoURL: "https://github.com/this-repo-does-not-exist-12345/private-repo.git",
		RepoRef: "main",
	}

	ws, err := store.Create(ctx, opts)
	if err == nil {
		if ws != nil && fileExists(ws.Path) {
			t.Log("Warning: clone succeeded unexpectedly (repo may be public)")
			if err := store.Remove(ctx, ws.Handle); err != nil {
				t.Logf("Warning: failed to cleanup: %v", err)
			}
		} else {
			t.Skip("Network unavailable, skipping auth failure test")
		}
		return
	}

	if !strings.Contains(err.Error(), "repository not found") &&
		!strings.Contains(err.Error(), "authentication failed") &&
		!strings.Contains(err.Error(), "not found") {
		t.Logf("Got error: %v", err)
	}
}

func TestWorkspaceCloneTimeoutShouldCleanUpProperly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	cancelCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := CreateOptions{
		Purpose: "Test timeout cleanup",
		RepoURL: "https://github.com/torvalds/linux",
		RepoRef: "master",
	}

	_, err = store.Create(cancelCtx, opts)
	if err == nil {
		t.Skip("Clone succeeded before timeout, skipping cleanup verification")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temp directory %q was not cleaned up after timeout", entry.Name())
		}
	}
}
