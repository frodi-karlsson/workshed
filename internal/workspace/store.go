package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/handle"
	"github.com/oklog/ulid/v2"
)

const metadataFileName = ".workshed.json"
const executionsDirName = "executions"
const capturesDirName = "captures"

// FSStore is a filesystem-based workspace store that manages workspace directories and metadata.
type FSStore struct {
	root string
	git  git.Git
}

// NewFSStore creates a new filesystem-based workspace store at the specified root directory.
func NewFSStore(root string, g ...git.Git) (*FSStore, error) {
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

	var gitClient git.Git = git.RealGit{}
	if len(g) > 0 && g[0] != nil {
		gitClient = g[0]
	}

	return &FSStore{root: absRoot, git: gitClient}, nil
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

	if opts.Template != "" {
		if err := validateTemplatePath(opts.Template); err != nil {
			return nil, fmt.Errorf("invalid template: %w", err)
		}
	}

	repos := opts.Repositories

	// If no repositories provided, use current directory
	if repos == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting current directory: %w", err)
		}
		repos = []RepositoryOption{
			{URL: cwd, Ref: ""},
		}
	}

	if len(repos) > 0 {
		if err := validateRepositories(repos, opts.InvocationCWD); err != nil {
			return nil, fmt.Errorf("invalid repositories: %w", err)
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

	clonedRepos := make([]Repository, len(repos))
	for i, opt := range repos {
		url := opt.URL
		if isLocalPath(url) {
			expandedPath, err := expandPath(url, opts.InvocationCWD)
			if err != nil {
				return nil, fmt.Errorf("expanding local path %s: %w", url, err)
			}
			absPath, err := filepath.Abs(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("resolving absolute local path %s: %w", url, err)
			}
			url = absPath
		}

		clonedRepos[i] = Repository{
			URL:  url,
			Ref:  opt.Ref,
			Name: extractRepoName(opt.URL, opts.InvocationCWD),
		}
	}

	ws := &Workspace{
		Version:      CurrentMetadataVersion,
		Handle:       h,
		Purpose:      opts.Purpose,
		Repositories: clonedRepos,
		CreatedAt:    time.Now(),
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

	if opts.Template != "" {
		if err := s.applyTemplate(ctx, opts.Template, opts.TemplateVars, tmpDir); err != nil {
			if cleanupErr != nil {
				return nil, fmt.Errorf("applying template: %w; %v", err, cleanupErr)
			}
			return nil, fmt.Errorf("applying template: %w", err)
		}
	}

	if err := s.cloneRepositories(ctx, clonedRepos, tmpDir, opts.InvocationCWD); err != nil {
		if cleanupErr != nil {
			return nil, fmt.Errorf("cloning repositories: %w; %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("cloning repositories: %w", err)
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

	var workspaces []*Workspace
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ws, err := s.Get(ctx, entry.Name())
		if err != nil {
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

// UpdatePurpose updates the purpose of an existing workspace.
func (s *FSStore) UpdatePurpose(ctx context.Context, handle string, purpose string) error {
	if purpose == "" {
		return errors.New("purpose cannot be empty")
	}

	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	ws.Purpose = purpose

	if err := s.writeMetadataToDir(ws, ws.Path); err != nil {
		return fmt.Errorf("updating purpose: %w", err)
	}

	return nil
}

// AddRepository adds a single repository to an existing workspace.
func (s *FSStore) AddRepository(ctx context.Context, handle string, repo RepositoryOption, invocationCWD string) error {
	return s.AddRepositories(ctx, handle, []RepositoryOption{repo}, invocationCWD)
}

// AddRepositories adds multiple repositories to an existing workspace.
func (s *FSStore) AddRepositories(ctx context.Context, handle string, repos []RepositoryOption, invocationCWD string) error {
	if len(repos) == 0 {
		return errors.New("no repositories specified")
	}

	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	if err := validateRepositories(repos, invocationCWD); err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	seenURLs := make(map[string]bool)
	seenNames := make(map[string]bool)
	for _, r := range ws.Repositories {
		seenURLs[r.URL] = true
		seenNames[r.Name] = true
	}

	for _, opt := range repos {
		if seenURLs[opt.URL] {
			return fmt.Errorf("repository already exists: %s", opt.URL)
		}
		name := extractRepoName(opt.URL, invocationCWD)
		if seenNames[name] {
			return fmt.Errorf("repository name already exists: %s", name)
		}
	}

	clonedRepos := make([]Repository, len(repos))
	for i, opt := range repos {
		url := opt.URL
		if isLocalPath(url) {
			expandedPath, err := expandPath(url, invocationCWD)
			if err != nil {
				return fmt.Errorf("expanding local path %s: %w", url, err)
			}
			absPath, err := filepath.Abs(expandedPath)
			if err != nil {
				return fmt.Errorf("resolving absolute local path %s: %w", url, err)
			}
			url = absPath
		}

		clonedRepos[i] = Repository{
			URL:  url,
			Ref:  opt.Ref,
			Name: extractRepoName(opt.URL, invocationCWD),
		}
	}

	success := false
	defer func() {
		if !success {
			for _, repo := range clonedRepos {
				repoDir := filepath.Join(ws.Path, repo.Name)
				_ = os.RemoveAll(repoDir)
			}
		}
	}()

	for _, repo := range clonedRepos {
		if err := s.cloneRepo(ctx, repo, ws.Path, invocationCWD); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
	}

	ws.Repositories = append(ws.Repositories, clonedRepos...)

	if err := s.writeMetadataToDir(ws, ws.Path); err != nil {
		return fmt.Errorf("updating metadata: %w", err)
	}

	success = true
	return nil
}

// RemoveRepository removes a repository from an existing workspace.
func (s *FSStore) RemoveRepository(ctx context.Context, handle string, repoName string) error {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	repo := ws.GetRepositoryByName(repoName)
	if repo == nil {
		return fmt.Errorf("repository not found: %s", repoName)
	}

	repoDir := filepath.Join(ws.Path, repo.Name)
	if _, err := os.Stat(repoDir); err == nil {
		if err := os.RemoveAll(repoDir); err != nil {
			return fmt.Errorf("removing repository directory: %w", err)
		}
	}

	newRepos := make([]Repository, 0, len(ws.Repositories)-1)
	for _, r := range ws.Repositories {
		if r.Name != repoName {
			newRepos = append(newRepos, r)
		}
	}
	ws.Repositories = newRepos

	if err := s.writeMetadataToDir(ws, ws.Path); err != nil {
		return fmt.Errorf("updating metadata: %w", err)
	}

	return nil
}

// FindWorkspace finds the workspace that contains the given directory.
// It walks up the directory tree looking for a .workshed.json file.
func (s *FSStore) FindWorkspace(ctx context.Context, dir string) (*Workspace, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path: %w", err)
	}

	for {
		metaPath := filepath.Join(absDir, metadataFileName)
		if _, err := os.Stat(metaPath); err == nil {
			return s.Get(ctx, filepath.Base(absDir))
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("checking for workspace: %w", err)
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			return nil, errors.New("not in a workspace directory")
		}
		absDir = parent
	}
}

type ExecOptions struct {
	Target   string
	Command  []string
	Parallel bool
}

type ExecResult struct {
	Repository string
	Dir        string
	ExitCode   int
	Output     []byte
	Duration   time.Duration
}

func (s *FSStore) Exec(ctx context.Context, handle string, opts ExecOptions) ([]ExecResult, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	var results []ExecResult

	if len(opts.Command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	if opts.Target == "" && len(ws.Repositories) == 0 {
		opts.Target = "root"
	}

	switch opts.Target {
	case "", "all":
		for _, repo := range ws.Repositories {
			result, err := s.execInRepository(ctx, repo, ws.Path, opts.Command)
			results = append(results, result)
			if err != nil {
				return results, err
			}
			if result.ExitCode != 0 {
				return results, fmt.Errorf("command failed in %s with exit code %d", repo.Name, result.ExitCode)
			}
		}
	case "root":
		result := ExecResult{
			Repository: "root",
			Dir:        ws.Path,
		}
		start := time.Now()
		cmd := exec.CommandContext(ctx, opts.Command[0], opts.Command[1:]...)
		cmd.Dir = ws.Path
		output, err := cmd.CombinedOutput()
		result.Duration = time.Since(start)

		result.Output = output
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			} else {
				result.ExitCode = 1
			}
		}
		results = append(results, result)
		if result.ExitCode != 0 {
			return results, fmt.Errorf("command failed with exit code %d", result.ExitCode)
		}
	default:
		repo := ws.GetRepositoryByName(opts.Target)
		if repo == nil {
			return nil, fmt.Errorf("repository not found: %s", opts.Target)
		}
		result, err := s.execInRepository(ctx, *repo, ws.Path, opts.Command)
		results = append(results, result)
		if err != nil {
			return results, err
		}
		if result.ExitCode != 0 {
			return results, fmt.Errorf("command failed in %s with exit code %d", repo.Name, result.ExitCode)
		}
	}

	return results, nil
}

func (s *FSStore) execInRepository(ctx context.Context, repo Repository, wsPath string, cmdArgs []string) (ExecResult, error) {
	if len(cmdArgs) == 0 {
		return ExecResult{}, errors.New("command cannot be empty")
	}

	repoDir := filepath.Join(wsPath, repo.Name)
	result := ExecResult{
		Repository: repo.Name,
		Dir:        repoDir,
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)

	result.Output = output

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, err
	}

	return result, nil
}

func (s *FSStore) GetRepositoryPath(ctx context.Context, handle, repoName string) (string, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return "", err
	}

	repo := ws.GetRepositoryByName(repoName)
	if repo == nil {
		return "", fmt.Errorf("repository not found: %s", repoName)
	}

	return filepath.Join(ws.Path, repo.Name), nil
}

func (s *FSStore) workspaceDir(handle string) string {
	return filepath.Join(s.root, handle)
}

func isLocalPath(path string) bool {
	if path == "" {
		return false
	}

	return !strings.HasPrefix(path, "git@") && !strings.Contains(path, "://")
}

func validateLocalRepository(path, invocationCWD string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	expandedPath, err := expandPath(path, invocationCWD)
	if err != nil {
		return fmt.Errorf("expanding path: %w", err)
	}

	cleanedPath := filepath.Clean(expandedPath)

	info, err := os.Stat(cleanedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("accessing path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	gitDir := filepath.Join(cleanedPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository (missing .git directory): %s", path)
		}
		return fmt.Errorf("checking .git directory: %w", err)
	}

	return nil
}

func validateRepoURL(url, invocationCWD string) error {
	if url == "" {
		return errors.New("repository URL cannot be empty")
	}

	if isLocalPath(url) {
		return validateLocalRepository(url, invocationCWD)
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

	return fmt.Errorf("unsupported URL (expected https://, git@, ssh://, git://, or a local path)")
}

func validateTemplatePath(path string) error {
	if path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("accessing template: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}

	return nil
}

func validateRepositories(repos []RepositoryOption, invocationCWD string) error {
	seenURLs := make(map[string]bool)
	seenNames := make(map[string]bool)

	for _, repo := range repos {
		if err := validateRepoURL(repo.URL, invocationCWD); err != nil {
			return fmt.Errorf("invalid repository URL %s: %w", repo.URL, err)
		}

		if seenURLs[repo.URL] {
			return fmt.Errorf("duplicate repository URL: %s", repo.URL)
		}
		seenURLs[repo.URL] = true

		name := extractRepoName(repo.URL, invocationCWD)
		if seenNames[name] {
			return fmt.Errorf("duplicate repository name: %s", name)
		}
		seenNames[name] = true
	}

	return nil
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

func (s *FSStore) cloneRepo(ctx context.Context, repo Repository, wsDir, invocationCWD string) error {
	url := repo.URL
	ref := repo.Ref

	// Convert relative local paths to absolute for git clone
	if isLocalPath(url) {
		expandedPath, err := expandPath(url, invocationCWD)
		if err != nil {
			return fmt.Errorf("expanding local path: %w", err)
		}
		absPath, err := filepath.Abs(expandedPath)
		if err != nil {
			return fmt.Errorf("resolving absolute local path: %w", err)
		}

		// Auto-detect current branch for local repos when no ref specified
		if ref == "" {
			branch, err := s.git.CurrentBranch(ctx, absPath)
			if err != nil {
				return fmt.Errorf("detecting current branch: %w", err)
			}
			ref = branch
		}

		url = absPath
	}

	// Fallback to "main" for remote repos or if branch detection failed
	if ref == "" {
		ref = "main"
	}

	repoDir := filepath.Join(wsDir, repo.Name)

	if err := s.git.Clone(ctx, url, repoDir, git.CloneOptions{}); err != nil {
		return err
	}

	if err := s.git.Checkout(ctx, repoDir, ref); err != nil {
		return err
	}

	return nil
}

func (s *FSStore) cloneRepositories(ctx context.Context, repos []Repository, wsDir, invocationCWD string) error {
	for _, repo := range repos {
		if err := s.cloneRepo(ctx, repo, wsDir, invocationCWD); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
	}
	return nil
}

func extractRepoName(url, invocationCWD string) string {
	url = strings.TrimSuffix(url, ".git")

	if isLocalPath(url) {
		expandedPath, err := expandPath(url, invocationCWD)
		if err != nil {
			return filepath.Base(url)
		}
		absPath, err := filepath.Abs(expandedPath)
		if err == nil {
			return filepath.Base(absPath)
		}
		return filepath.Base(expandedPath)
	}

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

func (s *FSStore) applyTemplate(ctx context.Context, templatePath string, vars map[string]string, wsDir string) error {
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("resolving template path: %w", err)
	}

	return filepath.Walk(absTemplatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absTemplatePath, path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}
		if relPath == "." {
			return nil
		}

		substitutedPath := substituteVars(relPath, vars)

		dstPath := filepath.Join(wsDir, substitutedPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info.Mode())
	})
}

func expandPath(path, invocationCWD string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting user home directory: %w", err)
		}
		return filepath.Join(homeDir, path[1:]), nil
	}

	if !filepath.IsAbs(path) {
		return filepath.Abs(filepath.Join(invocationCWD, path))
	}
	return path, nil
}

func substituteVars(path string, vars map[string]string) string {
	result := path
	for key, value := range vars {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("opening destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}

	return nil
}

func (s *FSStore) RecordExecution(ctx context.Context, handle string, record ExecutionRecord) error {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	workshedDir := filepath.Join(ws.Path, ".workshed")
	executionsDir := filepath.Join(workshedDir, executionsDirName)

	if err := os.MkdirAll(executionsDir, 0755); err != nil {
		return fmt.Errorf("creating executions directory: %w", err)
	}

	execDir := filepath.Join(executionsDir, record.ID)
	if err := os.MkdirAll(execDir, 0755); err != nil {
		return fmt.Errorf("creating execution directory: %w", err)
	}

	stdoutDir := filepath.Join(execDir, "stdout")
	stderrDir := filepath.Join(execDir, "stderr")
	if err := os.MkdirAll(stdoutDir, 0755); err != nil {
		return fmt.Errorf("creating stdout directory: %w", err)
	}
	if err := os.MkdirAll(stderrDir, 0755); err != nil {
		return fmt.Errorf("creating stderr directory: %w", err)
	}

	for i := range record.Results {
		result := &record.Results[i]
		if result.Repository != "" && result.Repository != "root" {
			result.OutputPath = filepath.Join(result.Repository + ".txt")
		}
	}

	record.Handle = handle
	record.Timestamp = time.Now()
	record.StartedAt = record.Timestamp

	recordPath := filepath.Join(execDir, "record.json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling execution record: %w", err)
	}
	if err := os.WriteFile(recordPath, data, 0644); err != nil {
		return fmt.Errorf("writing execution record: %w", err)
	}

	return nil
}

func (s *FSStore) GetExecution(ctx context.Context, handle, execID string) (*ExecutionRecord, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	execPath := filepath.Join(ws.Path, ".workshed", executionsDirName, execID, "record.json")
	data, err := os.ReadFile(execPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("execution not found: %s", execID)
		}
		return nil, fmt.Errorf("reading execution record: %w", err)
	}

	var record ExecutionRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("parsing execution record: %w", err)
	}

	return &record, nil
}

