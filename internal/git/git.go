package git

import (
	"context"
	"errors"
	"strings"
)

// CloneOptions configures how a repository clone operation behaves.
type CloneOptions struct {
	// Depth specifies how many commits to clone. Zero means full history.
	Depth int

	// Mirror creates a bare mirror repository.
	Mirror bool
}

// Git error types for common failure scenarios.
var (
	// ErrRepositoryNotFound indicates the repository URL is invalid or inaccessible.
	ErrRepositoryNotFound = errors.New("repository not found")

	// ErrAuthenticationFailed indicates credentials were rejected.
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrNetworkError indicates a network connectivity issue.
	ErrNetworkError = errors.New("network error")

	// ErrRefNotFound indicates the git reference (branch/tag/commit) doesn't exist.
	ErrRefNotFound = errors.New("ref not found")
)

// Git defines the interface for interacting with git repositories.
// Implementations provide either real git operations or mocked behavior for testing.
type Git interface {
	// Clone creates a local copy of a remote repository.
	Clone(ctx context.Context, url, dir string, opts CloneOptions) error

	// Checkout updates the working tree to a specific git reference.
	Checkout(ctx context.Context, dir, ref string) error

	// GetRemoteURL returns the URL of the origin remote for a repository.
	GetRemoteURL(ctx context.Context, dir string) (string, error)

	// CurrentBranch returns the name of the currently checked out branch.
	CurrentBranch(ctx context.Context, dir string) (string, error)
}

func ClassifyError(operation string, err error, output []byte) error {
	outputStr := string(output)
	var hint string

	switch {
	case strings.Contains(outputStr, "Repository not found") ||
		strings.Contains(outputStr, "repository not found") ||
		strings.Contains(outputStr, "not found"):
		hint = "repository not found"
	case strings.Contains(outputStr, "Authentication failed") ||
		strings.Contains(outputStr, "Permission denied") ||
		strings.Contains(outputStr, "could not read Username") ||
		strings.Contains(outputStr, "fatal: Could not read from remote repository"):
		hint = "authentication failed"
	case strings.Contains(outputStr, "Could not resolve host") ||
		strings.Contains(outputStr, "unable to access") ||
		strings.Contains(outputStr, "network") ||
		strings.Contains(outputStr, "Connection") ||
		strings.Contains(outputStr, "timeout"):
		hint = "network error"
	case strings.Contains(outputStr, "pathspec") && strings.Contains(outputStr, "did not match") ||
		strings.Contains(outputStr, "reference is not a tree"):
		hint = "ref not found"
	}

	if hint != "" {
		return classify(operation, hint, outputStr)
	}

	return classify(operation, "failed", err.Error())
}

func classify(operation, hint, details string) error {
	return &GitError{
		Operation: operation,
		Hint:      hint,
		Details:   details,
	}
}

// GitError represents a git operation failure with structured error information.
// Use ClassifyError to create instances from git command output.
type GitError struct {
	// Operation identifies which git operation failed.
	Operation string

	// Hint provides a user-friendly error category.
	Hint string

	// Details contains the original error output.
	Details string
}

func (e *GitError) Error() string {
	return gitErrorString(e.Operation, e.Hint, e.Details)
}

func (e *GitError) Unwrap() error {
	return nil
}

func gitErrorString(operation, hint, details string) string {
	return strings.TrimSpace(operation + " failed (" + hint + "): " + details)
}
