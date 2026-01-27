package export

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/frodi/workshed/internal/cli"
	fsutil "github.com/frodi/workshed/internal/fs"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var output string
	var compact bool

	cmd := &cobra.Command{
		Use:   "export [<handle>]",
		Short: "Export workspace configuration",
		Long: `Export workspace configuration including purpose and repositories.

Examples:
  workshed export
  workshed export --format json | jq '.captures'
  workshed export --output /tmp/context.json
  workshed export --compact --format json | jq '{purpose, repositories}'`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			wsPath, err := r.GetStore().Path(ctx, handle)
			if err != nil {
				return fmt.Errorf("failed to get workspace path: %w", err)
			}

			contextData, err := r.GetStore().ExportContext(ctx, handle)
			if err != nil {
				return fmt.Errorf("export failed: %w", err)
			}

			if compact {
				contextData.Captures = nil
			}

			outputPath := output
			if outputPath == "" {
				outputPath = filepath.Join(wsPath, ".workshed", "context.json")
			}

			data, err := json.MarshalIndent(contextData, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling context: %w", err)
			}

			if err := fsutil.WriteJson(outputPath, data); err != nil {
				return fmt.Errorf("writing context: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			switch format {
			case "json":
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			case "raw":
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), outputPath)
			default:
				return cli.RenderKeyValue(map[string]string{
					"path":  outputPath,
					"repos": strconv.Itoa(len(contextData.Repositories)),
				}, "table", cmd.OutOrStdout())
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output file path")
	cmd.Flags().BoolVar(&compact, "compact", false, "Exclude captures from export")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
