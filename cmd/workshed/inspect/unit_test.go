package inspect

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestInspectCommand(t *testing.T) {
	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("inspect should have --format flag")
		}
	})

	t.Run("accepts arbitrary args", func(t *testing.T) {
		cmd := Command()
		if cmd.Args == nil {
			t.Error("inspect should accept arguments")
		}
	})
}
