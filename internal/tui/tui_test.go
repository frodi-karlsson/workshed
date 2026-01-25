//go:build !integration
// +build !integration

package tui

import (
	"testing"
	"time"

	"github.com/frodi/workshed/internal/testutil"
	"github.com/frodi/workshed/internal/workspace"
)

func TestIsHumanMode(t *testing.T) {
	tests := []struct {
		name      string
		logFormat string
		want      bool
	}{
		{
			name:      "should return true when WORKSHED_LOG_FORMAT is empty",
			logFormat: "",
			want:      true,
		},
		{
			name:      "should return true when WORKSHED_LOG_FORMAT is human",
			logFormat: "human",
			want:      true,
		},
		{
			name:      "should return false when WORKSHED_LOG_FORMAT is json",
			logFormat: "json",
			want:      false,
		},
		{
			name:      "should return false for any other value",
			logFormat: "machine",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.WithEnvVar(t, "WORKSHED_LOG_FORMAT", tt.logFormat, func() {
				got := IsHumanMode()
				if got != tt.want {
					t.Errorf("IsHumanMode() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestFilterPurposeItems(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "OpenCode development",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws2",
			Purpose:      "API migration",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws3",
			Purpose:      "OpenCode testing",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	tests := []struct {
		name         string
		query        string
		wantPurposes []string
	}{
		{
			name:         "should return all purposes when query is empty",
			query:        "",
			wantPurposes: []string{"API migration", "OpenCode development", "OpenCode testing"},
		},
		{
			name:         "should filter purposes case-insensitively",
			query:        "opencode",
			wantPurposes: []string{"OpenCode development", "OpenCode testing"},
		},
		{
			name:         "should filter by partial match",
			query:        "API",
			wantPurposes: []string{"API migration"},
		},
		{
			name:         "should return empty when no match",
			query:        "nonexistent",
			wantPurposes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := filterPurposeItems(workspaces, tt.query)
			testutil.AssertEqual(t, len(items), len(tt.wantPurposes), "item count mismatch")

			purposeSet := make(map[string]bool)
			for _, want := range tt.wantPurposes {
				purposeSet[want] = false
			}

			for _, item := range items {
				purposeItem := item.(purposeItem)
				if _, ok := purposeSet[purposeItem.purpose]; ok {
					purposeSet[purposeItem.purpose] = true
				}
			}

			for want, found := range purposeSet {
				if !found {
					t.Errorf("purpose %q should be in results", want)
				}
			}
		})
	}
}

func TestFilterPurposeItemsRemovesDuplicates(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Development",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws2",
			Purpose:      "Development",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
		{
			Handle:       "ws3",
			Purpose:      "Testing",
			CreatedAt:    time.Now(),
			Repositories: []workspace.Repository{},
		},
	}

	items := filterPurposeItems(workspaces, "")
	testutil.AssertEqual(t, len(items), 2, "should remove duplicate purposes")
}

func TestWorkspaceItem(t *testing.T) {
	now := time.Now()
	ws := &workspace.Workspace{
		Handle:    "test-ws",
		Purpose:   "Test purpose",
		CreatedAt: now,
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
			{Name: "repo2", URL: "https://github.com/org/repo2"},
		},
	}

	item := WorkspaceItem{workspace: ws}

	t.Run("Title should return handle", func(t *testing.T) {
		testutil.AssertEqual(t, item.Title(), "test-ws", "title should match handle")
	})

	t.Run("Description should include purpose and repo count", func(t *testing.T) {
		desc := item.Description()
		testutil.AssertContains(t, desc, "Test purpose", "description should contain purpose")
		testutil.AssertContains(t, desc, "2 repos", "description should contain repo count")
	})

	t.Run("FilterValue should combine handle and purpose", func(t *testing.T) {
		filterValue := item.FilterValue()
		testutil.AssertContains(t, filterValue, "test-ws", "filter should contain handle")
		testutil.AssertContains(t, filterValue, "Test purpose", "filter should contain purpose")
	})
}

func TestWorkspaceItemSingularRepo(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "single-repo",
		Purpose:   "Single repo test",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
		},
	}

	item := WorkspaceItem{workspace: ws}
	desc := item.Description()

	testutil.AssertContains(t, desc, "1 repo", "description should use singular 'repo' for count of 1")
}

func TestPurposeItem(t *testing.T) {
	item := purposeItem{purpose: "Test purpose"}

	t.Run("Title should return purpose", func(t *testing.T) {
		testutil.AssertEqual(t, item.Title(), "Test purpose", "title should match purpose")
	})

	t.Run("Description should be empty", func(t *testing.T) {
		testutil.AssertEqual(t, item.Description(), "", "description should be empty")
	})

	t.Run("FilterValue should return purpose", func(t *testing.T) {
		testutil.AssertEqual(t, item.FilterValue(), "Test purpose", "filter should match purpose")
	})
}
