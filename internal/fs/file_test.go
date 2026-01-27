package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	t.Run("single level", func(t *testing.T) {
		root := t.TempDir()
		newDir := filepath.Join(root, "newdir")

		if err := EnsureDir(newDir); err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(newDir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected directory, got file")
		}
	})

	t.Run("nested", func(t *testing.T) {
		root := t.TempDir()
		newDir := filepath.Join(root, "a", "b", "c")

		if err := EnsureDir(newDir); err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(newDir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Expected directory, got file")
		}
	})

	t.Run("already exists", func(t *testing.T) {
		root := t.TempDir()
		existing := filepath.Join(root, "existing")

		if err := os.MkdirAll(existing, 0755); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if err := EnsureDir(existing); err != nil {
			t.Fatalf("EnsureDir should succeed for existing dir: %v", err)
		}
	})
}

func TestWriteJson(t *testing.T) {
	t.Run("basic write", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "output.json")

		if err := WriteJson(path, []byte(`{"key":"value"}`)); err != nil {
			t.Fatalf("WriteJson failed: %v", err)
		}

		content, _ := os.ReadFile(path)
		if !contains(string(content), "key") {
			t.Error("File content incorrect")
		}
	})

	t.Run("creates parent dirs", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a", "b", "output.json")

		if err := WriteJson(path, []byte(`{}`)); err != nil {
			t.Fatalf("WriteJson failed: %v", err)
		}

		if _, err := os.Stat(filepath.Dir(path)); err != nil {
			t.Error("Parent dir not created")
		}
	})
}

func TestWriteText(t *testing.T) {
	t.Run("basic write", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "output.txt")

		if err := WriteText(path, []byte("hello")); err != nil {
			t.Fatalf("WriteText failed: %v", err)
		}

		content, _ := os.ReadFile(path)
		if string(content) != "hello" {
			t.Errorf("Content mismatch: got %q", string(content))
		}
	})

	t.Run("creates parent dirs", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a", "b", "output.txt")

		if err := WriteText(path, []byte("text")); err != nil {
			t.Fatalf("WriteText failed: %v", err)
		}

		if _, err := os.Stat(filepath.Dir(path)); err != nil {
			t.Error("Parent dir not created")
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
