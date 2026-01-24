package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/frodi/workshed/internal/testutil"
)

func TestGetWorkshedRootShouldReturnCustomPathWhenWORKSHED_ROOTEnvVarSet(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		want    string
		wantErr bool
	}{
		{
			name:   "env var set",
			envVar: "/custom/path",
			want:   "/custom/path",
		},
		{
			name:   "env var not set",
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

func TestUsageShouldPrintUsageInformationToStderr(t *testing.T) {
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
}

func TestVersionShouldPrintVersionToStdout(t *testing.T) {
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
}

func TestParseRepoFlagShouldParseRepositoryURLsWithRefsCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantURL string
		wantRef string
	}{
		{
			name:    "url only",
			input:   "https://github.com/org/repo",
			wantURL: "https://github.com/org/repo",
			wantRef: "",
		},
		{
			name:    "url with ref",
			input:   "https://github.com/org/repo@main",
			wantURL: "https://github.com/org/repo",
			wantRef: "main",
		},
		{
			name:    "ssh url with ref",
			input:   "git@github.com:org/repo.git@v1.2.3",
			wantURL: "git@github.com:org/repo.git",
			wantRef: "v1.2.3",
		},
		{
			name:    "url with tag",
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

func TestTruncateShouldTruncateStringsToSpecifiedLength(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "truncation needed",
			input:  "this is a very long string",
			maxLen: 10,
			want:   "this is...",
		},
		{
			name:   "exact length",
			input:  "exactly10c",
			maxLen: 10,
			want:   "exactly10c",
		},
		{
			name:   "maxLen less than 3",
			input:  "test",
			maxLen: 2,
			want:   "te",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}