func (s *FSStore) ListExecutions(ctx context.Context, handle string, opts ListExecutionsOptions) ([]ExecutionRecord, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	executionsDir := filepath.Join(ws.Path, ".workshed", executionsDirName)
	entries, err := os.ReadDir(executionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ExecutionRecord{}, nil
		}
		return nil, fmt.Errorf("reading executions directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, entry.Name())
		}
	}

	if !opts.Reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	} else {
		sort.Strings(ids)
	}

	var records []ExecutionRecord
	for i, id := range ids {
		if i < opts.Offset {
			continue
		}
		if opts.Limit > 0 && len(records) >= opts.Limit {
			break
		}

		record, err := s.GetExecution(ctx, handle, id)
		if err != nil {
			continue
		}
		records = append(records, *record)
	}

	return records, nil
}

func (s *FSStore) CaptureState(ctx context.Context, handle string, opts CaptureOptions) (*Capture, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	if opts.Kind == "" && opts.Description == "" && len(opts.Tags) == 0 {
		return nil, fmt.Errorf("capture must have intent: provide --kind, --description, or --tag")
	}

	workshedDir := filepath.Join(ws.Path, ".workshed")
	capturesDir := filepath.Join(workshedDir, capturesDirName)

	if err := os.MkdirAll(capturesDir, 0755); err != nil {
		return nil, fmt.Errorf("creating captures directory: %w", err)
	}

	id := ulid.Make()
	captureDir := filepath.Join(capturesDir, id.String())

	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(captureDir)
		}
	}()

	if err := os.MkdirAll(captureDir, 0755); err != nil {
		return nil, fmt.Errorf("creating capture directory: %w", err)
	}

	capture := &Capture{
		ID:        id.String(),
		Timestamp: time.Now(),
		Handle:    handle,
		Name:      opts.Name,
		Kind:      opts.Kind,
		GitState:  make([]GitRef, 0, len(ws.Repositories)),
		Metadata: CaptureMetadata{
			Description: opts.Description,
			Tags:        opts.Tags,
			Custom:      opts.Custom,
		},
	}

	for _, repo := range ws.Repositories {
		repoDir := filepath.Join(ws.Path, repo.Name)
		ref, err := s.gitState(ctx, repoDir)
		if err != nil {
			return nil, fmt.Errorf("getting git state for %s: %w", repo.Name, err)
		}
		ref.Repository = repo.Name
		capture.GitState = append(capture.GitState, *ref)
	}

	capturePath := filepath.Join(captureDir, "capture.json")
	data, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling capture: %w", err)
	}
	if err := os.WriteFile(capturePath, data, 0644); err != nil {
		return nil, fmt.Errorf("writing capture: %w", err)
	}

	success = true
	return capture, nil
}

