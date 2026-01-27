package apply

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestApplyCommand(t *testing.T) {
	t.Run("has required flags", func(t *testing.T) {
		cmd := Command()
		requiredFlags := []string{"name", "dry-run", "format"}
		for _, f := range requiredFlags {
			if !flagExists(cmd, f) {
				t.Errorf("apply should have --%s flag", f)
			}
		}
	})

	t.Run("dry-run defaults to false", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("dry-run")
		if flag.DefValue != "false" {
			t.Errorf("dry-run default should be false, got: %s", flag.DefValue)
		}
	})

	t.Run("name is optional", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("name")
		if flag.DefValue != "" {
			t.Errorf("name default should be empty, got: %s", flag.DefValue)
		}
	})

	t.Run("use format matches documentation", func(t *testing.T) {
		cmd := Command()
		expected := "apply [<handle>] <capture-id>"
		if cmd.Use != expected {
			t.Errorf("Use = %q, want %q", cmd.Use, expected)
		}
	})
}
