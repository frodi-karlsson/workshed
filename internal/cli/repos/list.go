package repos

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [<handle>]",
		Short: "List repositories in a workspace",
		Long: `List repositories in a workspace.

Examples:
  workshed repos list
  workshed repos list my-workspace`,
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
				return fmt.Errorf("workspace not found: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()

			if len(ws.Repositories) == 0 {
				return cli.RenderEmptyList(format, "no repositories in workspace", cmd.OutOrStdout(), r.GetLogger())
			}

			if format == "raw" {
				for _, repo := range ws.Repositories {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), repo.Name)
				}
				return nil
			}

			var rows [][]string
			for _, repo := range ws.Repositories {
				var refInfo string
				if repo.Ref != "" {
					refInfo = " @ " + repo.Ref
				}
				rows = append(rows, []string{repo.Name, repo.URL + refInfo})
			}

			output := cli.Output{
				Columns: []cli.ColumnConfig{
					{Type: cli.Rigid, Name: "NAME", Min: 15, Max: 30},
					{Type: cli.Shrinkable, Name: "URL", Min: 30, Max: 0},
				},
				Rows: rows,
			}

			return cli.Render(output, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
