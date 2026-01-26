package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/oklog/ulid/v2"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Exec(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	target := fs.String("repo", "", "Target repository name (default: all repositories)")
	noRecord := fs.Bool("no-record", false, "Do not record this execution")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed exec [<handle>] -- <command>...\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed exec -- make test\n")
		logger.SafeFprintf(r.Stderr, "  workshed exec -- npm run build\n")
		logger.SafeFprintf(r.Stderr, "  workshed exec my-workspace -- go test ./...\n")
		logger.SafeFprintf(r.Stderr, "  workshed exec --repo api -- echo hello\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	sepIdx := -1
	for i, arg := range args {
		if arg == "--" {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		l.Error("missing '--' separator", "hint", "did you forget '--' before the command? Example: workshed exec -- make test")
		fs.Usage()
		r.ExitFunc(1)
	}

	if sepIdx+1 >= len(args) {
		l.Error("missing command to execute")
		fs.Usage()
		r.ExitFunc(1)
	}

	command := args[sepIdx+1:]

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	opts := workspace.ExecOptions{
		Target:  *target,
		Command: command,
	}

	startedAt := time.Now()
	results, err := s.Exec(ctx, handle, opts)
	if err != nil {
		l.Error("exec failed", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	for _, result := range results {
		if result.Repository != "root" {
			fmt.Printf("=== %s ===\n", result.Repository)
		}
		if _, err := os.Stdout.Write(result.Output); err != nil {
			l.Error("failed to write output", "error", err)
		}
		if len(results) > 1 {
			fmt.Println()
		}
	}

	if !*noRecord {
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
			Target:      *target,
			Command:     command,
			ExitCode:    maxExitCode,
			StartedAt:   startedAt,
			CompletedAt: time.Now(),
			Duration:    time.Since(startedAt).Milliseconds(),
			Results:     repoResults,
		}

		if err := s.RecordExecution(ctx, handle, record); err != nil {
			l.Debug("failed to record execution", "error", err)
		}
	}
}
