package export

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestExportCommand(t *testing.T) {
	t.Run("has --output flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "output") {
			t.Error("export should have --output flag")
		}
	})

	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("export should have --format flag")
		}
	})

	t.Run("output defaults to empty", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("output")
		if flag.DefValue != "" {
			t.Errorf("output default should be empty, got: %s", flag.DefValue)
		}
	})
}
