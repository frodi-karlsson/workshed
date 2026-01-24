package cli

import (
	"context"
	"flag"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

// Inspect displays detailed information about a workspace.
func Inspect(args []string) {
	l := logger.NewLogger(logger.INFO, "inspect")

	fs := flag.NewFlagSet("inspect", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed inspect <handle>\n")
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
	ws, err := store.Get(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	l.Info("workspace details", "handle", ws.Handle, "purpose", ws.Purpose, "created", ws.CreatedAt.Format("2006-01-02 15:04:05"), "path", ws.Path)

	if ws.RepoURL != "" {
		l.Info("repository info", "repo", ws.RepoURL)
		if ws.RepoRef != "" {
			l.Info("reference", "ref", ws.RepoRef)
		}
	}
}
