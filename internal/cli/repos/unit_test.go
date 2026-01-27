package repos

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestReposCommand(t *testing.T) {
	t.Run("has subcommands", func(t *testing.T) {
		cmd := Command()
		subcommands := []string{"list", "add", "remove"}
		for _, sub := range subcommands {
			found := false
			for _, c := range cmd.Commands() {
				if c.Name() == sub {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("repos should have %q subcommand", sub)
			}
		}
	})

	t.Run("add has --repo flag", func(t *testing.T) {
		cmd := Command()
		for _, c := range cmd.Commands() {
			if c.Name() == "add" {
				if !flagExists(c, "repo") {
					t.Error("repos add should have --repo flag")
				}
				return
			}
		}
		t.Error("repos add subcommand not found")
	})

	t.Run("add has -r shorthand for --repo", func(t *testing.T) {
		cmd := Command()
		for _, c := range cmd.Commands() {
			if c.Name() == "add" {
				flag := c.Flags().Lookup("repo")
				if flag != nil && string(flag.Shorthand) != "r" {
					t.Errorf("repos add --repo should have -r shorthand, got: %q", flag.Shorthand)
				}
				return
			}
		}
	})

	t.Run("remove has --repo flag", func(t *testing.T) {
		cmd := Command()
		for _, c := range cmd.Commands() {
			if c.Name() == "remove" {
				if !flagExists(c, "repo") {
					t.Error("repos remove should have --repo flag")
				}
				return
			}
		}
		t.Error("repos remove subcommand not found")
	})
}
