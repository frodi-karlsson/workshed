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
	"strings"
	"time"

	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/handle"
)

const metadataFileName = ".workshed.json"

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
		if err := validateRepositories(repos); err != nil {
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
		// Convert relative local paths to absolute for metadata storage
		if isLocalPath(url) {
			absPath, err := filepath.Abs(url)
			if err != nil {
				return nil, fmt.Errorf("resolving local path %s: %w", url, err)
			}
			url = absPath
		}

		clonedRepos[i] = Repository{
			URL:  url,
			Ref:  opt.Ref,
			Name: extractRepoName(opt.URL),
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

	if err := s.cloneRepositories(ctx, clonedRepos, tmpDir); err != nil {
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
func (s *FSStore) AddRepository(ctx context.Context, handle string, repo RepositoryOption) error {
	return s.AddRepositories(ctx, handle, []RepositoryOption{repo})
}

// AddRepositories adds multiple repositories to an existing workspace.
func (s *FSStore) AddRepositories(ctx context.Context, handle string, repos []RepositoryOption) error {
	if len(repos) == 0 {
		return errors.New("no repositories specified")
	}

	ws, err := s.Get(ctx, handle)
	if err != nil {
		return err
	}

	if err := validateRepositories(repos); err != nil {
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
		name := extractRepoName(opt.URL)
		if seenNames[name] {
			return fmt.Errorf("repository name already exists: %s", name)
		}
	}

	clonedRepos := make([]Repository, len(repos))
	for i, opt := range repos {
		url := opt.URL
		if isLocalPath(url) {
			absPath, err := filepath.Abs(url)
			if err != nil {
				return fmt.Errorf("resolving local path %s: %w", url, err)
			}
			url = absPath
		}

		clonedRepos[i] = Repository{
			URL:  url,
			Ref:  opt.Ref,
			Name: extractRepoName(opt.URL),
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
		if err := s.cloneRepo(ctx, repo, ws.Path); err != nil {
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

	if strings.HasPrefix(path, "git@") {
		return false
	}

	schemeEnd := strings.Index(path, "://")
	if schemeEnd != -1 {
		return false
	}

	if strings.HasPrefix(path, "/") || strings.Contains(path, string(filepath.Separator)) {
		return true
	}

	if path == "." || path == ".." {
		return true
	}

	return false
}

func validateLocalRepository(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("accessing path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository (missing .git directory): %s", path)
		}
		return fmt.Errorf("checking .git directory: %w", err)
	}

	return nil
}

func validateRepoURL(url string) error {
	if url == "" {
		return errors.New("repository URL cannot be empty")
	}

	if isLocalPath(url) {
		return validateLocalRepository(url)
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

func validateRepositories(repos []RepositoryOption) error {
	seenURLs := make(map[string]bool)
	seenNames := make(map[string]bool)

	for _, repo := range repos {
		if err := validateRepoURL(repo.URL); err != nil {
			return fmt.Errorf("invalid repository URL %s: %w", repo.URL, err)
		}

		if seenURLs[repo.URL] {
			return fmt.Errorf("duplicate repository URL: %s", repo.URL)
		}
		seenURLs[repo.URL] = true

		name := extractRepoName(repo.URL)
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

func (s *FSStore) cloneRepo(ctx context.Context, repo Repository, wsDir string) error {
	url := repo.URL
	ref := repo.Ref

	// Convert relative local paths to absolute for git clone
	if isLocalPath(url) {
		absPath, err := filepath.Abs(url)
		if err != nil {
			return fmt.Errorf("resolving local path: %w", err)
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

func (s *FSStore) cloneRepositories(ctx context.Context, repos []Repository, wsDir string) error {
	for _, repo := range repos {
		if err := s.cloneRepo(ctx, repo, wsDir); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
	}
	return nil
}

func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, ".git")

	if isLocalPath(url) {
		// Resolve relative paths like "." or ".." to get actual directory name
		if url == "." || url == ".." || strings.HasPrefix(url, "./") || strings.HasPrefix(url, "../") {
			absPath, err := filepath.Abs(url)
			if err == nil {
				return filepath.Base(absPath)
			}
		}
		return filepath.Base(url)
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
