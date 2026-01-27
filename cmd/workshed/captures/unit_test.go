package captures

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestCapturesCommand(t *testing.T) {
	t.Run("has --filter flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "filter") {
			t.Error("captures should have --filter flag")
		}
	})

	t.Run("has --reverse flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "reverse") {
			t.Error("captures should have --reverse flag")
		}
	})

	t.Run("reverse defaults to false", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("reverse")
		if flag.DefValue != "false" {
			t.Errorf("reverse default should be false, got: %s", flag.DefValue)
		}
	})

	t.Run("accepts arbitrary args", func(t *testing.T) {
		cmd := Command()
		if cmd.Args == nil {
			t.Error("captures should accept arguments")
		}
	})
}
