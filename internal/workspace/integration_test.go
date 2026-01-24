package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