func (s *FSStore) gitState(ctx context.Context, dir string) (*GitRef, error) {
	ref := &GitRef{}

	commit, err := s.git.RevParse(ctx, dir, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("getting commit: %w", err)
	}
	ref.Commit = commit

	branch, _ := s.git.CurrentBranch(ctx, dir)
	ref.Branch = branch

	output, err := s.git.StatusPorcelain(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("getting status: %w", err)
	}
	ref.Status = strings.TrimSpace(output)
	ref.Dirty = strings.TrimSpace(output) != ""

	return ref, nil
}

func (s *FSStore) ApplyCapture(ctx context.Context, handle string, captureID string) error {
	result, err := s.PreflightApply(ctx, handle, captureID)
	if err != nil {
		return err
	}
	if !result.Valid {
		return fmt.Errorf("apply blocked by preflight errors")
	}

	capture, err := s.GetCapture(ctx, handle, captureID)
	if err != nil {
		return err
	}

	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	for _, ref := range capture.GitState {
		repoDir := filepath.Join(ws.Path, ref.Repository)
		if err := s.git.Checkout(ctx, repoDir, ref.Commit); err != nil {
			return fmt.Errorf("checking out %s to %s: %w", ref.Repository, ref.Commit, err)
		}
	}

	return nil
}

