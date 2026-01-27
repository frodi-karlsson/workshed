package exec

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestExecCommand(t *testing.T) {
	t.Run("has --all flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "all") {
			t.Error("exec should have --all flag")
		}
	})

	t.Run("-a is shorthand for --all", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("all")
		if flag == nil {
			t.Error("exec should have --all flag")
		} else if string(flag.Shorthand) != "a" {
			t.Errorf("all flag should have -a shorthand, got: %q", flag.Shorthand)
		}
	})

	t.Run("has --repo flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "repo") {
			t.Error("exec should have --repo flag")
		}
	})

	t.Run("has --no-record flag", func(t *testing.T) {
		cmd := Command()
		if !flagExists(cmd, "no-record") {
			t.Error("exec should have --no-record flag")
		}
	})

	t.Run("all defaults to false", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("all")
		if flag.DefValue != "false" {
			t.Errorf("all default should be false, got: %s", flag.DefValue)
		}
	})

	t.Run("format defaults to stream", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("format")
		if flag == nil {
			t.Error("exec should have --format flag")
		} else if flag.DefValue != "stream" {
			t.Errorf("format default should be 'stream', got: %s", flag.DefValue)
		}
	})
}
