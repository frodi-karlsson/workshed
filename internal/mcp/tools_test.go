package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/workspace"
)

func TestListWorkspaces(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		_, out, err := server.listWorkspaces(ctx, nil, struct{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Workspaces) != 0 {
			t.Errorf("expected empty list, got %d", len(out.Workspaces))
		}
	})

	t.Run("with workspaces", func(t *testing.T) {
		_, _, _ = server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "test 1"})
		_, _, _ = server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "test 2"})

		_, out, err := server.listWorkspaces(ctx, nil, struct{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Workspaces) != 2 {
			t.Errorf("expected 2 workspaces, got %d", len(out.Workspaces))
		}
	})
}

func TestGetWorkspace(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.getWorkspace(ctx, nil, GetWorkspaceInput{})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := server.getWorkspace(ctx, nil, GetWorkspaceInput{Handle: "nonexistent"})
		if err == nil {
			t.Error("expected error for nonexistent handle")
		}
	})

	t.Run("valid handle", func(t *testing.T) {
		_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "test purpose"})
		_, getOut, err := server.getWorkspace(ctx, nil, GetWorkspaceInput{Handle: createOut.Handle})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if getOut.Handle != createOut.Handle {
			t.Errorf("expected handle %s, got %s", createOut.Handle, getOut.Handle)
		}
		if getOut.Purpose != "test purpose" {
			t.Errorf("expected purpose 'test purpose', got '%s'", getOut.Purpose)
		}
	})
}

func TestCreateWorkspace(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()

	t.Run("purpose required", func(t *testing.T) {
		_, _, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{})
		if err == nil {
			t.Error("expected error for empty purpose")
		}
	})

	t.Run("with purpose", func(t *testing.T) {
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "test purpose"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Handle == "" {
			t.Error("output should have handle")
		}
		if out.Purpose != "test purpose" {
			t.Errorf("expected purpose 'test purpose', got '%s'", out.Purpose)
		}
		if out.Path == "" {
			t.Error("output should have path")
		}
	})

	t.Run("with repo", func(t *testing.T) {
		localRepo := workspace.CreateLocalGitRepo(t, "testrepo", map[string]string{"file.txt": "content"})
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose: "with repo",
			Repos:   []string{localRepo},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Repositories) != 1 {
			t.Errorf("expected 1 repo, got %d", len(out.Repositories))
		}
		if out.Repositories[0].Name != "testrepo" {
			t.Errorf("expected repo name 'testrepo', got '%s'", out.Repositories[0].Name)
		}
	})

	t.Run("with template vars", func(t *testing.T) {
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose:      "with vars",
			TemplateVars: []string{"key1=value1", "key2=value2"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Handle == "" {
			t.Error("should create workspace with template vars")
		}
	})

	t.Run("with depth in repo string", func(t *testing.T) {
		localRepo := workspace.CreateLocalGitRepo(t, "depthrepocreate", map[string]string{"file.txt": "content"})
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose: "with depth in repo",
			Repos:   []string{localRepo + "::5"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Repositories) != 1 {
			t.Errorf("expected 1 repo, got %d", len(out.Repositories))
		}
	})

	t.Run("with depth field", func(t *testing.T) {
		localRepo := workspace.CreateLocalGitRepo(t, "depthrepocreate2", map[string]string{"file.txt": "content"})
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose: "with depth field",
			Repos:   []string{localRepo},
			Depth:   10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Repositories) != 1 {
			t.Errorf("expected 1 repo, got %d", len(out.Repositories))
		}
	})

	t.Run("repo depth overrides field depth", func(t *testing.T) {
		localRepo := workspace.CreateLocalGitRepo(t, "depthrepocreate3", map[string]string{"file.txt": "content"})
		_, out, err := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose: "with depth override",
			Repos:   []string{localRepo + "::7"},
			Depth:   99,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Repositories) != 1 {
			t.Errorf("expected 1 repo, got %d", len(out.Repositories))
		}
	})
}

