package repos

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Manage repositories in a workspace",
		Long: `Manage repositories in a workspace.

Examples:
  workshed repos list
  workshed repos add --repo github.com/org/repo@main
  workshed repos remove --repo my-repo`,
	}

	cmd.AddCommand(ListCommand())
	cmd.AddCommand(AddCommand())
	cmd.AddCommand(RemoveCommand())

	return cmd
}
