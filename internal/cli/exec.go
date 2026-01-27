package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/oklog/ulid/v2"
	flag "github.com/spf13/pflag"
)

type ExecResultOutput struct {
	Repository string `json:"repository"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output"`
	DurationMs int64  `json:"duration_ms"`
}

func (r *Runner) Exec(args []string) {
	l := r.getLogger()

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
		flagArgs = args
	} else {
		flagArgs = args[:sepIdx]
		command = args[sepIdx+1:]
	}

	fs := flag.NewFlagSet("exec", flag.ContinueOnError)
	fs.SetOutput(r.Stderr)
	target := fs.String("repo", "", "Target repository name (default: all repositories)")
	allRepos := fs.BoolP("all", "a", false, "Run command in all repositories (same as default)")
	noRecord := fs.Bool("no-record", false, "Do not record this execution")
	format := fs.String("format", "stream", "Output format (stream|json)")

	if err := fs.Parse(flagArgs); err != nil {
		if err == flag.ErrHelp {
			return
		}
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	if err := ValidateFormat(Format(*format), "exec"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	explicitAll := *allRepos
	if *target != "" {
		explicitAll = true
	}

	if sepIdx == -1 {
		nonFlags := fs.Args()
		if len(nonFlags) > 0 && strings.HasPrefix(nonFlags[0], "-") {
			l.Error("missing '--' separator", "hint", "did you forget '--' before the command? Example: workshed exec -- make test")
			fs.Usage()
			r.ExitFunc(1)
			return
		}
		command = nonFlags
	}

	if len(command) == 0 {
		l.Error("missing command to execute")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	ctx := context.Background()

	providedHandle := ""
	flagCount := fs.NFlag()

	if sepIdx == -1 && len(flagArgs) > 0 && flagCount < len(flagArgs) {
		for _, arg := range flagArgs {
			if !strings.HasPrefix(arg, "-") {
				providedHandle = arg
				break
			}
		}
	}

	handle := r.ResolveHandle(ctx, providedHandle, true, l)
	if handle == "" {
		return
	}

	s := r.getStore()
	opts := workspace.ExecOptions{
		Target:   *target,
		Command:  command,
		Parallel: explicitAll,
	}

	startedAt := time.Now()
	results, err := s.Exec(ctx, handle, opts)
	if err != nil {
		l.Error("exec failed", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if *format == "json" {
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
		logger.UncheckedFprintln(r.Stdout, string(data))
	} else {
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

		if err := s.RecordExecution(ctx, handle, record, nil); err != nil {
			l.Debug("failed to record execution", "error", err)
		}
	}
}
