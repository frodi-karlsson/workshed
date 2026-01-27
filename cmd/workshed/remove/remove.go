package remove

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/frodi/workshed/internal/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var yes bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "remove [<handle>]",
		Short: "Remove a workspace",
		Long: `Delete a workspace and all its repositories.

Examples:
  workshed remove
  workshed remove my-workspace
  workshed remove -y
  workshed remove --dry-run`,
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

			if dryRun {
				r.GetLogger().Info("dry run - would remove workspace", "handle", handle, "purpose", ws.Purpose)
				for _, repo := range ws.Repositories {
					r.GetLogger().Info("  - repository", "name", repo.Name)
				}
				return nil
			}

			if !yes {
				if !term.IsTerminal(os.Stdin.Fd()) {
					r.GetLogger().Warn("stdin is not a tty, cannot prompt", "hint", "use --yes to skip confirmation")
					r.GetLogger().Info("operation cancelled")
					return nil
				}

				prompt := fmt.Sprintf("Remove workspace %q (%s)? [y/N]: ", ws.Handle, ws.Purpose)
				if _, err := fmt.Fprint(cmd.OutOrStdout(), prompt); err != nil {
					return fmt.Errorf("failed to write prompt: %w", err)
				}

				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read user input: %w", err)
				}

				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					r.GetLogger().Info("operation cancelled")
					return nil
				}
			}

			if err := r.GetStore().Remove(ctx, handle); err != nil {
				return fmt.Errorf("failed to remove workspace: %w", err)
			}

			r.GetLogger().Success("workspace removed", "handle", handle)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed")

	return cmd
}
