package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/version"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func detectShell() (string, error) {
	if shell := os.Getenv("SHELL"); shell != "" {
		if _, err := os.Stat(shell); err == nil {
			return shell, nil
		}
	}

	for _, shell := range []string{"/bin/bash", "/bin/zsh", "/bin/sh"} {
		if _, err := os.Stat(shell); err == nil {
			return shell, nil
		}
	}

	return "", fmt.Errorf("no suitable shell found")
}

type Server struct {
	store        workspace.Store
	activeHandle *string
}

func NewServer(store workspace.Store) *Server {
	return &Server{store: store}
}

func (s *Server) resolveHandle(ctx context.Context, handle *string) (string, error) {
	if handle != nil {
		return *handle, nil
	}
	if s.activeHandle == nil {
		return "", NewToolError("no active workspace. Use enter_workspace({handle: \"...\"}) to set one, or pass handle explicitly to this command.")
	}
	if _, err := s.store.Get(ctx, *s.activeHandle); err != nil {
		return "", err
	}
	return *s.activeHandle, nil
}

func (s *Server) availableWorkspaces(ctx context.Context) ([]string, error) {
	list, err := s.store.List(ctx, workspace.ListOptions{})
	if err != nil {
		return nil, err
	}
	handles := make([]string, 0, len(list))
	for _, ws := range list {
		handles = append(handles, ws.Handle)
	}
	return handles, nil
}

func (s *Server) workspaceNotFoundError(ctx context.Context, handle string) error {
	handles, err := s.availableWorkspaces(ctx)
	if err != nil {
		return NewToolError(fmt.Sprintf("workspace %q not found", handle))
	}
	if len(handles) == 0 {
		return NewToolError(fmt.Sprintf("workspace %q not found. No workspaces exist - use create_workspace to create one", handle))
	}
	return NewToolError(fmt.Sprintf("workspace %q not found. Available: %s", handle, strings.Join(handles, ", ")))
}

func (s *Server) captureNotFoundError(ctx context.Context, handle, captureID string) error {
	captures, err := s.store.ListCaptures(ctx, handle)
	if err != nil {
		return NewToolError(fmt.Sprintf("capture %q not found in workspace %q", captureID, handle))
	}
	if len(captures) == 0 {
		return NewToolError(fmt.Sprintf("capture %q not found in workspace %q. No captures exist", captureID, handle))
	}
	captureNames := make([]string, 0, len(captures))
	for _, c := range captures {
		captureNames = append(captureNames, c.Name)
	}
	return NewToolError(fmt.Sprintf("capture %q not found in workspace %q. Available: %s", captureID, handle, strings.Join(captureNames, ", ")))
}

