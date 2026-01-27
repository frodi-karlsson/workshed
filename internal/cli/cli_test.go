//go:build !integration
// +build !integration

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/testutil"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
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
				r := NewRunner("")
				got := r.getWorkshedRoot()

				if tt.envVar != "" && got != tt.want {
					t.Errorf("getWorkshedRoot() = %v, want %v", got, tt.want)
				}

				if tt.envVar == "" && !strings.Contains(got, ".workshed/workspaces") {
					t.Errorf("getWorkshedRoot() = %v, should contain .workshed/workspaces", got)
				}
			})
		})
	}
}

func TestUsage(t *testing.T) {
	t.Run("should print usage information to stderr", func(t *testing.T) {
		var buf bytes.Buffer
		r := &Runner{Stderr: &buf}

		r.Usage()

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
	})
}

func TestVersion(t *testing.T) {
	t.Run("should print version to stdout", func(t *testing.T) {
		var buf bytes.Buffer
		r := &Runner{Stdout: &buf}

		r.Version()

		output := buf.String()
		if output == "" {
			t.Error("Version() should output version string")
		}
		if !strings.Contains(output, ".") {
			t.Errorf("Version() should contain version number with dot, got: %s", output)
		}
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
			gotURL, gotRef := workspace.ParseRepoFlag(tt.input)
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
		env.Runner().Exec([]string{"--", "echo", "hello"})

		if !env.ExitCalled() {
			t.Error("Exec should exit with error when handle is missing")
		}
	})

	t.Run("should exit with error when command separator is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Exec([]string{"handle", "echo", "hello"})

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
		env.Runner().Exec([]string{"nonexistent-handle", "--", "echo", "hello"})

		if !env.ExitCalled() {
			t.Error("Exec should exit with error when workspace does not exist")
		}
	})

	t.Run("should show exec in usage", func(t *testing.T) {
		var buf bytes.Buffer
		r := &Runner{Stderr: &buf}

		r.Usage()

		output := buf.String()
		if !strings.Contains(output, "exec") {
			t.Errorf("Usage() should contain 'exec', got: %s", output)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("should exit with error when purpose is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Update([]string{})

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
		env.Runner().Update([]string{"--purpose", "New purpose"})

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
		env.Runner().Update([]string{"--purpose", "New purpose", "nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("Update should exit with error when workspace does not exist")
		}
	})

	t.Run("should update purpose successfully", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().Create([]string{"--purpose", "Original purpose"})

		if !env.ExitCalled() {
			handle := env.ExtractHandleFromOutput(t)
			if handle != "" {
				env.ResetBuffers()
				env.Runner().Update([]string{"--purpose", "Updated purpose", handle})

				if env.ExitCalled() {
					t.Errorf("Update should succeed, but got error: %s", env.ErrorOutput())
				}

				env.AssertLastOutputRowContains(1, 0, "purpose")
				env.AssertLastOutputRowContains(1, 1, "Updated purpose")
			}
		}
	})

	t.Run("should show update in usage", func(t *testing.T) {
		var buf bytes.Buffer
		r := &Runner{Stderr: &buf}

		r.Usage()

		output := buf.String()
		if !strings.Contains(output, "update") {
			t.Errorf("Usage() should contain 'update', got: %s", output)
		}
	})
}

func TestValidFormatsForCommand(t *testing.T) {
	tests := []struct {
		cmd         string
		wantFormats []string
	}{
		{"exec", []string{"stream", "json"}},
		{"path", []string{"raw", "table", "json"}},
		{"list", []string{"table", "json", "raw"}},
		{"inspect", []string{"table", "json", "raw"}},
		{"captures", []string{"table", "json", "raw"}},
		{"repos", []string{"table", "json", "raw"}},
		{"unknown", []string{"table", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := ValidFormatsForCommand(tt.cmd)
			if len(got) != len(tt.wantFormats) {
				t.Errorf("ValidFormatsForCommand(%q) returned %d formats, want %d", tt.cmd, len(got), len(tt.wantFormats))
				return
			}
			for i, want := range tt.wantFormats {
				if got[i] != want {
					t.Errorf("ValidFormatsForCommand(%q)[%d] = %q, want %q", tt.cmd, i, got[i], want)
				}
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  Format
		cmd     string
		wantErr bool
	}{
		{"valid table for list", FormatTable, "list", false},
		{"valid json for list", FormatJSON, "list", false},
		{"valid raw for list", FormatRaw, "list", false},
		{"valid stream for exec", FormatStream, "exec", false},
		{"valid json for exec", FormatJSON, "exec", false},
		{"invalid stream for list", FormatStream, "list", true},
		{"invalid raw for exec", FormatRaw, "exec", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat(%q, %q) error = %v, wantErr %v", tt.format, tt.cmd, err, tt.wantErr)
			}
		})
	}
}

func TestReposList(t *testing.T) {
	t.Run("should show repos list in usage", func(t *testing.T) {
		var buf bytes.Buffer
		r := &Runner{Stderr: &buf}

		r.ReposUsage()

		output := buf.String()
		if !strings.Contains(output, "list") {
			t.Errorf("ReposUsage() should contain 'list', got: %s", output)
		}
		if !strings.Contains(output, "repos list") {
			t.Errorf("ReposUsage() should contain 'repos list', got: %s", output)
		}
	})

	t.Run("should exit with error when workspace handle is missing", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().ReposList([]string{})

		if !env.ExitCalled() {
			t.Error("ReposList should exit with error when workspace handle is missing")
		}
	})

	t.Run("should exit with error when workspace does not exist", func(t *testing.T) {
		env := NewCLITestEnvironment(t)
		defer env.Cleanup()

		env.ResetBuffers()
		env.Runner().ReposList([]string{"nonexistent-handle"})

		if !env.ExitCalled() {
			t.Error("ReposList should exit with error when workspace does not exist")
		}
	})
}

func TestCapturesFilter(t *testing.T) {
	t.Run("should parse filter flag", func(t *testing.T) {
		fs := NewFlagSetForTest(t, "captures")
		filterFlag := fs.String("filter", "", "Filter captures")

		if err := fs.Parse([]string{"--filter", "test-repo"}); err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if *filterFlag != "test-repo" {
			t.Errorf("filter = %q, want %q", *filterFlag, "test-repo")
		}
	})
}

func TestRemoveDryRun(t *testing.T) {
	t.Run("should parse dry-run flag", func(t *testing.T) {
		fs := NewFlagSetForTest(t, "remove")
		dryRun := fs.Bool("dry-run", false, "Dry run")

		if err := fs.Parse([]string{"--dry-run"}); err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if !*dryRun {
			t.Error("dry-run should be true after parsing --dry-run")
		}
	})
}

func TestReposRemoveDryRun(t *testing.T) {
	t.Run("should parse dry-run flag for repos remove", func(t *testing.T) {
		fs := NewFlagSetForTest(t, "repos remove")
		dryRun := fs.Bool("dry-run", false, "Dry run")

		if err := fs.Parse([]string{"--dry-run"}); err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if !*dryRun {
			t.Error("dry-run should be true after parsing --dry-run")
		}
	})
}

func TestListPagination(t *testing.T) {
	t.Run("should parse pagination flags", func(t *testing.T) {
		fs := NewFlagSetForTest(t, "list")
		page := fs.Int("page", 1, "Page number")
		pageSize := fs.Int("page-size", 20, "Page size")

		if err := fs.Parse([]string{"--page", "3", "--page-size", "50"}); err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if *page != 3 {
			t.Errorf("page = %d, want 3", *page)
		}
		if *pageSize != 50 {
			t.Errorf("pageSize = %d, want 50", *pageSize)
		}
	})
}

func NewFlagSetForTest(t *testing.T, name string) *flag.FlagSet {
	t.Helper()
	return flag.NewFlagSet(name, flag.ContinueOnError)
}
