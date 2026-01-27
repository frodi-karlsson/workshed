package main

import (
	"testing"

	"github.com/frodi/workshed/cmd/workshed/apply"
	"github.com/frodi/workshed/cmd/workshed/capture"
	"github.com/frodi/workshed/cmd/workshed/captures"
	"github.com/frodi/workshed/cmd/workshed/create"
	"github.com/frodi/workshed/cmd/workshed/exec"
	"github.com/frodi/workshed/cmd/workshed/export"
	"github.com/frodi/workshed/cmd/workshed/health"
	"github.com/frodi/workshed/cmd/workshed/importcmd"
	"github.com/frodi/workshed/cmd/workshed/inspect"
	"github.com/frodi/workshed/cmd/workshed/list"
	"github.com/frodi/workshed/cmd/workshed/path"
	"github.com/frodi/workshed/cmd/workshed/remove"
	"github.com/frodi/workshed/cmd/workshed/repos"
	"github.com/frodi/workshed/cmd/workshed/update"
	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestArgumentValidation(t *testing.T) {
	cmds := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"create", create.Command()},
		{"list", list.Command()},
		{"captures", captures.Command()},
		{"path", path.Command()},
		{"inspect", inspect.Command()},
		{"health", health.Command()},
		{"export", export.Command()},
		{"remove", remove.Command()},
		{"update", update.Command()},
		{"apply", apply.Command()},
		{"exec", exec.Command()},
		{"repos list", repos.ListCommand()},
		{"repos add", repos.AddCommand()},
		{"repos remove", repos.RemoveCommand()},
	}

	for _, tc := range cmds {
		t.Run(tc.name+" accepts arguments", func(t *testing.T) {
			if tc.cmd.Args == nil {
				t.Errorf("%s should accept arguments", tc.name)
			}
		})
	}
}

