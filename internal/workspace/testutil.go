package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type WorkspaceTestEnvironment struct {
	T       *testing.T
	Root    string
	Store   *FSStore
	Ctx     context.Context
	tmpDirs []string
}

func NewWorkspaceTestEnvironment(t *testing.T) *WorkspaceTestEnvironment {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}

	return &WorkspaceTestEnvironment{
		T:     t,
		Root:  root,
		Store: store,
		Ctx:   context.Background(),
	}
}

func (e *WorkspaceTestEnvironment) Cleanup() {
	for _, dir := range e.tmpDirs {
		if err := os.RemoveAll(dir); err != nil {
			e.T.Logf("Warning: failed to remove temp dir %s: %v", dir, err)
		}
	}
}

func (e *WorkspaceTestEnvironment) CreateWorkspace(purpose string) *Workspace {
	opts := CreateOptions{
		Purpose: purpose,
	}
	ws, err := e.Store.Create(e.Ctx, opts)
	if err != nil {
		e.T.Fatalf("Create failed: %v", err)
	}
	return ws
}

func (e *WorkspaceTestEnvironment) CreateWorkspaceWithRepo(purpose, repoURL, repoRef string) *Workspace {
	opts := CreateOptions{
		Purpose: purpose,
		RepoURL: repoURL,
		RepoRef: repoRef,
	}
	ws, err := e.Store.Create(e.Ctx, opts)
	if err != nil {
		e.T.Fatalf("Create failed: %v", err)
	}
	return ws
}

func (e *WorkspaceTestEnvironment) MustGet(handle string) *Workspace {
	ws, err := e.Store.Get(e.Ctx, handle)
	if err != nil {
		e.T.Fatalf("Get failed: %v", err)
	}
	return ws
}

func (e *WorkspaceTestEnvironment) MustRemove(handle string) {
	if err := e.Store.Remove(e.Ctx, handle); err != nil {
		e.T.Fatalf("Remove failed: %v", err)
	}
}

func VerifyNoTempDirectories(t *testing.T, root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temp directory %q found after operation", entry.Name())
		}
	}
}

func VerifyTempDirectoriesCleaned(t *testing.T, root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temp directory %q not cleaned up", entry.Name())
		}
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func MustHaveFile(t *testing.T, path string) {
	if !FileExists(path) {
		t.Errorf("File should exist: %s", path)
	}
}

func MustNotHaveFile(t *testing.T, path string) {
	if FileExists(path) {
		t.Errorf("File should not exist: %s", path)
	}
}

func MustNotHaveTempDirs(t *testing.T, root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temp directory found: %s", entry.Name())
		}
	}
}

func WorkspacePath(root, handle string) string {
	return filepath.Join(root, handle)
}

func CreateTestStore(t *testing.T) (*FSStore, string) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}
	return store, root
}
