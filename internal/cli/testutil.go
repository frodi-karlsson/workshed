package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

type CLITestEnvironment struct {
	T          *testing.T
	TempDir    string
	OutBuf     *bytes.Buffer
	ErrBuf     *bytes.Buffer
	exitCalled bool
	runner     *Runner
	stdin      *bytes.Reader
}

func NewCLITestEnvironment(t *testing.T) *CLITestEnvironment {
	env := &CLITestEnvironment{
		T:       t,
		TempDir: t.TempDir(),
		OutBuf:  &bytes.Buffer{},
		ErrBuf:  &bytes.Buffer{},
		stdin:   bytes.NewReader([]byte{}),
	}

	if err := os.Setenv("WORKSHED_ROOT", env.TempDir); err != nil {
		t.Fatalf("Failed to set WORKSHED_ROOT: %v", err)
	}
	if err := os.Setenv("WORKSHED_LOG_FORMAT", "json"); err != nil {
		t.Fatalf("Failed to set WORKSHED_LOG_FORMAT: %v", err)
	}

	env.runner = &Runner{
		Stderr:        env.ErrBuf,
		Stdout:        env.OutBuf,
		Stdin:         env.stdin,
		ExitFunc:      func(code int) { env.exitCalled = true },
		InvocationCWD: env.TempDir,
	}

	store, err := workspace.NewFSStore(env.TempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	env.runner.Store = store

	env.runner.Logger = logger.NewLogger(logger.INFO, "workshed", logger.WithTestOutput(env.OutBuf))

	return env
}

func (e *CLITestEnvironment) Cleanup() {
	if err := os.Unsetenv("WORKSHED_ROOT"); err != nil {
		e.T.Errorf("Failed to unset WORKSHED_ROOT: %v", err)
	}
	if err := os.Unsetenv("WORKSHED_LOG_FORMAT"); err != nil {
		e.T.Errorf("Failed to unset WORKSHED_LOG_FORMAT: %v", err)
	}
	e.runner.Logger = nil
}

func (e *CLITestEnvironment) ResetBuffers() {
	e.OutBuf.Reset()
	e.ErrBuf.Reset()
	e.exitCalled = false
	if e.runner != nil {
		e.runner.Stdout = e.OutBuf
		e.runner.Stderr = e.ErrBuf
	}
}

func (e *CLITestEnvironment) ExitCalled() bool {
	return e.exitCalled
}

func (e *CLITestEnvironment) Output() string {
	return e.OutBuf.String()
}

func (e *CLITestEnvironment) ErrorOutput() string {
	return e.ErrBuf.String()
}

func (e *CLITestEnvironment) Runner() *Runner {
	return e.runner
}

func (e *CLITestEnvironment) SetStdin(input string) {
	e.stdin = bytes.NewReader([]byte(input))
	if e.runner != nil {
		e.runner.Stdin = e.stdin
	}
}

func ExtractHandleFromLog(t *testing.T, output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "workspace created") {
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logData); err != nil {
				t.Fatalf("Failed to parse JSON log: %v", err)
			}
			handle, ok := logData["handle"].(string)
			if !ok {
				t.Fatalf("Handle not found in log data: %v", logData)
			}
			return handle
		}
	}
	t.Fatal("No workspace created message found in output")
	return ""
}

func ExtractFieldFromLog(t *testing.T, output, messageType, fieldName string) interface{} {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, messageType) {
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logData); err != nil {
				t.Fatalf("Failed to parse JSON log: %v", err)
			}
			field, ok := logData[fieldName]
			if !ok {
				t.Fatalf("Field %s not found in log data: %v", fieldName, logData)
			}
			return field
		}
	}
	t.Fatalf("No %s message found in output", messageType)
	return nil
}

func OutputContains(t *testing.T, output, substr string) bool {
	if !strings.Contains(output, substr) {
		t.Errorf("Output should contain %q, got: %s", substr, output)
		return false
	}
	return true
}
