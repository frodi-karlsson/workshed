//go:build integration
// +build integration

package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/workspace"
)

func TestCLICreateListRemoveWorkflowShouldExecuteCompleteWorkspaceLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	defer os.Unsetenv("WORKSHED_ROOT")

	// Reset CLI state
	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}
	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
	}()

	// Create a workspace
	outBuf.Reset()
	errBuf.Reset()
	exitCalled = false
	Create([]string{"--purpose", "Test workspace"})
	if exitCalled {
		t.Fatalf("Create exited: %s", errBuf.String())
	}

	output := outBuf.String()
	if !strings.Contains(output, "Created workspace:") {
		t.Errorf("Create output should contain 'Created workspace:', got: %s", output)
	}

	// Extract handle from output
	lines := strings.Split(output, "\n")
	var handle string
	for _, line := range lines {
		if strings.HasPrefix(line, "Created workspace:") {
			handle = strings.TrimSpace(strings.TrimPrefix(line, "Created workspace:"))
			break
		}
	}

	if handle == "" {
		t.Fatal("Could not extract handle from create output")
	}

	// List workspaces
	outBuf.Reset()
	errBuf.Reset()
	exitCalled = false
	List([]string{})
	if exitCalled {
		t.Fatalf("List exited: %s", errBuf.String())
	}

	output = outBuf.String()
	if !strings.Contains(output, handle) {
		t.Errorf("List output should contain handle %s, got: %s", handle, output)
	}
	if !strings.Contains(output, "Test workspace") {
		t.Errorf("List output should contain purpose, got: %s", output)
	}

	// Inspect workspace
	outBuf.Reset()
	errBuf.Reset()
	exitCalled = false
	Inspect([]string{handle})
	if exitCalled {
		t.Fatalf("Inspect exited: %s", errBuf.String())
	}

	output = outBuf.String()
	if !strings.Contains(output, handle) {
		t.Errorf("Inspect output should contain handle, got: %s", output)
	}
	if !strings.Contains(output, "Test workspace") {
		t.Errorf("Inspect output should contain purpose, got: %s", output)
	}

	// Remove workspace
	outBuf.Reset()
	errBuf.Reset()
	exitCalled = false
	Remove([]string{"--force", handle})
	if exitCalled {
		t.Fatalf("Remove exited: %s", errBuf.String())
	}

	// Verify workspace is gone
	store, err := workspace.NewFSStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	ctx := context.Background()
	_, err = store.Get(ctx, handle)
	if err == nil {
		t.Error("Workspace should have been removed")
	}
}

func TestCLICreateWithInvalidRepoURLShouldExitWithErrorForInvalidRepoURL(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	defer os.Unsetenv("WORKSHED_ROOT")

	// Reset CLI state
	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}
	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
	}()

	// Try to create with invalid repo URL
	Create([]string{"--purpose", "Test", "--repo", "https://github.com/nonexistent/repo12345@main"})

	if !exitCalled {
		t.Error("Create should have exited with error")
	}

	errOutput := errBuf.String()
	// Should contain git error with classification hint
	if !strings.Contains(errOutput, "Error creating workspace") {
		t.Errorf("Error output should mention workspace creation failed, got: %s", errOutput)
	}
}

func TestCLICreateContextCancellationShouldRespectContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	defer os.Unsetenv("WORKSHED_ROOT")

	store, err := workspace.NewFSStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure context is cancelled

	opts := workspace.CreateOptions{
		Purpose: "Test timeout",
		RepoURL: "https://github.com/torvalds/linux", // Large repo
		RepoRef: "master",
	}

	_, err = store.Create(ctx, opts)
	if err == nil {
		t.Error("Create should fail with cancelled context")
	}

	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "signal") {
		t.Logf("Expected context/signal error, got: %v", err)
	}

	// Verify no workspace directory was left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read tmpDir: %v", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Errorf("Temporary directory not cleaned up: %s", entry.Name())
		}
	}
}

func TestCLIListFilterIntegrationShouldFilterWorkspacesByPurpose(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	defer os.Unsetenv("WORKSHED_ROOT")

	store, err := workspace.NewFSStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	ctx := context.Background()

	// Create multiple workspaces
	_, err = store.Create(ctx, workspace.CreateOptions{Purpose: "Debug payment flow"})
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	_, err = store.Create(ctx, workspace.CreateOptions{Purpose: "Add login feature"})
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	_, err = store.Create(ctx, workspace.CreateOptions{Purpose: "Debug checkout bug"})
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitFunc = func(code int) {}
	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
	}()

	// List all workspaces
	outBuf.Reset()
	List([]string{})
	output := outBuf.String()
	if strings.Count(output, "Debug") != 2 {
		t.Errorf("Should show 2 debug workspaces, got: %s", output)
	}

	// Filter by purpose
	outBuf.Reset()
	List([]string{"--purpose", "debug"})
	output = outBuf.String()
	if !strings.Contains(output, "payment") {
		t.Errorf("Filtered list should contain 'payment', got: %s", output)
	}
	if !strings.Contains(output, "checkout") {
		t.Errorf("Filtered list should contain 'checkout', got: %s", output)
	}
	if strings.Contains(output, "login") {
		t.Errorf("Filtered list should not contain 'login', got: %s", output)
	}
}

func TestCLIRemoveNonExistentShouldExitWithErrorForNonexistentWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	defer os.Unsetenv("WORKSHED_ROOT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}
	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
	}()

	Remove([]string{"--force", "nonexistent-handle"})

	if !exitCalled {
		t.Error("Remove should exit with error for nonexistent workspace")
	}

	errOutput := errBuf.String()
	if !strings.Contains(errOutput, "not found") && !strings.Contains(errOutput, "Error") {
		t.Errorf("Error output should mention workspace not found, got: %s", errOutput)
	}
}
