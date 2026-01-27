package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var name string
	var kind string
	var description string
	var tags []string

	cmd := &cobra.Command{
		Use:   "capture [<handle>] --name <name>",
		Short: "Create a capture",
		Long: `Create a durable capture of git state for all repositories in a workspace.

Examples:
  workshed capture --name "Before refactor"
  workshed capture --name "Checkpoint 1" --description "API changes"
  workshed capture --name "Starting point" --tag test`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			if name == "" {
				return fmt.Errorf("missing required flag: --name")
			}

			if kind == "" {
				kind = workspace.CaptureKindManual
			}

			ctx := context.Background()

			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			capture, err := r.GetStore().CaptureState(ctx, handle, workspace.CaptureOptions{
				Name:        name,
				Kind:        kind,
				Description: description,
				Tags:        tags,
			})
			if err != nil {
				return fmt.Errorf("capture failed: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "json" {
				data, _ := json.MarshalIndent(capture, "", "  ")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			if format == "raw" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), capture.ID)
				return nil
			}

			return cli.RenderKeyValue(map[string]string{
				"id":    capture.ID,
				"name":  capture.Name,
				"kind":  capture.Kind,
				"repos": strconv.Itoa(len(capture.GitState)),
			}, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Capture name")
	cmd.Flags().StringVar(&kind, "kind", "", "Capture kind (default: state)")
	cmd.Flags().StringVar(&description, "description", "", "Capture description")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags for the capture")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
