package captures

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var reverse bool

	cmd := &cobra.Command{
		Use:   "captures [<handle>]",
		Short: "List captures",
		Long: `List captures for a workspace.

Examples:
  # List captures in current workspace
  workshed captures

  # List captures in specific workspace
  workshed captures my-workspace

  # Filter captures by name
  workshed captures --filter api

  # Filter captures by tag
  workshed captures --filter tag:debug`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			captures, err := r.GetStore().ListCaptures(ctx, handle)
			if err != nil {
				return fmt.Errorf("failed to list captures: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()

			if len(captures) == 0 {
				return cli.RenderEmptyList(format, "no captures found", cmd.OutOrStdout(), r.GetLogger())
			}

			var filteredCaptures []workspace.Capture
			if filter != "" {
				for _, cap := range captures {
					if cli.MatchesCaptureFilter(cap, filter) {
						filteredCaptures = append(filteredCaptures, cap)
					}
				}
			} else {
				filteredCaptures = captures
			}

			if len(filteredCaptures) == 0 {
				return cli.RenderEmptyList(format, "no captures match filter: "+filter, cmd.OutOrStdout(), r.GetLogger())
			}

			displayCaptures := filteredCaptures
			if reverse {
				for i, j := 0, len(displayCaptures)-1; i < j; i, j = i+1, j-1 {
					displayCaptures[i], displayCaptures[j] = displayCaptures[j], displayCaptures[i]
				}
			}

			if format == "raw" {
				for _, cap := range displayCaptures {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), cap.ID)
				}
				return nil
			}

			var rows [][]string
			for _, cap := range displayCaptures {
				created := cap.Timestamp.Format("2006-01-02 15:04")
				rows = append(rows, []string{cap.ID, cap.Name, cap.Kind, fmt.Sprintf("%d", len(cap.GitState)), created})
			}

			output := cli.Output{
				Columns: cli.CapturesColumns,
				Rows:    rows,
			}

			return cli.Render(output, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter captures by name or tag")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse order")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
