package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/git"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func MustHaveFile(t *testing.T, path string) {
	if !FileExists(path) {
		t.Errorf("File should exist: %s", path)
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

func CreateTestStore(t *testing.T) (*FSStore, string) {
	root := t.TempDir()
	store, err := NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}
	return store, root
}

func CreateMockedTestStore(t *testing.T) (*FSStore, string, *git.MockGit) {
	root := t.TempDir()
	mockGit := &git.MockGit{}
	mockGit.SetCurrentBranchResult("main")
	store, err := NewFSStore(root, mockGit)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}
	return store, root, mockGit
}

func CreateFakeRepo(t *testing.T, root, name string) string {
	repoDir := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
		t.Fatalf("Failed to create fake repo dir: %v", err)
	}
	return repoDir
}

func CreateLocalGitRepo(t *testing.T, name string, files map[string]string) string {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, name)

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to rename branch: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config user.email failed: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config user.name failed: %v\n%s", err, out)
	}

	if err := AddGitCommit(repoDir, "Initial commit", files); err != nil {
		t.Fatalf("Failed to add initial commit: %v", err)
	}

	return repoDir
}

func AddGitCommit(repoDir, message string, files map[string]string) error {
	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
		cmd := exec.Command("git", "add", relPath)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add %s failed: %v\n%s", relPath, err, out)
		}
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %v\n%s", err, out)
	}

	return nil
}

func CreateGitBranch(repoDir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b %s failed: %v\n%s", branchName, err, out)
	}
	return nil
}

func ChangeDefaultBranch(t *testing.T, repoDir, branchName string) {
	if err := CreateGitBranch(repoDir, branchName); err != nil {
		t.Fatalf("Failed to create branch %s: %v", branchName, err)
	}
	cmd := exec.Command("git", "branch", "-M", branchName)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to rename default branch to %s: %v\n%s", branchName, err, out)
	}
}
