package clitest

import (
	"strings"
	"testing"

	"github.com/frodi/workshed/cmd/workshed/apply"
	"github.com/frodi/workshed/cmd/workshed/capture"
	"github.com/frodi/workshed/cmd/workshed/captures"
	"github.com/frodi/workshed/cmd/workshed/export"
	"github.com/frodi/workshed/cmd/workshed/health"
	"github.com/frodi/workshed/cmd/workshed/inspect"
	"github.com/frodi/workshed/cmd/workshed/path"
	"github.com/frodi/workshed/cmd/workshed/remove"
	"github.com/frodi/workshed/cmd/workshed/repos"
	"github.com/frodi/workshed/cmd/workshed/update"
	"github.com/spf13/cobra"
)

func TestInvalidHandleErrors(t *testing.T) {
	tests := []struct {
		name    string
		cmd     func() *cobra.Command
		args    []string
		wantErr string
	}{
		{"captures invalid", captures.Command, []string{"nonexistent"}, "workspace"},
		{"path invalid", path.Command, []string{"nonexistent"}, "workspace"},
		{"health invalid", health.Command, []string{"nonexistent"}, "workspace"},
		{"inspect invalid", inspect.Command, []string{"nonexistent"}, "workspace"},
		{"export invalid", export.Command, []string{"nonexistent"}, "workspace"},
		{"remove invalid", remove.Command, []string{"-y", "nonexistent"}, "workspace"},
		{"update invalid", update.Command, []string{"--purpose", "x", "nonexistent"}, "workspace"},
		{"repos list invalid", repos.ListCommand, []string{"nonexistent"}, "workspace"},
		{"repos add invalid", repos.AddCommand, []string{"--repo", "url", "nonexistent"}, "workspace"},
		{"repos remove invalid", repos.RemoveCommand, []string{"--repo", "url", "nonexistent"}, "workspace"},
		{"apply invalid", apply.Command, []string{"nonexistent", "cap-id"}, "workspace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewCLIEnv(t)
			err := env.Run(tt.cmd(), tt.args)
			if err == nil {
				t.Error("Expected error for invalid handle")
			}
			if !strings.Contains(env.ErrorOutput(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got stderr: %s", tt.wantErr, env.ErrorOutput())
			}
		})
	}
}

func TestMissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cobra.Command
		args []string
	}{
		{"capture missing --name", capture.Command, []string{}},
		{"update missing --purpose", update.Command, []string{}},
		{"repos add missing --repo", repos.AddCommand, []string{}},
		{"repos remove missing --repo", repos.RemoveCommand, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewCLIEnv(t)
			err := env.Run(tt.cmd(), tt.args)
			if err == nil {
				t.Error("Expected error for missing required flag")
			}
		})
	}
}

func TestInvalidArgumentCombinations(t *testing.T) {
	env := NewCLIEnv(t)
	defer env.Cleanup()

	t.Run("update with empty purpose", func(t *testing.T) {
		ws := env.CreateWorkspace("test", nil)
		err := env.Run(update.Command(), []string{"--purpose", "", ws.Handle})
		if err == nil {
			t.Error("update with empty purpose should fail")
		}
	})
}
