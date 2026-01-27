package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/logger"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var name string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply [<handle>] <capture-id>",
		Short: "Apply a captured state",
		Long: `Apply a captured git state to all repositories in a workspace.

Examples:
  # Apply capture by ID in current workspace
  workshed apply 01HVABCDEFG

  # Apply capture by name
  workshed apply --name "Before refactor"

  # Apply capture in specific workspace
  workshed apply my-workspace 01HVABCDEFG`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			ctx := context.Background()

			providedHandle, remaining := cli.ExtractHandleFromArgs(args, "--name")
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			captureID := ""
			if name != "" {
				captures, err := r.GetStore().ListCaptures(ctx, handle)
				if err != nil {
					return fmt.Errorf("failed to list captures: %w", err)
				}
				found := false
				for _, c := range captures {
					if c.Name == name {
						captureID = c.ID
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("capture not found: %s", name)
				}
			} else if len(remaining) > 0 {
				captureID = remaining[0]
			} else {
				return fmt.Errorf("missing required argument: <capture-id>")
			}

			capture, err := r.GetStore().GetCapture(ctx, handle, captureID)
			if err != nil {
				return fmt.Errorf("failed to get capture: %w", err)
			}

			preflight, err := r.GetStore().PreflightApply(ctx, handle, captureID)
			if err != nil {
				return fmt.Errorf("preflight check failed: %w", err)
			}

			if !preflight.Valid {
				logger.UncheckedFprintf(cmd.ErrOrStderr(), "ERROR: apply blocked by preflight errors\n\n")
				logger.UncheckedFprintf(cmd.ErrOrStderr(), "Problems found:\n")
				for _, e := range preflight.Errors {
					hint := cli.PreflightErrorHint(e.Reason)
					logger.UncheckedFprintf(cmd.ErrOrStderr(), "  - %s: %s\n", e.Repository, e.Details)
					if hint != "" {
						logger.UncheckedFprintf(cmd.ErrOrStderr(), "    Hint: %s\n", hint)
					}
				}
				logger.UncheckedFprintf(cmd.ErrOrStderr(), "\n")
				return fmt.Errorf("preflight validation failed")
			}

			if dryRun {
				r.GetLogger().Info("dry run - would apply capture", "handle", handle, "capture", captureID)
				return nil
			}

			if err := r.GetStore().ApplyCapture(ctx, handle, captureID); err != nil {
				return fmt.Errorf("apply failed: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "json" {
				data, _ := json.MarshalIndent(capture, "", "  ")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			return cli.RenderKeyValue(map[string]string{
				"id":    captureID,
				"name":  capture.Name,
				"repos": strconv.Itoa(len(capture.GitState)),
			}, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Capture name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be applied")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}
