package workspace

import (
	"context"
	"time"
)

const CurrentMetadataVersion = 1

// Repository represents a git repository within a workspace.
type Repository struct {
	// URL is the clone URL of the repository.
	URL string `json:"url"`

	// Ref is the git reference (branch or tag) to use.
	// Empty means the default branch.
	Ref string `json:"ref,omitempty"`

	// Name is a human-readable identifier for the repository.
	Name string `json:"name"`
}

// RepositoryOption specifies a repository to add during workspace creation.
type RepositoryOption struct {
	// URL is the clone URL of the repository.
	URL string

	// Ref is the optional git reference to check out.
	Ref string
}

// Workspace represents a collection of repositories managed together.
type Workspace struct {
	// Version is the metadata version for future compatibility.
	Version int `json:"version"`

	// Handle is a unique identifier for the workspace.
	Handle string `json:"handle"`

	// Purpose describes what the workspace is used for.
	Purpose string `json:"purpose"`

	// Repositories contains the repositories in this workspace.
	Repositories []Repository `json:"repositories"`

	// CreatedAt records when the workspace was created.
	CreatedAt time.Time `json:"created_at"`

	// Path is the filesystem location of the workspace.
	// This field is not persisted to JSON.
	Path string `json:"-"`
}

func (ws *Workspace) GetRepositoryByName(name string) *Repository {
	for i := range ws.Repositories {
		if ws.Repositories[i].Name == name {
			return &ws.Repositories[i]
		}
	}
	return nil
}

// CreateOptions specifies the configuration for a new workspace.
type CreateOptions struct {
	// Purpose describes the intended use of the workspace.
	Purpose string

	// Template is an optional directory whose contents will be copied into the workspace.
	Template string

	// TemplateVars provides variable substitutions for template file/directory names.
	// Keys are matched against {{key}} patterns and replaced with their values.
	TemplateVars map[string]string

	// Repositories specifies the repositories to include in the workspace.
	Repositories []RepositoryOption

	InvocationCWD string
}

// ListOptions specifies filtering criteria for listing workspaces.
type ListOptions struct {
	// PurposeFilter returns only workspaces whose purpose contains this string.
	PurposeFilter string
}

// InvocationContext defines an interface for accessing the original invocation current working directory.
type InvocationContext interface {
	GetInvocationCWD() string
}

// Store defines the interface for persisting and retrieving workspaces.
// All operations accept a context for cancellation and timeout control.
type Store interface {
	// Create initializes a new workspace with the given options.
	Create(ctx context.Context, opts CreateOptions) (*Workspace, error)

	// Get retrieves a workspace by its unique handle.
	Get(ctx context.Context, handle string) (*Workspace, error)

	// List returns all workspaces, optionally filtered by the provided options.
	List(ctx context.Context, opts ListOptions) ([]*Workspace, error)

	// Remove deletes a workspace identified by its handle.
	Remove(ctx context.Context, handle string) error

	// Path returns the filesystem path where a workspace is stored.
	Path(ctx context.Context, handle string) (string, error)

	// UpdatePurpose modifies the purpose string for a given workspace.
	UpdatePurpose(ctx context.Context, handle string, purpose string) error

	// FindWorkspace locates a workspace based on a directory path.
	// Returns nil if no workspace is found for the given directory.
	FindWorkspace(ctx context.Context, dir string) (*Workspace, error)

	// Exec runs a command in all repositories belonging to a workspace.
	Exec(ctx context.Context, handle string, opts ExecOptions) ([]ExecResult, error)

	// AddRepository adds a repository to an existing workspace.
	AddRepository(ctx context.Context, handle string, repo RepositoryOption, invocationCWD string) error

	// AddRepositories adds multiple repositories to an existing workspace.
	AddRepositories(ctx context.Context, handle string, repos []RepositoryOption, invocationCWD string) error

	// RemoveRepository removes a repository from an existing workspace.
	RemoveRepository(ctx context.Context, handle string, repoName string) error
}
