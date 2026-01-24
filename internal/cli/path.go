package cli

import (
	"context"
	"flag"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

// Path prints the filesystem path for a workspace.
func Path(args []string) {
	l := logger.NewLogger(logger.INFO, "path")

	fs := flag.NewFlagSet("path", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed path <handle>\n")
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

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	ctx := context.Background()
	path, err := store.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		exitFunc(1)
	}

	l.Info("workspace path", "path", path)
}
