package snapshot_test

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/tui/snapshot"
	"github.com/frodi/workshed/internal/workspace"
)

func TestValidateView_Success(t *testing.T) {
	result := workspace.AgentsValidationResult{
		Valid: true,
		Sections: []workspace.AgentsSection{
			{Name: "Task", Line: 3, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Goal", Line: 8, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Commands", Line: 15, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Notes", Line: 25, Valid: true, Errors: 0, Warnings: 0},
		},
		Errors:      []workspace.AgentsError{},
		Warnings:    []workspace.AgentsWarning{},
		Explanation: "AGENTS.md contains all required sections",
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithValidationResult(result),
	})
	scenario.Enter("Open context menu")
	scenario.Key("v", "Navigate to validate")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestValidateView_MissingSections(t *testing.T) {
	result := workspace.AgentsValidationResult{
		Valid: false,
		Sections: []workspace.AgentsSection{
			{Name: "Task", Line: 3, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Goal", Line: 0, Valid: false, Errors: 1, Warnings: 0},
			{Name: "Commands", Line: 12, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Notes", Line: 0, Valid: false, Errors: 1, Warnings: 0},
		},
		Errors: []workspace.AgentsError{
			{Line: 0, Message: "missing required section: Goal", Field: "section"},
			{Line: 0, Message: "missing required section: Notes", Field: "section"},
		},
		Warnings:    []workspace.AgentsWarning{},
		Explanation: "AGENTS.md is missing required sections",
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithValidationResult(result),
	})
	scenario.Enter("Open context menu")
	scenario.Key("v", "Navigate to validate")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestValidateView_WithWarnings(t *testing.T) {
	result := workspace.AgentsValidationResult{
		Valid: true,
		Sections: []workspace.AgentsSection{
			{Name: "Task", Line: 3, Valid: true, Errors: 0, Warnings: 1},
			{Name: "Goal", Line: 8, Valid: true, Errors: 0, Warnings: 0},
			{Name: "Commands", Line: 15, Valid: true, Errors: 0, Warnings: 2},
			{Name: "Notes", Line: 25, Valid: true, Errors: 0, Warnings: 0},
		},
		Errors: []workspace.AgentsError{
			{Line: 3, Message: "Task section empty", Field: "section"},
		},
		Warnings: []workspace.AgentsWarning{
			{Line: 15, Message: "Commands section could be more detailed", Field: "content"},
			{Line: 17, Message: "Missing example command", Field: "content"},
		},
		Explanation: "AGENTS.md has warnings but no errors",
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithValidationResult(result),
	})
	scenario.Enter("Open context menu")
	scenario.Key("v", "Navigate to validate")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestValidateView_EmptyWorkspace(t *testing.T) {
	result := workspace.AgentsValidationResult{
		Valid:    false,
		Sections: []workspace.AgentsSection{},
		Errors: []workspace.AgentsError{
			{Line: 0, Message: "AGENTS.md not found", Field: "file"},
		},
		Warnings:    []workspace.AgentsWarning{},
		Explanation: "Create AGENTS.md with required sections",
	}
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithWorkspaces([]*workspace.Workspace{
			{
				Handle:       "test-ws",
				Purpose:      "Test workspace",
				CreatedAt:    time.Now(),
				Repositories: []workspace.Repository{},
			},
		}),
		snapshot.WithValidationResult(result),
	})
	scenario.Enter("Open context menu")
	scenario.Key("v", "Navigate to validate")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}
