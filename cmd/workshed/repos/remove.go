package repos

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func RemoveCommand() *cobra.Command {
	var repo string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "remove [<handle>] --repo <name>",
		Short: "Remove a repository from a workspace",
		Long: `Remove a repository from a workspace.

Examples:
  workshed repos remove --repo my-repo
  workshed repos remove my-workspace --repo my-repo
  workshed repos remove --repo my-repo --dry-run`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			if repo == "" {
				return fmt.Errorf("missing required flag: --repo")
			}

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			if dryRun {
				r.GetLogger().Info("dry run - would remove repository", "handle", handle, "repo", repo)
				return nil
			}

			if err := r.GetStore().RemoveRepository(ctx, handle, repo); err != nil {
				return fmt.Errorf("failed to remove repository: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "raw" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), repo)
				return nil
			}

			r.GetLogger().Success("repository removed", "handle", handle, "repo", repo)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository name to remove")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}
