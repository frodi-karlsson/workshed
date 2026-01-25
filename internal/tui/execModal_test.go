//go:build !integration
// +build !integration

package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/frodi/workshed/internal/workspace"
)

func TestExecModalInitial(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "test-ws",
		Purpose:   "Test",
		Path:      "/test/path",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestExecModalShowsRepos(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "multi-ws",
		Purpose:   "Multi repo",
		Path:      "/test/multi",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "frontend", URL: "https://github.com/org/frontend"},
			{Name: "backend", URL: "https://github.com/org/backend"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestExecModalCancelEsc(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "test-ws",
		Purpose:   "Test",
		Path:      "/test/path",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressEsc(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestExecModalCancelCtrlC(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "test-ws",
		Purpose:   "Test",
		Path:      "/test/path",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "repo1", URL: "https://github.com/org/repo1"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressCtrlC(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestExecModalWithRef(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "ref-ws",
		Purpose:   "Ref test",
		Path:      "/test/ref",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "feature", URL: "https://github.com/org/feature", Ref: "main"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestExecModalShowsAllReposSelected(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:    "select-ws",
		Purpose:   "Selection test",
		Path:      "/test/select",
		CreatedAt: time.Now(),
		Repositories: []workspace.Repository{
			{Name: "repo-a", URL: "https://github.com/org/a"},
			{Name: "repo-b", URL: "https://github.com/org/b"},
		},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestExecModalEmptyRepos(t *testing.T) {
	ws := &workspace.Workspace{
		Handle:       "empty-ws",
		Purpose:      "Empty",
		Path:         "/test/empty",
		CreatedAt:    time.Now(),
		Repositories: []workspace.Repository{},
	}

	m := newTestExecModel(ws)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}
