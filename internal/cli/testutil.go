package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/logger"
)

type CLITestEnvironment struct {
	T          *testing.T
	TempDir    string
	OutBuf     *bytes.Buffer
	ErrBuf     *bytes.Buffer
	exitCalled bool
	originals  struct {
		outWriter io.Writer
		errWriter io.Writer
		exitFunc  func(int)
	}
}

func NewCLITestEnvironment(t *testing.T) *CLITestEnvironment {
	env := &CLITestEnvironment{
		T:       t,
		TempDir: t.TempDir(),
		OutBuf:  &bytes.Buffer{},
		ErrBuf:  &bytes.Buffer{},
	}

	if err := os.Setenv("WORKSHED_ROOT", env.TempDir); err != nil {
		t.Fatalf("Failed to set WORKSHED_ROOT: %v", err)
	}
	if err := os.Setenv("WORKSHED_LOG_FORMAT", "json"); err != nil {
		t.Fatalf("Failed to set WORKSHED_LOG_FORMAT: %v", err)
	}

	env.originals.outWriter = outWriter
	env.originals.errWriter = errWriter
	env.originals.exitFunc = exitFunc

	outWriter = env.OutBuf
	errWriter = env.ErrBuf
	exitFunc = func(code int) {
		env.exitCalled = true
	}
	logger.SetTestOutputWriter(env.OutBuf)

	return env
}

func (e *CLITestEnvironment) Cleanup() {
	if err := os.Unsetenv("WORKSHED_ROOT"); err != nil {
		e.T.Errorf("Failed to unset WORKSHED_ROOT: %v", err)
	}
	if err := os.Unsetenv("WORKSHED_LOG_FORMAT"); err != nil {
		e.T.Errorf("Failed to unset WORKSHED_LOG_FORMAT: %v", err)
	}

	outWriter = e.originals.outWriter
	errWriter = e.originals.errWriter
	exitFunc = e.originals.exitFunc
	logger.ClearTestOutputWriter()
}

func (e *CLITestEnvironment) ResetBuffers() {
	e.OutBuf.Reset()
	e.ErrBuf.Reset()
	e.exitCalled = false
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
