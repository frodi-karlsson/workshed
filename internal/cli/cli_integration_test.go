//go:build integration
// +build integration

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

func TestCLICreateListRemoveWorkflowShouldExecuteCompleteWorkspaceLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json") // Use JSON logging for predictable output
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	// Reset CLI state
	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	// Configure logger to use the same output buffer for testing
	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
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
	if !strings.Contains(output, "workspace created") {
		t.Errorf("Create output should contain 'workspace created', got: %s", output)
	}

	// Extract handle from JSON log output
	lines := strings.Split(output, "\n")
	var handle string
	for _, line := range lines {
		if strings.Contains(line, "workspace created") {
			// Parse JSON to extract handle
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logData); err == nil {
				if handleValue, ok := logData["handle"]; ok {
					handle = handleValue.(string)
					break
				}
			}
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
	os.Setenv("WORKSHED_LOG_FORMAT", "json") // Use JSON logging for predictable output
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	// Reset CLI state
	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	// Configure logger to use the same output buffer for testing
	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	// Try to create with invalid repo URL
	Create([]string{"--purpose", "Test", "--repo", "https://github.com/nonexistent/repo12345@main"})

	if !exitCalled {
		t.Error("Create should have exited with error")
	}

	errOutput := outBuf.String()
	if !strings.Contains(errOutput, "workspace creation failed") {
		t.Errorf("Error output should mention workspace creation failed, got: %s", errOutput)
	}
}

func TestCLICreateContextCancellationShouldRespectContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

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
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

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

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
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
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	Remove([]string{"--force", "nonexistent-handle"})

	if !exitCalled {
		t.Error("Remove should exit with error for nonexistent workspace")
	}

	output := outBuf.String()
	if !strings.Contains(output, "not found") && !strings.Contains(output, "Error") {
		t.Errorf("Output should mention workspace not found, got: %s", output)
	}
}

func TestCLICreateInReadOnlyDirectoryShouldExitWithError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Skipf("Cannot make directory read-only, skipping test: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		os.Chmod(tmpDir, 0755)
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	Create([]string{"--purpose", "Test workspace"})

	if !exitCalled {
		t.Error("Create should exit with error in read-only directory")
	}
}

func TestCLICreateWithSpecialCharactersInPurposeShouldWork(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	purpose := "Debug: payment flow with café and naïve users"
	Create([]string{"--purpose", purpose})

	if exitCalled {
		t.Fatalf("Create exited unexpectedly: %s", errBuf.String())
	}

	output := outBuf.String()
	if !strings.Contains(output, "workspace created") {
		t.Errorf("Create output should mention workspace created, got: %s", output)
	}

	var outBuf2, errBuf2 bytes.Buffer
	outWriter = &outBuf2
	errWriter = &errBuf2
	exitCalled = false
	exitFunc = func(code int) {
		exitCalled = true
	}

	List([]string{})

	if exitCalled {
		t.Fatalf("List exited unexpectedly: %s", errBuf2.String())
	}

	listOutput := outBuf2.String()
	if !strings.Contains(listOutput, purpose) {
		t.Errorf("List output should contain purpose with special chars, got: %s", listOutput)
	}
}

func TestCLIListEmptyDirectoryShouldHandleGracefully(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitFunc = func(code int) {}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	List([]string{})

	output := outBuf.String()
	if !strings.Contains(output, "no workspaces") {
		t.Errorf("List should mention no workspaces found, got: %s", output)
	}
}

func TestCLIInspectWithNonexistentWorkspaceShouldFailCleanly(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	Inspect([]string{"nonexistent-handle"})

	if !exitCalled {
		t.Error("Inspect should exit with error for nonexistent workspace")
	}

	output := outBuf.String()
	if !strings.Contains(output, "not found") && !strings.Contains(output, "failed to get") {
		t.Errorf("Inspect output should mention workspace not found or failed to get, got: %s", output)
	}
}

func TestCLIPathWithNonexistentWorkspaceShouldFailCleanly(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("WORKSHED_ROOT", tmpDir)
	os.Setenv("WORKSHED_LOG_FORMAT", "json")
	defer os.Unsetenv("WORKSHED_ROOT")
	defer os.Unsetenv("WORKSHED_LOG_FORMAT")

	var outBuf, errBuf bytes.Buffer
	outWriter = &outBuf
	errWriter = &errBuf
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	logger.SetTestOutputWriter(&outBuf)

	defer func() {
		outWriter = os.Stdout
		errWriter = os.Stderr
		exitFunc = os.Exit
		logger.ClearTestOutputWriter()
	}()

	Path([]string{"nonexistent-handle"})

	if !exitCalled {
		t.Error("Path should exit with error for nonexistent workspace")
	}

	output := outBuf.String()
	if !strings.Contains(output, "not found") && !strings.Contains(output, "failed to get") {
		t.Errorf("Path output should mention workspace not found or failed to get, got: %s", output)
	}
}
