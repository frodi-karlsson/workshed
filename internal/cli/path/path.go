package path

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path [<handle>]",
		Short: "Print the workspace directory path",
		Long: `Print the workspace directory path.

Examples:
  workshed path
  workshed path my-workspace
  cd $(workshed path)
  ls $(workshed path)`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			path, err := r.GetStore().Path(ctx, handle)
			if err != nil {
				return fmt.Errorf("failed to get workspace path: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()

			if format == "raw" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
				return nil
			}

			return cli.RenderKeyValue(map[string]string{"path": path}, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().String("format", "raw", "Output format (raw|table|json)")

	return cmd
}
