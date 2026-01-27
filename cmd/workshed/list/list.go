package list

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var purpose string
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		Long: `List workspaces.

Examples:
  workshed list
  workshed list --purpose payment
  workshed list --purpose "API" --format json
  workshed list --page 2 --page-size 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()

			opts := workspace.ListOptions{
				PurposeFilter: purpose,
			}

			workspaces, err := r.GetStore().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(workspaces) == 0 {
				format := cmd.Flags().Lookup("format").Value.String()
				return cli.RenderEmptyList(format, "no workspaces found", cmd.OutOrStdout(), r.GetLogger())
			}

			total := len(workspaces)
			if page < 1 {
				page = 1
			}
			if pageSize < 1 {
				pageSize = 20
			}

			startIdx := (page - 1) * pageSize
			endIdx := startIdx + pageSize

			if startIdx >= total {
				format := cmd.Flags().Lookup("format").Value.String()
				return cli.RenderEmptyList(format, fmt.Sprintf("page %d is empty (total: %d items)", page, total), cmd.OutOrStdout(), r.GetLogger())
			}

			if endIdx > total {
				endIdx = total
			}

			pagedWorkspaces := workspaces[startIdx:endIdx]

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "raw" {
				for _, ws := range pagedWorkspaces {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), ws.Handle)
				}
				return nil
			}

			var rows [][]string
			for _, ws := range pagedWorkspaces {
				repoCount := len(ws.Repositories)
				var repoInfo string
				if repoCount == 1 {
					repoInfo = ws.Repositories[0].Name
				} else if repoCount > 1 {
					repoInfo = fmt.Sprintf("%d repos", repoCount)
				} else {
					repoInfo = "(empty)"
				}
				created := ws.CreatedAt.Format("2006-01-02 15:04")
				rows = append(rows, []string{ws.Handle, ws.Purpose, repoInfo, created})
			}

			output := cli.Output{
				Columns: cli.ListColumns,
				Rows:    rows,
			}

			if err := cli.Render(output, format, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to render output: %w", err)
			}

			if format != "json" && total > pageSize {
				r.GetLogger().Info(fmt.Sprintf("showing %d-%d of %d workspaces (page %d of %d)", startIdx+1, endIdx, total, page, (total+pageSize-1)/(pageSize)))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&purpose, "purpose", "", "Filter by purpose")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Items per page")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
