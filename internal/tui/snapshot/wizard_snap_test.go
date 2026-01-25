package snapshot_test

import (
	"errors"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/tui/snapshot"
)

func TestWizardView_EmptyRepoDetection(t *testing.T) {
	scenario := snapshot.NewScenario(t, []snapshot.GitOption{
		snapshot.WithGitRemoteURL("git@github.com/user/workshed"),
	}, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Enter("Empty repo input")
	scenario.Enter("Trigger create workspace")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_CreateSuccessEmptyInput(t *testing.T) {
	scenario := snapshot.NewScenario(t,
		[]snapshot.GitOption{snapshot.WithGitRemoteURL("git@github.com/user/workshed")},
		[]snapshot.StoreOption{snapshot.WithCreateDelay(100 * time.Millisecond)},
	)

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Enter("Empty repo input")
	scenario.Enter("Trigger create workspace")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_CustomRepoInput(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Type("git@github.com/org/repo", "Enter repo URL")
	scenario.Enter("Add repo")
	scenario.Enter("Trigger create workspace")
	scenario.Enter("Dismiss")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_CancelDuringGitDetection(t *testing.T) {
	mockErr := &git.GitError{
		Operation: "get-url",
		Hint:      "failed",
		Details:   "fatal: not a git repository",
	}

	scenario := snapshot.NewScenario(t, []snapshot.GitOption{
		snapshot.WithGitRemoteError(mockErr),
	}, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Enter("Trigger git detection")
	scenario.Key("c", "Cancel during detection")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_CreateError(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithCreateError(errors.New("failed to create workspace")),
	})

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Type("git@github.com/org/test-repo", "Enter repo URL")
	scenario.Enter("Add repo")
	scenario.Enter("Trigger create workspace (will fail)")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_CreateSuccess(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, []snapshot.StoreOption{
		snapshot.WithCreateDelay(100 * time.Millisecond),
	})

	scenario.Key("c", "Open create wizard")
	scenario.Type("My project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Type("git@github.com/org/test-repo", "Enter repo URL")
	scenario.Enter("Add repo")
	scenario.Enter("Trigger create workspace")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_MultipleRepos(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("API migration", "Enter purpose")
	scenario.Enter("Confirm purpose")

	scenario.Key("a", "Add first repo")
	scenario.Type("https://github.com/org/backend", "Enter backend URL")
	scenario.Enter("Add backend")

	scenario.Key("a", "Add second repo")
	scenario.Type("https://github.com/org/frontend", "Enter frontend URL")
	scenario.Enter("Add frontend")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_WithBranchRef(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Feature development", "Enter purpose")
	scenario.Enter("Confirm purpose")

	scenario.Key("a", "Add repo")
	scenario.Type("https://github.com/org/my-repo@feature/new-ui", "Enter repo with branch")
	scenario.Enter("Add repo with branch")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_GitDetectionFailure(t *testing.T) {
	mockErr := &git.GitError{
		Operation: "get-url",
		Hint:      "failed",
		Details:   "fatal: not a git repository (or any of the parent directories): .git",
	}

	scenario := snapshot.NewScenario(t, []snapshot.GitOption{
		snapshot.WithGitRemoteError(mockErr),
	}, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Test project", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Enter("Trigger git detection (empty input)")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_BackFromRepoStep(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Go back test", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Type("https://github.com/org/repo", "Enter repo URL")

	scenario.Key("esc", "Go back to purpose step")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_EscCancelsWizard(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Will cancel", "Enter purpose")
	scenario.Enter("Confirm purpose")
	scenario.Key("esc", "Cancel wizard")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_EmptyPurposeDoesNotAdvance(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Enter("Try to advance with empty purpose")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_LocalPathRepo(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Local repo project", "Enter purpose")
	scenario.Enter("Confirm purpose")

	scenario.Key("a", "Add local repo")
	scenario.Type("./my-local-repo", "Enter local path")
	scenario.Enter("Add local path")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}

func TestWizardView_TildePathRepo(t *testing.T) {
	scenario := snapshot.NewScenario(t, nil, nil)

	scenario.Key("c", "Open create wizard")
	scenario.Type("Home repo project", "Enter purpose")
	scenario.Enter("Confirm purpose")

	scenario.Key("a", "Add tilde path repo")
	scenario.Type("~/projects/my-repo", "Enter tilde path")
	scenario.Enter("Add tilde path")
	output := scenario.Record()
	snapshot.Match(t, t.Name(), output)
}
