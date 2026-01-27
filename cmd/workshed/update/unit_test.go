package update

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestUpdateCommand(t *testing.T) {
	t.Run("has --purpose flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "purpose") {
			t.Error("update should have --purpose flag")
		}
	})

	t.Run("accepts arbitrary args", func(t *testing.T) {
		cmd := Command()
		if cmd.Args == nil {
			t.Error("update should accept arguments")
		}
	})
}
