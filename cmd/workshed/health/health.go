package health

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health [<handle>]",
		Short: "Check workspace health",
		Long: `Check workspace health and report issues.

Examples:
  # Check health of current workspace
  workshed health

  # Check health of specific workspace
  workshed health my-workspace`,
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

			execs, err := r.GetStore().ListExecutions(ctx, handle, workspace.ListExecutionsOptions{Limit: 100})
			if err != nil {
				return fmt.Errorf("failed to list executions: %w", err)
			}

			captures, _ := r.GetStore().ListCaptures(ctx, handle)

			healthIssues := runHealthChecks(ctx, ws, execs, captures)

			status := "healthy"
			if len(healthIssues) > 0 {
				status = "issues found"
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "table" && len(healthIssues) > 0 {
				fmt.Printf("Issues found:\n\n")
				for _, issue := range healthIssues {
					fmt.Printf("  %s\n", issue)
				}
				fmt.Println()
			}

			return cli.RenderKeyValue(map[string]string{
				"handle": handle,
				"status": status,
			}, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().String("format", "table", "Output format (table|json|raw)")

	return cmd
}

func runHealthChecks(ctx context.Context, ws *workspace.Workspace, execs []workspace.ExecutionRecord, captures []workspace.Capture) []string {
	var issues []string

	staleThreshold := 30 * 24 * time.Hour
	staleCount := 0
	for _, e := range execs {
		if time.Since(e.Timestamp) > staleThreshold {
			staleCount++
		}
	}
	if staleCount > 0 {
		issues = append(issues, fmt.Sprintf("%d stale executions older than 30 days", staleCount))
	}

	gitClient := git.RealGit{}

	for _, repo := range ws.Repositories {
		repoDir := filepath.Join(ws.Path, repo.Name)
		_, err := os.Stat(repoDir)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("missing repository directory: %s", repo.Name))
			}
			continue
		}

		gitDir := filepath.Join(repoDir, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("%s is not a git repository", repo.Name))
			}
		} else {
			status, _ := gitClient.StatusPorcelain(ctx, repoDir)
			if strings.TrimSpace(status) != "" {
				issues = append(issues, fmt.Sprintf("%s has uncommitted changes", repo.Name))
			}
		}
	}

	for _, cap := range captures {
		for _, ref := range cap.GitState {
			repoDir := filepath.Join(ws.Path, ref.Repository)
			if _, err := os.Stat(repoDir); err != nil {
				if os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("capture '%s' references missing repository: %s", cap.Name, ref.Repository))
				}
			}
		}
	}

	return issues
}