func (s *Server) listWorkspaces(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ListWorkspacesOutput, error) {
	workspaces, err := s.store.List(ctx, workspace.ListOptions{})
	if err != nil {
		return nil, ListWorkspacesOutput{}, err
	}

	result := make([]WorkspaceInfo, 0, len(workspaces))
	for _, ws := range workspaces {
		result = append(result, WorkspaceInfo{
			Handle:    ws.Handle,
			Purpose:   ws.Purpose,
			RepoCount: len(ws.Repositories),
			CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return nil, ListWorkspacesOutput{Workspaces: result}, nil
}

func (s *Server) getWorkspace(ctx context.Context, req *mcp.CallToolRequest, input GetWorkspaceInput) (*mcp.CallToolResult, WorkspaceDetail, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, WorkspaceDetail{}, err
	}
	ws, err := s.store.Get(ctx, handle)
	if err != nil {
		return nil, WorkspaceDetail{}, s.workspaceNotFoundError(ctx, handle)
	}

	repos := make([]RepositoryInfo, 0, len(ws.Repositories))
	for _, r := range ws.Repositories {
		repos = append(repos, RepositoryInfo{
			Name: r.Name,
			URL:  r.URL,
			Ref:  r.Ref,
			Path: ws.Path + "/" + r.Name,
		})
	}

	return nil, WorkspaceDetail{
		Handle:       ws.Handle,
		Purpose:      ws.Purpose,
		Repositories: repos,
		CreatedAt:    ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *Server) createWorkspace(ctx context.Context, req *mcp.CallToolRequest, input CreateWorkspaceInput) (*mcp.CallToolResult, CreateWorkspaceOutput, error) {
	if input.Purpose == "" {
		return nil, CreateWorkspaceOutput{}, NewToolError("purpose is required. Provide a brief description of this workspace's purpose.\nExample: {purpose: \"Debug payment timeout\"}")
	}

	repoOpts := make([]workspace.RepositoryOption, 0, len(input.Repos))
	for _, repo := range input.Repos {
		url, ref, repoDepth := workspace.ParseRepoFlag(repo)
		d := input.Depth
		if repoDepth > 0 {
			d = repoDepth
		}
		repoOpts = append(repoOpts, workspace.RepositoryOption{URL: url, Ref: ref, Depth: d})
	}

	templateVars := make(map[string]string)
	for _, v := range input.TemplateVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			templateVars[parts[0]] = parts[1]
		}
	}

	ws, err := s.store.Create(ctx, workspace.CreateOptions{
		Purpose:      input.Purpose,
		Template:     input.Template,
		TemplateVars: templateVars,
		Repositories: repoOpts,
	})
	if err != nil {
		return nil, CreateWorkspaceOutput{}, err
	}

	repos := make([]RepositoryInfo, 0, len(ws.Repositories))
	for _, r := range ws.Repositories {
		repos = append(repos, RepositoryInfo{
			Name: r.Name,
			URL:  r.URL,
			Ref:  r.Ref,
			Path: ws.Path + "/" + r.Name,
		})
	}

	return nil, CreateWorkspaceOutput{
		Handle:       ws.Handle,
		Purpose:      ws.Purpose,
		Path:         ws.Path,
		Repositories: repos,
	}, nil
}

func (s *Server) removeWorkspace(ctx context.Context, req *mcp.CallToolRequest, input RemoveWorkspaceInput) (*mcp.CallToolResult, RemoveWorkspaceOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, RemoveWorkspaceOutput{}, err
	}

	ws, err := s.store.Get(ctx, handle)
	if err != nil {
		return nil, RemoveWorkspaceOutput{}, s.workspaceNotFoundError(ctx, handle)
	}

	if input.DryRun {
		return nil, RemoveWorkspaceOutput{
			WouldDelete:   true,
			ReposCount:    len(ws.Repositories),
			CapturesCount: 0,
			Message:       fmt.Sprintf("This will permanently delete workspace %q with %d repositories.", handle, len(ws.Repositories)),
		}, nil
	}

	if s.activeHandle != nil && *s.activeHandle == handle {
		s.activeHandle = nil
	}

	err = s.store.Remove(ctx, handle)
	if err != nil {
		return nil, RemoveWorkspaceOutput{}, s.workspaceNotFoundError(ctx, handle)
	}

	return nil, RemoveWorkspaceOutput{
		Success: true,
		Message: "Workspace removed",
	}, nil
}

func (s *Server) execCommand(ctx context.Context, req *mcp.CallToolRequest, input ExecCommandInput) (*mcp.CallToolResult, ExecCommandOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, ExecCommandOutput{}, err
	}
	if len(input.Command) == 0 {
		return nil, ExecCommandOutput{}, NewToolError("command is required. Provide an array of command and arguments.\nExample: {command: [\"make\", \"test\"]}")
	}

	shellPath, _ := detectShell()
	command := []string{shellPath, "-c", strings.Join(input.Command, " ")}

	execCtx := ctx
	var cancel context.CancelFunc
	if input.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(input.Timeout)*time.Millisecond)
		defer cancel()
	}

	opts := workspace.ExecOptions{
		Command:  command,
		Target:   input.Repo,
		Parallel: input.All,
	}

	results, err := s.store.Exec(execCtx, handle, opts)
	if err != nil {
		return nil, ExecCommandOutput{}, err
	}

	maxExitCode := 0
	resultInfos := make([]ExecResultInfo, 0, len(results))
	for _, r := range results {
		if r.ExitCode > maxExitCode {
			maxExitCode = r.ExitCode
		}
		output := strings.TrimSpace(string(r.Output))
		if input.OutputLimit > 0 && len(output) > input.OutputLimit {
			output = output[:input.OutputLimit] + "\n... (output truncated)"
		}
		resultInfos = append(resultInfos, ExecResultInfo{
			Repository: r.Repository,
			ExitCode:   r.ExitCode,
			Output:     output,
			DurationMs: r.Duration.Milliseconds(),
		})
	}

	return nil, ExecCommandOutput{
		Success:   maxExitCode == 0,
		ExitCode:  maxExitCode,
		Results:   resultInfos,
		TotalTime: 0,
	}, nil
}

