package list

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestListCommand(t *testing.T) {
	t.Run("has --purpose filter flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "purpose") {
			t.Error("list should have --purpose flag")
		}
	})

	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("list should have --format flag")
		}
	})

	t.Run("has --page flags", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "page") {
			t.Error("list should have --page flag")
		}
		if !flagExists(cmd, "page-size") {
			t.Error("list should have --page-size flag")
		}
	})
}
