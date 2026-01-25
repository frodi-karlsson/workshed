//go:build !integration
// +build !integration

package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

func newTestAlertModel(content string) *AlertModal {
	return &AlertModal{
		content:  content,
		quitting: false,
	}
}

func TestInspectModalRendersWorkspace(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "test-workspace",
		Purpose:   "Test purpose",
		Path:      "/test/path/to/workspace",
		CreatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
		},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestInspectModalShowsRepos(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "multi-repo-ws",
		Purpose:   "Multi repo workspace",
		Path:      "/test/workspaces/multi",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "frontend", URL: "https://github.com/org/frontend"},
			{Name: "backend", URL: "https://github.com/org/backend", Ref: "main"},
		},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestInspectModalClosesOnKey(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Testing",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressEnter(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestInspectModalClosesOnEsc(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "test-ws",
		Purpose:      "Testing",
		Path:         "/test/path",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressEsc(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestInspectModalShowsCreationDate(t *testing.T) {
	created := time.Date(2025, 1, 20, 14, 0, 0, 0, time.UTC)
	ws := &workspace.Workspace{
		Handle:       "dated-ws",
		Purpose:      "Dated testing",
		Path:         "/test/dated",
		CreatedAt:    created,
		Repositories: []workspace.Repository{},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestInspectModalWithNoRepos(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "empty-ws",
		Purpose:      "Empty workspace",
		Path:         "/test/empty",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestInspectModalWithRef(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "ref-ws",
		Purpose:   "Ref testing",
		Path:      "/test/ref",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "feature-repo", URL: "https://github.com/org/feature", Ref: "feature/new-feature"},
		},
	}

	content := buildWorkspaceDetailContent(ws)
	m := newTestAlertModel(content)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}
