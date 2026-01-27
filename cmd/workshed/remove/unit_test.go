package remove

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestRemoveCommand(t *testing.T) {
	t.Run("has --yes flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "yes") {
			t.Error("remove should have --yes flag")
		}
	})

	t.Run("-y is shorthand for --yes", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("yes")
		if flag == nil {
			t.Error("remove should have --yes flag")
		} else if string(flag.Shorthand) != "y" {
			t.Errorf("yes flag should have -y shorthand, got: %q", flag.Shorthand)
		}
	})

	t.Run("has --dry-run flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "dry-run") {
			t.Error("remove should have --dry-run flag")
		}
	})
}
