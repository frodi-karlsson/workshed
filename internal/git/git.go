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
type Git interface {
	// Init creates a new git repository.
	Init(ctx context.Context, dir string) error

	// Clone creates a local copy of a remote repository.
	Clone(ctx context.Context, url, dir string, opts CloneOptions) error

	// Checkout updates the working tree to a specific git reference.
	Checkout(ctx context.Context, dir, ref string) error

	// GetRemoteURL returns the URL of the origin remote for a repository.
	GetRemoteURL(ctx context.Context, dir string) (string, error)

	// CurrentBranch returns the name of the currently checked out branch.
	CurrentBranch(ctx context.Context, dir string) (string, error)

	// DefaultBranch returns the default branch name for a remote repository.
	// Returns empty string and error if the repository is inaccessible.
	DefaultBranch(ctx context.Context, url string) (string, error)

	// RevParse returns the commit hash for a given reference.
	RevParse(ctx context.Context, dir, ref string) (string, error)

	// StatusPorcelain returns the git status in porcelain format.
	StatusPorcelain(ctx context.Context, dir string) (string, error)
}

func ClassifyError(operation string, err error, output []byte) error {
	outputStr := string(output)
	var hint string
	var suggestion string

	switch {
	case strings.Contains(outputStr, "Repository not found") ||
		strings.Contains(outputStr, "repository not found") ||
		strings.Contains(outputStr, "not found") ||
		strings.Contains(outputStr, "could not resolve") ||
		strings.Contains(outputStr, "404"):
		hint = "repository not found"
		suggestion = "Check that the repository URL is correct and the repository exists."
	case strings.Contains(outputStr, "Authentication failed") ||
		strings.Contains(outputStr, "Permission denied") ||
		strings.Contains(outputStr, "could not read Username") ||
		strings.Contains(outputStr, "fatal: Could not read from remote repository") ||
		strings.Contains(outputStr, "no such identity") ||
		strings.Contains(outputStr, "identity"):
		hint = "authentication failed"
		suggestion = "Ensure SSH keys are configured (ssh-add -l) or use HTTPS with a valid token."
	case strings.Contains(outputStr, "Could not resolve host") ||
		strings.Contains(outputStr, "unable to access") ||
		strings.Contains(outputStr, "network") ||
		strings.Contains(outputStr, "Connection") ||
		strings.Contains(outputStr, "timeout") ||
		strings.Contains(outputStr, "Connection refused"):
		hint = "network error"
		suggestion = "Check your internet connection and firewall settings."
	case strings.Contains(outputStr, "pathspec") && strings.Contains(outputStr, "did not match") ||
		strings.Contains(outputStr, "reference is not a tree") ||
		strings.Contains(outputStr, "could not find") ||
		strings.Contains(outputStr, "remote branch") && strings.Contains(outputStr, "not found"):
		hint = "ref not found"
		suggestion = "Check that the branch or tag exists in the repository."
	}

	return classify(operation, hint, suggestion, outputStr)
}

func classify(operation, hint, suggestion, details string) error {
	return &GitError{
		Operation:  operation,
		Hint:       hint,
		Suggestion: suggestion,
		Details:    details,
	}
}

type GitError struct {
	Operation  string
	Hint       string
	Suggestion string
	Details    string
}

func (e *GitError) Error() string {
	result := gitErrorString(e.Operation, e.Hint, e.Details)
	if e.Suggestion != "" {
		result += "\nSuggestion: " + e.Suggestion
	}
	return result
}

func (e *GitError) Unwrap() error {
	return nil
}

func gitErrorString(operation, hint, details string) string {
	return strings.TrimSpace(operation + " failed (" + hint + "): " + details)
}
