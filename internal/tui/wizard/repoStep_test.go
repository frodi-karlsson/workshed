//go:build !integration
// +build !integration

package wizard

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/frodi/workshed/internal/workspace"
)

func TestRepoStep_Initialization(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	if step.purpose != "test purpose" {
		t.Errorf("Expected purpose 'test purpose', got '%s'", step.purpose)
	}

	if step.done {
		t.Error("Expected done to be false initially")
	}

	if step.cancelled {
		t.Error("Expected cancelled to be false initially")
	}

	if step.adding {
		t.Error("Expected adding to be false initially")
	}

	if step.focusedField != 0 {
		t.Error("Expected focusedField to be 0 initially")
	}

	if len(step.repositories) != 0 {
		t.Error("Expected repositories to be empty initially")
	}
}

func TestRepoStep_EnterSetsDone(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !updatedStep.done {
		t.Error("Expected done to be true after Enter")
	}
}

func TestRepoStep_AddMode(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	if step.adding {
		t.Error("Expected adding to be false initially")
	}

	updatedStep, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	if !updatedStep.adding {
		t.Error("Expected adding to be true after 'a'")
	}

	if updatedStep.mode != modeTyping {
		t.Error("Expected mode to be modeTyping after 'a'")
	}

	if updatedStep.focusedField != 0 {
		t.Error("Expected focusedField to be 0 after 'a'")
	}

	if cmd == nil {
		t.Error("Expected non-nil command (textinput.Blink) after 'a'")
	}
}

func TestRepoStep_RemoveLastRepo(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.repositories = []workspace.RepositoryOption{
		{URL: "https://github.com/org/repo1", Ref: "main"},
		{URL: "https://github.com/org/repo2", Ref: "develop"},
	}

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if len(updatedStep.repositories) != 1 {
		t.Errorf("Expected 1 repository after 'd', got %d", len(updatedStep.repositories))
	}

	if updatedStep.repositories[0].URL != "https://github.com/org/repo1" {
		t.Error("Expected first repository to remain")
	}
}

func TestRepoStep_RemoveLastRepoEmptyList(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if len(updatedStep.repositories) != 0 {
		t.Errorf("Expected 0 repositories when list is empty, got %d", len(updatedStep.repositories))
	}
}

func TestRepoStep_AddRepository(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 0
	step.mode = modeTyping
	step.urlInput.Focus()
	step.urlInput.SetValue("https://github.com/org/repo1")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if updatedStep.adding {
		t.Error("Expected adding to be false after adding repo")
	}

	if len(updatedStep.repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(updatedStep.repositories))
	}

	if updatedStep.repositories[0].URL != "https://github.com/org/repo1" {
		t.Errorf("Expected repository URL 'https://github.com/org/repo1', got '%s'", updatedStep.repositories[0].URL)
	}
}

func TestRepoStep_AddRepositoryWithRef(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 1
	step.mode = modeTyping
	step.urlInput.SetValue("https://github.com/org/repo1")
	step.urlInput.Blur()
	step.refInput.Focus()
	step.refInput.SetValue("develop")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if len(updatedStep.repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(updatedStep.repositories))
	}

	if updatedStep.repositories[0].Ref != "develop" {
		t.Errorf("Expected ref 'develop', got '%s'", updatedStep.repositories[0].Ref)
	}
}

func TestRepoStep_CancelAdding(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 0
	step.urlInput.Focus()

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updatedStep.adding {
		t.Error("Expected adding to be false after ESC")
	}

	if updatedStep.urlInput.Focused() {
		t.Error("Expected urlInput to be blurred after ESC")
	}
}

func TestRepoStep_TabNavigation(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 0
	step.mode = modeTyping

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if updatedStep.focusedField != 1 {
		t.Error("Expected focusedField to be 1 after Tab")
	}
}

func TestRepoStep_TabFromRefToSelecting(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:    "ws1",
			Purpose:   "test purpose",
			Path:      "/test/ws1",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 1
	step.mode = modeTyping

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyTab})

	if updatedStep.mode != modeSelecting {
		t.Error("Expected mode to switch to modeSelecting after Tab from ref")
	}
}

func TestRepoStep_SelectFromRecentList(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:    "ws1",
			Purpose:   "test purpose",
			Path:      "/test/ws1",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 1
	step.mode = modeSelecting
	step.recentList.Select(0)

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if updatedStep.adding {
		t.Error("Expected adding to be false after selecting from recent list")
	}

	if len(updatedStep.repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(updatedStep.repositories))
	}

	if updatedStep.repositories[0].URL != "https://github.com/org/repo1" {
		t.Errorf("Expected repository URL 'https://github.com/org/repo1', got '%s'", updatedStep.repositories[0].URL)
	}
}

func TestRepoStep_EmptyURLDoesNotAdd(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.urlInput.SetValue("")

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if len(updatedStep.repositories) != 0 {
		t.Errorf("Expected 0 repositories, got %d", len(updatedStep.repositories))
	}
}

func TestRepoStep_NavigationInAddingMode(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:    "ws1",
			Purpose:   "test purpose",
			Path:      "/test/ws1",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.focusedField = 0
	step.mode = modeTyping

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyDown})

	if updatedStep.mode != modeSelecting {
		t.Error("Expected mode to switch to modeSelecting on KeyDown")
	}
}