func (s *Server) captureState(ctx context.Context, req *mcp.CallToolRequest, input CaptureStateInput) (*mcp.CallToolResult, CaptureStateOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, CaptureStateOutput{}, err
	}
	if input.Name == "" {
		return nil, CaptureStateOutput{}, NewToolError("name is required. Provide a descriptive name for this state capture.\nExample: {name: \"Before refactoring\"}")
	}

	capture, err := s.store.CaptureState(ctx, handle, workspace.CaptureOptions{
		Name:        input.Name,
		Description: input.Description,
		Tags:        input.Tags,
	})
	if err != nil {
		return nil, CaptureStateOutput{}, err
	}

	gitState := make([]GitRefInfo, 0, len(capture.GitState))
	for _, g := range capture.GitState {
		gitState = append(gitState, GitRefInfo{
			Repository: g.Repository,
			Branch:     g.Branch,
			Commit:     g.Commit,
			Dirty:      g.Dirty,
		})
	}

	return nil, CaptureStateOutput{
		ID:        capture.ID,
		Name:      capture.Name,
		Timestamp: capture.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		GitState:  gitState,
	}, nil
}

func (s *Server) enterWorkspace(ctx context.Context, req *mcp.CallToolRequest, input EnterWorkspaceInput) (*mcp.CallToolResult, EnterWorkspaceOutput, error) {
	if input.Handle == nil {
		return nil, EnterWorkspaceOutput{}, NewToolError("handle is required. Use list_workspaces() to see available workspaces.")
	}
	ws, err := s.store.Get(ctx, *input.Handle)
	if err != nil {
		return nil, EnterWorkspaceOutput{}, NewToolError(fmt.Sprintf("workspace %q not found. Use list_workspaces() to see available workspaces.", *input.Handle))
	}
	s.activeHandle = input.Handle
	return nil, EnterWorkspaceOutput{
		Handle: *input.Handle,
		Path:   ws.Path,
	}, nil
}

func (s *Server) exitWorkspace(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ExitWorkspaceOutput, error) {
	s.activeHandle = nil
	return nil, ExitWorkspaceOutput{Message: "Exited active workspace"}, nil
}

func (s *Server) help(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, HelpOutput, error) {
	return nil, HelpOutput{
		Message: `# Workshed Concepts

## Key Parameters

### handle (Workspace Identifier)
A unique random identifier for a workspace (e.g., "aquatic-fish-motion"). Use list_workspaces() to see available workspaces.

### repo (Repository Name)
The name of a repository within a workspace (not the full URL). Find available repos via get_workspace().

### Active Workspace
Call enter_workspace({handle: "..."}) once to set an active workspace. Subsequent commands can omit the 'handle' parameter.

Use exit_workspace() to clear the active workspace when you're done.

## Use Cases

### Create a new workspace with multiple repositories

1. create_workspace({purpose: "My project", repos: ["github.com/org/repo1@main", "github.com/org/repo2@main"]})
2. enter_workspace({handle: "..."})
3. exec_command({command: ["make", "setup"]})

### Safely experiment with code (backup and restore)

1. enter_workspace({handle: "..."})
2. capture_state({name: "Before changes", description: "State before refactoring"})
3. exec_command({command: ["make", "changes"]})
4. If failed: apply_capture({capture_id: "..."})
5. If successful: capture_state({name: "After changes"})

### Run tests across all repositories

1. enter_workspace({handle: "..."})
2. exec_command({command: ["make", "test"], all: true})

### Run commands in a specific repository

1. enter_workspace({handle: "..."})
2. exec_command({command: ["npm", "test"], repo: "myrepo"})

### Export and import a workspace (backup/sharing)

1. export_workspace({})
2. import_workspace({context: {...}, preserve_handle: true})

### Add a repository to an existing workspace

1. enter_workspace({handle: "..."})
2. add_repository({repo: "github.com/org/newrepo@main"})
3. exec_command({command: ["git", "submodule", "update", "--init"], repo: "newrepo"})

### Get the workspace path for your IDE

1. enter_workspace({handle: "..."})
2. get_workspace_path({})
3. get_workspace_repo_path({repo_name: "myrepo"})`,
	}, nil
}

