package importcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var preserveHandle bool
	var force bool
	var file string

	cmd := &cobra.Command{
		Use:   "import [<file.json>]",
		Short: "Import workspace from JSON",
		Long: `Create a workspace from an exported JSON file.

Examples:
  workshed import workspace.json
  workshed import workspace.json --preserve-handle
  cat workspace.json | workshed import -
  workshed import --file workspace.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			inputFile := file
			if len(args) > 0 {
				inputFile = args[0]
			}

			if inputFile == "" {
				return fmt.Errorf("missing required argument: <file.json> or --file flag")
			}

			var data []byte
			var err error

			if inputFile == "-" {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading from stdin: %w", err)
				}
			} else {
				data, err = os.ReadFile(inputFile)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
			}

			if !json.Valid(data) {
				return fmt.Errorf("invalid JSON file: %s", inputFile)
			}

			var wsContext workspace.WorkspaceContext
			if err := json.Unmarshal(data, &wsContext); err != nil {
				return fmt.Errorf("parsing JSON: %w", err)
			}

			if wsContext.Purpose == "" {
				return fmt.Errorf("missing required field: purpose")
			}

			if wsContext.Repositories == nil {
				return fmt.Errorf("missing required field: repositories")
			}

			for _, repo := range wsContext.Repositories {
				if repo.URL == "" {
					return fmt.Errorf("invalid repository: URL is required")
				}
			}

			ctx := context.Background()

			ws, err := r.GetStore().ImportContext(ctx, workspace.ImportOptions{
				Context:        &wsContext,
				InvocationCWD:  r.GetInvocationCWD(),
				PreserveHandle: preserveHandle,
				Force:          force,
			})
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			output := cli.Output{
				Columns: []cli.ColumnConfig{
					{Type: cli.Rigid, Name: "KEY", Min: 10, Max: 20},
					{Type: cli.Rigid, Name: "VALUE", Min: 20, Max: 0},
				},
				Rows: [][]string{
					{"handle", ws.Handle},
					{"purpose", ws.Purpose},
					{"repos", strconv.Itoa(len(ws.Repositories))},
					{"path", ws.Path},
				},
			}

			if err := cli.Render(output, format, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to render output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&preserveHandle, "preserve-handle", false, "Preserve the handle from the imported file")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing workspace if it exists")
	cmd.Flags().StringVar(&file, "file", "", "Input file path (- for stdin)")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
