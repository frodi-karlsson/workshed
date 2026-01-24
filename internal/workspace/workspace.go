package workspace

import (
	"context"
	"time"
)

// Workspace represents a workshed workspace
type Workspace struct {
	Version   int       `json:"version"`
	Handle    string    `json:"handle"`
	Purpose   string    `json:"purpose"`
	RepoURL   string    `json:"repo_url,omitempty"`
	RepoRef   string    `json:"repo_ref,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Path      string    `json:"-"` // computed, not stored
}

const (
	CurrentMetadataVersion = 1
)

// CreateOptions contains options for creating a workspace
type CreateOptions struct {
	Purpose string
	RepoURL string
	RepoRef string
}

// ListOptions contains options for listing workspaces
type ListOptions struct {
	PurposeFilter string
}

// Store defines the interface for workspace storage operations
type Store interface {
	Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
	Get(ctx context.Context, handle string) (*Workspace, error)
	List(ctx context.Context, opts ListOptions) ([]*Workspace, error)
	Remove(ctx context.Context, handle string) error
	Path(ctx context.Context, handle string) (string, error)
}
