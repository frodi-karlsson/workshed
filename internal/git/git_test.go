package git

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyError(t *testing.T) {
	t.Run("should classify repository not found", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: repository not found"))
		errStr := err.Error()

		if !strings.Contains(errStr, "repository not found") {
			t.Errorf("Error should contain 'repository not found', got: %s", errStr)
		}
		if !strings.Contains(errStr, "Suggestion") {
			t.Errorf("Error should contain suggestion, got: %s", errStr)
		}
		if !strings.Contains(errStr, "URL") || !strings.Contains(errStr, "exists") {
			t.Errorf("Suggestion should mention checking URL and existence, got: %s", errStr)
		}
	})

	t.Run("should classify 404 not found", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: couldn't find remote ref 404"))
		errStr := err.Error()

		if !strings.Contains(errStr, "repository not found") {
			t.Errorf("Error should classify 404 as repository not found, got: %s", errStr)
		}
	})

	t.Run("should classify could not resolve as repository not found", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: could not resolve upstream"))
		errStr := err.Error()

		if !strings.Contains(errStr, "repository not found") {
			t.Errorf("Error should classify 'could not resolve' as repository not found, got: %s", errStr)
		}
	})

	t.Run("should classify authentication failed", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: Authentication failed"))
		errStr := err.Error()

		if !strings.Contains(errStr, "authentication failed") {
			t.Errorf("Error should contain 'authentication failed', got: %s", errStr)
		}
		if !strings.Contains(errStr, "SSH keys") && !strings.Contains(errStr, "token") {
			t.Errorf("Suggestion should mention SSH keys or token, got: %s", errStr)
		}
	})

	t.Run("should classify permission denied", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("Permission denied"))
		errStr := err.Error()

		if !strings.Contains(errStr, "authentication failed") {
			t.Errorf("Error should classify 'Permission denied' as authentication failed, got: %s", errStr)
		}
	})

	t.Run("should classify could not read username as auth failed", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("could not read Username for 'https://'"))
		errStr := err.Error()

		if !strings.Contains(errStr, "authentication failed") {
			t.Errorf("Error should classify 'could not read Username' as authentication failed, got: %s", errStr)
		}
	})

	t.Run("should classify network error", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: Could not resolve host"))
		errStr := err.Error()

		if !strings.Contains(errStr, "network error") {
			t.Errorf("Error should contain 'network error', got: %s", errStr)
		}
		if !strings.Contains(errStr, "internet") && !strings.Contains(errStr, "firewall") {
			t.Errorf("Suggestion should mention internet and firewall, got: %s", errStr)
		}
	})

	t.Run("should classify unable to access as network error", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("fatal: unable to access 'https://...'"))
		errStr := err.Error()

		if !strings.Contains(errStr, "network error") {
			t.Errorf("Error should classify 'unable to access' as network error, got: %s", errStr)
		}
	})

	t.Run("should classify connection refused as network error", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("Connection refused"))
		errStr := err.Error()

		if !strings.Contains(errStr, "network error") {
			t.Errorf("Error should classify 'Connection refused' as network error, got: %s", errStr)
		}
	})

	t.Run("should classify timeout as network error", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("Connection timed out"))
		errStr := err.Error()

		if !strings.Contains(errStr, "network error") {
			t.Errorf("Error should classify 'timed out' as network error, got: %s", errStr)
		}
	})

	t.Run("should classify ref not found", func(t *testing.T) {
		err := ClassifyError("checkout", errors.New("failed"), []byte("error: pathspec 'nonexistent' did not match"))
		errStr := err.Error()

		if !strings.Contains(errStr, "ref not found") {
			t.Errorf("Error should contain 'ref not found', got: %s", errStr)
		}
		if !strings.Contains(errStr, "branch") && !strings.Contains(errStr, "tag") {
			t.Errorf("Suggestion should mention branch or tag, got: %s", errStr)
		}
	})

	t.Run("should classify reference is not a tree as ref not found", func(t *testing.T) {
		err := ClassifyError("checkout", errors.New("failed"), []byte("error: reference is not a tree"))
		errStr := err.Error()

		if !strings.Contains(errStr, "ref not found") {
			t.Errorf("Error should classify 'reference is not a tree' as ref not found, got: %s", errStr)
		}
	})

	t.Run("should classify remote branch not found as repository not found", func(t *testing.T) {
		err := ClassifyError("fetch", errors.New("failed"), []byte("remote branch main not found"))
		errStr := err.Error()

		if !strings.Contains(errStr, "repository not found") {
			t.Errorf("Error should classify 'branch not found' as repository not found (matches 'not found'), got: %s", errStr)
		}
	})

	t.Run("should include original output in error", func(t *testing.T) {
		customOutput := "some custom git error message xyz123"
		err := ClassifyError("clone", errors.New("failed"), []byte(customOutput))
		errStr := err.Error()

		if !strings.Contains(errStr, "xyz123") {
			t.Errorf("Error should contain original output, got: %s", errStr)
		}
	})

	t.Run("should include operation name in error", func(t *testing.T) {
		err := ClassifyError("custom-operation", errors.New("failed"), []byte("some error"))
		errStr := err.Error()

		if !strings.Contains(errStr, "custom-operation") {
			t.Errorf("Error should contain operation name, got: %s", errStr)
		}
	})

	t.Run("should classify unknown errors as failed", func(t *testing.T) {
		err := ClassifyError("clone", errors.New("failed"), []byte("some unknown error that doesn't match any pattern"))
		errStr := err.Error()

		if !strings.Contains(errStr, "failed") {
			t.Errorf("Unknown error should contain 'failed', got: %s", errStr)
		}
	})
}

func TestGitErrorInterface(t *testing.T) {
	t.Run("GitError should implement error interface", func(t *testing.T) {
		err := &GitError{
			Operation:  "clone",
			Hint:       "network error",
			Suggestion: "Check your connection",
			Details:    "test details",
		}

		if err.Error() == "" {
			t.Error("Error() should return non-empty string")
		}
	})

	t.Run("GitError should unwrap to nil", func(t *testing.T) {
		err := &GitError{
			Operation: "clone",
		}

		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil")
		}
	})

	t.Run("GitError with suggestion should include it in Error()", func(t *testing.T) {
		err := &GitError{
			Operation:  "clone",
			Hint:       "authentication failed",
			Suggestion: "Run ssh-add to add your keys",
			Details:    "Permission denied",
		}

		errStr := err.Error()
		if !strings.Contains(errStr, "Run ssh-add to add your keys") {
			t.Errorf("Error() should include suggestion, got: %s", errStr)
		}
	})
}
