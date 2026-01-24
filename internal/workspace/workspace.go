package workspace

import (
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