func (s *Server) listCaptures(ctx context.Context, req *mcp.CallToolRequest, input ListCapturesInput) (*mcp.CallToolResult, ListCapturesOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, ListCapturesOutput{}, err
	}

	captures, err := s.store.ListCaptures(ctx, handle)
	if err != nil {
		return nil, ListCapturesOutput{}, err
	}

	result := make([]CaptureInfo, 0, len(captures))
	for _, c := range captures {
		result = append(result, CaptureInfo{
			ID:          c.ID,
			Name:        c.Name,
			Timestamp:   c.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Description: c.Metadata.Description,
			Tags:        c.Metadata.Tags,
			RepoCount:   len(c.GitState),
		})
	}

	return nil, ListCapturesOutput{Captures: result}, nil
}

func (s *Server) applyCapture(ctx context.Context, req *mcp.CallToolRequest, input ApplyCaptureInput) (*mcp.CallToolResult, ApplyCaptureOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, ApplyCaptureOutput{}, err
	}

	if input.CaptureID == "" {
		return nil, ApplyCaptureOutput{}, NewToolError("capture_id is required. Use list_captures() to see available captures.")
	}

	if input.DryRun {
		result, err := s.store.PreflightApply(ctx, handle, input.CaptureID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, ApplyCaptureOutput{}, s.captureNotFoundError(ctx, handle, input.CaptureID)
			}
			return nil, ApplyCaptureOutput{}, err
		}
		if !result.Valid {
			errors := make([]string, 0, len(result.Errors))
			for _, e := range result.Errors {
				errors = append(errors, fmt.Sprintf("%s: %s (%s)", e.Repository, e.Reason, e.Details))
			}
			return nil, ApplyCaptureOutput{
				Success: false,
				Message: "Preflight check failed",
				Errors:  errors,
			}, nil
		}
		return nil, ApplyCaptureOutput{
			Success: true,
			Message: "Preflight check passed - would apply cleanly",
		}, nil
	}

	err = s.store.ApplyCapture(ctx, handle, input.CaptureID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ApplyCaptureOutput{}, s.captureNotFoundError(ctx, handle, input.CaptureID)
		}
		return nil, ApplyCaptureOutput{}, err
	}

	return nil, ApplyCaptureOutput{
		Success: true,
		Message: "Capture applied successfully",
	}, nil
}

func (s *Server) exportWorkspace(ctx context.Context, req *mcp.CallToolRequest, input ExportWorkspaceInput) (*mcp.CallToolResult, ExportWorkspaceOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, ExportWorkspaceOutput{}, err
	}

	ctxData, err := s.store.ExportContext(ctx, handle)
	if err != nil {
		return nil, ExportWorkspaceOutput{}, err
	}

	if input.Compact {
		ctxData.Captures = nil
	}

	data, err := json.MarshalIndent(ctxData, "", "  ")
	if err != nil {
		return nil, ExportWorkspaceOutput{}, err
	}

	return nil, ExportWorkspaceOutput{
		JSON: string(data),
		Metadata: ExportMetadata{
			Handle:       ctxData.Handle,
			Purpose:      ctxData.Purpose,
			RepoCount:    len(ctxData.Repositories),
			CaptureCount: len(ctxData.Captures),
		},
	}, nil
}

func (s *Server) importWorkspace(ctx context.Context, req *mcp.CallToolRequest, input ImportWorkspaceInput) (*mcp.CallToolResult, ImportWorkspaceOutput, error) {
	if input.Context == nil {
		return nil, ImportWorkspaceOutput{}, NewToolError("context is required. Use export_workspace() to get a valid workspace context, then pass it here.")
	}

	ctxBytes, err := json.Marshal(input.Context)
	if err != nil {
		return nil, ImportWorkspaceOutput{}, NewToolError(fmt.Sprintf("invalid context: %v", err))
	}

	var ctxData workspace.WorkspaceContext
	if err := json.Unmarshal(ctxBytes, &ctxData); err != nil {
		return nil, ImportWorkspaceOutput{}, NewToolError(fmt.Sprintf("invalid context: %v. Use export_workspace for valid format", err))
	}

	ws, err := s.store.ImportContext(ctx, workspace.ImportOptions{
		Context:        &ctxData,
		PreserveHandle: input.PreserveHandle,
	})
	if err != nil {
		return nil, ImportWorkspaceOutput{}, err
	}

	return nil, ImportWorkspaceOutput{
		Handle:    ws.Handle,
		Purpose:   ws.Purpose,
		Path:      ws.Path,
		RepoCount: len(ws.Repositories),
	}, nil
}

