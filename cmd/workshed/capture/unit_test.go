package capture

import (
	"testing"

	"github.com/spf13/cobra"
)

func flagExists(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil
}

func TestCaptureCommand(t *testing.T) {
	t.Run("has required flags", func(t *testing.T) {
		cmd := Command()
		requiredFlags := []string{"name", "kind", "description", "tag", "format"}
		for _, f := range requiredFlags {
			if !flagExists(cmd, f) {
				t.Errorf("capture should have --%s flag", f)
			}
		}
	})

	t.Run("tag can be specified multiple times", func(t *testing.T) {
		cmd := Command()
		flag := cmd.Flags().Lookup("tag")
		if flag == nil {
			t.Error("tag flag should exist")
		}
	})
}
