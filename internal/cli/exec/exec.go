package exec

import (
	"context"
	"encoding/json"
	"fmt"
	osexec "os/exec"
	"time"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

type ExecResultOutput struct {
	Repository string `json:"repository"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output"`
	DurationMs int64  `json:"duration_ms"`
}

func Command() *cobra.Command {
	var repo string
	var all bool
	var noRecord bool

	cmd := &cobra.Command{
		Use:   "exec [<handle>] <command> [args...]",
		Short: "Run a command in repositories",
		Long: `Run a command in repositories.

Examples:
  workshed exec make test
  workshed exec -a go test ./...
  workshed exec my-workspace make build`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			if len(args) == 0 {
				return fmt.Errorf("missing command to execute")
			}

			sepIdx := -1
			for i, arg := range args {
				if arg == "--" {
					sepIdx = i
					break
				}
			}

			var flagArgs []string
			var command []string
			if sepIdx == -1 {
				providedHandle, remaining := cli.ExtractHandleFromArgs(args)
				if len(remaining) > 0 {
					flagArgs = []string{providedHandle}
					command = remaining
				} else {
					flagArgs = args
				}
			} else {
				flagArgs = args[:sepIdx]
				command = args[sepIdx+1:]
			}

			if len(command) == 0 {
				firstArg := args[0]
				if _, err := osexec.LookPath(firstArg); err == nil {
					command = []string{firstArg}
				} else {
					return fmt.Errorf("missing command to execute")
				}
			}

			format := cmd.Flags().Lookup("format").Value.String()

			explicitAll := all
			if repo != "" {
				explicitAll = true
			}

			ctx := context.Background()

			providedHandle, _ := cli.ExtractHandleFromArgs(flagArgs)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				if err.Error() == "not in a workspace directory" {
					return fmt.Errorf("not in a workspace directory: run from workspace dir or use -- <handle>")
				}
				return fmt.Errorf("failed to resolve workspace: %w", err)
			}

			opts := workspace.ExecOptions{
				Target:   repo,
				Command:  command,
				Parallel: explicitAll,
			}

			startedAt := time.Now()
			results, err := r.GetStore().Exec(ctx, handle, opts)
			if err != nil {
				return fmt.Errorf("exec failed: %w", err)
			}

			switch format {
			case "json":
				var outputResults []ExecResultOutput
				for _, result := range results {
					outputResults = append(outputResults, ExecResultOutput{
						Repository: result.Repository,
						ExitCode:   result.ExitCode,
						Output:     string(result.Output),
						DurationMs: result.Duration.Milliseconds(),
					})
				}
				data, _ := json.MarshalIndent(outputResults, "", "  ")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			case "raw":
				var outputResults []ExecResultOutput
				for _, result := range results {
					outputResults = append(outputResults, ExecResultOutput{
						Repository: result.Repository,
						ExitCode:   result.ExitCode,
						Output:     string(result.Output),
						DurationMs: result.Duration.Milliseconds(),
					})
				}
				data, _ := json.Marshal(outputResults)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			default:
				for _, result := range results {
					if result.Repository != "root" {
						fmt.Printf("=== %s ===\n", result.Repository)
					}
					if _, err := cmd.OutOrStdout().Write(result.Output); err != nil {
						r.GetLogger().Error("failed to write output", "error", err)
					}
					if len(results) > 1 {
						fmt.Println()
					}
				}
			}

			if !noRecord {
				var maxExitCode int
				repoResults := make([]workspace.ExecutionRepoResult, 0, len(results))
				for _, result := range results {
					if result.ExitCode > maxExitCode {
						maxExitCode = result.ExitCode
					}
					repoResults = append(repoResults, workspace.ExecutionRepoResult{
						Repository: result.Repository,
						ExitCode:   result.ExitCode,
						Duration:   result.Duration.Milliseconds(),
					})
				}

				record := workspace.ExecutionRecord{
					ID:          ulid.Make().String(),
					Timestamp:   startedAt,
					Handle:      handle,
					Target:      repo,
					Command:     command,
					ExitCode:    maxExitCode,
					StartedAt:   startedAt,
					CompletedAt: time.Now(),
					Duration:    time.Since(startedAt).Milliseconds(),
					Results:     repoResults,
				}

				if err := r.GetStore().RecordExecution(ctx, handle, record, nil); err != nil {
					r.GetLogger().Debug("failed to record execution", "error", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Repository name to exec in")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Exec in all repositories")
	cmd.Flags().BoolVar(&noRecord, "no-record", false, "Don't record command execution")
	cmd.Flags().String("format", "stream", "Output format (stream|json|raw)")

	return cmd
}