func (s *Server) getWorkspacePath(ctx context.Context, req *mcp.CallToolRequest, input GetWorkspacePathInput) (*mcp.CallToolResult, GetWorkspacePathOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, err
	}

	path, err := s.store.Path(ctx, handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, err
	}

	return nil, GetWorkspacePathOutput{Path: path}, nil
}

func (s *Server) getWorkspaceRepoPath(ctx context.Context, req *mcp.CallToolRequest, input GetWorkspaceRepoPathInput) (*mcp.CallToolResult, GetWorkspacePathOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, err
	}

	if input.RepoName == "" {
		return nil, GetWorkspacePathOutput{}, NewToolError("repo_name is required. Use get_workspace() to see available repositories, then provide the repository name.")
	}

	ws, err := s.store.Get(ctx, handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, s.workspaceNotFoundError(ctx, handle)
	}

	repoPath := ws.Path + "/" + input.RepoName
	return nil, GetWorkspacePathOutput{Path: repoPath}, nil
}

func (s *Server) addRepository(ctx context.Context, req *mcp.CallToolRequest, input AddRepositoryInput) (*mcp.CallToolResult, AddRepositoryOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, AddRepositoryOutput{}, err
	}

	if input.Repo == "" {
		return nil, AddRepositoryOutput{}, NewToolError("repo is required. Provide a git repository URL with optional @ref.\nExample: {repo: \"github.com/org/repo@main\"}")
	}

	_, err = s.store.Get(ctx, handle)
	if err != nil {
		return nil, AddRepositoryOutput{}, s.workspaceNotFoundError(ctx, handle)
	}

	url, ref, repoDepth := workspace.ParseRepoFlag(input.Repo)
	d := input.Depth
	if repoDepth > 0 {
		d = repoDepth
	}
	invocationCWD := ""
	err = s.store.AddRepository(ctx, handle, workspace.RepositoryOption{URL: url, Ref: ref, Depth: d}, invocationCWD)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not a git repository") || strings.Contains(errMsg, "missing .git") {
			return nil, AddRepositoryOutput{}, NewToolError(fmt.Sprintf("failed to add repository %q: not a git repository (missing .git directory). Workshed requires git-tracked repositories.", input.Repo))
		}
		if strings.Contains(errMsg, "already exists") {
			return nil, AddRepositoryOutput{}, NewToolError(fmt.Sprintf("failed to add repository %q: repository already exists in workspace", input.Repo))
		}
		return nil, AddRepositoryOutput{}, NewToolError(fmt.Sprintf("failed to add repository: %v", err))
	}

	ws, _ := s.store.Get(ctx, handle)
	var newRepo *workspace.Repository
	for _, r := range ws.Repositories {
		if r.URL == url && (ref == "" || r.Ref == ref) {
			newRepo = &r
			break
		}
	}

	if newRepo == nil {
		return nil, AddRepositoryOutput{}, NewToolError("repository added but could not find it in workspace")
	}

	return nil, AddRepositoryOutput{
		Name: newRepo.Name,
		URL:  newRepo.URL,
		Ref:  newRepo.Ref,
		Path: ws.Path + "/" + newRepo.Name,
	}, nil
}

