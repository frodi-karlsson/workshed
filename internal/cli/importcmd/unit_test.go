package importcmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestImportCommand(t *testing.T) {
	t.Run("has --file flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "file") {
			t.Error("import should have --file flag")
		}
	})

	t.Run("has --format flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "format") {
			t.Error("import should have --format flag")
		}
	})

	t.Run("has --preserve-handle flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "preserve-handle") {
			t.Error("import should have --preserve-handle flag")
		}
	})

	t.Run("has --force flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "force") {
			t.Error("import should have --force flag")
		}
	})
}
