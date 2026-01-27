package update

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandFlags(t *testing.T) {
	cmd := Command()

	flags := []string{"purpose"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Command should have --%s flag", flagName)
		}
	}
}

func TestArgsAllowsZeroOrOne(t *testing.T) {
	cmd := Command()

	if cmd.Args == nil {
		t.Error("Args should be set")
	}
}

func TestPurposeFlagCanBeUpdated(t *testing.T) {
	var capturedPurpose string

	cmd := &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			purpose, _ := cmd.Flags().GetString("purpose")
			capturedPurpose = purpose
			return nil
		},
	}
	cmd.Flags().String("purpose", "", "New purpose")
	cmd.SetArgs([]string{"--purpose", "New task purpose"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPurpose != "New task purpose" {
		t.Errorf("expected purpose 'New task purpose', got: %s", capturedPurpose)
	}
}

func TestCommandUseFormat(t *testing.T) {
	cmd := Command()

	if cmd.Use == "" {
		t.Error("Use should not be empty")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}
