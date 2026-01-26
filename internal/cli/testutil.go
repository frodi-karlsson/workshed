package cli

import (
	"bytes"
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
		Stderr:         env.ErrBuf,
		Stdout:         env.OutBuf,
		Stdin:          env.stdin,
		ExitFunc:       func(code int) { env.exitCalled = true },
		InvocationCWD:  env.TempDir,
		TableRenderer:  &MockTableRenderer{},
		OutputRenderer: &MockOutputRenderer{},
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

func (e *CLITestEnvironment) TableRenderer() *MockTableRenderer {
	return e.runner.TableRenderer.(*MockTableRenderer)
}

func (e *CLITestEnvironment) OutputRenderer() *MockOutputRenderer {
	return e.runner.OutputRenderer.(*MockOutputRenderer)
}

func (e *CLITestEnvironment) LastOutput() Output {
	calls := e.OutputRenderer().Calls
	if len(calls) == 0 {
		e.T.Fatal("No output renderer calls recorded")
	}
	return calls[len(calls)-1].Output
}

func (e *CLITestEnvironment) LastFormat() Format {
	calls := e.OutputRenderer().Calls
	if len(calls) == 0 {
		e.T.Fatal("No output renderer calls recorded")
	}
	return calls[len(calls)-1].Format
}

func (e *CLITestEnvironment) OutputCallCount() int {
	return len(e.OutputRenderer().Calls)
}

func (e *CLITestEnvironment) ResetOutputRenderer() {
	e.OutputRenderer().Reset()
}

func (e *CLITestEnvironment) AssertLastOutputRowContains(rowIdx int, colIdx int, substr string) {
	lastOutput := e.LastOutput()
	if rowIdx >= len(lastOutput.Rows) {
		e.T.Fatalf("Row index %d out of range (have %d rows)", rowIdx, len(lastOutput.Rows))
	}
	row := lastOutput.Rows[rowIdx]
	if colIdx >= len(row) {
		e.T.Fatalf("Column index %d out of range (row has %d columns)", colIdx, len(row))
	}
	if !strings.Contains(row[colIdx], substr) {
		e.T.Errorf("Row %d, column %d should contain %q, got: %q", rowIdx, colIdx, substr, row[colIdx])
	}
}

func (e *CLITestEnvironment) AssertLastOutputRowsEqual(expectedRows [][]string) {
	lastOutput := e.LastOutput()
	if len(lastOutput.Rows) != len(expectedRows) {
		e.T.Errorf("Expected %d rows, got %d", len(expectedRows), len(lastOutput.Rows))
		return
	}
	for i, expectedRow := range expectedRows {
		actualRow := lastOutput.Rows[i]
		if len(actualRow) != len(expectedRow) {
			e.T.Errorf("Row %d: expected %d columns, got %d", i, len(expectedRow), len(actualRow))
			continue
		}
		for j, expectedCell := range expectedRow {
			if actualRow[j] != expectedCell {
				e.T.Errorf("Row %d, column %d: expected %q, got %q", i, j, expectedCell, actualRow[j])
			}
		}
	}
}

func (e *CLITestEnvironment) AssertLastFormat(expected Format) {
	lastFormat := e.LastFormat()
	if lastFormat != expected {
		e.T.Errorf("Expected format %q, got %q", expected, lastFormat)
	}
}

func (e *CLITestEnvironment) ExtractHandleFromOutput(t *testing.T) string {
	mockRenderer := e.OutputRenderer()
	calls := mockRenderer.Calls
	if len(calls) == 0 {
		t.Fatal("No output renderer calls recorded")
	}

	lastCall := calls[len(calls)-1]
	for _, row := range lastCall.Output.Rows {
		if len(row) >= 2 && row[0] == "handle" {
			return row[1]
		}
	}
	t.Fatal("Handle not found in output rows")
	return ""
}

func (e *CLITestEnvironment) SetStdin(input string) {
	e.stdin = bytes.NewReader([]byte(input))
	if e.runner != nil {
		e.runner.Stdin = e.stdin
	}
}

func OutputContains(t *testing.T, output, substr string) bool {
	if !strings.Contains(output, substr) {
		t.Errorf("Output should contain %q, got: %s", substr, output)
		return false
	}
	return true
}
