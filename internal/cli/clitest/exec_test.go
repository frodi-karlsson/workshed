package clitest

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/cli/exec"
)

func TestExecCommand(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test", nil)

	t.Run("command with handle and simple command", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{ws.Handle, "pwd"})
		if err != nil {
			t.Errorf("exec with valid handle should succeed, got: %v", err)
		}
		if strings.Contains(env.ErrorOutput(), "unknown command") {
			t.Errorf("exec should not produce unknown command error, stderr: %s", env.ErrorOutput())
		}
	})

	t.Run("command with -a flag from workspace directory", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{ws.Handle, "-a", "pwd"})
		if err != nil {
			t.Errorf("exec with -a should work: %v", err)
		}
	})

	t.Run("missing command error", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{ws.Handle})
		if err == nil {
			t.Error("exec with no command should fail")
		}
		if !strings.Contains(env.ErrorOutput(), "missing") && !strings.Contains(env.ErrorOutput(), "command") {
			t.Errorf("exec missing command should mention command, stderr: %s", env.ErrorOutput())
		}
	})

	t.Run("format json", func(t *testing.T) {
		if err := env.Run(exec.Command(), []string{ws.Handle, "pwd", "--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var results []exec.ExecResultOutput
		err := json.Unmarshal([]byte(output), &results)
		if err != nil {
			t.Errorf("Expected valid JSON output: %v, got: %s", err, output)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("format raw", func(t *testing.T) {
		if err := env.Run(exec.Command(), []string{ws.Handle, "pwd", "--format", "raw"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var results []exec.ExecResultOutput
		err := json.Unmarshal([]byte(output), &results)
		if err != nil {
			t.Errorf("Expected valid JSON output: %v, got: %s", err, output)
		}
	})

	t.Run("with --repo flag", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{ws.Handle, "--repo", "testrepo", "pwd"})
		if err != nil {
			t.Errorf("exec with --repo should work: %v", err)
		}
	})

	t.Run("with --no-record flag", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{ws.Handle, "pwd", "--no-record"})
		if err != nil {
			t.Errorf("exec with --no-record should work: %v", err)
		}
	})
}

func TestExecCommandNoWorkspace(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	t.Run("command without handle from non-workspace directory", func(t *testing.T) {
		err := env.Run(exec.Command(), []string{"pwd"})
		if err == nil {
			t.Error("exec without workspace should fail")
		}
		if !strings.Contains(env.ErrorOutput(), "workspace") && !strings.Contains(env.ErrorOutput(), "missing command") {
			t.Errorf("exec should mention workspace or command, stderr: %s", env.ErrorOutput())
		}
	})
}

func TestExecCommandFromWorkspaceDir(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test", nil)

	t.Run("command without handle from workspace directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd failed: %v", err)
		}
		defer func() {
			if err := os.Chdir(cwd); err != nil {
				t.Errorf("Chdir back failed: %v", err)
			}
		}()
		if err := os.Chdir(ws.Path); err != nil {
			t.Fatalf("Chdir failed: %v", err)
		}

		err = env.Run(exec.Command(), []string{"pwd"})
		if err != nil {
			t.Errorf("exec pwd from workspace dir should work: %v", err)
		}
		if strings.Contains(env.ErrorOutput(), "missing command") {
			t.Errorf("exec should not produce missing command error, stderr: %s", env.ErrorOutput())
		}
	})
}
