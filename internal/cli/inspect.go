package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Inspect(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("inspect", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed inspect [<handle>]\n\n")
		logger.SafeFprintf(r.Stderr, "Show workspace details including repositories and creation time.\n\n")
		logger.SafeFprintf(r.Stderr, "Examples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed inspect\n")
		logger.SafeFprintf(r.Stderr, "  workshed inspect aquatic-fish-motion\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	ws, err := s.Get(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	logger.SafeFprintf(r.Stdout, "Workspace: %s\n", ws.Handle)
	logger.SafeFprintf(r.Stdout, "Purpose: %s\n", ws.Purpose)
	logger.SafeFprintf(r.Stdout, "Path: %s\n", ws.Path)
	logger.SafeFprintf(r.Stdout, "Created: %s\n", ws.CreatedAt.Format("2006-01-02 15:04:05"))
	logger.SafeFprintln(r.Stdout, "Repositories:")
	for _, repo := range ws.Repositories {
		if repo.Ref != "" {
			logger.SafeFprintf(r.Stdout, "  • %s @ %s\n", repo.Name, repo.Ref)
		} else {
			logger.SafeFprintf(r.Stdout, "  • %s\n", repo.Name)
		}
	}
}