func TestRepoStep_QCancelsAdding(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	step := NewRepoStep(workspaces, "test purpose")

	step.adding = true
	step.urlInput.Focus()

	updatedStep, _ := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if updatedStep.adding {
		t.Error("Expected adding to be false after 'q'")
	}
}

func TestRepoStep_IsDone(t *testing.T) {
	step := NewRepoStep(nil, "test")

	if step.IsDone() {
		t.Error("Expected IsDone() to return false initially")
	}

	step.done = true

	if !step.IsDone() {
		t.Error("Expected IsDone() to return true after done is set")
	}
}

func TestRepoStep_IsCancelled(t *testing.T) {
	step := NewRepoStep(nil, "test")

	if step.IsCancelled() {
		t.Error("Expected IsCancelled() to return false initially")
	}

	step.cancelled = true

	if !step.IsCancelled() {
		t.Error("Expected IsCancelled() to return true after cancelled is set")
	}
}

func TestRepoStep_GetResult(t *testing.T) {
	step := NewRepoStep(nil, "test")

	result := step.GetResult()

	repos, ok := result.([]workspace.RepositoryOption)
	if !ok {
		t.Error("Expected GetResult() to return []workspace.RepositoryOption")
	}

	if len(repos) != 0 {
		t.Errorf("Expected 0 repositories initially, got %d", len(repos))
	}

	step.repositories = []workspace.RepositoryOption{
		{URL: "https://github.com/org/repo1", Ref: "main"},
	}

	result = step.GetResult()

	repos, ok = result.([]workspace.RepositoryOption)
	if !ok {
		t.Error("Expected GetResult() to return []workspace.RepositoryOption")
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}
}

func TestRepoStep_GetRepositories(t *testing.T) {
	step := NewRepoStep(nil, "test")

	if len(step.GetRepositories()) != 0 {
		t.Error("Expected empty repositories initially")
	}

	step.repositories = []workspace.RepositoryOption{
		{URL: "https://github.com/org/repo1", Ref: "main"},
	}

	repos := step.GetRepositories()

	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}
}

func TestRepoStep_ViewNotAdding(t *testing.T) {
	step := NewRepoStep(nil, "test")

	output := step.View()

	if output == "" {
		t.Error("Expected non-empty view output")
	}

	if !contains(output, "Add Repositories") {
		t.Error("View output should contain 'Add Repositories' title")
	}

	if !contains(output, "Repositories:") {
		t.Error("View output should contain 'Repositories:' label")
	}

	if !contains(output, "a] Add repository") {
		t.Error("View output should contain add help")
	}
}

func TestRepoStep_ViewAdding(t *testing.T) {
	step := NewRepoStep(nil, "test")

	step.adding = true
	step.focusedField = 0
	step.mode = modeTyping

	output := step.View()

	if !contains(output, "Add Repository") {
		t.Error("View output should contain 'Add Repository' when adding")
	}

	if !contains(output, "URL:") {
		t.Error("View output should contain 'URL:' label when adding")
	}

	if !contains(output, "Ref:") {
		t.Error("View output should contain 'Ref:' label when adding")
	}
}

func TestRepoStep_ViewWithRepositories(t *testing.T) {
	step := NewRepoStep(nil, "test")

	step.repositories = []workspace.RepositoryOption{
		{URL: "https://github.com/org/repo1", Ref: "main"},
		{URL: "https://github.com/org/repo2", Ref: ""},
	}

	output := step.View()

	if !contains(output, "repo1") {
		t.Error("View output should contain first repository")
	}

	if !contains(output, "repo2") {
		t.Error("View output should contain second repository")
	}
}

func TestRepoStep_ViewWithRecentRepos(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:    "ws1",
			Purpose:   "test",
			Path:      "/test",
			CreatedAt: time.Now(),
			Repositories: []workspace.Repository{
				{URL: "https://github.com/org/repo1", Ref: "main", Name: "repo1"},
			},
		},
	}

	step := NewRepoStep(workspaces, "test")

	step.adding = true
	step.focusedField = 1
	step.mode = modeSelecting

	output := step.View()

	if !contains(output, "Recent repos") {
		t.Error("View output should contain 'Recent repos' when in selecting mode")
	}
}

func TestRepoStep_Init(t *testing.T) {
	step := NewRepoStep(nil, "test")

	cmd := step.Init()

	if cmd == nil {
		t.Error("Expected Init to return non-nil command (textinput.Blink)")
	}
}

func TestExtractRecentReposFromWorkspaces(t *testing.T) {
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

	seen := make(map[string]bool)
	for _, repo := range repos {
		key := repo.url
		if repo.ref != "" {
			key = repo.url + "@" + repo.ref
		}
		if seen[key] {
			t.Errorf("Duplicate repo found: %s", key)
		}
		seen[key] = true
	}
}

func TestExtractRecentReposEmptyList(t *testing.T) {
	workspaces := []*workspace.Workspace{}

	repos := extractRecentRepos(workspaces)

	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}
}

func TestExtractRecentReposNoRepos(t *testing.T) {
	workspaces := []*workspace.Workspace{
		{
			Handle:       "ws1",
			Purpose:      "Test",
			Repositories: []workspace.Repository{},
		},
	}

	repos := extractRecentRepos(workspaces)

	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}
}

var _ = textinput.Model{}