func (s *FSStore) PreflightApply(ctx context.Context, handle string, captureID string) (ApplyPreflightResult, error) {
	result := ApplyPreflightResult{Valid: true}

	capture, err := s.GetCapture(ctx, handle, captureID)
	if err != nil {
		return ApplyPreflightResult{}, err
	}

	ws, err := s.Get(ctx, handle)
	if err != nil {
		return ApplyPreflightResult{}, err
	}

	repoSet := make(map[string]bool)
	for _, repo := range ws.Repositories {
		repoSet[repo.Name] = true
	}

	for _, ref := range capture.GitState {
		repoDir := filepath.Join(ws.Path, ref.Repository)

		if !repoSet[ref.Repository] {
			result.Valid = false
			result.Errors = append(result.Errors, ApplyPreflightError{
				Repository: ref.Repository,
				Reason:     ReasonMissingRepository,
				Details:    "repository defined in capture does not exist in workspace",
			})
			continue
		}

		gitDir := filepath.Join(repoDir, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ApplyPreflightError{
				Repository: ref.Repository,
				Reason:     ReasonRepositoryNotGit,
				Details:    "directory is not a git repository",
			})
			continue
		}

		dirty, _ := s.git.StatusPorcelain(ctx, repoDir)
		if strings.TrimSpace(dirty) != "" {
			result.Valid = false
			result.Errors = append(result.Errors, ApplyPreflightError{
				Repository: ref.Repository,
				Reason:     ReasonDirtyWorkingTree,
				Details:    "working tree has uncommitted changes",
			})
		}
	}

	return result, nil
}

