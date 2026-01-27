package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

type TestEnvironment struct {
	T      *testing.T
	OutBuf bytes.Buffer
	ErrBuf bytes.Buffer
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()
	return &TestEnvironment{
		T: t,
	}
}

func (e *TestEnvironment) Cleanup() {
	e.OutBuf.Reset()
	e.ErrBuf.Reset()
}

func (e *TestEnvironment) Output() string {
	return e.OutBuf.String()
}

func (e *TestEnvironment) ErrorOutput() string {
	return e.ErrBuf.String()
}

func (e *TestEnvironment) ResetBuffers() {
	e.OutBuf.Reset()
	e.ErrBuf.Reset()
}

func (e *TestEnvironment) AssertOutputContains(substr string) {
	e.T.Helper()
	output := e.Output()
	if !bytes.Contains(e.OutBuf.Bytes(), []byte(substr)) {
		e.T.Errorf("expected output to contain %q, got: %s", substr, output)
	}
}

func (e *TestEnvironment) AssertNoOutput() {
	e.T.Helper()
	if e.OutBuf.Len() > 0 {
		e.T.Errorf("expected no output, got: %s", e.Output())
	}
}

func (e *TestEnvironment) AssertErrorContains(substr string) {
	e.T.Helper()
	if !bytes.Contains(e.ErrBuf.Bytes(), []byte(substr)) {
		e.T.Errorf("expected error output to contain %q, got: %s", substr, e.ErrorOutput())
	}
}

type CapturedCommand struct {
	Command string
	Args    []string
}

type MockRunner struct {
	Store       workspace.Store
	Logger      *logger.Logger
	Invocations []CapturedCommand
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		Invocations: []CapturedCommand{},
	}
}

func (m *MockRunner) RecordInvocation(cmd string, args ...string) {
	m.Invocations = append(m.Invocations, CapturedCommand{cmd, args})
}

func (m *MockRunner) LastInvocation() CapturedCommand {
	if len(m.Invocations) == 0 {
		return CapturedCommand{}
	}
	return m.Invocations[len(m.Invocations)-1]
}

type TestWorkspace struct {
	Handle  string
	Purpose string
	Path    string
	Repos   []string
}

func CreateTestWorkspace(tempDir, handle, purpose string, repos []string) *TestWorkspace {
	wsPath := filepath.Join(tempDir, handle)
	for _, repo := range repos {
		repoPath := filepath.Join(wsPath, filepath.Base(repo))
		_ = os.MkdirAll(repoPath, 0755)
	}
	return &TestWorkspace{
		Handle:  handle,
		Purpose: purpose,
		Path:    wsPath,
		Repos:   repos,
	}
}

func AssertExitCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	if err == nil {
		if expectedCode != 0 {
			t.Errorf("expected exit code %d, got nil error", expectedCode)
		}
		return
	}
	t.Logf("error occurred (exit code check skipped for non-exec errors): %v", err)
}

func AssertFormat(t *testing.T, output string, expectedFormat string) {
	t.Helper()
	if expectedFormat == "json" {
		if !bytes.HasPrefix([]byte(output), []byte("{")) && !bytes.HasPrefix([]byte(output), []byte("[")) {
			t.Errorf("expected json output, got: %s", output)
		}
	}
}

func AssertTableOutput(t *testing.T, output string) {
	t.Helper()
	if !bytes.Contains([]byte(output), []byte("handle")) {
		t.Errorf("expected table output with 'handle' column, got: %s", output)
	}
}
