package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/handle"
	"github.com/frodi/workshed/internal/logger"
)

const metadataFileName = ".workshed.json"

// FSStore is a filesystem-based workspace store that manages workspace directories and metadata.
type FSStore struct {
	root string
}

// NewFSStore creates a new filesystem-based workspace store at the specified root directory.
func NewFSStore(root string) (*FSStore, error) {
	if root == "" {
		return nil, errors.New("root directory cannot be empty")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, fmt.Errorf("creating root directory: %w", err)
	}

	return &FSStore{root: absRoot}, nil
}

// Root returns the root directory for this store.
func (s *FSStore) Root() string {
	return s.root
}

// Create creates a new workspace with the given options and returns the workspace metadata.
func (s *FSStore) Create(ctx context.Context, opts CreateOptions) (*Workspace, error) {
	if opts.Purpose == "" {
		return nil, errors.New("purpose is required")
	}

	if opts.RepoURL != "" {
		if err := validateRepoURL(opts.RepoURL); err != nil {
			return nil, fmt.Errorf("invalid repository URL: %w", err)
		}
	}

	gen := handle.NewGenerator()
	h, err := gen.GenerateUnique(func(h string) bool {
		_, err := s.Get(ctx, h)
		return err == nil
	})
	if err != nil {
		return nil, fmt.Errorf("generating handle: %w", err)
	}

	ws := &Workspace{
		Version:   CurrentMetadataVersion,
		Handle:    h,
		Purpose:   opts.Purpose,
		RepoURL:   opts.RepoURL,
		RepoRef:   opts.RepoRef,
		CreatedAt: time.Now(),
	}

	tmpDir, err := os.MkdirTemp(s.root, ".tmp-")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}

	success := false
	var cleanupErr error
	defer func() {
		if !success {
			if err := os.RemoveAll(tmpDir); err != nil {
				// Store the cleanup error - we'll combine it with the main error
				cleanupErr = fmt.Errorf("cleanup of temp directory %s failed: %w", tmpDir, err)
			}
		}
	}()

	if err := s.writeMetadataToDir(ws, tmpDir); err != nil {
		if cleanupErr != nil {
			return nil, fmt.Errorf("writing metadata: %w; %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("writing metadata: %w", err)
	}

	if opts.RepoURL != "" {
		if err := s.cloneRepo(ctx, ws, tmpDir); err != nil {
			if cleanupErr != nil {
				return nil, fmt.Errorf("cloning repository: %w; %v", err, cleanupErr)
			}
			return nil, fmt.Errorf("cloning repository: %w", err)
		}
	}

	finalDir := s.workspaceDir(h)
	if err := os.Rename(tmpDir, finalDir); err != nil {
		if cleanupErr != nil {
			return nil, fmt.Errorf("finalizing workspace: %w; %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("finalizing workspace: %w", err)
	}

	success = true
	ws.Path = finalDir
	return ws, nil
}

// Get retrieves workspace metadata by handle.
func (s *FSStore) Get(ctx context.Context, handle string) (*Workspace, error) {
	metaPath := filepath.Join(s.workspaceDir(handle), metadataFileName)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("workspace not found: %s", handle)
		}
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	var ws Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}

	ws.Path = s.workspaceDir(handle)
	return &ws, nil
}

// List returns all workspaces matching the given filter options.
func (s *FSStore) List(ctx context.Context, opts ListOptions) ([]*Workspace, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Workspace{}, nil
		}
		return nil, fmt.Errorf("reading workspaces directory: %w", err)
	}

	l := logger.NewLogger(logger.ERROR, "workspace")
	var workspaces []*Workspace
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ws, err := s.Get(ctx, entry.Name())
		if err != nil {
			l.Error("skipping corrupted workspace directory", "handle", entry.Name(), "error", err)
			continue
		}

		if opts.PurposeFilter != "" && !strings.Contains(strings.ToLower(ws.Purpose), strings.ToLower(opts.PurposeFilter)) {
			continue
		}

		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
}

// Remove deletes the workspace with the given handle.
func (s *FSStore) Remove(ctx context.Context, handle string) error {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(ws.Path); err != nil {
		return fmt.Errorf("removing workspace directory: %w", err)
	}

	return nil
}

// Path returns the filesystem path for the workspace with the given handle.
func (s *FSStore) Path(ctx context.Context, handle string) (string, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return "", err
	}
	return ws.Path, nil
}

func (s *FSStore) workspaceDir(handle string) string {
	return filepath.Join(s.root, handle)
}

func validateRepoURL(url string) error {
	if url == "" {
		return errors.New("repository URL cannot be empty")
	}

	if strings.HasPrefix(url, "git@") {
		if !strings.Contains(url, ":") {
			return errors.New("invalid SSH URL format")
		}
		return nil
	}

	validSchemes := []string{"https://", "http://", "git://", "ssh://"}
	for _, scheme := range validSchemes {
		if strings.HasPrefix(url, scheme) {
			return nil
		}
	}

	return fmt.Errorf("unsupported URL scheme (expected https://, git@, ssh://, or git://)")
}

func (s *FSStore) writeMetadataToDir(ws *Workspace, dir string) error {
	metaPath := filepath.Join(dir, metadataFileName)

	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("writing metadata file: %w", err)
	}

	return nil
}

func (s *FSStore) cloneRepo(ctx context.Context, ws *Workspace, wsDir string) error {
	url := ws.RepoURL
	ref := ws.RepoRef
	if ref == "" {
		ref = "main"
	}

	repoName := extractRepoName(url)
	if repoName == "" {
		return errors.New("could not extract repository name from URL")
	}

	repoDir := filepath.Join(wsDir, repoName)

	cmd := exec.CommandContext(ctx, "git", "clone", url, repoDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return classifyGitError("clone", err, output)
	}

	cmd = exec.CommandContext(ctx, "git", "checkout", ref)
	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return classifyGitError("checkout", err, output)
	}

	return nil
}

func classifyGitError(operation string, err error, output []byte) error {
	outputStr := string(output)
	var hint string

	// Check for common error patterns
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
		return fmt.Errorf("git %s failed (%s): %s", operation, hint, outputStr)
	}

	return fmt.Errorf("git %s failed: %w\n%s", operation, err, outputStr)
}

func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, ".git")

	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}

	if idx := strings.LastIndex(url, ":"); idx != -1 {
		parts := strings.Split(url[idx+1:], "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	return ""
}