func (s *Server) removeRepository(ctx context.Context, req *mcp.CallToolRequest, input RemoveRepositoryInput) (*mcp.CallToolResult, RemoveRepositoryOutput, error) {
	handle, err := s.resolveHandle(ctx, input.Handle)
	if err != nil {
		return nil, RemoveRepositoryOutput{}, err
	}

	if input.RepoName == "" {
		return nil, RemoveRepositoryOutput{}, NewToolError("repo_name is required. Use get_workspace() to see available repositories.")
	}

	ws, err := s.store.Get(ctx, handle)
	if err != nil {
		return nil, RemoveRepositoryOutput{}, s.workspaceNotFoundError(ctx, handle)
	}

	found := false
	for _, r := range ws.Repositories {
		if r.Name == input.RepoName {
			found = true
			break
		}
	}
	if !found {
		names := make([]string, 0, len(ws.Repositories))
		for _, r := range ws.Repositories {
			names = append(names, r.Name)
		}
		return nil, RemoveRepositoryOutput{}, NewToolError(fmt.Sprintf("repository %q not found in workspace %q. Available: %s", input.RepoName, handle, strings.Join(names, ", ")))
	}

	err = s.store.RemoveRepository(ctx, handle, input.RepoName)
	if err != nil {
		return nil, RemoveRepositoryOutput{}, NewToolError(fmt.Sprintf("failed to remove repository: %v", err))
	}

	return nil, RemoveRepositoryOutput{
		Success: true,
		Message: fmt.Sprintf("Repository %q removed from workspace %q", input.RepoName, handle),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "workshed",
			Version: version.Version,
		},
		nil,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workspaces",
		Description: "List all Workshed workspaces with their handles, purposes, and repository counts. Use this to discover available workspaces.",
	}, s.listWorkspaces)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "help",
		Description: "Get example workflows and common use cases for working with Workshed workspaces.",
	}, s.help)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace",
		Description: "Get workspace details. Parameters: handle (workspace identifier, e.g., \"aquatic-fish-motion\"). If not provided, uses active workspace. Returns purpose, list of repositories (with their names, URLs, refs, and paths).",
	}, s.getWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_workspace",
		Description: "Create a new workspace. Parameters: purpose (required, brief description), repos (array of git URLs with optional @ref, e.g., \"github.com/org/repo@main\"), template, template_vars. Returns a new workspace handle (random identifier like \"aquatic-fish-motion\"), path, and repository details.",
	}, s.createWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "enter_workspace",
		Description: "Set the active workspace handle. All subsequent commands will use this workspace unless a handle is explicitly provided. Returns the workspace path.",
	}, s.enterWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "exit_workspace",
		Description: "Clear the active workspace handle. Subsequent commands will require an explicit handle.",
	}, s.exitWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_workspace",
		Description: "Delete a Workshed workspace by its handle. If handle is not provided, uses the active workspace (set with enter_workspace). Use with caution as this action cannot be undone.",
	}, s.removeWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "exec_command",
		Description: "Execute a command in a workspace. Parameters: handle (workspace identifier), repo (repository name), all (run in all repos), timeout (max milliseconds), output_limit (max output characters). Command runs in a shell with detected $SHELL, falling back to /bin/sh.",
	}, s.execCommand)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_state",
		Description: "Create a git state snapshot (capture) for a workspace. If handle is not provided, uses the active workspace (set with enter_workspace). Records branch, commit, and dirty status. Takes a name and optional description and tags.",
	}, s.captureState)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_captures",
		Description: "List all captures for a workspace. If handle is not provided, uses the active workspace (set with enter_workspace). Returns capture IDs, names, timestamps, descriptions, tags, and repo counts.",
	}, s.listCaptures)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "apply_capture",
		Description: "Apply (restore) git state from a capture. If handle is not provided, uses the active workspace (set with enter_workspace). Takes a capture ID. Set dry_run to true to check preflight without applying.",
	}, s.applyCapture)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_workspace",
		Description: "Export a workspace to portable JSON format. If handle is not provided, uses the active workspace (set with enter_workspace). Includes metadata, repository config, and optionally captures. Set compact to exclude captures.",
	}, s.exportWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "import_workspace",
		Description: "Create a workspace from exported context. Takes the context object from export_workspace output. Optionally preserve the original handle with preserve_handle. Repositories are cloned based on the exported configuration.",
	}, s.importWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace_path",
		Description: "Get the filesystem path for a workspace. If handle is not provided, uses the active workspace (set with enter_workspace). Returns the absolute path where the workspace directory is located.",
	}, s.getWorkspacePath)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace_repo_path",
		Description: "Get the filesystem path for a specific repository within a workspace. If handle is not provided, uses the active workspace (set with enter_workspace). Takes a repository name and returns its directory path.",
	}, s.getWorkspaceRepoPath)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_repository",
		Description: "Add a repository to an existing workspace. If handle is not provided, uses the active workspace (set with enter_workspace). Takes a repo URL with optional @ref (e.g., github.com/org/repo@main). Returns the added repository details.",
	}, s.addRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_repository",
		Description: "Remove a repository from a workspace by name. If handle is not provided, uses the active workspace (set with enter_workspace). Takes a repository name. Use get_workspace to see available repository names.",
	}, s.removeRepository)

	return server.Run(ctx, &mcp.StdioTransport{})
}
