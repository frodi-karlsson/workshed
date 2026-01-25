package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
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

	store := GetOrCreateStore(l)
	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := ResolveHandle(ctx, store, providedHandle, l)

	path, err := store.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	l.Info("workspace path", "path", path)
}