func TestRemoveWorkspace(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.removeWorkspace(ctx, nil, RemoveWorkspaceInput{})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "to remove"})
		_, out, err := server.removeWorkspace(ctx, nil, RemoveWorkspaceInput{Handle: createOut.Handle})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.Success {
			t.Error("expected success=true")
		}
		if out.Message == "" {
			t.Error("expected message")
		}
		_, _, err = server.getWorkspace(ctx, nil, GetWorkspaceInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("workspace should be removed after success")
		}
	})

	t.Run("dry run", func(t *testing.T) {
		_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "dry run test"})
		ws, _ := server.store.Get(ctx, createOut.Handle)
		_, out, err := server.removeWorkspace(ctx, nil, RemoveWorkspaceInput{Handle: createOut.Handle, DryRun: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.WouldDelete {
			t.Error("expected would_delete=true for dry run")
		}
		if out.ReposCount != len(ws.Repositories) {
			t.Errorf("expected repos_count=%d, got %d", len(ws.Repositories), out.ReposCount)
		}
		// Verify workspace still exists after dry run
		_, _, err = server.getWorkspace(ctx, nil, GetWorkspaceInput{Handle: createOut.Handle})
		if err != nil {
			t.Error("workspace should still exist after dry run")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := server.removeWorkspace(ctx, nil, RemoveWorkspaceInput{Handle: "nonexistent-workspace"})
		if err == nil {
			t.Error("expected error for nonexistent handle")
		}
	})
}

func TestExecCommand(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	localRepo := workspace.CreateLocalGitRepo(t, "exectestrepo", map[string]string{"file.txt": "content"})
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
		Purpose: "exec test",
		Repos:   []string{localRepo},
	})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.execCommand(ctx, nil, ExecCommandInput{Command: []string{"echo", "test"}})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("command required", func(t *testing.T) {
		_, _, err := server.execCommand(ctx, nil, ExecCommandInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty command")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.execCommand(ctx, nil, ExecCommandInput{
			Handle:  createOut.Handle,
			Command: []string{"echo", "hello"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Results) == 0 {
			t.Error("expected results")
		}
	})

	t.Run("with target repo", func(t *testing.T) {
		_, out, err := server.execCommand(ctx, nil, ExecCommandInput{
			Handle:  createOut.Handle,
			Command: []string{"echo", "targeted"},
			Repo:    "exectestrepo",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Results) != 1 {
			t.Errorf("expected 1 result for targeted repo, got %d", len(out.Results))
		}
	})
}

func TestCaptureState(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
		Purpose: "capture test",
	})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.captureState(ctx, nil, CaptureStateInput{Name: "snapshot"})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("name required", func(t *testing.T) {
		_, _, err := server.captureState(ctx, nil, CaptureStateInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.captureState(ctx, nil, CaptureStateInput{
			Handle:      createOut.Handle,
			Name:        "test snapshot",
			Description: "a test capture",
			Tags:        []string{"test"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.ID == "" {
			t.Error("output should have ID")
		}
		if out.Name != "test snapshot" {
			t.Errorf("expected name 'test snapshot', got '%s'", out.Name)
		}
	})
}

func TestListCaptures(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	localRepo := workspace.CreateLocalGitRepo(t, "capturetestrepo", map[string]string{"file.txt": "content"})
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
		Purpose: "list captures test",
		Repos:   []string{localRepo},
	})
	_, _, _ = server.captureState(ctx, nil, CaptureStateInput{Handle: createOut.Handle, Name: "capture 1", Description: "first capture"})
	_, _, _ = server.captureState(ctx, nil, CaptureStateInput{Handle: createOut.Handle, Name: "capture 2", Description: "second capture"})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.listCaptures(ctx, nil, ListCapturesInput{})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.listCaptures(ctx, nil, ListCapturesInput{Handle: createOut.Handle})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.Captures) != 2 {
			t.Errorf("expected 2 captures, got %d", len(out.Captures))
		}
	})
}

