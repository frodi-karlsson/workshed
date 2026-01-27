package create

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestCreateCommand(t *testing.T) {
	t.Run("has --purpose flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "purpose") {
			t.Error("create should have --purpose flag")
		}
	})

	t.Run("has --repo flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "repo") {
			t.Error("create should have --repo flag")
		}
	})

	t.Run("has -r shorthand for --repo", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("repo")
		if flag == nil {
			t.Error("create should have --repo flag")
		} else if string(flag.Shorthand) != "r" {
			t.Errorf("repo flag should have -r shorthand, got: %q", flag.Shorthand)
		}
	})

	t.Run("has --template flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "template") {
			t.Error("create should have --template flag")
		}
	})

	t.Run("has --map flag for template variables", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "map") {
			t.Error("create should have --map flag")
		}
	})

	t.Run("has --local-map flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "local-map") {
			t.Error("create should have --local-map flag")
		}
	})
}
