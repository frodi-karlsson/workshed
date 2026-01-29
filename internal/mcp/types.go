package mcp

type WorkspaceInfo struct {
	Handle    string `json:"handle"`
	Purpose   string `json:"purpose"`
	RepoCount int    `json:"repo_count"`
	CreatedAt string `json:"created_at"`
}

type RepositoryInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
	Path string `json:"path"`
}

type WorkspaceDetail struct {
	Handle       string           `json:"handle"`
	Purpose      string           `json:"purpose"`
	Repositories []RepositoryInfo `json:"repositories"`
	CreatedAt    string           `json:"created_at"`
}

type CreateWorkspaceInput struct {
	Purpose      string   `json:"purpose"`
	Repos        []string `json:"repos,omitempty"`
	Template     string   `json:"template,omitempty"`
	TemplateVars []string `json:"template_vars,omitempty"`
	Depth        int      `json:"depth,omitempty"`
}

type CreateWorkspaceOutput struct {
	Handle       string           `json:"handle"`
	Purpose      string           `json:"purpose"`
	Path         string           `json:"path"`
	Repositories []RepositoryInfo `json:"repositories"`
}

type RemoveWorkspaceInput struct {
	Handle *string `json:"handle,omitempty"`
	DryRun bool    `json:"dry_run,omitempty"`
	Yes    bool    `json:"yes,omitempty"`
}

type RemoveWorkspaceOutput struct {
	Success       bool   `json:"success"`
	WouldDelete   bool   `json:"would_delete,omitempty"`
	Message       string `json:"message,omitempty"`
	ReposCount    int    `json:"repos_count,omitempty"`
	CapturesCount int    `json:"captures_count,omitempty"`
}

type ExecCommandInput struct {
	Handle      *string  `json:"handle,omitempty"`
	Command     []string `json:"command"`
	Repo        string   `json:"repo,omitempty"`
	All         bool     `json:"all,omitempty"`
	NoRecord    bool     `json:"no_record,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
	OutputLimit int      `json:"output_limit,omitempty"`
}

type ExecResultInfo struct {
	Repository string `json:"repository"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

type ExecCommandOutput struct {
	Success   bool             `json:"success"`
	ExitCode  int              `json:"exit_code,omitempty"`
	Results   []ExecResultInfo `json:"results,omitempty"`
	TotalTime int64            `json:"total_time_ms,omitempty"`
}

type CaptureStateInput struct {
	Handle      *string  `json:"handle,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type GitRefInfo struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Dirty      bool   `json:"dirty"`
}

type CaptureStateOutput struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Timestamp string       `json:"timestamp"`
	GitState  []GitRefInfo `json:"git_state"`
}

type ListCapturesInput struct {
	Handle *string `json:"handle,omitempty"`
}

type CaptureInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Timestamp   string   `json:"timestamp"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RepoCount   int      `json:"repo_count"`
}

type ListCapturesOutput struct {
	Captures []CaptureInfo `json:"captures"`
}

type ListWorkspacesOutput struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

type ApplyCaptureInput struct {
	Handle    *string `json:"handle,omitempty"`
	CaptureID string  `json:"capture_id"`
	DryRun    bool    `json:"dry_run,omitempty"`
}

type ApplyCaptureOutput struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type ExportWorkspaceInput struct {
	Handle  *string `json:"handle,omitempty"`
	Compact bool    `json:"compact,omitempty"`
}

type ExportWorkspaceOutput struct {
	JSON     string         `json:"json"`
	Metadata ExportMetadata `json:"metadata"`
}

type ExportMetadata struct {
	Handle       string `json:"handle"`
	Purpose      string `json:"purpose"`
	RepoCount    int    `json:"repo_count"`
	CaptureCount int    `json:"capture_count,omitempty"`
}

type ImportWorkspaceInput struct {
	Context        map[string]any `json:"context"`
	PreserveHandle bool           `json:"preserve_handle,omitempty"`
}

type AddRepositoryInput struct {
	Handle *string `json:"handle,omitempty"`
	Repo   string  `json:"repo"`
	Depth  int     `json:"depth,omitempty"`
}

type AddRepositoryOutput struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
	Path string `json:"path"`
}

type RemoveRepositoryInput struct {
	Handle   *string `json:"handle,omitempty"`
	RepoName string  `json:"repo_name"`
}

type RemoveRepositoryOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ImportWorkspaceOutput struct {
	Handle    string `json:"handle"`
	Purpose   string `json:"purpose"`
	Path      string `json:"path"`
	RepoCount int    `json:"repo_count"`
}

type GetWorkspacePathInput struct {
	Handle *string `json:"handle,omitempty"`
}

type GetWorkspacePathOutput struct {
	Path string `json:"path"`
}

type GetWorkspaceRepoPathInput struct {
	Handle   *string `json:"handle,omitempty"`
	RepoName string  `json:"repo_name"`
}

type GetWorkspaceInput struct {
	Handle *string `json:"handle,omitempty"`
}

type EnterWorkspaceInput struct {
	Handle *string `json:"handle,omitempty"`
}

type EnterWorkspaceOutput struct {
	Handle string `json:"handle"`
	Path   string `json:"path"`
}

type ExitWorkspaceOutput struct {
	Message string `json:"message"`
}

type HelpOutput struct {
	Message string `json:"message"`
}

type ToolError struct {
	Message string `json:"message"`
}

func (e *ToolError) Error() string {
	return e.Message
}

func NewToolError(msg string) error {
	return &ToolError{Message: msg}
}
