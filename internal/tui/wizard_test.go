//go:build !integration
// +build !integration

package tui

import (
	"context"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/workspace"
)

func TestWizard_CreateWorkspace(t *testing.T) {
	t.Run("should create workspace with just purpose when skipping repos", func(t *testing.T) {
		t.Skip("Skipping interactive test that enables mouse tracking and doesn't clean up terminal state properly")
		ctx := context.Background()

		done := make(chan *WizardResult, 1)
		errChan := make(chan error, 1)

		go func() {
			result, err := RunCreateWizard(ctx, &workspace.FSStore{})
			if err != nil {
				errChan <- err
				return
			}
			done <- result
		}()

		time.Sleep(100 * time.Millisecond)

		select {
		case result := <-done:
			if result.Purpose != "" {
				t.Logf("Got result: %+v", result)
			}
		case err := <-errChan:
			t.Logf("Got error (expected for this test): %v", err)
		case <-time.After(500 * time.Millisecond):
			t.Log("Wizard timed out (expected for interactive test)")
		}
	})
}

func TestExtractRecentRepos(t *testing.T) {
	t.Run("should extract unique repos from workspaces", func(t *testing.T) {
		workspaces := []*workspace.Workspace{
			{
				Handle:  "ws1",
				Purpose: "Test 1",
				Repositories: []workspace.Repository{
					{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
					{URL: "https://github.com/org/repo2", Ref: "develop", Name: "repo2"},
				},
			},
			{
				Handle:  "ws2",
				Purpose: "Test 2",
				Repositories: []workspace.Repository{
					{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
					{URL: "https://github.com/org/repo3", Ref: "", Name: "repo3"},
				},
			},
		}

		repos := extractRecentRepos(workspaces)

		expectedCount := 3
		if len(repos) != expectedCount {
			t.Errorf("Expected %d unique repos, got %d", expectedCount, len(repos))
		}

		repoURLs := make(map[string]bool)
		for _, repo := range repos {
			key := repo.url
			if repo.ref != "" {
				key = repo.url + "@" + repo.ref
			}
			if repoURLs[key] {
				t.Errorf("Duplicate repo found: %s", key)
			}
			repoURLs[key] = true
		}
	})

	t.Run("should handle workspaces with no repos", func(t *testing.T) {
		workspaces := []*workspace.Workspace{
			{
				Handle:       "ws1",
				Purpose:      "Test 1",
				Repositories: []workspace.Repository{},
			},
		}

		repos := extractRecentRepos(workspaces)

		if len(repos) != 0 {
			t.Errorf("Expected 0 repos, got %d", len(repos))
		}
	})

	t.Run("should handle empty workspace list", func(t *testing.T) {
		workspaces := []*workspace.Workspace{}

		repos := extractRecentRepos(workspaces)

		if len(repos) != 0 {
			t.Errorf("Expected 0 repos, got %d", len(repos))
		}
	})
}