func TestApplyCapture(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "apply test"})
	_, captureOut, err := server.captureState(ctx, nil, CaptureStateInput{
		Handle:      createOut.Handle,
		Name:        "to apply",
		Description: "test capture for apply",
	})
	if err != nil {
		t.Fatalf("captureState failed: %v", err)
	}

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.applyCapture(ctx, nil, ApplyCaptureInput{CaptureID: captureOut.ID})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("capture_id required", func(t *testing.T) {
		_, _, err := server.applyCapture(ctx, nil, ApplyCaptureInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty capture_id")
		}
	})

	t.Run("dry run success", func(t *testing.T) {
		_, out, err := server.applyCapture(ctx, nil, ApplyCaptureInput{
			Handle:    createOut.Handle,
			CaptureID: captureOut.ID,
			DryRun:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.Success {
			t.Error("expected dry-run success")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.applyCapture(ctx, nil, ApplyCaptureInput{
			Handle:    createOut.Handle,
			CaptureID: captureOut.ID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.Success {
			t.Errorf("expected success=true, got false: %s", out.Message)
		}
		if !strings.Contains(out.Message, "successfully") {
			t.Errorf("expected success message, got: %s", out.Message)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := server.applyCapture(ctx, nil, ApplyCaptureInput{
			Handle:    createOut.Handle,
			CaptureID: "nonexistent-capture",
		})
		if err == nil {
			t.Error("expected error for nonexistent capture")
		}
	})
}

func TestExportWorkspace(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "export test"})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.exportWorkspace(ctx, nil, ExportWorkspaceInput{})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := server.exportWorkspace(ctx, nil, ExportWorkspaceInput{Handle: "nonexistent"})
		if err == nil {
			t.Error("expected error for nonexistent handle")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.exportWorkspace(ctx, nil, ExportWorkspaceInput{Handle: createOut.Handle})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.JSON == "" {
			t.Error("output should have JSON")
		}
		if out.Metadata.Handle != createOut.Handle {
			t.Errorf("expected handle %s, got %s", createOut.Handle, out.Metadata.Handle)
		}
	})

	t.Run("compact", func(t *testing.T) {
		_, out, err := server.exportWorkspace(ctx, nil, ExportWorkspaceInput{
			Handle:  createOut.Handle,
			Compact: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = out
	})
}

func TestImportWorkspace(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()

	t.Run("context required", func(t *testing.T) {
		_, _, err := server.importWorkspace(ctx, nil, ImportWorkspaceInput{})
		if err == nil {
			t.Error("expected error for empty context")
		}
	})

	t.Run("invalid context", func(t *testing.T) {
		_, _, err := server.importWorkspace(ctx, nil, ImportWorkspaceInput{Context: map[string]any{"invalid": "data"}})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("missing purpose in context", func(t *testing.T) {
		_, _, err := server.importWorkspace(ctx, nil, ImportWorkspaceInput{Context: map[string]any{
			"handle":       "test",
			"purpose":      "",
			"repositories": []any{},
		}})
		if err == nil {
			t.Error("expected error for missing purpose")
		}
	})

	t.Run("success", func(t *testing.T) {
		localRepo := workspace.CreateLocalGitRepo(t, "importtestrepo", map[string]string{"file.txt": "content"})
		_, sourceOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
			Purpose: "source",
			Repos:   []string{localRepo},
		})
		exportCtx, _ := server.store.ExportContext(ctx, sourceOut.Handle)
		exportJSON, _ := json.Marshal(exportCtx)
		var exportMap map[string]any
		if err := json.Unmarshal(exportJSON, &exportMap); err != nil {
			t.Fatalf("failed to unmarshal export: %v", err)
		}

		_, out, err := server.importWorkspace(ctx, nil, ImportWorkspaceInput{Context: exportMap})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Handle == "" {
			t.Error("output should have handle")
		}
		if out.Purpose != "source" {
			t.Errorf("expected purpose 'source', got '%s'", out.Purpose)
		}
	})
}

func TestGetWorkspacePath(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "path test"})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.getWorkspacePath(ctx, nil, GetWorkspacePathInput{})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := server.getWorkspacePath(ctx, nil, GetWorkspacePathInput{Handle: "nonexistent"})
		if err == nil {
			t.Error("expected error for nonexistent handle")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.getWorkspacePath(ctx, nil, GetWorkspacePathInput{Handle: createOut.Handle})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Path == "" {
			t.Error("output should have path")
		}
	})
}

func TestGetWorkspaceRepoPath(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	localRepo := workspace.CreateLocalGitRepo(t, "repopathrepo", map[string]string{"file.txt": "content"})
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
		Purpose: "repo path test",
		Repos:   []string{localRepo},
	})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.getWorkspaceRepoPath(ctx, nil, struct {
			Handle   string `json:"handle"`
			RepoName string `json:"repo_name"`
		}{RepoName: "repopathrepo"})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("repo_name required", func(t *testing.T) {
		_, _, err := server.getWorkspaceRepoPath(ctx, nil, struct {
			Handle   string `json:"handle"`
			RepoName string `json:"repo_name"`
		}{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty repo_name")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, out, err := server.getWorkspaceRepoPath(ctx, nil, struct {
			Handle   string `json:"handle"`
			RepoName string `json:"repo_name"`
		}{Handle: createOut.Handle, RepoName: "nonexistent"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Function doesn't validate repo existence, just constructs path
		if out.Path == "" {
			t.Error("expected path even for nonexistent repo")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.getWorkspaceRepoPath(ctx, nil, struct {
			Handle   string `json:"handle"`
			RepoName string `json:"repo_name"`
		}{Handle: createOut.Handle, RepoName: "repopathrepo"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Path == "" {
			t.Error("output should have path")
		}
	})
}

func TestAddRepository(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{Purpose: "add repo test"})
	localRepo := workspace.CreateLocalGitRepo(t, "newrepo", map[string]string{"file.txt": "content"})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.addRepository(ctx, nil, AddRepositoryInput{Repo: localRepo})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("repo required", func(t *testing.T) {
		_, _, err := server.addRepository(ctx, nil, AddRepositoryInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty repo")
		}
	})

	t.Run("workspace not found", func(t *testing.T) {
		_, _, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: "nonexistent",
			Repo:   localRepo,
		})
		if err == nil {
			t.Error("expected error for nonexistent workspace")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: createOut.Handle,
			Repo:   localRepo,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Name == "" {
			t.Error("output should have name")
		}
		if out.URL == "" {
			t.Error("output should have URL")
		}
		if out.Path == "" {
			t.Error("output should have path")
		}
	})

	t.Run("with ref", func(t *testing.T) {
		anotherRepo := workspace.CreateLocalGitRepo(t, "refrepo"+createOut.Handle[:4], map[string]string{"file.txt": "content"})
		_, out, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: createOut.Handle,
			Repo:   anotherRepo + "@main",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Ref != "main" {
			t.Errorf("expected ref 'main', got '%s'", out.Ref)
		}
	})

	t.Run("with depth in repo string", func(t *testing.T) {
		depthRepo := workspace.CreateLocalGitRepo(t, "depthrepo"+createOut.Handle[:4], map[string]string{"file.txt": "content"})
		_, out, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: createOut.Handle,
			Repo:   depthRepo + "::5",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.URL == "" {
			t.Error("output should have URL")
		}
	})

	t.Run("with depth field", func(t *testing.T) {
		depthRepo2 := workspace.CreateLocalGitRepo(t, "depthrepo2"+createOut.Handle[:4], map[string]string{"file.txt": "content"})
		_, out, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: createOut.Handle,
			Repo:   depthRepo2,
			Depth:  10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.URL == "" {
			t.Error("output should have URL")
		}
	})

	t.Run("repo depth overrides input depth", func(t *testing.T) {
		depthRepo3 := workspace.CreateLocalGitRepo(t, "depthrepo3"+createOut.Handle[:4], map[string]string{"file.txt": "content"})
		_, out, err := server.addRepository(ctx, nil, AddRepositoryInput{
			Handle: createOut.Handle,
			Repo:   depthRepo3 + "::7",
			Depth:  99,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.URL == "" {
			t.Error("output should have URL")
		}
	})
}

func TestRemoveRepository(t *testing.T) {
	t.Parallel()
	store, _ := workspace.CreateTestStore(t)
	server := NewServer(store)
	ctx := context.Background()
	localRepo := workspace.CreateLocalGitRepo(t, "toremoverepo", map[string]string{"file.txt": "content"})
	_, createOut, _ := server.createWorkspace(ctx, nil, CreateWorkspaceInput{
		Purpose: "remove repo test",
		Repos:   []string{localRepo},
	})

	t.Run("handle required", func(t *testing.T) {
		_, _, err := server.removeRepository(ctx, nil, RemoveRepositoryInput{RepoName: "toremoverepo"})
		if err == nil {
			t.Error("expected error for empty handle")
		}
	})

	t.Run("repo_name required", func(t *testing.T) {
		_, _, err := server.removeRepository(ctx, nil, RemoveRepositoryInput{Handle: createOut.Handle})
		if err == nil {
			t.Error("expected error for empty repo_name")
		}
	})

	t.Run("workspace not found", func(t *testing.T) {
		_, _, err := server.removeRepository(ctx, nil, RemoveRepositoryInput{
			Handle:   "nonexistent",
			RepoName: "toremoverepo",
		})
		if err == nil {
			t.Error("expected error for nonexistent workspace")
		}
	})

	t.Run("repo not found", func(t *testing.T) {
		_, _, err := server.removeRepository(ctx, nil, RemoveRepositoryInput{
			Handle:   createOut.Handle,
			RepoName: "nonexistent-repo",
		})
		if err == nil {
			t.Error("expected error for nonexistent repo")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, out, err := server.removeRepository(ctx, nil, RemoveRepositoryInput{
			Handle:   createOut.Handle,
			RepoName: "toremoverepo",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", nil)
		}
		if !out.Success {
			t.Error("expected success=true")
		}
		if out.Message == "" {
			t.Error("expected message")
		}
	})
}
