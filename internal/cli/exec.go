package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/store"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Exec(args []string) {
	l := logger.NewLogger(logger.INFO, "exec")

	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	target := fs.String("repo", "", "Target repository name (default: all repositories)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed exec [<handle>] -- <command>...\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
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
		l.Error("missing command separator", "separator", "--")
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
	opts := store.ExecOptions{
		Target:  *target,
		Command: command,
	}

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
}
