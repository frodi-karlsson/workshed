package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

func Exec(args []string) {
	l := logger.NewLogger(logger.INFO, "exec")

	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	target := fs.String("repo", "", "Target repository name (default: all repositories)")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed exec <handle> -- <command>...\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	if fs.NArg() < 1 {
		l.Error("missing required argument", "argument", "handle")
		fs.Usage()
		exitFunc(1)
	}

	handle := fs.Arg(0)

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
		exitFunc(1)
	}

	if sepIdx+1 >= len(args) {
		l.Error("missing command to execute")
		fs.Usage()
		exitFunc(1)
	}

	command := args[sepIdx+1:]

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	ctx := context.Background()
	ws, err := store.Get(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	if len(ws.Repositories) == 0 {
		l.Error("workspace has no repositories")
		exitFunc(1)
	}

	opts := workspace.ExecOptions{
		Target:  *target,
		Command: command,
	}

	results, err := store.Exec(ctx, handle, opts)
	if err != nil {
		l.Error("exec failed", "handle", handle, "error", err)
		exitFunc(1)
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
