package clitest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/cli/capture"
	"github.com/frodi/workshed/internal/cli/captures"
	"github.com/frodi/workshed/internal/cli/export"
	"github.com/frodi/workshed/internal/cli/list"
	"github.com/frodi/workshed/internal/cli/path"
)

func TestCapturesOutputFormats(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)
	if err := env.Run(capture.Command(), []string{"--name", "test capture", ws.Handle}); err != nil {
		t.Errorf("Run failed: %v", err)
	}

	t.Run("table format has columns", func(t *testing.T) {
		if err := env.Run(captures.Command(), []string{ws.Handle}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		if !strings.Contains(output, "ID") && !strings.Contains(output, "NAME") {
			t.Errorf("Expected table format with columns, got: %s", output)
		}
	})

	t.Run("json format is valid", func(t *testing.T) {
		if err := env.Run(captures.Command(), []string{ws.Handle, "--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var data interface{}
		err := json.Unmarshal([]byte(output), &data)
		if err != nil {
			t.Errorf("Expected valid JSON output, got: %s", output)
		}
	})

	t.Run("raw format is just IDs", func(t *testing.T) {
		if err := env.Run(captures.Command(), []string{ws.Handle, "--format", "raw"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		if strings.Contains(output, "ID") || strings.Contains(output, "NAME") {
			t.Errorf("Raw format should not contain headers, got: %s", output)
		}
		if len(output) < 26 {
			t.Errorf("Raw format should contain capture IDs (26 chars), got: %s", output)
		}
	})
}

func TestExportOutputFormats(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("export json format", func(t *testing.T) {
		if err := env.Run(export.Command(), []string{ws.Handle, "--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var data map[string]interface{}
		err := json.Unmarshal([]byte(output), &data)
		if err != nil {
			t.Errorf("Expected valid JSON: %v", err)
		}
		if _, ok := data["handle"]; !ok {
			t.Errorf("Expected 'handle' in export output")
		}
		if _, ok := data["purpose"]; !ok {
			t.Errorf("Expected 'purpose' in export output")
		}
		if _, ok := data["repositories"]; !ok {
			t.Errorf("Expected 'repositories' in export output")
		}
	})

	t.Run("export compact excludes captures", func(t *testing.T) {
		if err := env.Run(export.Command(), []string{ws.Handle, "--compact", "--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(output), &data); err != nil {
			t.Errorf("Unmarshal failed: %v, output: %s", err, output)
		}
		if captures, ok := data["captures"]; ok {
			if captures != nil {
				t.Errorf("compact export should exclude captures, got: %v", captures)
			}
		}
	})
}

func TestListOutputFormats(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	env.CreateWorkspace("test purpose 1", nil)
	env.CreateWorkspace("test purpose 2", nil)

	t.Run("list json format", func(t *testing.T) {
		if err := env.Run(list.Command(), []string{"--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var data []interface{}
		err := json.Unmarshal([]byte(output), &data)
		if err != nil {
			t.Errorf("Expected valid JSON array: %v", err)
		}
		if len(data) < 2 {
			t.Errorf("Expected at least 2 workspaces, got %d", len(data))
		}
	})

	t.Run("list raw format", func(t *testing.T) {
		if err := env.Run(list.Command(), []string{"--format", "raw"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, "HANDLE") || strings.Contains(line, "PURPOSE") {
				t.Errorf("Raw format should not contain headers, got: %s", output)
				break
			}
		}
	})
}

func TestPathOutputFormats(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	ws := env.CreateWorkspace("test purpose", nil)

	t.Run("path raw format is just path", func(t *testing.T) {
		if err := env.Run(path.Command(), []string{ws.Handle, "--format", "raw"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		output = strings.TrimSpace(output)
		if strings.Contains(output, "\n") {
			t.Errorf("Raw format should be single line, got: %s", output)
		}
		if !strings.HasPrefix(output, ws.Path) {
			t.Errorf("Raw format should start with path, got: %s", output)
		}
	})

	t.Run("path json format", func(t *testing.T) {
		if err := env.Run(path.Command(), []string{ws.Handle, "--format", "json"}); err != nil {
			t.Errorf("Run failed: %v", err)
		}
		output := env.Output()
		var data map[string]interface{}
		err := json.Unmarshal([]byte(output), &data)
		if err != nil {
			t.Errorf("Expected valid JSON: %v", err)
		}
		if _, ok := data["path"]; !ok {
			t.Errorf("Expected 'path' in json output")
		}
	})
}