func TestCommandDescriptions(t *testing.T) {
	t.Run("create has description", func(t *testing.T) {
		cmd := create.Command()
		if cmd.Use == "" {
			t.Error("create Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("create Short should not be empty")
		}
	})

	t.Run("list has description", func(t *testing.T) {
		cmd := list.Command()
		if cmd.Use == "" {
			t.Error("list Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("list Short should not be empty")
		}
	})

	t.Run("inspect has description", func(t *testing.T) {
		cmd := inspect.Command()
		if cmd.Use == "" {
			t.Error("inspect Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("inspect Short should not be empty")
		}
	})

	t.Run("path has description", func(t *testing.T) {
		cmd := path.Command()
		if cmd.Use == "" {
			t.Error("path Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("path Short should not be empty")
		}
	})

	t.Run("capture has description", func(t *testing.T) {
		cmd := capture.Command()
		if cmd.Use == "" {
			t.Error("capture Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("capture Short should not be empty")
		}
	})

	t.Run("captures has description", func(t *testing.T) {
		cmd := captures.Command()
		if cmd.Use == "" {
			t.Error("captures Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("captures Short should not be empty")
		}
	})

	t.Run("apply has description", func(t *testing.T) {
		cmd := apply.Command()
		if cmd.Use == "" {
			t.Error("apply Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("apply Short should not be empty")
		}
	})

	t.Run("remove has description", func(t *testing.T) {
		cmd := remove.Command()
		if cmd.Use == "" {
			t.Error("remove Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("remove Short should not be empty")
		}
	})

	t.Run("update has description", func(t *testing.T) {
		cmd := update.Command()
		if cmd.Use == "" {
			t.Error("update Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("update Short should not be empty")
		}
	})

	t.Run("health has description", func(t *testing.T) {
		cmd := health.Command()
		if cmd.Use == "" {
			t.Error("health Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("health Short should not be empty")
		}
	})

	t.Run("export has description", func(t *testing.T) {
		cmd := export.Command()
		if cmd.Use == "" {
			t.Error("export Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("export Short should not be empty")
		}
	})

	t.Run("repos has description", func(t *testing.T) {
		cmd := repos.Command()
		if cmd.Use == "" {
			t.Error("repos Use should not be empty")
		}
		if cmd.Short == "" {
			t.Error("repos Short should not be empty")
		}
	})
}

func TestFormatFlagConsistency(t *testing.T) {
	t.Run("create format default", func(t *testing.T) {
		cmd := create.Command()
		f := cmd.Flags().Lookup("format")
		if f == nil {
			t.Error("create should have --format flag")
			return
		}
		if f.DefValue != "table" {
			t.Errorf("create --format default = %q, want %q", f.DefValue, "table")
		}
	})

	t.Run("list format default", func(t *testing.T) {
		cmd := list.Command()
		f := cmd.Flags().Lookup("format")
		if f == nil {
			t.Error("list should have --format flag")
			return
		}
		if f.DefValue != "table" {
			t.Errorf("list --format default = %q, want %q", f.DefValue, "table")
		}
	})

	t.Run("exec format default", func(t *testing.T) {
		cmd := exec.Command()
		f := cmd.Flags().Lookup("format")
		if f == nil {
			t.Error("exec should have --format flag")
			return
		}
		if f.DefValue != "stream" {
			t.Errorf("exec --format default = %q, want %q", f.DefValue, "stream")
		}
	})
}

func TestFlagValueParsing(t *testing.T) {
	t.Run("create purpose flag", func(t *testing.T) {
		cmd := create.Command()
		err := cmd.Flags().Set("purpose", "test purpose")
		if err != nil {
			t.Errorf("Set purpose = %q failed: %v", "test purpose", err)
		}
	})

	t.Run("create repo flag", func(t *testing.T) {
		cmd := create.Command()
		err := cmd.Flags().Set("repo", "https://github.com/org/repo")
		if err != nil {
			t.Errorf("Set repo = %q failed: %v", "https://github.com/org/repo", err)
		}
	})

	t.Run("capture name flag", func(t *testing.T) {
		cmd := capture.Command()
		err := cmd.Flags().Set("name", "my capture")
		if err != nil {
			t.Errorf("Set name = %q failed: %v", "my capture", err)
		}
	})

	t.Run("captures filter flag", func(t *testing.T) {
		cmd := captures.Command()
		err := cmd.Flags().Set("filter", "some-filter")
		if err != nil {
			t.Errorf("Set filter = %q failed: %v", "some-filter", err)
		}
	})
}

func TestLongFormFlagsExist(t *testing.T) {
	tests := []struct {
		name  string
		cmd   *cobra.Command
		flags []string
	}{
		{"list", list.Command(), []string{"format", "page", "page-size", "purpose"}},
		{"captures", captures.Command(), []string{"format", "filter", "reverse"}},
		{"create", create.Command(), []string{"format", "purpose", "repo", "template", "map", "local-map"}},
		{"export", export.Command(), []string{"format", "output"}},
		{"import", importcmd.Command(), []string{"format", "file", "preserve-handle", "force"}},
		{"capture", capture.Command(), []string{"format", "name", "kind", "description", "tag"}},
		{"apply", apply.Command(), []string{"format", "name", "dry-run"}},
		{"health", health.Command(), []string{"format"}},
		{"inspect", inspect.Command(), []string{"format"}},
		{"path", path.Command(), []string{"format"}},
		{"remove", remove.Command(), []string{"yes", "dry-run"}},
		{"update", update.Command(), []string{"purpose"}},
		{"repos list", repos.ListCommand(), []string{"format"}},
		{"repos add", repos.AddCommand(), []string{"format", "repo"}},
		{"repos remove", repos.RemoveCommand(), []string{"format", "repo", "dry-run"}},
	}

	for _, tc := range tests {
		for _, flag := range tc.flags {
			t.Run(tc.name+" "+flag, func(t *testing.T) {
				if !flagExists(tc.cmd, flag) {
					t.Errorf("%s should have --%s flag", tc.name, flag)
				}
			})
		}
	}
}
