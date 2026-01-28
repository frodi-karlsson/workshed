package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/frodi/workshed/internal/version"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	store workspace.Store
}

func NewServer(store workspace.Store) *Server {
	return &Server{store: store}
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
	if input.Handle == "" {
		return nil, WorkspaceDetail{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	ws, err := s.store.Get(ctx, input.Handle)
	if err != nil {
		return nil, WorkspaceDetail{}, s.workspaceNotFoundError(ctx, input.Handle)
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
		return nil, CreateWorkspaceOutput{}, NewToolError("purpose is required. Example: purpose: \"Debug payment timeout\"")
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
	if input.Handle == "" {
		return nil, RemoveWorkspaceOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}

	ws, err := s.store.Get(ctx, input.Handle)
	if err != nil {
		return nil, RemoveWorkspaceOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	if input.DryRun {
		return nil, RemoveWorkspaceOutput{
			WouldDelete:   true,
			ReposCount:    len(ws.Repositories),
			CapturesCount: 0,
			Message:       fmt.Sprintf("This will permanently delete workspace %q with %d repositories.", input.Handle, len(ws.Repositories)),
		}, nil
	}

	err = s.store.Remove(ctx, input.Handle)
	if err != nil {
		return nil, RemoveWorkspaceOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	return nil, RemoveWorkspaceOutput{
		Success: true,
		Message: "Workspace removed",
	}, nil
}

func (s *Server) execCommand(ctx context.Context, req *mcp.CallToolRequest, input ExecCommandInput) (*mcp.CallToolResult, ExecCommandOutput, error) {
	if input.Handle == "" {
		return nil, ExecCommandOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if len(input.Command) == 0 {
		return nil, ExecCommandOutput{}, NewToolError("command is required. Example: command: [\"make\", \"test\"]")
	}

	opts := workspace.ExecOptions{
		Command:  input.Command,
		Target:   input.Repo,
		Parallel: input.All,
	}

	results, err := s.store.Exec(ctx, input.Handle, opts)
	if err != nil {
		return nil, ExecCommandOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	maxExitCode := 0
	resultInfos := make([]ExecResultInfo, 0, len(results))
	for _, r := range results {
		if r.ExitCode > maxExitCode {
			maxExitCode = r.ExitCode
		}
		output := strings.TrimSpace(string(r.Output))
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
	if input.Handle == "" {
		return nil, CaptureStateOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if input.Name == "" {
		return nil, CaptureStateOutput{}, NewToolError("name is required. Example: name: \"Before refactoring\"")
	}

	capture, err := s.store.CaptureState(ctx, input.Handle, workspace.CaptureOptions{
		Name:        input.Name,
		Description: input.Description,
		Tags:        input.Tags,
	})
	if err != nil {
		return nil, CaptureStateOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
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

func (s *Server) listCaptures(ctx context.Context, req *mcp.CallToolRequest, input ListCapturesInput) (*mcp.CallToolResult, ListCapturesOutput, error) {
	if input.Handle == "" {
		return nil, ListCapturesOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}

	captures, err := s.store.ListCaptures(ctx, input.Handle)
	if err != nil {
		return nil, ListCapturesOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
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
	if input.Handle == "" {
		return nil, ApplyCaptureOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if input.CaptureID == "" {
		return nil, ApplyCaptureOutput{}, NewToolError("capture_id is required. Use list_captures to see available captures")
	}

	if input.DryRun {
		result, err := s.store.PreflightApply(ctx, input.Handle, input.CaptureID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, ApplyCaptureOutput{}, s.captureNotFoundError(ctx, input.Handle, input.CaptureID)
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

	err := s.store.ApplyCapture(ctx, input.Handle, input.CaptureID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ApplyCaptureOutput{}, s.captureNotFoundError(ctx, input.Handle, input.CaptureID)
		}
		return nil, ApplyCaptureOutput{}, err
	}

	return nil, ApplyCaptureOutput{
		Success: true,
		Message: "Capture applied successfully",
	}, nil
}

func (s *Server) exportWorkspace(ctx context.Context, req *mcp.CallToolRequest, input ExportWorkspaceInput) (*mcp.CallToolResult, ExportWorkspaceOutput, error) {
	if input.Handle == "" {
		return nil, ExportWorkspaceOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}

	ctxData, err := s.store.ExportContext(ctx, input.Handle)
	if err != nil {
		return nil, ExportWorkspaceOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
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
		return nil, ImportWorkspaceOutput{}, NewToolError("context is required. Use export_workspace to get a valid workspace context")
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
	if input.Handle == "" {
		return nil, GetWorkspacePathOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}

	path, err := s.store.Path(ctx, input.Handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	return nil, GetWorkspacePathOutput{Path: path}, nil
}

func (s *Server) getWorkspaceRepoPath(ctx context.Context, req *mcp.CallToolRequest, input struct {
	Handle   string `json:"handle"`
	RepoName string `json:"repo_name"`
}) (*mcp.CallToolResult, GetWorkspacePathOutput, error) {
	if input.Handle == "" {
		return nil, GetWorkspacePathOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if input.RepoName == "" {
		return nil, GetWorkspacePathOutput{}, NewToolError("repo_name is required. Use get_workspace to see repository names in a workspace")
	}

	ws, err := s.store.Get(ctx, input.Handle)
	if err != nil {
		return nil, GetWorkspacePathOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	repoPath := ws.Path + "/" + input.RepoName
	return nil, GetWorkspacePathOutput{Path: repoPath}, nil
}

func (s *Server) addRepository(ctx context.Context, req *mcp.CallToolRequest, input AddRepositoryInput) (*mcp.CallToolResult, AddRepositoryOutput, error) {
	if input.Handle == "" {
		return nil, AddRepositoryOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if input.Repo == "" {
		return nil, AddRepositoryOutput{}, NewToolError("repo is required. Use format: url or url@ref or url@ref::depth (e.g., github.com/org/repo@main::10)")
	}

	_, err := s.store.Get(ctx, input.Handle)
	if err != nil {
		return nil, AddRepositoryOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
	}

	url, ref, repoDepth := workspace.ParseRepoFlag(input.Repo)
	d := input.Depth
	if repoDepth > 0 {
		d = repoDepth
	}
	invocationCWD := ""
	err = s.store.AddRepository(ctx, input.Handle, workspace.RepositoryOption{URL: url, Ref: ref, Depth: d}, invocationCWD)
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

	ws, _ := s.store.Get(ctx, input.Handle)
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
	if input.Handle == "" {
		return nil, RemoveRepositoryOutput{}, NewToolError("handle is required. Use list_workspaces to see available workspaces")
	}
	if input.RepoName == "" {
		return nil, RemoveRepositoryOutput{}, NewToolError("repo_name is required. Use get_workspace to see repository names")
	}

	ws, err := s.store.Get(ctx, input.Handle)
	if err != nil {
		return nil, RemoveRepositoryOutput{}, s.workspaceNotFoundError(ctx, input.Handle)
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
		return nil, RemoveRepositoryOutput{}, NewToolError(fmt.Sprintf("repository %q not found in workspace %q. Available: %s", input.RepoName, input.Handle, strings.Join(names, ", ")))
	}

	err = s.store.RemoveRepository(ctx, input.Handle, input.RepoName)
	if err != nil {
		return nil, RemoveRepositoryOutput{}, NewToolError(fmt.Sprintf("failed to remove repository: %v", err))
	}

	return nil, RemoveRepositoryOutput{
		Success: true,
		Message: fmt.Sprintf("Repository %q removed from workspace %q", input.RepoName, input.Handle),
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
		Name:        "get_workspace",
		Description: "Get detailed information about a specific workspace, including its purpose, repositories, and creation date. Returns the workspace handle, purpose, list of repositories with their names, URLs, refs, and paths.",
	}, s.getWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_workspace",
		Description: "Create a new workspace with purpose and optional repositories. Use format 'url' or 'url@ref' (e.g., github.com/org/repo@main). Template variables use key=value pairs. Returns handle, path, and repository details.",
	}, s.createWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_workspace",
		Description: "Delete a Workshed workspace by its handle. This removes the workspace directory and all its repositories. Use with caution as this action cannot be undone.",
	}, s.removeWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "exec_command",
		Description: "Execute a shell command in a Workshed workspace. Takes the workspace handle and a command array (e.g., [\"make\", \"test\"]). Optionally specify a repository name with 'repo' or set 'all' to true to run in all repositories. Returns exit code and command output for each repository.",
	}, s.execCommand)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "capture_state",
		Description: "Create a git state snapshot (capture) for a workspace. Records the current branch, commit, and dirty status for each repository. Useful for documenting state before making changes. Takes a name and optional description and tags.",
	}, s.captureState)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_captures",
		Description: "List all captures for a workspace. Returns capture IDs, names, timestamps, descriptions, tags, and repository counts. Captures document git state at a point in time.",
	}, s.listCaptures)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "apply_capture",
		Description: "Apply (restore) git state from a capture. Takes a workspace handle and capture ID. By default, this will fail if repositories have uncommitted changes. Set dry_run to true to check preflight without applying.",
	}, s.applyCapture)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_workspace",
		Description: "Export a workspace to portable JSON format. Includes workspace metadata, repository configuration, and optionally captures. The JSON can be used with import_workspace to recreate the workspace elsewhere. Set compact to exclude captures.",
	}, s.exportWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "import_workspace",
		Description: "Create a workspace from exported context. Takes the context object from export_workspace output. Optionally preserve the original handle with preserve_handle. Repositories are cloned based on the exported configuration.",
	}, s.importWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace_path",
		Description: "Get the filesystem path for a workspace. Returns the absolute path where the workspace directory is located. Useful for opening the workspace in an IDE or running commands directly.",
	}, s.getWorkspacePath)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace_repo_path",
		Description: "Get the filesystem path for a specific repository within a workspace. Takes workspace handle and repository name. Returns the absolute path to that repository's directory.",
	}, s.getWorkspaceRepoPath)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_repository",
		Description: "Add a repository to an existing workspace. Takes workspace handle and repo URL with optional @ref (e.g., github.com/org/repo@main). Returns the added repository details.",
	}, s.addRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_repository",
		Description: "Remove a repository from a workspace by name. Takes workspace handle and repository name. Use get_workspace to see available repository names.",
	}, s.removeRepository)

	return server.Run(ctx, &mcp.StdioTransport{})
}
