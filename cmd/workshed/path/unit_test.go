package path

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestPathCommand(t *testing.T) {
	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("path should have --format flag")
		}
	})

	t.Run("format defaults to raw", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("format")
		if flag == nil {
			t.Error("path should have --format flag")
		} else if flag.DefValue != "raw" {
			t.Errorf("format default should be 'raw', got: %s", flag.DefValue)
		}
	})

	t.Run("accepts arbitrary args", func(t *testing.T) {
		cmd := Command()
		if cmd.Args == nil {
			t.Error("path should accept arguments")
		}
	})
}
