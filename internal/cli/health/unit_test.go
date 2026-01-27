package health

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestHealthCommand(t *testing.T) {
	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("health should have --format flag")
		}
	})

	t.Run("use format is correct", func(t *testing.T) {
		cmd := Command()
		expected := "health [<handle>]"
		if cmd.Use != expected {
			t.Errorf("Use = %q, want %q", cmd.Use, expected)
		}
	})
}
