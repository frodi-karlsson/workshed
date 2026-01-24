package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/testutil"
)

func TestGetWorkshedRoot(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		want    string
		wantErr bool
	}{
		{
			name:   "should return custom path when env var is set",
			envVar: "/custom/path",
			want:   "/custom/path",
		},
		{
			name:   "should return default path when env var is not set",
			envVar: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.WithEnvVar(t, "WORKSHED_ROOT", tt.envVar, func() {
				got := GetWorkshedRoot()

				if tt.envVar != "" && got != tt.want {
					t.Errorf("GetWorkshedRoot() = %v, want %v", got, tt.want)
				}

				if tt.envVar == "" && !strings.Contains(got, ".workshed/workspaces") {
					t.Errorf("GetWorkshedRoot() = %v, should contain .workshed/workspaces", got)
				}
			})
		})
	}
}

func TestUsage(t *testing.T) {
	t.Run("should print usage information to stderr", func(t *testing.T) {
		var buf bytes.Buffer
		errWriter = &buf

		Usage()

		output := buf.String()
		if !strings.Contains(output, "workshed") {
			t.Errorf("Usage() should contain 'workshed', got: %s", output)
		}
		if !strings.Contains(output, "create") {
			t.Errorf("Usage() should contain 'create', got: %s", output)
		}
		if !strings.Contains(output, "list") {
			t.Errorf("Usage() should contain 'list', got: %s", output)
		}

		// Reset to default
		errWriter = os.Stderr
	})
}

func TestVersion(t *testing.T) {
	t.Run("should print version to stdout", func(t *testing.T) {
		var buf bytes.Buffer
		outWriter = &buf

		Version()

		output := buf.String()
		if output == "" {
			t.Error("Version() should output version string")
		}
		if !strings.Contains(output, ".") {
			t.Errorf("Version() should contain version number with dot, got: %s", output)
		}

		// Reset to default
		outWriter = os.Stdout
	})
}

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantURL string
		wantRef string
	}{
		{
			name:    "should parse URL without ref",
			input:   "https://github.com/org/repo",
			wantURL: "https://github.com/org/repo",
			wantRef: "",
		},
		{
			name:    "should parse URL with ref",
			input:   "https://github.com/org/repo@main",
			wantURL: "https://github.com/org/repo",
			wantRef: "main",
		},
		{
			name:    "should parse SSH URL with ref",
			input:   "git@github.com:org/repo.git@v1.2.3",
			wantURL: "git@github.com:org/repo.git",
			wantRef: "v1.2.3",
		},
		{
			name:    "should parse URL with branch ref",
			input:   "https://github.com/org/repo@feature/branch",
			wantURL: "https://github.com/org/repo",
			wantRef: "feature/branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotRef := parseRepoFlag(tt.input)
			if gotURL != tt.wantURL {
				t.Errorf("parseRepoFlag() URL = %v, want %v", gotURL, tt.wantURL)
			}
			if gotRef != tt.wantRef {
				t.Errorf("parseRepoFlag() Ref = %v, want %v", gotRef, tt.wantRef)
			}
		})
	}
}

func TestExecErrors(t *testing.T) {
	t.Run("should exit with error when workspace handle is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Exec([]string{"--", "echo", "hello"})

		if !env.ExitCalled() {
			t.Error("Exec should exit with error when handle is missing")
		}
	})

	t.Run("should exit with error when command separator is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Exec([]string{"handle", "echo", "hello"})

		if !env.ExitCalled() {
			t.Error("Exec should exit with error when separator is missing")
		}

		errOutput := env.ErrorOutput()
		if !strings.Contains(errOutput, "Usage") {
			t.Errorf("Error output should contain usage, got: %s", errOutput)
		}
	})

	t.Run("should exit with error when workspace does not exist", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Exec([]string{"nonexistent-handle", "--", "echo", "hello"})

		if !env.ExitCalled() {
			t.Error("Exec should exit with error when workspace does not exist")
		}
	})

	t.Run("should show exec in usage", func(t *testing.T) {
		var buf bytes.Buffer
		errWriter = &buf

		Usage()

		output := buf.String()
		if !strings.Contains(output, "exec") {
			t.Errorf("Usage() should contain 'exec', got: %s", output)
		}

		errWriter = os.Stderr
	})
}

func TestUpdate(t *testing.T) {
	t.Run("should exit with error when purpose is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Update([]string{})

		if !env.ExitCalled() {
			t.Error("Update should exit with error when purpose is missing")
		}

		errOutput := env.ErrorOutput()
		if !strings.Contains(errOutput, "--purpose") {
			t.Errorf("Error output should mention --purpose flag, got: %s", errOutput)
		}
	})

	t.Run("should exit with error when handle is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Update([]string{"--purpose", "New purpose"})

		if !env.ExitCalled() {
			t.Error("Update should exit with error when handle is missing")
		}

		output := env.Output()
		if !strings.Contains(output, "workspace") {
			t.Errorf("Output should mention workspace, got: %s", output)
		}
	})

	t.Run("should exit with error when workspace does not exist", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Update([]string{"--purpose", "New purpose", "nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Update should exit with error when workspace does not exist")
		}
	})

	t.Run("should update purpose successfully", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		Create([]string{"--purpose", "Original purpose"})

		if !env.ExitCalled() {
			handle := ExtractHandleFromLog(t, env.Output())
			if handle != "" {
				env.ResetBuffers()
				Update([]string{"--purpose", "Updated purpose", handle})

				if env.ExitCalled() {
					t.Errorf("Update should succeed, but got error: %s", env.ErrorOutput())
				}

				output := env.Output()
				if !strings.Contains(output, "purpose updated") {
					t.Errorf("Output should contain 'purpose updated', got: %s", output)
				}
				if !strings.Contains(output, "Updated purpose") {
					t.Errorf("Output should contain new purpose, got: %s", output)
				}
			}
		}
	})

	t.Run("should show update in usage", func(t *testing.T) {
		var buf bytes.Buffer
		errWriter = &buf

		Usage()

		output := buf.String()
		if !strings.Contains(output, "update") {
			t.Errorf("Usage() should contain 'update', got: %s", output)
		}

		errWriter = os.Stderr
	})
}
