package inspect

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [<handle>]",
		Short: "Show workspace details",
		Long: `Show workspace details including repositories and creation time.

Examples:
  workshed inspect
  workshed inspect aquatic-fish-motion`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			ws, err := r.GetStore().Get(ctx, handle)
			if err != nil {
				return fmt.Errorf("failed to get workspace: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()

			data := map[string]string{
				"handle":  ws.Handle,
				"purpose": ws.Purpose,
				"path":    ws.Path,
				"created": ws.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			for _, repo := range ws.Repositories {
				var repoInfo string
				if repo.Ref != "" {
					repoInfo = repo.Name + " @ " + repo.Ref
				} else {
					repoInfo = repo.Name
				}
				data["repo"] = repoInfo
			}

			return cli.RenderKeyValue(data, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
