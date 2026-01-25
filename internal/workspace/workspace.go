package workspace

import (
	"context"
	"time"
)

const CurrentMetadataVersion = 1

type Repository struct {
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
	Name string `json:"name"`
}

type RepositoryOption struct {
	URL string
	Ref string
}

type Workspace struct {
	Version      int          `json:"version"`
	Handle       string       `json:"handle"`
	Purpose      string       `json:"purpose"`
	Repositories []Repository `json:"repositories"`
	CreatedAt    time.Time    `json:"created_at"`
	Path         string       `json:"-"`
}

func (ws *Workspace) GetRepositoryByName(name string) *Repository {
	for i := range ws.Repositories {
		if ws.Repositories[i].Name == name {
			return &ws.Repositories[i]
		}
	}
	return nil
}

type CreateOptions struct {
	Purpose      string
	Repositories []RepositoryOption
}

type ListOptions struct {
	PurposeFilter string
}

type Store interface {
	Create(ctx context.Context, opts CreateOptions) (*Workspace, error)
	Get(ctx context.Context, handle string) (*Workspace, error)
	List(ctx context.Context, opts ListOptions) ([]*Workspace, error)
	Remove(ctx context.Context, handle string) error
	Path(ctx context.Context, handle string) (string, error)
	UpdatePurpose(ctx context.Context, handle string, purpose string) error
	FindWorkspace(ctx context.Context, dir string) (*Workspace, error)
	Exec(ctx context.Context, handle string, opts ExecOptions) ([]ExecResult, error)
}
