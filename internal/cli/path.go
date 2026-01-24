package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

// Path prints the filesystem path for a workspace.
func Path(args []string) {
	l := logger.NewLogger(logger.INFO, "path")

	fs := flag.NewFlagSet("path", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed path [<handle>]\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	ctx := context.Background()

	var handle string
	if fs.NArg() >= 1 {
		handle = fs.Arg(0)
	} else {
		ws, err := store.FindWorkspace(ctx, ".")
		if err != nil {
			l.Error("failed to find workspace", "error", err)
			exitFunc(1)
			return
		}
		handle = ws.Handle
	}

	path, err := store.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	l.Info("workspace path", "path", path)
}
