package workspace

import (
	"time"
)

const ContextVersion = 1

type ExecutionRecord struct {
	ID          string                `json:"id"`
	Timestamp   time.Time             `json:"timestamp"`
	Handle      string                `json:"handle"`
	Target      string                `json:"target"`
	Command     []string              `json:"command"`
	Parallel    bool                  `json:"parallel"`
	ExitCode    int                   `json:"exit_code"`
	StartedAt   time.Time             `json:"started_at"`
	CompletedAt time.Time             `json:"completed_at"`
	Duration    int64                 `json:"duration_ms"`
	Results     []ExecutionRepoResult `json:"results"`
}

type ExecutionRepoResult struct {
	Repository string `json:"repository"`
	ExitCode   int    `json:"exit_code"`
	Duration   int64  `json:"duration_ms"`
	OutputPath string `json:"output_path,omitempty"`
	Error      string `json:"error,omitempty"`
}

type Capture struct {
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Handle    string          `json:"handle"`
	Name      string          `json:"name"`
	Kind      string          `json:"kind"`
	GitState  []GitRef        `json:"git_state"`
	Metadata  CaptureMetadata `json:"metadata"`
}

// CaptureKind describes the intent behind a capture.
// Captures are descriptive snapshots, not authoritative state.
const (
	CaptureKindManual     = "manual"     // User-initiated capture for documentation
	CaptureKindExecution  = "execution"  // Capture created from an execution record
	CaptureKindCheckpoint = "checkpoint" // Periodic state snapshot
)

type GitRef struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Dirty      bool   `json:"dirty"`
	Status     string `json:"status"`
}

type CaptureMetadata struct {
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

type WorkspaceContext struct {
	Version      int              `json:"version"`
	GeneratedAt  time.Time        `json:"generated_at"`
	Handle       string           `json:"handle"`
	Purpose      string           `json:"purpose"`
	Repositories []ContextRepo    `json:"repositories"`
	Captures     []ContextCapture `json:"captures,omitempty"`
	Metadata     ContextMetadata  `json:"metadata"`
}

type ContextCapture struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind,omitempty"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	RepoCount   int       `json:"repo_count"`
}

type ContextRepo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	RootPath string `json:"root_path"`
}

type ContextMetadata struct {
	WorkshedVersion string     `json:"workshed_version"`
	ExecutionsCount int        `json:"executions_count"`
	CapturesCount   int        `json:"captures_count"`
	LastExecutedAt  *time.Time `json:"last_executed_at,omitempty"`
	LastCapturedAt  *time.Time `json:"last_captured_at,omitempty"`
}

type AgentsValidationResult struct {
	Valid       bool            `json:"valid"`
	Errors      []AgentsError   `json:"errors,omitempty"`
	Warnings    []AgentsWarning `json:"warnings,omitempty"`
	Sections    []AgentsSection `json:"sections,omitempty"`
	Explanation string          `json:"explanation,omitempty"`
}

type AgentsError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type AgentsWarning struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type AgentsSection struct {
	Name     string `json:"name"`
	Line     int    `json:"line"`
	Valid    bool   `json:"valid"`
	Warnings int    `json:"warnings"`
	Errors   int    `json:"errors"`
}

type ListExecutionsOptions struct {
	Limit   int
	Offset  int
	Reverse bool
}

type CaptureOptions struct {
	Name        string
	Kind        string
	Description string
	Tags        []string
	Custom      map[string]string
}

type ApplyPreflightError struct {
	Repository string `json:"repository"`
	Reason     string `json:"reason"`
	Details    string `json:"details,omitempty"`
}

type ApplyPreflightResult struct {
	Valid  bool                  `json:"valid"`
	Errors []ApplyPreflightError `json:"errors,omitempty"`
}

const (
	ReasonDirtyWorkingTree  = "dirty_working_tree"
	ReasonMissingRepository = "missing_repository"
	ReasonCheckoutFailed    = "checkout_failed"
	ReasonHeadMismatch      = "head_mismatch"
	ReasonRepositoryNotGit  = "not_a_git_repository"
)