func (s *FSStore) GetCapture(ctx context.Context, handle, captureID string) (*Capture, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	capturePath := filepath.Join(ws.Path, ".workshed", capturesDirName, captureID, "capture.json")
	data, err := os.ReadFile(capturePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("capture not found: %s", captureID)
		}
		return nil, fmt.Errorf("reading capture: %w", err)
	}

	var capture Capture
	if err := json.Unmarshal(data, &capture); err != nil {
		return nil, fmt.Errorf("parsing capture: %w", err)
	}

	return &capture, nil
}

func (s *FSStore) ListCaptures(ctx context.Context, handle string) ([]Capture, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	capturesDir := filepath.Join(ws.Path, ".workshed", capturesDirName)
	entries, err := os.ReadDir(capturesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Capture{}, nil
		}
		return nil, fmt.Errorf("reading captures directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, entry.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(ids)))

	var captures []Capture
	for _, id := range ids {
		capture, err := s.GetCapture(ctx, handle, id)
		if err != nil {
			continue
		}
		captures = append(captures, *capture)
	}

	return captures, nil
}

func (s *FSStore) DeriveContext(ctx context.Context, handle string) (*WorkspaceContext, error) {
	ws, err := s.Get(ctx, handle)
	if err != nil {
		return nil, err
	}

	executions, err := s.ListExecutions(ctx, handle, ListExecutionsOptions{Limit: 1})
	if err != nil {
		return nil, err
	}

	captures, err := s.ListCaptures(ctx, handle)
	if err != nil {
		return nil, err
	}

	repos := make([]ContextRepo, len(ws.Repositories))
	for i, repo := range ws.Repositories {
		repos[i] = ContextRepo{
			Name:     repo.Name,
			Path:     filepath.Join(ws.Path, repo.Name),
			URL:      repo.URL,
			RootPath: repo.Name,
		}
	}

	var lastExecuted *time.Time
	if len(executions) > 0 {
		t := executions[0].Timestamp
		lastExecuted = &t
	}

	var lastCaptured *time.Time
	if len(captures) > 0 {
		lastCaptured = &captures[0].Timestamp
	}

	// Note: captures are ordered newest-first (see ListCaptures)
	contextCaptures := make([]ContextCapture, 0, len(captures))
	for _, cap := range captures {
		contextCaptures = append(contextCaptures, ContextCapture{
			ID:          cap.ID,
			Timestamp:   cap.Timestamp,
			Name:        cap.Name,
			Kind:        cap.Kind,
			Description: cap.Metadata.Description,
			Tags:        cap.Metadata.Tags,
			RepoCount:   len(cap.GitState),
		})
	}

	return &WorkspaceContext{
		Version:      ContextVersion,
		GeneratedAt:  time.Now(),
		Handle:       handle,
		Purpose:      ws.Purpose,
		Repositories: repos,
		Captures:     contextCaptures,
		Metadata: ContextMetadata{
			WorkshedVersion: "0.3.0",
			ExecutionsCount: len(executions),
			CapturesCount:   len(captures),
			LastExecutedAt:  lastExecuted,
			LastCapturedAt:  lastCaptured,
		},
	}, nil
}

func (s *FSStore) ValidateAgents(ctx context.Context, handle string, agentsPath string) (AgentsValidationResult, error) {
	result := AgentsValidationResult{
		Valid:       true,
		Errors:      make([]AgentsError, 0),
		Warnings:    make([]AgentsWarning, 0),
		Sections:    make([]AgentsSection, 0),
		Explanation: "",
	}

	data, err := os.ReadFile(agentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return AgentsValidationResult{
				Valid:       false,
				Explanation: "AGENTS.md file not found at " + agentsPath,
			}, nil
		}
		return AgentsValidationResult{}, fmt.Errorf("reading AGENTS.md: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	sectionPattern := regexp.MustCompile(`^##\s+(.+)$`)
	subsections := []string{"Running", "Philosophy", "Code Guidelines", "Testing Philosophy", "Design Smells to Watch For", "Final Note"}

	var currentSection string
	sectionStartLine := 0
	sectionContent := make(map[string][]string)
	seenSections := make(map[string]bool)

	for i, line := range lines {
		match := sectionPattern.FindStringSubmatch(line)
		if match != nil {
			if currentSection != "" {
				seenSections[currentSection] = true
				result.Sections = append(result.Sections, AgentsSection{
					Name:     currentSection,
					Line:     sectionStartLine + 1,
					Valid:    true,
					Warnings: 0,
					Errors:   0,
				})
			}
			currentSection = match[1]
			sectionStartLine = i
			sectionContent[currentSection] = make([]string, 0)
		} else if currentSection != "" {
			sectionContent[currentSection] = append(sectionContent[currentSection], strings.TrimSpace(line))
		}
	}

	if currentSection != "" {
		seenSections[currentSection] = true
		result.Sections = append(result.Sections, AgentsSection{
			Name:     currentSection,
			Line:     sectionStartLine + 1,
			Valid:    true,
			Warnings: 0,
			Errors:   0,
		})
	}

	for _, expected := range subsections {
		if !seenSections[expected] {
			result.Valid = false
			result.Errors = append(result.Errors, AgentsError{
				Line:    0,
				Message: "Missing required section: " + expected,
				Field:   "structure",
			})
		}
	}

	sectionCount := len(result.Sections)
	if sectionCount < 6 {
		result.Valid = false
		result.Explanation = fmt.Sprintf("AGENTS.md has %d sections, expected at least 6 (Running, Philosophy, Code Guidelines, Testing Philosophy, Design Smells to Watch For, Final Note)", sectionCount)
	} else {
		result.Explanation = fmt.Sprintf("AGENTS.md has %d sections with required sections present.", sectionCount)
	}

	return result, nil
}
