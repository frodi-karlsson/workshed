package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

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

	for _, repo := range ws.Repositories {
		if repo.Ref != "" {
			l.Info("repository", "name", repo.Name, "url", repo.URL, "ref", repo.Ref)
		} else {
			l.Info("repository", "name", repo.Name, "url", repo.URL)
		}
	}
}
