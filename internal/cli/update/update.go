package update

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var purpose string

	cmd := &cobra.Command{
		Use:   "update [<handle>]",
		Short: "Update workspace purpose",
		Long: `Update the purpose of a workspace.

Examples:
  workshed update --purpose "New focus area"
  workshed update --purpose "Completed" my-workspace`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			if purpose == "" {
				return fmt.Errorf("missing required flag: --purpose")
			}

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			if err := r.GetStore().UpdatePurpose(ctx, handle, purpose); err != nil {
				return fmt.Errorf("failed to update workspace purpose: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			return cli.RenderKeyValue(map[string]string{
				"handle":  handle,
				"purpose": purpose,
			}, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&purpose, "purpose", "", "New workspace purpose")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")
	_ = cmd.MarkFlagRequired("purpose")

	return cmd
}
